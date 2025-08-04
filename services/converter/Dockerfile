FROM alpine:3.18

# 安裝 FFmpeg 和必要的工具
RUN apk add --no-cache \
    ffmpeg \
    curl \
    wget \
    bash \
    bc \
    coreutils \
    && rm -rf /var/cache/apk/*

# 安裝 MinIO Client
RUN wget https://dl.min.io/client/mc/release/linux-amd64/mc -O /usr/local/bin/mc \
    && chmod +x /usr/local/bin/mc

# 創建工作目錄
RUN mkdir -p /scripts /tmp/transcoding

# 複製腳本
COPY scripts/ /scripts/
RUN chmod +x /scripts/*.sh

# 設置工作目錄
WORKDIR /scripts

# 健康檢查
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD ffmpeg -version > /dev/null 2>&1 || exit 1

# 默認命令 - 保持容器運行
CMD ["tail", "-f", "/dev/null"] 