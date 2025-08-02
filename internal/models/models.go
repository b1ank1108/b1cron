package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID                  uint   `gorm:"primaryKey" json:"id"`
	Username            string `gorm:"uniqueIndex;not null" json:"username"`
	PasswordHash        string `gorm:"not null" json:"-"`
	ForcePasswordChange bool   `gorm:"default:false" json:"force_password_change"`
	CreatedAt           time.Time
	UpdatedAt           time.Time
	DeletedAt           gorm.DeletedAt `gorm:"index"`
}

type Task struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Name         string    `gorm:"not null" json:"name"`
	Command      string    `json:"command"`
	ScriptType   string    `gorm:"default:'command'" json:"script_type"` // command, shell, python
	ScriptPath   string    `json:"script_path"`                          // relative path to script file
	ScheduleSpec string    `gorm:"not null" json:"schedule_spec"`
	ScheduleType string    `gorm:"default:'cron'" json:"schedule_type"`  // cron, once, range, dynamic
	ExecuteAt    *time.Time `json:"execute_at"`                         // for one-time execution
	IsEnabled    bool      `gorm:"default:true" json:"is_enabled"`
	GocronJobID  uuid.UUID `gorm:"type:char(36)" json:"gocron_job_id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

type TaskExecution struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	TaskID      uint      `gorm:"not null;index" json:"task_id"`
	Task        Task      `gorm:"foreignKey:TaskID" json:"task,omitempty"`
	Status      string    `gorm:"not null" json:"status"` // success, failed, running
	StartedAt   time.Time `gorm:"not null" json:"started_at"`
	CompletedAt *time.Time `json:"completed_at"`
	Duration    int64     `json:"duration"` // milliseconds
	Output      string    `gorm:"type:text" json:"output"`
	ErrorMsg    string    `gorm:"type:text" json:"error_msg"`
	CreatedAt   time.Time `json:"created_at"`
}