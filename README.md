# 🎬 Stream Demo - 串流平台專案

現代化全棧串流平台，提供影片上傳、自動轉碼、直播間管理和公開直播功能。採用微服務架構，完全容器化部署。

## 🚀 快速啟動

### 前置需求
- Docker Desktop
- 8GB+ 可用記憶體
- 端口 1935, 5173, 8080-8085, 9000-9001 可用

### 一鍵啟動
```bash
# 1. 啟動所有服務
cd deploy
./scripts/docker-manage.sh start

# 2. 初始化服務
./scripts/docker-manage.sh init-live

# 3. 檢查狀態
./scripts/docker-manage.sh status
```

### 訪問地址
- **統一入口**: http://localhost:8084
- **前端應用**: http://localhost:5173
- **後端 API**: http://localhost:8080
- **MinIO Console**: http://localhost:9001 (minioadmin/minioadmin)
- **RTMP 推流**: rtmp://localhost:1935/live

## 🏗️ 系統架構

### 微服務組件
- **API 服務** (8080): Go + Gin 後端 API
- **前端服務** (5173): Vue 3 + TypeScript
- **Receiver** (1935): nginx-rtmp RTMP 接收
- **Live-CDN** (8085): HLS 靜態緩存分發
- **Puller** (8083): 外部直播源拉取
- **Converter**: FFmpeg 影片轉碼服務
- **Gateway** (8084): nginx 反向代理

### 基礎設施
- **PostgreSQL** (5432): 主資料庫
- **Redis** (6379): 緩存與訊息隊列
- **MinIO** (9000-9001): S3 兼容對象存儲
- **MySQL** (3306): 備用資料庫

## 📺 功能特色

### ✅ 直播功能
- **RTMP 推流**: 支援 OBS 等推流軟體
- **LL-HLS 播放**: 低延遲直播播放
- **直播間管理**: 創建、開始/結束、聊天室
- **自動轉碼**: RTMP → HLS 即時轉換

### ✅ 影片功能
- **影片上傳**: 多格式支援
- **自動轉碼**: 多品質 HLS + MP4 輸出
- **縮圖生成**: 自動提取影片縮圖
- **播放統計**: 觀看次數和時長記錄

### ✅ 公開直播
- **外部流拉取**: 支援 HLS/RTMP/RTSP 源
- **分類管理**: 測試、太空、新聞、體育等
- **狀態監控**: 實時監控直播源狀態

## 🎮 使用指南

### 直播推流設置
1. **創建直播間**: 登入後點擊"創建直播間"
2. **獲取推流資訊**: 複製串流金鑰和 RTMP 地址
3. **OBS 設置**:
   - 服務: 自訂
   - 伺服器: `rtmp://localhost:1935/live`
   - 串流金鑰: `stream_xxxxxxxx`
4. **開始推流**: OBS 開始串流，前端點擊"開始直播"

### 影片上傳
1. 選擇影片檔案上傳
2. 系統自動轉碼處理
3. 轉碼完成後可播放多品質版本

## 🔧 管理命令

```bash
# 服務管理
./scripts/docker-manage.sh start      # 啟動所有服務
./scripts/docker-manage.sh stop       # 停止所有服務
./scripts/docker-manage.sh restart    # 重啟所有服務
./scripts/docker-manage.sh status     # 查看服務狀態
./scripts/docker-manage.sh logs       # 查看服務日誌

# 初始化
./scripts/docker-manage.sh init       # 初始化 MinIO
./scripts/docker-manage.sh init-live  # 初始化直播服務

# 診斷
./scripts/docker-manage.sh diagnose   # 系統診斷
```

## 🛠️ 技術棧

### 後端
- **語言**: Go 1.24
- **框架**: Gin 1.10
- **ORM**: GORM 1.30
- **認證**: JWT
- **直播**: nginx-rtmp + FFmpeg

### 前端
- **框架**: Vue 3.4 + TypeScript 5.3
- **UI**: Element Plus 2.5
- **播放器**: hls.js 1.6
- **狀態管理**: Pinia 2.1

### 基礎設施
- **資料庫**: PostgreSQL 15, Redis 7, MySQL 8.0
- **存儲**: MinIO (S3 兼容)
- **容器**: Docker & Docker Compose
- **代理**: nginx

## 📊 服務狀態

檢查所有服務是否正常運行：
```bash
./scripts/docker-manage.sh status
```

預期看到所有服務狀態為 "Up" 且健康檢查通過。

## 🐛 故障排除

### 常見問題
1. **端口衝突**: 確保所需端口未被占用
2. **記憶體不足**: 確保至少 8GB 可用記憶體
3. **Docker 未啟動**: 先啟動 Docker Desktop

### 診斷工具
```bash
./scripts/docker-manage.sh diagnose   # 完整診斷
./scripts/docker-manage.sh logs api   # 查看特定服務日誌
```

## 📚 文檔

- [開發指南](./docs/DEVELOPMENT.md) - 詳細開發環境設置
- [部署指南](./docs/DEPLOYMENT.md) - 生產環境部署
- [配置說明](./docs/CONFIGURATION.md) - 環境變數配置

## 🎯 專案狀態

- ✅ **核心功能完成**: 直播、影片、用戶管理
- ✅ **完全容器化**: Docker Compose 一鍵部署
- ✅ **生產就緒**: 健康檢查、日誌、監控
- 🔄 **持續優化**: 性能調優、功能擴展

---

**開發團隊**: 現代化串流平台解決方案  
**技術支援**: 完整的微服務架構與容器化部署