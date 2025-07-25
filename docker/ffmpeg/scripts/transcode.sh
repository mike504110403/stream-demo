#!/bin/bash

# FFmpeg 影片轉碼腳本
# 支援：多品質 HLS 串流、MP4 轉換、縮圖生成

set -e

# 參數檢查
if [ "$#" -ne 4 ]; then
    echo "用法: $0 <input_key> <output_prefix> <user_id> <video_id>"
    echo "範例: $0 videos/original/1/input.mov videos/processed/1/1 1 1"
    exit 1
fi

INPUT_KEY="$1"
OUTPUT_PREFIX="$2"
USER_ID="$3"
VIDEO_ID="$4"

# 環境變數
MINIO_ENDPOINT="${MINIO_ENDPOINT:-http://minio:9000}"
MINIO_ACCESS_KEY="${MINIO_ACCESS_KEY:-minioadmin}"
MINIO_SECRET_KEY="${MINIO_SECRET_KEY:-minioadmin}"
MINIO_BUCKET="${MINIO_BUCKET:-stream-demo-videos}"
MINIO_PROCESSED_BUCKET="${MINIO_PROCESSED_BUCKET:-stream-demo-processed}"

# 工作目錄
WORK_DIR="/tmp/transcoding/${VIDEO_ID}"
mkdir -p "$WORK_DIR"

# MinIO Client 配置
echo "🔧 配置 MinIO Client..."
mc alias set s3 "$MINIO_ENDPOINT" "$MINIO_ACCESS_KEY" "$MINIO_SECRET_KEY"

# 下載原始文件
echo "📥 下載原始影片: $INPUT_KEY"
INPUT_FILE="$WORK_DIR/input.$(echo $INPUT_KEY | rev | cut -d. -f1 | rev)"
mc cp "s3/$MINIO_BUCKET/$INPUT_KEY" "$INPUT_FILE"

if [ ! -f "$INPUT_FILE" ]; then
    echo "❌ 下載失敗: $INPUT_FILE"
    exit 1
fi

echo "✅ 下載完成: $(ls -lh $INPUT_FILE)"

# 獲取影片資訊
echo "📊 分析影片資訊..."
ffprobe -v quiet -print_format json -show_format -show_streams "$INPUT_FILE" > "$WORK_DIR/info.json"

# 提取影片時長和尺寸
DURATION=$(ffprobe -v quiet -show_entries format=duration -of csv="p=0" "$INPUT_FILE")
WIDTH=$(ffprobe -v quiet -select_streams v:0 -show_entries stream=width -of csv="p=0" "$INPUT_FILE")
HEIGHT=$(ffprobe -v quiet -select_streams v:0 -show_entries stream=height -of csv="p=0" "$INPUT_FILE")

echo "📏 影片資訊: ${WIDTH}x${HEIGHT}, 時長: ${DURATION}秒"

# 創建輸出目錄
HLS_DIR="$WORK_DIR/hls"
MP4_DIR="$WORK_DIR/mp4"
THUMB_DIR="$WORK_DIR/thumbnails"

mkdir -p "$HLS_DIR" "$MP4_DIR" "$THUMB_DIR"

# 1. 生成 MP4 版本（網頁播放）
echo "🎬 轉換為 MP4..."
ffmpeg -i "$INPUT_FILE" \
    -c:v libx264 -profile:v high -level 4.0 \
    -c:a aac -ac 2 -b:a 128k \
    -movflags +faststart \
    -f mp4 \
    "$MP4_DIR/video.mp4" \
    -y

# 2. 生成多品質 HLS 串流
echo "📺 生成 HLS 串流..."

# 根據原始尺寸決定品質
declare -a QUALITIES=()

if [ "$HEIGHT" -ge 720 ]; then
    QUALITIES+=("720p:1280:720:2500k")
fi
if [ "$HEIGHT" -ge 480 ]; then
    QUALITIES+=("480p:854:480:1200k")
fi
QUALITIES+=("360p:640:360:800k")

# 生成 HLS 主播放列表
MASTER_PLAYLIST="$HLS_DIR/index.m3u8"
echo "#EXTM3U" > "$MASTER_PLAYLIST"
echo "#EXT-X-VERSION:3" >> "$MASTER_PLAYLIST"

