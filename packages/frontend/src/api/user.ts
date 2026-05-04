import { get, post, put, patch } from './request'
import type { User, UserQueryParams, PaginatedData, UserStatus } from '@/types'
import type { CreateUserParams, UpdateUserParams } from '@/types'

// 获取用户列表
export const getUsers = (params?: UserQueryParams) =>
  get<PaginatedData<User>>('/users', params)

// 获取用户详情
export const getUserDetail = (id: number) =>
  get<User>(`/users/${id}`)

// 创建用户
export const createUser = (data: CreateUserParams) =>
  post<User>('/users', data)

// 更新用户
export const updateUser = (id: number, data: UpdateUserParams) =>
  put<User>(`/users/${id}`, data)

// 更新用户状态
export const updateUserStatus = (id: number, status: UserStatus) =>
  patch<User>(`/users/${id}/status`, { status })

// 获取角色列表
export const getRoles = () =>
  get<{ name: string; description: string }[]>('/roles')
