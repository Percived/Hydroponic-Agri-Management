const TOKEN_KEY = 'hydroponic_token'
const USER_KEY = 'hydroponic_user'

// Token 相关
export function getToken(): string | null {
  return localStorage.getItem(TOKEN_KEY)
}

export function setToken(token: string): void {
  localStorage.setItem(TOKEN_KEY, token)
}

export function removeToken(): void {
  localStorage.removeItem(TOKEN_KEY)
}

// 用户信息相关
export function getUserInfo(): string | null {
  return localStorage.getItem(USER_KEY)
}

export function setUserInfo(user: string): void {
  localStorage.setItem(USER_KEY, user)
}

export function removeUserInfo(): void {
  localStorage.removeItem(USER_KEY)
}

// 清除所有认证信息
export function clearAuth(): void {
  removeToken()
  removeUserInfo()
}