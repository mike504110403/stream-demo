FROM golang:1.24.3-alpine AS builder

# 安裝必要的工具
RUN apk add --no-cache git

# 設置工作目錄
WORKDIR /app

# 複製整個後端目錄
COPY . .

# 下載依賴
RUN go mod download

# 編譯 stream-puller
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o stream-puller ./cmd/stream_puller

# 運行階段
FROM alpine:latest

# 安裝 FFmpeg 和 curl
RUN apk add --no-cache ffmpeg curl

# 創建非 root 用戶
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# 設置工作目錄
WORKDIR /app

# 從構建階段複製二進制文件
COPY --from=builder /app/stream-puller .

# 創建輸出目錄
RUN mkdir -p /tmp/public_streams && \
    chown -R appuser:appgroup /tmp/public_streams

# 切換到非 root 用戶
USER appuser

# 暴露端口
EXPOSE 8081

# 健康檢查
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8081/health || exit 1

# 啟動命令
CMD ["./stream-puller", "-output", "/tmp/public_streams", "-port", "8081"] 