/**
 * B1Cron Execution Records Management
 * 
 * 处理执行记录的搜索和分页功能
 */

// 全局变量
let executionCurrentPage = 1;
let executionPageSize = 20;
let executionSearchTerm = '';
let executionTotalPages = 1;
let executionTotalItems = 0;

// 初始化执行记录功能
document.addEventListener('DOMContentLoaded', function() {
    // 初始加载执行记录
    loadExecutions();
    
    // 搜索输入框防抖处理
    let searchTimeout;
    const searchInput = document.getElementById('executionSearchInput');
    if (searchInput) {
        searchInput.addEventListener('input', function() {
            clearTimeout(searchTimeout);
            searchTimeout = setTimeout(() => {
                executionSearchTerm = this.value.trim();
                executionCurrentPage = 1; // 重置到第一页
                loadExecutions();
            }, 300); // 300ms防抖
        });
    }
});

// 加载执行记录
async function loadExecutions() {
    const pageSize = document.getElementById('executionPageSize');
    if (pageSize) {
        executionPageSize = parseInt(pageSize.value);
    }
    
    try {
        const params = new URLSearchParams({
            page: executionCurrentPage,
            page_size: executionPageSize
        });
        
        if (executionSearchTerm) {
            params.append('search', executionSearchTerm);
        }
        
        const response = await fetch(`/api/executions/recent?${params}`);
        if (!response.ok) {
            throw new Error('Failed to load executions');
        }
        
        const data = await response.json();
        
        // 检查是否是新的分页API格式
        if (data.executions && data.pagination) {
            renderExecutions(data.executions);
            updatePagination(data.pagination);
        } else {
            // 处理旧格式数据（向后兼容）
            renderExecutions(data);
            updatePagination({
                current_page: 1,
                page_size: executionPageSize,
                total_items: data.length,
                total_pages: 1,
                has_next: false,
                has_prev: false
            });
        }
    } catch (error) {
        console.error('Error loading executions:', error);
        SimpleToast.error('加载执行记录失败');
    }
}

// 渲染执行记录表格
function renderExecutions(executions) {
    const tbody = document.getElementById('executionsTableBody');
    if (!tbody) return;
    
    if (executions.length === 0) {
        tbody.innerHTML = `
            <tr>
                <td colspan="5" class="px-6 py-12 text-center">
                    <div class="text-center">
                        <div class="text-4xl mb-4 opacity-50">📊</div>
                        <p class="text-slate-500">${executionSearchTerm ? '没有找到匹配的执行记录' : '暂无执行记录'}</p>
                    </div>
                </td>
            </tr>
        `;
        return;
    }
    
    tbody.innerHTML = executions.map(execution => {
        const statusClass = getStatusClass(execution.status);
        const statusText = getStatusText(execution.status);
        const durationText = execution.duration ? 
            (execution.duration < 1000 ? `${execution.duration}ms` : `${(execution.duration / 1000).toFixed(2)}s`) : 
            '--';
        
        const outputHtml = execution.output ? 
            `<div class="bg-slate-100 rounded-md p-2 text-xs font-mono text-slate-700 max-h-10 overflow-y-auto hover:max-h-16 transition-all duration-200" title="${escapeHtml(execution.output)}">${escapeHtml(execution.output)}</div>` :
            (execution.error_msg ? 
                `<div class="bg-red-100 rounded-md p-2 text-xs font-mono text-red-700 max-h-10 overflow-y-auto hover:max-h-16 transition-all duration-200" title="${escapeHtml(execution.error_msg)}">${escapeHtml(execution.error_msg)}</div>` :
                '<span class="text-slate-400">--</span>');
        
        return `
            <tr class="hover:bg-slate-50 transition-colors duration-150 cursor-pointer" 
                data-execution-id="${execution.id}" 
                onclick="showExecutionDetailFromElement(this)"
                data-execution='${JSON.stringify({
                    id: execution.id,
                    taskName: execution.task.name,
                    status: execution.status,
                    startTime: formatDateTime(execution.started_at),
                    endTime: execution.completed_at ? formatDateTime(execution.completed_at) : '',
                    duration: execution.duration || 0,
                    output: execution.output || '',
                    error: execution.error_msg || ''
                })}'>
                <td class="px-6 py-4 whitespace-nowrap">
                    <div class="text-sm font-medium text-slate-900 truncate max-w-xs">${escapeHtml(execution.task.name)}</div>
                </td>
                <td class="px-6 py-4 whitespace-nowrap">
                    <span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${statusClass}">${statusText}</span>
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-slate-500 font-mono">${formatDateTime(execution.started_at)}</td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-slate-500 font-mono">${durationText}</td>
                <td class="px-6 py-4">
                    ${outputHtml}
                </td>
            </tr>
        `;
    }).join('');
}

