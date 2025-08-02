package service

import (
	"b1cron/internal/config"
	"b1cron/internal/database"
	"b1cron/internal/models"
	"b1cron/internal/scheduler"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TaskService struct {
	schedulerService *scheduler.SchedulerService
	scriptService    *ScriptFileService
}

func NewTaskService(schedulerService *scheduler.SchedulerService, cfg *config.Config) *TaskService {
	return &TaskService{
		schedulerService: schedulerService,
		scriptService:    NewScriptFileService(cfg),
	}
}

func (s *TaskService) CreateTask(name, command, scheduleSpec string, isEnabled bool) (*models.Task, error) {
	return s.CreateTaskWithScript(name, command, "command", scheduleSpec, isEnabled)
}

func (s *TaskService) CreateTaskWithScript(name, command, scriptType, scheduleSpec string, isEnabled bool) (*models.Task, error) {
	// 验证脚本内容
	if err := s.scriptService.ValidateScriptContent(scriptType, command); err != nil {
		return nil, fmt.Errorf("invalid script content: %w", err)
	}

	// 创建任务记录（先不保存到数据库）
	task := &models.Task{
		Name:         name,
		ScriptType:   scriptType,
		ScheduleSpec: scheduleSpec,
		IsEnabled:    isEnabled,
	}

	// 先保存到数据库获取ID
	if err := database.GetDB().Create(task).Error; err != nil {
		return nil, fmt.Errorf("failed to create task in database: %w", err)
	}

	// 创建脚本文件并获取执行命令
	scriptPath, execCommand, err := s.scriptService.CreateScriptFile(task.ID, name, scriptType, command)
	if err != nil {
		// 回滚数据库记录
		database.GetDB().Delete(task)
		return nil, fmt.Errorf("failed to create script file: %w", err)
	}

	// 更新任务的脚本路径和执行命令
	task.ScriptPath = scriptPath
	task.Command = execCommand

	// 如果启用任务，进行调度
	if isEnabled {
		jobUUID, err := s.schedulerService.ScheduleTask(task)
		if err != nil {
			// 清理脚本文件和数据库记录
			s.scriptService.DeleteScriptFile(scriptPath)
			database.GetDB().Delete(task)
			return nil, fmt.Errorf("failed to schedule task: %w", err)
		}
		task.GocronJobID = jobUUID
	}

	// 保存更新的任务信息
	if err := database.GetDB().Save(task).Error; err != nil {
		// 清理调度任务和脚本文件
		if isEnabled && task.GocronJobID != uuid.Nil {
			s.schedulerService.UnscheduleTask(task.GocronJobID)
		}
		s.scriptService.DeleteScriptFile(scriptPath)
		database.GetDB().Delete(task)
		return nil, fmt.Errorf("failed to update task in database: %w", err)
	}

	return task, nil
}

// CreateTaskFull 创建完整的任务（支持调度类型和执行时间）
func (s *TaskService) CreateTaskFull(name, command, scriptType, scheduleSpec, scheduleType string, executeAt *time.Time, isEnabled bool) (*models.Task, error) {
	// 验证脚本内容
	if err := s.scriptService.ValidateScriptContent(scriptType, command); err != nil {
		return nil, fmt.Errorf("invalid script content: %w", err)
	}

	// 验证一次性任务的执行时间
	if scheduleType == "once" {
		if executeAt == nil {
			return nil, fmt.Errorf("execute_at is required for one-time tasks")
		}
		if executeAt.Before(time.Now()) {
			return nil, fmt.Errorf("execute_at must be in the future")
		}
	}

	// 创建任务记录
	task := &models.Task{
		Name:         name,
		ScriptType:   scriptType,
		ScheduleSpec: scheduleSpec,
		ScheduleType: scheduleType,
		ExecuteAt:    executeAt,
		IsEnabled:    isEnabled,
	}

	// 先保存到数据库获取ID
	if err := database.GetDB().Create(task).Error; err != nil {
		return nil, fmt.Errorf("failed to create task in database: %w", err)
	}

	// 创建脚本文件并获取执行命令
	scriptPath, execCommand, err := s.scriptService.CreateScriptFile(task.ID, name, scriptType, command)
	if err != nil {
		// 回滚数据库记录
		database.GetDB().Delete(task)
		return nil, fmt.Errorf("failed to create script file: %w", err)
	}

	// 更新任务的脚本路径和执行命令
	task.ScriptPath = scriptPath
	task.Command = execCommand

	// 如果启用任务，进行调度
	if isEnabled {
		jobUUID, err := s.schedulerService.ScheduleTask(task)
		if err != nil {
			// 清理脚本文件和数据库记录
			s.scriptService.DeleteScriptFile(scriptPath)
			database.GetDB().Delete(task)
			return nil, fmt.Errorf("failed to schedule task: %w", err)
		}
		task.GocronJobID = jobUUID
	}

	// 保存更新的任务信息
	if err := database.GetDB().Save(task).Error; err != nil {
		// 清理调度任务和脚本文件
		if isEnabled && task.GocronJobID != uuid.Nil {
			s.schedulerService.UnscheduleTask(task.GocronJobID)
		}
		s.scriptService.DeleteScriptFile(scriptPath)
		database.GetDB().Delete(task)
		return nil, fmt.Errorf("failed to update task in database: %w", err)
	}

	return task, nil
}

func (s *TaskService) GetAllTasks() ([]models.Task, error) {
	var tasks []models.Task
	if err := database.GetDB().Find(&tasks).Error; err != nil {
		return nil, fmt.Errorf("failed to get tasks: %w", err)
	}
	return tasks, nil
}

func (s *TaskService) GetTaskByID(id uint) (*models.Task, error) {
	var task models.Task
	if err := database.GetDB().First(&task, id).Error; err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}
	return &task, nil
}

