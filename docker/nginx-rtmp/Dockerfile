FROM tiangolo/nginx-rtmp

# 安裝必要的工具
RUN apt-get update && apt-get install -y \
    curl \
    && rm -rf /var/lib/apt/lists/*

# 複製 nginx 配置
COPY nginx.conf /etc/nginx/nginx.conf

# 創建 HLS 輸出目錄
RUN mkdir -p /tmp/live

EXPOSE 1935 8080

CMD ["nginx", "-g", "daemon off;"] 