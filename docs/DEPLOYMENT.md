# 🚀 部署指南

## 📋 概述

本指南詳細說明如何將串流平台部署到生產環境，包含 Docker 部署、環境配置、監控設置等。

## 🐳 Docker 部署

### 前置需求

#### 伺服器要求
- **作業系統**: Linux (Ubuntu 20.04+ 推薦)
- **Docker**: 20.10+
- **Docker Compose**: 2.0+
- **硬碟空間**: 至少 50GB
- **記憶體**: 至少 4GB RAM
- **網路**: 穩定的網路連接

#### 網路要求
- **HTTP/HTTPS**: 80/443 端口
- **RTMP**: 1935 端口（直播推流）
- **HLS**: 8083 端口（直播播放）

### 快速部署

#### 1. 克隆專案
```bash
git clone <repository-url>
cd stream-demo
```

#### 2. 配置環境變數
```bash
# 複製環境變數範例
cp deploy/env/env.example deploy/env/.env

# 編輯環境變數
nano deploy/env/.env
```

#### 3. 啟動服務
```bash
# 使用部署腳本
./deploy/scripts/deploy.sh

# 或手動啟動
cd infrastructure
docker-compose -f docker-compose.yml up -d
```

#### 4. 初始化服務
```bash
# 初始化 MinIO 桶
./deploy/scripts/docker-manage.sh init

# 初始化直播服務
./deploy/scripts/docker-manage.sh init-live
```

### 生產環境配置

#### 環境變數配置
```bash
# 資料庫配置
DATABASES__POSTGRESQL__MASTER__HOST=postgresql
DATABASES__POSTGRESQL__MASTER__PORT=5432
DATABASES__POSTGRESQL__MASTER__USERNAME=stream_user
DATABASES__POSTGRESQL__MASTER__PASSWORD=<strong_password>
DATABASES__POSTGRESQL__MASTER__DBNAME=stream_demo

# Redis 配置
REDIS__MASTER__HOST=redis
REDIS__MASTER__PORT=6379
REDIS__MASTER__PASSWORD=<redis_password>

# MinIO 配置
STORAGE__S3__ENDPOINT=http://minio:9000
STORAGE__S3__ACCESS_KEY=<access_key>
STORAGE__S3__SECRET_KEY=<secret_key>
STORAGE__S3__BUCKET=stream-demo-videos

# JWT 配置
JWT__SECRET=<jwt_secret_key>

# 服務配置
GIN__HOST=0.0.0.0
GIN__PORT=8080
GIN__MODE=release
```

#### 安全配置
```bash
# 修改預設密碼
# PostgreSQL
POSTGRES_PASSWORD=<strong_postgres_password>

# MySQL
MYSQL_ROOT_PASSWORD=<strong_mysql_password>

# MinIO
MINIO_ROOT_USER=<minio_admin_user>
MINIO_ROOT_PASSWORD=<strong_minio_password>

# Redis
REDIS_PASSWORD=<strong_redis_password>
```

### 服務管理

#### 啟動服務
```bash
# 啟動所有服務
./deploy/scripts/docker-manage.sh start

# 啟動特定服務
./deploy/scripts/docker-manage.sh start postgresql
./deploy/scripts/docker-manage.sh start redis
./deploy/scripts/docker-manage.sh start minio
```

#### 停止服務
```bash
# 停止所有服務
./deploy/scripts/docker-manage.sh stop

# 停止特定服務
./deploy/scripts/docker-manage.sh stop api
./deploy/scripts/docker-manage.sh stop frontend
```

#### 重啟服務
```bash
# 重啟所有服務
./deploy/scripts/docker-manage.sh restart

# 重啟特定服務
./deploy/scripts/docker-manage.sh restart api
```

#### 查看狀態
```bash
# 查看所有服務狀態
./deploy/scripts/docker-manage.sh status

# 查看特定服務狀態
./deploy/scripts/docker-manage.sh status api
```

#### 查看日誌
```bash
# 查看所有服務日誌
./deploy/scripts/docker-manage.sh logs

# 查看特定服務日誌
./deploy/scripts/docker-manage.sh logs api
./deploy/scripts/docker-manage.sh logs postgresql
```

## 🔒 安全配置

### SSL/TLS 配置

#### 使用 Let's Encrypt
```bash
# 安裝 Certbot
sudo apt update
sudo apt install certbot

# 獲取 SSL 證書
sudo certbot certonly --standalone -d your-domain.com

# 配置 Nginx SSL
# 編輯 nginx/nginx-reverse-proxy-prod.conf
```

#### 自簽名證書（測試環境）
```bash
# 生成自簽名證書
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout nginx/ssl/nginx.key \
  -out nginx/ssl/nginx.crt
```

### 防火牆配置
```bash
# 開放必要端口
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw allow 1935/tcp
sudo ufw allow 8083/tcp

# 限制管理端口
sudo ufw allow from <your_ip> to any port 22
```