for quality in "${QUALITIES[@]}"; do
    IFS=':' read -r name width height bitrate <<< "$quality"
    
    echo "🎯 生成 $name 品質..."
    
    # 創建品質目錄
    mkdir -p "$HLS_DIR/$name"
    
    # 計算緩衝區大小 (bitrate * 2)
    bitrate_num=$(echo "$bitrate" | sed 's/k//')
    bufsize=$((bitrate_num * 2))k
    bandwidth=$((bitrate_num * 1000))
    
    # 檢查是否有音頻軌道
    has_audio=$(ffprobe -v quiet -select_streams a:0 -show_entries stream=codec_type -of csv=p=0 "$INPUT_FILE" 2>/dev/null || echo "")
    
    # FFmpeg 轉碼 - 根據是否有音頻調整參數
    if [ -n "$has_audio" ]; then
        # 有音頻軌道
        ffmpeg -i "$INPUT_FILE" \
            -c:v libx264 -preset medium -profile:v high \
            -vf "scale=$width:-1" \
            -b:v "$bitrate" -maxrate "$bitrate" -bufsize "$bufsize" \
            -c:a aac -b:a 128k -ac 2 \
            -f hls \
            -hls_time 10 \
            -hls_list_size 0 \
            -hls_segment_filename "$HLS_DIR/$name/segment_%03d.ts" \
            "$HLS_DIR/$name/index.m3u8" \
            -y
    else
        # 沒有音頻軌道
        ffmpeg -i "$INPUT_FILE" \
            -c:v libx264 -preset medium -profile:v high \
            -vf "scale=$width:-1" \
            -b:v "$bitrate" -maxrate "$bitrate" -bufsize "$bufsize" \
            -an \
            -f hls \
            -hls_time 10 \
            -hls_list_size 0 \
            -hls_segment_filename "$HLS_DIR/$name/segment_%03d.ts" \
            "$HLS_DIR/$name/index.m3u8" \
            -y
    fi
    
    # 添加到主播放列表
    echo "#EXT-X-STREAM-INF:BANDWIDTH=$bandwidth,RESOLUTION=${width}x${height}" >> "$MASTER_PLAYLIST"
    echo "$name/index.m3u8" >> "$MASTER_PLAYLIST"
done

echo "✅ HLS 串流生成完成"

# 3. 生成縮圖
echo "🖼️ 生成縮圖..."

# 取中間時間點
THUMB_TIME=$(echo "$DURATION / 2" | bc)

# 生成多個縮圖
ffmpeg -i "$INPUT_FILE" -ss "$THUMB_TIME" -vframes 1 -f image2 -s 320x240 "$THUMB_DIR/thumb_320x240.jpg" -y
ffmpeg -i "$INPUT_FILE" -ss "$THUMB_TIME" -vframes 1 -f image2 -s 640x480 "$THUMB_DIR/thumb_640x480.jpg" -y
ffmpeg -i "$INPUT_FILE" -ss "$THUMB_TIME" -vframes 1 -f image2 -s 1280x720 "$THUMB_DIR/thumb_1280x720.jpg" -y

# 生成時間軸縮圖（每 10 秒一張）
INTERVAL=10
COUNT=0
for ((i=0; i<$(echo "$DURATION" | cut -d. -f1); i+=INTERVAL)); do
    ffmpeg -i "$INPUT_FILE" -ss "$i" -vframes 1 -f image2 -s 320x240 "$THUMB_DIR/timeline_${COUNT}.jpg" -y 2>/dev/null || true
    COUNT=$((COUNT + 1))
done

echo "✅ 縮圖生成完成"

# 4. 上傳處理後的文件到處理後桶
echo "📤 上傳處理後的文件到處理後桶..."
echo "🔧 使用處理後桶: $MINIO_PROCESSED_BUCKET"
echo "🔧 輸出前綴: $OUTPUT_PREFIX"

# 上傳 MP4
echo "📤 上傳 MP4..."
mc cp "$MP4_DIR/video.mp4" "s3/$MINIO_PROCESSED_BUCKET/${OUTPUT_PREFIX}/video.mp4"

# 上傳 HLS 文件
echo "📤 上傳 HLS 串流..."
mc cp --recursive "$HLS_DIR/" "s3/$MINIO_PROCESSED_BUCKET/${OUTPUT_PREFIX}/hls/"

# 上傳縮圖
echo "📤 上傳縮圖..."
mc cp --recursive "$THUMB_DIR/" "s3/$MINIO_PROCESSED_BUCKET/${OUTPUT_PREFIX}/thumbnails/"

# 生成轉碼報告
REPORT_FILE="$WORK_DIR/transcode_report.json"
cat > "$REPORT_FILE" << EOF
{
  "status": "completed",
  "input_file": "$INPUT_KEY",
  "output_prefix": "$OUTPUT_PREFIX",
  "original_info": {
    "duration": $DURATION,
    "width": $WIDTH,
    "height": $HEIGHT,
    "file_size": $(stat -f%z "$INPUT_FILE" 2>/dev/null || stat -c%s "$INPUT_FILE")
  },
  "outputs": {
    "mp4": "${OUTPUT_PREFIX}/video.mp4",
    "hls_master": "${OUTPUT_PREFIX}/hls/index.m3u8",
    "thumbnail": "${OUTPUT_PREFIX}/thumbnails/thumb_640x480.jpg"
  },
  "qualities": [$(printf '"%s",' "${QUALITIES[@]}" | sed 's/,$//')],
  "completed_at": "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
}
EOF

# 上傳報告到處理後桶
mc cp "$REPORT_FILE" "s3/$MINIO_PROCESSED_BUCKET/${OUTPUT_PREFIX}/transcode_report.json"

# 清理工作目錄
echo "🧹 清理臨時文件..."
rm -rf "$WORK_DIR"

echo "🎉 轉碼完成！"
echo "📺 HLS 主播放列表: ${OUTPUT_PREFIX}/hls/index.m3u8"
echo "🎬 MP4 影片: ${OUTPUT_PREFIX}/video.mp4"
echo "🖼️ 縮圖: ${OUTPUT_PREFIX}/thumbnails/thumb_640x480.jpg" 