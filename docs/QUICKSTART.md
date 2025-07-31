# 🚀 快速啟動指南

## 📋 一鍵啟動開發環境

```bash
# 啟動周邊服務
./dev.sh start

# 查看狀態
./dev.sh status

# 停止環境
./dev.sh stop
```

## 🎯 分步驟啟動

### 1. 啟動周邊服務
```bash
./dev.sh start
# 或
./manage.sh start-dev
```

### 2. 啟動前後端 (在 IDE 中)
```bash
# 後端
cd backend && go run main.go

# 前端
cd frontend && npm run dev
```

## 🌐 訪問地址

- **統一入口**: http://localhost:8084
- **前端 (IDE)**: http://localhost:5173
- **後端 (IDE)**: http://localhost:8080
- **MinIO Console**: http://localhost:9001
- **HLS 播放**: http://localhost:8083/[stream_name]/index.m3u8
- **RTMP 推流**: rtmp://localhost:1935/live

## 🔧 常用命令

```bash
# 開發環境管理
./dev.sh start      # 啟動完整開發環境
./dev.sh stop       # 停止開發環境
./dev.sh status     # 查看狀態
./dev.sh logs       # 查看日誌

# 服務管理
./manage.sh start-dev    # 啟動周邊服務
./manage.sh stop         # 停止所有服務
./manage.sh dev-status   # 查看服務狀態
./manage.sh dev-logs     # 查看服務日誌

# 查看特定服務日誌
./dev.sh logs nginx-reverse-proxy
./dev.sh logs postgresql
```

## 📚 詳細文檔

- [README.md](./README.md) - 完整專案說明
- [DEVELOPMENT.md](./DEVELOPMENT.md) - 開發模式詳細指南 