func (s *TaskService) UpdateTask(id uint, name, command, scheduleSpec string, isEnabled bool) (*models.Task, error) {
	return s.UpdateTaskWithScript(id, name, command, "command", scheduleSpec, isEnabled)
}

func (s *TaskService) UpdateTaskWithScript(id uint, name, command, scriptType, scheduleSpec string, isEnabled bool) (*models.Task, error) {
	// 验证脚本内容
	if err := s.scriptService.ValidateScriptContent(scriptType, command); err != nil {
		return nil, fmt.Errorf("invalid script content: %w", err)
	}

	task, err := s.GetTaskByID(id)
	if err != nil {
		return nil, err
	}

	// 取消当前调度
	if task.GocronJobID != uuid.Nil {
		if err := s.schedulerService.UnscheduleTask(task.GocronJobID); err != nil {
			fmt.Printf("Warning: failed to unschedule old task: %v\n", err)
		}
		task.GocronJobID = uuid.Nil
	}

	// 更新脚本文件
	oldScriptPath := task.ScriptPath
	scriptPath, execCommand, err := s.scriptService.UpdateScriptFile(id, name, scriptType, command, oldScriptPath)
	if err != nil {
		return nil, fmt.Errorf("failed to update script file: %w", err)
	}

	// 更新任务信息
	task.Name = name
	task.Command = execCommand
	task.ScriptType = scriptType
	task.ScriptPath = scriptPath
	task.ScheduleSpec = scheduleSpec
	task.IsEnabled = isEnabled

	// 如果启用任务，重新调度
	if isEnabled {
		jobUUID, err := s.schedulerService.ScheduleTask(task)
		if err != nil {
			return nil, fmt.Errorf("failed to schedule updated task: %w", err)
		}
		task.GocronJobID = jobUUID
	}

	// 保存到数据库
	if err := database.GetDB().Save(task).Error; err != nil {
		// 回滚调度
		if isEnabled && task.GocronJobID != uuid.Nil {
			if unscheduleErr := s.schedulerService.UnscheduleTask(task.GocronJobID); unscheduleErr != nil {
				fmt.Printf("Warning: failed to rollback scheduled task: %v\n", unscheduleErr)
			}
		}
		return nil, fmt.Errorf("failed to update task in database: %w", err)
	}

	return task, nil
}

// UpdateTaskFull 更新完整的任务（支持调度类型和执行时间）
func (s *TaskService) UpdateTaskFull(id uint, name, command, scriptType, scheduleSpec, scheduleType string, executeAt *time.Time, isEnabled bool) (*models.Task, error) {
	// 验证脚本内容
	if err := s.scriptService.ValidateScriptContent(scriptType, command); err != nil {
		return nil, fmt.Errorf("invalid script content: %w", err)
	}

	// 验证一次性任务的执行时间
	if scheduleType == "once" {
		if executeAt == nil {
			return nil, fmt.Errorf("execute_at is required for one-time tasks")
		}
		if executeAt.Before(time.Now()) {
			return nil, fmt.Errorf("execute_at must be in the future")
		}
	}

	task, err := s.GetTaskByID(id)
	if err != nil {
		return nil, err
	}

	// 取消当前调度
	if task.GocronJobID != uuid.Nil {
		if err := s.schedulerService.UnscheduleTask(task.GocronJobID); err != nil {
			fmt.Printf("Warning: failed to unschedule old task: %v\n", err)
		}
		task.GocronJobID = uuid.Nil
	}

	// 更新脚本文件
	oldScriptPath := task.ScriptPath
	scriptPath, execCommand, err := s.scriptService.UpdateScriptFile(id, name, scriptType, command, oldScriptPath)
	if err != nil {
		return nil, fmt.Errorf("failed to update script file: %w", err)
	}

	// 更新任务信息
	task.Name = name
	task.Command = execCommand
	task.ScriptType = scriptType
	task.ScriptPath = scriptPath
	task.ScheduleSpec = scheduleSpec
	task.ScheduleType = scheduleType
	task.ExecuteAt = executeAt
	task.IsEnabled = isEnabled

	// 如果启用任务，重新调度
	if isEnabled {
		jobUUID, err := s.schedulerService.ScheduleTask(task)
		if err != nil {
			return nil, fmt.Errorf("failed to schedule updated task: %w", err)
		}
		task.GocronJobID = jobUUID
	}

	// 保存到数据库
	if err := database.GetDB().Save(task).Error; err != nil {
		// 回滚调度
		if isEnabled && task.GocronJobID != uuid.Nil {
			if unscheduleErr := s.schedulerService.UnscheduleTask(task.GocronJobID); unscheduleErr != nil {
				fmt.Printf("Warning: failed to rollback scheduled task: %v\n", unscheduleErr)
			}
		}
		return nil, fmt.Errorf("failed to update task in database: %w", err)
	}

	return task, nil
}

