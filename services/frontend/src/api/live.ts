import request from "@/utils/request";
import type { Live, CreateLiveRequest, UpdateLiveRequest } from "@/types";

// 獲取直播列表
export const getLives = (params?: { offset?: number; limit?: number }) => {
  return request.get<Live[]>("/lives", { params });
};

// 創建直播
export const createLive = (data: CreateLiveRequest) => {
  return request.post<Live>("/lives", data);
};

// 獲取單個直播
export const getLive = (id: number) => {
  return request.get<Live>(`/lives/${id}`);
};

// 獲取用戶的直播
export const getUserLives = (userId: number) => {
  return request.get<Live[]>(`/users/${userId}/lives`);
};

// 更新直播
export const updateLive = (id: number, data: UpdateLiveRequest) => {
  return request.put<Live>(`/lives/${id}`, data);
};

// 刪除直播
export const deleteLive = (id: number) => {
  return request.delete(`/lives/${id}`);
};

// 開始直播
export const startLive = (id: number) => {
  return request.post(`/lives/${id}/start`);
};

// 結束直播
export const endLive = (id: number) => {
  return request.post(`/lives/${id}/end`);
};

// 獲取串流金鑰
export const getStreamKey = (id: number) => {
  return request.get<{ stream_key: string }>(`/lives/${id}/stream-key`);
};

// 切換聊天功能
export const toggleChat = (id: number, enabled: boolean) => {
  return request.post(`/lives/${id}/chat/toggle`, { enabled });
};
