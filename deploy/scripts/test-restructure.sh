#!/bin/bash

# 重構測試腳本
# 驗證新的服務導向架構是否正常工作

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

# 檢查目錄結構
check_directory_structure() {
    log_info "檢查目錄結構..."
    
    local required_dirs=(
        "services/api"
        "services/frontend"
        "services/rtmp-service"
        "services/stream-puller"
        "services/media-service"
        "services/gateway"
        "infrastructure/postgresql"
        "infrastructure/mysql"
        "infrastructure/redis"
        "infrastructure/minio"
        "deploy/scripts"
        "deploy/env"
    )
    
    for dir in "${required_dirs[@]}"; do
        if [ -d "$dir" ]; then
            log_success "✓ $dir"
        else
            log_error "✗ $dir (缺失)"
            return 1
        fi
    done
    
    log_success "目錄結構檢查完成"
}

# 檢查配置文件
check_config_files() {
    log_info "檢查配置文件..."
    
    local required_files=(
        "infrastructure/docker-compose.yml"
        "services/api/docker-compose.yml"
        "services/frontend/docker-compose.yml"
        "services/rtmp-service/docker-compose.yml"
        "services/stream-puller/docker-compose.yml"
        "services/media-service/docker-compose.yml"
        "services/gateway/docker-compose.yml"
        "deploy/docker-compose.yml"
        "deploy/docker-compose.dev.yml"
        "deploy/scripts/docker-manage.sh"
        "deploy/scripts/deploy.sh"
        "deploy/scripts/start.sh"
        "deploy/scripts/diagnose.sh"
    )
    
    for file in "${required_files[@]}"; do
        if [ -f "$file" ]; then
            log_success "✓ $file"
        else
            log_error "✗ $file (缺失)"
            return 1
        fi
    done
    
    log_success "配置文件檢查完成"
}

# 檢查 Docker 網路
check_docker_network() {
    log_info "檢查 Docker 網路..."
    
    if docker network ls | grep -q "stream-demo-network"; then
        log_success "✓ stream-demo-network 網路存在"
    else
        log_warning "⚠️  stream-demo-network 網路不存在，將在啟動服務時創建"
    fi
}

# 測試基礎設施服務
test_infrastructure() {
    log_info "測試基礎設施服務..."
    
    # 啟動基礎設施服務
    cd deploy
    docker-compose -f docker-compose.dev.yml up -d postgresql redis minio
    
    # 等待服務啟動
    log_info "等待服務啟動..."
    sleep 10
    
    # 檢查服務狀態
    if docker-compose -f docker-compose.dev.yml ps | grep -q "postgresql.*Up"; then
        log_success "✓ PostgreSQL 運行正常"
    else
        log_error "✗ PostgreSQL 啟動失敗"
        return 1
    fi
    
    if docker-compose -f docker-compose.dev.yml ps | grep -q "redis.*Up"; then
        log_success "✓ Redis 運行正常"
    else
        log_error "✗ Redis 啟動失敗"
        return 1
    fi
    
    if docker-compose -f docker-compose.dev.yml ps | grep -q "minio.*Up"; then
        log_success "✓ MinIO 運行正常"
    else
        log_error "✗ MinIO 啟動失敗"
        return 1
    fi
    
    # 停止服務
    docker-compose -f docker-compose.dev.yml down
    cd ..
    
    log_success "基礎設施服務測試完成"
}

# 檢查腳本路徑
check_script_paths() {
    log_info "檢查腳本路徑..."
    
    # 檢查 deploy.sh 中的路徑
    if grep -q "deploy/env/.env" deploy/scripts/deploy.sh; then
        log_success "✓ deploy.sh 路徑已更新"
    else
        log_error "✗ deploy.sh 路徑未更新"
        return 1
    fi
    
    # 檢查 docker-manage.sh 中的路徑
    if grep -q "deploy/docker-compose.yml" deploy/scripts/docker-manage.sh; then
        log_success "✓ docker-manage.sh 路徑已更新"
    else
        log_error "✗ docker-manage.sh 路徑未更新"
        return 1
    fi
    
    log_success "腳本路徑檢查完成"
}

# 主函數
main() {
    echo "🚀 開始重構測試..."
    echo ""
    
    local tests=(
        "check_directory_structure"
        "check_config_files"
        "check_docker_network"
        "check_script_paths"
        "test_infrastructure"
    )
    
    local failed_tests=()
    
    for test in "${tests[@]}"; do
        echo "📋 執行測試: $test"
        if $test; then
            log_success "測試通過: $test"
        else
            log_error "測試失敗: $test"
            failed_tests+=("$test")
        fi
        echo ""
    done
    
    # 總結
    if [ ${#failed_tests[@]} -eq 0 ]; then
        log_success "🎉 所有測試通過！重構成功！"
        echo ""
        echo "📋 重構完成清單:"
        echo "  ✓ 目錄結構重組"
        echo "  ✓ 服務分離"
        echo "  ✓ 配置文件更新"
        echo "  ✓ 路徑引用更新"
        echo "  ✓ 基礎設施服務測試"
        echo ""
        echo "🚀 現在可以使用以下命令啟動服務:"
        echo "  開發模式: ./deploy/scripts/start.sh"
        echo "  生產模式: ./deploy/scripts/deploy.sh"
        echo "  管理服務: ./deploy/scripts/manage.sh"
    else
        log_error "❌ 以下測試失敗:"
        for test in "${failed_tests[@]}"; do
            echo "  - $test"
        done
        echo ""
        echo "請檢查失敗的測試並修復問題"
        exit 1
    fi
}

# 執行主函數
main "$@" 