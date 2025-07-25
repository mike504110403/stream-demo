# 🐳 Docker 環境使用指南

## 📋 概述

本專案使用 Docker 和 Docker Compose 來管理本地開發環境，包含 **PostgreSQL**、**MySQL**、**Redis** 三個核心資料庫服務。

## 🎯 支援的服務

### 🗄️ 核心資料庫服務
- **PostgreSQL 15** - 主要資料庫 (端口: 5432)
- **MySQL 8.0** - 備選資料庫 (端口: 3306)
- **Redis 7** - 緩存和訊息服務 (端口: 6379)

## 🚀 快速開始

### 1. 前置需求

確保您的系統已安裝：
- **Docker** (版本 20.10+)
- **Docker Compose** (版本 2.0+)

### 2. 初始化環境

```bash
# 給腳本執行權限
chmod +x docker-manage.sh

# 初始化開發環境（一鍵設置）
./docker-manage.sh init
```

### 3. 基本操作

```bash
# 啟動所有服務
./docker-manage.sh start

# 查看服務狀態
./docker-manage.sh status

# 查看日誌
./docker-manage.sh logs

# 停止所有服務
./docker-manage.sh stop
```

## 📊 服務配置詳情

### PostgreSQL 配置

```yaml
服務名稱: postgresql
容器名稱: stream-demo-postgres
映像: postgres:15-alpine
端口: 5432
```

**連接信息:**
- 主機: `localhost:5432`
- 資料庫: `stream_demo`
- 用戶: `stream_user`
- 密碼: `stream_password`
- 測試資料庫: `stream_demo_test`

**特殊功能:**
- ✅ 自動創建測試資料庫
- ✅ 預裝擴展 (uuid-ossp, pg_trgm, btree_gin, btree_gist)
- ✅ 中文全文搜索支援
- ✅ 時區設定為 Asia/Taipei

### MySQL 配置

```yaml
服務名稱: mysql
容器名稱: stream-demo-mysql
映像: mysql:8.0
端口: 3306
```

**連接信息:**
- 主機: `localhost:3306`
- 資料庫: `stream_demo`
- 用戶: `stream_user`
- 密碼: `stream_password`
- 測試資料庫: `stream_demo_test`
- Root 密碼: `root_password`

**特殊功能:**
- ✅ UTF8MB4 字符集支援
- ✅ 自動創建測試資料庫
- ✅ 性能優化配置
- ✅ 慢查詢日誌啟用

### Redis 配置

```yaml
服務名稱: redis
容器名稱: stream-demo-redis
映像: redis:7-alpine
端口: 6379
```

**連接信息:**
- 主機: `localhost:6379`
- 密碼: (無)
- 資料庫數量: 16

**DB 分配:**
- `DB 0`: 默認資料庫
- `DB 1`: 應用緩存
- `DB 2`: 訊息佇列
- `DB 13-15`: 測試環境

## 🔧 進階操作

### 查看特定服務日誌

```bash
# 查看 PostgreSQL 日誌
./docker-manage.sh logs postgresql

# 查看 MySQL 日誌
./docker-manage.sh logs mysql

# 查看 Redis 日誌
./docker-manage.sh logs redis
```

### 測試服務連接

```bash
# 測試所有服務連接
./docker-manage.sh test
```

### 備份和恢復

```bash
# 備份所有資料庫
./docker-manage.sh backup

# 備份檔案位置
ls ./backups/
```

### 資料重置

```bash
# 重置所有數據（危險操作）
./docker-manage.sh reset
```

### 清理環境

```bash
# 清理未使用的容器和映像
./docker-manage.sh clean
```

## 🧪 與應用程式整合

### 配置文件設置

確保您的 `config/config.local.yaml` 使用相同的連接參數：

```yaml
databases:
  postgresql:
    type: "postgresql"
    master:
      host: "localhost"
      port: 5432
      username: "stream_user"
      password: "stream_password"
      dbname: "stream_demo"
      sslmode: "disable"
  
  mysql:
    type: "mysql"  
    master:
      host: "localhost"
      port: 3306
      username: "stream_user"
      password: "stream_password"
      dbname: "stream_demo"
      sslmode: "false"

redis:
  master:
    host: "localhost"
    port: 6379
    password: ""
    db: 0
```

### 環境變數設置

```bash
# 複製環境變數範例
cp env.example .env

# 根據需要修改 .env 文件
vim .env
```

### 應用程式啟動順序

