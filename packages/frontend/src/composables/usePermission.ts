import { useAuthStore } from '@/stores/auth'
import { Role } from '@/types'

export function usePermission() {
  const authStore = useAuthStore()

  // 判断是否有指定角色
  function hasRole(role: Role): boolean {
    return authStore.hasRole(role)
  }

  // 判断是否有任一角色
  function hasAnyRole(roles: Role[]): boolean {
    return authStore.hasAnyRole(roles)
  }

  // 判断是否可以编辑设备
  function canEditDevice(): boolean {
    return hasRole(Role.ADMIN)
  }

  // 判断是否可以控制设备
  function canControlDevice(): boolean {
    return hasAnyRole([Role.ADMIN, Role.OPERATOR])
  }

  // 判断是否可以管理用户
  function canManageUser(): boolean {
    return hasRole(Role.ADMIN)
  }

  return {
    hasRole,
    hasAnyRole,
    canEditDevice,
    canControlDevice,
    canManageUser
  }
}