#!/bin/bash

# 一鍵啟動腳本
# 支援自動安裝依賴和智能啟動

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

# 檢查命令是否存在
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# 檢查 Docker 是否運行
check_docker() {
    if ! command_exists docker; then
        log_error "Docker 未安裝，請先安裝 Docker"
        return 1
    fi
    
    if ! docker info >/dev/null 2>&1; then
        log_error "Docker 未運行，請啟動 Docker Desktop"
        return 1
    fi
    
    log_success "Docker 已啟動"
    return 0
}

# 檢查 Node.js 和 npm
check_node() {
    if ! command_exists node; then
        log_warning "Node.js 未安裝，正在安裝..."
        if [[ "$OSTYPE" == "darwin"* ]]; then
            # macOS
            if command_exists brew; then
                brew install node
            else
                log_error "請先安裝 Homebrew，然後運行: brew install node"
                return 1
            fi
        elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
            # Linux
            curl -fsSL https://deb.nodesource.com/setup_lts.x | sudo -E bash -
            sudo apt-get install -y nodejs
        else
            log_error "不支援的操作系統，請手動安裝 Node.js"
            return 1
        fi
    fi
    
    if ! command_exists npm; then
        log_error "npm 未安裝，請重新安裝 Node.js"
        return 1
    fi
    
    log_success "Node.js $(node --version) 和 npm $(npm --version) 已安裝"
    return 0
}

# 檢查 Go
check_go() {
    if ! command_exists go; then
        log_warning "Go 未安裝，正在安裝..."
        if [[ "$OSTYPE" == "darwin"* ]]; then
            # macOS
            if command_exists brew; then
                brew install go
            else
                log_error "請先安裝 Homebrew，然後運行: brew install go"
                return 1
            fi
        elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
            # Linux
            wget https://go.dev/dl/go1.24.3.linux-amd64.tar.gz
            sudo tar -C /usr/local -xzf go1.24.3.linux-amd64.tar.gz
            echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
            source ~/.bashrc
            rm go1.24.3.linux-amd64.tar.gz
        else
            log_error "不支援的操作系統，請手動安裝 Go"
            return 1
        fi
    fi
    
    log_success "Go $(go version) 已安裝"
    return 0
}

# 安裝前端依賴
install_frontend_deps() {
    log_step "安裝前端依賴..."
    
    if [ ! -d "frontend/node_modules" ]; then
        cd frontend
        npm install
        cd ..
        log_success "前端依賴安裝完成"
    else
        log_success "前端依賴已存在"
    fi
}

# 安裝後端依賴
install_backend_deps() {
    log_step "安裝後端依賴..."
    
    cd backend
    go mod download
    go mod tidy
    cd ..
    log_success "後端依賴安裝完成"
}

# 檢查並啟動周邊服務
start_peripheral_services() {
    log_step "檢查並啟動周邊服務..."
    
    # 檢查服務是否已運行
    if docker ps --format "table {{.Names}}" | grep -q "postgres"; then
        log_success "周邊服務已運行"
        return 0
    fi
    
    # 檢查是否有同名容器但未運行
    local containers_to_clean=(
        "postgres"
        "redis"
        "minio"
        "nginx-reverse-proxy"
        "nginx-rtmp"
        "stream-puller"
        "ffmpeg-transcoder"
    )
    
    local need_cleanup=false
    for container in "${containers_to_clean[@]}"; do
        if docker ps -a --format "table {{.Names}}" | grep -q "$container"; then
            if ! docker ps --format "table {{.Names}}" | grep -q "$container"; then
                log_warning "發現未運行的容器: $container"
                need_cleanup=true
            fi
        fi
    done
    
    # 如果需要清理，詢問用戶是否要重新構建
    if [ "$need_cleanup" = true ]; then
        echo ""
        log_info "發現未運行的容器，請選擇操作："
        echo "  [1] 跳過 - 直接啟動現有容器"
        echo "  [2] 清理重建 - 清理舊容器並重新構建映像檔"
        echo "  [3] 強制重建 - 清理所有容器並強制重新構建（無快取）"
        echo ""
        read -p "請選擇 (1/2/3，預設為 1): " choice
        
        case "${choice:-1}" in
            1)
                log_info "選擇跳過，直接啟動現有容器..."
                ;;
            2)
                log_info "選擇清理重建..."
                cleanup_and_rebuild false
                ;;
            3)
                log_info "選擇強制重建（無快取）..."
                cleanup_and_rebuild true
                ;;
            *)
                log_info "無效選擇，使用預設值：跳過"
                ;;
        esac
    fi
    
    # 啟動周邊服務
    log_info "啟動周邊服務..."
    cd docker
    docker-compose -f docker-compose.dev.yml --project-name stream-demo up -d postgresql redis minio nginx-reverse-proxy
    cd ..
    
    # 等待服務啟動
    log_info "等待服務啟動..."
    sleep 15
    
    # 檢查服務狀態
    if docker ps --format "table {{.Names}}" | grep -q "postgres"; then
        log_success "周邊服務啟動成功"
        return 0
    else
        log_error "周邊服務啟動失敗"
        return 1
    fi
}

