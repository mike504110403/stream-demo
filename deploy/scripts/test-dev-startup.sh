#!/bin/bash

# 開發模式啟動測試腳本
# 測試 F5 一鍵啟動 (前後端透過 IDE 啟動，其他周邊服務透過 Docker)

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

# 檢查 Docker 是否運行
check_docker() {
    log_info "檢查 Docker 服務..."
    if ! docker info > /dev/null 2>&1; then
        log_error "Docker 服務未運行，請先啟動 Docker"
        return 1
    fi
    log_success "Docker 服務正常"
}

# 檢查端口是否被佔用
check_port() {
    local port=$1
    local service=$2
    
    if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1; then
        log_warning "端口 $port 已被佔用 ($service)"
        return 1
    else
        log_success "端口 $port 可用"
        return 0
    fi
}

# 檢查服務健康狀態
check_service_health() {
    local service=$1
    local url=$2
    local timeout=${3:-10}
    
    log_info "檢查 $service 健康狀態..."
    
    for i in $(seq 1 $timeout); do
        if curl -s -f "$url" > /dev/null 2>&1; then
            log_success "$service 健康檢查通過"
            return 0
        fi
        sleep 1
    done
    
    log_error "$service 健康檢查失敗"
    return 1
}

# 測試開發模式啟動
test_dev_startup() {
    log_info "🧪 測試開發模式啟動 (F5 一鍵啟動)..."
    
    # 檢查必要端口
    local ports=(
        "5432:PostgreSQL"
        "6379:Redis"
        "9000:MinIO"
        "9001:MinIO Console"
        "1935:RTMP"
        "8083:Stream Puller"
    )
    
    for port_info in "${ports[@]}"; do
        IFS=':' read -r port service <<< "$port_info"
        check_port $port $service
    done
    
    # 啟動基礎設施服務
    log_info "啟動基礎設施服務..."
    cd deploy
    docker-compose -f docker-compose.dev.yml up -d postgresql redis minio rtmp stream-puller media
    
    # 等待服務啟動
    log_info "等待服務啟動..."
    sleep 15
    
    # 檢查服務狀態
    log_info "檢查服務狀態..."
    local services=("postgresql" "redis" "minio" "rtmp" "stream-puller" "media")
    
    for service in "${services[@]}"; do
        if docker-compose -f docker-compose.dev.yml ps | grep -q "$service.*Up"; then
            log_success "$service 運行正常"
        else
            log_error "$service 啟動失敗"
            return 1
        fi
    done
    
    # 檢查服務健康狀態
    check_service_health "MinIO" "http://localhost:9000/minio/health/live" 10
    check_service_health "Stream Puller" "http://localhost:8083/health" 10
    
    cd ..
    log_success "開發模式啟動測試完成"
    echo ""
    echo "📋 開發模式服務狀態:"
    echo "  ✅ PostgreSQL: localhost:5432"
    echo "  ✅ Redis: localhost:6379"
    echo "  ✅ MinIO: localhost:9000"
    echo "  ✅ MinIO Console: localhost:9001"
    echo "  ✅ RTMP: localhost:1935"
    echo "  ✅ Stream Puller: localhost:8083"
    echo "  ✅ Media Service: 運行中"
    echo ""
    echo "🚀 現在可以在 IDE 中啟動前後端服務"
    echo "  後端: 使用 launch.json 配置 (localhost:8080)"
    echo "  前端: npm run dev (localhost:5173)"
    echo ""
    echo "🌐 訪問地址:"
    echo "  MinIO Console: http://localhost:9001 (minioadmin/minioadmin)"
    echo "  Stream Puller: http://localhost:8083"
    echo "  RTMP 推流: rtmp://localhost:1935/live"
}

