// Login page functionality
document.addEventListener('DOMContentLoaded', function() {
    const loginForm = document.getElementById('loginForm');
    if (loginForm) {
        loginForm.addEventListener('submit', handleLogin);
    }
});

// 显示强制修改密码模态框
function showForcePasswordChangeModal() {
    // 创建模态框HTML
    const modalHTML = `
        <div id="forcePasswordChangeModal" class="modal-overlay fixed inset-0 bg-black bg-opacity-50 backdrop-blur-sm flex items-center justify-center z-50">
            <div class="modal bg-white rounded-xl shadow-xl max-w-md w-full mx-4">
                <div class="flex justify-between items-center p-6 border-b border-slate-200">
                    <h3 class="text-xl font-semibold text-slate-900">首次登录 - 修改密码</h3>
                </div>
                <div class="p-6">
                    <form id="forcePasswordChangeForm" class="space-y-4">
                        <div class="space-y-2">
                            <label class="block text-sm font-medium text-slate-700">当前密码</label>
                            <input type="password" id="forceCurrentPassword" required 
                                   class="w-full px-3 py-2 border border-slate-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-primary-500">
                        </div>
                        <div class="space-y-2">
                            <label class="block text-sm font-medium text-slate-700">新密码 (至少6位)</label>
                            <input type="password" id="forceNewPassword" required minlength="6"
                                   class="w-full px-3 py-2 border border-slate-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-primary-500">
                        </div>
                        <div class="space-y-2">
                            <label class="block text-sm font-medium text-slate-700">确认新密码</label>
                            <input type="password" id="forceConfirmPassword" required 
                                   class="w-full px-3 py-2 border border-slate-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-primary-500">
                        </div>
                        <div class="flex justify-end space-x-3 pt-4">
                            <button type="submit" class="px-4 py-2 bg-primary-600 hover:bg-primary-700 text-white font-medium rounded-lg">
                                修改密码
                            </button>
                        </div>
                    </form>
                </div>
            </div>
        </div>
    `;
    
    // 添加到页面
    document.body.insertAdjacentHTML('beforeend', modalHTML);
    
    // 绑定表单提交事件
    const form = document.getElementById('forcePasswordChangeForm');
    form.addEventListener('submit', handleForcePasswordChange);
}

// 处理强制修改密码
async function handleForcePasswordChange(e) {
    e.preventDefault();
    
    const currentPassword = document.getElementById('forceCurrentPassword').value;
    const newPassword = document.getElementById('forceNewPassword').value;
    const confirmPassword = document.getElementById('forceConfirmPassword').value;
    
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
            window.b1cron.showToast('密码修改成功，正在跳转...', 'success');
            // 移除模态框
            document.getElementById('forcePasswordChangeModal').remove();
            // 跳转到dashboard
            setTimeout(() => {
                window.location.href = '/dashboard';
            }, 1500);
        } else {
            window.b1cron.showToast(data.message || '密码修改失败', 'error');
        }
    } catch (error) {
        window.b1cron.showToast('网络错误，请重试', 'error');
    }
}

async function handleLogin(e) {
    e.preventDefault();
    
    const form = e.target;
    const submitBtn = form.querySelector('button[type="submit"]');
    const btnText = submitBtn.querySelector('.btn-text');
    const btnLoading = submitBtn.querySelector('.btn-loading');
    
    const username = document.getElementById('username').value;
    const password = document.getElementById('password').value;
    
    // 显示加载状态
    btnText.classList.add('hidden');
    btnLoading.classList.remove('hidden');
    submitBtn.disabled = true;
    
    try {
        const response = await fetch('/login', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                username: username,
                password: password
            })
        });
        
        const data = await response.json();
        
        if (response.ok) {
            // 检查是否需要强制修改密码
            if (data.force_password_change) {
                window.b1cron.showToast(data.message || '首次登录，请修改密码', 'warning');
                // 显示修改密码模态框
                showForcePasswordChangeModal();
            } else {
                // 正常登录成功
                window.b1cron.showToast('登录成功!', 'success');
                setTimeout(() => {
                    window.location.href = '/dashboard';
                }, 1000);
            }
        } else {
            // 显示错误信息
            window.b1cron.showToast(data.message || '登录失败', 'error');
        }
    } catch (error) {
        window.b1cron.showToast('网络错误，请重试', 'error');
    } finally {
        // 恢复按钮状态
        btnText.classList.remove('hidden');
        btnLoading.classList.add('hidden');
        submitBtn.disabled = false;
    }
}