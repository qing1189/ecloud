import { defineStore } from 'pinia'
import { authAPI } from '@/api'

export const useUserStore = defineStore('user', {
  state: () => ({
    token: localStorage.getItem('token') || '',
    isLoggedIn: !!localStorage.getItem('token')
  }),

  actions: {
    // 登录
    async login(password) {
      const res = await authAPI.login(password)
      this.token = res.data.token
      this.isLoggedIn = true
      localStorage.setItem('token', res.data.token)
      return res
    },

    // 登出
    logout() {
      this.token = ''
      this.isLoggedIn = false
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
    }
  }
})
