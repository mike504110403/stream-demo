-- PostgreSQL 初始化腳本 - 串流平台專用
-- 此腳本在容器首次啟動時執行

-- 設置時區
SET timezone = 'Asia/Taipei';

-- 創建必要的擴展
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";
CREATE EXTENSION IF NOT EXISTS "btree_gin";
CREATE EXTENSION IF NOT EXISTS "btree_gist";

-- 設置資料庫配置
ALTER DATABASE stream_demo SET timezone TO 'Asia/Taipei';
ALTER DATABASE stream_demo SET log_statement TO 'all';
ALTER DATABASE stream_demo SET log_min_duration_statement TO 1000;

-- 創建測試資料庫
CREATE DATABASE stream_demo_test WITH 
    OWNER stream_user 
    ENCODING 'UTF8' 
    LC_COLLATE = 'C' 
    LC_CTYPE = 'C' 
    TEMPLATE template0;

-- 為測試資料庫創建擴展
\c stream_demo_test;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";
CREATE EXTENSION IF NOT EXISTS "btree_gin";
CREATE EXTENSION IF NOT EXISTS "btree_gist";

-- 設置測試資料庫配置
ALTER DATABASE stream_demo_test SET timezone TO 'Asia/Taipei';

-- 切換回主資料庫
\c stream_demo;

-- 創建全文搜索配置（中文支援）
CREATE TEXT SEARCH CONFIGURATION chinese_simple (COPY = simple);

-- 顯示初始化完成訊息
SELECT 'PostgreSQL initialization completed successfully!' AS status; 