/**
 * 调度规则管理组件 - 统一处理Cron表达式生成和解析
 */
class ScheduleManager {
    constructor(containerSelector, options = {}) {
        this.container = typeof containerSelector === 'string' ? 
            document.querySelector(containerSelector) : containerSelector;
        this.options = {
            frequencySelector: '#frequency-select',
            intervalSelector: '#interval-value',
            timeSelector: '#execution-time',
            outputSelector: '#final-schedule',
            manualSelector: 'input[name="schedule_spec_manual"]',
            previewSelector: '#cron-preview',
            ...options
        };
        this.init();
    }

    init() {
        if (!this.container) return;
        this.bindEvents();
        this.updateSchedule();
    }

    bindEvents() {
        const selectors = [
            this.options.frequencySelector,
            this.options.intervalSelector,
            this.options.timeSelector
        ];

        selectors.forEach(selector => {
            const element = this.container.querySelector(selector);
            if (element) {
                element.addEventListener('change', () => this.updateSchedule());
                element.addEventListener('input', () => this.updateSchedule());
            }
        });
    }

    updateSchedule() {
        const frequency = this.getValue(this.options.frequencySelector);
        const interval = this.getValue(this.options.intervalSelector);
        const time = this.getValue(this.options.timeSelector);
        const output = this.container.querySelector(this.options.outputSelector);

        if (!frequency || !output) return;

        let spec = '';
        switch (frequency) {
            case 'minutes':
                spec = `@every ${interval}m`;
                break;
            case 'hours':
                spec = `@every ${interval}h`;
                break;
            case 'daily':
                if (time) {
                    const [hour, minute] = time.split(':');
                    spec = `${minute} ${hour} * * *`;
                }
                break;
            case 'weekly':
                if (time) {
                    const [h, m] = time.split(':');
                    spec = `${m} ${h} * * 1`;
                }
                break;
        }

        output.value = spec;
    }

    getValue(selector) {
        const element = this.container.querySelector(selector);
        return element ? element.value : '';
    }

    getScheduleSpec(tabManager) {
        if (!tabManager) return '';
        
        const activeTab = tabManager.getActiveTab();
        if (activeTab === 'advanced' || activeTab?.includes('advanced')) {
            // 高级模式：直接返回手动输入的表达式
            const manualInput = this.container.querySelector(this.options.manualSelector);
            return manualInput ? manualInput.value : '';
        } else {
            // 简单模式：返回自动生成的表达式
            const output = this.container.querySelector(this.options.outputSelector);
            return output ? output.value : '';
        }
    }
}

// 导出组件
window.B1Components = window.B1Components || {};
window.B1Components.ScheduleManager = ScheduleManager;