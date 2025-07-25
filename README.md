# 🎬 串流平台專案完整技術文檔

## 📋 專案概述

這是一個現代化的全棧串流平台專案，提供完整的影片上傳、自動轉檔、直播、用戶管理和支付功能。專案採用 **PostgreSQL + Redis 混合架構**，結合 **MinIO 對象存儲** 和 **FFmpeg 本地轉碼服務**，實現高效能的影片處理和播放體驗。

### 🎯 專案核心特色
- ✅ **混合架構設計**：PostgreSQL 作為主資料庫，Redis 作為緩存和訊息佇列
- ✅ **本地化存儲與轉碼**：整合 MinIO S3 兼容對象存儲和 FFmpeg 本地轉碼服務
- ✅ **智能自動轉碼系統**：背景服務自動處理上傳影片，生成多品質 HLS 串流和 MP4 播放版本
- ✅ **雙桶存儲架構**：原始檔案存儲於 `stream-demo-videos`，轉碼後檔案存儲於 `stream-demo-processed`
- ✅ **實時通信**：Redis Pub/Sub + WebSocket 即時聊天和直播互動
- ✅ **現代化前端**：Vue 3 + TypeScript + Element Plus + hls.js
- ✅ **智能播放體驗**：自動品質切換、垂直影片比例保持、即時載入
- ✅ **完整 Docker 化**：包含 FFmpeg 轉碼容器的完整開發環境

## 🏗️ 系統架構

### 整體架構圖
```mermaid
graph TB
    subgraph "前端層"
        FE[Vue 3 前端應用]
        FE --> |HTTP/HTTPS| LB[負載均衡器]
        FE --> |WebSocket| WS[WebSocket 連接]
    end
    
    subgraph "後端服務層"
        LB --> BE[Go 後端服務]
        BE --> |JWT 認證| AUTH[認證中間件]
        BE --> WS
        
        subgraph "核心服務"
            US[用戶服務]
            VS[影片服務]
            LS[直播服務]
            PS[支付服務]
            TW[轉碼背景服務]
        end
        
        BE --> US
        BE --> VS
        BE --> LS
        BE --> PS
        BE --> TW
    end
    
    subgraph "資料儲存層"
        subgraph "PostgreSQL 資料庫"
            PG[(PostgreSQL 主資料庫)]
            PGS[(PostgreSQL 從資料庫)]
        end
        
        subgraph "Redis 緩存與訊息"
            REDIS_CACHE[Redis DB 1<br/>緩存系統]
            REDIS_MSG[Redis DB 2<br/>Pub/Sub 訊息]
        end
        
        subgraph "MinIO 對象存儲"
            MINIO_ORIG[(MinIO 原始存儲<br/>stream-demo-videos)]
            MINIO_PROC[(MinIO 處理存儲<br/>stream-demo-processed)]
        end
    end
    
    subgraph "轉碼服務層"
        FFMPEG[FFmpeg 轉碼容器]
        FFMPEG --> |下載原始| MINIO_ORIG
        FFMPEG --> |上傳處理| MINIO_PROC
        FFMPEG --> |生成| HLS[多品質 HLS 串流]
        FFMPEG --> |生成| MP4[MP4 網頁播放版本]
        FFMPEG --> |生成| THUMB[縮圖和時間軸預覽]
    end
    
    subgraph "外部服務"
        EMAIL[郵件服務]
        PAYMENT[第三方支付]
        
        US --> EMAIL
        PS --> PAYMENT
    end
```

### 技術棧詳細說明

#### 前端技術棧
```mermaid
graph LR
    subgraph "前端技術棧"
        VUE[Vue 3.x] --> TS[TypeScript]
        VUE --> EP[Element Plus UI]
        VUE --> VR[Vue Router 4]
        VUE --> PINIA[Pinia 狀態管理]
        VUE --> AXIOS[Axios HTTP 客戶端]
        VUE --> VITE[Vite 構建工具]
        VUE --> HLSJS[hls.js 串流播放]
        
        TS --> TYPES[類型定義]
        EP --> COMPONENTS[UI 組件]
        VR --> GUARDS[路由守衛]
        PINIA --> STORES[狀態倉庫]
        AXIOS --> API[API 服務]
        VITE --> BUILD[構建配置]
        HLSJS --> STREAMING[HLS 串流播放]
    end
```

