# 📁 專案結構說明

## 🏗️ 整體架構

```
stream-demo/
├── 📁 services/          # 服務目錄
│   ├── 📁 api/           # 後端應用 (Go)
│   ├── 📁 frontend/      # 前端應用 (Vue 3)
│   ├── 📁 puller/        # 串流拉取器
│   ├── 📁 receiver/      # 串流接收器
│   ├── 📁 converter/     # 媒體轉換器
│   └── 📁 gateway/       # API 網關
├── 📁 infrastructure/    # 基礎設施配置
│   ├── 📁 nginx/         # Nginx 配置
│   ├── 📁 postgresql/    # PostgreSQL 配置
│   ├── 📁 mysql/         # MySQL 配置
│   ├── 📁 redis/         # Redis 配置
│   ├── 📁 minio/         # MinIO 配置
│   ├── 📁 ffmpeg/        # FFmpeg 配置
│   ├── docker-compose.yml        # 生產模式配置
│   ├── docker-compose.dev.yml    # 開發模式配置
│   └── docker-manage.sh          # Docker 服務管理腳本
├── 📁 logs/              # 日誌文件 (自動創建)
├── manage.sh              # 簡化管理腳本 (快捷方式)
├── dev.sh                 # 開發環境快速啟動腳本
├── QUICKSTART.md          # 快速啟動指南
├── DEVELOPMENT.md         # 開發模式詳細指南
├── README.md              # 專案完整說明
└── PROJECT_STRUCTURE.md   # 本文件
```

## 🎯 核心文件說明

### 📋 管理腳本
- **`manage.sh`** - 簡化管理腳本，作為 `docker/docker-manage.sh` 的快捷方式
- **`start.sh`** - 開發環境一鍵啟動腳本，智能檢查並啟動開發環境
- **`deploy.sh`** - 生產環境部署腳本，完整容器化部署
- **`docker/docker-manage.sh`** - 完整的 Docker 服務管理腳本

### 🐳 Docker 配置
- **`infrastructure/docker-compose.yml`** - 生產模式配置 (包含前後端容器)
- **`deploy/docker-compose.dev.yml`** - 開發模式配置 (只包含周邊服務)
- **`services/gateway/`** - Nginx 反向代理配置
- **`infrastructure/*/`** - 各服務的 Docker 配置

### 📚 文檔
- **`README.md`** - 專案完整說明 (包含快速開始)
- **`docs/DEVELOPMENT.md`** - 詳細開發指南
- **`docs/DEPLOYMENT.md`** - 生產部署指南
- **`docs/PROJECT_STRUCTURE.md`** - 專案結構說明

## 🚀 使用流程

### 開發模式 (推薦)
```bash
# 一鍵啟動開發環境
./deploy/scripts/docker-manage.sh start

# 在 IDE 中啟動前後端
cd services/api && go run main.go
cd services/frontend && npm run dev
```

### 生產模式
```bash
./deploy/scripts/deploy.sh deploy   # 部署生產環境
```

## 🔧 腳本功能對比

| 腳本 | 功能 | 適用場景 |
|------|------|----------|
| **`start.sh`** | 一鍵啟動開發環境 | 日常開發 |
| **`deploy.sh`** | 生產環境部署 | 生產部署 |
| **`manage.sh`** | Docker 服務管理 | 服務管理 |
| **`docker/docker-manage.sh`** | 完整 Docker 管理 | 高級用戶 |

## 📁 目錄職責

### `services/api/` - 後端應用
- Go 1.24.3 + Gin 框架
- 提供 RESTful API 和 WebSocket 服務
- 支援 PostgreSQL 和 MySQL 資料庫

### `services/frontend/` - 前端應用
- Vue 3 + TypeScript + Vite
- Element Plus UI 框架
- 支援熱重載開發

### `infrastructure/` - 基礎設施配置
- 所有 Docker 相關配置集中管理
- 支援開發模式和生產模式
- 包含各服務的配置和腳本

### `logs/` - 日誌文件
- 自動創建的日誌目錄
- 存放前後端運行日誌
- 已加入 .gitignore

## 🎯 優化特色

### ✅ 結構清晰
- Docker 相關文件統一放在 `docker/` 目錄
- 管理腳本分層設計，滿足不同需求
- 文檔分類明確，便於查找

### ✅ 使用簡潔
- `start.sh` 一鍵啟動開發環境
- `deploy.sh` 一鍵部署生產環境
- `manage.sh` 簡化管理命令
- 路徑引用統一，避免混淆

### ✅ 開發友好
- 專注於周邊服務管理
- 前後端由 IDE 控制，支援熱重載
- 狀態檢查和故障排除

### ✅ 部署靈活
- 開發模式和生產模式分離
- 支援 IDE 開發和容器部署
- 配置可根據需求調整 