# 🎬 OBS 推流設置指南

## 📋 概述

本指南將教您如何使用 OBS Studio 或其他推流軟體向我們的串流平台推流。

## 🎯 推流架構

```
OBS Studio → RTMP (1935) → nginx-rtmp → stream-puller → HLS (8083) → 前端播放器
```

## 📱 步驟 1: 獲取推流資訊

### 1.1 登入平台
1. 打開瀏覽器，訪問: http://localhost:5173
2. 登入您的帳號

### 1.2 創建或進入直播間
1. 點擊"直播間" → "創建直播間"
2. 填寫直播間標題和描述
3. 點擊"創建"

### 1.3 獲取推流資訊
1. 在直播間頁面，點擊"串流資訊"按鈕
2. 複製以下資訊：
   - **串流金鑰**: `stream_xxxxxxxx` (每個直播間唯一)
   - **RTMP 推流地址**: `rtmp://localhost:1935/live/stream_xxxxxxxx`

## 🖥️ 步驟 2: OBS Studio 設置

### 2.1 下載並安裝 OBS Studio
- 官方網站: https://obsproject.com/
- 下載適合您作業系統的版本

### 2.2 基本設置
1. **打開 OBS Studio**
2. **設置 → 串流**
3. **服務**: 選擇"自訂"
4. **伺服器**: 填入 `rtmp://localhost:1935/live`
5. **串流金鑰**: 填入您的串流金鑰（如：`stream_xxxxxxxx`）

### 2.3 視頻設置（推薦）
1. **設置 → 視頻**
2. **基礎解析度**: 1920x1080
3. **輸出解析度**: 1280x720
4. **縮放過濾器**: Lanczos
5. **FPS**: 30

### 2.4 音頻設置
1. **設置 → 音頻**
2. **採樣率**: 44.1kHz
3. **聲道**: 立體聲

### 2.5 輸出設置
1. **設置 → 輸出**
2. **輸出模式**: 進階
3. **編碼器**: x264
4. **比特率**: 2500 Kbps
5. **關鍵幀間隔**: 2
6. **CPU 使用預設**: veryfast
7. **配置**: main
8. **調優**: zerolatency

## 🎬 步驟 3: 添加場景和來源

### 3.1 添加視頻來源
1. 在"來源"區域點擊"+"號
2. 選擇"視頻捕獲設備"（攝像頭）
3. 選擇您的攝像頭
4. 調整位置和大小

### 3.2 添加音頻來源
1. 在"來源"區域點擊"+"號
2. 選擇"音頻輸入捕獲"
3. 選擇您的麥克風

### 3.3 添加桌面捕獲（可選）
1. 在"來源"區域點擊"+"號
2. 選擇"顯示器捕獲"
3. 選擇要捕獲的顯示器

## 🚀 步驟 4: 開始推流

### 4.1 測試推流
1. 在 OBS 中點擊"開始串流"
2. 觀察狀態欄是否顯示"串流中"
3. 如果出現錯誤，檢查設置

### 4.2 在平台開始直播
1. 回到瀏覽器中的直播間
2. 點擊"開始直播"按鈕
3. 等待 5-10 秒鐘
4. 直播畫面應該會出現在播放器中

### 4.3 檢查直播狀態
- 前端播放器應該顯示您的直播畫面
- 聊天室應該顯示"直播已開始"通知
- 觀眾可以加入直播間觀看

## 🛠️ 其他推流軟體

### Streamlabs OBS
設置方式與 OBS Studio 完全相同：
1. 設置 → 串流
2. 服務: 自訂
3. 伺服器: `rtmp://localhost:1935/live`
4. 串流金鑰: 您的 stream_key

### XSplit Broadcaster
1. 工具 → 帳號
2. 添加自訂串流服務
3. 伺服器: `rtmp://localhost:1935/live`
4. 串流金鑰: 您的 stream_key

### 手機 App
支援 RTMP 推流的手機 App：
- **Larix Broadcaster** (iOS/Android)
- **OBS Camera** (iOS/Android)
- **BroadcastMe** (iOS)

設置方式：
- 伺服器: `rtmp://localhost:1935/live`
- 串流金鑰: 您的 stream_key

## 🔧 故障排除

### 常見問題

#### 1. 推流失敗
**症狀**: OBS 顯示"串流失敗"
**解決方案**:
- 檢查 nginx-rtmp 服務是否運行: `docker-compose ps nginx-rtmp`
- 檢查端口 1935 是否開放: `lsof -i :1935`
- 重啟服務: `./docker-manage.sh restart nginx-rtmp`

#### 2. 前端無法播放
**症狀**: 推流成功但前端播放器黑屏
**解決方案**:
- 檢查 stream-puller 服務: `docker-compose ps stream-puller`
- 查看轉換日誌: `docker-compose logs stream-puller`
- 等待 10-15 秒讓轉換完成

#### 3. 延遲過高
**症狀**: 直播延遲超過 5 秒
**解決方案**:
- 降低 OBS 比特率到 1500 Kbps
- 使用 "ultrafast" CPU 預設
- 檢查網路連接

#### 4. 畫質不佳
**症狀**: 直播畫質模糊
**解決方案**:
- 提高 OBS 比特率到 3000-4000 Kbps
- 使用 "faster" CPU 預設
- 確保輸出解析度為 1280x720 或更高

### 監控命令

```bash
# 檢查服務狀態
./docker-manage.sh status

# 查看 nginx-rtmp 日誌
docker-compose logs nginx-rtmp

# 查看 stream-puller 日誌
docker-compose logs stream-puller

# 檢查 RTMP 狀態
curl http://localhost:1935/stat

# 檢查 HLS 文件
ls -la /tmp/public_streams/
```

## 📊 最佳實踐

### 1. 網路設置
- 使用有線網路而非 WiFi
- 確保上傳頻寬至少 5 Mbps
- 關閉其他佔用頻寬的應用

### 2. 硬體設置
- 使用 USB 3.0 攝像頭
- 使用外接麥克風
- 確保電腦有足夠的 CPU 和記憶體

### 3. 軟體設置
- 定期更新 OBS Studio
- 使用推薦的編碼設置
- 測試推流前先進行本地錄製測試

### 4. 直播準備
- 提前 10 分鐘開始推流測試
- 準備備用網路連接
- 測試音頻和視頻設備

## 🎯 下一步

成功設置推流後，您可以：
1. 邀請觀眾加入直播間
2. 使用聊天功能與觀眾互動
3. 嘗試不同的場景和來源設置
4. 探索更多 OBS 功能（如綠幕、濾鏡等）

祝您直播愉快！🎉 