import { createRouter, createWebHistory } from 'vue-router'
import { useUserStore } from '@/stores/user'

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/Login.vue'),
    meta: { requiresAuth: false }
  },
  {
    path: '/',
    component: () => import('@/views/Layout.vue'),
    redirect: '/dashboard',
    meta: { requiresAuth: true },
    children: [
      {
        path: 'dashboard',
        name: 'Dashboard',
        component: () => import('@/views/Dashboard.vue'),
        meta: { title: '仪表盘' }
      },
      {
        path: 'accounts',
        name: 'Accounts',
        component: () => import('@/views/Accounts.vue'),
        meta: { title: '账号管理' }
      },
      {
        path: 'monitor',
        name: 'Monitor',
        component: () => import('@/views/Monitor.vue'),
        meta: { title: '实时监控' }
      },
      {
        path: 'logs',
        name: 'Logs',
        component: () => import('@/views/Logs.vue'),
        meta: { title: '操作日志' }
      },
      {
        path: 'users',
        name: 'Users',
        component: () => import('@/views/Users.vue'),
        meta: { title: '用户管理', requiresAdmin: true }
      }
    ]
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// 路由守卫
router.beforeEach(async (to, from, next) => {
  const userStore = useUserStore()

  if (to.meta.requiresAuth) {
    // 需要认证的页面
    if (userStore.isLoggedIn) {
      // 如果用户信息为空，先获取用户信息
      if (!userStore.userInfo.id) {
        try {
          await userStore.fetchUserInfo()
        } catch (error) {
          userStore.logout()
          next('/login')
          return
        }
      }

      // 检查是否需要管理员权限
      if (to.meta.requiresAdmin && !userStore.isAdmin) {
        next('/dashboard')
        return
      }

      next()
    } else {
      next('/login')
    }
  } else {
    // 不需要认证的页面（登录页）
    if (to.path === '/login' && userStore.isLoggedIn) {
      next('/')
    } else {
      next()
    }
  }
})

export default router
