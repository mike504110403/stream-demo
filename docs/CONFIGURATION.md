# ⚙️ 配置說明

## 📋 概述

本文檔詳細說明串流平台的配置選項，包含環境變數、Docker 配置、網路設置等。

## 🗄️ 資料庫配置

### PostgreSQL 配置

| 變數名 | 預設值 | 說明 |
|--------|--------|------|
| `POSTGRES_DB` | `stream_demo` | 資料庫名稱 |
| `POSTGRES_USER` | `stream_user` | 資料庫用戶名 |
| `POSTGRES_PASSWORD` | `stream_password` | 資料庫密碼 |
| `POSTGRES_HOST` | `localhost` | 資料庫主機 |
| `POSTGRES_PORT` | `5432` | 資料庫端口 |

### MySQL 配置

| 變數名 | 預設值 | 說明 |
|--------|--------|------|
| `MYSQL_ROOT_PASSWORD` | `root_password` | Root 用戶密碼 |
| `MYSQL_DATABASE` | `stream_demo` | 資料庫名稱 |
| `MYSQL_USER` | `stream_user` | 資料庫用戶名 |
| `MYSQL_PASSWORD` | `stream_password` | 資料庫密碼 |
| `MYSQL_HOST` | `localhost` | 資料庫主機 |
| `MYSQL_PORT` | `3306` | 資料庫端口 |

### Redis 配置

| 變數名 | 預設值 | 說明 |
|--------|--------|------|
| `REDIS_HOST` | `localhost` | Redis 主機 |
| `REDIS_PORT` | `6379` | Redis 端口 |
| `REDIS_PASSWORD` | `` | Redis 密碼（可選） |
| `REDIS_DB` | `0` | Redis 資料庫編號 |

## 🔧 應用程序配置

### 基本配置

| 變數名 | 預設值 | 說明 |
|--------|--------|------|
| `DATABASE_TYPE` | `postgresql` | 使用的資料庫類型 |
| `APP_ENV` | `local` | 運行環境 |
| `JWT_SECRET` | `your_jwt_secret_key_here` | JWT 簽名密鑰 |
| `SERVER_HOST` | `0.0.0.0` | 服務器監聽地址 |
| `SERVER_PORT` | `8080` | 服務器端口 |

### 安全配置

| 變數名 | 預設值 | 說明 |
|--------|--------|------|
| `CORS_ALLOW_ORIGIN` | `*` | CORS 允許的來源 |
| `CORS_ALLOW_METHODS` | `GET,POST,PUT,DELETE,OPTIONS` | CORS 允許的方法 |
| `CORS_ALLOW_HEADERS` | `Content-Type,Authorization` | CORS 允許的標頭 |
| `RATE_LIMIT_REQUESTS` | `100` | 速率限制請求數 |
| `RATE_LIMIT_WINDOW` | `1m` | 速率限制時間窗口 |

## 📦 存儲配置

### MinIO 配置

| 變數名 | 預設值 | 說明 |
|--------|--------|------|
| `MINIO_ROOT_USER` | `minioadmin` | MinIO 管理員用戶 |
| `MINIO_ROOT_PASSWORD` | `minioadmin` | MinIO 管理員密碼 |
| `MINIO_SERVER_URL` | `http://localhost:9000` | MinIO 服務器 URL |
| `MINIO_BROWSER_REDIRECT_URL` | `http://localhost:9001` | MinIO 瀏覽器重定向 URL |

### S3 兼容配置

| 變數名 | 預設值 | 說明 |
|--------|--------|------|
| `S3_ENDPOINT` | `http://localhost:9000` | S3 端點 |
| `S3_ACCESS_KEY` | `minioadmin` | S3 訪問密鑰 |
| `S3_SECRET_KEY` | `minioadmin` | S3 秘密密鑰 |
| `S3_BUCKET` | `stream-demo-videos` | 原始影片桶 |
| `S3_PROCESSED_BUCKET` | `stream-demo-processed` | 處理後影片桶 |

## 📺 直播配置

### RTMP 配置

| 變數名 | 預設值 | 說明 |
|--------|--------|------|
| `RTMP_PORT` | `1935` | RTMP 推流端口 |
| `RTMP_APP` | `live` | RTMP 應用名稱 |

### HLS 配置

| 變數名 | 預設值 | 說明 |
|--------|--------|------|
| `HLS_PORT` | `8083` | HLS 播放端口 |
| `HLS_PATH` | `/tmp/hls` | HLS 文件路徑 |
| `PUBLIC_STREAMS_PATH` | `/tmp/public_streams` | 公開直播路徑 |

## 🌐 網路配置

### 服務端口

| 服務 | 端口 | 說明 |
|------|------|------|
| 前端 | `5173` | Vue 開發服務器 |
| 後端 | `8080` | Go API 服務器 |
| 統一入口 | `8084` | Nginx 反向代理 |
| MinIO API | `9000` | MinIO 服務器 |
| MinIO Console | `9001` | MinIO 管理界面 |
| RTMP | `1935` | 直播推流 |
| HLS | `8083` | 直播播放 |

