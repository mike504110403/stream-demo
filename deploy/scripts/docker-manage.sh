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
    echo "  start     啟動所有服務 (生產模式)"
    echo "  start-dev 啟動周邊服務 (開發模式，前後端由 IDE 啟動)"
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
    echo "  nginx     管理 nginx 反向代理"
    echo "  frontend  管理前端應用"
    echo "  backend   管理後端 API"
    echo "  test      運行 Go 測試"
    echo "  help      顯示此幫助信息"
    echo ""
    echo "開發模式命令:"
    echo "  start-dev 啟動周邊服務 (資料庫、Redis、MinIO、直播服務等)"
    echo "  dev-status 查看開發模式狀態"
    echo "  dev-logs  查看開發模式日誌"
    echo ""
    echo "流拉取服務命令:"
    echo "  stream-puller start    啟動流拉取服務"
    echo "  stream-puller stop     停止流拉取服務"
    echo "  stream-puller restart  重啟流拉取服務"
    echo "  stream-puller status   查看流拉取服務狀態"
    echo "  stream-puller logs     查看流拉取服務日誌"
    echo "  stream-puller test     測試流播放"
    echo ""
    echo "Nginx 反向代理命令:"
    echo "  nginx start    啟動 nginx 反向代理"
    echo "  nginx stop     停止 nginx 反向代理"
    echo "  nginx restart  重啟 nginx 反向代理"
    echo "  nginx status   查看 nginx 反向代理狀態"
    echo "  nginx logs     查看 nginx 反向代理日誌"
    echo "  nginx test     測試反向代理功能"
    echo ""
    echo "範例:"
    echo "  $0 start      # 啟動所有服務 (生產模式)"
    echo "  $0 start-dev  # 啟動周邊服務 (開發模式)"
    echo "  $0 logs       # 查看日誌"
    echo "  $0 status     # 查看狀態"
}

# 啟動服務 (生產模式)
start_services() {
    log_info "啟動所有服務 (生產模式)..."
    docker-compose -f docker-compose.yml --project-name stream-demo up -d
    log_success "服務啟動完成"
    
    # 等待服務啟動
    log_info "等待服務啟動..."
    sleep 10
    
    # 檢查服務狀態
    check_services_status
}

# 啟動開發模式服務 (只啟動周邊服務)
start_dev_services() {
    log_info "啟動開發模式服務 (周邊服務)..."
    log_info "前後端將由 IDE 啟動，nginx 會代理到主機的 5173 和 8080 端口"
    
    # 使用開發模式配置啟動服務
    docker-compose -f deploy/docker-compose.dev.yml --project-name stream-demo up -d
    log_success "開發模式服務啟動完成"
    
    # 等待服務啟動
    log_info "等待服務啟動..."
    sleep 10
    
    # 檢查開發模式服務狀態
    check_dev_services_status
}

# 檢查開發模式服務狀態
check_dev_services_status() {
    log_info "檢查開發模式服務狀態..."
    
    # 檢查容器狀態
    echo ""
    echo "📊 開發模式容器狀態:"
    docker-compose -f deploy/docker-compose.dev.yml --project-name stream-demo ps
    
    # 檢查健康狀態
    echo ""
    echo "🏥 健康檢查:"
    for service in postgresql redis minio receiver puller gateway; do
        if docker-compose -f deploy/docker-compose.dev.yml --project-name stream-demo ps | grep -q "$service.*Up"; then
            log_success "$service: 運行中"
        else
            log_error "$service: 未運行"
        fi
    done
    
    # 檢查開發模式配置
    echo ""
    echo "🔧 開發模式配置:"
    if curl -s "http://localhost:8084/dev-status" > /dev/null 2>&1; then
        log_success "Nginx 開發模式: 正常"
        echo "  開發模式狀態: $(curl -s http://localhost:8084/dev-status)"
    else
        log_error "Nginx 開發模式: 異常"
    fi
    
    # 檢查 IDE 啟動的服務
    echo ""
    echo "💻 IDE 服務檢查:"
    if curl -s "http://localhost:5173" > /dev/null 2>&1; then
        log_success "前端 (IDE): 運行中 (http://localhost:5173)"
    else
        log_warning "前端 (IDE): 未運行 (http://localhost:5173)"
    fi
    
    if curl -s "http://localhost:8080/api/health" > /dev/null 2>&1; then
        log_success "後端 (IDE): 運行中 (http://localhost:8080)"
    else
        log_warning "後端 (IDE): 未運行 (http://localhost:8080)"
    fi
    
    echo ""
    echo "📋 開發模式訪問地址:"
    echo "  統一入口: http://localhost:8084"
    echo "  前端 (IDE): http://localhost:5173"
    echo "  後端 (IDE): http://localhost:8080"
    echo "  MinIO Console: http://localhost:9001"
    echo "  HLS 播放: http://localhost:8083/[stream_name]/index.m3u8"
    echo "  RTMP 推流: rtmp://localhost:1935/live"
}