#### 後端技術棧
```mermaid
graph LR
    subgraph "後端核心技術"
        GO[Go 1.24.3] --> GIN[Gin Web 框架]
        GO --> GORM[GORM ORM]
        GO --> JWT[JWT 認證]
        GO --> WS[Gorilla WebSocket]
        GO --> PQ[lib/pq PostgreSQL 驅動]
        GO --> REDIS[Redis 客戶端]
        
        GIN --> MW[中間件]
        GORM --> REPO[Repository 層]
        JWT --> AUTH[身份認證]
        WS --> CHAT[即時聊天]
        PQ --> DB[資料庫連接]
        REDIS --> CACHE[緩存系統]
    end
    
    subgraph "存儲與轉碼技術"
        MINIO[MinIO S3 API] --> STORAGE[對象存儲]
        FFMPEG[FFmpeg 6.0.1] --> TRANSCODE[影片轉碼]
        DOCKER[Docker 容器] --> FFMPEG
        
        STORAGE --> BUCKET[雙桶存儲]
        TRANSCODE --> HLS[HLS 串流]
        TRANSCODE --> MP4[MP4 轉換]
        TRANSCODE --> THUMB[縮圖生成]
    end
```

## ⚡ 技術架構評估與實現

### PostgreSQL + Redis 混合架構設計

本專案採用 **PostgreSQL + Redis 混合架構**，結合兩者優勢實現最佳化的效能：

#### Redis 緩存與訊息系統的性能特點
```mermaid
graph LR
    subgraph "Redis 緩存系統"
        REDIS_CACHE[Redis Cache]
        REDIS_CACHE --> MEMORY[記憶體存儲]
        REDIS_CACHE --> ATOMIC[原子操作]
        REDIS_CACHE --> EXPIRE[自動過期]
        
        MEMORY --> PERF1[讀取: <1ms]
        ATOMIC --> PERF2[寫入: <1ms]
        EXPIRE --> PERF3[過期: 即時]
    end
    
    subgraph "Redis Pub/Sub 訊息系統"
        REDIS_MSG[Redis Pub/Sub]
        REDIS_MSG --> PUBSUB[發布/訂閱]
        REDIS_MSG --> CHANNELS[多頻道支援]
        REDIS_MSG --> REALTIME[即時廣播]
        
        PUBSUB --> PERF4[延遲: <1ms]
        CHANNELS --> PERF5[吞吐量: >10K msg/s]
        REALTIME --> PERF6[廣播: 即時]
    end
    
    subgraph "PostgreSQL 資料持久化"
        PG_DB[PostgreSQL]
        PG_DB --> ACID[ACID 事務]
        PG_DB --> COMPLEX[複雜查詢]
        PG_DB --> PERSIST[資料持久化]
        
        ACID --> PERF7[一致性: 強]
        COMPLEX --> PERF8[查詢: 靈活]
        PERSIST --> PERF9[儲存: 可靠]
    end
```

#### Redis 架構優勢分析

**Redis 緩存優勢:**
- ✅ 極低延遲（<1ms）
- ✅ 高吞吐量（>100K ops/s）
- ✅ 記憶體存儲，高速存取
- ✅ 豐富的資料結構支援
- ✅ 自動過期機制

**Redis Pub/Sub 優勢:**
- ✅ 即時訊息傳遞（<1ms 延遲）
- ✅ 高並發訊息處理
- ✅ 多頻道隔離
- ✅ 水平擴展支援
- ✅ 跨實例通信簡單

#### 實現的混合架構

本專案實現的架構充分利用兩種技術的優勢：

```mermaid
graph TB
    subgraph "完整架構實現"
        subgraph "應用層"
            API[REST API]
            WS[WebSocket]
            FRONTEND[前端應用]
        end
        
        subgraph "緩存與訊息層 (Redis)"
            subgraph "Redis DB 分離"
                REDIS_CACHE["Redis DB 1<br/>用戶緩存"]
                REDIS_MSG["Redis DB 2<br/>訊息佇列"]
            end
            
            REDIS_CACHE --> CACHE_FEATURES[會話管理<br/>API緩存<br/>計數器]
            REDIS_MSG --> MSG_FEATURES[聊天訊息<br/>直播通知<br/>系統廣播]
        end
        
        subgraph "資料持久化層 (PostgreSQL)"
            PG_MAIN[主資料庫]
            PG_FEATURES[用戶資料<br/>影片資料<br/>交易記錄<br/>聊天歷史]
        end
        
        API --> REDIS_CACHE
        API --> PG_MAIN
        WS --> REDIS_MSG
        REDIS_MSG --> WS
        FRONTEND --> API
        FRONTEND --> WS
    end
```

#### 架構實現特色

1. **資料庫隔離**
   - Redis DB 1: 緩存資料
   - Redis DB 2: 訊息佇列
   - PostgreSQL: 持久化資料

2. **智能緩存策略**
   - 會話資料：Redis 緩存（快速驗證）
   - 用戶資料：Redis + PostgreSQL（讀寫分離）
   - 即時計數：Redis 原子操作

