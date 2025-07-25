# 轉碼系統重構指南

## 📋 概述

轉碼系統已經重構為更簡潔、更易維護的架構。主要改進包括：

1. **簡化流程**: 減少不必要的分層，使轉碼流程更清晰
2. **Go 測試**: 使用 Go 測試替代複雜的腳本測試
3. **簡化管理**: 只保留必要的 Docker Compose 管理腳本

## 🏗️ 架構設計

### 核心組件

1. **轉碼工作服務** (`TranscodeWorker`)
   - 簡化的背景服務
   - 使用 `SELECT FOR UPDATE` 鎖機制
   - 清晰的狀態管理

2. **FFmpeg 轉碼服務** (`ffmpeg-transcoder`)
   - 負責實際的影片轉碼工作
   - 支援多種輸出格式（HLS、MP4）

3. **MinIO 存儲**
   - `stream-demo-videos`: 原始影片桶
   - `stream-demo-processed`: 處理後影片桶

4. **PostgreSQL 資料庫**
   - 存儲影片元資料和狀態
   - 支援事務和鎖機制

## 🏗️ 架構設計

### 核心組件

1. **轉碼排程服務** (`transcode-scheduler`)
   - 獨立的 Docker 容器
   - 持續監控資料庫中的影片狀態
   - 使用事務和鎖機制確保資料一致性

2. **FFmpeg 轉碼服務** (`ffmpeg-transcoder`)
   - 負責實際的影片轉碼工作
   - 支援多種輸出格式（HLS、MP4）
   - 生成多品質版本

3. **MinIO 存儲**
   - `stream-demo-videos`: 原始影片桶
   - `stream-demo-processed`: 處理後影片桶

4. **PostgreSQL 資料庫**
   - 存儲影片元資料和狀態
   - 支援事務和鎖機制

## 🚀 快速開始

### 1. 啟動服務

```bash
# 啟動所有服務
./docker-manage.sh start

# 或者使用 docker-compose
docker-compose up -d
```

### 2. 檢查服務狀態

```bash
# 查看服務狀態
./docker-manage.sh status

# 查看日誌
./docker-manage.sh logs

# 查看特定服務日誌
./docker-manage.sh logs ffmpeg-transcoder
```

### 3. 運行測試

```bash
# 運行 Go 測試
./docker-manage.sh test

# 或者直接運行
cd backend && go test ./services -v
```

## 🔧 配置說明

### 環境變數

| 變數名 | 說明 | 預設值 |
|--------|------|--------|
| `CONFIG_PATH` | 配置文件路徑 | `config/config.local.yaml` |
| `ENV` | 運行環境 | `local` |

### 配置文件

轉碼相關配置在 `backend/config/config.local.yaml` 中：

```yaml
transcode:
  type: "ffmpeg"  # 轉碼類型：ffmpeg 或 media_convert
  ffmpeg:
    enabled: true
    container_name: "stream-demo-transcoder"

media_convert:
  enabled: false
  region: "us-west-2"
  role_arn: "arn:aws:iam::..."
  endpoint: "https://mediaconvert.us-west-2.amazonaws.com"

storage:
  type: "s3"
  s3:
    region: "us-east-1"
    bucket: "stream-demo-videos"
    access_key: "minioadmin"
    secret_key: "minioadmin"
    endpoint: "http://localhost:9000"
    cdn_domain: "localhost:9000"
```

## 📊 監控和日誌

### 日誌關鍵字

- `🚀 啟動背景轉碼工作服務` - 服務啟動
- `🔍 服務啟動時檢查待轉碼影片...` - 啟動時檢查
- `📋 發現 X 個待轉碼影片` - 發現待處理影片
- `🎬 開始處理影片 ID: X` - 開始處理影片
- `✅ 影片轉碼完成` - 轉碼完成
- `❌ 影片轉碼失敗` - 轉碼失敗

### 健康檢查

服務包含健康檢查機制：

