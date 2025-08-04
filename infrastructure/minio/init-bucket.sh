#!/bin/bash

# MinIO 初始化腳本
echo "🚀 初始化 MinIO 桶..."

# 設置 MinIO 客戶端別名
mc alias set local http://localhost:9000 minioadmin minioadmin

# 創建原始影片桶（如果不存在）
echo "📦 創建原始影片桶..."
mc mb local/stream-demo-videos --ignore-existing

# 創建處理後影片桶（如果不存在）
echo "📦 創建處理後影片桶..."
mc mb local/stream-demo-processed --ignore-existing

# 設置桶的公開讀取權限
echo "🔓 設置桶權限..."
mc anonymous set public local/stream-demo-videos
mc anonymous set public local/stream-demo-processed

# MinIO 不需要預先創建目錄，會在需要時自動創建
echo "📁 目錄結構會在需要時自動創建..."

echo "✅ MinIO 初始化完成！"
echo "📊 桶列表："
mc ls local

echo "📁 目錄結構："
mc ls local/stream-demo-videos --recursive
mc ls local/stream-demo-processed --recursive 