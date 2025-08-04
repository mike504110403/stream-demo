#!/bin/bash

# 生產環境部署腳本
# 用於完整容器化部署

# 顏色定義
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
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

log_step() {
    echo -e "${PURPLE}🔧 $1${NC}"
}

# 顯示幫助信息
show_help() {
    echo -e "${CYAN}🚀 生產環境部署腳本${NC}"
    echo "=================================="
    echo ""
    echo "用法: $0 [命令] [選項]"
    echo ""
    echo "命令:"
    echo "  deploy    部署生產環境"
    echo "  start     啟動生產服務"
    echo "  stop      停止生產服務"
    echo "  restart   重啟生產服務"
    echo "  status    查看生產服務狀態"
    echo "  logs      查看服務日誌"
    echo "  update    更新應用"
    echo "  backup    備份數據"
    echo "  help      顯示此幫助信息"
    echo ""
    echo "選項:"
    echo "  --force   強制重新部署"
    echo "  --clean   清理並重建"
    echo "  --env     指定環境文件"
    echo ""
    echo "範例:"
    echo "  $0 deploy        # 部署生產環境"
    echo "  $0 deploy --force # 強制重新部署"
    echo "  $0 status        # 查看狀態"
    echo "  $0 stop          # 停止服務"
}

# 檢查 Docker 是否運行
check_docker() {
    if ! command -v docker >/dev/null 2>&1; then
        log_error "Docker 未安裝"
        echo ""
        log_info "安裝 Docker："
        echo "  macOS: https://docs.docker.com/desktop/install/mac-install/"
        echo "  Linux: https://docs.docker.com/engine/install/"
        echo "  Windows: https://docs.docker.com/desktop/install/windows-install/"
        return 1
    fi
    
    if ! docker info >/dev/null 2>&1; then
        log_error "Docker 未運行"
        echo ""
        log_info "請啟動 Docker："
        echo "  macOS: 啟動 Docker Desktop 應用"
        echo "  Linux: sudo systemctl start docker"
        echo "  Windows: 啟動 Docker Desktop 應用"
        return 1
    fi
    
    log_success "Docker 已啟動"
    return 0
}

# 檢查 Docker Compose
check_docker_compose() {
    if ! command -v docker-compose >/dev/null 2>&1; then
        log_error "Docker Compose 未安裝，請先安裝 Docker Compose"
        return 1
    fi
    
    log_success "Docker Compose 已安裝"
    return 0
}

# 檢查環境變數文件
check_env_file() {
    local env_file=${1:-"docker/.env"}
    
    if [ ! -f "$env_file" ]; then
        log_warning "環境變數文件 $env_file 不存在"
        log_info "正在複製範例文件..."
        if [ -f "docker/env.example" ]; then
            cp docker/env.example "$env_file"
            log_success "已創建環境變數文件，請編輯 $env_file 配置生產環境"
            return 1
        else
            log_error "找不到環境變數範例文件"
            return 1
        fi
    fi
    
    log_success "環境變數文件已配置"
    return 0
}

