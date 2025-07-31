# 🚀 快速開始指南

## 📋 概述

這是一個完整的串流平台專案，支援影片上傳、直播串流、即時聊天等功能。本指南將幫助您快速啟動開發環境。

## 🎯 一鍵啟動（推薦）

### 開發模式：F5 一鍵啟動（最簡單）
1. **確保 Docker 已啟動**
2. **按 F5 或 Fn+F5**：自動執行 `🚀 F5 一鍵啟動 (推薦)`

這個配置會自動：
- ✅ 檢查並安裝必要的依賴（Node.js、Go、npm 包、Go 模組）
- ✅ 智能檢查並啟動周邊服務（資料庫、Redis、MinIO、Nginx）
- ✅ 啟動前後端服務
- ✅ 提供詳細的狀態報告

### 生產模式：完整容器化部署
```bash
# 啟動所有服務（包含前後端容器）
./cmd/manage.sh start

# 或使用 Docker Compose
cd docker
docker-compose -f docker-compose.yml --project-name stream-demo up -d
```

#### 🎯 互動式選項
腳本會根據情況提供選擇：
- **周邊服務**：跳過/清理重建/強制重建（無快取）
- **後端服務**：自動重啟/跳過
- **前端服務**：自動重啟/跳過

#### 🏷️ 容器和映像檔命名
**容器命名**（簡化，無前綴）：
- `postgres`、`redis`、`minio`
- `nginx-reverse-proxy`、`nginx-rtmp`
- `stream-puller`、`ffmpeg-transcoder`

**映像檔命名**（使用專案前綴）：
- `stream-demo-nginx-reverse-proxy`
- `stream-demo-stream-puller`
- `stream-demo-ffmpeg-transcoder`
- `stream-demo-nginx-rtmp`

**分層結構**：
- `stream-demo` (父群組，使用 `--project-name stream-demo` 強制指定)
  - `postgres`、`redis`、`minio` (子容器)
  - `nginx-reverse-proxy`、`nginx-rtmp` (子容器)
  - `stream-puller`、`ffmpeg-transcoder` (子容器)

### 方式二：VS Code 智能啟動
1. **確保 Docker 已啟動**
2. **使用以下任一配置**：
   - `🎯 智能啟動完整環境 (PostgreSQL)` - 推薦
   - `🎯 智能啟動完整環境 (MySQL)`
   - `🎯 智能啟動完整環境 (默認)`

這些配置會自動：
- ✅ 檢查並啟動周邊服務（資料庫、Redis、MinIO、Nginx）
- ✅ 啟動前後端服務
- ✅ 提供詳細的狀態報告

### 方式二：命令行智能啟動
```bash
# 智能啟動（自動檢查並啟動所有服務）
./cmd/smart-dev.sh
```

## 🔧 手動啟動

### 1. 啟動周邊服務
```bash
# 啟動 Docker 周邊服務
./cmd/dev.sh start

# 或使用管理腳本
./cmd/manage.sh start-dev
```

### 2. 啟動前後端
在 VS Code 中使用以下配置：

#### **後端啟動選項**
- `🚀 啟動後端 (PostgreSQL)` - 使用 PostgreSQL 資料庫
- `🚀 啟動後端 (MySQL)` - 使用 MySQL 資料庫
- `🚀 啟動後端 (默認配置)` - 使用配置文件默認設定

#### **前端啟動**
- `🎨 啟動前端 (開發模式)` - 啟動 Vite 開發服務器

#### **一鍵啟動前後端**
- `🎬 一鍵啟動前後端 (PostgreSQL)`
- `🎬 一鍵啟動前後端 (MySQL)`
- `🎬 一鍵啟動前後端 (默認)`

## 📊 訪問地址

啟動成功後，您可以訪問以下地址：

| 服務 | 地址 | 說明 |
|------|------|------|
| **統一入口** | http://localhost:8084 | 主要應用入口 |
| **前端 (IDE)** | http://localhost:5173 | Vue 開發服務器 |
| **後端 (IDE)** | http://localhost:8080 | Go API 服務器 |
| **MinIO Console** | http://localhost:9001 | 對象存儲管理 |
| **HLS 播放** | http://localhost:8083/[stream_name]/index.m3u8 | 直播串流播放 |
| **RTMP 推流** | rtmp://localhost:1935/live | 直播推流地址 |

## 🛠️ 開發工具

### VS Code 任務
- `🎯 智能啟動開發環境` - 自動檢查並啟動所有服務
- `檢查開發環境` - 檢查服務狀態

## 🔧 環境配置說明

### 開發模式配置
- **nginx-reverse-proxy-dev.conf**: 連接到 `host.docker.internal` (主機服務)
- **Dockerfile.reverse-proxy**: 開發模式專用映像檔
- **前後端**: 由 IDE 啟動，支援熱重載

