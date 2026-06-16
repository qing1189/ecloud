import { defineStore } from 'pinia'
import { authAPI, userAPI } from '@/api'

export const useUserStore = defineStore('user', {
  state: () => ({
    token: localStorage.getItem('token') || '',
    isLoggedIn: !!localStorage.getItem('token'),
    userInfo: {
      id: '',
      username: '',
      role: '',
      display_name: '',
      email: ''
    }
  }),

  getters: {
    isAdmin: (state) => state.userInfo.role === 'admin'
  },

  actions: {
    // 登录
    async login(username, password) {
      const res = await authAPI.login(username, password)
      this.token = res.data.token
      this.isLoggedIn = true
      localStorage.setItem('token', res.data.token)

      // 保存用户信息
      if (res.data.user) {
        this.userInfo = res.data.user
      }

      return res
    },

    // 登出
    logout() {
      this.token = ''
      this.isLoggedIn = false
      this.userInfo = {
        id: '',
        username: '',
        role: '',
        display_name: '',
        email: ''
      }
      localStorage.removeItem('token')
    },

    // 验证 Token
    async verify() {
      try {
        await authAPI.verify()
        return true
      } catch (error) {
        this.logout()
        return false
      }
    },

    // 获取当前用户信息
    async fetchUserInfo() {
      try {
        const res = await userAPI.getCurrentUser()
        this.userInfo = res.data
        return res.data
      } catch (error) {
        console.error('获取用户信息失败:', error)
        throw error
      }
    }
  }
})
