package database

import (
	"b1cron/internal/models"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDatabase(dbPath string) error {
	// 确保数据库目录存在
	dbDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return fmt.Errorf("failed to create database directory: %w", err)
	}
	
	var err error
	
	DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// 先进行 AutoMigrate
	err = DB.AutoMigrate(&models.User{}, &models.Task{}, &models.TaskExecution{})
	if err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	// 手动处理 Command 字段的迁移
	if err := migrateTaskCommand(); err != nil {
		return fmt.Errorf("failed to migrate task command: %w", err)
	}

	log.Printf("Database initialized successfully at: %s", dbPath)
	return nil
}

func migrateTaskCommand() error {
	// 检查 command 字段是否存在
	if !DB.Migrator().HasColumn(&models.Task{}, "command") {
		// 如果字段不存在，添加它
		if err := DB.Migrator().AddColumn(&models.Task{}, "command"); err != nil {
			return err
		}
	}
	
	// 为所有没有 command 的现有任务设置默认值
	var count int64
	if err := DB.Model(&models.Task{}).Where("command IS NULL OR command = ''").Count(&count).Error; err != nil {
		return err
	}
	
	if count > 0 {
		if err := DB.Model(&models.Task{}).Where("command IS NULL OR command = ''").Update("command", "echo 'No command specified'").Error; err != nil {
			return err
		}
		log.Printf("Updated %d tasks with default command", count)
	}
	
	return nil
}

func GetDB() *gorm.DB {
	return DB
}