package scheduler

import (
	"b1cron/internal/database"
	"b1cron/internal/models"
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
)

type SchedulerService struct {
	scheduler gocron.Scheduler
}

func NewSchedulerService() (*SchedulerService, error) {
	s, err := gocron.NewScheduler()
	if err != nil {
		return nil, err
	}

	return &SchedulerService{
		scheduler: s,
	}, nil
}

func (s *SchedulerService) Start() error {
	s.scheduler.Start()
	
	if err := s.loadExistingTasks(); err != nil {
		return fmt.Errorf("failed to load existing tasks: %w", err)
	}
	
	log.Println("Scheduler service started and existing tasks loaded")
	return nil
}

func (s *SchedulerService) Stop() error {
	return s.scheduler.Shutdown()
}

func (s *SchedulerService) ScheduleTask(task *models.Task) (uuid.UUID, error) {
	job, err := s.createJob(task)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to create job: %w", err)
	}

	jobUUID := job.ID()
	log.Printf("Task '%s' scheduled with ID: %s", task.Name, jobUUID)
	return jobUUID, nil
}

func (s *SchedulerService) UnscheduleTask(jobID uuid.UUID) error {
	err := s.scheduler.RemoveJob(jobID)
	if err != nil {
		return fmt.Errorf("failed to remove job %s: %w", jobID, err)
	}
	
	log.Printf("Task with ID %s unscheduled", jobID)
	return nil
}

func (s *SchedulerService) UpdateTask(task *models.Task) (uuid.UUID, error) {
	if task.GocronJobID != uuid.Nil {
		if err := s.UnscheduleTask(task.GocronJobID); err != nil {
			log.Printf("Warning: failed to unschedule old task: %v", err)
		}
	}
	
	return s.ScheduleTask(task)
}

func (s *SchedulerService) createJob(task *models.Task) (gocron.Job, error) {
	taskFunc := func() {
		s.executeTask(task)
	}

	// 处理一次性任务
	if task.ScheduleType == "once" {
		if task.ExecuteAt == nil {
			return nil, fmt.Errorf("execute_at is required for one-time tasks")
		}
		
		// 检查执行时间是否已过，如果已过则不调度
		if task.ExecuteAt.Before(time.Now()) {
			log.Printf("One-time task '%s' execution time has passed, skipping scheduling", task.Name)
			// 自动禁用已过期的一次性任务
			if err := database.GetDB().Model(task).Update("is_enabled", false).Error; err != nil {
				log.Printf("Failed to disable expired one-time task %s: %v", task.Name, err)
			}
			return nil, fmt.Errorf("execution time has passed")
		}
		
		return s.scheduler.NewJob(
			gocron.OneTimeJob(gocron.OneTimeJobStartDateTime(*task.ExecuteAt)),
			gocron.NewTask(taskFunc),
		)
	}

	if strings.HasPrefix(task.ScheduleSpec, "@every") {
		duration, err := s.parseDuration(task.ScheduleSpec)
		if err != nil {
			return nil, err
		}
		return s.scheduler.NewJob(
			gocron.DurationJob(duration),
			gocron.NewTask(taskFunc),
		)
	}

	// 处理标准Cron表达式
	cronSpec := s.normalizeCronSpec(task.ScheduleSpec)
	return s.scheduler.NewJob(
		gocron.CronJob(cronSpec, false),
		gocron.NewTask(taskFunc),
	)
}

func (s *SchedulerService) executeTask(task *models.Task) {
	log.Printf("Executing task: %s (ID: %d)", task.Name, task.ID)
	
	// 检查命令是否为空
	if task.Command == "" {
		log.Printf("Task '%s' has no command to execute", task.Name)
		return
	}
	
	startTime := time.Now()
	
	// 创建执行记录
	execution := &models.TaskExecution{
		TaskID:    task.ID,
		Status:    "running",
		StartedAt: startTime,
	}
	
	// 保存执行记录到数据库
	if err := database.GetDB().Create(execution).Error; err != nil {
		log.Printf("Failed to create execution record: %v", err)
	}
	
	// 执行命令 - 统一通过shell执行以支持重定向、管道等操作
	var cmd *exec.Cmd
	
	// 检测操作系统并使用相应的shell
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", task.Command)
	} else {
		cmd = exec.Command("/bin/sh", "-c", task.Command)
	}
	
	output, err := cmd.CombinedOutput()
	completedAt := time.Now()
	duration := completedAt.Sub(startTime)
	
	// 更新执行记录
	execution.CompletedAt = &completedAt
	execution.Duration = duration.Milliseconds()
	execution.Output = string(output)
	
	if err != nil {
		execution.Status = "failed"
		execution.ErrorMsg = err.Error()
		log.Printf("Task '%s' failed after %v: %v\nOutput: %s", 
			task.Name, duration, err, string(output))
	} else {
		execution.Status = "success"
		log.Printf("Task '%s' completed successfully in %v\nOutput: %s", 
			task.Name, duration, string(output))
	}
	
	// 保存更新后的执行记录
	if err := database.GetDB().Save(execution).Error; err != nil {
		log.Printf("Failed to update execution record: %v", err)
	}

	// 如果是一次性任务，执行完成后自动禁用并从调度器中移除
	if task.ScheduleType == "once" {
		if err := database.GetDB().Model(task).Update("is_enabled", false).Error; err != nil {
			log.Printf("Failed to disable one-time task %s: %v", task.Name, err)
		} else {
			log.Printf("One-time task '%s' completed and disabled", task.Name)
		}
		
		// 从调度器中移除任务
		if task.GocronJobID != uuid.Nil {
			if err := s.UnscheduleTask(task.GocronJobID); err != nil {
				log.Printf("Failed to unschedule completed one-time task %s: %v", task.Name, err)
			} else {
				log.Printf("One-time task '%s' removed from scheduler", task.Name)
			}
		}
	}
}

// normalizeCronSpec 确保Cron表达式是5字段格式
func (s *SchedulerService) normalizeCronSpec(spec string) string {
	parts := strings.Fields(spec)
	
	// 如果是6字段格式 (秒 分 时 日 月 周)，转换为5字段格式 (分 时 日 月 周)
	if len(parts) == 6 {
		// 移除秒字段，保留后5个字段
		return strings.Join(parts[1:], " ")
	}
	
	// 如果已经是5字段格式，直接返回
	if len(parts) == 5 {
		return spec
	}
	
	// 格式错误的情况
	return spec
}

func (s *SchedulerService) parseDuration(spec string) (time.Duration, error) {
	parts := strings.Fields(spec)
	if len(parts) != 2 || parts[0] != "@every" {
		return 0, fmt.Errorf("invalid duration format: %s", spec)
	}
	
	return time.ParseDuration(parts[1])
}

func (s *SchedulerService) loadExistingTasks() error {
	var tasks []models.Task
	if err := database.GetDB().Where("is_enabled = ?", true).Find(&tasks).Error; err != nil {
		return err
	}

	for _, task := range tasks {
		jobUUID, err := s.ScheduleTask(&task)
		if err != nil {
			log.Printf("Failed to reschedule task %s: %v", task.Name, err)
			continue
		}

		if err := database.GetDB().Model(&task).Update("gocron_job_id", jobUUID).Error; err != nil {
			log.Printf("Failed to update job ID for task %s: %v", task.Name, err)
		}
	}

	log.Printf("Loaded %d existing tasks", len(tasks))
	return nil
}