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
  username: string
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

// 影片品質類型
export interface VideoQuality {
  id: number
  video_id: number
  quality: string
  width: number
  height: number
  bitrate: number
  file_url: string
  file_key: string
  status: string
  created_at: string
  updated_at: string
}

// 影片相關類型
export interface Video {
  id: number
  title: string
  description?: string
  user_id: number
  username?: string
  original_url?: string // 原始影片 URL
  video_url?: string // 舊字段，保持兼容性
  thumbnail_url?: string // 縮圖 URL
  hls_master_url?: string // HLS 播放列表 URL
  mp4_url?: string // MP4 轉碼版本 URL（網頁播放）
  status: 'processing' | 'ready' | 'failed' | 'uploading' | 'transcoding'
  processing_progress?: number
  duration?: number // 影片長度（秒）
  file_size?: number // 檔案大小（字節）
  views: number
  likes: number
  created_at: string
  updated_at: string
  user?: User
  qualities?: VideoQuality[] // 影片品質列表
}

export interface UploadVideoRequest {
  title: string
  description?: string
}

// 分離式上傳相關類型
export interface GenerateUploadURLRequest {
  title: string
  description?: string
  filename: string
  file_size: number
}

export interface GenerateUploadURLResponse {
  upload_url: string
  form_data: Record<string, string>
  key: string
  video: Video
}

export interface ConfirmUploadRequest {
  video_id: number
  s3_key: string
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

// 直播間相關類型
export interface LiveRoomInfo {
  id: string
  title: string
  description: string
  creator_id: number
  status: 'created' | 'live' | 'ended' | 'cancelled'
  stream_key: string
  viewer_count: number
  max_viewers: number
  started_at: string
  created_at: string
  updated_at: string
}

export interface CreateRoomRequest {
  title: string
  description?: string
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
