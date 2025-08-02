// B1Cron 现代化交互功能 - 简化版（无暗黑模式）
class B1Cron {
    constructor() {
        this.init();
    }

    init() {
        this.initModal();
        this.initFormValidation();
        this.initScriptTypeSelector();
        this.bindEvents();
        
        // 页面加载动画
        document.body.classList.add('animate-fade-in');
    }

    // 模态框管理
    initModal() {
        document.addEventListener('click', (e) => {
            if (e.target.matches('.modal-overlay')) {
                this.closeModal(e.target);
            }
            if (e.target.matches('.modal-close')) {
                const modal = e.target.closest('.modal-overlay');
                this.closeModal(modal);
            }
        });
    }

    showModal(modalId) {
        const modal = document.getElementById(modalId);
        if (modal) {
            modal.style.display = 'flex';
            modal.classList.add('show');
            document.body.style.overflow = 'hidden';
        }
    }

    closeModal(modal) {
        if (typeof modal === 'string') {
            modal = document.getElementById(modal);
        }
        if (modal) {
            modal.classList.remove('show');
            setTimeout(() => {
                modal.style.display = 'none';
            }, 250);
            document.body.style.overflow = '';
        }
    }

    // 表单验证
    initFormValidation() {
        document.addEventListener('submit', (e) => {
            const form = e.target;
            if (form.classList.contains('validate')) {
                if (!this.validateForm(form)) {
                    e.preventDefault();
                }
            }
        });
    }

    validateForm(form) {
        const inputs = form.querySelectorAll('input[required], select[required], textarea[required]');
        let isValid = true;

        inputs.forEach(input => {
            const isFieldValid = this.validateField(input);
            if (!isFieldValid) isValid = false;
        });

        return isValid;
    }

    validateField(field) {
        const value = field.value.trim();
        const type = field.type;
        let isValid = true;
        let message = '';

        // 必填验证
        if (field.hasAttribute('required') && !value) {
            isValid = false;
            message = '此字段为必填项';
        }

        // 类型验证
        if (value && type === 'email' && !this.isValidEmail(value)) {
            isValid = false;
            message = '请输入有效的邮箱地址';
        }

        if (value && type === 'password' && value.length < 6) {
            isValid = false;
            message = '密码长度至少6位';
        }

        // Cron表达式验证
        if (field.name === 'schedule_spec' && value && !this.isValidCron(value)) {
            isValid = false;
            message = '请输入有效的Cron表达式或间隔格式';
        }

        this.showFieldValidation(field, isValid, message);
        return isValid;
    }

    showFieldValidation(field, isValid, message) {
        // 移除之前的验证样式
        field.classList.remove('is-valid', 'is-invalid');
        const existingFeedback = field.parentNode.querySelector('.field-feedback');
        if (existingFeedback) {
            existingFeedback.remove();
        }

        if (!isValid) {
            field.classList.add('is-invalid');
            const feedback = document.createElement('div');
            feedback.className = 'field-feedback text-error';
            feedback.textContent = message;
            field.parentNode.appendChild(feedback);
        } else if (field.value.trim()) {
            field.classList.add('is-valid');
        }
    }

    // 工具函数
    isValidEmail(email) {
        const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
        return emailRegex.test(email);
    }

    isValidCron(cron) {
        // 简单的Cron验证
        if (cron.startsWith('@every ')) {
            const duration = cron.substring(7);
            return /^\d+[smhd]$/.test(duration);
        }
        
        // 标准Cron表达式验证 (5段)
        const parts = cron.split(' ');
        return parts.length === 5;
    }

    // Cron表达式预览
    getCronPreview(cronExpression) {
        try {
            if (cronExpression.startsWith('@every ')) {
                const duration = cronExpression.substring(7);
                return `每${this.formatDuration(duration)}执行一次`;
            }
            
            // 简单的Cron描述
            return this.describeCron(cronExpression);
        } catch (error) {
            return '无效的表达式';
        }
    }

    formatDuration(duration) {
        const number = parseInt(duration);
        const unit = duration.slice(-1);
        const units = { s: '秒', m: '分钟', h: '小时', d: '天' };
        return `${number}${units[unit] || ''}`;
    }

    describeCron(cron) {
        const parts = cron.split(' ');
        if (parts.length !== 5) return '格式错误，需要5个字段';
        
        const [minute, hour, day, month, dayOfWeek] = parts;
        
        // 简单的描述生成
        if (minute === '*' && hour === '*') return '每分钟';
        if (hour === '*') return `每小时的第${minute}分钟`;
        if (day === '*' && month === '*' && dayOfWeek === '*') {
            return `每天${hour}:${minute.padStart(2, '0')}`;
        }
        if (day === '*' && month === '*' && dayOfWeek !== '*') {
            const dayNames = ['周日', '周一', '周二', '周三', '周四', '周五', '周六'];
            return `每${dayNames[dayOfWeek]}的${hour}:${minute.padStart(2, '0')}`;
        }
        
        return '自定义时间表';
    }

