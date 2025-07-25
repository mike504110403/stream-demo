# 🎬 轉碼流程調試指南

## 📋 轉碼流程概覽

### 完整流程
1. **用戶上傳影片** → 前端獲取預簽名 URL
2. **直接上傳到 MinIO** → 原始檔案存儲在 `videos/original/{user_id}/{uuid}.{ext}`
3. **創建影片記錄** → 狀態為 `uploading`
4. **確認上傳完成** → 調用 `ConfirmUploadAndStartProcessingWithKey`
5. **檢查檔案大小** → 大於 1MB 的檔案進行轉碼
6. **啟動 FFmpeg 轉碼** → 在 Docker 容器中執行轉碼腳本
7. **下載原始檔案** → 從 MinIO 下載到轉碼容器
8. **多格式轉碼** → 同時生成 MP4、HLS、縮圖
9. **上傳轉碼結果** → 存儲到 `videos/processed/{user_id}/{video_id}/`
10. **更新資料庫** → 設置播放 URL 和狀態

## 🔧 調試步驟

### 1. 檢查服務狀態
```bash
# 檢查所有容器
docker ps

# 檢查 FFmpeg 容器日誌
docker logs stream-demo-transcoder

# 檢查後端日誌
tail -f backend/logs/app-$(date +%Y-%m-%d).log
```

### 2. 檢查配置
```bash
# 檢查轉碼配置
cat backend/config/config.local.yaml | grep -A 10 "transcode:"

# 檢查 S3 配置
cat backend/config/config.local.yaml | grep -A 10 "storage:"
```

### 3. 測試 FFmpeg 服務
```bash
# 測試 FFmpeg 容器連接
docker exec stream-demo-transcoder ffmpeg -version

# 測試 MinIO 客戶端
docker exec stream-demo-transcoder mc ls s3/stream-demo-videos

# 手動測試轉碼腳本
docker exec stream-demo-transcoder /scripts/transcode.sh \
  "videos/original/1/test.mov" \
  "videos/processed/1/1" \
  "1" \
  "1"
```

### 4. 檢查 MinIO 檔案
```bash
# 列出所有檔案
docker exec stream-demo-minio mc ls local/stream-demo-videos --recursive

# 檢查原始檔案
docker exec stream-demo-minio mc ls local/stream-demo-videos/videos/original/

# 檢查處理後檔案
docker exec stream-demo-minio mc ls local/stream-demo-videos/videos/processed/
```

### 5. 檢查資料庫狀態
```bash
# 連接到資料庫
docker exec -it stream-demo-postgresql psql -U postgres -d stream_demo

# 查詢影片狀態
SELECT id, title, status, processing_progress, original_key, mp4_url, hls_master_url 
FROM videos 
ORDER BY created_at DESC 
LIMIT 10;
```

## 🐛 常見問題

### 問題 1: 轉碼服務未啟動
**症狀**: 後端日誌顯示 "沒有可用的轉碼服務"
**解決方案**:
```bash
# 重新啟動 FFmpeg 容器
docker-compose restart ffmpeg-transcoder

# 檢查容器狀態
docker ps | grep transcoder
```

### 問題 2: MinIO 連接失敗
**症狀**: FFmpeg 容器無法下載或上傳檔案
**解決方案**:
```bash
# 重新配置 MinIO 客戶端
docker exec stream-demo-transcoder mc alias set s3 http://minio:9000 minioadmin minioadmin

# 測試連接
docker exec stream-demo-transcoder mc ls s3/stream-demo-videos
```

### 問題 3: 檔案權限問題
**症狀**: 轉碼腳本執行失敗
**解決方案**:
```bash
# 設置腳本執行權限
docker exec stream-demo-transcoder chmod +x /scripts/transcode.sh

# 檢查權限
docker exec stream-demo-transcoder ls -la /scripts/
```

### 問題 4: 轉碼超時
**症狀**: 轉碼任務長時間無響應
**解決方案**:
```bash
# 檢查 FFmpeg 容器資源使用
docker stats stream-demo-transcoder

# 重啟轉碼容器
docker-compose restart ffmpeg-transcoder
```

## 📊 監控 API

### 檢查轉碼狀態
```bash
# 獲取影片轉碼狀態
curl -H "Authorization: Bearer YOUR_TOKEN" \
  http://localhost:8080/api/videos/1/transcode-status
```

### 響應格式
```json
{
  "code": 200,
  "data": {
    "video_id": 1,
    "status": "ready",
    "processing_progress": 100,
    "original_url": "http://localhost:9000/stream-demo-videos/videos/original/1/video.mp4",
    "mp4_url": "http://localhost:9000/stream-demo-videos/videos/processed/1/1/video.mp4",
    "hls_master_url": "http://localhost:9000/stream-demo-videos/videos/processed/1/1/hls/index.m3u8",
    "thumbnail_url": "http://localhost:9000/stream-demo-videos/videos/processed/1/1/thumbnails/thumb_640x480.jpg",
    "file_size": 1048576,
    "original_format": "mp4"
  }
}
```

## 🎯 狀態說明

- **uploading**: 上傳中
- **processing**: 處理中（檢查檔案）
- **transcoding**: 轉碼中
- **ready**: 轉碼完成，可以播放
- **failed**: 轉碼失敗
- **completed**: 小檔案，跳過轉碼

## 📝 日誌關鍵字

### 成功流程
- `🔄 開始轉碼流程`
- `🎯 選擇 FFmpeg 轉碼服務`
- `🎬 創建 FFmpeg 轉碼任務`
- `🚀 開始執行 FFmpeg 轉碼`
- `✅ 轉碼任務完成`
- `🎉 處理 FFmpeg 轉碼完成`

### 錯誤流程
- `❌ FFmpeg 轉碼任務創建失敗`
- `❌ 轉碼任務失敗`
- `❌ 沒有可用的轉碼服務`

## 🚀 快速測試

使用提供的測試腳本：
```bash
./test_transcode.sh
```

這個腳本會自動檢查所有關鍵組件並提供診斷信息。 