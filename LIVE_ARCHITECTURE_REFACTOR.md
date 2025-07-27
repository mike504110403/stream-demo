# 直播架構重構方案

## 問題分析

您提到的三個問題都很關鍵：

### 1. 一般上線直播服務的架構選擇

**一般上線的直播服務通常不會完全自己做轉檔及推流拉流**，而是會選擇：

- **雲端直播服務**：AWS MediaLive、阿里雲直播、騰訊雲直播等
- **CDN 服務商**：Cloudflare、Akamai 等提供直播加速
- **專業直播平台**：Twitch、YouTube Live 等

**自建直播服務的情況**：
- 成本敏感的中小企業
- 需要完全控制數據和流程
- 特殊行業需求（如教育、企業內訓）

### 2. 關於推流測試

**您說得對！** 確實有免費公開的流可以拉來測試：

- **公開 HLS 流**：NASA、新聞台等
- **測試流**：Big Buck Bunny 等公開測試影片
- **本地測試**：使用 FFmpeg 生成測試流

### 3. 架構重構建議

您的想法很好！建議將直播相關服務整合進專案，並做成可配置的模組。

## 重構方案

### 1. 配置驅動的服務選擇

```yaml
# config.local.yaml
live:
  enabled: true
  type: "local"  # local, cloud, hybrid
  local:
    enabled: true
    rtmp_server: "localhost"
    rtmp_server_port: 1935
    transcoder_enabled: true
    hls_output_dir: "/tmp/live"
    http_port: 8081
  cloud:
    provider: "aws"  # aws, aliyun, tencent
    rtmp_ingest_url: ""
    hls_playback_url: ""
    api_key: ""
    api_secret: ""
    transcode_enabled: false
```

### 2. 服務層抽象

```go
// LiveService 介面
type LiveService interface {
    Start() error
    Stop() error
    GetStreamURL(streamKey string) (string, error)
    GetPushURL(streamKey string) (string, error)
    CheckStreamStatus(streamKey string) (bool, error)
    GetActiveStreams() ([]string, error)
}

// 本地實現
type LocalLiveService struct {
    config     LocalLiveConfig
    rtmpServer *RTMPServer
    transcoder *LiveTranscoder
}

// 雲端實現
type CloudLiveService struct {
    config CloudLiveConfig
}
```

### 3. 工廠模式創建服務

```go
func LiveServiceFactory(serviceType string, config interface{}) (LiveService, error) {
    switch serviceType {
    case "local":
        return NewLocalLiveService(config.(LocalLiveConfig)), nil
    case "cloud":
        return NewCloudLiveService(config.(CloudLiveConfig)), nil
    default:
        return nil, fmt.Errorf("不支援的直播服務類型: %s", serviceType)
    }
}
```

## 優勢

### 1. 開發環境
- **本地模式**：使用 Docker 容器提供 RTMP 和轉碼服務
- **快速測試**：使用公開測試流驗證功能
- **成本控制**：避免雲端服務費用

### 2. 生產環境
- **雲端模式**：使用專業直播服務
- **混合模式**：本地 + 雲端備援
- **靈活切換**：根據需求選擇服務

### 3. 維護性
- **統一介面**：所有直播功能通過相同介面
- **配置驅動**：通過配置文件控制行為
- **模組化**：各組件獨立，易於維護

## 實施步驟

### 第一階段：本地整合
1. ✅ 創建直播服務介面
2. ✅ 實現本地直播服務
3. ✅ 整合到現有服務架構
4. ✅ 配置驅動的服務選擇

### 第二階段：雲端整合
1. 🔄 實現雲端直播服務
2. 🔄 添加混合模式支援
3. 🔄 實現服務切換邏輯

### 第三階段：優化
1. 🔄 添加監控和日誌
2. 🔄 實現自動故障轉移
3. 🔄 性能優化

## 測試方案

### 1. 公開測試流
```bash
# 測試公開 HLS 流
./test_public_streams.sh
```

### 2. 本地推流測試
```bash
# 使用公開影片推流到本地 RTMP
ffmpeg -re -i /tmp/test_video.mp4 \
       -c:v libx264 -preset ultrafast \
       -c:a aac -b:a 128k \
       -f flv rtmp://localhost:1935/live/test
```

### 3. 可用的測試流
- **Big Buck Bunny**: https://commondatastorage.googleapis.com/gtv-videos-bucket/sample/BigBuckBunny.mp4
- **Elephants Dream**: https://commondatastorage.googleapis.com/gtv-videos-bucket/sample/ElephantsDream.mp4
- **Sintel**: https://commondatastorage.googleapis.com/gtv-videos-bucket/sample/Sintel.mp4

## 配置示例

### 本地開發配置
```yaml
live:
  enabled: true
  type: "local"
  local:
    enabled: true
    rtmp_server: "localhost"
    rtmp_server_port: 1935
    transcoder_enabled: true
    hls_output_dir: "/tmp/live"
    http_port: 8081
```

### 雲端生產配置
```yaml
live:
  enabled: true
  type: "cloud"
  cloud:
    provider: "aws"
    rtmp_ingest_url: "rtmp://your-rtmp-endpoint.amazonaws.com/live"
    hls_playback_url: "https://your-cdn.amazonaws.com/live"
    api_key: "your-api-key"
    api_secret: "your-api-secret"
    transcode_enabled: true
```

### 混合模式配置
```yaml
live:
  enabled: true
  type: "hybrid"
  hybrid:
    local_enabled: true
    cloud_enabled: true
    fallback_to_local: true
    cloud_provider: "aws"
```

## 總結

這個重構方案解決了您提到的所有問題：

1. **成本控制**：開發時使用本地服務，生產時可選擇雲端
2. **測試便利**：使用公開測試流，無需自己推流
3. **維護性**：模組化設計，易於維護和擴展
4. **靈活性**：配置驅動，可根據需求切換服務

這樣的架構既適合開發環境的快速迭代，也適合生產環境的穩定運行。 