func (s *TaskService) DeleteTask(id uint) error {
	task, err := s.GetTaskByID(id)
	if err != nil {
		return err
	}

	// 取消调度
	if task.GocronJobID != uuid.Nil {
		if err := s.schedulerService.UnscheduleTask(task.GocronJobID); err != nil {
			return fmt.Errorf("failed to unschedule task: %w", err)
		}
	}

	// 删除脚本文件
	if err := s.scriptService.DeleteScriptFile(task.ScriptPath); err != nil {
		fmt.Printf("Warning: failed to delete script file: %v\n", err)
	}

	// 从数据库删除
	if err := database.GetDB().Delete(task).Error; err != nil {
		return fmt.Errorf("failed to delete task from database: %w", err)
	}

	return nil
}

func (s *TaskService) ToggleTask(id uint) (*models.Task, error) {
	task, err := s.GetTaskByID(id)
	if err != nil {
		return nil, err
	}

	return s.UpdateTaskWithScript(id, task.Name, s.GetTaskScriptContent(task), task.ScriptType, task.ScheduleSpec, !task.IsEnabled)
}

func (s *TaskService) GetTaskScriptContent(task *models.Task) string {
	if task.ScriptType == "command" {
		return task.Command
	}
	
	content, err := s.scriptService.GetScriptContent(task.ScriptPath)
	if err != nil {
		fmt.Printf("Warning: failed to read script content: %v\n", err)
		return ""
	}
	return content
}

func (s *TaskService) GetTaskExecutions(taskID uint, limit int) ([]models.TaskExecution, error) {
	var executions []models.TaskExecution
	query := database.GetDB().Where("task_id = ?", taskID).Order("started_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if err := query.Find(&executions).Error; err != nil {
		return nil, fmt.Errorf("failed to get task executions: %w", err)
	}
	return executions, nil
}

func (s *TaskService) GetRecentExecutions(limit int) ([]models.TaskExecution, error) {
	var executions []models.TaskExecution
	query := database.GetDB().Preload("Task", func(db *gorm.DB) *gorm.DB {
		return db.Unscoped()
	}).Order("started_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if err := query.Find(&executions).Error; err != nil {
		return nil, fmt.Errorf("failed to get recent executions: %w", err)
	}
	return executions, nil
}

func (s *TaskService) GetExecutionStats() (map[string]interface{}, error) {
	var stats map[string]interface{} = make(map[string]interface{})
	
	// 总执行次数
	var totalExecutions int64
	if err := database.GetDB().Model(&models.TaskExecution{}).Count(&totalExecutions).Error; err != nil {
		return nil, fmt.Errorf("failed to count total executions: %w", err)
	}
	stats["total_executions"] = totalExecutions
	
	// 成功执行次数
	var successExecutions int64
	if err := database.GetDB().Model(&models.TaskExecution{}).Where("status = ?", "success").Count(&successExecutions).Error; err != nil {
		return nil, fmt.Errorf("failed to count success executions: %w", err)
	}
	stats["success_executions"] = successExecutions
	
	// 失败执行次数
	var failedExecutions int64
	if err := database.GetDB().Model(&models.TaskExecution{}).Where("status = ?", "failed").Count(&failedExecutions).Error; err != nil {
		return nil, fmt.Errorf("failed to count failed executions: %w", err)
	}
	stats["failed_executions"] = failedExecutions
	
	// 计算成功率
	var successRate float64 = 0
	if totalExecutions > 0 {
		successRate = float64(successExecutions) / float64(totalExecutions) * 100
	}
	stats["success_rate"] = successRate
	
	return stats, nil
}