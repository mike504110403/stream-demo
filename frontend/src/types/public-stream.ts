// 公開流相關類型定義

export interface PublicStreamInfo {
  name: string
  title: string
  description: string
  url: string
  status: string
  last_update: string
  viewer_count: number
  category: string
}

export interface PublicStreamResponse {
  success: boolean
  data: {
    streams: PublicStreamInfo[]
    total: number
  }
}

// 響應攔截器處理後的格式
export interface PublicStreamData {
  streams: PublicStreamInfo[]
  total: number
}

export interface PublicStreamDetailResponse {
  success: boolean
  data: PublicStreamInfo
}

export interface PublicStreamURLResponse {
  success: boolean
  data: {
    stream_name: string
    urls: {
      hls: string
    }
  }
}

export interface PublicStreamStats {
  stream_name: string
  title: string
  status: string
  viewer_count: number
  last_update: string
  category: string
}
