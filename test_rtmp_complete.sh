#!/bin/bash

# 完整的 RTMP 推流測試
echo "🎬 完整 RTMP 推流測試"
echo "===================="

# 檢查服務狀態
echo "📡 檢查服務狀態..."

# 檢查後端
if ! curl -s http://localhost:8080/api/live-rooms > /dev/null 2>&1; then
    echo "❌ 後端服務未運行，請先啟動後端"
    exit 1
fi

# 檢查 NGINX RTMP
if ! curl -s http://localhost:1935/stat > /dev/null 2>&1; then
    echo "❌ NGINX RTMP 服務未運行"
    exit 1
fi

# 檢查 Stream Puller
if ! curl -s http://localhost:8083/health > /dev/null 2>&1; then
    echo "❌ Stream Puller 服務未運行"
    exit 1
fi

echo "✅ 所有服務運行正常"
echo ""

echo "🎯 完整推流流程："
echo "1. 前端創建直播間 → 獲取 stream_key"
echo "2. OBS 推流到 rtmp://localhost:1935/live/[stream_key]"
echo "3. NGINX RTMP 接收推流 → 自動生成 HLS"
echo "4. Stream Puller 監控 HLS 文件"
echo "5. 前端播放 http://localhost:8083/[stream_key]/index.m3u8"
echo ""

echo "🔧 技術架構："
echo "OBS → RTMP (1935) → NGINX → HLS → Stream Puller (8083) → 前端播放器"
echo ""

echo "📊 服務端口："
echo "- RTMP 推流: rtmp://localhost:1935/live"
echo "- HLS 播放: http://localhost:8083"
echo "- NGINX 狀態: http://localhost:1935/stat"
echo ""

echo "🧪 測試步驟："
echo "1. 啟動前端和後端："
echo "   cd backend && go run main.go"
echo "   cd frontend && npm run dev"
echo ""
echo "2. 登入前端: http://localhost:5173"
echo "3. 創建直播間，複製推流資訊"
echo "4. 使用 OBS 推流："
echo "   - 推流地址: rtmp://localhost:1935/live"
echo "   - 串流金鑰: [你的 stream_key]"
echo "5. 在直播間中點擊'開始直播'"
echo "6. 檢查是否能正常播放"
echo ""

echo "📈 監控命令："
echo "- 查看 NGINX 狀態: curl http://localhost:1935/stat"
echo "- 查看 Stream Puller 日誌: docker-compose logs stream-puller"
echo "- 查看 NGINX 日誌: docker-compose logs nginx-rtmp"
echo ""

echo "🚀 準備就緒！請開始測試。" 