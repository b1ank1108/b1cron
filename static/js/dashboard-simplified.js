// 简化的Dashboard JavaScript - 使用组件化架构

// 全局变量
let createTaskForm = null;
let editTaskForm = null;

// 页面初始化
document.addEventListener('DOMContentLoaded', function() {
    initializeForms();
    bindGlobalEvents();
});

/**
 * 初始化表单组件
 */
function initializeForms() {
    // 初始化创建任务表单
    createTaskForm = new B1Components.TaskForm('#createTaskForm', {
        isEdit: false,
        apiEndpoint: '/api/tasks'
    });

    // 初始化编辑任务表单  
    editTaskForm = new B1Components.TaskForm('#editTaskForm', {
        isEdit: true,
        apiEndpoint: '/api/tasks'
    });
}

/**
 * 绑定全局事件
 */
function bindGlobalEvents() {
    // 其他不需要组件化的功能
    bindPasswordChange();
    bindExecutionDetails();
}

/**
 * 修改密码功能
 */
function bindPasswordChange() {
    const form = document.getElementById('changePasswordForm');
    if (form) {
        form.addEventListener('submit', handlePasswordChange);
    }
}

async function handlePasswordChange(e) {
    e.preventDefault();
    
    const form = e.target;
    const formData = new FormData(form);
    
    const currentPassword = formData.get('current_password');
    const newPassword = formData.get('new_password');
    const confirmPassword = formData.get('confirm_password');
    
    if (newPassword !== confirmPassword) {
        window.b1cron.showToast('两次输入的密码不一致', 'error');
        return;
    }
    
    try {
        const response = await fetch('/api/change-password', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                current_password: currentPassword,
                new_password: newPassword
            })
        });
        
        const data = await response.json();
        
        if (response.ok) {
            window.b1cron.showToast('密码修改成功', 'success');
            window.b1cron.closeModal('changePasswordModal');
            form.reset();
        } else {
            window.b1cron.showToast(data.message || '密码修改失败', 'error');
        }
    } catch (error) {
        window.b1cron.showToast('网络错误，请重试', 'error');
    }
}

/**
 * 执行详情功能
 */
function bindExecutionDetails() {
    // 这部分功能相对独立，暂时保持原有实现
}


// 显示执行详情模态框 - 从元素读取数据
function showExecutionDetailFromElement(element) {
    try {
        const executionData = JSON.parse(element.dataset.execution || '{}');
        const taskName = executionData.taskName || '';
        const status = executionData.status || '';
        const startTime = executionData.startTime || '';
        const endTime = executionData.endTime || '';
        const duration = parseInt(executionData.duration) || 0;
        const output = executionData.output || '';
        const errorMsg = executionData.error || '';
        
        showExecutionDetail(taskName, status, startTime, endTime, duration, output, errorMsg);
    } catch (error) {
        console.error('Error parsing execution data:', error);
        // 回退到旧的数据格式（如果存在）
        const taskName = element.dataset.taskName || '';
        const status = element.dataset.status || '';
        const startTime = element.dataset.startTime || '';
        const endTime = element.dataset.endTime || '';
        const duration = parseInt(element.dataset.duration) || 0;
        const output = element.dataset.output || '';
        const errorMsg = element.dataset.error || '';
        
        showExecutionDetail(taskName, status, startTime, endTime, duration, output, errorMsg);
    }
}

// 显示执行详情模态框
function showExecutionDetail(taskName, status, startTime, endTime, duration, output, errorMsg) {
    // 设置基本信息
    document.getElementById('detailTaskName').textContent = taskName;
    
    // 设置状态徽章
    const statusElement = document.getElementById('detailStatus');
    let statusHtml = '';
    switch(status) {
        case 'success':
            statusHtml = '<span class="badge badge-success">成功</span>';
            break;
        case 'failed':
            statusHtml = '<span class="badge badge-danger">失败</span>';
            break;
        case 'running':
            statusHtml = '<span class="badge badge-warning">运行中</span>';
            break;
        default:
            statusHtml = `<span class="badge badge-secondary">${status}</span>`;
    }
    statusElement.innerHTML = statusHtml;
    
    // 设置时间信息
    document.getElementById('detailStartTime').textContent = startTime;
    document.getElementById('detailEndTime').textContent = endTime || '--';
    
    // 设置执行时长
    const durationElement = document.getElementById('detailDuration');
    if (duration && duration > 0) {
        if (duration < 1000) {
            durationElement.textContent = duration + 'ms';
        } else {
            durationElement.textContent = (duration / 1000).toFixed(2) + 's';
        }
    } else {
        durationElement.textContent = '--';
    }
    
    // 处理输出和错误信息
    const outputSection = document.getElementById('outputSection');
    const errorSection = document.getElementById('errorSection');
    const outputElement = document.getElementById('detailOutput');
    const errorElement = document.getElementById('detailError');
    
    if (output && output.trim()) {
        outputSection.style.display = 'block';
        outputElement.textContent = output;
    } else {
        outputSection.style.display = 'none';
    }
    
    if (errorMsg && errorMsg.trim()) {
        errorSection.style.display = 'block';
        errorElement.textContent = errorMsg;
    } else {
        errorSection.style.display = 'none';
    }
    
    // 显示模态框
    window.b1cron.showModal('executionDetailModal');
}

/**
 * 编辑任务功能 - 从元素读取数据 
 */
function editTaskFromElement(element) {
    const id = element.dataset.taskId;
    // 通过API获取完整的任务信息
    fetchTaskAndEdit(id);
}

/**
 * 通过API获取任务信息并编辑
 */
async function fetchTaskAndEdit(taskId) {
    try {
        const response = await fetch(`/api/tasks/${taskId}`);
        if (response.ok) {
            const task = await response.json();
            
            // 使用组件加载任务数据
            if (editTaskForm) {
                editTaskForm.loadTaskData(task);
                window.b1cron.showModal('editTaskModal');
                
                // 模态框显示后刷新编辑器确保内容正确显示
                setTimeout(() => {
                    if (editTaskForm.codeEditor) {
                        editTaskForm.codeEditor.refresh();
                    }
                }, 200);
            }
        } else {
            window.b1cron.showToast('获取任务信息失败', 'error');
        }
    } catch (error) {
        window.b1cron.showToast('网络错误，请重试', 'error');
    }
}

// // 为了向后兼容，保留一些全局函数引用
// window.editTaskFromElement = editTaskFromElement;
// window.showExecutionDetailFromElement = showExecutionDetailFromElement;

// 将函数添加到全局作用域，供HTML调用
window.editTaskFromElement = editTaskFromElement;
window.showExecutionDetailFromElement = showExecutionDetailFromElement;