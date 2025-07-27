FROM jrottenberg/ffmpeg:latest

# 安裝必要的工具
RUN apt-get update && apt-get install -y \
    curl \
    wget \
    netcat \
    python3 \
    && rm -rf /var/lib/apt/lists/*

# 安裝 MinIO Client
RUN wget https://dl.min.io/client/mc/release/linux-amd64/mc && \
    chmod +x mc && \
    mv mc /usr/local/bin/

# 覆蓋 ENTRYPOINT
ENTRYPOINT []

# 創建工作目錄
WORKDIR /app

# 複製轉碼腳本和配置
COPY scripts/transcode_live.sh /app/transcode_live.sh
COPY config/streams.conf /app/config/streams.conf
RUN chmod +x /app/transcode_live.sh

# 創建 HLS 輸出目錄
RUN mkdir -p /tmp/live

# 暴露端口
EXPOSE 8080

# 啟動轉碼服務
CMD ["bash", "/app/transcode_live.sh"] 