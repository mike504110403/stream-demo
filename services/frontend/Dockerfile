# 前端 Dockerfile (使用預構建版本)
FROM nginx:alpine

# 複製已經構建好的 dist 目錄
COPY dist/ /usr/share/nginx/html/

# 複製 nginx 配置
COPY nginx.conf /etc/nginx/conf.d/default.conf

# 暴露端口
EXPOSE 80

# 啟動 nginx
CMD ["nginx", "-g", "daemon off;"] 