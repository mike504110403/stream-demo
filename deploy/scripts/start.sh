#!/bin/bash

# 開發環境一鍵啟動腳本
# 整合了智能檢查、依賴安裝、服務啟動等功能

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
    echo -e "${CYAN}🚀 開發環境一鍵啟動腳本${NC}"
    echo "=================================="
    echo ""
    echo "用法: $0 [命令] [選項]"
    echo ""
    echo "命令:"
    echo "  start     啟動開發環境 (智能檢查並啟動)"
    echo "  stop      停止開發環境"
    echo "  restart   重啟開發環境"
    echo "  status    查看開發環境狀態"
    echo "  logs      查看服務日誌"
    echo "  check     檢查環境依賴"
    echo "  health    執行全面健康檢查"
    echo "  ports     檢查端口佔用情況"
    echo "  help      顯示此幫助信息"
    echo ""
    echo "選項:"
    echo "  --force   強制重新啟動服務"
    echo "  --clean   清理並重建服務"
    echo ""
    echo "範例:"
    echo "  $0 start         # 智能啟動開發環境"
    echo "  $0 start --force # 強制重新啟動"
    echo "  $0 status        # 查看狀態"
    echo "  $0 stop          # 停止服務"
}

# 檢查命令是否存在
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# 檢查端口是否被佔用
check_port() {
    local port=$1
    local service_name=$2
    
    if lsof -i ":$port" >/dev/null 2>&1; then
        log_warning "$service_name 端口 $port 已被佔用"
        echo ""
        log_info "解決方案："
        echo "  1. 停止佔用端口的服務"
        echo "  2. 修改配置使用其他端口"
        echo "  3. 使用 --force 選項強制啟動"
        return 1
    fi
    
    return 0
}

# 檢查關鍵端口
check_ports() {
    log_info "檢查端口佔用情況..."
    
    local ports_to_check=(
        "8080:後端 API"
        "5173:前端開發服務器"
        "5432:PostgreSQL"
        "3306:MySQL"
        "6379:Redis"
        "9000:MinIO API"
        "9001:MinIO Console"
        "1935:RTMP"
        "8083:HLS"
        "8084:統一入口"
    )
    
    local has_conflict=false
    
    for port_info in "${ports_to_check[@]}"; do
        local port=$(echo "$port_info" | cut -d':' -f1)
        local service=$(echo "$port_info" | cut -d':' -f2)
        
        if ! check_port "$port" "$service"; then
            has_conflict=true
        fi
    done
    
    if [ "$has_conflict" = true ]; then
        echo ""
        log_warning "檢測到端口衝突，可能會影響服務啟動"
        return 1
    fi
    
    log_success "端口檢查通過"
    return 0
}

