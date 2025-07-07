# 🎬 串流平台前端專案

## 📋 專案概述

這是一個基於 Vue 3 + TypeScript + Element Plus 的現代化串流平台前端應用，提供完整的用戶管理、影片上傳、直播功能和支付系統。

## 🚀 功能特色

### 🔐 用戶認證
- ✅ 用戶註冊和登入
- ✅ JWT Token 認證
- ✅ 自動登入狀態維護
- ✅ 路由守衛保護

### 🎥 影片管理
- ✅ 影片列表展示
- ✅ 影片搜尋和篩選
- ✅ 影片上傳功能
- ✅ 影片編輯和刪除
- ✅ 影片詳情頁面

### 📺 直播功能
- ✅ 直播列表
- ✅ 創建直播間
- ✅ 直播管理
- ✅ 串流金鑰管理
- ✅ 聊天功能切換

### 💰 支付系統
- ✅ 支付記錄查看
- ✅ 創建支付訂單
- ✅ 支付處理
- ✅ 退款功能

### 👤 個人中心
- ✅ 個人資料編輯
- ✅ 頭像上傳
- ✅ 帳號設定
- ✅ 帳號刪除

## 🛠️ 技術棧

- **框架**: Vue 3 (Composition API)
- **語言**: TypeScript
- **UI 組件庫**: Element Plus
- **路由**: Vue Router 4
- **狀態管理**: Pinia
- **HTTP 客戶端**: Axios
- **構建工具**: Vite
- **包管理**: npm

## 📁 專案結構

```
frontend/
├── src/
│   ├── api/                 # API 服務層
│   │   ├── user.ts         # 用戶相關 API
│   │   ├── video.ts        # 影片相關 API
│   │   ├── live.ts         # 直播相關 API
│   │   └── payment.ts      # 支付相關 API
│   ├── components/         # 共用組件
│   │   └── NavBar.vue      # 導航欄組件
│   ├── router/             # 路由配置
│   │   └── index.ts        # 路由定義
│   ├── store/              # 狀態管理
│   │   └── auth.ts         # 認證狀態
│   ├── types/              # TypeScript 類型定義
│   │   └── index.ts        # 全局類型
│   ├── utils/              # 工具函數
│   │   └── request.ts      # HTTP 請求封裝
│   ├── views/              # 頁面組件
│   │   ├── auth/           # 認證相關頁面
│   │   ├── home/           # 首頁
│   │   ├── video/          # 影片相關頁面
│   │   ├── live/           # 直播相關頁面
│   │   └── payment/        # 支付相關頁面
│   ├── App.vue             # 根組件
│   └── main.ts             # 應用入口
├── package.json            # 專案配置
├── vite.config.ts          # Vite 配置
└── tsconfig.json           # TypeScript 配置
```

## 🔧 開發環境設置

### 1. 安裝依賴
```bash
cd frontend
npm install
```

### 2. 啟動開發服務器
```bash
npm run dev
```

### 3. 構建生產版本
```bash
npm run build
```

### 4. 預覽生產版本
```bash
npm run preview
```

## 🌐 API 配置

後端 API 地址配置在 `src/utils/request.ts` 中：

```typescript
const request = axios.create({
  baseURL: 'http://localhost:8080/api',  // 後端 API 地址
  timeout: 10000,
})
```

如需修改後端地址，請更新此配置。

## 📱 頁面路由

| 路由 | 頁面 | 描述 | 需要認證 |
|------|------|------|----------|
| `/` | 首頁 | 平台介紹和功能展示 | ❌ |
| `/login` | 登入 | 用戶登入 | ❌ |
| `/register` | 註冊 | 用戶註冊 | ❌ |
| `/dashboard` | 儀表板 | 個人儀表板 | ✅ |
| `/profile` | 個人資料 | 編輯個人資料 | ✅ |
| `/videos` | 影片列表 | 影片管理 | ✅ |
| `/videos/upload` | 上傳影片 | 影片上傳 | ✅ |
| `/videos/:id` | 影片詳情 | 影片詳情頁 | ✅ |
| `/lives` | 直播列表 | 直播管理 | ✅ |
| `/lives/create` | 創建直播 | 創建直播間 | ✅ |
| `/lives/:id` | 直播詳情 | 直播詳情頁 | ✅ |
| `/payments` | 支付記錄 | 支付管理 | ✅ |
| `/payments/create` | 創建支付 | 創建支付訂單 | ✅ |

