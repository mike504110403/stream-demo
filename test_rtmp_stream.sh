#!/bin/bash

# 測試 RTMP 推流功能
echo "🎬 測試 RTMP 推流功能"
echo "===================="

# 檢查服務狀態
echo "📡 檢查服務狀態..."
if ! curl -s http://localhost:8080/api/live-rooms > /dev/null 2>&1; then
    echo "❌ 後端服務未運行，請先啟動後端"
    exit 1
fi

if ! curl -s http://localhost:8083/health > /dev/null 2>&1; then
    echo "❌ Stream Puller 服務未運行，請先啟動"
    echo "   ./docker-manage.sh stream-puller start"
    exit 1
fi

echo "✅ 所有服務運行正常"
echo ""

echo "🎯 RTMP 推流測試步驟："
echo "1. 登入前端: http://localhost:5173"
echo "2. 創建一個直播間"
echo "3. 複製推流密鑰 (stream_key)"
echo "4. 使用 OBS 或其他推流軟體："
echo "   - 推流地址: rtmp://localhost:1935/live"
echo "   - 串流金鑰: [你的推流密鑰]"
echo "5. 開始推流"
echo "6. 在直播間中點擊'開始直播'"
echo "7. 檢查是否能正常播放"
echo ""

echo "🔧 技術說明："
echo "- RTMP 推流到 rtmp://localhost:1935/live"
echo "- Stream Puller 自動拉取並轉換為 HLS"
echo "- 前端通過 HLS 播放 (http://localhost:8083/[stream_key]/index.m3u8)"
echo ""

echo "📊 檢查當前直播流："
if docker-compose ps stream-puller | grep -q "Up"; then
    streams=$(docker exec stream-demo-stream-puller ls /tmp/public_streams/ 2>/dev/null || true)
    if [ -n "$streams" ]; then
        echo "當前直播流："
        for stream in $streams; do
            echo "  - $stream"
            echo "    HLS: http://localhost:8083/$stream/index.m3u8"
        done
    else
        echo "目前沒有直播流"
    fi
else
    echo "Stream Puller 容器未運行"
fi

echo ""
echo "🚀 準備就緒！請開始測試 RTMP 推流。" 