# 查看開發模式日誌
show_dev_logs() {
    local service=${1:-""}
    
    if [ -z "$service" ]; then
        log_info "查看開發模式服務日誌 (按 Ctrl+C 退出)..."
        docker-compose -f deploy/docker-compose.dev.yml --project-name stream-demo logs -f
    else
        log_info "查看開發模式 $service 服務日誌 (按 Ctrl+C 退出)..."
        docker-compose -f deploy/docker-compose.dev.yml --project-name stream-demo logs -f "$service"
    fi
}

# 停止服務
stop_services() {
    log_info "停止所有服務..."
    docker-compose -f docker-compose.yml --project-name stream-demo down
    log_success "服務停止完成"
}

# 重啟服務
restart_services() {
    log_info "重啟所有服務..."
    docker-compose -f docker-compose.yml --project-name stream-demo restart
    log_success "服務重啟完成"
}

# 檢查服務狀態
check_services_status() {
    log_info "檢查服務狀態..."
    
    # 檢查容器狀態
    echo ""
    echo "📊 容器狀態:"
    docker-compose -f docker-compose.yml --project-name stream-demo ps
    
    # 檢查健康狀態
    echo ""
    echo "🏥 健康檢查:"
    for service in postgresql redis minio receiver puller gateway; do
        if docker-compose -f docker-compose.yml --project-name stream-demo ps | grep -q "$service.*Up"; then
            log_success "$service: 運行中"
        else
            log_error "$service: 未運行"
        fi
    done
    
    # 檢查流拉取服務
    echo ""
    echo "🎬 流拉取服務狀態:"
    if docker-compose ps stream-puller | grep -q "Up"; then
        log_success "stream-puller: 運行中"
        if curl -s "http://localhost:8083/health" > /dev/null 2>&1; then
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
        docker-compose -f docker-compose.yml --project-name stream-demo logs -f
    else
        log_info "查看 $service 服務日誌 (按 Ctrl+C 退出)..."
        docker-compose -f docker-compose.yml --project-name stream-demo logs -f "$service"
    fi
}

# 重新構建服務
build_services() {
    log_info "重新構建服務..."
    docker-compose -f docker-compose.yml --project-name stream-demo build --no-cache
    log_success "服務構建完成"
}

# 清理資源
clean_resources() {
    log_warning "清理 Docker 資源..."
    
    # 停止並移除容器
    docker-compose -f docker-compose.yml --project-name stream-demo down --remove-orphans
    docker-compose -f deploy/docker-compose.dev.yml --project-name stream-demo down --remove-orphans
    
    # 清理未使用的映像
    docker image prune -f
    
    # 清理未使用的卷
    docker volume prune -f
    
    log_success "清理完成"
}

