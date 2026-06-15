import request from '@/utils/request'

// 认证相关
export const authAPI = {
  // 登录
  login(password) {
    return request.post('/auth/login', { password })
  },

  // 验证 Token
  verify() {
    return request.get('/auth/verify')
  },

  // 修改密码
  changePassword(oldPassword, newPassword) {
    return request.post('/auth/change-password', {
      old_password: oldPassword,
      new_password: newPassword
    })
  }
}

// 账号管理
export const accountAPI = {
  // 获取账号列表
  list() {
    return request.get('/accounts')
  },

  // 添加账号
  add(data) {
    return request.post('/accounts', data)
  },

  // 验证设备（提交验证码）
  verifyDevice(data) {
    return request.post('/accounts/verify', data)
  },

  // 重新发送验证码
  resendCode(data) {
    return request.post('/accounts/resend-code', data)
  },

  // 获取账号详情
  get(id) {
    return request.get(`/accounts/${id}`)
  },

  // 更新账号
  update(id, data) {
    return request.put(`/accounts/${id}`, data)
  },

  // 删除账号
  delete(id) {
    return request.delete(`/accounts/${id}`)
  },

  // 切换监控状态
  toggle(id) {
    return request.post(`/accounts/${id}/toggle`)
  }
}

// 监控状态
export const monitorAPI = {
  // 获取所有任务状态
  getAllStatus() {
    return request.get('/monitor/status')
  },

  // 获取单个任务状态
  getStatus(id) {
    return request.get(`/monitor/status/${id}`)
  },

  // 重新加载所有任务
  reload() {
    return request.post('/monitor/reload')
  }
}

// 操作日志
export const logAPI = {
  // 获取日志列表
  list(params) {
    return request.get('/logs', { params })
  }
}
