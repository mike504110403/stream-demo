#!/bin/bash

# 手動啟動 RTMP 到 HLS 轉換
STREAM_KEY="stream_ca6f8c94-8c1"

echo "🔄 啟動 RTMP 到 HLS 轉換: $STREAM_KEY"

# 在 stream-puller 容器中啟動 FFmpeg 轉換
docker exec stream-demo-stream-puller ffmpeg \
    -i "rtmp://nginx-rtmp:1935/live/$STREAM_KEY" \
    -c:v libx264 \
    -preset ultrafast \
    -c:a aac \
    -b:a 128k \
    -f hls \
    -hls_time 2 \
    -hls_list_size 6 \
    -hls_flags delete_segments+independent_segments \
    -hls_segment_type mpegts \
    -hls_segment_filename "/tmp/public_streams/$STREAM_KEY/segment_%03d.ts" \
    -hls_playlist_type event \
    "/tmp/public_streams/$STREAM_KEY/index.m3u8" &

echo "✅ RTMP 轉換已啟動"
echo "📺 播放地址: http://localhost:8083/$STREAM_KEY/index.m3u8" 