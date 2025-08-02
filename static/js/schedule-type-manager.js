/**
 * 调度类型管理组件 - 处理周期性和一次性任务切换
 */
class ScheduleTypeManager {
    constructor(containerSelector, options = {}) {
        this.container = typeof containerSelector === 'string' ? 
            document.querySelector(containerSelector) : containerSelector;
        this.options = {
            typeButtonSelector: '.schedule-type-btn',
            contentSelector: '.schedule-type-content',
            typeOutputSelector: '#final-schedule-type',
            activeClass: 'text-primary-600 border-b-2 border-primary-600 bg-primary-50',
            inactiveClass: 'text-slate-500 hover:text-slate-700',
            ...options
        };
        this.init();
    }

    init() {
        if (!this.container) return;
        this.bindEvents();
        this.setDefaultType();
    }

    bindEvents() {
        this.container.addEventListener('click', (e) => {
            const typeBtn = e.target.closest(this.options.typeButtonSelector);
            if (typeBtn) {
                this.switchType(typeBtn);
            }
        });
    }

    switchType(activeBtn) {
        const scheduleType = activeBtn.dataset.type;
        
        // 更新按钮状态
        this.container.querySelectorAll(this.options.typeButtonSelector).forEach(btn => {
            btn.className = btn.className
                .replace(this.options.activeClass, '')
                .replace(this.options.inactiveClass, '') + 
                (btn === activeBtn ? ` ${this.options.activeClass}` : ` ${this.options.inactiveClass}`);
        });

        // 切换内容显示
        this.container.querySelectorAll(this.options.contentSelector).forEach(content => {
            content.classList.add('hidden');
        });

        const targetContent = this.container.querySelector('#' + scheduleType + '-schedule');
        if (targetContent) {
            targetContent.classList.remove('hidden');
        }

        // 更新隐藏字段
        const typeOutput = this.container.querySelector(this.options.typeOutputSelector);
        if (typeOutput) {
            typeOutput.value = scheduleType;
        }
    }

    setDefaultType() {
        const defaultBtn = this.container.querySelector(this.options.typeButtonSelector);
        if (defaultBtn) {
            this.switchType(defaultBtn);
        }
    }

    getActiveType() {
        const activeBtn = this.container.querySelector(this.options.typeButtonSelector + '.text-primary-600');
        return activeBtn ? activeBtn.dataset.type : 'cron';
    }

    // 设置指定的调度类型为激活状态
    setActiveType(type) {
        const targetBtn = this.container.querySelector(`[data-type="${type}"]`);
        if (targetBtn) {
            this.switchType(targetBtn);
        }
    }
}

// 导出组件
window.B1Components = window.B1Components || {};
window.B1Components.ScheduleTypeManager = ScheduleTypeManager;