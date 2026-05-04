import { post } from './request'
import { LoginRequest, LoginResponse } from '@/types'

// 登录
export function login(data: LoginRequest): Promise<LoginResponse> {
  return post<LoginResponse>('/auth/login', data)
}

// 登出
export function logout(): Promise<void> {
  return post<void>('/auth/logout')
}