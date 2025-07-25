# 🚀 Docker 快速啟動指南

## 📋 快速設置（只需要 3 步）

### 1️⃣ 確認前置條件
```bash
# 確認 Docker 已安裝並運行
docker --version
docker compose version
```

### 2️⃣ 啟動所有服務
```bash
# 一鍵啟動所有資料庫服務
./docker-manage.sh init
```

### 3️⃣ 測試連接
```bash
# 測試所有服務是否正常
./docker-manage.sh test
```

## ✅ 成功後你將得到

### 🗄️ 可用的資料庫服務
- **PostgreSQL**: `localhost:5432`
- **MySQL**: `localhost:3306`  
- **Redis**: `localhost:6379`

### 🔑 統一的登入資訊
```
用戶名: stream_user
密碼: stream_password
資料庫: stream_demo
測試資料庫: stream_demo_test
```

## 🎯 立即開始使用

### 啟動應用程式
```bash
# 使用 PostgreSQL
go run main.go

# 使用 MySQL
go run main.go -db mysql
```

### 運行測試
```bash
# 運行所有測試
go test ./tests

# 運行特定資料庫測試  
DATABASE_TYPE=mysql go test ./tests
```

## 🛠️ 常用操作

```bash
# 查看服務狀態
./docker-manage.sh status

# 查看日誌
./docker-manage.sh logs

# 停止服務
./docker-manage.sh stop

# 重啟服務
./docker-manage.sh restart

# 備份資料
./docker-manage.sh backup
```

## 🚨 問題排除

### 端口被佔用？
```bash
# 停止本地資料庫服務
sudo service postgresql stop
sudo service mysql stop  
sudo service redis-server stop
```

### 服務啟動失敗？
```bash
# 重置並重新啟動
./docker-manage.sh reset
./docker-manage.sh start
```

### 需要幫助？
```bash
# 查看完整幫助
./docker-manage.sh help

# 查看詳細文檔
cat DOCKER_GUIDE.md
```

## 💡 就是這麼簡單！

現在你可以：
- ✅ 在 PostgreSQL 和 MySQL 之間快速切換
- ✅ 使用 Redis 進行緩存和訊息傳遞
- ✅ 運行完整的多資料庫測試
- ✅ 輕鬆備份和恢復資料

開始編碼吧！🎉 