    // 事件绑定
    bindEvents() {
        // 实时表单验证
        document.addEventListener('blur', (e) => {
            if (e.target.matches('input, select, textarea')) {
                this.validateField(e.target);
            }
        }, true);

        // Cron表达式预览
        document.addEventListener('input', (e) => {
            if (e.target.name === 'schedule_spec') {
                this.updateCronPreview(e.target);
            }
        });

        // HTMX增强
        this.enhanceHTMX();
    }

    updateCronPreview(input) {
        const preview = document.querySelector('#cron-preview');
        if (preview && input.value.trim()) {
            const description = this.getCronPreview(input.value.trim());
            preview.textContent = description;
            preview.style.display = 'block';
        } else if (preview) {
            preview.style.display = 'none';
        }
    }

    // HTMX集成增强
    enhanceHTMX() {
        // HTMX请求前显示加载状态
        document.addEventListener('htmx:beforeRequest', (e) => {
            const button = e.detail.elt;
            if (button.tagName === 'BUTTON') {
                button.classList.add('loading');
                button.disabled = true;
            }
        });

        // HTMX请求完成后恢复状态
        document.addEventListener('htmx:afterRequest', (e) => {
            const button = e.detail.elt;
            if (button.tagName === 'BUTTON') {
                button.classList.remove('loading');
                button.disabled = false;
            }

            // 根据响应状态显示通知
            const xhr = e.detail.xhr;
            if (xhr.status >= 200 && xhr.status < 300) {
                window.toast.success('操作成功');
            } else if (xhr.status >= 400) {
                window.toast.error('操作失败');
            }
        });

        // HTMX错误处理
        document.addEventListener('htmx:responseError', (e) => {
            window.toast.error('网络错误，请重试');
        });
    }

    // 脚本类型切换功能
    initScriptTypeSelector() {
        // 创建任务模态框的脚本类型选择器
        const createScriptSelect = document.getElementById('script-type-select');
        if (createScriptSelect) {
            createScriptSelect.addEventListener('change', (e) => {
                this.handleScriptTypeChange(e.target.value, 'create');
            });
            // 初始化时就显示默认选项的模板
            this.handleScriptTypeChange(createScriptSelect.value, 'create');
        }

        // 编辑任务模态框的脚本类型选择器
        const editScriptSelect = document.getElementById('edit-script-type-select');
        if (editScriptSelect) {
            editScriptSelect.addEventListener('change', (e) => {
                this.handleScriptTypeChange(e.target.value, 'edit');
            });
        }
    }

    handleScriptTypeChange(scriptType, modalType) {
        const prefix = modalType === 'edit' ? 'edit-' : '';
        const commandLabel = document.getElementById(`${prefix}command-label`);
        const commandHelp = document.getElementById(`${prefix}command-help`);

        // 获取对应的元素
        const actualCommandInput = modalType === 'edit' ? 
            document.getElementById('editTaskCommand') : 
            document.querySelector('#createTaskForm textarea[name="command"]');

        if (!commandLabel || !actualCommandInput || !commandHelp) return;

        switch(scriptType) {
            case 'shell':
                commandLabel.textContent = 'Shell脚本内容 *';
                commandHelp.textContent = '脚本将保存为.sh文件并通过bash执行';
                // 如果当前内容为空或者是其他类型的模板，则填入Shell模板
                if (!actualCommandInput.value.trim() || this.isTemplateContent(actualCommandInput.value)) {
                    actualCommandInput.value = `#!/bin/bash

# 在这里编写您的Shell脚本
`;
                }
                break;
            
            case 'python':
                commandLabel.textContent = 'Python脚本内容 *';
                commandHelp.textContent = '脚本将保存为.py文件并通过python3执行';
                // 如果当前内容为空或者是其他类型的模板，则填入Python模板
                if (!actualCommandInput.value.trim() || this.isTemplateContent(actualCommandInput.value)) {
                    actualCommandInput.value = `#!/usr/bin/env python3
# -*- coding: utf-8 -*-
# 在这里编写您的Python脚本
`;
                }
                break;
            
            default: // command
                commandLabel.textContent = '执行命令 *';
                commandHelp.textContent = '支持任何 shell 命令，如脚本路径、系统命令等';
                // 如果当前内容是脚本模板，则清除
                if (this.isTemplateContent(actualCommandInput.value)) {
                    actualCommandInput.value = '';
                }
                actualCommandInput.placeholder = "例如: echo 'Hello World' 或 /path/to/script.sh";
                break;
        }
    }

    // 检查是否为模板内容
    isTemplateContent(content) {
        if (!content) return false;
        
        // 检查是否包含我们的模板标识内容
        return content.includes('# 在这里编写您的Shell脚本') || 
               content.includes('# 在这里编写您的Python脚本') ||
               (content.includes('#!/bin/bash') && content.trim().split('\n').length <= 4) ||
               (content.includes('#!/usr/bin/env python3') && content.trim().split('\n').length <= 4);
    }

    // Toast 便捷方法（使用新的 SimpleToast）
    showToast(message, type = 'info', duration = 3000) {
        return window.toast.show(message, type, duration);
    }
}

// 页面加载完成后初始化
document.addEventListener('DOMContentLoaded', () => {
    window.b1cron = new B1Cron();
});

// 导出给全局使用
window.B1Cron = B1Cron;