### 生產模式配置  
- **nginx-reverse-proxy-prod.conf**: 連接到容器內服務 (`frontend:80`, `backend:8080`)
- **Dockerfile.reverse-proxy-prod**: 生產模式專用映像檔
- **前後端**: 容器化部署，完整隔離
- `啟動周邊服務` - 啟動 Docker 服務
- `npm-install-frontend` - 安裝前端依賴
- `go-mod-tidy` - 整理 Go 模組
- `build-frontend` - 構建前端

### 命令行工具
```bash
# 開發環境管理
./cmd/dev.sh start      # 啟動開發環境
./cmd/dev.sh stop       # 停止開發環境
./cmd/dev.sh restart    # 重啟開發環境
./cmd/dev.sh status     # 檢查狀態
./cmd/dev.sh logs       # 查看日誌

# Docker 服務管理
./cmd/manage.sh start-dev    # 啟動開發模式服務
./cmd/manage.sh stop         # 停止所有服務
./cmd/manage.sh dev-status   # 檢查開發服務狀態
./cmd/manage.sh dev-logs     # 查看開發服務日誌
```

## 🔍 故障排除

### 常見問題

#### 1. 端口被佔用
```bash
# 檢查端口使用情況
lsof -i :8080  # 檢查後端端口
lsof -i :5173  # 檢查前端端口
lsof -i :5432  # 檢查 PostgreSQL 端口
```

#### 2. Docker 服務未啟動
```bash
# 檢查 Docker 狀態
docker ps

# 重啟 Docker 服務
./cmd/manage.sh stop
./cmd/manage.sh start-dev
```

#### 3. 資料庫連接失敗
```bash
# 檢查資料庫容器狀態
docker ps | grep postgres
docker ps | grep mysql

# 查看資料庫日誌
./cmd/manage.sh dev-logs postgresql
./cmd/manage.sh dev-logs mysql
```

#### 4. 前端依賴問題
```bash
# 重新安裝前端依賴
cd frontend
rm -rf node_modules package-lock.json
npm install
```

#### 5. 後端依賴問題
```bash
# 重新整理 Go 模組
cd backend
go mod tidy
go mod download
```

### 日誌查看
```bash
# 查看所有服務日誌
./cmd/dev.sh logs

# 查看特定服務日誌
./cmd/dev.sh logs postgresql
./cmd/dev.sh logs redis
./cmd/dev.sh logs minio
```

## 🎯 開發流程

### 日常開發
1. **啟動開發環境**：按 F5 或使用智能啟動配置
2. **修改代碼**：前後端支援熱重載
3. **測試功能**：訪問 http://localhost:8084
4. **停止服務**：在 VS Code 中停止或使用 `./cmd/dev.sh stop`

### 除錯開發
1. **使用除錯配置**：
   - `🧪 後端除錯模式 (PostgreSQL)`
   - `🧪 後端除錯模式 (MySQL)`
2. **設置斷點**：在 VS Code 中設置斷點
3. **開始除錯**：按 F5 開始除錯

### 測試
1. **運行測試**：
   - `🧪 運行後端測試 (PostgreSQL)`
   - `🧪 運行後端測試 (MySQL)`
2. **查看測試結果**：在 VS Code 測試面板中查看

## 📁 專案結構

```
stream-demo/
├── cmd/                    # 管理腳本
│   ├── dev.sh             # 開發環境管理
│   ├── manage.sh          # Docker 服務管理
│   ├── smart-dev.sh       # 智能啟動腳本
│   └── smart-dev-simple.sh # 簡化智能啟動
├── docker/                 # Docker 配置
│   ├── docker-compose.yml      # 生產環境配置
│   ├── docker-compose.dev.yml  # 開發環境配置
│   └── docker-manage.sh        # Docker 管理腳本
├── backend/               # Go 後端
├── frontend/              # Vue 前端
├── docs/                  # 文檔
└── .vscode/              # VS Code 配置
    ├── launch.json        # 啟動配置
    └── tasks.json         # 任務配置
```

## 🔧 配置說明

### 後端配置
- **配置文件**：`backend/config/config.local.yaml`
- **支援的資料庫**：PostgreSQL、MySQL
- **環境變數**：可覆蓋配置文件設定
- **命令行參數**：最高優先級

### 前端配置
- **開發服務器**：Vite
- **端口**：5173
- **API 基礎 URL**：http://localhost:8080/api

### Docker 配置
- **開發模式**：只啟動周邊服務，前後端由 IDE 啟動
- **生產模式**：啟動所有服務
- **網路**：使用 Docker 內部網路通信

## 🎉 成功啟動檢查清單

- [ ] Docker 已啟動
- [ ] 周邊服務運行中（PostgreSQL、Redis、MinIO、Nginx）
- [ ] 後端服務運行中（http://localhost:8080）
- [ ] 前端服務運行中（http://localhost:5173）
- [ ] 統一入口可訪問（http://localhost:8084）
- [ ] MinIO Console 可訪問（http://localhost:9001）

## 📞 支援

如果遇到問題，請：
1. 檢查本指南的故障排除部分
2. 查看服務日誌
3. 確認 Docker 和相關服務狀態
4. 重新啟動開發環境 