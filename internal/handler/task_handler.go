package handler

import (
	"b1cron/internal/database"
	"b1cron/internal/models"
	"b1cron/internal/service"
	"net/http"
	"strconv"
	"time"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type TaskHandler struct {
	taskService *service.TaskService
}

type CreateTaskRequest struct {
	Name         string `json:"name" binding:"required"`
	Command      string `json:"command" binding:"required"`
	ScriptType   string `json:"script_type"`
	ScheduleSpec string `json:"schedule_spec"`
	ScheduleType string `json:"schedule_type"`
	ExecuteAt    string `json:"execute_at"`    // RFC3339 format datetime string
	IsEnabled    bool   `json:"is_enabled"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=6"`
}

func NewTaskHandler(taskService *service.TaskService) *TaskHandler {
	return &TaskHandler{
		taskService: taskService,
	}
}

func (h *TaskHandler) ShowDashboard(c *gin.Context) {
	// 检查用户是否需要强制修改密码
	claims := jwt.ExtractClaims(c)
	username, ok := claims["username"].(string)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}
	
	var user models.User
	if err := database.GetDB().Where("username = ?", username).First(&user).Error; err != nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}
	
	if user.ForcePasswordChange {
		// 如果需要强制修改密码，重定向到登录页面
		c.Redirect(http.StatusFound, "/login")
		return
	}
	
	tasks, err := h.taskService.GetAllTasks()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 计算任务统计信息
	var enabledCount, disabledCount int
	for _, task := range tasks {
		if task.IsEnabled {
			enabledCount++
		} else {
			disabledCount++
		}
	}

	// 获取执行统计
	executionStats, err := h.taskService.GetExecutionStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 获取最近执行记录
	recentExecutions, err := h.taskService.GetRecentExecutions(20)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.HTML(http.StatusOK, "dashboard.html", gin.H{
		"title":            "Dashboard",
		"tasks":            tasks,
		"totalTasks":       len(tasks),
		"enabledTasks":     enabledCount,
		"disabledTasks":    disabledCount,
		"executionStats":   executionStats,
		"recentExecutions": recentExecutions,
	})
}

func (h *TaskHandler) CreateTask(c *gin.Context) {
	var req CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 设置默认值
	if req.ScriptType == "" {
		req.ScriptType = "command"
	}
	if req.ScheduleType == "" {
		req.ScheduleType = "cron"
	}

	// 解析执行时间（如果是一次性任务）
	var executeAt *time.Time
	if req.ScheduleType == "once" {
		if req.ExecuteAt == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "execute_at is required for one-time tasks"})
			return
		}
		
		parsedTime, err := time.ParseInLocation("2006-01-02T15:04", req.ExecuteAt, time.Local)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid execute_at format, expected YYYY-MM-DDTHH:MM"})
			return
		}
		executeAt = &parsedTime
	} else {
		// 周期性任务需要schedule_spec
		if req.ScheduleSpec == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "schedule_spec is required for cron tasks"})
			return
		}
	}

	task, err := h.taskService.CreateTaskFull(req.Name, req.Command, req.ScriptType, req.ScheduleSpec, req.ScheduleType, executeAt, req.IsEnabled)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if c.GetHeader("HX-Request") == "true" {
		c.HTML(http.StatusOK, "_task_row.html", task)
		return
	}

	c.JSON(http.StatusCreated, task)
}

func (h *TaskHandler) GetTasks(c *gin.Context) {
	tasks, err := h.taskService.GetAllTasks()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tasks)
}

func (h *TaskHandler) GetTask(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	task, err := h.taskService.GetTaskByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	// 如果是脚本类型，获取脚本内容
	taskData := map[string]interface{}{
		"id":            task.ID,
		"name":          task.Name,
		"command":       h.taskService.GetTaskScriptContent(task),
		"script_type":   task.ScriptType,
		"script_path":   task.ScriptPath,
		"schedule_spec": task.ScheduleSpec,
		"is_enabled":    task.IsEnabled,
		"created_at":    task.CreatedAt,
		"updated_at":    task.UpdatedAt,
	}

	c.JSON(http.StatusOK, taskData)
}