# 檢查 Docker 是否運行
check_docker() {
    if ! command_exists docker; then
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

# 檢查 Node.js 和 npm
check_node() {
    if ! command_exists node; then
        log_error "Node.js 未安裝"
        echo ""
        log_info "安裝 Node.js："
        echo "  macOS: brew install node 或 https://nodejs.org/"
        echo "  Linux: curl -fsSL https://deb.nodesource.com/setup_lts.x | sudo -E bash - && sudo apt-get install -y nodejs"
        echo "  Windows: https://nodejs.org/"
        return 1
    fi
    
    if ! command_exists npm; then
        log_error "npm 未安裝，請重新安裝 Node.js"
        return 1
    fi
    
    # 檢查版本
    local node_version=$(node --version | cut -d'v' -f2)
    local npm_version=$(npm --version)
    
    # 檢查 Node.js 版本是否 >= 18
    if [[ $(echo "$node_version" | cut -d'.' -f1) -lt 18 ]]; then
        log_warning "Node.js 版本過舊 ($node_version)，建議使用 18+ 版本"
    fi
    
    log_success "Node.js v$node_version 和 npm v$npm_version 已安裝"
    return 0
}

# 檢查 Go
check_go() {
    if ! command_exists go; then
        log_error "Go 未安裝"
        echo ""
        log_info "安裝 Go："
        echo "  macOS: brew install go 或 https://go.dev/dl/"
        echo "  Linux: https://go.dev/dl/ 或使用包管理器"
        echo "  Windows: https://go.dev/dl/"
        return 1
    fi
    
    # 檢查版本
    local go_version=$(go version | awk '{print $3}' | sed 's/go//')
    local major_version=$(echo "$go_version" | cut -d'.' -f1)
    local minor_version=$(echo "$go_version" | cut -d'.' -f2)
    
    # 檢查 Go 版本是否 >= 1.24
    if [[ $major_version -lt 1 ]] || [[ $major_version -eq 1 && $minor_version -lt 24 ]]; then
        log_warning "Go 版本過舊 ($go_version)，建議使用 1.24+ 版本"
    fi
    
    log_success "Go $go_version 已安裝"
    return 0
}

# 安裝前端依賴
install_frontend_deps() {
    log_step "安裝前端依賴..."
    
    if [ ! -d "services/frontend/node_modules" ]; then
        log_info "前端依賴未安裝，正在安裝..."
        cd services/frontend
        npm install
        cd ../..
        log_success "前端依賴安裝完成"
    else
        log_success "前端依賴已安裝"
    fi
}

# 安裝後端依賴
install_backend_deps() {
    log_step "安裝後端依賴..."
    
    cd services/api
    go mod download
    go mod tidy
    cd ../..
    log_success "後端依賴安裝完成"
}

# 檢查 Docker 服務狀態
check_docker_services() {
    log_info "檢查 Docker 服務狀態..."
    
    local services_running=true
    
    # 檢查 PostgreSQL
    if ! docker ps --format "table {{.Names}}" | grep -q "stream-demo-postgresql"; then
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
    
    # 檢查 Nginx 反向代理 (Gateway)
    if ! docker ps --format "table {{.Names}}" | grep -q "stream-demo-gateway"; then
        log_warning "Nginx 反向代理 (Gateway) 未運行"
        services_running=false
    else
        log_success "Nginx 反向代理 (Gateway) 運行中"
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
    
    # 檢查是否要強制重建
    if [[ "$*" == *"--clean"* ]]; then
        log_info "清理並重建服務..."
        ./deploy/scripts/manage.sh stop
        docker system prune -f
    fi
    
    # 啟動開發模式服務
            ./deploy/scripts/manage.sh start-dev
    
    # 等待服務啟動
    log_info "等待服務啟動..."
    sleep 10
    
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

# 檢查服務健康狀態
check_service_health() {
    local service_name=$1
    local health_url=$2
    
    if curl -s "$health_url" > /dev/null 2>&1; then
        log_success "$service_name: 健康"
        return 0
    else
        log_warning "$service_name: 不健康"
        return 1
    fi
}

# 全面健康檢查
perform_health_check() {
    log_step "執行全面健康檢查..."
    
    local services=(
        "後端 API:http://localhost:8080/api/health"
        "前端服務:http://localhost:5173"
        "統一入口:http://localhost:8084"
        "MinIO Console:http://localhost:9001"
        "HLS 服務:http://localhost:8083"
    )
    
    local all_healthy=true
    
    for service_info in "${services[@]}"; do
        local service_name=$(echo "$service_info" | cut -d':' -f1)
        local health_url=$(echo "$service_info" | cut -d':' -f2)
        
        if ! check_service_health "$service_name" "$health_url"; then
            all_healthy=false
        fi
    done
    
    echo ""
    
    if [ "$all_healthy" = true ]; then
        log_success "🎉 所有服務運行正常！"
        return 0
    else
        log_warning "部分服務可能未正常運行"
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

# 啟動開發環境
start_dev_environment() {
    echo -e "${CYAN}🎯 智能開發環境啟動器${NC}"
    echo "=================================="
    echo ""
    
    # 檢查依賴
    log_step "檢查環境依賴..."
    if ! check_docker; then
        exit 1
    fi
    
    if ! check_node; then
        exit 1
    fi
    
    if ! check_go; then
        exit 1
    fi
    
    echo ""
    
    # 檢查端口衝突（除非使用 --force）
    if [[ "$*" != *"--force"* ]]; then
        if ! check_ports; then
            echo ""
            log_info "使用 --force 選項跳過端口檢查："
            echo "  ./deploy/scripts/start.sh start --force"
            echo ""
            read -p "是否繼續啟動？(y/N): " -n 1 -r
            echo
            if [[ ! $REPLY =~ ^[Yy]$ ]]; then
                log_info "啟動已取消"
                exit 0
            fi
        fi
    else
        log_info "跳過端口檢查（--force 模式）"
    fi
    
    echo ""
    
    # 安裝依賴
    install_frontend_deps
    install_backend_deps
    
    echo ""
    
    # 檢查周邊服務
    if check_docker_services; then
        log_success "周邊服務已運行"
    else
        log_warning "檢測到周邊服務未運行，正在啟動..."
        if ! start_peripheral_services "$@"; then
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

# 停止開發環境
stop_dev_environment() {
    log_info "🛑 停止開發環境..."
    
    # 停止周邊服務
            ./deploy/scripts/manage.sh stop
    
    log_success "開發環境已停止"
}

# 重啟開發環境
restart_dev_environment() {
    log_info "🔄 重啟開發環境..."
    stop_dev_environment
    sleep 2
    start_dev_environment "$@"
}

# 查看開發環境狀態
check_dev_status() {
    log_info "📊 查看開發環境狀態..."
    
    # 檢查周邊服務狀態
            ./deploy/scripts/manage.sh dev-status
    
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
        ./deploy/scripts/manage.sh dev-logs
    else
        log_info "查看 $service 服務日誌..."
        ./deploy/scripts/manage.sh dev-logs "$service"
    fi
}

# 檢查環境依賴
check_environment() {
    log_info "🔍 檢查環境依賴..."
    
    echo "Docker: $(docker --version 2>/dev/null || echo '未安裝')"
    echo "Node.js: $(node --version 2>/dev/null || echo '未安裝')"
    echo "npm: $(npm --version 2>/dev/null || echo '未安裝')"
    echo "Go: $(go version 2>/dev/null || echo '未安裝')"
    
    echo ""
    log_info "檢查 Docker 服務狀態..."
    check_docker_services
    
    echo ""
    log_info "檢查 IDE 服務狀態..."
    check_ide_services
}

# 主函數
main() {
    case "${1:-help}" in
        start)
            start_dev_environment "$@"
            ;;
        stop)
            stop_dev_environment
            ;;
        restart)
            restart_dev_environment "$@"
            ;;
        status)
            check_dev_status
            ;;
        logs)
            show_logs "$2"
            ;;
        check)
            check_environment
            ;;
        health)
            perform_health_check
            ;;
        ports)
            check_ports
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