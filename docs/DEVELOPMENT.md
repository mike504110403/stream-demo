# 🛠️ 開發指南

## 📋 概述

本指南詳細說明如何設置和維護串流平台的開發環境，包含環境配置、調試技巧、測試方法等。

## 🚀 開發環境設置

### 前置需求

#### 必需軟體
- **Docker Desktop** - 容器化環境
- **Go 1.24+** - 後端開發
- **Node.js 18+** - 前端開發
- **VS Code** - 推薦的 IDE

#### 可選軟體
- **Postman** - API 測試
- **DBeaver** - 資料庫管理
- **Redis Desktop Manager** - Redis 管理

### 環境檢查

```bash
# 檢查 Docker
docker --version
docker-compose --version

# 檢查 Go
go version

# 檢查 Node.js
node --version
npm --version
```

## 🔧 開發工作流程

### 1. 啟動開發環境

#### 方式一：F5 一鍵啟動（推薦）
1. 確保 Docker Desktop 已啟動
2. 在 VS Code 中按 `F5` 或 `Fn+F5`
3. 選擇 `🚀 F5 一鍵啟動 (推薦)`

#### 方式二：命令行啟動
```bash
# 啟動周邊服務
./deploy/scripts/docker-manage.sh start

# 在 IDE 中啟動前後端
# 後端: 使用 launch.json 配置
# 前端: npm run dev
```

### 2. 開發模式配置

#### 後端配置
- **配置文件**: `services/api/config/config.local.yaml`
- **環境變數**: `services/api/.env`
- **支援資料庫**: PostgreSQL、MySQL
- **熱重載**: 支援，修改代碼後自動重啟

#### 前端配置
- **開發服務器**: Vite
- **端口**: 5173
- **熱重載**: 支援，修改代碼後自動更新
- **API 基礎 URL**: http://localhost:8080/api

### 3. 服務訪問地址

| 服務 | 地址 | 說明 |
|------|------|------|
| **統一入口** | http://localhost:8084 | 主要應用入口 |
| **前端 (IDE)** | http://localhost:5173 | Vue 開發服務器 |
| **後端 (IDE)** | http://localhost:8080 | Go API 服務器 |
| **MinIO Console** | http://localhost:9001 | 對象存儲管理 |
| **HLS 播放** | http://localhost:8083/[stream_name]/index.m3u8 | 直播串流播放 |
| **RTMP 推流** | rtmp://localhost:1935/live | 直播推流地址 |

## 🐛 調試技巧

### 後端調試

#### 使用 VS Code 調試
1. 在 VS Code 中設置斷點
2. 使用 `🧪 後端除錯模式 (PostgreSQL)` 配置
3. 按 `F5` 開始調試

#### 日誌調試
```bash
# 查看後端日誌
./deploy/scripts/docker-manage.sh logs api

# 查看特定服務日誌
./deploy/scripts/docker-manage.sh logs postgresql
./deploy/scripts/docker-manage.sh logs redis
```

#### 資料庫調試
```bash
# 連接 PostgreSQL
psql -h localhost -p 5432 -U stream_user -d stream_demo

# 連接 MySQL
mysql -h localhost -P 3306 -u stream_user -p stream_demo

# 使用 Docker 連接
docker exec -it stream-demo-postgresql psql -U stream_user -d stream_demo
docker exec -it stream-demo-mysql mysql -u stream_user -p stream_demo
```

### 前端調試

#### 瀏覽器開發者工具
- **Vue DevTools**: 安裝 Vue DevTools 擴展
- **Network**: 檢查 API 請求
- **Console**: 查看錯誤和日誌

#### 熱重載調試
- 修改 Vue 組件後自動更新
- 修改 TypeScript 代碼後自動編譯
- 修改樣式後即時預覽

### 直播調試

#### RTMP 推流測試
```bash
# 使用 FFmpeg 測試推流
ffmpeg -re -f lavfi -i testsrc2 -f lavfi -i sine=frequency=1000:sample_rate=44100 -c:v libx264 -c:a aac -f flv rtmp://localhost:1935/live/test

# 檢查推流狀態
curl http://localhost:1935/stat
```

#### HLS 播放測試
```bash
# 檢查 HLS 文件
curl http://localhost:8083/test/index.m3u8

# 使用 VLC 播放
vlc http://localhost:8083/test/index.m3u8
```

## 🧪 測試

### 後端測試

#### 單元測試
```bash
cd services/api
go test -v ./...
```

#### 整合測試
```bash
cd backend
go test -v -tags=integration ./...
```

#### 測試覆蓋率
```bash
cd backend
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

### 前端測試

#### 單元測試
```bash
cd frontend
npm run test
```

#### 建置測試
```bash
cd frontend
npm run build
```

#### 類型檢查
```bash
cd frontend
npm run type-check
```

### 端到端測試

#### 手動測試流程
1. **用戶註冊/登入**
2. **影片上傳和轉碼**
3. **直播間創建和管理**
4. **直播推流和播放**
5. **聊天功能**

#### 自動化測試（未來規劃）
- 使用 Playwright 或 Cypress
- 包含完整的用戶流程測試
- 整合到 CI/CD 流程

## 🔍 性能調優

### 後端性能

#### 資料庫優化
- 使用索引優化查詢
- 監控慢查詢
- 定期清理無用數據

#### 緩存策略
- Redis 緩存熱門數據
- 使用 GORM 查詢緩存
- 實現合理的緩存失效策略

### 前端性能

#### 打包優化
- 代碼分割
- 懶加載組件
- 圖片壓縮

#### 運行時優化
- 虛擬滾動
- 防抖和節流
- 記憶化組件

## 🚨 常見問題

### 端口衝突
```bash
# 檢查端口使用情況
lsof -i :8080  # 後端端口
lsof -i :5173  # 前端端口
lsof -i :5432  # PostgreSQL 端口

# 殺死佔用端口的進程
kill -9 <PID>
```

### Docker 服務問題
```bash
# 檢查 Docker 狀態
docker ps

# 重啟服務
./deploy/scripts/docker-manage.sh restart

# 清理容器
docker system prune -f
```

### 依賴問題
```bash
# 後端依賴
cd services/api
go mod tidy
go mod download

# 前端依賴
cd services/frontend
rm -rf node_modules package-lock.json
npm install
```

### 資料庫連接問題
```bash
# 檢查資料庫容器
docker ps | grep stream-demo-postgresql
docker ps | grep stream-demo-mysql

# 查看資料庫日誌
./deploy/scripts/docker-manage.sh logs postgresql
./deploy/scripts/docker-manage.sh logs mysql
```

## 📚 開發資源

### 文檔
- [Go 官方文檔](https://golang.org/doc/)
- [Gin 框架文檔](https://gin-gonic.com/docs/)
- [Vue 3 文檔](https://vuejs.org/)
- [TypeScript 文檔](https://www.typescriptlang.org/)

### 工具
- [Postman](https://www.postman.com/) - API 測試
- [DBeaver](https://dbeaver.io/) - 資料庫管理
- [Redis Desktop Manager](https://rdm.dev/) - Redis 管理

### 學習資源
- [Go 最佳實踐](https://github.com/golang/go/wiki/CodeReviewComments)
- [Vue 3 組合式 API](https://vuejs.org/guide/extras/composition-api-faq.html)
- [Docker 最佳實踐](https://docs.docker.com/develop/dev-best-practices/) 