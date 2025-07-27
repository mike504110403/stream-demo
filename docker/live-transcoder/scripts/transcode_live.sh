#!/bin/bash

# 直播轉碼服務 - 接收公開 RTMP 流 + 外部 RTMP 推流
set -e

# 環境變數
HLS_OUTPUT_DIR=${HLS_OUTPUT_DIR:-"/tmp/live"}
MINIO_ENDPOINT=${MINIO_ENDPOINT:-"http://minio:9000"}
MINIO_ACCESS_KEY=${MINIO_ACCESS_KEY:-"minioadmin"}
MINIO_SECRET_KEY=${MINIO_SECRET_KEY:-"minioadmin"}
MINIO_LIVE_BUCKET=${MINIO_LIVE_BUCKET:-"stream-demo-live"}
RTMP_SERVER=${RTMP_SERVER:-"rtmp://nginx-rtmp:1935"}
NGINX_RTMP_STATUS_URL=${NGINX_RTMP_STATUS_URL:-"http://nginx-rtmp:8080/stat"}

# 讀取公開 RTMP 流配置
load_stream_config() {
    local config_file="/app/config/streams.conf"
    
    if [ -f "$config_file" ]; then
        echo "📋 載入流配置: $config_file"
        while IFS='=' read -r stream_name rtmp_url; do
            # 跳過註釋和空行
            if [[ ! "$stream_name" =~ ^#.*$ ]] && [[ -n "$stream_name" ]]; then
                PUBLIC_STREAMS["$stream_name"]="$rtmp_url"
                echo "  - $stream_name: $rtmp_url"
            fi
        done < "$config_file"
    else
        echo "⚠️  配置文件不存在，使用預設配置"
        # 預設配置
        PUBLIC_STREAMS["test"]="rtmp://localhost:1935/live/test"
    fi
}

# 初始化公開流配置
declare -A PUBLIC_STREAMS
load_stream_config

# 配置 MinIO 客戶端
configure_minio() {
    echo "🔧 配置 MinIO 客戶端..."
    mc alias set s3 $MINIO_ENDPOINT $MINIO_ACCESS_KEY $MINIO_SECRET_KEY
    
    # 創建直播桶
    mc mb s3/$MINIO_LIVE_BUCKET --ignore-existing
    mc policy set download s3/$MINIO_LIVE_BUCKET
}

# 啟動 HTTP 服務（用於 HLS 分發）
start_http_server() {
    echo "🌐 啟動 HTTP 服務..."
    
    # 創建簡單的 HTTP 服務器腳本
    cat > /tmp/simple_server.py << 'EOF'
#!/usr/bin/env python3
import http.server
import socketserver
import os

class SimpleHTTPRequestHandler(http.server.SimpleHTTPRequestHandler):
    def do_GET(self):
        if self.path == '/health':
            self.send_response(200)
            self.send_header('Content-Type', 'text/plain')
            self.end_headers()
            self.wfile.write(b'healthy')
        else:
            # 設置 HLS 相關 headers
            self.send_header('Access-Control-Allow-Origin', '*')
            self.send_header('Cache-Control', 'no-cache')
            super().do_GET()
    
    def log_message(self, format, *args):
        # 減少日誌輸出
        pass

if __name__ == '__main__':
    os.chdir('/tmp/live')
    httpd = socketserver.TCPServer(("", 8080), SimpleHTTPRequestHandler)
    print("Simple HTTP server started on port 8080")
    httpd.serve_forever()
EOF

    # 啟動 HTTP 服務器
    python3 /tmp/simple_server.py &
    
    # 等待服務啟動
    sleep 2
}

# 轉碼 RTMP 流
transcode_stream() {
    local stream_name=$1
    local rtmp_url=$2
    local output_dir="$HLS_OUTPUT_DIR/$stream_name"
    
    echo "🎬 開始轉碼流: $stream_name ($rtmp_url)"
    
    # 創建輸出目錄
    mkdir -p "$output_dir"
    
    # 啟動 FFmpeg 轉碼
    ffmpeg -i "$rtmp_url" \
        -c:v libx264 -preset ultrafast -tune zerolatency \
        -c:a aac -b:a 128k \
        -f hls \
        -hls_time 2 \
        -hls_list_size 10 \
        -hls_flags delete_segments \
        -hls_segment_filename "$output_dir/segment_%03d.ts" \
        "$output_dir/index.m3u8" \
        -y
    
    echo "✅ 流 $stream_name 轉碼完成"
}

# 轉碼公開 RTMP 流
transcode_public_stream() {
    local stream_name=$1
    local rtmp_url=$2
    transcode_stream "$stream_name" "$rtmp_url"
}

# 轉碼 RTMP 推流
transcode_rtmp_push() {
    local stream_name=$1
    local rtmp_url="$RTMP_SERVER/live/$stream_name"
    transcode_stream "$stream_name" "$rtmp_url"
}

# 獲取 RTMP 推流列表
get_rtmp_streams() {
    local streams=()
    
    # 嘗試從 nginx-rtmp 獲取狀態
    if curl -s "$NGINX_RTMP_STATUS_URL" > /tmp/nginx_stat.xml 2>/dev/null; then
        # 檢查是否有活躍的推流
        local has_streams=$(grep -c '<stream>' /tmp/nginx_stat.xml 2>/dev/null || echo "0")
        
        if [ "$has_streams" -gt 0 ]; then
            # 解析 XML 獲取流名稱
            if command -v xmllint > /dev/null; then
                # 使用 xmllint 解析 XML
                local stream_names=$(xmllint --xpath "//stream/name/text()" /tmp/nginx_stat.xml 2>/dev/null || true)
                if [ -n "$stream_names" ]; then
                    for name in $stream_names; do
                        # 過濾掉已轉碼的流
                        if [[ "$name" != "live_transcoded"* ]]; then
                            streams+=("$name")
                        fi
                    done
                fi
            else
                # 使用 grep 簡單解析
                local stream_names=$(grep -o 'name="[^"]*"' /tmp/nginx_stat.xml | cut -d'"' -f2 | grep -v "live_transcoded" || true)
                if [ -n "$stream_names" ]; then
                    for name in $stream_names; do
                        streams+=("$name")
                    done
                fi
            fi
        else
            # 如果沒有活躍流，檢查是否有最近的推流記錄
            # 這裡可以添加更複雜的檢測邏輯
            echo "📊 沒有檢測到活躍的 RTMP 流" >&2
        fi
    else
        echo "❌ 無法連接到 Nginx-RTMP 狀態頁面" >&2
    fi
    
    # 只返回流名稱數組
    printf '%s\n' "${streams[@]}"
}

# 監控並啟動轉碼
monitor_and_transcode() {
    echo "📡 開始監控 RTMP 流..."
    
    # 記錄已啟動的流
    declare -A active_streams
    
    # 啟動所有配置的公開流
    for stream_name in "${!PUBLIC_STREAMS[@]}"; do
        rtmp_url="${PUBLIC_STREAMS[$stream_name]}"
        
        echo "🎬 啟動公開流轉碼: $stream_name"
        transcode_public_stream "$stream_name" "$rtmp_url" &
        active_streams["$stream_name"]=1
    done
    
    # 監控轉碼進程和 RTMP 推流
    while true; do
        echo "🔄 監控循環執行中..."
        
        # 檢查公開流轉碼進程
        for stream_name in "${!PUBLIC_STREAMS[@]}"; do
            if ! pgrep -f "ffmpeg.*$stream_name" > /dev/null; then
                echo "🔄 重新啟動公開流轉碼: $stream_name"
                rtmp_url="${PUBLIC_STREAMS[$stream_name]}"
                transcode_public_stream "$stream_name" "$rtmp_url" &
            fi
        done
        
        # 檢查 RTMP 推流
        echo "🔍 檢查 RTMP 推流..."
        local rtmp_streams=($(get_rtmp_streams))
        echo "📊 發現的 RTMP 流: ${rtmp_streams[*]}"
        
        for stream_name in "${rtmp_streams[@]}"; do
            if [[ -z "${active_streams[$stream_name]}" ]]; then
                echo "🎬 發現新的 RTMP 推流: $stream_name"
                transcode_rtmp_push "$stream_name" &
                active_streams["$stream_name"]=1
            fi
        done
        
        # 清理已停止的流
        for stream_name in "${!active_streams[@]}"; do
            if [[ ! " ${rtmp_streams[@]} " =~ " ${stream_name} " ]] && [[ ! "${PUBLIC_STREAMS[$stream_name]}" ]]; then
                echo "🧹 清理已停止的流: $stream_name"
                unset active_streams["$stream_name"]
                # 停止對應的 FFmpeg 進程
                pkill -f "ffmpeg.*$stream_name" 2>/dev/null || true
            fi
        done
        
        echo "⏰ 等待 10 秒..."
        sleep 10
    done
}

# 列出可用的直播流
list_streams() {
    echo "📋 可用的直播流:"
    for stream_name in "${!PUBLIC_STREAMS[@]}"; do
        echo "  - $stream_name: ${PUBLIC_STREAMS[$stream_name]}"
        echo "    HLS: http://localhost:8080/$stream_name/index.m3u8"
    done
    echo ""
    echo "📡 RTMP 推流地址: $RTMP_SERVER/live/[stream_key]"
    echo "📺 HLS 播放地址: http://localhost:8080/[stream_name]/index.m3u8"
}

# 主函數
main() {
    echo "🚀 啟動 RTMP 流轉碼服務..."
    
    # 配置 MinIO
    configure_minio
    
    # 啟動 HTTP 服務
    start_http_server
    
    # 列出可用流
    list_streams
    
    # 開始監控和轉碼
    monitor_and_transcode
}

# 執行主函數
main "$@" 