// API 响应通用类型
export interface ApiResponse<T = unknown> {
  code: number
  message: string
  data: T
  request_id: string
}

// 分页响应
export interface PaginatedData<T> {
  page: number
  page_size: number
  total: number
  items: T[]
}

// 业务错误码
export enum ErrorCode {
  SUCCESS = 0,
  VALIDATION_ERROR = 10001,
  UNAUTHORIZED = 10002,
  FORBIDDEN = 10003,
  NOT_FOUND = 10004,
  CONFLICT = 10005,
  RATE_LIMIT = 10006,
  DEVICE_OFFLINE = 10007,
  RULE_CONFLICT = 10008,
  DATA_OUT_OF_RANGE = 10009,
  DEVICE_CODE_DUPLICATE = 10010
}