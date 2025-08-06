# TDD 遷移計劃

## 概述

將現有的串流平台後端從傳統開發模式轉換到 TDD (Test-Driven Development) 模式。

## 當前狀況分析

### 優勢
- ✅ 清晰的分層架構 (API → Service → Repository → Database)
- ✅ 依賴注入容器已實現
- ✅ 部分接口已定義 (UserServiceInterface)
- ✅ 測試基礎設施已建立 (Mock 框架、測試工具)

### 劣勢
- ❌ 測試覆蓋率低 (2.4%)
- ❌ 缺乏完整的接口定義
- ❌ 測試驅動開發文化尚未建立
- ❌ 重構阻力較大

## TDD 轉換計劃

### 第一階段：基礎設施準備 (1-2 週)

#### 1. 完善接口定義
```go
// 需要創建的接口
type VideoServiceInterface interface {
    UploadVideo(file *multipart.FileHeader, userID uint) (*dto.VideoDTO, error)
    GetVideoByID(videoID uint) (*dto.VideoDTO, error)
    GetVideosByUser(userID uint, page, pageSize int) ([]*dto.VideoDTO, int64, error)
    DeleteVideo(videoID, userID uint) error
}

type LiveServiceInterface interface {
    CreateLiveStream(userID uint, title, description string) (*dto.LiveDTO, error)
    StartLiveStream(liveID, userID uint) error
    StopLiveStream(liveID, userID uint) error
    GetLiveStream(liveID uint) (*dto.LiveDTO, error)
}

type PaymentServiceInterface interface {
    CreatePayment(userID uint, amount float64, paymentType string) (*dto.PaymentDTO, error)
    ProcessPayment(paymentID uint) error
    GetPaymentByID(paymentID uint) (*dto.PaymentDTO, error)
}
```

#### 2. 更新依賴注入容器
```go
// 修改 Container 結構
type Container struct {
    // 服務層 - 使用接口
    UserService         services.UserServiceInterface
    VideoService        services.VideoServiceInterface
    LiveService         services.LiveServiceInterface
    PaymentService      services.PaymentServiceInterface
    
    // 處理器層 - 使用接口
    UserHandler         *api.UserHandler
    VideoHandler        *api.VideoHandler
    LiveHandler         *api.LiveHandler
    PaymentHandler      *api.PaymentHandler
}
```

#### 3. 建立 TDD 開發環境
```bash
# 創建 TDD 開發腳本
#!/bin/bash
# scripts/tdd.sh

# TDD 開發流程
# 1. 寫測試 (Red)
# 2. 寫代碼 (Green)
# 3. 重構 (Refactor)

case $1 in
    "test")
        echo "🔴 寫測試階段"
        go test -v ./...
        ;;
    "code")
        echo "🟢 寫代碼階段"
        go build ./...
        go test -v ./...
        ;;
    "refactor")
        echo "🔄 重構階段"
        go vet ./...
        go fmt ./...
        go test -v ./...
        ;;
    *)
        echo "用法: $0 {test|code|refactor}"
        ;;
esac
```

### 第二階段：逐步轉換現有功能 (4-6 週)

#### 1. 用戶模組 TDD 重構
```bash
# 步驟 1: 為現有功能寫測試
./scripts/tdd.sh test

# 步驟 2: 確保測試失敗 (Red)
# 步驟 3: 實現最小功能 (Green)
# 步驟 4: 重構代碼 (Refactor)
```

#### 2. 影片模組 TDD 重構
```go
// 先寫測試
func TestVideoHandler_UploadVideo(t *testing.T) {
    // 測試用例
    tests := []struct {
        name           string
        file           *multipart.FileHeader
        userID         uint
        mockSetup      func()
        expectedStatus int
        expectedError  bool
    }{
        {
            name: "成功上傳影片",
            file: createMockFile("test.mp4"),
            userID: 1,
            mockSetup: func() {
                mockVideoService.On("UploadVideo", mock.Anything, uint(1)).
                    Return(&dto.VideoDTO{ID: 1, Title: "test.mp4"}, nil)
            },
            expectedStatus: http.StatusOK,
            expectedError:  false,
        },
        // 更多測試用例...
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // 實現測試
        })
    }
}
```