# 清理並重建函數
cleanup_and_rebuild() {
    local no_cache=$1
    
    log_info "清理舊容器和映像檔..."
    cd docker
    
    # 停止並移除所有相關容器
    docker-compose -f docker-compose.dev.yml --project-name stream-demo down --remove-orphans
    
            # 強制移除可能殘留的容器
        local containers_to_clean=(
            "postgres"
            "redis"
            "minio"
            "nginx-reverse-proxy"
            "nginx-rtmp"
            "stream-puller"
            "ffmpeg-transcoder"
        )
    
    for container in "${containers_to_clean[@]}"; do
        docker rm -f "$container" 2>/dev/null || true
    done
    
    # 清理網路
    docker network rm docker_stream-demo-network 2>/dev/null || true
    docker network rm stream-demo_stream-demo-network 2>/dev/null || true
    
    # 清理相關映像檔
    log_info "清理相關映像檔..."
    docker rmi stream-demo-nginx-reverse-proxy:latest 2>/dev/null || true
    docker rmi stream-demo-stream-puller:latest 2>/dev/null || true
    docker rmi stream-demo-ffmpeg-transcoder:latest 2>/dev/null || true
    docker rmi stream-demo-nginx-rtmp:latest 2>/dev/null || true
    
    cd ..
    
    # 重新構建映像檔
    log_info "重新構建映像檔..."
    cd docker
    
    if [ "$no_cache" = true ]; then
        log_info "使用 --no-cache 強制重新構建..."
        docker-compose -f docker-compose.dev.yml --project-name stream-demo build --no-cache nginx-reverse-proxy stream-puller ffmpeg-transcoder
    else
        log_info "使用快取重新構建..."
        docker-compose -f docker-compose.dev.yml --project-name stream-demo build nginx-reverse-proxy stream-puller ffmpeg-transcoder
    fi
    
    cd ..
    
    log_success "清理重建完成"
}

# 啟動後端
start_backend() {
    log_step "啟動後端服務..."
    
    # 檢查後端是否已運行
    if curl -s "http://localhost:8080/api/health" > /dev/null 2>&1; then
        log_success "後端已運行"
        return 0
    fi
    
    # 檢查是否有舊的後端進程或端口被佔用
    local need_restart=false
    if [ -f "logs/backend.pid" ]; then
        local old_pid=$(cat logs/backend.pid)
        if ps -p $old_pid > /dev/null 2>&1; then
            log_warning "發現舊的後端進程 (PID: $old_pid)"
            need_restart=true
        fi
    fi
    
    if lsof -i :8080 > /dev/null 2>&1; then
        log_warning "端口 8080 被佔用"
        need_restart=true
    fi
    
    # 如果需要重啟，詢問用戶
    if [ "$need_restart" = true ]; then
        echo ""
        log_info "後端服務需要重啟，請選擇操作："
        echo "  [1] 自動重啟 - 停止舊進程並重新啟動"
        echo "  [2] 跳過 - 保持現狀"
        echo ""
        read -p "請選擇 (1/2，預設為 1): " choice
        
        case "${choice:-1}" in
            1)
                log_info "選擇自動重啟後端..."
                # 停止舊進程
                if [ -f "logs/backend.pid" ]; then
                    local old_pid=$(cat logs/backend.pid)
                    if ps -p $old_pid > /dev/null 2>&1; then
                        log_info "停止舊的後端進程 (PID: $old_pid)..."
                        kill $old_pid 2>/dev/null || true
                        sleep 2
                    fi
                    rm -f logs/backend.pid
                fi
                
                # 清理端口
                if lsof -i :8080 > /dev/null 2>&1; then
                    log_info "清理端口 8080..."
                    lsof -ti :8080 | xargs kill -9 2>/dev/null || true
                    sleep 2
                fi
                ;;
            2)
                log_info "選擇跳過後端重啟"
                return 0
                ;;
            *)
                log_info "無效選擇，使用預設值：自動重啟"
                ;;
        esac
    fi
    
    # 啟動後端
    cd backend
    nohup go run main.go -config config/config.local.yaml -env local -db postgresql > ../logs/backend.log 2>&1 &
    BACKEND_PID=$!
    cd ..
    
    # 等待後端啟動
    log_info "等待後端啟動..."
    for i in {1..30}; do
        if curl -s "http://localhost:8080/api/health" > /dev/null 2>&1; then
            log_success "後端啟動成功 (PID: $BACKEND_PID)"
            echo $BACKEND_PID > logs/backend.pid
            return 0
        fi
        sleep 1
    done
    
    log_error "後端啟動失敗"
    return 1
}

