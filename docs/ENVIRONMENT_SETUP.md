# 環境變數配置說明

## 📋 概述

本專案採用分層環境變數配置，支援本地開發和生產環境部署。

## 🏗️ 配置架構

```
stream-demo/
├── deploy/env/
│   └── env.example          # Docker 環境變數範例
├── services/api/
│   └── .env.example         # 後端 API 環境變數範例
├── services/frontend/
│   └── .env.example         # 前端環境變數範例
└── .gitignore               # 忽略 .env 文件
```

## 🚀 快速設置

### 1. 後端 API 環境變數

```bash
# 進入後端目錄
cd services/api

# 複製環境變數範例
cp .env.example .env

# 根據需要修改配置
```

### 2. 前端環境變數

```bash
# 進入前端目錄
cd services/frontend

# 複製環境變數範例
cp .env.example .env

# 根據需要修改配置
```

### 3. Docker 環境變數

```bash
# 進入部署目錄
cd deploy/env

# 複製環境變數範例
cp env.example .env

# 根據需要修改配置
```

## 🔧 環境配置

### 本地開發環境

#### 後端 API (.env)
```bash
# 資料庫配置
DATABASE_TYPE=postgresql
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_DB=stream_demo
POSTGRES_USER=stream_user
POSTGRES_PASSWORD=stream_password

# Redis 配置
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# 應用配置
APP_ENV=local
SERVER_HOST=0.0.0.0
SERVER_PORT=8080
GIN_MODE=debug

# 存儲配置
S3_ENDPOINT=http://localhost:9000
S3_ACCESS_KEY=minioadmin
S3_SECRET_KEY=minioadmin
S3_BUCKET=stream-demo-videos

# 網路配置
FRONTEND_HOST=host.docker.internal
FRONTEND_PORT=5173
BACKEND_HOST=host.docker.internal
BACKEND_PORT=8080
```

#### 前端 (.env)
```bash
# 開發環境配置
NODE_ENV=development
VITE_API_BASE_URL=http://localhost:8080/api
VITE_DEV_PORT=5173

# 應用配置
VITE_APP_NAME=Stream Demo
VITE_DEBUG=true

# 直播配置
VITE_RTMP_URL=rtmp://localhost:1935/live
VITE_HLS_BASE_URL=http://localhost:8083
VITE_WS_URL=ws://localhost:8080/ws
```

### 生產環境

#### 後端 API (.env)
```bash
# 資料庫配置
DATABASE_TYPE=postgresql
POSTGRES_HOST=postgresql
POSTGRES_PORT=5432
POSTGRES_DB=stream_demo
POSTGRES_USER=stream_user
POSTGRES_PASSWORD=your_secure_password

# Redis 配置
REDIS_HOST=redis
REDIS_PORT=6379
REDIS_PASSWORD=your_secure_password
REDIS_DB=0

# 應用配置
APP_ENV=production
SERVER_HOST=0.0.0.0
SERVER_PORT=8080
GIN_MODE=release

# 存儲配置
S3_ENDPOINT=http://minio:9000
S3_ACCESS_KEY=your_access_key
S3_SECRET_KEY=your_secret_key
S3_BUCKET=stream-demo-videos

# 網路配置
FRONTEND_HOST=frontend
FRONTEND_PORT=80
BACKEND_HOST=api
BACKEND_PORT=8080
```

#### 前端 (.env)
```bash
# 生產環境配置
NODE_ENV=production
VITE_API_BASE_URL=http://localhost:8084/api
VITE_DEV_PORT=80

# 應用配置
VITE_APP_NAME=Stream Demo
VITE_DEBUG=false

# 直播配置
VITE_RTMP_URL=rtmp://localhost:1935/live
VITE_HLS_BASE_URL=http://localhost:8083
VITE_WS_URL=ws://localhost:8084/ws
```

## 🔒 安全注意事項

### 1. 密碼安全
- 生產環境必須修改所有預設密碼
- 使用強密鑰生成器生成 JWT_SECRET
- 定期更換資料庫密碼

### 2. 網路安全
- 生產環境限制 CORS 來源
- 啟用 HTTPS
- 配置防火牆規則

### 3. 環境變數管理
- 不要將 .env 文件提交到版本控制
- 使用環境變數管理工具（如 Docker Secrets）
- 定期審查環境變數配置

## 🚀 F5 一鍵啟動配置

### launch.json 配置
```json
{
  "name": "🚀 F5 一鍵啟動 (本地環境)",
  "configurations": [
    "🚀 啟動後端 (本地環境)",
    "🎨 啟動前端 (本地環境)"
  ],
  "preLaunchTask": "🎯 智能啟動開發環境"
}
```

### 環境變數注入
- 後端：通過 `envFile` 和 `env` 配置注入
- 前端：通過 `envFile` 和 `env` 配置注入
- 自動連接到 Docker 周邊服務

## 📝 配置優先級

1. **環境變數** (最高優先級)
2. **.env 文件**
3. **預設值** (最低優先級)

## 🔍 故障排除

### 常見問題

1. **環境變數未生效**
   - 檢查 .env 文件是否存在
   - 確認文件格式正確（無空格、引號等）
   - 重新啟動服務

2. **資料庫連接失敗**
   - 檢查資料庫服務是否運行
   - 確認連接參數正確
   - 檢查防火牆設置

3. **前端 API 調用失敗**
   - 確認 VITE_API_BASE_URL 配置正確
   - 檢查後端服務是否運行
   - 確認 CORS 配置

### 調試命令

```bash
# 檢查環境變數
echo $DATABASE_TYPE

# 檢查服務狀態
docker-compose ps

# 查看服務日誌
docker-compose logs [service_name]

# 測試資料庫連接
docker-compose exec postgresql psql -U stream_user -d stream_demo
```

## 📚 相關文檔

- [開發指南](./DEVELOPMENT.md) - 詳細的開發環境設置
- [部署指南](./DEPLOYMENT.md) - 生產環境部署說明
- [配置說明](./CONFIGURATION.md) - 詳細配置選項 