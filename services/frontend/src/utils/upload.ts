import { generateUploadURL, confirmUpload } from "@/api/video";
import type { GenerateUploadURLRequest } from "@/types";

export interface UploadProgress {
  percent: number;
  message: string;
  status: "uploading" | "success" | "error" | "pending";
}

export type ProgressCallback = (progress: UploadProgress) => void;

/**
 * 分離式影片上傳函數
 * @param file 要上傳的檔案
 * @param metadata 影片metadata（標題、描述等）
 * @param onProgress 進度回調函數
 * @returns Promise<影片信息>
 */
export async function uploadVideoSeparately(
  file: File,
  metadata: Omit<GenerateUploadURLRequest, "filename" | "file_size">,
  onProgress?: ProgressCallback,
) {
  try {
    // 第一步：獲取預簽名上傳 URL
    onProgress?.({
      percent: 10,
      message: "正在獲取上傳 URL...",
      status: "uploading",
    });

    const uploadUrlResponse = await generateUploadURL({
      ...metadata,
      filename: file.name,
      file_size: file.size,
    });

    // Axios 攔截器已經提取了 data 字段，直接解構
    const { upload_url, key, video } = uploadUrlResponse;

    onProgress?.({
      percent: 20,
      message: "開始上傳檔案...",
      status: "uploading",
    });

    // 第二步：直接上傳檔案到 S3 (使用 PUT 方式)
    // 使用 XMLHttpRequest 實現進度追蹤
    await new Promise<void>((resolve, reject) => {
      const xhr = new XMLHttpRequest();

      // 上傳進度
      xhr.upload.addEventListener("progress", (e) => {
        if (e.lengthComputable) {
          const percent = 20 + Math.round((e.loaded / e.total) * 60); // 20-80%
          onProgress?.({
            percent,
            message: `上傳中... ${Math.round((e.loaded / e.total) * 100)}%`,
            status: "uploading",
          });
        }
      });

      // 上傳完成
      xhr.addEventListener("load", () => {
        if (xhr.status >= 200 && xhr.status < 300) {
          resolve();
        } else {
          console.error(
            "S3 上傳失敗:",
            xhr.status,
            xhr.statusText,
            xhr.responseText,
          );
          reject(new Error(`S3 上傳失敗: ${xhr.statusText} (${xhr.status})`));
        }
      });

      // 上傳錯誤
      xhr.addEventListener("error", () => {
        reject(new Error("網路錯誤"));
      });

      // 上傳超時
      xhr.addEventListener("timeout", () => {
        reject(new Error("上傳超時"));
      });

      xhr.timeout = 300000; // 5分鐘超時

      // 設置請求頭（簡化版本，避免簽名問題）
      xhr.open("PUT", upload_url);
      // 不設置 Content-Type，讓瀏覽器自動處理

      // 直接發送文件數據，不使用 FormData
      xhr.send(file);
    });

    onProgress?.({
      percent: 90,
      message: "確認上傳完成...",
      status: "uploading",
    });

    // 第三步：確認上傳完成
    const confirmResponse = await confirmUpload({
      video_id: video.id,
      s3_key: key,
    });

    onProgress?.({
      percent: 100,
      message: "上傳完成！轉碼處理已開始，請稍後查看影片列表",
      status: "success",
    });

    return confirmResponse;
  } catch (error) {
    const errorMessage = error instanceof Error ? error.message : "上傳失敗";

    onProgress?.({
      percent: -1,
      message: `上傳失敗: ${errorMessage}`,
      status: "error",
    });

    throw error;
  }
}

/**
 * 簡化版上傳函數（不需要進度追蹤）
 */
export async function uploadVideo(
  file: File,
  title: string,
  description?: string,
) {
  return uploadVideoSeparately(file, { title, description });
}
