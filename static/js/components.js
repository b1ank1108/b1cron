/**
 * B1Cron 组件库 - 统一入口文件
 * 
 * 这个文件作为组件库的主入口点，负责：
 * 1. 初始化全局B1Components命名空间
 * 2. 提供组件注册和管理功能
 * 3. 确保组件加载顺序
 * 
 * 使用方式：
 * - 在HTML中依次加载各个组件文件
 * - 然后加载此文件
 * - 通过 B1Components.ComponentName 访问组件
 */

// 初始化全局命名空间
window.B1Components = window.B1Components || {};

// 组件加载状态跟踪
window.B1Components._loadedComponents = new Set();
window.B1Components._componentDependencies = {
    'TaskForm': ['TabManager', 'ScheduleManager', 'ScheduleTypeManager', 'CodeEditor']
};

/**
 * 注册组件
 * @param {string} name - 组件名称
 * @param {Function} component - 组件构造函数
 */
window.B1Components.register = function(name, component) {
    this[name] = component;
    this._loadedComponents.add(name);
    console.log(`Component ${name} registered`);
};

/**
 * 检查组件依赖是否满足
 * @param {string} componentName - 组件名称
 * @returns {boolean} - 依赖是否满足
 */
window.B1Components.checkDependencies = function(componentName) {
    const deps = this._componentDependencies[componentName] || [];
    return deps.every(dep => this._loadedComponents.has(dep));
};

/**
 * 获取已加载的组件列表
 * @returns {Array} - 已加载的组件名称数组
 */
window.B1Components.getLoadedComponents = function() {
    return Array.from(this._loadedComponents);
};

/**
 * 检查所有组件是否已加载
 * @returns {boolean} - 是否所有必要组件都已加载
 */
window.B1Components.isReady = function() {
    const requiredComponents = ['TabManager', 'ScheduleManager', 'ScheduleTypeManager', 'CodeEditor', 'TaskForm'];
    return requiredComponents.every(comp => this._loadedComponents.has(comp));
};

// 组件就绪事件
window.B1Components.onReady = function(callback) {
    if (this.isReady()) {
        callback();
    } else {
        document.addEventListener('B1ComponentsReady', callback);
    }
};

// 检查并触发就绪事件
function checkAndFireReady() {
    if (window.B1Components.isReady()) {
        document.dispatchEvent(new CustomEvent('B1ComponentsReady'));
    }
}

// 定期检查组件加载状态
const checkInterval = setInterval(() => {
    if (window.B1Components.isReady()) {
        clearInterval(checkInterval);
        checkAndFireReady();
    }
}, 100);

// 兼容性：如果组件已经通过其他方式加载，自动注册
setTimeout(() => {
    ['TabManager', 'ScheduleManager', 'ScheduleTypeManager', 'CodeEditor', 'TaskForm'].forEach(name => {
        if (window.B1Components[name] && !window.B1Components._loadedComponents.has(name)) {
            window.B1Components._loadedComponents.add(name);
        }
    });
    checkAndFireReady();
}, 200);