```bash
# 1. 啟動 Docker 服務
./docker-manage.sh start

# 2. 等待服務就緒
./docker-manage.sh test

# 3. 啟動應用程式（PostgreSQL）
go run main.go

# 4. 或啟動應用程式（MySQL）
go run main.go -db mysql
```

## 🔄 資料庫切換

### 使用命令行參數

```bash
# 使用 PostgreSQL
go run main.go -db postgresql

# 使用 MySQL  
go run main.go -db mysql
```

### 使用環境變數

```bash
# 設置環境變數
export DATABASE_TYPE=mysql
go run main.go

# 或者
DATABASE_TYPE=postgresql go run main.go
```

## 🚨 故障排除

### 常見問題

#### 1. 端口佔用錯誤

```bash
# 檢查端口使用情況
lsof -i :5432  # PostgreSQL
lsof -i :3306  # MySQL
lsof -i :6379  # Redis

# 停止衝突的服務
sudo service postgresql stop
sudo service mysql stop
sudo service redis-server stop
```

#### 2. 容器啟動失敗

```bash
# 查看詳細錯誤信息
./docker-manage.sh logs

# 重置環境
./docker-manage.sh reset
./docker-manage.sh start
```

#### 3. 連接被拒絕

```bash
# 檢查服務健康狀態
./docker-manage.sh status

# 等待服務完全啟動
sleep 30
./docker-manage.sh test
```

#### 4. 數據持久化問題

```bash
# 檢查 Docker 卷
docker volume ls | grep stream-demo

# 如果卷損壞，重置數據
./docker-manage.sh reset
```

### 日誌分析

```bash
# PostgreSQL 連接日誌
./docker-manage.sh logs postgresql | grep "connection"

# MySQL 錯誤日誌
./docker-manage.sh logs mysql | grep "ERROR"

# Redis 命令日誌
./docker-manage.sh logs redis | grep "COMMAND"
```

### 性能監控

```bash
# 檢查容器資源使用
docker stats stream-demo-postgres stream-demo-mysql stream-demo-redis

# 檢查資料庫連接數
docker exec stream-demo-postgres psql -U stream_user -d stream_demo -c "SELECT count(*) FROM pg_stat_activity;"
docker exec stream-demo-mysql mysql -u stream_user -pstream_password -e "SHOW STATUS LIKE 'Threads_connected';"
```

## 🔐 安全最佳實踐

### 開發環境

- ✅ 使用非默認密碼
- ✅ 限制網路訪問範圍
- ✅ 定期更新容器映像
- ✅ 啟用日誌監控

### 生產環境注意事項

⚠️ **本配置僅適用於開發環境**

生產環境建議：
- 🔒 啟用 SSL/TLS 加密
- 🔒 使用強密碼
- 🔒 限制網路訪問
- 🔒 啟用審計日誌
- 🔒 定期備份策略

## 📚 參考資源

### Docker 官方文檔
- [Docker 安裝指南](https://docs.docker.com/get-docker/)
- [Docker Compose 文檔](https://docs.docker.com/compose/)

### 資料庫文檔
- [PostgreSQL Docker](https://hub.docker.com/_/postgres)
- [MySQL Docker](https://hub.docker.com/_/mysql)
- [Redis Docker](https://hub.docker.com/_/redis)

---

## 💡 快速參考

### 常用命令

```bash
# 基本操作
./docker-manage.sh start     # 啟動服務
./docker-manage.sh stop      # 停止服務
./docker-manage.sh status    # 查看狀態
./docker-manage.sh test      # 測試連接

# 查看日誌
./docker-manage.sh logs      # 所有服務日誌
./docker-manage.sh logs postgresql  # PostgreSQL 日誌
./docker-manage.sh logs mysql       # MySQL 日誌
./docker-manage.sh logs redis       # Redis 日誌

# 維護操作
./docker-manage.sh backup    # 備份數據
./docker-manage.sh clean     # 清理環境
./docker-manage.sh reset     # 重置數據
```

### 服務地址

```bash
# 資料庫服務
PostgreSQL: localhost:5432
MySQL:       localhost:3306
Redis:       localhost:6379
```

### 資料庫切換

```bash
# 命令行參數
go run main.go -db postgresql
go run main.go -db mysql

# 環境變數
export DATABASE_TYPE=mysql
go run main.go
```

有任何 Docker 相關問題，請參考此指南或聯繫開發團隊！🐳 