# 📁 專案結構說明

## 🏗️ 整體架構

```
stream-demo/
├── 📁 backend/           # 後端應用 (Go)
├── 📁 frontend/          # 前端應用 (Vue 3)
├── 📁 docker/            # Docker 相關配置
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
- **`dev.sh`** - 開發環境快速啟動腳本，一鍵啟動完整開發環境
- **`docker/docker-manage.sh`** - 完整的 Docker 服務管理腳本

### 🐳 Docker 配置
- **`docker/docker-compose.yml`** - 生產模式配置 (包含前後端容器)
- **`docker/docker-compose.dev.yml`** - 開發模式配置 (只包含周邊服務)
- **`docker/nginx/`** - Nginx 反向代理配置
- **`docker/*/`** - 各服務的 Docker 配置

### 📚 文檔
- **`QUICKSTART.md`** - 快速啟動指南 (推薦新用戶)
- **`DEVELOPMENT.md`** - 開發模式詳細指南
- **`README.md`** - 專案完整說明
- **`PROJECT_STRUCTURE.md`** - 專案結構說明

## 🚀 使用流程

### 開發模式 (推薦)
```bash
# 一鍵啟動周邊服務
./dev.sh start

# 在 IDE 中啟動前後端
cd backend && go run main.go
cd frontend && npm run dev
```

### 生產模式
```bash
./manage.sh start        # 啟動所有服務
```

## 🔧 腳本功能對比

| 腳本 | 功能 | 適用場景 |
|------|------|----------|
| **`dev.sh`** | 一鍵啟動周邊服務 | 日常開發 |
| **`manage.sh`** | Docker 服務管理 | 服務管理、部署 |
| **`docker/docker-manage.sh`** | 完整 Docker 管理 | 高級用戶 |

## 📁 目錄職責

### `backend/` - 後端應用
- Go 1.24.3 + Gin 框架
- 提供 RESTful API 和 WebSocket 服務
- 支援 PostgreSQL 和 MySQL 資料庫

### `frontend/` - 前端應用
- Vue 3 + TypeScript + Vite
- Element Plus UI 框架
- 支援熱重載開發

### `docker/` - Docker 配置
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
- `dev.sh` 一鍵啟動周邊服務
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