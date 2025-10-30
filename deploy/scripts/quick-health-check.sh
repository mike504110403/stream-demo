#!/bin/bash

# 簡單的周邊服務檢查腳本 - 專為 VSCode 調試器設計
# 只檢查服務狀態，失敗時提供啟動命令

set -e

# 靜默模式檢查函數
check_service_silent() {
    local service=$1
    if docker ps --format "table {{.Names}}\t{{.Status}}" | grep -q "stream-demo-${service}.*Up" 2>/dev/null; then
        return 0
    else
        return 1
    fi
}

# 主要邏輯
main() {
    # 檢查 Docker 是否運行（靜默）
    if ! docker info > /dev/null 2>&1; then
        echo ""
        echo "❌ Docker 未運行"
        echo "請先啟動 Docker 應用程式"
        echo ""
        exit 1
    fi

    # 檢查核心服務狀態
    local missing_services=()
    
    if ! check_service_silent "postgresql"; then
        missing_services+=("PostgreSQL")
    fi
    
    if ! check_service_silent "redis"; then
        missing_services+=("Redis")
    fi
    
    # 如果有服務未運行
    if [ ${#missing_services[@]} -gt 0 ]; then
        echo ""
        echo "⚠️  以下周邊服務未運行: ${missing_services[*]}"
        echo ""
        echo "🚀 請複製並執行以下命令啟動周邊服務:"
        echo ""
        echo "   ./deploy/scripts/docker-manage.sh start-dev"
        echo ""
        echo "然後重新按 F5 啟動應用"
        echo ""
        exit 1
    else
        echo "✅ 周邊服務已運行"
        exit 0
    fi
}

# 執行檢查
main "$@"