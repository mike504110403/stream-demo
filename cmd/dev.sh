#!/bin/bash

# 開發環境快速啟動腳本
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

# 顯示幫助信息
show_help() {
    echo "🚀 開發環境快速啟動腳本"
    echo ""
    echo "用法: $0 [命令]"
    echo ""
    echo "命令:"
    echo "  start     啟動開發環境 (周邊服務)"
    echo "  stop      停止開發環境"
    echo "  restart   重啟開發環境"
    echo "  status    查看開發環境狀態"
    echo "  logs      查看服務日誌"
    echo "  help      顯示此幫助信息"
    echo ""
    echo "範例:"
    echo "  $0 start    # 啟動周邊服務"
    echo "  $0 status   # 查看狀態"
    echo "  $0 stop     # 停止服務"
}

# 啟動開發環境 (周邊服務)
start_dev_environment() {
    log_info "🚀 啟動開發環境 (周邊服務)..."
    
    # 啟動周邊服務
    log_info "啟動周邊服務..."
    ./cmd/manage.sh start-dev
    
    # 等待服務啟動
    log_info "等待服務啟動..."
    sleep 5
    
    log_success "🎉 開發環境啟動完成！"
    echo ""
    echo "📋 訪問地址:"
    echo "  統一入口: http://localhost:8084"
    echo "  前端 (IDE): http://localhost:5173"
    echo "  後端 (IDE): http://localhost:8080"
    echo "  MinIO Console: http://localhost:9001"
    echo "  HLS 播放: http://localhost:8083/[stream_name]/index.m3u8"
    echo "  RTMP 推流: rtmp://localhost:1935/live"
    echo ""
    echo "💡 請在 IDE 中啟動前後端服務:"
    echo "  後端: cd backend && go run main.go"
    echo "  前端: cd frontend && npm run dev"
}

# 停止開發環境
stop_dev_environment() {
    log_info "🛑 停止開發環境..."
    
    # 停止周邊服務
./cmd/manage.sh stop
    
    log_success "開發環境已停止"
}

# 重啟開發環境
restart_dev_environment() {
    log_info "🔄 重啟開發環境..."
    stop_dev_environment
    sleep 2
    start_dev_environment
}

# 查看開發環境狀態
check_dev_status() {
    log_info "📊 查看開發環境狀態..."
    
    # 檢查周邊服務狀態
./cmd/manage.sh dev-status
    
    # 檢查 IDE 服務
    echo ""
    echo "💻 IDE 服務狀態:"
    
    # 檢查後端
    if curl -s "http://localhost:8080/api/health" > /dev/null 2>&1; then
        log_success "後端 (IDE): 運行中"
    else
        log_warning "後端 (IDE): 未運行 (請在 IDE 中啟動)"
    fi
    
    # 檢查前端
    if curl -s "http://localhost:5173" > /dev/null 2>&1; then
        log_success "前端 (IDE): 運行中"
    else
        log_warning "前端 (IDE): 未運行 (請在 IDE 中啟動)"
    fi
}



# 查看日誌
show_logs() {
    local service=${1:-""}
    
    if [ -z "$service" ]; then
    log_info "查看周邊服務日誌..."
    ./cmd/manage.sh dev-logs
else
    log_info "查看 $service 服務日誌..."
    ./cmd/manage.sh dev-logs "$service"
fi
}

# 主函數
main() {
    case "${1:-help}" in
        start)
            start_dev_environment
            ;;
        stop)
            stop_dev_environment
            ;;
        restart)
            restart_dev_environment
            ;;
        status)
            check_dev_status
            ;;
        logs)
            show_logs "$2"
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