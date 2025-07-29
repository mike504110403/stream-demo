#!/bin/bash

# 簡化的 Docker 管理腳本
set -e

# 顏色定義
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 函數定義
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

# 檢查 Docker 是否運行
check_docker() {
    if ! docker info > /dev/null 2>&1; then
        log_error "Docker 未運行，請先啟動 Docker"
        exit 1
    fi
}

# 顯示幫助信息
show_help() {
    echo "🚀 Stream Demo Docker 管理腳本"
    echo ""
    echo "用法: $0 [命令]"
    echo ""
    echo "命令:"
    echo "  start     啟動所有服務"
    echo "  stop      停止所有服務"
    echo "  restart   重啟所有服務"
    echo "  status    查看服務狀態"
    echo "  logs      查看服務日誌"
    echo "  build     重新構建服務"
    echo "  clean     清理容器和映像"
    echo "  init      初始化 MinIO 桶"
    echo "  init-live 初始化直播服務"
    echo "  live-status 查看直播狀態"
    echo "  stream-puller 管理流拉取服務"
    echo "  test      運行 Go 測試"
    echo "  help      顯示此幫助信息"
    echo ""
    echo "流拉取服務命令:"
    echo "  stream-puller start    啟動流拉取服務"
    echo "  stream-puller stop     停止流拉取服務"
    echo "  stream-puller restart  重啟流拉取服務"
    echo "  stream-puller status   查看流拉取服務狀態"
    echo "  stream-puller logs     查看流拉取服務日誌"
    echo "  stream-puller test     測試流播放"
    echo ""
    echo "範例:"
    echo "  $0 start    # 啟動所有服務"
    echo "  $0 logs     # 查看日誌"
    echo "  $0 status   # 查看狀態"
}

# 啟動服務
start_services() {
    log_info "啟動所有服務..."
    docker-compose up -d
    log_success "服務啟動完成"
    
    # 等待服務啟動
    log_info "等待服務啟動..."
    sleep 10
    
    # 檢查服務狀態
    check_services_status
}

# 停止服務
stop_services() {
    log_info "停止所有服務..."
    docker-compose down
    log_success "服務停止完成"
}

# 重啟服務
restart_services() {
    log_info "重啟所有服務..."
    docker-compose restart
    log_success "服務重啟完成"
}

# 檢查服務狀態
check_services_status() {
    log_info "檢查服務狀態..."
    
    # 檢查容器狀態
    echo ""
    echo "📊 容器狀態:"
    docker-compose ps
    
    # 檢查健康狀態
    echo ""
    echo "🏥 健康檢查:"
    for service in postgresql redis minio ffmpeg-transcoder stream-puller; do
        if docker-compose ps | grep -q "$service.*Up"; then
            log_success "$service: 運行中"
        else
            log_error "$service: 未運行"
        fi
    done
    
    # 檢查流拉取服務
    echo ""
    echo "🎬 流拉取服務狀態:"
    if pgrep -f "stream-puller" > /dev/null; then
        log_success "stream-puller: 運行中"
        if curl -s "http://localhost:8083" > /dev/null 2>&1; then
            log_success "HLS 服務器: 正常"
        else
            log_error "HLS 服務器: 異常"
        fi
    else
        log_error "stream-puller: 未運行"
    fi
}

# 查看日誌
show_logs() {
    local service=${1:-""}
    
    if [ -z "$service" ]; then
        log_info "查看所有服務日誌 (按 Ctrl+C 退出)..."
        docker-compose logs -f
    else
        log_info "查看 $service 服務日誌 (按 Ctrl+C 退出)..."
        docker-compose logs -f "$service"
    fi
}

# 重新構建服務
build_services() {
    log_info "重新構建服務..."
    docker-compose build --no-cache
    log_success "服務構建完成"
}

# 清理資源
clean_resources() {
    log_warning "清理 Docker 資源..."
    
    # 停止並移除容器
    docker-compose down --remove-orphans
    
    # 清理未使用的映像
    docker image prune -f
    
    # 清理未使用的卷
    docker volume prune -f
    
    log_success "清理完成"
}

# 初始化 MinIO 桶
init_minio() {
    log_info "初始化 MinIO 桶..."
    if [ -f "./docker/minio/init-bucket.sh" ]; then
        ./docker/minio/init-bucket.sh
        log_success "MinIO 桶初始化完成"
    else
        log_error "MinIO 初始化腳本不存在"
    fi
}

# 初始化直播服務
init_live() {
    log_info "初始化直播服務..."
    
    # 創建直播桶
    if command -v mc &> /dev/null; then
        mc alias set s3 http://localhost:9000 minioadmin minioadmin
        mc mb s3/stream-demo-live --ignore-existing
        mc policy set download s3/stream-demo-live
        log_success "直播桶初始化完成"
    else
        log_warning "MinIO Client (mc) 未安裝，請手動創建 stream-demo-live 桶"
    fi
}

# 查看直播狀態
show_live_status() {
    log_info "查看直播狀態..."
    
    echo ""
    echo "📡 直播服務狀態:"
    log_info "Stream Puller 統一處理所有直播流"
    log_info "支援 HLS 拉流和 RTMP 推流轉換"
    
    echo ""
    echo "🎬 直播流服務狀態:"
    if curl -s http://localhost:8083/health > /dev/null 2>&1; then
        log_success "Stream Puller: 運行中"
        echo "HLS 播放地址: http://localhost:8083/[stream_name]/index.m3u8"
    else
        log_error "Stream Puller: 未運行"
    fi
    
    echo ""
    echo "🎬 當前直播流:"
    if [ -d "/tmp/public_streams" ]; then
        streams=$(ls /tmp/public_streams/ 2>/dev/null || true)
        if [ -n "$streams" ]; then
            for stream in $streams; do
                if [ -f "/tmp/public_streams/$stream/index.m3u8" ]; then
                    log_success "直播中: $stream"
                    echo "  HLS: http://localhost:8083/$stream/index.m3u8"
                fi
            done
        else
            log_info "目前沒有直播流"
        fi
    else
        log_error "直播目錄不存在"
    fi
    
    echo ""
    echo "📊 流服務狀態:"
    log_info "Stream Puller 統一處理所有直播流"
    log_info "支援 HLS 拉流和 RTMP 推流轉換"
}