# 建置 Docker 映像
build_images() {
    log_step "建置 Docker 映像..."
    
    cd docker
    
    local total_images=6
    local current_image=0
    
    # 建置後端映像
    ((current_image++))
    log_info "[$current_image/$total_images] 建置後端映像..."
    if docker build -t stream-demo-backend ../backend; then
        log_success "後端映像建置完成"
    else
        log_error "後端映像建置失敗"
        return 1
    fi
    
    # 建置前端映像
    ((current_image++))
    log_info "[$current_image/$total_images] 建置前端映像..."
    if docker build -t stream-demo-frontend ../frontend; then
        log_success "前端映像建置完成"
    else
        log_error "前端映像建置失敗"
        return 1
    fi
    
    # 建置 nginx 反向代理映像
    ((current_image++))
    log_info "[$current_image/$total_images] 建置 Nginx 反向代理映像..."
    if docker build -f nginx/Dockerfile.reverse-proxy-prod -t stream-demo-nginx-reverse-proxy nginx/; then
        log_success "Nginx 反向代理映像建置完成"
    else
        log_error "Nginx 反向代理映像建置失敗"
        return 1
    fi
    
    # 建置 nginx-rtmp 映像
    ((current_image++))
    log_info "[$current_image/$total_images] 建置 Nginx RTMP 映像..."
    if docker build -f nginx/Dockerfile.rtmp -t stream-demo-nginx-rtmp nginx/; then
        log_success "Nginx RTMP 映像建置完成"
    else
        log_error "Nginx RTMP 映像建置失敗"
        return 1
    fi
    
    # 建置 stream-puller 映像
    ((current_image++))
    log_info "[$current_image/$total_images] 建置 Stream Puller 映像..."
    if docker build -t stream-demo-stream-puller ../backend/cmd/stream_puller; then
        log_success "Stream Puller 映像建置完成"
    else
        log_error "Stream Puller 映像建置失敗"
        return 1
    fi
    
    # 建置 FFmpeg 轉碼器映像
    ((current_image++))
    log_info "[$current_image/$total_images] 建置 FFmpeg 轉碼器映像..."
    if docker build -t stream-demo-ffmpeg-transcoder ffmpeg/; then
        log_success "FFmpeg 轉碼器映像建置完成"
    else
        log_error "FFmpeg 轉碼器映像建置失敗"
        return 1
    fi
    
    cd ..
    
    log_success "🎉 所有映像建置完成！"
    return 0
}

# 部署生產環境
deploy_production() {
    echo -e "${CYAN}🚀 生產環境部署${NC}"
    echo "=================================="
    echo ""
    
    # 檢查依賴
    log_step "檢查部署依賴..."
    if ! check_docker; then
        exit 1
    fi
    
    if ! check_docker_compose; then
        exit 1
    fi
    
    # 檢查環境變數
    local env_file="docker/.env"
    if [[ "$*" == *"--env"* ]]; then
        env_file="$2"
    fi
    
    if ! check_env_file "$env_file"; then
        log_error "請先配置環境變數文件"
        exit 1
    fi
    
    echo ""
    
    # 建置映像
    if [[ "$*" == *"--force"* ]] || [[ "$*" == *"--clean"* ]]; then
        build_images
    fi
    
    echo ""
    
    # 啟動服務
    log_step "啟動生產服務..."
    cd docker
    
    if [[ "$*" == *"--clean"* ]]; then
        log_info "清理並重建服務..."
        docker-compose -f docker-compose.yml down -v
        docker system prune -f
    fi
    
    # 啟動所有服務
    docker-compose -f docker-compose.yml up -d
    
    cd ..
    
    # 等待服務啟動
    log_info "等待服務啟動..."
    
    local max_attempts=30
    local attempt=0
    
    while [ $attempt -lt $max_attempts ]; do
        ((attempt++))
        echo -n "."
        sleep 1
        
        # 每 5 秒顯示進度
        if [ $((attempt % 5)) -eq 0 ]; then
            echo " ($attempt/$max_attempts)"
        fi
        
        # 檢查關鍵服務是否啟動
        if curl -s "http://localhost:8084" > /dev/null 2>&1; then
            echo ""
            log_success "服務啟動完成！"
            break
        fi
    done
    
    if [ $attempt -eq $max_attempts ]; then
        echo ""
        log_warning "服務啟動超時，但可能仍在啟動中..."
    fi
    
    # 檢查服務狀態
    check_production_status
    
    echo ""
    log_success "🎉 生產環境部署完成！"
    
    # 顯示訪問信息
    show_production_info
}

# 啟動生產服務
start_production() {
    log_step "啟動生產服務..."
    
    cd docker
    docker-compose -f docker-compose.yml up -d
    cd ..
    
    log_success "生產服務已啟動"
}

# 停止生產服務
stop_production() {
    log_step "停止生產服務..."
    
    cd docker
    docker-compose -f docker-compose.yml down
    cd ..
    
    log_success "生產服務已停止"
}

# 重啟生產服務
restart_production() {
    log_step "重啟生產服務..."
    
    stop_production
    sleep 5
    start_production
    
    log_success "生產服務已重啟"
}

