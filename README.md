# 🎬 串流平台專案

## 📋 專案概述

現代化全棧串流平台，提供影片上傳、自動轉碼、直播和用戶管理功能。採用 **PostgreSQL + Redis 混合架構**，整合 **MinIO 對象存儲** 和 **FFmpeg 本地轉碼**。

### 🎯 核心特色
- ✅ **混合架構**: PostgreSQL 主資料庫 + Redis 緩存與訊息
- ✅ **智能轉碼**: 背景服務自動生成多品質 HLS 和 MP4
- ✅ **雙桶存儲**: 原始檔案與轉碼後檔案分離
- ✅ **直播系統**: RTMP 推流 + HLS 播放 + 低延遲 + 自動化轉換
- ✅ **即時通信**: WebSocket + Redis Pub/Sub
- ✅ **現代前端**: Vue 3 + TypeScript + Element Plus + hls.js
- ✅ **完整 Docker**: 一鍵啟動開發環境
- ✅ **模組化架構**: 依賴注入 + 統一路由管理
- ✅ **自動化推流**: RTMP 推流自動觸發 HLS 轉換

## 🏗️ 技術架構

### 整體架構
```
前端 (Vue 3) → 後端 (Go/Gin) → 資料庫 (PostgreSQL + Redis) → 存儲 (MinIO) → 轉碼 (FFmpeg)
```

### 直播架構
```
OBS/推流軟體 → nginx-rtmp (1935) → on_publish 事件 → stream-puller (8083) → FFmpeg 轉換 → HLS 文件 → 前端播放器 (hls.js)
```

### 技術棧
- **前端**: Vue 3, TypeScript, Element Plus, hls.js
- **後端**: Go 1.24.3, Gin, GORM, JWT, 依賴注入
- **資料庫**: PostgreSQL 15, Redis 7, MySQL 8.0
- **存儲**: MinIO (S3 兼容)
- **轉碼**: FFmpeg 6.0.1
- **直播**: nginx-rtmp, stream-puller
- **容器**: Docker & Docker Compose

## 🚀 快速開始

### 1. 啟動周邊服務
```bash
./docker-manage.sh start
```

### 2. 初始化服務
```bash
./docker-manage.sh init      # 初始化 MinIO 桶
./docker-manage.sh init-live # 初始化直播服務
```

### 3. 啟動應用
```bash
# 後端 (Go)
cd backend && go run main.go

# 前端 (Vue)
cd frontend && npm run dev
```

### 4. 訪問應用
- **前端**: http://localhost:5173
- **後端 API**: http://localhost:8080
- **MinIO Console**: http://localhost:9001 (minioadmin/minioadmin)
- **直播流服務**: http://localhost:8083
- **RTMP 推流**: rtmp://localhost:1935/live
- **HLS 播放**: http://localhost:8083/[stream_key]/index.m3u8

## 📺 直播間使用

#### 生命週期概念
- **開始/結束直播** = 狀態切換（在同一直播間內循環）
- **創建/關閉直播間** = 生命週期流程（創建→關閉）

#### 狀態流程
```
created → live → ended → live → ended → ... (循環)
    ↓
closed (完全刪除)
```

### 基本操作
1. **創建直播間**: 登入後點擊"創建直播間"，填寫標題和描述
2. **開始直播**: 使用推流密鑰在 OBS 等軟體中推流，點擊"開始直播"
3. **結束直播**: 點擊"結束直播"，直播間保留，可重新開始
4. **重新開始**: 在已結束的直播間中點擊"重新開始直播"
5. **加入直播間**: 瀏覽列表，點擊加入感興趣的直播間
6. **離開直播間**: 觀眾可主動離開，自動跳轉回列表
7. **關閉直播間**: 只有創建者可關閉，完全刪除數據

## 🎬 OBS 推流設置

### 1. 獲取推流資訊
1. 登入前端: http://localhost:5173
2. 創建直播間或進入現有直播間
3. 點擊"串流資訊"按鈕
4. 複製以下資訊：
   - **串流金鑰**: `stream_xxxxxxxx`
   - **RTMP 推流地址**: `rtmp://localhost:1935/live/stream_xxxxxxxx`

### 2. OBS 設置
1. **打開 OBS Studio**
2. **設置 → 串流**
3. **服務**: 選擇"自訂"
4. **伺服器**: 填入 `rtmp://localhost:1935/live`
5. **串流金鑰**: 填入您的串流金鑰（如：`stream_xxxxxxxx`）

### 3. 開始推流
1. 在 OBS 中點擊"開始串流"
2. 回到前端直播間，點擊"開始直播"
3. 系統會自動：
   - NGINX RTMP 接收推流
   - 觸發 `on_publish` 事件
   - `stream-puller` 自動啟動 FFmpeg 轉換
   - 生成 HLS 文件
   - 前端 `hls.js` 自動播放（支援自動重試）
