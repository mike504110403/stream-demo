# 後端 Dockerfile
FROM golang:1.24-alpine AS builder

# 設置工作目錄
WORKDIR /app

# 安裝必要的工具
RUN apk add --no-cache git

# 複製 go mod 文件
COPY go.mod go.sum ./

# 下載依賴
RUN go mod download

# 複製源代碼
COPY . .

# 構建應用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./main.go

# 生產階段
FROM alpine:latest

# 安裝 ca-certificates 用於 HTTPS 請求
RUN apk --no-cache add ca-certificates

# 創建非 root 用戶
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# 設置工作目錄
WORKDIR /app

# 複製構建結果
COPY --from=builder /app/main .

# 複製配置文件
COPY config/ ./config/

# 創建必要的目錄
RUN mkdir -p /app/logs && \
    chown -R appuser:appgroup /app

# 切換到非 root 用戶
USER appuser

# 暴露端口
EXPOSE 8080

# 啟動應用
CMD ["./main"] 