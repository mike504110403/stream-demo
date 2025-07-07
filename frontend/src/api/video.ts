import request from '@/utils/request'
import type { 
  Video, 
  UploadVideoRequest, 
  UpdateVideoRequest,
  SearchVideoRequest 
} from '@/types'

// 獲取影片列表
export const getVideos = (params?: { offset?: number; limit?: number }) => {
  return request.get<Video[]>('/videos', { params })
}

// 上傳影片
export const uploadVideo = (data: UploadVideoRequest) => {
  return request.post<Video>('/videos', data)
}

// 獲取單個影片
export const getVideo = (id: number) => {
  return request.get<Video>(`/videos/${id}`)
}

// 獲取用戶的影片
export const getUserVideos = (userId: number) => {
  return request.get<Video[]>(`/users/${userId}/videos`)
}

// 更新影片
export const updateVideo = (id: number, data: UpdateVideoRequest) => {
  return request.put<Video>(`/videos/${id}`, data)
}

// 刪除影片
export const deleteVideo = (id: number) => {
  return request.delete(`/videos/${id}`)
}

// 搜尋影片
export const searchVideos = (params: SearchVideoRequest) => {
  return request.get<Video[]>('/videos/search', { params })
}

// 按讚影片
export const likeVideo = (id: number) => {
  return request.post(`/videos/${id}/like`)
}
