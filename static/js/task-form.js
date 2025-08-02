/**
 * 任务表单组件 - 统一处理任务创建和编辑表单
 */
class TaskForm {
    constructor(formSelector, options = {}) {
        this.form = typeof formSelector === 'string' ? 
            document.querySelector(formSelector) : formSelector;
        this.options = {
            isEdit: false,
            apiEndpoint: '/api/tasks',
            ...options
        };
        
        this.tabManager = null;
        this.scheduleManager = null;
        this.scheduleTypeManager = null;
        this.codeEditor = null;
        
        this.init();
    }

    init() {
        if (!this.form) return;

        this.initComponents();
        this.bindEvents();
    }

    initComponents() {
        const prefix = this.options.isEdit ? 'edit-' : '';
        
        // 初始化调度类型管理
        this.scheduleTypeManager = new B1Components.ScheduleTypeManager(this.form, {
            typeButtonSelector: '.schedule-type-btn',
            contentSelector: '.schedule-type-content',
            typeOutputSelector: `#${prefix}final-schedule-type`
        });
        
        // 初始化标签页管理
        this.tabManager = new B1Components.TabManager(this.form, {
            defaultTab: this.form.querySelector(`[data-tab="${prefix}simple"]`)
        });

        // 初始化调度规则管理
        this.scheduleManager = new B1Components.ScheduleManager(this.form, {
            frequencySelector: `#${prefix}frequency-select`,
            intervalSelector: `#${prefix}interval-value`,
            timeSelector: `#${prefix}execution-time`,
            outputSelector: `#${prefix}final-schedule`,
            manualSelector: `#${prefix}ScheduleSpecManual`
        });

        // 初始化代码编辑器
        const textareaSelector = this.options.isEdit ? '#editTaskCommand' : '#command-input';
        this.codeEditor = new B1Components.CodeEditor(textareaSelector);
    }

    bindEvents() {
        this.form.addEventListener('submit', (e) => this.handleSubmit(e));
        
        // 脚本类型切换
        const scriptTypeSelector = this.options.isEdit ? 
            '#edit-script-type-select' : '#script-type-select';
        const scriptSelect = this.form.querySelector(scriptTypeSelector);
        
        if (scriptSelect) {
            scriptSelect.addEventListener('change', (e) => {
                this.handleScriptTypeChange(e.target.value);
            });
        }
    }

    handleScriptTypeChange(scriptType) {
        if (!this.codeEditor) return;

        let mode = 'shell';
        switch (scriptType) {
            case 'python':
                mode = 'python';
                break;
            case 'shell':
                mode = 'shell';
                break;
            default:
                mode = 'shell';
        }
        
        this.codeEditor.setMode(mode);
    }

    async handleSubmit(e) {
        e.preventDefault();
        
        // 确保编辑器内容同步
        if (this.codeEditor) {
            this.codeEditor.save();
        }

        const formData = new FormData(this.form);
        const scheduleType = this.scheduleTypeManager.getActiveType();
        
        let taskData = {
            name: formData.get('name'),
            command: formData.get('command'),
            script_type: formData.get('script_type'),
            schedule_type: scheduleType,
            schedule_spec: '',
            is_enabled: formData.has('is_enabled')
        };

        if (scheduleType === 'once') {
            // 一次性任务
            const executeAt = formData.get('execute_at');
            if (!executeAt) {
                window.b1cron.showToast('请选择执行时间', 'error');
                return;
            }
            taskData.execute_at = executeAt;
            taskData.schedule_spec = ''; // 一次性任务不需要schedule_spec
        } else {
            // 周期性任务
            const scheduleSpec = this.scheduleManager.getScheduleSpec(this.tabManager);
            taskData.schedule_spec = scheduleSpec;
        }

        try {
            const method = this.options.isEdit ? 'PUT' : 'POST';
            const url = this.options.isEdit ? 
                `${this.options.apiEndpoint}/${formData.get('id')}` : 
                this.options.apiEndpoint;

            const response = await fetch(url, {
                method,
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(taskData)
            });

            if (response.ok) {
                const message = this.options.isEdit ? '任务修改成功' : '任务创建成功';
                window.b1cron.showToast(message, 'success');
                this.onSuccess();
            } else {
                const data = await response.json();
                window.b1cron.showToast(data.error || '操作失败', 'error');
            }
        } catch (error) {
            window.b1cron.showToast('网络错误，请重试', 'error');
        }
    }

    onSuccess() {
        // 关闭模态框，重置表单，刷新页面
        const modalId = this.options.isEdit ? 'editTaskModal' : 'createTaskModal';
        window.b1cron.closeModal(modalId);
        this.form.reset();
        setTimeout(() => location.reload(), 1000);
    }

    loadTaskData(taskData) {
        if (!this.options.isEdit) return;

        // 填充表单数据
        Object.keys(taskData).forEach(key => {
            let input = this.form.querySelector(`[name="${key}"]`);
            
            // 特殊处理编辑表单的命名
            if (!input && key === 'id') {
                input = this.form.querySelector('#editTaskId');
            } else if (!input && key === 'name') {
                input = this.form.querySelector('#editTaskName');
            } else if (!input && key === 'is_enabled') {
                input = this.form.querySelector('#editTaskEnabled');
            }
            
            if (input) {
                if (input.type === 'checkbox') {
                    input.checked = taskData[key];
                } else {
                    input.value = taskData[key];
                }
            }
        });

        // 设置脚本类型
        const scriptTypeSelect = this.form.querySelector('#edit-script-type-select');
        if (scriptTypeSelect && taskData.script_type) {
            scriptTypeSelect.value = taskData.script_type;
            // 触发脚本类型变化
            this.handleScriptTypeChange(taskData.script_type);
        }

        // 设置编辑器内容
        if (this.codeEditor && taskData.command) {
            this.codeEditor.setValue(taskData.command);
            // 确保编辑器内容正确显示
            this.codeEditor.refresh();
        }

        // 设置调度类型和相关字段
        if (taskData.schedule_type) {
            // 设置调度类型
            if (this.scheduleTypeManager) {
                this.scheduleTypeManager.setActiveType(taskData.schedule_type);
            }

            // 如果是一次性任务，设置执行时间
            if (taskData.schedule_type === 'once' && taskData.execute_at) {
                const executeAtInput = this.form.querySelector('#editExecuteAt');
                if (executeAtInput) {
                    // 将ISO时间转换为datetime-local格式
                    const date = new Date(taskData.execute_at);
                    const year = date.getFullYear();
                    const month = String(date.getMonth() + 1).padStart(2, '0');
                    const day = String(date.getDate()).padStart(2, '0');
                    const hours = String(date.getHours()).padStart(2, '0');
                    const minutes = String(date.getMinutes()).padStart(2, '0');
                    executeAtInput.value = `${year}-${month}-${day}T${hours}:${minutes}`;
                }
            }
        }

        // 设置调度规则（对于周期性任务）
        if (taskData.schedule_spec && (!taskData.schedule_type || taskData.schedule_type === 'cron')) {
            // 设置高级模式的cron表达式
            const manualInput = this.form.querySelector('#editScheduleSpecManual');
            if (manualInput) {
                manualInput.value = taskData.schedule_spec;
            }
        }
    }
}

// 导出组件
window.B1Components = window.B1Components || {};
window.B1Components.TaskForm = TaskForm;