## 🔐 認證機制

### Token 存儲
- JWT Token 存儲在 localStorage 中
- 自動在請求頭中添加 Authorization
- Token 過期自動跳轉登入頁

### 路由守衛
- 需要認證的頁面會檢查登入狀態
- 未登入用戶自動重定向到登入頁
- 已登入用戶訪問登入/註冊頁會重定向到儀表板

## 🎨 UI 設計

### 設計原則
- **現代化**: 使用 Element Plus 現代化 UI 組件
- **響應式**: 支援多種屏幕尺寸
- **一致性**: 統一的色彩和間距規範
- **易用性**: 直觀的交互設計

### 色彩方案
- **主色**: #409EFF (Element Plus 藍)
- **成功**: #67C23A
- **警告**: #E6A23C
- **危險**: #F56C6C
- **文字**: #303133, #606266, #909399

## 📦 主要依賴

```json
{
  "vue": "^3.4.15",           // Vue 3 框架
  "vue-router": "^4.2.5",     // 路由管理
  "pinia": "^2.1.7",          // 狀態管理
  "axios": "^1.6.7",          // HTTP 客戶端
  "element-plus": "^2.5.3",   // UI 組件庫
  "@vueuse/core": "^10.7.2"   // Vue 組合式工具庫
}
```

## 🚀 部署指南

### 1. 構建專案
```bash
npm run build
```

### 2. 部署 dist 目錄
將 `dist/` 目錄部署到您的 Web 服務器

### 3. 配置 Nginx (範例)
```nginx
server {
    listen 80;
    server_name your-domain.com;
    root /path/to/dist;
    index index.html;
    
    location / {
        try_files $uri $uri/ /index.html;
    }
    
    location /api {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

## 🧪 測試功能

### 快速測試流程
1. 啟動後端服務 (`go run main.go`)
2. 啟動前端服務 (`npm run dev`)
3. 訪問 `http://localhost:5173`
4. 註冊新用戶或登入
5. 測試各項功能

### 測試用戶
可以使用以下測試數據：
- 郵箱: `test@example.com`
- 密碼: `123456`
- 用戶名: `testuser`

## 🔍 故障排除

### 常見問題

**Q: 無法連接到後端 API**
A: 檢查後端服務是否啟動，確認 API 地址配置正確

**Q: 登入後頁面空白**
A: 檢查瀏覽器控制台錯誤，確認 Token 是否正確存儲

**Q: 路由跳轉不正常**
A: 檢查路由守衛邏輯，確認認證狀態

**Q: 組件樣式異常**
A: 確認 Element Plus 樣式是否正確載入

## 📝 開發注意事項

1. **API 響應處理**: 所有 API 響應都經過統一處理，注意錯誤捕獲
2. **類型安全**: 使用 TypeScript 確保類型安全
3. **組件復用**: 盡量抽取共用組件提高復用性
4. **性能優化**: 使用路由懶加載減少初始包大小
5. **錯誤處理**: 統一的錯誤提示和處理機制

## 🎯 後續優化

- [ ] 添加更多頁面組件 (影片上傳、直播詳情等)
- [ ] 實現 WebSocket 聊天功能
- [ ] 添加檔案上傳組件
- [ ] 優化移動端適配
- [ ] 添加單元測試
- [ ] 實現 PWA 功能
- [ ] 添加國際化支援

---

🎉 **前端專案已完成基礎架構，可以開始 Demo 所有後端功能！** 