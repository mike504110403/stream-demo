#!/bin/bash

echo "🔍 測試 API 服務狀態..."

# 檢查後端服務是否運行
echo "📡 檢查後端服務 (localhost:8080)..."
if curl -s http://localhost:8080/api/health > /dev/null; then
    echo "✅ 後端服務正在運行"
    curl -s http://localhost:8080/api/health | jq . 2>/dev/null || curl -s http://localhost:8080/api/health
else
    echo "❌ 後端服務未運行或無法訪問"
fi

echo ""
echo "📡 檢查前端服務 (localhost:5173)..."
if curl -s http://localhost:5173 > /dev/null; then
    echo "✅ 前端服務正在運行"
else
    echo "❌ 前端服務未運行或無法訪問"
fi

echo ""
echo "📡 測試 API 代理..."
if curl -s http://localhost:5173/api/health > /dev/null; then
    echo "✅ API 代理工作正常"
    curl -s http://localhost:5173/api/health | jq . 2>/dev/null || curl -s http://localhost:5173/api/health
else
    echo "❌ API 代理有問題"
fi 