# 啟動前端
start_frontend() {
    log_step "啟動前端服務..."
    
    # 檢查前端是否已運行
    if curl -s "http://localhost:5173" > /dev/null 2>&1; then
        log_success "前端已運行"
        return 0
    fi
    
    # 檢查是否有舊的前端進程或端口被佔用
    local need_restart=false
    if [ -f "logs/frontend.pid" ]; then
        local old_pid=$(cat logs/frontend.pid)
        if ps -p $old_pid > /dev/null 2>&1; then
            log_warning "發現舊的前端進程 (PID: $old_pid)"
            need_restart=true
        fi
    fi
    
    if lsof -i :5173 > /dev/null 2>&1; then
        log_warning "端口 5173 被佔用"
        need_restart=true
    fi
    
    # 如果需要重啟，詢問用戶
    if [ "$need_restart" = true ]; then
        echo ""
        log_info "前端服務需要重啟，請選擇操作："
        echo "  [1] 自動重啟 - 停止舊進程並重新啟動"
        echo "  [2] 跳過 - 保持現狀"
        echo ""
        read -p "請選擇 (1/2，預設為 1): " choice
        
        case "${choice:-1}" in
            1)
                log_info "選擇自動重啟前端..."
                # 停止舊進程
                if [ -f "logs/frontend.pid" ]; then
                    local old_pid=$(cat logs/frontend.pid)
                    if ps -p $old_pid > /dev/null 2>&1; then
                        log_info "停止舊的前端進程 (PID: $old_pid)..."
                        kill $old_pid 2>/dev/null || true
                        sleep 2
                    fi
                    rm -f logs/frontend.pid
                fi
                
                # 清理端口
                if lsof -i :5173 > /dev/null 2>&1; then
                    log_info "清理端口 5173..."
                    lsof -ti :5173 | xargs kill -9 2>/dev/null || true
                    sleep 2
                fi
                ;;
            2)
                log_info "選擇跳過前端重啟"
                return 0
                ;;
            *)
                log_info "無效選擇，使用預設值：自動重啟"
                ;;
        esac
    fi
    
    # 啟動前端
    cd frontend
    nohup npm run dev -- --port 5173 --host 0.0.0.0 > ../logs/frontend.log 2>&1 &
    FRONTEND_PID=$!
    cd ..
    
    # 等待前端啟動
    log_info "等待前端啟動..."
    for i in {1..30}; do
        if curl -s "http://localhost:5173" > /dev/null 2>&1; then
            log_success "前端啟動成功 (PID: $FRONTEND_PID)"
            echo $FRONTEND_PID > logs/frontend.pid
            return 0
        fi
        sleep 1
    done
    
    log_error "前端啟動失敗"
    return 1
}

# 顯示啟動結果
show_startup_result() {
    echo ""
    log_success "🎉 開發環境啟動完成！"
    echo ""
    echo "📊 訪問地址："
    echo "   統一入口: http://localhost:8084"
    echo "   前端: http://localhost:5173"
    echo "   後端: http://localhost:8080"
    echo "   MinIO Console: http://localhost:9001"
    echo ""
    echo "🔧 管理命令："
    echo "   停止服務: ./cmd/dev.sh stop"
    echo "   查看狀態: ./cmd/dev.sh status"
    echo "   查看日誌: ./cmd/dev.sh logs"
    echo ""
    echo "📝 日誌文件："
    echo "   後端日誌: logs/backend.log"
    echo "   前端日誌: logs/frontend.log"
}

# 主函數
main() {
    echo -e "${CYAN}🚀 一鍵啟動開發環境${NC}"
    echo "================================"
    echo ""
    
    # 創建日誌目錄
    mkdir -p logs
    
    # 檢查環境
    log_step "檢查開發環境..."
    
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
    
    # 安裝依賴
    log_step "安裝項目依賴..."
    
    if ! install_frontend_deps; then
        log_error "前端依賴安裝失敗"
        exit 1
    fi
    
    if ! install_backend_deps; then
        log_error "後端依賴安裝失敗"
        exit 1
    fi
    
    echo ""
    
    # 啟動服務
    log_step "啟動開發服務..."
    
    if ! start_peripheral_services; then
        log_error "周邊服務啟動失敗"
        exit 1
    fi
    
    if ! start_backend; then
        log_error "後端啟動失敗"
        exit 1
    fi
    
    if ! start_frontend; then
        log_error "前端啟動失敗"
        exit 1
    fi
    
    echo ""
    
    # 顯示結果
    show_startup_result
}

# 執行主函數
main "$@" 