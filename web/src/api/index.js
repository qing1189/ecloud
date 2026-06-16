import request from '@/utils/request'

// 认证相关
export const authAPI = {
  // 登录
  login(username, password) {
    return request.post('/auth/login', { username, password })
  },

  // 验证 Token
  verify() {
    return request.get('/auth/verify')
  }
}

// 用户管理
export const userAPI = {
  // 获取当前用户信息
  getCurrentUser() {
    return request.get('/users/me')
  },

  // 获取用户列表（仅管理员）
  list() {
    return request.get('/users')
  },

  // 创建用户（仅管理员）
  create(data) {
    return request.post('/users', data)
  },

  // 获取用户详情
  get(id) {
    return request.get(`/users/${id}`)
  },

  // 更新用户信息
  update(id, data) {
    return request.put(`/users/${id}`, data)
  },

  // 删除用户（仅管理员）
  delete(id) {
    return request.delete(`/users/${id}`)
  },

  // 修改密码
  changePassword(id, oldPassword, newPassword) {
    return request.put(`/users/${id}/password`, {
      old_password: oldPassword,
      new_password: newPassword
    })
  },

  // 重置密码（管理员）
  resetPassword(id, newPassword) {
    return request.put(`/users/${id}/password`, {
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