```yaml
healthcheck:
  test: ["CMD", "pgrep", "-f", "transcode_scheduler"]
  interval: 30s
  timeout: 10s
  retries: 3
```

## 🔒 並發安全機制

### SELECT FOR UPDATE 鎖

```go
// 使用事務和鎖來避免並發問題
tx := w.videoService.Repo.GetDB().Begin()
if err := tx.Where("status IN ?", []string{"uploading", "processing"}).
    Order("created_at ASC").
    Limit(5).
    Clauses(clause.Locking{Strength: "UPDATE"}).
    Find(&videos).Error; err != nil {
    // 處理錯誤
}
```

### 事務管理

所有資料庫操作都使用事務確保一致性：

```go
// 開始事務
tx := w.videoService.Repo.GetDB().Begin()
defer func() {
    if r := recover(); r != nil {
        tx.Rollback()
    }
}()

// 執行操作
if err := tx.Model(&video).Updates(updates).Error; err != nil {
    tx.Rollback()
    return
}

// 提交事務
if err := tx.Commit().Error; err != nil {
    return
}
```

## 📈 性能優化

### 批量處理

- 每次最多處理 5 個影片
- 避免系統過載
- 可配置處理間隔

### 資源控制

- 使用連接池管理資料庫連接
- 限制並發轉碼任務數量
- 自動清理完成的任務

## 🛠️ 故障排除

### 常見問題

1. **服務無法啟動**
   ```bash
   # 檢查 Docker 是否運行
   docker info
   
   # 檢查網絡
   docker network ls | grep stream-demo-network
   
   # 查看詳細日誌
   docker logs stream-demo-transcode-scheduler
   ```

2. **轉碼失敗**
   ```bash
   # 檢查 FFmpeg 容器
   docker logs stream-demo-transcoder
   
   # 檢查 MinIO 連接
   docker exec stream-demo-minio mc ls local/
   
   # 檢查資料庫連接
   docker exec stream-demo-postgresql psql -U postgres -d stream_demo -c "SELECT 1;"
   ```

3. **影片狀態不更新**
   ```bash
   # 檢查資料庫中的影片狀態
   docker exec stream-demo-postgresql psql -U postgres -d stream_demo -c "
   SELECT id, title, status, processing_progress, updated_at 
   FROM videos 
   ORDER BY updated_at DESC;
   "
   ```

### 重啟服務

```bash
# 重啟轉碼排程服務
docker-compose restart transcode-scheduler

# 重啟所有服務
docker-compose restart

# 完全重建
docker-compose down
docker-compose up -d --build
```

## 📝 API 端點

### 轉碼狀態查詢

```http
GET /api/videos/{id}/transcode-status
```

回應範例：

```json
{
  "success": true,
  "data": {
    "id": 1,
    "title": "測試影片",
    "status": "transcoding",
    "processing_progress": 75,
    "original_url": "http://localhost:9000/stream-demo-videos/videos/original/1/test.mp4",
    "mp4_url": "http://localhost:9000/stream-demo-processed/videos/processed/1/999/output.mp4",
    "hls_master_url": "http://localhost:9000/stream-demo-processed/videos/processed/1/999/hls/index.m3u8",
    "thumbnail_url": "http://localhost:9000/stream-demo-processed/videos/processed/1/999/thumbnails/thumb_640x480.jpg"
  }
}
```

## 🔄 部署流程

### 開發環境

1. 啟動基礎服務
2. 構建轉碼排程服務
3. 啟動轉碼排程服務
4. 監控服務狀態

### 生產環境

1. 使用 Docker Swarm 或 Kubernetes
2. 配置健康檢查和自動重啟
3. 設置監控和告警
4. 配置日誌收集

## 📚 相關文檔

- [Docker 指南](./DOCKER_GUIDE.md)
- [MinIO 指南](./MINIO_GUIDE.md)
- [轉碼調試指南](./TRANSCODE_DEBUG.md)
- [API 文檔](./API_DOCS.md) 