4. 等待幾秒鐘，直播畫面應該會出現在前端播放器中

### 4. 其他推流軟體
- **Streamlabs OBS**: 設置方式相同
- **XSplit**: 設置方式相同
- **手機 App**: 支援 RTMP 推流的 App 都可以使用

## 🎬 影片管理

1. **上傳影片**: 選擇檔案，系統自動轉碼
2. **多品質播放**: 自動生成 720p, 480p, 360p HLS 串流
3. **縮圖生成**: 自動提取影片縮圖
4. **播放統計**: 記錄播放次數和時長

## 🔧 開發調試

```bash
# 查看服務狀態
./docker-manage.sh status

# 查看日誌
./docker-manage.sh logs [service]

# 管理直播流服務
./docker-manage.sh stream-puller start
./docker-manage.sh stream-puller status
./docker-manage.sh stream-puller test

# 查看直播狀態
./docker-manage.sh live-status

# 運行測試
./docker-manage.sh test

# 測試 RTMP 推流
./test_rtmp_complete.sh
```

### 自動化流程驗證
```bash
# 1. 檢查 RTMP 推流狀態
curl http://localhost:1935/stat

# 2. 檢查 HLS 文件生成
curl http://localhost:8083/[stream_key]/index.m3u8

# 3. 檢查 stream-puller 日誌
docker-compose logs stream-puller --tail=20
```

## 📊 功能完成度

### 高優先級
- [x] **直播間基礎功能**: 創建、加入、開始/結束直播 ✅
- [x] **直播間聊天系統**: 實時聊天、用戶加入/離開通知 ✅
- [x] **角色權限管理**: 創建者/觀眾權限區分 ✅
- [x] **實時通知系統**: WebSocket 實時通知直播狀態變化 ✅
- [x] **統一踢出功能**: 關閉直播間時自動踢出所有用戶 ✅
- [x] **離開直播間功能**: 用戶離開時自動跳轉回列表 ✅
- [x] **直播間持久化**: 結束直播時保留直播間，只有關閉才刪除 ✅
- [x] **直播間生命週期**: 開始/結束直播為狀態切換，創建/關閉為生命週期 ✅
- [x] **重新開始直播**: 已結束的直播間可以重新開始 ✅
- [x] **RTMP 推流支援**: nginx-rtmp 接收推流，stream-puller 轉換 HLS ✅
- [x] **自動化推流處理**: RTMP 推流自動觸發 HLS 轉換 ✅
- [x] **前端 HLS 播放**: hls.js 整合，支援自動重試和低延遲 ✅
- [x] **服務管理整合**: docker-manage.sh 統一管理所有服務 ✅
- [ ] **角色 API 修復**: 解決角色判斷 API 404 問題

### 中優先級
- [x] **影片上傳**: 支援多格式上傳 ✅
- [x] **自動轉碼**: 背景服務自動處理 ✅
- [x] **多品質播放**: HLS 自適應串流 ✅
- [x] **用戶認證**: JWT 登入註冊 ✅
- [x] **檔案管理**: 影片列表和刪除 ✅
- [ ] **播放統計**: 觀看次數和時長統計
- [ ] **搜尋功能**: 影片標題和標籤搜尋

### 低優先級
- [ ] **表情系統**: 聊天表情和禮物
- [ ] **錄製功能**: 直播錄製和回放
- [ ] **CDN 整合**: 外部 CDN 支援
- [ ] **多語言**: 國際化支援

## 🐛 已知問題

- **角色 API 404**: 偶爾出現角色判斷 API 404 錯誤，需要重啟後端
- **HLS 文件訪問**: 已修復 stream-puller 路由配置，HLS 文件現在可以正常訪問

## ✅ 最近修復

- **自動化推流**: RTMP 推流現在會自動觸發 HLS 轉換
- **前端播放**: 整合 hls.js，支援自動重試和低延遲播放
- **服務管理**: 統一整合到 docker-manage.sh，一鍵管理所有服務
- **HLS 路由**: 修復 stream-puller 的 HLS 文件服務路由

## 📝 開發筆記

- 使用 `docker-manage.sh` 統一管理所有服務
- 直播間狀態通過 Redis 管理，確保實時性
- WebSocket 用於實時通知和聊天功能
- 影片轉碼使用 FFmpeg 背景服務處理
- RTMP 推流通過 nginx-rtmp 接收，stream-puller 轉換為 HLS
- 前端使用 hls.js 播放 HLS 流，支援自動重試和低延遲
- 自動化流程：RTMP 推流 → on_publish 事件 → FFmpeg 轉換 → HLS 生成 → 前端播放