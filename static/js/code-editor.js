/**
 * 代码编辑器组件 - 基于CodeMirror的简化封装
 */
class CodeEditor {
    constructor(textareaSelector, options = {}) {
        this.textarea = typeof textareaSelector === 'string' ? 
            document.querySelector(textareaSelector) : textareaSelector;
        this.options = {
            lineNumbers: false,
            mode: 'shell',
            theme: 'default',
            lineWrapping: true,
            ...options
        };
        this.editor = null;
        this.init();
    }

    init() {
        if (!this.textarea || typeof CodeMirror === 'undefined') return;
        
        try {
            this.editor = CodeMirror.fromTextArea(this.textarea, this.options);
            this.bindEvents();
        } catch (error) {
            console.warn('CodeMirror initialization failed:', error);
        }
    }

    bindEvents() {
        if (!this.editor) return;
        
        // 监听内容变化
        this.editor.on('change', () => {
            this.save();
        });
    }

    setValue(value) {
        if (this.editor) {
            this.editor.setValue(value || '');
            // 确保CodeMirror正确刷新显示内容
            setTimeout(() => {
                this.editor.refresh();
            }, 10);
        } else if (this.textarea) {
            this.textarea.value = value || '';
        }
    }

    getValue() {
        if (this.editor) {
            return this.editor.getValue();
        } else if (this.textarea) {
            return this.textarea.value;
        }
        return '';
    }

    save() {
        if (this.editor && this.textarea) {
            this.textarea.value = this.editor.getValue();
        }
    }

    setMode(mode) {
        if (this.editor) {
            this.editor.setOption('mode', mode);
        }
    }

    refresh() {
        if (this.editor) {
            // 使用更长的延迟确保DOM完全渲染
            setTimeout(() => {
                this.editor.refresh();
                // 如果编辑器在隐藏容器中，可能需要多次刷新
                setTimeout(() => {
                    this.editor.refresh();
                }, 50);
            }, 100);
        }
    }

    focus() {
        if (this.editor) {
            this.editor.focus();
        } else if (this.textarea) {
            this.textarea.focus();
        }
    }
}

// 导出组件
window.B1Components = window.B1Components || {};
window.B1Components.CodeEditor = CodeEditor;