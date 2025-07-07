// 用戶相關類型
export interface User {
  id: number
  username: string
  email: string
  avatar?: string
  bio?: string
  created_at: string
  updated_at: string
}

export interface LoginRequest {
  email: string
  password: string
}

export interface RegisterRequest {
  username: string
  email: string
  password: string
}

export interface UpdateUserRequest {
  username?: string
  email?: string
  avatar?: string
  bio?: string
}

// 影片相關類型
export interface Video {
  id: number
  title: string
  description?: string
  user_id: number
  video_url: string
  thumbnail_url?: string
  status: 'processing' | 'ready' | 'failed'
  views: number
  likes: number
  created_at: string
  updated_at: string
  user?: User
}

export interface UploadVideoRequest {
  title: string
  description?: string
}

export interface UpdateVideoRequest {
  title?: string
  description?: string
}

export interface SearchVideoRequest {
  q: string
  offset?: number
  limit?: number
}

// 直播相關類型
export interface Live {
  id: number
  title: string
  description?: string
  user_id: number
  status: 'scheduled' | 'live' | 'ended'
  start_time: string
  end_time?: string
  stream_key: string
  viewer_count: number
  chat_enabled: boolean
  created_at: string
  updated_at: string
  user?: User
}

export interface CreateLiveRequest {
  title: string
  description?: string
  start_time: string
}

export interface UpdateLiveRequest {
  title?: string
  description?: string
  start_time?: string
}

// 支付相關類型
export interface Payment {
  id: number
  user_id: number
  amount: number
  currency: string
  status: 'pending' | 'completed' | 'failed' | 'refunded'
  payment_method: string
  transaction_id: string
  description?: string
  refund_reason?: string
  created_at: string
  updated_at: string
  user?: User
}

export interface CreatePaymentRequest {
  amount: number
  currency: string
  payment_method: string
  description?: string
}

export interface ProcessPaymentRequest {
  payment_method: string
  transaction_id: string
}

export interface RefundPaymentRequest {
  reason: string
}

// 聊天相關類型
export interface ChatMessage {
  id: number
  live_id: number
  user_id: number
  username: string
  content: string
  type: 'text' | 'system' | 'gift'
  created_at: string
}

// API 響應類型
export interface ApiResponse<T = any> {
  success: boolean
  data?: T
  message?: string
  code?: number
}

// 分頁類型
export interface PaginationParams {
  page?: number
  limit?: number
  offset?: number
}

export interface PaginationResponse<T> {
  data: T[]
  total: number
  page: number
  limit: number
} 