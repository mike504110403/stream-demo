import request from '@/utils/request'
import type { LiveRoomInfo } from '@/types'

// 獲取活躍直播間列表
export const getActiveRooms = (params?: { limit?: number }) => {
  return request.get<LiveRoomInfo[]>('/live-rooms', { params })
}

// 獲取所有直播間列表（包括已結束的）
export const getAllRooms = (params?: { limit?: number }) => {
  return request.get<LiveRoomInfo[]>('/live-rooms/all', { params })
}

// 創建直播間
export const createRoom = (data: { title: string; description?: string }) => {
  return request.post<LiveRoomInfo>('/live-rooms', data)
}

// 獲取單個直播間信息
export const getRoomById = (roomId: string) => {
  return request.get<LiveRoomInfo>(`/live-rooms/${roomId}`)
}

// 加入直播間
export const joinRoom = (roomId: string) => {
  return request.post(`/live-rooms/${roomId}/join`)
}

// 離開直播間
export const leaveRoom = (roomId: string) => {
  return request.post(`/live-rooms/${roomId}/leave`)
}

// 開始直播
export const startLive = (roomId: string) => {
  return request.post(`/live-rooms/${roomId}/start`)
}

// 結束直播
export const endLive = (roomId: string) => {
  return request.post(`/live-rooms/${roomId}/end`)
}

// 關閉直播間
export const closeRoom = (roomId: string) => {
  return request.delete(`/live-rooms/${roomId}`)
}

// 獲取用戶在房間中的角色
export const getUserRole = (roomId: string) => {
  return request.get<{ role: string }>(`/live-rooms/${roomId}/role`)
}