# 初始化 MinIO 桶
init_minio() {
    log_info "初始化 MinIO 桶..."
    if [ -f "./infrastructure/minio/init-bucket.sh" ]; then
        ./infrastructure/minio/init-bucket.sh
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
                if curl -s "http://localhost:8083/health" > /dev/null 2>&1; then
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

# 管理前端應用
manage_frontend() {
    local action=${1:-help}
    
    case "$action" in
        start)
            log_info "啟動前端應用..."
            
            docker-compose up -d frontend
            
            sleep 5
            
            if docker-compose ps frontend | grep -q "Up"; then
                log_success "前端應用啟動成功"
                log_info "前端地址: http://localhost:5173"
            else
                log_error "前端應用啟動失敗"
                return 1
            fi
            ;;
        stop)
            log_info "停止前端應用..."
            
            docker-compose stop frontend
            
            if ! docker-compose ps frontend | grep -q "Up"; then
                log_success "前端應用已停止"
            else
                log_error "停止前端應用失敗"
                return 1
            fi
            ;;
        restart)
            log_info "重啟前端應用..."
            docker-compose restart frontend
            sleep 5
            
            if docker-compose ps frontend | grep -q "Up"; then
                log_success "前端應用重啟成功"
            else
                log_error "前端應用重啟失敗"
                return 1
            fi
            ;;
        build)
            log_info "構建前端應用..."
            docker-compose build --no-cache frontend
            log_success "前端應用構建完成"
            ;;
        status)
            log_info "前端應用狀態:"
            echo "=================="
            
            docker-compose ps frontend
            
            if docker-compose ps frontend | grep -q "Up"; then
                echo -e "狀態: ${GREEN}運行中${NC}"
                echo "前端地址: http://localhost:5173"
                
                # 檢查健康狀態
                if curl -s "http://localhost:5173/" > /dev/null 2>&1; then
                    echo -e "健康檢查: ${GREEN}正常${NC}"
                else
                    echo -e "健康檢查: ${RED}異常${NC}"
                fi
            else
                echo -e "狀態: ${RED}未運行${NC}"
            fi
            ;;
        logs)
            log_info "顯示前端應用日誌 (按 Ctrl+C 退出):"
            echo "=================="
            docker-compose logs -f frontend
            ;;
        test)
            log_info "測試前端應用功能..."
            echo "=================="
            
            if ! docker-compose ps frontend | grep -q "Up"; then
                log_error "前端應用容器未運行"
                return 1
            fi
            
            echo "🧪 測試項目:"
            
            # 測試前端頁面
            echo "1. 前端頁面:"
            if curl -s -I "http://localhost:5173/" | grep -q "200 OK"; then
                echo -e "   ${GREEN}✓${NC} 前端頁面正常"
            else
                echo -e "   ${RED}✗${NC} 前端頁面異常"
            fi
            
            echo ""
            echo "📋 服務地址:"
            echo "  前端應用: http://localhost:5173"
            ;;
        help|--help|-h)
            echo "🎨 前端應用管理"
            echo ""
            echo "用法: $0 frontend [命令]"
            echo ""
            echo "命令:"
            echo "  start     啟動前端應用"
            echo "  stop      停止前端應用"
            echo "  restart   重啟前端應用"
            echo "  build     構建前端應用"
            echo "  status    顯示前端應用狀態"
            echo "  logs      顯示前端應用日誌"
            echo "  test      測試前端應用功能"
            echo "  help      顯示幫助"
            ;;
        *)
            log_error "未知命令: $action"
            manage_frontend help
            return 1
            ;;
    esac
}

# 管理後端 API
manage_backend() {
    local action=${1:-help}
    
    case "$action" in
        start)
            log_info "啟動後端 API..."
            
            docker-compose up -d backend
            
            sleep 10
            
            if docker-compose ps backend | grep -q "Up"; then
                log_success "後端 API 啟動成功"
                log_info "API 地址: http://localhost:8080"
            else
                log_error "後端 API 啟動失敗"
                return 1
            fi
            ;;
        stop)
            log_info "停止後端 API..."
            
            docker-compose stop backend
            
            if ! docker-compose ps backend | grep -q "Up"; then
                log_success "後端 API 已停止"
            else
                log_error "停止後端 API 失敗"
                return 1
            fi
            ;;
        restart)
            log_info "重啟後端 API..."
            docker-compose restart backend
            sleep 10
            
            if docker-compose ps backend | grep -q "Up"; then
                log_success "後端 API 重啟成功"
            else
                log_error "後端 API 重啟失敗"
                return 1
            fi
            ;;
        build)
            log_info "構建後端 API..."
            docker-compose build --no-cache backend
            log_success "後端 API 構建完成"
            ;;
        status)
            log_info "後端 API 狀態:"
            echo "=================="
            
            docker-compose ps backend
            
            if docker-compose ps backend | grep -q "Up"; then
                echo -e "狀態: ${GREEN}運行中${NC}"
                echo "API 地址: http://localhost:8080"
                
                # 檢查健康狀態
                if curl -s "http://localhost:8080/health" > /dev/null 2>&1; then
                    echo -e "健康檢查: ${GREEN}正常${NC}"
                else
                    echo -e "健康檢查: ${RED}異常${NC}"
                fi
            else
                echo -e "狀態: ${RED}未運行${NC}"
            fi
            ;;
        logs)
            log_info "顯示後端 API 日誌 (按 Ctrl+C 退出):"
            echo "=================="
            docker-compose logs -f backend
            ;;
        test)
            log_info "測試後端 API 功能..."
            echo "=================="
            
            if ! docker-compose ps backend | grep -q "Up"; then
                log_error "後端 API 容器未運行"
                return 1
            fi
            
            echo "🧪 測試項目:"
            
            # 測試健康檢查
            echo "1. 健康檢查:"
            if curl -s "http://localhost:8080/health" > /dev/null 2>&1; then
                echo -e "   ${GREEN}✓${NC} 健康檢查正常"
            else
                echo -e "   ${RED}✗${NC} 健康檢查失敗"
            fi
            
            # 測試 API 端點
            echo "2. API 端點:"
            if curl -s -I "http://localhost:8080/api/" | grep -q "404\|200\|401"; then
                echo -e "   ${GREEN}✓${NC} API 端點正常"
            else
                echo -e "   ${RED}✗${NC} API 端點異常"
            fi
            
            echo ""
            echo "📋 服務地址:"
            echo "  後端 API: http://localhost:8080"
            echo "  API 文檔: http://localhost:8080/api/"
            ;;
        help|--help|-h)
            echo "🔧 後端 API 管理"
            echo ""
            echo "用法: $0 backend [命令]"
            echo ""
            echo "命令:"
            echo "  start     啟動後端 API"
            echo "  stop      停止後端 API"
            echo "  restart   重啟後端 API"
            echo "  build     構建後端 API"
            echo "  status    顯示後端 API 狀態"
            echo "  logs      顯示後端 API 日誌"
            echo "  test      測試後端 API 功能"
            echo "  help      顯示幫助"
            ;;
        *)
            log_error "未知命令: $action"
            manage_backend help
            return 1
            ;;
    esac
}

