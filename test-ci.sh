#!/bin/bash

# GitHub Actions 本地測試腳本
set -e

echo "🚀 本地 CI 測試"
echo "==============="

# 檢查依賴
if ! command -v act &> /dev/null; then
    echo "❌ 請先安裝 act: brew install act"
    exit 1
fi

if ! docker ps &> /dev/null; then
    echo "🔧 啟動 Docker..."
    open -a Docker
    sleep 10
fi

# 執行測試
echo "🔧 執行 GitHub Actions 測試..."
act --container-architecture linux/amd64

echo "✅ 測試完成！" 