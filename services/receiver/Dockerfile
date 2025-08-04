FROM tiangolo/nginx-rtmp:latest

# 複製配置文件
COPY nginx-llhls.conf /etc/nginx/nginx.conf

# 創建必要的目錄
RUN mkdir -p /tmp/hls /tmp/hls_standard

# 暴露端口
EXPOSE 1935 80

# 使用官方 nginx-rtmp 的啟動命令
CMD ["nginx", "-g", "daemon off;"] 