// 更新分页控件
function updatePagination(pagination) {
    executionTotalPages = pagination.total_pages;
    executionTotalItems = pagination.total_items;
    
    // 更新分页信息
    const infoElement = document.getElementById('executionPaginationInfo');
    if (infoElement) {
        const startItem = (pagination.current_page - 1) * pagination.page_size + 1;
        const endItem = Math.min(pagination.current_page * pagination.page_size, pagination.total_items);
        infoElement.textContent = `显示第 ${startItem}-${endItem} 条，共 ${pagination.total_items} 条记录`;
    }
    
    // 更新按钮状态
    const prevBtn = document.getElementById('prevPageBtn');
    const nextBtn = document.getElementById('nextPageBtn');
    
    if (prevBtn) {
        prevBtn.disabled = !pagination.has_prev;
    }
    if (nextBtn) {
        nextBtn.disabled = !pagination.has_next;
    }
    
    // 更新页码按钮
    renderPageNumbers(pagination.current_page, pagination.total_pages);
}

// 渲染页码按钮
function renderPageNumbers(currentPage, totalPages) {
    const pageNumbersContainer = document.getElementById('pageNumbers');
    if (!pageNumbersContainer) return;
    
    const maxVisiblePages = 5;
    let startPage = Math.max(1, currentPage - Math.floor(maxVisiblePages / 2));
    let endPage = Math.min(totalPages, startPage + maxVisiblePages - 1);
    
    if (endPage - startPage + 1 < maxVisiblePages) {
        startPage = Math.max(1, endPage - maxVisiblePages + 1);
    }
    
    let html = '';
    
    // 第一页
    if (startPage > 1) {
        html += `<button onclick="goToPage(1)" class="px-2 py-1 border border-slate-300 rounded-md text-sm font-medium text-slate-700 bg-white hover:bg-slate-50 transition-colors duration-150">1</button>`;
        if (startPage > 2) {
            html += `<span class="px-2 py-1 text-slate-500">...</span>`;
        }
    }
    
    // 页码
    for (let i = startPage; i <= endPage; i++) {
        const isActive = i === currentPage;
        html += `<button onclick="goToPage(${i})" class="px-2 py-1 border rounded-md text-sm font-medium transition-colors duration-150 ${isActive ? 'bg-primary-500 text-white border-primary-500' : 'border-slate-300 text-slate-700 bg-white hover:bg-slate-50'}">${i}</button>`;
    }
    
    // 最后一页
    if (endPage < totalPages) {
        if (endPage < totalPages - 1) {
            html += `<span class="px-2 py-1 text-slate-500">...</span>`;
        }
        html += `<button onclick="goToPage(${totalPages})" class="px-2 py-1 border border-slate-300 rounded-md text-sm font-medium text-slate-700 bg-white hover:bg-slate-50 transition-colors duration-150">${totalPages}</button>`;
    }
    
    pageNumbersContainer.innerHTML = html;
}

// 分页函数
function goToPage(page) {
    executionCurrentPage = page;
    loadExecutions();
}

function previousPage() {
    if (executionCurrentPage > 1) {
        executionCurrentPage--;
        loadExecutions();
    }
}

function nextPage() {
    if (executionCurrentPage < executionTotalPages) {
        executionCurrentPage++;
        loadExecutions();
    }
}

// 搜索函数（用于input的onkeyup事件）
function searchExecutions() {
    const searchInput = document.getElementById('executionSearchInput');
    if (searchInput) {
        executionSearchTerm = searchInput.value.trim();
        executionCurrentPage = 1; // 重置到第一页
        loadExecutions();
    }
}

// 辅助函数
function getStatusClass(status) {
    switch (status) {
        case 'success':
            return 'bg-success-100 text-success-800';
        case 'failed':
            return 'bg-red-100 text-red-800';
        case 'running':
            return 'bg-warning-100 text-warning-800';
        default:
            return 'bg-slate-100 text-slate-800';
    }
}

function getStatusText(status) {
    switch (status) {
        case 'success':
            return '成功';
        case 'failed':
            return '失败';
        case 'running':
            return '运行中';
        default:
            return status;
    }
}

function formatDateTime(dateString) {
    if (!dateString) return '';
    const date = new Date(dateString);
    return date.toLocaleString('zh-CN', {
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit'
    });
}

function escapeHtml(text) {
    if (!text) return '';
    return String(text).replace(/&/g, '&amp;')
                      .replace(/</g, '&lt;')
                      .replace(/>/g, '&gt;')
                      .replace(/"/g, '&quot;')
                      .replace(/'/g, '&#39;');
}

function escapeJs(text) {
    if (!text) return '';
    return String(text).replace(/\\/g, '\\\\')
                      .replace(/'/g, "\\'")
                      .replace(/"/g, '\\"')
                      .replace(/</g, '&lt;')
                      .replace(/>/g, '&gt;')
                      .replace(/&/g, '&amp;')
                      .replace(/\n/g, '\\n')
                      .replace(/\r/g, '\\r')
                      .replace(/\t/g, '\\t');
}