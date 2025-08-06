#!/bin/bash

echo "🔍 測試資料庫連接和資料..."

# 檢查 PostgreSQL 是否運行
echo "📡 檢查 PostgreSQL 服務..."
if docker ps | grep -q "stream-demo-postgresql"; then
    echo "✅ PostgreSQL 容器正在運行"
else
    echo "❌ PostgreSQL 容器未運行"
    exit 1
fi

# 檢查資料庫連接
echo "📡 測試資料庫連接..."
if docker exec stream-demo-postgresql pg_isready -U stream_user -d stream_demo; then
    echo "✅ 資料庫連接正常"
else
    echo "❌ 資料庫連接失敗"
    exit 1
fi

# 檢查 public_streams 表
echo "📡 檢查 public_streams 表..."
docker exec stream-demo-postgresql psql -U stream_user -d stream_demo -c "\dt public_streams"

# 查詢所有記錄
echo "📡 查詢所有記錄..."
docker exec stream-demo-postgresql psql -U stream_user -d stream_demo -c "SELECT id, name, title, enabled FROM public_streams;"

# 查詢啟用的記錄
echo "📡 查詢啟用的記錄..."
docker exec stream-demo-postgresql psql -U stream_user -d stream_demo -c "SELECT id, name, title, enabled FROM public_streams WHERE enabled = true;"

echo "✅ 資料庫測試完成" 