# 運行測試
run_tests() {
    log_info "運行 Go 測試..."
    cd backend
    go test ./services -v
    cd ..
    log_success "測試完成"
}

# 管理流拉取服務
manage_stream_puller() {
    local action=${1:-help}
    
    case "$action" in
        start)
            log_info "啟動流拉取服務..."
            
            # 使用 Docker Compose 啟動 stream-puller
            docker-compose up -d stream-puller
            
            # 等待服務啟動
            sleep 5
            
            if docker-compose ps stream-puller | grep -q "Up"; then
                log_success "流拉取服務啟動成功"
                log_info "HTTP 服務器: http://localhost:8083"
                log_info "輸出目錄: /tmp/public_streams (Docker volume)"
            else
                log_error "服務啟動失敗"
                return 1
            fi
            ;;
        stop)
            log_info "停止流拉取服務..."
            
            docker-compose stop stream-puller
            
            if ! docker-compose ps stream-puller | grep -q "Up"; then
                log_success "服務已停止"
            else
                log_error "停止服務失敗"
                return 1
            fi
            ;;
        restart)
            log_info "重啟流拉取服務..."
            docker-compose restart stream-puller
            sleep 5
            
            if docker-compose ps stream-puller | grep -q "Up"; then
                log_success "服務重啟成功"
            else
                log_error "服務重啟失敗"
                return 1
            fi
            ;;
        status)
            log_info "流拉取服務狀態:"
            echo "=================="
            
            docker-compose ps stream-puller
            
            if docker-compose ps stream-puller | grep -q "Up"; then
                echo -e "狀態: ${GREEN}運行中${NC}"
                echo "HTTP 服務器: http://localhost:8083"
                echo "容器名稱: stream-demo-stream-puller"
                
                # 檢查 HTTP 服務
                if curl -s "http://localhost:8083" > /dev/null 2>&1; then
                    echo -e "HTTP 服務: ${GREEN}正常${NC}"
                else
                    echo -e "HTTP 服務: ${RED}異常${NC}"
                fi
                
                # 顯示 HLS 文件 (從 Docker volume)
                echo "HLS 文件:"
                docker exec stream-demo-stream-puller ls -la /tmp/public_streams/ 2>/dev/null || echo "無 HLS 文件"
            else
                echo -e "狀態: ${RED}未運行${NC}"
            fi
            ;;
        logs)
            log_info "顯示服務日誌 (按 Ctrl+C 退出):"
            echo "=================="
            docker-compose logs -f stream-puller
            ;;
        test)
            log_info "測試流播放..."
            echo "=================="
            
            # 檢查容器是否運行
            if ! docker-compose ps stream-puller | grep -q "Up"; then
                log_error "stream-puller 容器未運行"
                return 1
            fi
            
            # 從容器內檢查 HLS 文件
            streams=$(docker exec stream-demo-stream-puller ls /tmp/public_streams/ 2>/dev/null || true)
            
            if [ -n "$streams" ]; then
                for stream_name in $streams; do
                    hls_url="http://localhost:8083/$stream_name/index.m3u8"
                    
                    echo "測試流: $stream_name"
                    if curl -s -I "$hls_url" | grep -q "200 OK"; then
                        echo -e "  ${GREEN}✓${NC} HLS 播放列表可訪問"
                    else
                        echo -e "  ${RED}✗${NC} HLS 播放列表無法訪問"
                    fi
                    
                    # 檢查片段文件
                    ts_count=$(docker exec stream-demo-stream-puller find "/tmp/public_streams/$stream_name" -name "*.ts" 2>/dev/null | wc -l)
                    echo "  片段文件: $ts_count 個"
                done
            else
                log_info "目前沒有直播流"
            fi
            ;;
        help|--help|-h)
            echo "🎬 流拉取服務管理"
            echo ""
            echo "用法: $0 stream-puller [命令]"
            echo ""
            echo "命令:"
            echo "  start     啟動服務"
            echo "  stop      停止服務"
            echo "  restart   重啟服務"
            echo "  status    顯示狀態"
            echo "  logs      顯示日誌"
            echo "  test      測試流播放"
            echo "  help      顯示幫助"
            ;;
        *)
            log_error "未知命令: $action"
            manage_stream_puller help
            return 1
            ;;
    esac
}

# 主函數
main() {
    check_docker
    
    case "${1:-help}" in
        start)
            start_services
            ;;
        stop)
            stop_services
            ;;
        restart)
            restart_services
            ;;
        status)
            check_services_status
            ;;
        logs)
            show_logs "$2"
            ;;
        build)
            build_services
            ;;
        clean)
            clean_resources
            ;;
        init)
            init_minio
            ;;
        init-live)
            init_live
            ;;
        live-status)
            show_live_status
            ;;
        stream-puller)
            manage_stream_puller "$2"
            ;;
        test)
            run_tests
            ;;
        help|--help|-h)
            show_help
            ;;
        *)
            log_error "未知命令: $1"
            echo ""
            show_help
            exit 1
            ;;
    esac
}

# 執行主函數
main "$@" 