#### 3. 直播模組 TDD 重構
```go
// 先寫測試
func TestLiveHandler_CreateLiveStream(t *testing.T) {
    // 測試用例
    tests := []struct {
        name           string
        requestBody    request.CreateLiveRequest
        userID         uint
        mockSetup      func()
        expectedStatus int
        expectedError  bool
    }{
        {
            name: "成功創建直播",
            requestBody: request.CreateLiveRequest{
                Title:       "測試直播",
                Description: "這是一個測試直播",
            },
            userID: 1,
            mockSetup: func() {
                mockLiveService.On("CreateLiveStream", uint(1), "測試直播", "這是一個測試直播").
                    Return(&dto.LiveDTO{ID: 1, Title: "測試直播"}, nil)
            },
            expectedStatus: http.StatusOK,
            expectedError:  false,
        },
        // 更多測試用例...
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // 實現測試
        })
    }
}
```

### 第三階段：新功能開發 (持續)

#### 1. 新功能開發流程
```bash
# 1. 寫測試 (Red)
./scripts/tdd.sh test

# 2. 寫代碼 (Green)
./scripts/tdd.sh code

# 3. 重構 (Refactor)
./scripts/tdd.sh refactor
```

#### 2. 測試驅動的 API 設計
```go
// 先定義 API 測試
func TestAPI_UserRegistration(t *testing.T) {
    // 測試用戶註冊 API
    // 這會驅動 API 設計
}

func TestAPI_VideoUpload(t *testing.T) {
    // 測試影片上傳 API
    // 這會驅動 API 設計
}
```

## TDD 實施工具

### 1. 測試工具
- **Go Test**: 基礎測試框架
- **Testify**: 斷言和 Mock 框架
- **SQLMock**: 資料庫 Mock
- **HTTPTest**: HTTP 測試

### 2. 開發工具
- **GoLand/VS Code**: IDE 支持
- **Delve**: 調試工具
- **Air**: 熱重載

### 3. CI/CD 工具
- **GitHub Actions**: 自動化測試
- **Coverage**: 覆蓋率檢查
- **Linting**: 代碼質量檢查

## TDD 最佳實踐

### 1. 測試命名規範
```go
// 格式: Test[模組]_[功能]_[場景]
func TestUserService_Register_WithValidData(t *testing.T) {}
func TestUserService_Register_WithInvalidEmail(t *testing.T) {}
func TestUserService_Register_WithDuplicateUsername(t *testing.T) {}
```

### 2. 測試結構
```go
func TestFunction(t *testing.T) {
    // Arrange (準備)
    mockService := &mocks.MockService{}
    handler := NewHandler(mockService)
    
    // Act (執行)
    result, err := handler.DoSomething()
    
    // Assert (斷言)
    assert.NoError(t, err)
    assert.Equal(t, expected, result)
}
```

### 3. Mock 使用原則
```go
// 只 Mock 外部依賴
mockDB := &mocks.MockDatabase{}
mockRedis := &mocks.MockRedis{}

// 不要 Mock 業務邏輯
service := NewService(mockDB, mockRedis)
```

## 預期效果

### 短期效果 (1-2 個月)
- 測試覆蓋率提升到 50%+
- 代碼質量提升
- Bug 數量減少

### 中期效果 (3-6 個月)
- 測試覆蓋率達到 80%+
- 重構信心提升
- 開發速度穩定

### 長期效果 (6+ 個月)
- 測試覆蓋率達到 90%+
- 代碼維護性大幅提升
- 新功能開發效率提升

## 風險與挑戰

### 1. 學習成本
- 團隊需要學習 TDD 思維
- 初期開發速度可能下降

### 2. 重構阻力
- 現有代碼重構工作量較大
- 需要平衡新功能和重構

### 3. 工具適應
- 需要適應新的開發工具
- CI/CD 流程需要調整

## 成功指標

### 1. 技術指標
- 測試覆蓋率 > 80%
- 代碼重複率 < 5%
- 圈複雜度 < 10

### 2. 業務指標
- Bug 數量減少 50%
- 新功能開發時間穩定
- 代碼審查通過率提升

### 3. 團隊指標
- 團隊對 TDD 的接受度
- 代碼審查效率提升
- 新人上手速度提升

## 下一步行動

1. **立即開始**：完善接口定義
2. **本週目標**：建立 TDD 開發環境
3. **本月目標**：完成用戶模組 TDD 重構
4. **下月目標**：完成影片和直播模組重構 