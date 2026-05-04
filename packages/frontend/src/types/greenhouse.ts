// 温室
export interface Greenhouse {
  id: number
  name: string
  location: string | null
  description: string | null
  created_at: string
  updated_at: string
}

// 温室查询参数
export interface GreenhouseQueryParams {
  page?: number
  page_size?: number
  keyword?: string
}

// 创建/更新温室参数
export interface GreenhouseFormData {
  name: string
  location?: string
  description?: string
}