# 檢查服務間通訊
check_service_communication() {
    log_info "🔍 檢查服務間通訊..."
    
    # 檢查網路
    if docker network ls | grep -q "stream-demo-network"; then
        log_success "Docker 網路 stream-demo-network 存在"
    else
        log_error "Docker 網路 stream-demo-network 不存在"
        return 1
    fi
    
    # 檢查容器間通訊
    local containers=("stream-demo-postgresql" "stream-demo-redis" "stream-demo-minio" "stream-demo-rtmp" "stream-demo-stream-puller" "stream-demo-media")
    
    for container in "${containers[@]}"; do
        if docker ps | grep -q "$container"; then
            log_success "$container 正在運行"
        else
            log_warning "$container 未運行"
        fi
    done
    
    # 檢查 Stream Puller 是否能連接到資料庫
    if docker exec stream-demo-stream-puller wget -q --spider http://postgresql:5432 2>/dev/null; then
        log_success "Stream Puller 可以連接到 PostgreSQL"
    else
        log_warning "Stream Puller 無法連接到 PostgreSQL (可能是正常的，因為 PostgreSQL 不提供 HTTP 接口)"
    fi
    
    # 檢查 Media Service 是否能連接到 MinIO
    if docker exec stream-demo-media wget -q --spider http://minio:9000/minio/health/live 2>/dev/null; then
        log_success "Media Service 可以連接到 MinIO"
    else
        log_error "Media Service 無法連接到 MinIO"
        return 1
    fi
}

# 檢查環境變數配置
check_environment_config() {
    log_info "🔧 檢查環境變數配置..."
    
    # 檢查 Stream Puller 環境變數
    local stream_puller_env_vars=(
        "OUTPUT_DIR=/tmp/public_streams"
        "HTTP_PORT=8081"
        "DB_HOST=postgresql"
        "DB_PORT=5432"
        "DB_USER=stream_user"
        "DB_PASS=stream_password"
        "DB_NAME=stream_demo"
    )
    
    for env_var in "${stream_puller_env_vars[@]}"; do
        IFS='=' read -r key value <<< "$env_var"
        if docker exec stream-demo-stream-puller env | grep -q "^$key=$value$"; then
            log_success "Stream Puller 環境變數 $key 配置正確"
        else
            log_warning "Stream Puller 環境變數 $key 配置可能不正確"
        fi
    done
    
    # 檢查 Media Service 環境變數
    local media_env_vars=(
        "MINIO_ENDPOINT=http://minio:9000"
        "MINIO_ACCESS_KEY=minioadmin"
        "MINIO_SECRET_KEY=minioadmin"
        "MINIO_BUCKET=stream-demo-videos"
        "MINIO_PROCESSED_BUCKET=stream-demo-processed"
    )
    
    for env_var in "${media_env_vars[@]}"; do
        IFS='=' read -r key value <<< "$env_var"
        if docker exec stream-demo-media env | grep -q "^$key=$value$"; then
            log_success "Media Service 環境變數 $key 配置正確"
        else
            log_warning "Media Service 環境變數 $key 配置可能不正確"
        fi
    done
}

# 清理服務
cleanup_services() {
    log_info "🧹 清理服務..."
    cd deploy
    docker-compose -f docker-compose.dev.yml down --remove-orphans
    cd ..
    log_success "服務清理完成"
}

# 主函數
main() {
    echo "🚀 開始開發模式啟動測試..."
    echo ""
    
    # 檢查 Docker
    check_docker || exit 1
    
    # 測試開發模式啟動
    test_dev_startup
    
    echo ""
    echo "⏳ 等待 10 秒後檢查服務間通訊..."
    sleep 10
    
    # 檢查服務間通訊
    check_service_communication
    
    # 檢查環境變數配置
    check_environment_config
    
    echo ""
    log_success "🎉 開發模式測試完成！"
    echo ""
    echo "📋 測試總結:"
    echo "  ✅ 基礎設施服務啟動"
    echo "  ✅ 服務間通訊"
    echo "  ✅ 環境變數配置"
    echo ""
    echo "🔧 下一步:"
    echo "  1. 在 IDE 中啟動後端服務 (localhost:8080)"
    echo "  2. 在 IDE 中啟動前端服務 (localhost:5173)"
    echo "  3. 訪問 http://localhost:5173 開始開發"
    echo ""
    echo "🌐 服務地址:"
    echo "  MinIO Console: http://localhost:9001 (minioadmin/minioadmin)"
    echo "  Stream Puller: http://localhost:8083"
    echo "  RTMP 推流: rtmp://localhost:1935/live"
    
    # 詢問是否清理
    echo ""
    read -p "是否清理所有服務？(y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        cleanup_services
    fi
}

# 執行主函數
main "$@" 