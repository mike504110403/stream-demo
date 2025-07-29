#!/bin/bash

# 直播系統測試腳本
# 測試重構後的直播系統功能

set -e

# 顏色定義
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日誌函數
log_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

log_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

log_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

log_error() {
    echo -e "${RED}❌ $1${NC}"
}

echo "🎬 直播系統測試開始..."
echo "================================"

# 1. 檢查基礎服務
log_info "1. 檢查基礎服務狀態..."
if docker ps --format "{{.Names}}" | grep -q "stream-demo-postgres"; then
    log_success "PostgreSQL 運行中"
else
    log_error "PostgreSQL 未運行"
    exit 1
fi

if docker ps --format "{{.Names}}" | grep -q "stream-demo-redis"; then
    log_success "Redis 運行中"
else
    log_error "Redis 未運行"
    exit 1
fi

if docker ps --format "{{.Names}}" | grep -q "stream-demo-minio"; then
    log_success "MinIO 運行中"
else
    log_error "MinIO 未運行"
    exit 1
fi

# 2. 檢查 Stream Puller 服務
log_info "2. 檢查 Stream Puller 服務..."
if curl -s http://localhost:8083/health > /dev/null; then
    log_success "Stream Puller 健康檢查通過"
else
    log_error "Stream Puller 健康檢查失敗"
    exit 1
fi

# 3. 檢查後端 API
log_info "3. 檢查後端 API..."
if curl -s http://localhost:8080/api/public-streams | jq -e '.success' > /dev/null; then
    log_success "後端 API 正常"
else
    log_error "後端 API 異常"
    exit 1
fi

# 4. 檢查前端代理
log_info "4. 檢查前端代理..."
if curl -s http://localhost:5173/api/public-streams | jq -e '.success' > /dev/null; then
    log_success "前端代理正常"
else
    log_error "前端代理異常"
    exit 1
fi

# 5. 檢查 HLS 流
log_info "5. 檢查 HLS 流..."
if curl -s -I http://localhost:8083/tears_of_steel/index.m3u8 | grep -q "200 OK"; then
    log_success "tears_of_steel 流正常"
else
    log_warning "tears_of_steel 流異常"
fi

if curl -s -I http://localhost:8083/mux_test/index.m3u8 | grep -q "200 OK"; then
    log_success "mux_test 流正常"
else
    log_warning "mux_test 流異常"
fi

# 6. 檢查前端服務
log_info "6. 檢查前端服務..."
if curl -s -I http://localhost:5173 | grep -q "200 OK"; then
    log_success "前端服務正常"
else
    log_error "前端服務異常"
    exit 1
fi

# 7. 檢查 FFmpeg 進程
log_info "7. 檢查 FFmpeg 進程..."
ffmpeg_count=$(ps aux | grep ffmpeg | grep -v grep | wc -l)
if [ "$ffmpeg_count" -ge 1 ]; then
    log_success "FFmpeg 進程運行中 ($ffmpeg_count 個)"
else
    log_warning "沒有 FFmpeg 進程運行"
fi

# 8. 檢查輸出目錄
log_info "8. 檢查輸出目錄..."
if [ -d "/tmp/public_streams" ]; then
    stream_count=$(ls /tmp/public_streams/ | wc -l)
    log_success "輸出目錄存在 ($stream_count 個流目錄)"
    
    # 檢查具體的流目錄
    for stream in tears_of_steel mux_test; do
        if [ -f "/tmp/public_streams/$stream/index.m3u8" ]; then
            log_success "$stream 流文件存在"
        else
            log_warning "$stream 流文件不存在"
        fi
    done
else
    log_error "輸出目錄不存在"
fi

# 9. 測試 API 響應
log_info "9. 測試 API 響應..."
api_response=$(curl -s http://localhost:8080/api/public-streams)
stream_count=$(echo "$api_response" | jq '.data.total')
log_success "API 返回 $stream_count 個流"

# 10. 顯示服務端口
log_info "10. 服務端口信息..."
echo "  後端 API: http://localhost:8080"
echo "  前端服務: http://localhost:5173"
echo "  Stream Puller: http://localhost:8083"
echo "  MinIO Console: http://localhost:9001"
echo "  PostgreSQL: localhost:5432"
echo "  Redis: localhost:6379"

echo ""
echo "🎉 直播系統測試完成！"
echo "================================"
echo ""
echo "📺 可以訪問以下頁面測試："
echo "  - 前端首頁: http://localhost:5173"
echo "  - 公開流列表: http://localhost:5173/public-streams"
echo "  - 直接播放流: http://localhost:8083/tears_of_steel/index.m3u8"
echo ""
echo "🔧 管理命令："
echo "  - 查看狀態: ./docker-manage.sh status"
echo "  - 重啟服務: ./docker-manage.sh restart"
echo "  - 查看日誌: tail -f backend/cmd/stream_puller/stream-puller.log" 