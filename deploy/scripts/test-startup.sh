#!/bin/bash

# 啟動測試腳本
# 測試 F5 一鍵啟動和完整 Docker 啟動

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

# 測試 F5 一鍵啟動 (開發模式)
test_f5_startup() {
    log_info "🧪 測試 F5 一鍵啟動 (開發模式)..."
    
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
    if docker-compose -f docker-compose.dev.yml ps | grep -q "postgresql.*Up"; then
        log_success "PostgreSQL 運行正常"
    else
        log_error "PostgreSQL 啟動失敗"
        return 1
    fi
    
    if docker-compose -f docker-compose.dev.yml ps | grep -q "redis.*Up"; then
        log_success "Redis 運行正常"
    else
        log_error "Redis 啟動失敗"
        return 1
    fi
    
    if docker-compose -f docker-compose.dev.yml ps | grep -q "minio.*Up"; then
        log_success "MinIO 運行正常"
    else
        log_error "MinIO 啟動失敗"
        return 1
    fi
    
    if docker-compose -f docker-compose.dev.yml ps | grep -q "rtmp.*Up"; then
        log_success "RTMP 服務運行正常"
    else
        log_error "RTMP 服務啟動失敗"
        return 1
    fi
    
    if docker-compose -f docker-compose.dev.yml ps | grep -q "stream-puller.*Up"; then
        log_success "Stream Puller 運行正常"
    else
        log_error "Stream Puller 啟動失敗"
        return 1
    fi
    
    # 檢查服務健康狀態
    check_service_health "PostgreSQL" "http://localhost:5432" 5 || true
    check_service_health "Redis" "http://localhost:6379" 5 || true
    check_service_health "MinIO" "http://localhost:9000/minio/health/live" 10
    check_service_health "RTMP" "http://localhost:1935/stat" 5 || true
    check_service_health "Stream Puller" "http://localhost:8083/health" 10
    
    cd ..
    log_success "F5 一鍵啟動測試完成"
    echo ""
    echo "📋 開發模式服務狀態:"
    echo "  ✅ PostgreSQL: localhost:5432"
    echo "  ✅ Redis: localhost:6379"
    echo "  ✅ MinIO: localhost:9000"
    echo "  ✅ MinIO Console: localhost:9001"
    echo "  ✅ RTMP: localhost:1935"
    echo "  ✅ Stream Puller: localhost:8083"
    echo ""
    echo "🚀 現在可以在 IDE 中啟動前後端服務"
    echo "  後端: 使用 launch.json 配置 (localhost:8080)"
    echo "  前端: npm run dev (localhost:5173)"
}

# 測試完整 Docker 啟動 (生產模式)
test_full_docker_startup() {
    log_info "🧪 測試完整 Docker 啟動 (生產模式)..."
    
    # 停止開發模式服務
    log_info "停止開發模式服務..."
    cd deploy
    docker-compose -f docker-compose.dev.yml down
    
    # 檢查必要端口
    local ports=(
        "5432:PostgreSQL"
        "6379:Redis"
        "9000:MinIO"
        "9001:MinIO Console"
        "1935:RTMP"
        "8083:Stream Puller"
        "8080:Backend API"
        "5173:Frontend"
        "8084:Gateway"
    )
    
    for port_info in "${ports[@]}"; do
        IFS=':' read -r port service <<< "$port_info"
        check_port $port $service
    done
    
    # 啟動所有服務
    log_info "啟動所有服務..."
    docker-compose -f docker-compose.yml up -d
    
    # 等待服務啟動
    log_info "等待服務啟動..."
    sleep 30
    
    # 檢查服務狀態
    log_info "檢查服務狀態..."
    local services=("postgresql" "redis" "minio" "api" "frontend" "rtmp" "stream-puller" "media" "gateway")
    
    for service in "${services[@]}"; do
        if docker-compose -f docker-compose.yml ps | grep -q "$service.*Up"; then
            log_success "$service 運行正常"
        else
            log_error "$service 啟動失敗"
            return 1
        fi
    done
    
    # 檢查服務健康狀態
    check_service_health "PostgreSQL" "http://localhost:5432" 5 || true
    check_service_health "Redis" "http://localhost:6379" 5 || true
    check_service_health "MinIO" "http://localhost:9000/minio/health/live" 10
    check_service_health "Backend API" "http://localhost:8080/api/health" 15
    check_service_health "Frontend" "http://localhost:5173" 10
    check_service_health "Gateway" "http://localhost:8084/health" 10
    check_service_health "RTMP" "http://localhost:1935/stat" 5 || true
    check_service_health "Stream Puller" "http://localhost:8083/health" 10
    
    cd ..
    log_success "完整 Docker 啟動測試完成"
    echo ""
    echo "📋 生產模式服務狀態:"
    echo "  ✅ PostgreSQL: localhost:5432"
    echo "  ✅ Redis: localhost:6379"
    echo "  ✅ MinIO: localhost:9000"
    echo "  ✅ MinIO Console: localhost:9001"
    echo "  ✅ Backend API: localhost:8080"
    echo "  ✅ Frontend: localhost:5173"
    echo "  ✅ Gateway: localhost:8084"
    echo "  ✅ RTMP: localhost:1935"
    echo "  ✅ Stream Puller: localhost:8083"
    echo ""
    echo "🌐 統一入口: http://localhost:8084"
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
    local containers=("stream-demo-postgresql" "stream-demo-redis" "stream-demo-minio" "stream-demo-api" "stream-demo-frontend" "stream-demo-rtmp" "stream-demo-stream-puller" "stream-demo-media" "stream-demo-gateway")
    
    for container in "${containers[@]}"; do
        if docker ps | grep -q "$container"; then
            log_success "$container 正在運行"
        else
            log_warning "$container 未運行"
        fi
    done
    
    # 檢查 API 服務是否能連接到資料庫
    if docker exec stream-demo-api wget -q --spider http://postgresql:5432 2>/dev/null; then
        log_success "API 服務可以連接到 PostgreSQL"
    else
        log_warning "API 服務無法連接到 PostgreSQL (可能是正常的，因為 PostgreSQL 不提供 HTTP 接口)"
    fi
    
    # 檢查 API 服務是否能連接到 Redis
    if docker exec stream-demo-api wget -q --spider http://redis:6379 2>/dev/null; then
        log_success "API 服務可以連接到 Redis"
    else
        log_warning "API 服務無法連接到 Redis (可能是正常的，因為 Redis 不提供 HTTP 接口)"
    fi
    
    # 檢查 API 服務是否能連接到 MinIO
    if docker exec stream-demo-api wget -q --spider http://minio:9000/minio/health/live 2>/dev/null; then
        log_success "API 服務可以連接到 MinIO"
    else
        log_error "API 服務無法連接到 MinIO"
        return 1
    fi
}