# 管理 nginx 反向代理
manage_nginx() {
    local action=${1:-help}
    
    case "$action" in
        start)
            log_info "啟動 nginx 反向代理..."
            
            # 使用 Docker Compose 啟動 nginx-reverse-proxy
            docker-compose up -d nginx-reverse-proxy
            
            # 等待服務啟動
            sleep 5
            
            if docker-compose ps nginx-reverse-proxy | grep -q "Up"; then
                log_success "nginx 反向代理啟動成功"
                log_info "統一入口: http://localhost:80"
                log_info "前端應用: http://localhost/"
                log_info "後端 API: http://localhost/api/"
                log_info "HLS 播放: http://localhost/hls/"
                log_info "WebSocket: ws://localhost/ws/"
            else
                log_error "服務啟動失敗"
                return 1
            fi
            ;;
        stop)
            log_info "停止 nginx 反向代理..."
            
            docker-compose stop nginx-reverse-proxy
            
            if ! docker-compose ps nginx-reverse-proxy | grep -q "Up"; then
                log_success "服務已停止"
            else
                log_error "停止服務失敗"
                return 1
            fi
            ;;
        restart)
            log_info "重啟 nginx 反向代理..."
            docker-compose restart nginx-reverse-proxy
            sleep 5
            
            if docker-compose ps nginx-reverse-proxy | grep -q "Up"; then
                log_success "服務重啟成功"
            else
                log_error "服務重啟失敗"
                return 1
            fi
            ;;
        status)
            log_info "nginx 反向代理狀態:"
            echo "=================="
            
            docker-compose ps nginx-reverse-proxy
            
            if docker-compose ps nginx-reverse-proxy | grep -q "Up"; then
                echo -e "狀態: ${GREEN}運行中${NC}"
                echo "統一入口: http://localhost:80"
                echo "容器名稱: stream-demo-nginx-reverse-proxy"
                
                # 檢查健康狀態
                if curl -s "http://localhost/health" > /dev/null 2>&1; then
                    echo -e "健康檢查: ${GREEN}正常${NC}"
                else
                    echo -e "健康檢查: ${RED}異常${NC}"
                fi
                
                # 檢查各項服務
                echo ""
                echo "🔍 服務檢查:"
                
                # 檢查前端代理
                if curl -s -I "http://localhost/" | grep -q "200 OK\|302 Found"; then
                    echo -e "  前端代理: ${GREEN}正常${NC}"
                else
                    echo -e "  前端代理: ${RED}異常${NC}"
                fi
                
                # 檢查後端 API 代理
                if curl -s -I "http://localhost/api/" | grep -q "404\|200\|401"; then
                    echo -e "  後端 API 代理: ${GREEN}正常${NC}"
                else
                    echo -e "  後端 API 代理: ${RED}異常${NC}"
                fi
                
                # 檢查 HLS 代理
                if curl -s -I "http://localhost/hls/" | grep -q "200 OK\|404 Not Found"; then
                    echo -e "  HLS 代理: ${GREEN}正常${NC}"
                else
                    echo -e "  HLS 代理: ${RED}異常${NC}"
                fi
                
            else
                echo -e "狀態: ${RED}未運行${NC}"
            fi
            ;;
        logs)
            log_info "顯示 nginx 反向代理日誌 (按 Ctrl+C 退出):"
            echo "=================="
            docker-compose logs -f nginx-reverse-proxy
            ;;
        test)
            log_info "測試 nginx 反向代理功能..."
            echo "=================="
            
            # 檢查容器是否運行
            if ! docker-compose ps nginx-reverse-proxy | grep -q "Up"; then
                log_error "nginx-reverse-proxy 容器未運行"
                return 1
            fi
            
            echo "🧪 測試項目:"
            
            # 測試健康檢查
            echo "1. 健康檢查:"
            if curl -s "http://localhost/health" | grep -q "healthy"; then
                echo -e "   ${GREEN}✓${NC} 健康檢查正常"
            else
                echo -e "   ${RED}✗${NC} 健康檢查失敗"
            fi
            
            # 測試前端代理
            echo "2. 前端代理:"
            if curl -s -I "http://localhost/" | grep -q "200 OK\|302 Found"; then
                echo -e "   ${GREEN}✓${NC} 前端代理正常"
            else
                echo -e "   ${RED}✗${NC} 前端代理失敗"
            fi
            
            # 測試後端 API 代理
            echo "3. 後端 API 代理:"
            if curl -s -I "http://localhost/api/" | grep -q "404\|200\|401"; then
                echo -e "   ${GREEN}✓${NC} 後端 API 代理正常"
            else
                echo -e "   ${RED}✗${NC} 後端 API 代理失敗"
            fi
            
            # 測試 HLS 代理
            echo "4. HLS 代理:"
            if curl -s -I "http://localhost/hls/" | grep -q "200 OK\|404 Not Found"; then
                echo -e "   ${GREEN}✓${NC} HLS 代理正常"
            else
                echo -e "   ${RED}✗${NC} HLS 代理失敗"
            fi
            
            # 測試具體的 HLS 流
            echo "5. HLS 流測試:"
            streams=$(docker exec stream-demo-nginx-rtmp ls /tmp/hls/ 2>/dev/null || true)
            if [ -n "$streams" ]; then
                for stream_name in $streams; do
                    hls_url="http://localhost/hls/$stream_name/index.m3u8"
                    if curl -s -I "$hls_url" | grep -q "200 OK"; then
                        echo -e "   ${GREEN}✓${NC} $stream_name HLS 流正常"
                    else
                        echo -e "   ${RED}✗${NC} $stream_name HLS 流異常"
                    fi
                done
            else
                echo "   目前沒有 HLS 流"
            fi
            
            echo ""
            echo "📋 服務地址:"
            echo "  統一入口: http://localhost:80"
            echo "  前端應用: http://localhost/"
            echo "  後端 API: http://localhost/api/"
            echo "  HLS 播放: http://localhost/hls/[stream_name]/index.m3u8"
            echo "  WebSocket: ws://localhost/ws/"
            ;;
        help|--help|-h)
            echo "🌐 nginx 反向代理管理"
            echo ""
            echo "用法: $0 nginx [命令]"
            echo ""
            echo "命令:"
            echo "  start     啟動服務"
            echo "  stop      停止服務"
            echo "  restart   重啟服務"
            echo "  status    顯示狀態"
            echo "  logs      顯示日誌"
            echo "  test      測試反向代理功能"
            echo "  help      顯示幫助"
            ;;
        *)
            log_error "未知命令: $action"
            manage_nginx help
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
        start-dev)
            start_dev_services
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
        dev-status)
            check_dev_services_status
            ;;
        logs)
            show_logs "$2"
            ;;
        dev-logs)
            show_dev_logs "$2"
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
        nginx)
            manage_nginx "$2"
            ;;
        frontend)
            manage_frontend "$2"
            ;;
        backend)
            manage_backend "$2"
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