func (h *TaskHandler) UpdateTask(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	var req CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 设置默认值
	if req.ScriptType == "" {
		req.ScriptType = "command"
	}
	if req.ScheduleType == "" {
		req.ScheduleType = "cron"
	}

	// 解析执行时间（如果是一次性任务）
	var executeAt *time.Time
	if req.ScheduleType == "once" {
		if req.ExecuteAt == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "execute_at is required for one-time tasks"})
			return
		}
		
		parsedTime, err := time.ParseInLocation("2006-01-02T15:04", req.ExecuteAt, time.Local)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid execute_at format, expected YYYY-MM-DDTHH:MM"})
			return
		}
		executeAt = &parsedTime
	} else {
		// 周期性任务需要schedule_spec
		if req.ScheduleSpec == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "schedule_spec is required for cron tasks"})
			return
		}
	}

	task, err := h.taskService.UpdateTaskFull(uint(id), req.Name, req.Command, req.ScriptType, req.ScheduleSpec, req.ScheduleType, executeAt, req.IsEnabled)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, task)
}

func (h *TaskHandler) DeleteTask(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	if err := h.taskService.DeleteTask(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if c.GetHeader("HX-Request") == "true" {
		c.Status(http.StatusOK)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task deleted successfully"})
}

func (h *TaskHandler) ToggleTask(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	task, err := h.taskService.ToggleTask(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if c.GetHeader("HX-Request") == "true" {
		c.HTML(http.StatusOK, "_task_row.html", task)
		return
	}

	c.JSON(http.StatusOK, task)
}

func (h *TaskHandler) ChangePassword(c *gin.Context) {
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取当前用户信息
	claims := jwt.ExtractClaims(c)
	username, ok := claims["username"].(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user"})
		return
	}

	// 查找用户
	var user models.User
	if err := database.GetDB().Where("username = ?", username).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// 验证当前密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.CurrentPassword)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "当前密码不正确"})
		return
	}

	// 生成新密码哈希
	newHashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// 更新密码和ForcePasswordChange标志
	updates := map[string]interface{}{
		"password_hash":         string(newHashedPassword),
		"force_password_change": false,
	}
	if err := database.GetDB().Model(&user).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password updated successfully"})
}

func (h *TaskHandler) GetTaskExecutions(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	limit := 50 // 默认限制50条记录
	if limitParam := c.Query("limit"); limitParam != "" {
		if parsedLimit, err := strconv.Atoi(limitParam); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	executions, err := h.taskService.GetTaskExecutions(uint(id), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, executions)
}

func (h *TaskHandler) GetRecentExecutions(c *gin.Context) {
	// 检查是否使用新的分页API
	search := c.Query("search")
	pageStr := c.Query("page")
	pageSizeStr := c.Query("page_size")
	
	if search != "" || pageStr != "" || pageSizeStr != "" {
		// 使用新的分页API
		page := 1
		if pageStr != "" {
			if parsedPage, err := strconv.Atoi(pageStr); err == nil && parsedPage > 0 {
				page = parsedPage
			}
		}
		
		pageSize := 20
		if pageSizeStr != "" {
			if parsedPageSize, err := strconv.Atoi(pageSizeStr); err == nil && parsedPageSize > 0 && parsedPageSize <= 100 {
				pageSize = parsedPageSize
			}
		}
		
		executions, total, err := h.taskService.GetRecentExecutionsWithPagination(search, page, pageSize)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		
		// 计算分页信息
		totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))
		
		c.JSON(http.StatusOK, gin.H{
			"executions": executions,
			"pagination": gin.H{
				"current_page": page,
				"page_size":    pageSize,
				"total_items":  total,
				"total_pages":  totalPages,
				"has_next":     page < totalPages,
				"has_prev":     page > 1,
			},
		})
		return
	}
	
	// 保持向后兼容的旧API
	limit := 20 // 默认限制20条记录
	if limitParam := c.Query("limit"); limitParam != "" {
		if parsedLimit, err := strconv.Atoi(limitParam); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	executions, err := h.taskService.GetRecentExecutions(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, executions)
}