### 資料庫安全
```bash
# PostgreSQL 安全配置
# 編輯 postgresql/conf/postgresql.conf
listen_addresses = 'localhost'
max_connections = 100
shared_buffers = 256MB

# MySQL 安全配置
# 編輯 mysql/conf/my.cnf
bind-address = 127.0.0.1
max_connections = 200
```

## 📊 監控和日誌

### 日誌管理

#### 日誌配置
```bash
# 配置日誌輪轉
sudo nano /etc/logrotate.d/stream-demo

# 日誌輪轉配置
/var/log/stream-demo/*.log {
    daily
    missingok
    rotate 7
    compress
    delaycompress
    notifempty
    create 644 root root
}
```

#### 日誌收集
```bash
# 使用 Docker 日誌驅動
# 在 docker-compose.yml 中配置
services:
  backend:
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"
```

### 性能監控

#### 系統監控
```bash
# 安裝監控工具
sudo apt install htop iotop nethogs

# 監控系統資源
htop
iotop
nethogs
```

#### 應用監控
```bash
# 健康檢查端點
curl http://localhost:8080/api/health

# 服務狀態檢查
./deploy/scripts/docker-manage.sh status
```

### 備份策略

#### 資料庫備份
```bash
# PostgreSQL 備份
pg_dump -h localhost -U stream_user -d stream_demo > backup_$(date +%Y%m%d_%H%M%S).sql

# MySQL 備份
mysqldump -h localhost -u stream_user -p stream_demo > backup_$(date +%Y%m%d_%H%M%S).sql

# 自動備份腳本
#!/bin/bash
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_DIR="/backup/database"

# PostgreSQL 備份
docker exec stream-demo-postgresql pg_dump -U stream_user -d stream_demo > $BACKUP_DIR/postgres_$DATE.sql

# MySQL 備份
docker exec stream-demo-mysql mysqldump -u stream_user -pstream_password stream_demo > $BACKUP_DIR/mysql_$DATE.sql
```

#### 檔案備份
```bash
# MinIO 備份
mc mirror minio/stream-demo-videos /backup/storage

# 配置檔案備份
tar -czf config_backup_$(date +%Y%m%d_%H%M%S).tar.gz deploy/ infrastructure/
```

## 🔄 更新和維護

### 應用更新
```bash
# 拉取最新代碼
git pull origin main

# 重新建置映像
docker-compose build

# 重啟服務
docker-compose up -d
```

### 資料庫遷移
```bash
# 執行資料庫遷移
cd services/api
go run main.go migrate
```

### 清理維護
```bash
# 清理無用映像
docker image prune -f

# 清理無用容器
docker container prune -f

# 清理無用資料卷
docker volume prune -f

# 清理無用網路
docker network prune -f
```

## 🚨 故障排除

### 常見問題

#### 服務無法啟動
```bash
# 檢查 Docker 狀態
docker ps -a

# 查看服務日誌
./deploy/scripts/docker-manage.sh logs

# 檢查端口衝突
netstat -tulpn | grep :8080
```

#### 資料庫連接失敗
```bash
# 檢查資料庫容器
docker ps | grep postgres
docker ps | grep mysql

# 測試資料庫連接
docker exec -it stream-demo-postgresql psql -U stream_user -d stream_demo
```

#### 直播服務問題
```bash
# 檢查 RTMP 服務
curl http://localhost:1935/stat

# 檢查 HLS 服務
curl http://localhost:8083/test/index.m3u8

# 查看 stream-puller 日誌
./deploy/scripts/docker-manage.sh logs puller
```

### 性能問題

#### 高 CPU 使用率
```bash
# 檢查進程
top
htop

# 檢查 Docker 資源使用
docker stats
```

#### 高記憶體使用率
```bash
# 檢查記憶體使用
free -h

# 檢查 Docker 記憶體
docker stats --no-stream
```

#### 網路問題
```bash
# 檢查網路連接
ping google.com

# 檢查端口監聽
netstat -tulpn

# 檢查防火牆
sudo ufw status
```

## 📈 擴展和優化

### 水平擴展
```bash
# 擴展後端服務
docker-compose up -d --scale backend=3

# 使用負載均衡器
# 配置 Nginx 負載均衡
```

### 垂直擴展
```bash
# 增加系統資源
# 調整 Docker 資源限制
# 優化資料庫配置
```

### CDN 整合
```bash
# 配置 CDN
# 將靜態資源部署到 CDN
# 配置 HLS 串流 CDN
```

## 📞 支援

### 日誌收集
```bash
# 收集系統日誌
sudo journalctl -u docker.service

# 收集應用日誌
./deploy/scripts/docker-manage.sh logs > app_logs.txt

# 收集系統信息
uname -a > system_info.txt
```

### 聯繫方式
- **GitHub Issues**: 報告 Bug 和功能請求
- **文檔**: 查看詳細文檔
- **社群**: 加入開發者社群 