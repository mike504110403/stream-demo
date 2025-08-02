# GitHub CI 整合測試設定指南

## 概述

本專案包含多個服務，需要進行完整的整合測試。以下是各服務的測試策略：

## 服務架構

### 後端服務 (Go)
- **主要服務**: `backend/main.go`
- **串流拉取器**: `backend/cmd/stream_puller/main.go`
- **測試命令**: `go test -v -race ./...`

### 前端服務 (Vue.js + TypeScript)
- **框架**: Vue 3 + Vite + TypeScript
- **測試框架**: Vitest
- **測試命令**: `npm run test`

### 基礎設施服務
- **資料庫**: PostgreSQL
- **快取**: Redis
- **物件儲存**: MinIO
- **反向代理**: Nginx

## CI 工作流程

### 1. 後端測試 (backend-test)
```yaml
- 設定 PostgreSQL 和 Redis 服務
- 執行 Go 單元測試
- 執行程式碼覆蓋率分析
- 執行 linter 檢查
```

### 2. 前端建置測試 (frontend-build)
```yaml
- 安裝 Node.js 依賴
- 執行 ESLint 檢查
- 執行 TypeScript 型別檢查
- 建置前端專案
```

### 3. Docker 建置測試 (docker-build)
```yaml
- 建置後端 Docker 映像
- 建置串流拉取器 Docker 映像
- 建置前端 Docker 映像
```

### 4. 整合測試 (integration-test)
```yaml
- 啟動後端服務
- 測試 API 端點
- 驗證服務間通訊
```

## 設定 GitHub Branch Protection Rules

### 步驟 1: 前往專案設定
1. 在 GitHub 專案頁面點擊 "Settings"
2. 在左側選單中點擊 "Branches"

### 步驟 2: 建立 Branch Protection Rule
1. 點擊 "Add rule"
2. 在 "Branch name pattern" 中輸入 `main`
3. 勾選以下選項：

#### 基本保護
- ✅ **Require a pull request before merging**
- ✅ **Require approvals** (建議設定 1-2 個審查者)
- ✅ **Dismiss stale PR approvals when new commits are pushed**

#### 狀態檢查
- ✅ **Require status checks to pass before merging**
- ✅ **Require branches to be up to date before merging**

在狀態檢查中新增以下檢查：
- `Backend Tests`
- `Frontend Build Test`
- `Docker Build Test`
- `Integration Tests`

#### 其他保護
- ✅ **Block force pushes**
- ✅ **Do not allow bypassing the above settings**

### 步驟 3: 儲存設定
點擊 "Create" 按鈕儲存設定。

## 本地測試

### 後端測試
```bash
cd backend
go test -v -race ./...
go vet ./...
go fmt ./...
```

### 前端建置測試
```bash
cd frontend
npm install
npm run lint
npm run type-check
npm run build
```

### 整合測試
```bash
# 啟動基礎設施服務
cd docker
./docker-manage.sh start

# 啟動後端服務
cd backend
go run main.go

# 啟動前端服務
cd frontend
npm run dev

# 執行 API 測試
curl http://localhost:8080/health
curl http://localhost:5173
```

## 測試覆蓋率

### 後端覆蓋率
```bash
cd backend
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

### 前端覆蓋率
```bash
cd frontend
npm run test:coverage
```

## 故障排除

### 常見問題

1. **狀態檢查失敗**
   - 檢查 GitHub Actions 日誌
   - 確認本地測試通過
   - 檢查依賴版本是否一致

2. **整合測試失敗**
   - 確認基礎設施服務正常運行
   - 檢查環境變數設定
   - 確認網路連接正常

3. **Docker 建置失敗**
   - 檢查 Dockerfile 語法
   - 確認依賴檔案存在
   - 檢查映像名稱衝突

### 重新執行檢查
如果檢查失敗，可以：
1. 推送新的提交
2. 在 PR 頁面點擊 "Re-run checks"
3. 檢查並修復問題後重新提交

## 進階設定

### 自定義狀態檢查
可以在 `.github/workflows/` 目錄中新增自定義工作流程：

```yaml
name: Custom Check
on: [pull_request]
jobs:
  custom-test:
    runs-on: ubuntu-latest
    steps:
      - name: Custom test step
        run: echo "Custom test"
```

### 條件執行
使用 `if` 條件來控制工作流程執行：

```yaml
- name: Run expensive tests
  if: github.event_name == 'pull_request'
  run: npm run test:integration
```

### 快取優化
使用 GitHub Actions 快取來加速建置：

```yaml
- name: Cache dependencies
  uses: actions/cache@v3
  with:
    path: |
      ~/.npm
      node_modules
    key: ${{ runner.os }}-node-${{ hashFiles('**/package-lock.json') }}
``` 