# 檢查生產服務狀態
check_production_status() {
    log_info "📊 檢查生產服務狀態..."
    
    cd docker
    docker-compose -f docker-compose.yml ps
    cd ..
    
    # 檢查關鍵服務
    echo ""
    log_info "檢查關鍵服務..."
    
    # 檢查後端 API
    if curl -s "http://localhost:8084/api/health" > /dev/null 2>&1; then
        log_success "後端 API: 運行中"
    else
        log_warning "後端 API: 未運行"
    fi
    
    # 檢查前端
    if curl -s "http://localhost:8084" > /dev/null 2>&1; then
        log_success "前端: 運行中"
    else
        log_warning "前端: 未運行"
    fi
    
    # 檢查 MinIO
    if curl -s "http://localhost:9001" > /dev/null 2>&1; then
        log_success "MinIO: 運行中"
    else
        log_warning "MinIO: 未運行"
    fi
}

# 查看服務日誌
show_logs() {
    local service=${1:-""}
    
    cd docker
    
    if [ -z "$service" ]; then
        log_info "查看所有服務日誌..."
        docker-compose -f docker-compose.yml logs -f
    else
        log_info "查看 $service 服務日誌..."
        docker-compose -f docker-compose.yml logs -f "$service"
    fi
    
    cd ..
}

# 更新應用
update_application() {
    log_step "更新應用..."
    
    # 拉取最新代碼
    log_info "拉取最新代碼..."
    git pull origin main
    
    # 重新建置映像
    build_images
    
    # 重啟服務
    restart_production
    
    log_success "應用更新完成"
}

# 備份數據
backup_data() {
    log_step "備份數據..."
    
    local backup_dir="backup/$(date +%Y%m%d_%H%M%S)"
    mkdir -p "$backup_dir"
    
    # 備份資料庫
    log_info "備份 PostgreSQL 數據..."
    docker exec stream-demo-postgresql pg_dump -U stream_user -d stream_demo > "$backup_dir/postgres_backup.sql"
    
    # 備份 MySQL 數據
    log_info "備份 MySQL 數據..."
    docker exec stream-demo-mysql mysqldump -u stream_user -pstream_password stream_demo > "$backup_dir/mysql_backup.sql"
    
    # 備份 MinIO 數據
    log_info "備份 MinIO 數據..."
    docker run --rm -v "$backup_dir:/backup" --network stream-demo_stream-demo-network \
        minio/mc mirror minio/stream-demo-videos /backup/minio
    
    # 備份配置
    log_info "備份配置..."
    tar -czf "$backup_dir/config_backup.tar.gz" docker/
    
    log_success "數據備份完成: $backup_dir"
}

# 顯示生產環境信息
show_production_info() {
    echo ""
    log_info "📊 生產環境訪問信息："
    echo ""
    echo "🌐 統一入口: http://localhost:8084"
    echo "🎬 前端應用: http://localhost:8084"
    echo "🔧 後端 API: http://localhost:8084/api"
    echo "📦 MinIO Console: http://localhost:9001"
    echo "📺 直播流服務: http://localhost:8083"
    echo "📡 RTMP 推流: rtmp://localhost:1935/live"
    echo "🎥 HLS 播放: http://localhost:8083/[stream_key]/index.m3u8"
    echo ""
    echo "🔧 管理命令："
    echo "  查看狀態: ./cmd/deploy.sh status"
    echo "  查看日誌: ./cmd/deploy.sh logs"
    echo "  停止服務: ./cmd/deploy.sh stop"
    echo "  重啟服務: ./cmd/deploy.sh restart"
}

# 主函數
main() {
    case "${1:-help}" in
        deploy)
            deploy_production "$@"
            ;;
        start)
            start_production
            ;;
        stop)
            stop_production
            ;;
        restart)
            restart_production
            ;;
        status)
            check_production_status
            ;;
        logs)
            show_logs "$2"
            ;;
        update)
            update_application
            ;;
        backup)
            backup_data
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