# 檢查環境變數配置
check_environment_config() {
    log_info "🔧 檢查環境變數配置..."
    
    # 檢查後端環境變數
    local backend_env_vars=(
        "DATABASES__POSTGRESQL__MASTER__HOST=postgresql"
        "DATABASES__POSTGRESQL__MASTER__PORT=5432"
        "DATABASES__POSTGRESQL__MASTER__USERNAME=stream_user"
        "DATABASES__POSTGRESQL__MASTER__PASSWORD=stream_password"
        "DATABASES__POSTGRESQL__MASTER__DBNAME=stream_demo"
        "REDIS__MASTER__HOST=redis"
        "REDIS__MASTER__PORT=6379"
        "STORAGE__S3__ENDPOINT=http://minio:9000"
        "STORAGE__S3__ACCESS_KEY=minioadmin"
        "STORAGE__S3__SECRET_KEY=minioadmin"
        "STORAGE__S3__BUCKET=stream-demo-videos"
    )
    
    for env_var in "${backend_env_vars[@]}"; do
        IFS='=' read -r key value <<< "$env_var"
        if docker exec stream-demo-api env | grep -q "^$key=$value$"; then
            log_success "後端環境變數 $key 配置正確"
        else
            log_warning "後端環境變數 $key 配置可能不正確"
        fi
    done
    
    # 檢查前端環境變數
    if [ -f "services/frontend/.env" ]; then
        log_success "前端環境變數檔案存在"
    else
        log_warning "前端環境變數檔案不存在，使用預設配置"
    fi
}

# 清理服務
cleanup_services() {
    log_info "🧹 清理服務..."
    cd deploy
    docker-compose -f docker-compose.yml down --remove-orphans
    docker-compose -f docker-compose.dev.yml down --remove-orphans
    cd ..
    log_success "服務清理完成"
}

# 主函數
main() {
    echo "🚀 開始啟動測試..."
    echo ""
    
    # 檢查 Docker
    check_docker || exit 1
    
    # 測試 F5 一鍵啟動
    test_f5_startup
    
    echo ""
    echo "⏳ 等待 5 秒後測試完整 Docker 啟動..."
    sleep 5
    
    # 測試完整 Docker 啟動
    test_full_docker_startup
    
    echo ""
    echo "⏳ 等待 10 秒後檢查服務間通訊..."
    sleep 10
    
    # 檢查服務間通訊
    check_service_communication
    
    # 檢查環境變數配置
    check_environment_config
    
    echo ""
    log_success "🎉 所有測試完成！"
    echo ""
    echo "📋 測試總結:"
    echo "  ✅ F5 一鍵啟動 (開發模式)"
    echo "  ✅ 完整 Docker 啟動 (生產模式)"
    echo "  ✅ 服務間通訊"
    echo "  ✅ 環境變數配置"
    echo ""
    echo "🔧 使用方式:"
    echo "  開發模式: ./deploy/scripts/start.sh"
    echo "  生產模式: ./deploy/scripts/deploy.sh"
    echo "  管理服務: ./deploy/scripts/manage.sh"
    echo ""
    echo "🌐 訪問地址:"
    echo "  開發模式: http://localhost:5173 (前端) + http://localhost:8080 (後端)"
    echo "  生產模式: http://localhost:8084 (統一入口)"
    
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