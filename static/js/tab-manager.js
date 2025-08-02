/**
 * 标签页管理组件 - 处理选项卡切换和内容显示
 */
class TabManager {
    constructor(containerSelector, options = {}) {
        // 如果传入的是DOM元素，直接使用；如果是字符串，查询
        this.container = typeof containerSelector === 'string' ? 
            document.querySelector(containerSelector) : containerSelector;
        this.options = {
            tabSelector: '.tab-btn',
            contentSelector: '.tab-content',
            activeClass: 'active',
            defaultTab: null,
            ...options
        };
        this.init();
    }

    init() {
        if (!this.container) return;
        
        this.bindEvents();
        this.setDefaultTab();
    }

    bindEvents() {
        this.container.addEventListener('click', (e) => {
            const tabBtn = e.target.closest(this.options.tabSelector);
            if (tabBtn) {
                this.switchTab(tabBtn);
            }
        });
    }

    switchTab(activeBtn) {
        // 定义选项卡的CSS类
        const activeClasses = 'text-primary-600 border-b-2 border-primary-600 bg-primary-50';
        const inactiveClasses = 'text-slate-500 hover:text-slate-700';
        
        // 移除所有活跃状态并设置为非活跃状态
        this.container.querySelectorAll(this.options.tabSelector).forEach(btn => {
            btn.className = btn.className
                .replace(/text-primary-600|border-b-2|border-primary-600|bg-primary-50/g, '')
                .replace(/text-slate-500|hover:text-slate-700/g, '')
                .trim();
            btn.className += ` ${inactiveClasses}`;
        });
        
        this.container.querySelectorAll(this.options.contentSelector)
            .forEach(content => content.style.display = 'none');

        // 激活当前标签
        activeBtn.className = activeBtn.className
            .replace(/text-slate-500|hover:text-slate-700/g, '')
            .trim();
        activeBtn.className += ` ${activeClasses}`;
        
        const tabId = activeBtn.dataset.tab;
        const content = document.getElementById(tabId + '-tab');
        if (content) {
            content.style.display = 'block';
        }
    }

    setDefaultTab() {
        const defaultTab = this.options.defaultTab || 
            this.container.querySelector(this.options.tabSelector);
        
        if (defaultTab && !this.container.querySelector('.text-primary-600')) {
            this.switchTab(defaultTab);
        }
    }

    getActiveTab() {
        const activeBtn = this.container.querySelector('.text-primary-600');
        return activeBtn ? activeBtn.dataset.tab : null;
    }
}

// 导出组件
window.B1Components = window.B1Components || {};
window.B1Components.TabManager = TabManager;