3. **即時通信系統**
   - 聊天訊息：Redis Pub/Sub 即時廣播
   - 直播通知：多頻道隔離
   - 系統訊息：統一發布機制

#### 效能提升對比

| 功能 | 純 PostgreSQL | PostgreSQL + Redis | 提升幅度 |
|------|---------------|-------------------|----------|
| 緩存讀取 | 5-50ms | <1ms | **50-500倍** |
| 聊天延遲 | 10-100ms | <1ms | **10-100倍** |
| 訊息吞吐 | 1K-5K msg/s | >10K msg/s | **2-10倍** |
| 會話驗證 | 每次查詢DB | 記憶體驗證 | **100-1000倍** |

## 📊 資料庫設計

### PostgreSQL + Redis 架構圖
```mermaid
graph TB
    subgraph "混合資料庫架構"
        subgraph "PostgreSQL 資料持久化"
            USERS[users - 用戶表]
            VIDEOS[videos - 影片表]
            LIVES[lives - 直播表]
            PAYMENTS[payments - 支付表]
            CHAT[chat_messages - 聊天記錄]
            VQ[video_qualities - 影片品質]
        end
        
        subgraph "Redis 緩存與訊息系統"
            subgraph "Redis DB 1 - 緩存"
                USER_CACHE[用戶會話緩存]
                API_CACHE[API 回應緩存]
                COUNTER_CACHE[計數器緩存]
            end
            
            subgraph "Redis DB 2 - 訊息佇列"
                CHAT_CHANNEL[chat_messages 頻道]
                LIVE_CHANNEL[live_updates 頻道]
                VIDEO_CHANNEL[video_processing 頻道]
                USER_CHANNEL[user_notifications 頻道]
            end
        end
        
        subgraph "PostgreSQL 特殊功能"
            JSONB[JSONB 欄位 - 結構化資料]
            TRIGGER[觸發器 - 自動更新]
            INDEX[GIN 索引 - 全文搜尋]
        end
        
        subgraph "PostgreSQL 擴展"
            UUID[uuid-ossp - UUID 生成]
            TRGM[pg_trgm - 模糊搜尋]
            BTREE[btree_gin - 複合索引]
        end
    end
    
    USERS --> |1:N| VIDEOS
    USERS --> |1:N| LIVES
    USERS --> |1:N| PAYMENTS
    VIDEOS --> |1:N| VQ
    LIVES --> |1:N| CHAT
    
    USER_CACHE -.-> |緩存| USERS
    API_CACHE -.-> |快取| VIDEOS
    CHAT_CHANNEL -.-> |即時| CHAT
    LIVE_CHANNEL -.-> |廣播| LIVES
```

### 核心資料表結構

