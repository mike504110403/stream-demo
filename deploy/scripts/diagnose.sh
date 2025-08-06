#!/bin/bash

# 快速診斷腳本
# 用於檢查和解決常見問題

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
    echo -e "${CYAN}🔍 快速診斷工具${NC}"
    echo "=================================="
    echo ""
    echo "用法: $0 [命令]"
    echo ""
    echo "命令:"
    echo "  all        執行完整診斷"
    echo "  env        檢查環境依賴"
    echo "  ports      檢查端口衝突"
    echo "  docker     檢查 Docker 狀態"
    echo "  services   檢查服務狀態"
    echo "  network    檢查網路連接"
    echo "  logs       檢查錯誤日誌"
    echo "  fix        自動修復常見問題"
    echo "  help       顯示此幫助信息"
    echo ""
    echo "範例:"
    echo "  $0 all      # 執行完整診斷"
    echo "  $0 ports    # 檢查端口衝突"
    echo "  $0 fix      # 自動修復問題"
}

# 檢查環境依賴
check_environment() {
    log_step "檢查環境依賴..."
    
    local issues=0
    
    # 檢查 Docker
    if ! command -v docker >/dev/null 2>&1; then
        log_error "Docker 未安裝"
        ((issues++))
    elif ! docker info >/dev/null 2>&1; then
        log_error "Docker 未運行"
        ((issues++))
    else
        log_success "Docker 正常"
    fi
    
    # 檢查 Node.js
    if ! command -v node >/dev/null 2>&1; then
        log_error "Node.js 未安裝"
        ((issues++))
    else
        local node_version=$(node --version | cut -d'v' -f2)
        if [[ $(echo "$node_version" | cut -d'.' -f1) -lt 18 ]]; then
            log_warning "Node.js 版本過舊 ($node_version)，建議使用 18+"
        else
            log_success "Node.js v$node_version 正常"
        fi
    fi
    
    # 檢查 Go
    if ! command -v go >/dev/null 2>&1; then
        log_error "Go 未安裝"
        ((issues++))
    else
        local go_version=$(go version | awk '{print $3}' | sed 's/go//')
        log_success "Go $go_version 正常"
    fi
    
    return $issues
}

# 檢查端口衝突
check_port_conflicts() {
    log_step "檢查端口衝突..."
    
    local ports=("8080" "5173" "5432" "3306" "6379" "9000" "9001" "1935" "8083" "8084")
    local conflicts=0
    
    for port in "${ports[@]}"; do
        if lsof -i ":$port" >/dev/null 2>&1; then
            local process=$(lsof -i ":$port" | tail -n +2 | awk '{print $1}' | head -1)
            log_warning "端口 $port 被 $process 佔用"
            ((conflicts++))
        else
            log_success "端口 $port 可用"
        fi
    done
    
    return $conflicts
}

# 檢查 Docker 狀態
check_docker_status() {
    log_step "檢查 Docker 狀態..."
    
    local issues=0
    
    # 檢查 Docker 服務
    if ! docker info >/dev/null 2>&1; then
        log_error "Docker 服務未運行"
        ((issues++))
        return $issues
    fi
    
    # 檢查容器狀態
    local containers=(
        "stream-demo-postgresql"
        "stream-demo-redis"
        "stream-demo-minio"
        "stream-demo-nginx-reverse-proxy"
    )
    
    for container in "${containers[@]}"; do
        if docker ps --format "{{.Names}}" | grep -q "^$container$"; then
            local status=$(docker ps --format "{{.Status}}" --filter "name=^$container$")
            log_success "$container: $status"
        else
            log_warning "$container: 未運行"
            ((issues++))
        fi
    done
    
    return $issues
}

# 檢查服務狀態
check_services() {
    log_step "檢查服務狀態..."
    
    local services=(
        "後端 API:http://localhost:8080/api/health"
        "前端服務:http://localhost:5173"
        "統一入口:http://localhost:8084"
        "MinIO Console:http://localhost:9001"
    )
    
    local issues=0
    
    for service_info in "${services[@]}"; do
        local service_name=$(echo "$service_info" | cut -d':' -f1)
        local health_url=$(echo "$service_info" | cut -d':' -f2)
        
        if curl -s "$health_url" > /dev/null 2>&1; then
            log_success "$service_name: 正常"
        else
            log_warning "$service_name: 無響應"
            ((issues++))
        fi
    done
    
    return $issues
}

