# 🗃️ MinIO 整合指南

MinIO 是一個高性能的對象存儲服務器，與 Amazon S3 API 完全兼容，為我們的串流平台提供本地文件存儲服務。

## 🚀 快速開始

### 1. 啟動所有服務
```bash
./docker-manage.sh start
```

### 2. 初始化開發環境（推薦）
```bash
./docker-manage.sh init
```

### 3. 手動初始化 MinIO 桶
```bash
# 在容器內執行
docker exec -it stream-demo-minio /bin/bash
mc alias set local http://localhost:9000 minioadmin minioadmin
mc mb local/stream-demo-videos
mc mb local/stream-demo-processed
```

## 🌐 訪問 MinIO

### Web Console
- **URL**: http://localhost:9001
- **用戶名**: minioadmin
- **密碼**: minioadmin

### API 端點
- **URL**: http://localhost:9000
- **Access Key**: minioadmin
- **Secret Key**: minioadmin

## 📁 桶結構

```
stream-demo-videos/          # 原始影片文件
├── users/1/video1.mp4
├── users/1/video2.mp4
└── users/2/video3.mp4

stream-demo-processed/       # 處理後的影片文件
├── users/1/video1/
│   ├── index.m3u8
│   ├── 720p.m3u8
│   ├── 480p.m3u8
│   └── thumbnails/
└── users/2/video3/
    ├── index.m3u8
    └── 360p.m3u8
```

## 🔧 配置說明

### 後端配置 (config.local.yaml)
```yaml
storage:
  type: "s3"
  s3:
    region: "us-east-1"
    bucket: "stream-demo-videos"
    access_key: "minioadmin"
    secret_key: "minioadmin"
    endpoint: "http://localhost:9000"
    cdn_domain: "http://localhost:9000"
```

### Docker Compose
```yaml
minio:
  image: minio/minio:latest
  container_name: stream-demo-minio
  ports:
    - "9000:9000"   # API
    - "9001:9001"   # Console
  environment:
    MINIO_ROOT_USER: minioadmin
    MINIO_ROOT_PASSWORD: minioadmin
  command: server /data --console-address ":9001"
```

## 🛠️ 開發工具

### MinIO Client (mc)
```bash
# 安裝 mc client
curl -fsSL https://dl.min.io/client/mc/release/darwin-amd64/mc -o /usr/local/bin/mc
chmod +x /usr/local/bin/mc

# 配置別名
mc alias set local http://localhost:9000 minioadmin minioadmin

# 基本操作
mc ls local/                           # 列出桶
mc ls local/stream-demo-videos/        # 列出桶內容
mc cp video.mp4 local/stream-demo-videos/users/1/  # 上傳文件
mc rm local/stream-demo-videos/users/1/video.mp4   # 刪除文件
```

### 使用 AWS CLI (兼容模式)
```bash
# 配置 AWS CLI
aws configure set aws_access_key_id minioadmin
aws configure set aws_secret_access_key minioadmin
aws configure set default.region us-east-1

# 使用 MinIO 端點
aws --endpoint-url http://localhost:9000 s3 ls
aws --endpoint-url http://localhost:9000 s3 ls s3://stream-demo-videos/
```

## 🔍 監控和調試

### 健康檢查
```bash
# API 健康檢查
curl http://localhost:9000/minio/health/live

# 服務狀態
./docker-manage.sh status

# 查看日誌
./docker-manage.sh logs minio
```

### 常見問題

#### 1. 桶不存在錯誤
```bash
# 手動創建桶
mc mb local/stream-demo-videos
```

#### 2. 權限錯誤 (403)
```bash
# 檢查憑證配置
mc admin info local
```

#### 3. 連接錯誤
```bash
# 檢查服務狀態
docker ps | grep minio
```

## 🔄 生產環境切換

要切換到 AWS S3，只需更新配置：

```yaml
# config.production.yaml
storage:
  s3:
    region: "ap-northeast-1"
    bucket: "your-production-bucket"
    access_key: "AKIA..."
    secret_key: "..."
    endpoint: ""              # 留空使用 AWS
    cdn_domain: "https://..."  # CloudFront URL
```

## 📊 性能提示

### 1. 多部分上傳
大文件會自動使用多部分上傳，提高上傳速度。

### 2. CDN 配置
在生產環境中，建議使用 CloudFront 或其他 CDN 服務。

### 3. 桶策略
為公開內容設置適當的桶策略：
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {"AWS": ["*"]},
      "Action": ["s3:GetObject"],
      "Resource": ["arn:aws:s3:::stream-demo-videos/*"]
    }
  ]
}
```

## 🔐 安全考慮

### 開發環境
- 使用默認憑證 (minioadmin/minioadmin)
- 允許所有來源訪問
- 適合本地開發和測試

### 生產環境
- 使用強密碼和 IAM 權限
- 配置適當的桶策略
- 啟用 HTTPS 和訪問日誌
- 定期備份重要數據

## 🎯 總結

MinIO 為我們提供了：
- ✅ **S3 兼容性**: 無需修改代碼
- ✅ **本地開發**: 無需 AWS 帳號
- ✅ **快速部署**: Docker 一鍵啟動
- ✅ **完整功能**: 支援所有 S3 功能
- ✅ **易於切換**: 生產環境無縫遷移

現在你可以在本地環境中享受完整的對象存儲功能！🎉 