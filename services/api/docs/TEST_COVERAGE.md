# 測試覆蓋率提升指南

## 概述

本文檔說明如何將 Go 測試覆蓋率從當前的 2.4% 提升到 100%。

## 當前狀況

- **整體覆蓋率**: 2.4%
- **API 包覆蓋率**: 3.4%
- **Utils 包覆蓋率**: 21.7%
- **其他包覆蓋率**: 0%

## 測試工具

### 測試腳本

我們提供了一個便捷的測試腳本來運行測試和檢查覆蓋率：

```bash
# 運行所有測試並生成覆蓋率報告
./scripts/test.sh --all --coverage --check-coverage

# 運行單元測試
./scripts/test.sh --unit --coverage

# 設置自定義覆蓋率閾值
./scripts/test.sh --all --coverage --check-coverage --threshold 80
```

### CI/CD 配置

GitHub Actions 已配置為：
- 運行所有測試
- 生成覆蓋率報告
- 檢查覆蓋率閾值（目前設為 50%）
- 上傳 HTML 覆蓋率報告作為構建產物

## 覆蓋率提升計劃

### 第一階段：核心功能測試（目標：30%）

#### 1. API 層測試
- [x] 用戶 API 測試（註冊、登入、獲取資料）
- [ ] 直播 API 測試
- [ ] 影片 API 測試
- [ ] 支付 API 測試
- [ ] 公共串流 API 測試

#### 2. 服務層測試
- [ ] 用戶服務測試
- [ ] 直播服務測試
- [ ] 影片服務測試
- [ ] 支付服務測試
- [ ] 串流服務測試

#### 3. 工具函數測試
- [x] 錯誤處理測試
- [x] JWT 工具測試
- [x] 日誌工具測試
- [x] 密碼工具測試
- [x] 訊息處理測試
- [ ] 緩存工具測試
- [ ] Redis 工具測試

### 第二階段：資料庫層測試（目標：60%）

#### 1. Repository 測試
- [ ] PostgreSQL Repository 測試
- [ ] MySQL Repository 測試
- [ ] Redis Repository 測試

#### 2. 模型測試
- [ ] 用戶模型測試
- [ ] 直播模型測試
- [ ] 影片模型測試
- [ ] 支付模型測試

### 第三階段：整合測試（目標：80%）

#### 1. 端到端測試
- [ ] 用戶註冊到登入流程
- [ ] 影片上傳到播放流程
- [ ] 直播創建到結束流程
- [ ] 支付流程測試

#### 2. 中間件測試
- [ ] 認證中間件測試
- [ ] 錯誤處理中間件測試
- [ ] 日誌中間件測試

### 第四階段：邊界情況測試（目標：100%）

#### 1. 錯誤處理測試
- [ ] 資料庫連接失敗
- [ ] Redis 連接失敗
- [ ] 外部 API 失敗
- [ ] 檔案上傳失敗

#### 2. 效能測試
- [ ] 高併發測試
- [ ] 記憶體洩漏測試
- [ ] 資料庫效能測試

## 測試最佳實踐

### 1. 使用 Mock 和 Stub

對於外部依賴，使用 Mock 來隔離測試：

```go
// 創建模擬服務
mockUserService := &mocks.MockUserService{}

// 設置期望行為
mockUserService.On("Register", "testuser", "test@example.com", "password123").
    Return(&dto.UserDTO{ID: 1, Username: "testuser"}, nil)

// 創建處理器
handler := &UserHandler{userService: mockUserService}
```

### 2. 測試資料庫操作

使用測試資料庫或 SQL Mock：

```go
// 使用 SQL Mock
db, mock, err := sqlmock.New()
if err != nil {
    t.Fatalf("Failed to create mock: %v", err)
}
defer db.Close()

// 設置期望的 SQL 查詢
mock.ExpectQuery("SELECT (.+) FROM users").
    WithArgs("testuser").
    WillReturnRows(sqlmock.NewRows([]string{"id", "username"}).
        AddRow(1, "testuser"))
```

### 3. 測試 HTTP 處理器

使用 `httptest` 包測試 HTTP 處理器：

```go
// 創建測試請求
req, _ := http.NewRequest("POST", "/api/v1/users/register", bytes.NewBuffer(jsonData))
req.Header.Set("Content-Type", "application/json")

// 創建響應記錄器
w := httptest.NewRecorder()

// 執行請求
router.ServeHTTP(w, req)

// 驗證結果
assert.Equal(t, http.StatusOK, w.Code)
```

### 4. 測試 WebSocket

使用 `gorilla/websocket` 的測試工具：

```go
// 創建測試服務器
server := httptest.NewServer(http.HandlerFunc(handler.ServeWS))
defer server.Close()

// 連接到 WebSocket
wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
if err != nil {
    t.Fatalf("Failed to connect: %v", err)
}
defer conn.Close()
```

## 測試檔案命名規範

- 單元測試：`*_test.go`
- 整合測試：`*_integration_test.go`
- 效能測試：`*_benchmark_test.go`
- 測試輔助檔案：`test_helpers.go`

## 覆蓋率檢查

### 本地檢查

```bash
# 生成覆蓋率報告
go test -coverprofile=coverage.out ./...

# 查看函數級別覆蓋率
go tool cover -func=coverage.out

# 生成 HTML 報告
go tool cover -html=coverage.out -o coverage.html
```

### CI/CD 檢查

GitHub Actions 會自動：
1. 運行所有測試
2. 生成覆蓋率報告
3. 檢查是否達到閾值（50%）
4. 上傳 HTML 報告

## 常見問題

### Q: 如何處理外部依賴？

A: 使用 Mock 和 Stub 來隔離外部依賴，確保測試的獨立性。

### Q: 如何測試資料庫操作？

A: 使用測試資料庫或 SQL Mock，避免影響生產資料。

### Q: 如何測試 WebSocket？

A: 使用 `httptest` 創建測試服務器，然後使用 WebSocket 客戶端進行測試。

### Q: 如何提高測試執行速度？

A: 使用並行測試、測試快取，並避免不必要的 I/O 操作。

## 下一步行動

1. **立即開始**：為 API 層添加更多測試
2. **本週目標**：達到 30% 覆蓋率
3. **本月目標**：達到 80% 覆蓋率
4. **長期目標**：達到 100% 覆蓋率

## 參考資源

- [Go Testing Package](https://golang.org/pkg/testing/)
- [Testify Framework](https://github.com/stretchr/testify)
- [SQL Mock](https://github.com/DATA-DOG/go-sqlmock)
- [Gin Testing](https://github.com/gin-gonic/gin#testing) 