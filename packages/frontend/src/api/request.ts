import axios, { AxiosInstance, AxiosRequestConfig, AxiosResponse, InternalAxiosRequestConfig } from 'axios'
import { ElMessage } from 'element-plus'
import { getToken, clearAuth } from '@/utils/storage'
import { ApiResponse, ErrorCode } from '@/types'

// 创建 Axios 实例
const instance: AxiosInstance = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '/api',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json'
  }
})

// 请求拦截器
instance.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    const token = getToken()
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// 响应拦截器
instance.interceptors.response.use(
  (response: AxiosResponse<ApiResponse>) => {
    const { data } = response

    // 业务错误处理
    if (data.code !== ErrorCode.SUCCESS) {
      console.error('[API Error]', data.message, data)
      ElMessage.error(data.message || '请求失败')
      return Promise.reject(new Error(data.message))
    }

    return response
  },
  (error) => {
    const { response } = error

    if (response) {
      const { status, data } = response

      console.error('[HTTP Error]', status, data?.message || error.message)

      switch (status) {
        case 401:
          ElMessage.error('登录已过期，请重新登录')
          clearAuth()
          // 动态导入路由避免循环依赖
          import('@/router').then((r) => r.default.push('/login'))
          break
        case 403:
          ElMessage.error('没有权限访问')
          break
        case 404:
          ElMessage.error('资源不存在')
          break
        case 429:
          ElMessage.error('请求过于频繁，请稍后再试')
          break
        default:
          ElMessage.error(data?.message || '请求失败')
      }
    } else {
      console.error('[Network Error]', error.message)
      ElMessage.error('网络异常，请检查网络连接')
    }

    return Promise.reject(error)
  }
)

// 封装请求方法
export async function request<T>(config: AxiosRequestConfig): Promise<T> {
  const response = await instance.request<ApiResponse<T>>(config)
  return response.data.data
}

export function get<T>(url: string, params?: object, config?: AxiosRequestConfig): Promise<T> {
  return request<T>({ ...config, method: 'GET', url, params })
}

export function post<T>(url: string, data?: object, config?: AxiosRequestConfig): Promise<T> {
  return request<T>({ ...config, method: 'POST', url, data })
}

export function put<T>(url: string, data?: object, config?: AxiosRequestConfig): Promise<T> {
  return request<T>({ ...config, method: 'PUT', url, data })
}

export function patch<T>(url: string, data?: object, config?: AxiosRequestConfig): Promise<T> {
  return request<T>({ ...config, method: 'PATCH', url, data })
}

export function del<T>(url: string, config?: AxiosRequestConfig): Promise<T> {
  return request<T>({ ...config, method: 'DELETE', url })
}

export default instance