# 檢查網路連接
check_network() {
    log_step "檢查網路連接..."
    
    local issues=0
    
    # 檢查 Docker 網路
    if docker network ls | grep -q "stream-demo-network"; then
        log_success "Docker 網路 stream-demo-network 存在"
    else
        log_warning "Docker 網路 stream-demo-network 不存在"
        ((issues++))
    fi
    
    # 檢查網路連接
    if ping -c 1 google.com >/dev/null 2>&1; then
        log_success "網路連接正常"
    else
        log_warning "網路連接異常"
        ((issues++))
    fi
    
    return $issues
}

# 檢查錯誤日誌
check_logs() {
    log_step "檢查錯誤日誌..."
    
    local issues=0
    
    # 檢查 Docker 容器日誌
    local containers=(
        "stream-demo-postgresql"
        "stream-demo-redis"
        "stream-demo-minio"
        "stream-demo-nginx-reverse-proxy"
    )
    
    for container in "${containers[@]}"; do
        if docker ps --format "{{.Names}}" | grep -q "^$container$"; then
            local error_count=$(docker logs "$container" 2>&1 | grep -i "error\|failed\|exception" | wc -l)
            if [ "$error_count" -gt 0 ]; then
                log_warning "$container: 發現 $error_count 個錯誤"
                ((issues++))
            else
                log_success "$container: 日誌正常"
            fi
        fi
    done
    
    return $issues
}

# 自動修復常見問題
auto_fix() {
    log_step "自動修復常見問題..."
    
    local fixes_applied=0
    
    # 修復 1: 重啟 Docker 容器
    log_info "嘗試重啟異常容器..."
    local containers=(
        "stream-demo-postgresql"
        "stream-demo-redis"
        "stream-demo-minio"
        "stream-demo-nginx-reverse-proxy"
    )
    
    for container in "${containers[@]}"; do
        if docker ps --format "{{.Names}}" | grep -q "^$container$"; then
            local status=$(docker inspect --format='{{.State.Status}}' "$container")
            if [ "$status" != "running" ]; then
                log_info "重啟容器 $container..."
                docker restart "$container" >/dev/null 2>&1
                ((fixes_applied++))
            fi
        fi
    done
    
    # 修復 2: 清理無用資源
    log_info "清理 Docker 無用資源..."
    docker system prune -f >/dev/null 2>&1
    ((fixes_applied++))
    
    # 修復 3: 重建網路（如果不存在）
    if ! docker network ls | grep -q "stream-demo-network"; then
        log_info "重建 Docker 網路..."
        docker network create stream-demo-network >/dev/null 2>&1
        ((fixes_applied++))
    fi
    
    if [ $fixes_applied -gt 0 ]; then
        log_success "應用了 $fixes_applied 個修復"
    else
        log_info "沒有發現需要修復的問題"
    fi
}

# 執行完整診斷
full_diagnosis() {
    echo -e "${CYAN}🔍 執行完整診斷${NC}"
    echo "=================================="
    echo ""
    
    local total_issues=0
    local checks=(
        "check_environment"
        "check_port_conflicts"
        "check_docker_status"
        "check_services"
        "check_network"
        "check_logs"
    )
    
    for check in "${checks[@]}"; do
        echo ""
        $check
        total_issues=$((total_issues + $?))
    done
    
    echo ""
    echo "=================================="
    if [ $total_issues -eq 0 ]; then
        log_success "🎉 診斷完成，沒有發現問題！"
    else
        log_warning "發現 $total_issues 個問題"
        echo ""
        log_info "建議執行自動修復："
        echo "  $0 fix"
    fi
    
    return $total_issues
}

# 主函數
main() {
    case "${1:-help}" in
        all)
            full_diagnosis
            ;;
        env)
            check_environment
            ;;
        ports)
            check_port_conflicts
            ;;
        docker)
            check_docker_status
            ;;
        services)
            check_services
            ;;
        network)
            check_network
            ;;
        logs)
            check_logs
            ;;
        fix)
            auto_fix
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