package service

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"b1cron/internal/config"
)

// ScriptFileService 脚本文件管理服务
type ScriptFileService struct {
	dataDir string
	config  *config.Config
}

// NewScriptFileService 创建脚本文件服务
func NewScriptFileService(cfg *config.Config) *ScriptFileService {
	dataDir := "data"
	if cfg != nil && cfg.Database.Path != "" {
		dataDir = filepath.Dir(cfg.Database.Path)
	}
	return &ScriptFileService{
		dataDir: dataDir,
		config:  cfg,
	}
}

// CreateScriptFile 创建脚本文件
func (s *ScriptFileService) CreateScriptFile(taskID uint, taskName, scriptType, content string) (string, string, error) {
	if scriptType == "command" {
		return "", content, nil
	}

	// 创建脚本目录
	scriptDir := filepath.Join(s.dataDir, scriptType)
	if err := os.MkdirAll(scriptDir, 0755); err != nil {
		return "", "", fmt.Errorf("failed to create script directory: %w", err)
	}

	// 生成文件名和路径
	ext := s.getScriptExtension(scriptType)
	if ext == "" {
		return "", "", fmt.Errorf("unsupported script type: %s", scriptType)
	}

	filename := fmt.Sprintf("task_%d.%s", taskID, ext)
	scriptPath := filepath.Join(scriptType, filename)
	fullPath := filepath.Join(s.dataDir, scriptPath)

	// 生成脚本内容
	scriptContent := s.generateScriptContent(taskID, taskName, scriptType, content)

	// 写入脚本文件
	if err := os.WriteFile(fullPath, []byte(scriptContent), 0755); err != nil {
		return "", "", fmt.Errorf("failed to write script file: %w", err)
	}

	// 生成执行命令
	execCommand := s.generateExecCommand(scriptType, fullPath)

	return scriptPath, execCommand, nil
}

// UpdateScriptFile 更新脚本文件
func (s *ScriptFileService) UpdateScriptFile(taskID uint, taskName, scriptType, content, oldScriptPath string) (string, string, error) {
	// 删除旧脚本文件
	if oldScriptPath != "" {
		s.DeleteScriptFile(oldScriptPath)
	}

	// 更新脚本文件时，检查内容是否已经包含头部注释
	// 如果包含我们的头部注释格式，直接使用；否则添加头部注释
	if s.hasScriptHeader(content) {
		return s.CreateScriptFileWithContent(taskID, scriptType, content)
	} else {
		return s.CreateScriptFile(taskID, taskName, scriptType, content)
	}
}

// CreateScriptFileWithContent 使用完整内容创建脚本文件（不添加头部注释）
func (s *ScriptFileService) CreateScriptFileWithContent(taskID uint, scriptType, content string) (string, string, error) {
	if scriptType == "command" {
		return "", content, nil
	}

	// 创建脚本目录
	scriptDir := filepath.Join(s.dataDir, scriptType)
	if err := os.MkdirAll(scriptDir, 0755); err != nil {
		return "", "", fmt.Errorf("failed to create script directory: %w", err)
	}

	// 生成文件名和路径
	ext := s.getScriptExtension(scriptType)
	if ext == "" {
		return "", "", fmt.Errorf("unsupported script type: %s", scriptType)
	}

	filename := fmt.Sprintf("task_%d.%s", taskID, ext)
	scriptPath := filepath.Join(scriptType, filename)
	fullPath := filepath.Join(s.dataDir, scriptPath)

	// 直接写入用户提供的完整内容
	if err := os.WriteFile(fullPath, []byte(content), 0755); err != nil {
		return "", "", fmt.Errorf("failed to write script file: %w", err)
	}

	// 生成执行命令
	execCommand := s.generateExecCommand(scriptType, fullPath)

	return scriptPath, execCommand, nil
}

// hasScriptHeader 检查内容是否已包含我们的脚本头部注释
func (s *ScriptFileService) hasScriptHeader(content string) bool {
	return strings.Contains(content, "# B1Cron Task:") || 
		   strings.Contains(content, "Task ID:")
}

// DeleteScriptFile 删除脚本文件
func (s *ScriptFileService) DeleteScriptFile(scriptPath string) error {
	if scriptPath == "" {
		return nil
	}

	fullPath := filepath.Join(s.dataDir, scriptPath)
	if err := os.Remove(fullPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete script file: %w", err)
	}

	return nil
}

// getScriptExtension 获取脚本文件扩展名
func (s *ScriptFileService) getScriptExtension(scriptType string) string {
	extensions := map[string]string{
		"shell":  "sh",
		"python": "py",
	}
	return extensions[scriptType]
}

// generateScriptContent 生成脚本内容
func (s *ScriptFileService) generateScriptContent(taskID uint, taskName, scriptType, userContent string) string {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	
	var header string
	switch scriptType {
	case "shell":
		header = fmt.Sprintf(`#!/bin/bash
# B1Cron Task: %s
# Created: %s
# Task ID: %d

`, taskName, timestamp, taskID)
	case "python":
		header = fmt.Sprintf(`#!/usr/bin/env python3
# -*- coding: utf-8 -*-
# B1Cron Task: %s
# Created: %s
# Task ID: %d

`, taskName, timestamp, taskID)
	default:
		return userContent
	}

	return header + userContent
}

// generateExecCommand 生成执行命令
func (s *ScriptFileService) generateExecCommand(scriptType, fullPath string) string {
	switch scriptType {
	case "shell":
		return fmt.Sprintf("/bin/bash %s", fullPath)
	case "python":
		// 使用 python3 命令，依赖 shell 的 PATH 环境变量查找
		// 在Docker Alpine环境中更加可靠
		return fmt.Sprintf("python3 %s", fullPath)
	default:
		return fullPath
	}
}

// ValidateScriptContent 验证脚本内容
func (s *ScriptFileService) ValidateScriptContent(scriptType, content string) error {
	if strings.TrimSpace(content) == "" {
		return fmt.Errorf("script content cannot be empty")
	}

	// 检查内容长度限制（防止过大文件）
	if len(content) > 1024*1024 { // 1MB limit
		return fmt.Errorf("script content too large (max 1MB)")
	}

	// 基本安全检查：禁止包含可能的路径遍历
	if strings.Contains(content, "../") {
		return fmt.Errorf("script content contains potentially unsafe path traversal")
	}

	return nil
}

// GetScriptContent 读取脚本文件内容
func (s *ScriptFileService) GetScriptContent(scriptPath string) (string, error) {
	if scriptPath == "" {
		return "", nil
	}

	fullPath := filepath.Join(s.dataDir, scriptPath)
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to read script file: %w", err)
	}

	// 直接返回完整的文件内容，包括头部注释
	return string(content), nil
}