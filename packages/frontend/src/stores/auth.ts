import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { login as loginApi, logout as logoutApi } from '@/api/auth'
import { setToken, setUserInfo, clearAuth, getToken, getUserInfo } from '@/utils/storage'
import { User, Role, LoginRequest } from '@/types'

export const useAuthStore = defineStore('auth', () => {
  // 状态
  const user = ref<User | null>(null)
  const token = ref<string | null>(null)

  // 初始化 - 从 localStorage 恢复状态
  function init() {
    const savedToken = getToken()
    const savedUser = getUserInfo()
    if (savedToken && savedUser) {
      token.value = savedToken
      try {
        user.value = JSON.parse(savedUser)
      } catch {
        user.value = null
      }
    }
  }

  // 登录
  async function login(credentials: LoginRequest) {
    const response = await loginApi(credentials)
    token.value = response.token
    user.value = response.user
    setToken(response.token)
    setUserInfo(JSON.stringify(response.user))
    return response
  }

  // 登出
  async function logout() {
    try {
      await logoutApi()
    } catch (error) {
      // 登出接口失败不影响清除本地状态
      console.warn('[Auth] Logout API failed:', error)
    } finally {
      user.value = null
      token.value = null
      clearAuth()
    }
  }

  // 计算属性
  const isLoggedIn = computed(() => !!token.value && !!user.value)
  const roles = computed(() => user.value?.roles || [])
  const isAdmin = computed(() => roles.value.includes(Role.ADMIN))
  const isOperator = computed(() => roles.value.includes(Role.OPERATOR))
  const isViewer = computed(() => roles.value.includes(Role.VIEWER))

  // 权限判断
  function hasRole(role: Role): boolean {
    return roles.value.includes(role)
  }

  function hasAnyRole(checkRoles: Role[]): boolean {
    return checkRoles.some((r) => roles.value.includes(r))
  }

  return {
    user,
    token,
    isLoggedIn,
    roles,
    isAdmin,
    isOperator,
    isViewer,
    init,
    login,
    logout,
    hasRole,
    hasAnyRole
  }
})