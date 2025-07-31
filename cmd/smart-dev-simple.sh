#!/bin/bash

# 簡化版智能開發環境啟動腳本
# 自動檢查周邊服務狀態，如果未運行則啟動

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

# 檢查服務是否運行
check_service() {
    local service_name=$1
    local port=$2
    
    if curl -s "http://localhost:$port" > /dev/null 2>&1; then
        return 0  # 服務運行中
    else
        return 1  # 服務未運行
    fi
}

# 檢查 Docker 服務狀態
check_docker_services() {
    log_info "檢查 Docker 服務狀態..."
    
    local services_running=true
    
    # 檢查 PostgreSQL
    if ! docker ps --format "table {{.Names}}" | grep -q "stream-demo-postgres"; then
        log_warning "PostgreSQL 未運行"
        services_running=false
    else
        log_success "PostgreSQL 運行中"
    fi
    
    # 檢查 Redis
    if ! docker ps --format "table {{.Names}}" | grep -q "stream-demo-redis"; then
        log_warning "Redis 未運行"
        services_running=false
    else
        log_success "Redis 運行中"
    fi
    
    # 檢查 MinIO
    if ! docker ps --format "table {{.Names}}" | grep -q "stream-demo-minio"; then
        log_warning "MinIO 未運行"
        services_running=false
    else
        log_success "MinIO 運行中"
    fi
    
    # 檢查 Nginx 反向代理
    if ! docker ps --format "table {{.Names}}" | grep -q "stream-demo-nginx-reverse-proxy"; then
        log_warning "Nginx 反向代理未運行"
        services_running=false
    else
        log_success "Nginx 反向代理運行中"
    fi
    
    if [ "$services_running" = true ]; then
        return 0
    else
        return 1
    fi
}

# 啟動周邊服務
start_peripheral_services() {
    log_step "啟動周邊服務..."
    
    # 使用簡化的啟動方式
    cd ../docker
    docker-compose -f docker-compose.dev.yml up -d postgresql redis minio nginx-reverse-proxy
    cd ../cmd
    
    # 等待服務啟動
    log_info "等待服務啟動..."
    sleep 15
    
    # 再次檢查服務狀態
    if check_docker_services; then
        log_success "周邊服務啟動成功！"
        return 0
    else
        log_error "周邊服務啟動失敗，請檢查 Docker 狀態"
        return 1
    fi
}

# 檢查 IDE 服務
check_ide_services() {
    log_info "檢查 IDE 服務狀態..."
    
    local backend_running=false
    local frontend_running=false
    
    # 檢查後端
    if curl -s "http://localhost:8080/api/health" > /dev/null 2>&1; then
        log_success "後端 (IDE): 運行中"
        backend_running=true
    else
        log_warning "後端 (IDE): 未運行"
    fi
    
    # 檢查前端
    if curl -s "http://localhost:5173" > /dev/null 2>&1; then
        log_success "前端 (IDE): 運行中"
        frontend_running=true
    else
        log_warning "前端 (IDE): 未運行"
    fi
    
    echo ""
    
    if [ "$backend_running" = true ] && [ "$frontend_running" = true ]; then
        log_success "🎉 開發環境完全就緒！"
        return 0
    else
        return 1
    fi
}

# 顯示啟動指導
show_startup_guide() {
    echo ""
    log_info "📋 下一步操作："
    echo ""
    
    if ! curl -s "http://localhost:8080/api/health" > /dev/null 2>&1; then
        echo "🚀 啟動後端："
        echo "   在 VS Code 中使用以下任一配置："
        echo "   - 🚀 啟動後端 (PostgreSQL)"
        echo "   - 🚀 啟動後端 (MySQL)"
        echo "   - 🚀 啟動後端 (默認配置)"
        echo ""
    fi
    
    if ! curl -s "http://localhost:5173" > /dev/null 2>&1; then
        echo "🎨 啟動前端："
        echo "   在 VS Code 中使用："
        echo "   - 🎨 啟動前端 (開發模式)"
        echo ""
    fi
    
    echo "🎬 一鍵啟動前後端："
    echo "   在 VS Code 中使用以下任一配置："
    echo "   - 🎬 一鍵啟動前後端 (PostgreSQL)"
    echo "   - 🎬 一鍵啟動前後端 (MySQL)"
    echo "   - 🎬 一鍵啟動前後端 (默認)"
    echo ""
    
    echo "📊 訪問地址："
    echo "   統一入口: http://localhost:8084"
    echo "   前端 (IDE): http://localhost:5173"
    echo "   後端 (IDE): http://localhost:8080"
    echo "   MinIO Console: http://localhost:9001"
    echo "   HLS 播放: http://localhost:8083/[stream_name]/index.m3u8"
    echo "   RTMP 推流: rtmp://localhost:1935/live"
}

# 主函數
main() {
    echo -e "${CYAN}🎯 簡化版智能開發環境啟動器${NC}"
    echo "=================================="
    echo ""
    
    # 檢查周邊服務
    if check_docker_services; then
        log_success "周邊服務已運行"
    else
        log_warning "檢測到周邊服務未運行，正在啟動..."
        if ! start_peripheral_services; then
            log_error "無法啟動周邊服務，請手動檢查"
            exit 1
        fi
    fi
    
    echo ""
    
    # 檢查 IDE 服務
    if check_ide_services; then
        log_success "開發環境已完全就緒！"
    else
        show_startup_guide
    fi
}

# 執行主函數
main "$@" 