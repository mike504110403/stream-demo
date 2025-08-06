FROM nginx:alpine

# 複製 nginx 配置
COPY nginx.conf /etc/nginx/nginx.conf

# 創建 HLS 目錄並設置權限
RUN mkdir -p /var/www/hls && \
    chown -R nginx:nginx /var/www/hls && \
    chmod -R 755 /var/www/hls

# 暴露端口
EXPOSE 80

# 健康檢查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost/health || exit 1 