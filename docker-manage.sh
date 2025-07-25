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
    echo "  help      顯示此幫助信息"
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
    for service in postgresql redis minio ffmpeg-transcoder; do
        if docker-compose ps | grep -q "$service.*Up"; then
            log_success "$service: 運行中"
        else
            log_error "$service: 未運行"
        fi
    done
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

# 運行測試
run_tests() {
    log_info "運行 Go 測試..."
    cd backend
    go test ./services -v
    cd ..
    log_success "測試完成"
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