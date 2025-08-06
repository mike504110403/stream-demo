FROM golang:1.21-alpine AS builder

# 安裝必要的工具
RUN apk add --no-cache git

# 設置工作目錄
WORKDIR /app

# 複製 go.mod 和 go.sum
COPY go.mod go.sum ./

# 下載依賴
RUN go mod download

# 複製源碼
COPY . .

# 編譯應用程式
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o converter .

# 運行階段
FROM alpine:3.18

# 安裝 FFmpeg 和必要的工具
RUN apk add --no-cache \
    ffmpeg \
    curl \
    wget \
    bash \
    bc \
    coreutils \
    ca-certificates \
    && rm -rf /var/cache/apk/*

# 安裝 MinIO Client
RUN wget https://dl.min.io/client/mc/release/linux-amd64/mc -O /usr/local/bin/mc \
    && chmod +x /usr/local/bin/mc

# 創建工作目錄
RUN mkdir -p /scripts /tmp/transcoding

# 複製腳本
COPY scripts/ /scripts/
RUN chmod +x /scripts/*.sh

# 從 builder 階段複製編譯好的應用程式
COPY --from=builder /app/converter /usr/local/bin/converter

# 設置工作目錄
WORKDIR /scripts

# 健康檢查
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD converter --health-check || exit 1

# 默認命令 - 運行 Go 應用程式
CMD ["converter"] 