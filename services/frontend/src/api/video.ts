import request from "@/utils/request";
import type {
  Video,
  UploadVideoRequest,
  UpdateVideoRequest,
  SearchVideoRequest,
  GenerateUploadURLRequest,
  GenerateUploadURLResponse,
  ConfirmUploadRequest,
} from "@/types";

// 獲取影片列表
export const getVideos = (params?: { offset?: number; limit?: number }) => {
  return request.get<Video[]>("/videos", { params });
};

// 分離式上傳：第一步 - 獲取預簽名上傳 URL
export const generateUploadURL = (data: GenerateUploadURLRequest) => {
  return request.post<GenerateUploadURLResponse>("/videos/upload-url", data);
};

// 分離式上傳：第三步 - 確認上傳完成
export const confirmUpload = (data: ConfirmUploadRequest) => {
  return request.post("/videos/confirm-upload", data);
};

// 傳統上傳影片（保留作為備用）
export const uploadVideo = (data: UploadVideoRequest) => {
  return request.post<Video>("/videos", data);
};

// 獲取單個影片
export const getVideo = (id: number) => {
  return request.get<Video>(`/videos/${id}`);
};

// 獲取用戶的影片
export const getUserVideos = (userId: number) => {
  return request.get<Video[]>(`/users/${userId}/videos`);
};

// 更新影片
export const updateVideo = (id: number, data: UpdateVideoRequest) => {
  return request.put<Video>(`/videos/${id}`, data);
};

// 刪除影片
export const deleteVideo = (id: number) => {
  return request.delete(`/videos/${id}`);
};

// 搜尋影片
export const searchVideos = (params: SearchVideoRequest) => {
  return request.get<Video[]>("/videos/search", { params });
};

// 按讚影片
export const likeVideo = (id: number) => {
  return request.post(`/videos/${id}/like`);
};