#### 1. 用戶表 (users)
```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    avatar VARCHAR(500),
    bio TEXT,
    role VARCHAR(20) DEFAULT 'user',
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

#### 2. 影片表 (videos)
```sql
CREATE TABLE videos (
    id SERIAL PRIMARY KEY,
    title VARCHAR(100) NOT NULL,
    description TEXT,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    
    -- 原始影片資訊
    original_url VARCHAR(500) NOT NULL,
    original_key VARCHAR(500),
    thumbnail_url VARCHAR(500),
    
    -- HLS 串流資訊
    hls_master_url VARCHAR(500),
    hls_key VARCHAR(500),
    
    -- MP4 轉碼版本
    mp4_url VARCHAR(500),
    mp4_key VARCHAR(500),
    
    -- 影片屬性
    duration INTEGER DEFAULT 0,
    file_size BIGINT DEFAULT 0,
    original_format VARCHAR(10),
    
    -- 狀態管理
    status VARCHAR(20) NOT NULL,  -- uploading, processing, transcoding, ready, failed
    processing_progress INTEGER DEFAULT 0,  -- 0-100
    error_message VARCHAR(500),
    
    -- 統計資料
    views BIGINT DEFAULT 0,
    likes BIGINT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 創建索引
CREATE INDEX idx_videos_user_status ON videos(user_id, status);
CREATE INDEX idx_videos_status_created ON videos(status, created_at);
CREATE INDEX idx_videos_user_created ON videos(user_id, created_at);
```

#### 3. 影片品質表 (video_qualities)
```sql
CREATE TABLE video_qualities (
    id SERIAL PRIMARY KEY,
    video_id INTEGER REFERENCES videos(id) ON DELETE CASCADE,
    quality VARCHAR(10) NOT NULL,  -- 720p, 480p, 360p
    width INTEGER NOT NULL,
    height INTEGER NOT NULL,
    bitrate INTEGER NOT NULL,
    file_url VARCHAR(500) NOT NULL,
    file_size BIGINT DEFAULT 0,
    status VARCHAR(20) DEFAULT 'ready',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 創建索引
CREATE INDEX idx_video_qualities_video_id ON video_qualities(video_id);
CREATE INDEX idx_video_qualities_quality ON video_qualities(quality);
```

#### 4. 直播表 (lives)
```sql
CREATE TABLE lives (
    id SERIAL PRIMARY KEY,
    title VARCHAR(100) NOT NULL,
    description TEXT,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL,
    start_time TIMESTAMP,
    end_time TIMESTAMP,
    stream_key VARCHAR(100) UNIQUE,
    viewer_count BIGINT DEFAULT 0,
    chat_enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 創建索引
CREATE INDEX idx_lives_user_status ON lives(user_id, status);
CREATE INDEX idx_lives_status_start ON lives(status, start_time);
```

#### 5. Redis 緩存配置

Redis 作為緩存和訊息系統，使用不同的資料庫來隔離功能：

```yaml
# config/config.local.yaml
redis:
  master:
    host: "localhost"
    port: 6379
    password: ""
    db: 0
  pool:
    max_active: 100
    max_idle: 20
    idle_timeout: 300

cache:
  type: "redis"
  db: 1                     # Redis DB 1 用於緩存
  key_prefix: "cache:"      # 緩存鍵前綴
  default_expiration: 3600  # 默認過期時間（秒）

messaging:
  type: "redis"
  db: 2                     # Redis DB 2 用於訊息佇列
  channels:
    - "video_processing"    # 影片處理通知
    - "live_updates"        # 直播更新通知
    - "user_notifications"  # 用戶通知
    - "chat_messages"       # 聊天訊息
```

## 🔧 功能模組地圖

### 功能架構圖
```mermaid
mindmap
  root((串流平台))
    認證系統
      用戶註冊
      用戶登入
      JWT Token 管理
      角色權限控制
      密碼加密
    影片模組
      影片上傳
        MinIO 預簽名 URL
        檔案格式驗證
        大小限制檢查
      自動轉碼處理
        背景服務監控
        FFmpeg 本地轉碼
        HLS 多品質生成
        MP4 網頁版本
        縮圖自動生成
      影片管理
        列表展示
        搜尋功能
        編輯資訊
        刪除影片
        觀看統計
      智能播放
        自動品質切換
        垂直影片比例保持
        即時載入優化
    直播模組
      直播管理
        創建直播間
        串流金鑰管理
        直播狀態控制
      即時互動
        WebSocket 聊天
        觀眾人數統計
        聊天室管理
      直播推流
        RTMP 推流接收
        HLS 直播流分發
        CDN 加速
    支付模組
      訂單管理
        創建支付訂單
        訂單狀態追蹤
        支付記錄查詢
      支付處理
        第三方支付整合
        支付結果通知
        退款處理
    系統服務
      緩存系統
        Redis 記憶體緩存
        自動過期機制
        高速存取
      訊息佇列
        Redis Pub/Sub
        即時訊息廣播
        多頻道隔離
      日誌系統
        結構化日誌
        錯誤追蹤
        性能監控
```

## 🎨 前端頁面地圖

### 前端路由架構
```mermaid
graph TB
    subgraph "公開頁面 (無需認證)"
        HOME[首頁 /]
        LOGIN[登入 /login]
        REGISTER[註冊 /register]
    end
    
    subgraph "認證頁面 (需要登入)"
        DASHBOARD[儀表板 /dashboard]
        PROFILE[個人資料 /profile]
        
        subgraph "影片相關"
            VIDEO_LIST[影片列表 /videos]
            VIDEO_UPLOAD[影片上傳 /videos/upload]
            VIDEO_DETAIL[影片詳情 /videos/:id]
        end
        
        subgraph "直播相關"
            LIVE_LIST[直播列表 /lives]
            LIVE_CREATE[創建直播 /lives/create]
            LIVE_DETAIL[直播詳情 /lives/:id]
            LIVE_STREAM[直播間 /lives/:id/stream]
        end
        
        subgraph "支付相關"
            PAYMENT_LIST[支付記錄 /payments]
            PAYMENT_CREATE[創建支付 /payments/create]
            PAYMENT_DETAIL[支付詳情 /payments/:id]
        end
    end
    
    subgraph "錯誤頁面"
        NOT_FOUND[404 頁面 /*]
    end
    
    HOME --> LOGIN
    HOME --> REGISTER
    LOGIN --> DASHBOARD
    REGISTER --> DASHBOARD
    DASHBOARD --> VIDEO_LIST
    DASHBOARD --> LIVE_LIST
    DASHBOARD --> PAYMENT_LIST
    DASHBOARD --> PROFILE
```

### 前端組件架構
```mermaid
graph TB
    subgraph "佈局組件"
        APP[App.vue - 根組件]
        NAVBAR[NavBar.vue - 導航欄]
        LAYOUT[Layout.vue - 主佈局]
    end
    
    subgraph "通用組件"
        BUTTON[Button.vue - 按鈕]
        INPUT[Input.vue - 輸入框]
        MODAL[Modal.vue - 彈窗]
        LOADING[Loading.vue - 載入動畫]
    end
    
    subgraph "業務組件"
        subgraph "影片組件"
            VIDEO_LIST_COMP[VideoList.vue - 影片列表]
            VIDEO_PLAYER[VideoPlayer.vue - 影片播放器]
            VIDEO_UPLOAD_COMP[VideoUpload.vue - 上傳組件]
            VIDEO_DETAIL[VideoDetailView.vue - 影片詳情]
        end
        
        subgraph "直播組件"
            LIVE_PLAYER[LivePlayer.vue - 直播播放器]
            LIVE_CHAT[LiveChat.vue - 聊天組件]
            LIVE_CREATE_COMP[LiveCreate.vue - 創建直播]
        end
    end
    
    APP --> NAVBAR
    APP --> LAYOUT
    LAYOUT --> VIDEO_LIST_COMP
    LAYOUT --> LIVE_PLAYER
    VIDEO_LIST_COMP --> VIDEO_PLAYER
    LIVE_PLAYER --> LIVE_CHAT
    VIDEO_DETAIL --> VIDEO_PLAYER
```

## 📈 業務流程圖

### 影片上傳與自動轉檔流程
```mermaid
sequenceDiagram
    participant U as 用戶
    participant FE as 前端
    participant BE as 後端
    participant MINIO_ORIG as MinIO 原始桶
    participant MINIO_PROC as MinIO 處理桶
    participant FFMPEG as FFmpeg 轉碼器
    participant DB as PostgreSQL
    participant TW as 轉碼背景服務
    
    U->>FE: 選擇影片檔案
    FE->>BE: 請求上傳 URL
    BE->>BE: 驗證檔案格式/大小
    BE->>MINIO_ORIG: 生成預簽名 URL
    MINIO_ORIG-->>BE: 返回上傳 URL
    BE-->>FE: 返回上傳 URL 和 Key
    FE->>MINIO_ORIG: 直接上傳檔案
    MINIO_ORIG-->>FE: 上傳完成
    FE->>BE: 確認上傳完成 (ConfirmUploadOnly)
    BE->>DB: 創建影片記錄 (status: uploading)
    BE->>TW: 啟動異步轉碼
    
    Note over TW: 背景服務監控
    TW->>DB: SELECT FOR UPDATE 查詢待轉碼影片
    DB-->>TW: 返回待處理影片列表
    
    loop 每個待轉碼影片
        TW->>FFMPEG: 觸發轉碼任務
        FFMPEG->>MINIO_ORIG: 下載原始影片
        MINIO_ORIG-->>FFMPEG: 返回影片檔案
        
        FFMPEG->>FFMPEG: 多品質轉碼處理
        Note right of FFMPEG: 720p, 480p, 360p HLS<br/>MP4 網頁版本<br/>縮圖生成
        
        FFMPEG->>MINIO_PROC: 上傳轉碼結果
        MINIO_PROC-->>FFMPEG: 上傳完成
        FFMPEG-->>TW: 轉碼完成通知
        
        TW->>DB: 更新影片狀態 (status: ready)
        TW->>DB: 創建品質記錄
        Note right of DB: 更新 URL 為 stream-demo-processed 路徑
    end
    
    BE->>FE: 回應上傳確認成功
    FE->>U: 顯示上傳成功，提示可能正在轉碼
```

### 背景轉碼服務流程

```mermaid
sequenceDiagram
    participant TW as 轉碼背景服務
    participant DB as PostgreSQL
    participant FFMPEG as FFmpeg 容器
    participant MINIO_ORIG as MinIO 原始桶
    participant MINIO_PROC as MinIO 處理桶
    
    Note over TW: 服務啟動時自動檢查
    TW->>DB: 查詢待轉碼影片
    Note right of TW: SELECT FOR UPDATE<br/>status IN ('uploading', 'processing', 'transcoding')
    
    DB-->>TW: 返回待處理影片列表
    
    loop 每個影片
        TW->>TW: 驗證影片檔案存在性
        TW->>DB: 更新狀態為 'transcoding'
        
        TW->>FFMPEG: 觸發轉碼任務
        Note right of FFMPEG: 輸入: videos/original/{user_id}/{uuid}.{ext}<br/>輸出: videos/processed/{user_id}/{video_id}/
        
        FFMPEG->>MINIO_ORIG: 下載原始影片
        MINIO_ORIG-->>FFMPEG: 返回影片檔案
        
        FFMPEG->>FFMPEG: 分析影片資訊
        Note right of FFMPEG: 解析尺寸、時長、格式<br/>保持原始比例 (scale=width:-1)
        
        parallel 多格式轉碼
            FFMPEG->>FFMPEG: 生成 MP4 (H.264+AAC)
            and FFMPEG->>FFMPEG: 生成 HLS 720p
            and FFMPEG->>FFMPEG: 生成 HLS 480p  
            and FFMPEG->>FFMPEG: 生成 HLS 360p
            and FFMPEG->>FFMPEG: 生成縮圖 (多尺寸)
            and FFMPEG->>FFMPEG: 生成時間軸縮圖
        end
        
        FFMPEG->>MINIO_PROC: 上傳 MP4 版本
        FFMPEG->>MINIO_PROC: 上傳 HLS 串流檔案
        FFMPEG->>MINIO_PROC: 上傳縮圖檔案
        FFMPEG->>MINIO_PROC: 上傳轉碼報告
        
        FFMPEG-->>TW: 轉碼完成通知
        TW->>DB: 更新影片 URL 和狀態
        Note right of DB: MP4URL, HLSMasterURL<br/>ThumbnailURL, Status: ready<br/>所有 URL 指向 stream-demo-processed
        
        TW->>DB: 創建品質記錄
        Note right of DB: VideoQuality: 720p, 480p, 360p<br/>file_url 指向 stream-demo-processed
    end
    
    Note over TW: 每 30 秒重複檢查
```

### 前端智能播放流程

```mermaid
sequenceDiagram
    participant U as 用戶
    participant FE as 前端
    participant BE as 後端
    participant MINIO_PROC as MinIO 處理桶
    
    U->>FE: 進入影片詳情頁
    FE->>BE: 獲取影片資訊
    BE->>BE: 檢查影片狀態
    alt 影片狀態為 'ready'
        BE-->>FE: 返回影片資訊 + 品質列表
        FE->>FE: 設置預設品質為 'auto'
        FE->>FE: 自動載入影片資源
        
        alt 選擇 MP4 播放
            FE->>MINIO_PROC: 載入 MP4 檔案
            MINIO_PROC-->>FE: 返回 MP4 串流
            FE->>FE: 設置 video.src
            FE->>FE: 監聽載入事件
            FE->>FE: 自動播放
        else 選擇 HLS 播放
            FE->>MINIO_PROC: 載入 HLS master playlist
            MINIO_PROC-->>FE: 返回 m3u8 檔案
            FE->>FE: 初始化 hls.js
            FE->>FE: 自動選擇最佳品質
            FE->>FE: 開始播放
        end
        
        Note over FE: 智能品質監控
        loop 播放過程中
            FE->>FE: 監控緩衝狀態
            alt 緩衝過多或載入緩慢
                FE->>FE: 自動切換到較低品質
                FE->>U: 顯示品質切換通知
            end
        end
        
    else 影片狀態為 'uploading' 或 'transcoding'
        BE-->>FE: 返回影片資訊 (無播放 URL)
        FE->>U: 顯示轉碼中狀態
        FE->>U: 提示用戶刷新列表
    end
    
    U->>FE: 手動切換品質
    FE->>FE: 重新載入對應品質
    FE->>U: 顯示載入動畫
    FE->>U: 播放新品質
```

### 檔案存儲結構

```
MinIO Bucket: stream-demo-videos/ (原始檔案)
├── videos/
│   └── original/              # 原始上傳檔案
│       └── {user_id}/
│           └── {uuid}.{ext}   # 例：431254c8-6bdc-4137-969b-5fa3d9ae9788.mov

MinIO Bucket: stream-demo-processed/ (轉碼後檔案)
├── videos/
│   └── processed/             # 轉碼後檔案
│       └── {user_id}/
│           └── {video_id}/
│               ├── video.mp4                    # MP4 播放版本
│               ├── hls/                        # HLS 串流
│               │   ├── index.m3u8              # 主播放列表
│               │   ├── 720p/
│               │   │   ├── index.m3u8
│               │   │   └── segment_*.ts
│               │   ├── 480p/
│               │   │   ├── index.m3u8
│               │   │   └── segment_*.ts
│               │   └── 360p/
│               │       ├── index.m3u8
│               │       └── segment_*.ts
│               ├── thumbnails/                 # 縮圖
│               │   ├── thumb_320x240.jpg
│               │   ├── thumb_640x480.jpg
│               │   ├── thumb_1280x720.jpg
│               │   └── timeline_*.jpg          # 時間軸縮圖
│               └── transcode_report.json       # 轉碼報告
```

### 直播聊天流程
```mermaid
sequenceDiagram
    participant U1 as 用戶A
    participant U2 as 用戶B
    participant FE1 as 前端A
    participant FE2 as 前端B
    participant WS as WebSocket Hub
    participant PG as PostgreSQL
    
    U1->>FE1: 加入直播間
    FE1->>WS: WebSocket 連接
    WS->>PG: LISTEN chat_messages
    WS->>WS: 註冊用戶到聊天室
    
    U2->>FE2: 加入同個直播間
    FE2->>WS: WebSocket 連接
    WS->>WS: 註冊用戶到聊天室
    
    U1->>FE1: 發送聊天訊息
    FE1->>WS: 傳送訊息
    WS->>PG: NOTIFY chat_messages
    PG->>WS: 訊息廣播
    WS->>FE1: 廣播給所有用戶
    WS->>FE2: 廣播給所有用戶
    FE1->>U1: 顯示訊息
    FE2->>U2: 顯示訊息
```

### 用戶認證流程
```mermaid
sequenceDiagram
    participant U as 用戶
    participant FE as 前端
    participant BE as 後端
    participant DB as PostgreSQL
    participant JWT as JWT Service
    
    U->>FE: 輸入登入資訊
    FE->>BE: 發送登入請求
    BE->>DB: 查詢用戶資料
    DB-->>BE: 返回用戶資料
    BE->>BE: 驗證密碼
    BE->>JWT: 生成 JWT Token
    JWT-->>BE: 返回 Token
    BE-->>FE: 返回 Token 和用戶資訊
    FE->>FE: 儲存 Token 到 LocalStorage
    FE->>FE: 更新全域認證狀態
    FE-->>U: 登入成功，跳轉儀表板
    
    Note over FE,BE: 後續請求
    U->>FE: 訪問受保護頁面
    FE->>BE: 帶 JWT Token 的請求
    BE->>JWT: 驗證 Token
    JWT-->>BE: Token 有效
    BE-->>FE: 返回受保護資料
    FE-->>U: 顯示頁面內容
```

## 🛠️ 部署架構

### Docker 容器部署
```mermaid
graph TB
    subgraph "Docker Compose 完整架構"
        subgraph "前端容器"
            FE_CONTAINER[Vue.js Frontend<br/>Nginx]
        end
        
        subgraph "後端容器"
            BE_CONTAINER[Go Backend<br/>Gin Server]
        end
        
        subgraph "資料庫容器"
            PG_MASTER[PostgreSQL Master<br/>端口: 5432]
            MYSQL_SLAVE[MySQL Slave<br/>端口: 3306]
            REDIS_CONTAINER[Redis<br/>端口: 6379]
        end
        
        subgraph "存儲與轉碼容器"
            MINIO_CONTAINER[MinIO 對象存儲<br/>API: 9000, Console: 9001]
            FFMPEG_CONTAINER[FFmpeg 轉碼器<br/>Alpine + FFmpeg 6.0.1]
        end
        
        subgraph "監控容器 (可選)"
            PROMETHEUS[Prometheus 監控]
            GRAFANA[Grafana 儀表板]
            LOG_CONTAINER[Log Aggregator]
        end
    end
    
    FE_CONTAINER --> BE_CONTAINER
    BE_CONTAINER --> PG_MASTER
    BE_CONTAINER --> MYSQL_SLAVE
    BE_CONTAINER --> REDIS_CONTAINER
    BE_CONTAINER --> MINIO_CONTAINER
    BE_CONTAINER --> FFMPEG_CONTAINER
    
    FFMPEG_CONTAINER --> MINIO_CONTAINER
    PG_MASTER -.-> MYSQL_SLAVE
```

### 🚀 快速開始

#### 環境要求
- Docker & Docker Compose
- Go 1.24.3+
- Node.js 18+

#### 啟動完整開發環境

```bash
# 克隆專案
git clone <repository-url>
cd stream-demo

# 啟動所有 Docker 服務
docker-compose up -d

# 檢查服務狀態
docker-compose ps
```

#### 服務端口說明

| 服務 | 端口 | 描述 |
|------|------|------|
| PostgreSQL | 5432 | 主資料庫 |
| MySQL | 3306 | 從資料庫 |
| Redis | 6379 | 緩存與訊息佇列 |
| MinIO API | 9000 | S3 兼容 API |
| MinIO Console | 9001 | 管理界面 |
| Go 後端 | 8080 | REST API 服務 |
| Vue 前端 | 3000 | 開發伺服器 |

#### MinIO 初始設置

```bash
# MinIO 管理界面
http://localhost:9001
# 默認帳號: minioadmin / minioadmin

# 創建初始儲存桶
docker exec stream-demo-minio mc alias set local http://localhost:9000 minioadmin minioadmin
docker exec stream-demo-minio mc mb local/stream-demo-videos
docker exec stream-demo-minio mc mb local/stream-demo-processed
docker exec stream-demo-minio mc anonymous set public local/stream-demo-videos
docker exec stream-demo-minio mc anonymous set public local/stream-demo-processed
```

#### 測試轉碼功能

```bash
# 1. 啟動後端服務
cd backend && go run main.go

# 2. 上傳測試影片（透過前端或 API）
# 前端: http://localhost:3000
# 後端 API: http://localhost:8080

# 3. 檢查轉碼狀態
docker logs stream-demo-transcoder

# 4. 查看轉碼結果
docker exec stream-demo-transcoder mc ls s3/stream-demo-processed/videos/processed/ --recursive

# 5. 手動測試轉碼（可選）
docker exec stream-demo-transcoder /scripts/transcode.sh \
  "videos/original/1/test.mov" \
  "videos/processed/1/1" \
  "1" \
  "1"
```

#### 轉碼後檔案格式

- **MP4 版本**: `stream-demo-processed/videos/processed/{user_id}/{video_id}/video.mp4` - 最佳瀏覽器相容性
- **HLS 串流**: `stream-demo-processed/videos/processed/{user_id}/{video_id}/hls/index.m3u8` - 多品質適應性串流
- **縮圖**: `stream-demo-processed/videos/processed/{user_id}/{video_id}/thumbnails/` - 多尺寸預覽圖

---

## 🚀 **最新更新 (2025-01)**

### ✨ **完整的影片上傳與自動轉檔系統**

我們已經實現了完整的影片上傳與自動轉檔系統，包括：

#### 🔧 **核心功能實現**
- ✅ **雙桶存儲架構**: 原始檔案存儲於 `stream-demo-videos`，轉碼後檔案存儲於 `stream-demo-processed`
- ✅ **背景轉碼服務**: `TranscodeWorker` 自動監控資料庫，處理待轉碼影片
- ✅ **智能檔案管理**: 使用 UUID 命名原始檔案，避免衝突
- ✅ **多品質轉碼**: 720p、480p、360p HLS 串流 + MP4 網頁版本
- ✅ **比例保持**: 使用 `scale=width:-1` 保持原始影片比例，支援垂直影片
- ✅ **縮圖生成**: 多尺寸縮圖 + 時間軸預覽圖

#### 🎯 **轉碼流程優化**
```mermaid
graph LR
    A[用戶上傳] --> B[MinIO 原始桶]
    B --> C[背景服務監控]
    C --> D[FFmpeg 自動轉碼]
    D --> E[多品質 HLS]
    D --> F[MP4 網頁版本]
    D --> G[縮圖生成]
    E --> H[MinIO 處理桶]
    F --> H
    G --> H
    H --> I[前端智能播放]
```

#### 📂 **檔案結構優化**
- **原始檔案**: `stream-demo-videos/videos/original/{user_id}/{uuid}.{ext}` - 永久保留
- **轉碼產出**: `stream-demo-processed/videos/processed/{user_id}/{video_id}/` - 多格式組織
- **智能播放**: 前端優先使用 MP4 → HLS → 原始檔案

#### 🐳 **Docker 完整化**
- **PostgreSQL + MySQL + Redis**: 完整資料庫支援
- **MinIO**: S3 兼容對象存儲 (API: 9000, Console: 9001)
- **FFmpeg Transcoder**: 專用轉碼容器，支援 Alpine + MinIO Client

#### ⚡ **效能提升**
- **本地化處理**: 無需 AWS 服務，降低成本和延遲
- **並行轉碼**: 同時生成多品質版本
- **智能觸發**: 所有上傳影片都會進行轉碼
- **瀏覽器優化**: MP4 優先確保最佳相容性

#### 🎮 **前端播放體驗**
- **自動載入**: 進入影片詳情頁自動載入影片資源
- **智能品質切換**: 根據網路狀況自動切換品質
- **比例保持**: 垂直影片正確顯示，不會被壓縮
- **即時反饋**: 轉碼中狀態提示，完成後自動刷新

#### 🔄 **背景服務特色**
- **SELECT FOR UPDATE**: 防止並發處理同一影片
- **事務管理**: 確保資料庫操作原子性
- **錯誤處理**: 完善的錯誤記錄和重試機制
- **狀態追蹤**: 詳細的處理進度和狀態更新

---

**開發環境現在只需一個命令即可完整啟動！** 🎉
```bash
docker-compose up -d
```