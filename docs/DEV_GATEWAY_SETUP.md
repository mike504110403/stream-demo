# 開發環境 Gateway 設置

## 問題描述

用戶發現開發環境缺少 gateway 服務，無法提供反向代理功能。在開發模式下，前後端通過 IDE 啟動，需要一個統一的入口來訪問所有服務。

## 解決方案

### 1. 添加開發環境 Gateway 服務

**修改文件**: `deploy/docker-compose.dev.yml`

**新增服務配置**:
```yaml
# 開發環境反向代理 (Gateway)
gateway:
  build:
    context: ../services/gateway
    dockerfile: Dockerfile.reverse-proxy-dev
  image: stream-demo-gateway:latest
  container_name: stream-demo-gateway
  restart: unless-stopped
  ports:
    - "8084:80"  # 開發環境使用 8084 端口
  networks:
    - stream-demo-network
  depends_on:
    - rtmp
  healthcheck:
    test: ["CMD", "curl", "-f", "http://localhost/health"]
    interval: 30s
    timeout: 10s
    retries: 3
```

### 2. 創建開發環境專用 Dockerfile

**新增文件**: `services/gateway/Dockerfile.reverse-proxy-dev`

```dockerfile
FROM nginx:alpine

# 安裝必要的工具
RUN apk add --no-cache curl envsubst

# 複製開發模式配置文件
COPY nginx-reverse-proxy-dev.conf /etc/nginx/conf.d/default.conf

# 創建日誌目錄
RUN mkdir -p /var/log/nginx

# 暴露端口
EXPOSE 80

# 使用官方 nginx 的啟動命令
CMD ["nginx", "-g", "daemon off;"]
```

### 3. 開發環境 Nginx 配置

**文件**: `services/gateway/nginx-reverse-proxy-dev.conf`

**主要特點**:
- 前端代理到 `host.docker.internal:5173` (IDE 啟動的前端)
- 後端代理到 `host.docker.internal:8080` (IDE 啟動的後端)
- 串流服務代理到容器內服務
- 提供開發模式狀態檢查端點

**關鍵配置**:
```nginx
upstream frontend {
    # 開發模式: 直接連接到主機的 5173 端口 (IDE 啟動的前端)
    server host.docker.internal:5173;
}

upstream backend {
    # 開發模式: 直接連接到主機的 8080 端口 (IDE 啟動的後端)
    server host.docker.internal:8080;
}

# 開發模式狀態頁面
location /dev-status {
    access_log off;
    return 200 "Development Mode: true\nFrontend: host.docker.internal:5173\nBackend: host.docker.internal:8080\n";
    add_header Content-Type text/plain;
}
```

### 4. 更新腳本

**修改文件**: `deploy/scripts/docker-manage.sh`

**更新內容**:
- 修復路徑問題 (移除 `deploy/` 前綴)
- 在開發環境健康檢查中添加 `gateway` 服務
- 確保所有 docker-compose 命令使用正確的項目名稱

## 開發環境架構

### 服務組成
- **基礎設施**: PostgreSQL, MySQL, Redis, MinIO
- **串流服務**: RTMP, Stream-Puller, Media
- **反向代理**: Gateway (Nginx)

### 端口分配
- **Gateway**: `8084` (統一入口)
- **前端 (IDE)**: `5173`
- **後端 (IDE)**: `8080`
- **RTMP**: `1935`
- **Stream-Puller**: `8083`
- **MinIO**: `9000/9001`
- **PostgreSQL**: `5432`
- **MySQL**: `3306`
- **Redis**: `6379`

### 訪問地址
- **統一入口**: http://localhost:8084
- **前端 (IDE)**: http://localhost:5173
- **後端 (IDE)**: http://localhost:8080
- **MinIO Console**: http://localhost:9001
- **HLS 播放**: http://localhost:8083/[stream_name]/index.m3u8
- **RTMP 推流**: rtmp://localhost:1935/live

## 使用方式

### 啟動開發環境
```bash
./scripts/docker-manage.sh start-dev
```

### 檢查狀態
```bash
./scripts/docker-manage.sh dev-status
```

### 查看日誌
```bash
./scripts/docker-manage.sh dev-logs
```

## 驗證結果

✅ Gateway 服務成功啟動  
✅ 健康檢查正常  
✅ 開發模式狀態頁面正常  
✅ 反向代理配置正確  
✅ 腳本路徑修復完成  
✅ 所有服務健康狀態正常  

## 注意事項

1. 開發環境的 Gateway 使用端口 8084，避免與生產環境衝突
2. 前端和後端需要通過 IDE 啟動，Gateway 會代理到主機的相應端口
3. 使用 `host.docker.internal` 來訪問主機上 IDE 啟動的服務
4. 開發模式狀態頁面提供服務連接信息
5. 所有容器名稱都有 `stream-demo-` 前綴，符合命名規範 