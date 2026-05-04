import { useAuthStore } from '@/stores/auth'
import { storeToRefs } from 'pinia'

export function useAuth() {
  const authStore = useAuthStore()
  const { user, isLoggedIn, roles, isAdmin, isOperator, isViewer } = storeToRefs(authStore)

  return {
    user,
    isLoggedIn,
    roles,
    isAdmin,
    isOperator,
    isViewer,
    login: authStore.login,
    logout: authStore.logout,
    hasRole: authStore.hasRole,
    hasAnyRole: authStore.hasAnyRole
  }
}