### 容器網路

| 網路名稱 | 類型 | 子網 | 說明 |
|----------|------|------|------|
| `stream-demo-network` | bridge | `172.20.0.0/16` | 專案內部網路 |

## 🐳 Docker 配置

### 容器命名規範

所有容器都使用 `stream-demo-` 前綴：

- `stream-demo-postgresql` - PostgreSQL 資料庫
- `stream-demo-mysql` - MySQL 資料庫
- `stream-demo-redis` - Redis 緩存
- `stream-demo-minio` - MinIO 存儲
- `stream-demo-backend` - 後端 API
- `stream-demo-frontend` - 前端應用
- `stream-demo-nginx-reverse-proxy` - Nginx 反向代理
- `stream-demo-nginx-rtmp` - Nginx RTMP
- `stream-demo-stream-puller` - 串流拉取器
- `stream-demo-ffmpeg-transcoder` - FFmpeg 轉碼器

### 資料卷

| 資料卷名稱 | 用途 |
|------------|------|
| `postgres_data` | PostgreSQL 數據 |
| `mysql_data` | MySQL 數據 |
| `redis_data` | Redis 數據 |
| `minio_data` | MinIO 數據 |
| `ffmpeg_temp` | FFmpeg 臨時文件 |
| `public_streams` | 公開直播文件 |
| `hls_streams` | HLS 串流文件 |
| `hls_standard` | 標準 HLS 文件 |

## 📊 監控配置

### 健康檢查

| 變數名 | 預設值 | 說明 |
|--------|--------|------|
| `HEALTH_CHECK_INTERVAL` | `30s` | 健康檢查間隔 |
| `HEALTH_CHECK_TIMEOUT` | `10s` | 健康檢查超時 |
| `HEALTH_CHECK_RETRIES` | `3` | 健康檢查重試次數 |

### 日誌配置

| 變數名 | 預設值 | 說明 |
|--------|--------|------|
| `LOG_LEVEL` | `info` | 日誌級別 |
| `VERBOSE_LOGGING` | `false` | 是否啟用詳細日誌 |
| `SLOW_QUERY_THRESHOLD` | `1000` | 慢查詢閾值（毫秒） |

## 🧪 測試配置

### 測試資料庫

| 變數名 | 預設值 | 說明 |
|--------|--------|------|
| `TEST_DATABASE_TYPE` | `postgresql` | 測試資料庫類型 |
| `TEST_POSTGRES_DB` | `stream_demo_test` | 測試 PostgreSQL 資料庫 |
| `TEST_MYSQL_DB` | `stream_demo_test` | 測試 MySQL 資料庫 |

### 測試 Redis

| 變數名 | 預設值 | 說明 |
|--------|--------|------|
| `TEST_REDIS_DB` | `15` | 測試 Redis 資料庫 |
| `TEST_CACHE_DB` | `14` | 測試緩存資料庫 |
| `TEST_MESSAGING_DB` | `13` | 測試訊息資料庫 |

## 🔒 安全配置

### 生產環境安全檢查清單

- [ ] 修改所有預設密碼
- [ ] 使用強密鑰作為 JWT_SECRET
- [ ] 限制 CORS 來源
- [ ] 啟用 HTTPS
- [ ] 配置防火牆規則
- [ ] 定期更新依賴
- [ ] 啟用日誌監控
- [ ] 配置備份策略

### 密碼強度要求

- **最小長度**: 12 字符
- **包含**: 大小寫字母、數字、特殊字符
- **避免**: 常見密碼、個人信息

## 📝 配置優先級

配置的優先級順序（從高到低）：

1. **環境變數** - 系統環境變數
2. **.env 文件** - 專案環境變數文件
3. **預設值** - 代碼中的預設值

## 🚀 快速配置

### 開發環境

```bash
# 複製配置範例
cp deploy/env/env.example deploy/env/.env

# 啟動開發環境
./deploy/scripts/docker-manage.sh start
```

### 生產環境

```bash
# 複製配置範例
cp deploy/env/env.example deploy/env/.env

# 編輯配置
nano deploy/env/.env

# 部署生產環境
./deploy/scripts/deploy.sh deploy
```

## 🔧 配置驗證

### 檢查配置

```bash
# 檢查環境變數
./deploy/scripts/docker-manage.sh check

# 檢查服務狀態
./deploy/scripts/docker-manage.sh status

# 檢查網路連接
docker network ls | grep stream-demo
```

### 常見問題

1. **端口衝突**: 檢查端口是否被佔用
2. **權限問題**: 確保 Docker 有足夠權限
3. **網路問題**: 檢查防火牆設置
4. **配置錯誤**: 驗證環境變數格式 