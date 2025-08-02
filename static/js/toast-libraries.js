/**
 * 第三方Toast库集成方案
 * 推荐使用 Notyf - 轻量级、现代化的通知库
 */

// 方案1: Notyf (推荐) - 2.3KB gzipped
// CDN: <script src="https://cdn.jsdelivr.net/npm/notyf@3/notyf.min.js"></script>
// CDN: <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/notyf@3/notyf.min.css">

class NotyfToastWrapper {
    constructor() {
        if (typeof Notyf !== 'undefined') {
            this.notyf = new Notyf({
                duration: 4000,
                position: { x: 'right', y: 'top' },
                types: [
                    {
                        type: 'warning',
                        background: 'orange',
                        icon: {
                            className: 'material-icons',
                            tagName: 'i',
                            text: 'warning'
                        }
                    },
                    {
                        type: 'info', 
                        background: 'blue',
                        icon: {
                            className: 'material-icons',
                            tagName: 'i', 
                            text: 'info'
                        }
                    }
                ]
            });
        } else {
            console.warn('Notyf library not loaded, falling back to browser alerts');
        }
    }

    show(message, type = 'info', duration = 4000) {
        if (this.notyf) {
            return this.notyf.open({
                type,
                message,
                duration
            });
        } else {
            // 降级处理
            alert(`${type.toUpperCase()}: ${message}`);
        }
    }

    success(message, duration = 4000) {
        return this.notyf ? this.notyf.success(message) : this.show(message, 'success', duration);
    }

    error(message, duration = 4000) {
        return this.notyf ? this.notyf.error(message) : this.show(message, 'error', duration);
    }

    warning(message, duration = 4000) {
        return this.show(message, 'warning', duration);
    }

    info(message, duration = 4000) {
        return this.show(message, 'info', duration);
    }
}

// 方案2: SweetAlert2 - 功能丰富但体积较大 (40KB+)
// CDN: <script src="https://cdn.jsdelivr.net/npm/sweetalert2@11"></script>

class SweetAlertToastWrapper {
    constructor() {
        if (typeof Swal !== 'undefined') {
            this.Toast = Swal.mixin({
                toast: true,
                position: 'top-end',
                showConfirmButton: false,
                timer: 3000,
                timerProgressBar: true,
                didOpen: (toast) => {
                    toast.addEventListener('mouseenter', Swal.stopTimer);
                    toast.addEventListener('mouseleave', Swal.resumeTimer);
                }
            });
        }
    }

    show(message, type = 'info', duration = 3000) {
        if (this.Toast) {
            return this.Toast.fire({
                icon: type,
                title: message,
                timer: duration
            });
        }
    }

    success(message, duration) {
        return this.show(message, 'success', duration);
    }

    error(message, duration) {
        return this.show(message, 'error', duration);
    }

    warning(message, duration) {
        return this.show(message, 'warning', duration);
    }

    info(message, duration) {
        return this.show(message, 'info', duration);
    }
}

// 方案3: 基于CSS Transition的简化版本 (最轻量)
class SimpleToast {
    constructor() {
        this.container = this.createContainer();
        this.addStyles();
    }

    createContainer() {
        const container = document.createElement('div');
        container.id = 'simple-toast-container';
        container.className = 'fixed top-4 right-4 z-50 flex flex-col gap-2 pointer-events-none';
        document.body.appendChild(container);
        return container;
    }

    addStyles() {
        if (document.getElementById('simple-toast-styles')) return;
        
        const style = document.createElement('style');
        style.id = 'simple-toast-styles';
        style.textContent = `
            .simple-toast {
                pointer-events: auto;
                transform: translateX(100%);
                opacity: 0;
                transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
            }
            
            .simple-toast.show {
                transform: translateX(0);
                opacity: 1;
            }
            
            .simple-toast.hide {
                transform: translateX(100%);
                opacity: 0;
            }
        `;
        document.head.appendChild(style);
    }

    show(message, type = 'info', duration = 4000) {
        const colors = {
            success: 'bg-green-500 text-white',
            error: 'bg-red-500 text-white', 
            warning: 'bg-yellow-500 text-black',
            info: 'bg-blue-500 text-white'
        };

        const icons = {
            success: '✅',
            error: '❌',
            warning: '⚠️', 
            info: 'ℹ️'
        };

        const toast = document.createElement('div');
        toast.className = `simple-toast flex items-center gap-3 px-4 py-3 rounded-lg shadow-lg ${colors[type]}`;
        toast.innerHTML = `
            <span>${icons[type]}</span>
            <span class="flex-1">${message}</span>
            <button class="hover:opacity-70" onclick="this.parentElement.remove()">×</button>
        `;

        this.container.appendChild(toast);

        // 触发动画
        requestAnimationFrame(() => {
            toast.classList.add('show');
        });

        // 自动移除
        if (duration > 0) {
            setTimeout(() => {
                toast.classList.add('hide');
                setTimeout(() => toast.remove(), 300);
            }, duration);
        }

        return toast;
    }

    success(message, duration) { return this.show(message, 'success', duration); }
    error(message, duration) { return this.show(message, 'error', duration); }
    warning(message, duration) { return this.show(message, 'warning', duration); }
    info(message, duration) { return this.show(message, 'info', duration); }
}

// 导出配置
window.ToastLibraries = {
    NotyfToastWrapper,
    SweetAlertToastWrapper,
    SimpleToast
};

// 使用建议：根据项目需求选择
/*
使用指南：

1. 轻量级项目 (< 5KB):
   const toast = new SimpleToast();

2. 中等项目，需要更多功能 (< 10KB):
   引入 Notyf CDN，然后：
   const toast = new NotyfToastWrapper();

3. 大型项目，需要复杂交互 (< 50KB):
   引入 SweetAlert2 CDN，然后：
   const toast = new SweetAlertToastWrapper();

4. 现代浏览器 (支持 Web Components):
   使用 toast-modern.js

5. 完全自定义控制:
   使用 toast-improved.js
*/