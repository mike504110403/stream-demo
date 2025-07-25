-- MySQL 初始化腳本 - 串流平台專用
-- 此腳本在容器首次啟動時執行

-- 設置時區
SET GLOBAL time_zone = '+08:00';
SET time_zone = '+08:00';

-- 創建測試資料庫
CREATE DATABASE IF NOT EXISTS stream_demo_test 
    CHARACTER SET utf8mb4 
    COLLATE utf8mb4_unicode_ci;

-- 為用戶授權主資料庫權限
GRANT ALL PRIVILEGES ON stream_demo.* TO 'stream_user'@'%';

-- 為用戶授權測試資料庫權限
GRANT ALL PRIVILEGES ON stream_demo_test.* TO 'stream_user'@'%';

-- 創建只讀用戶（用於查詢優化）
CREATE USER IF NOT EXISTS 'stream_reader'@'%' IDENTIFIED BY 'reader_password';
GRANT SELECT ON stream_demo.* TO 'stream_reader'@'%';
GRANT SELECT ON stream_demo_test.* TO 'stream_reader'@'%';

-- 設置 MySQL 性能優化參數
SET GLOBAL innodb_buffer_pool_size = 268435456; -- 256MB
SET GLOBAL max_connections = 200;
SET GLOBAL wait_timeout = 28800;
SET GLOBAL interactive_timeout = 28800;

-- 創建性能監控相關的設置
SET GLOBAL slow_query_log = 'ON';
SET GLOBAL long_query_time = 2;
SET GLOBAL log_queries_not_using_indexes = 'ON';

-- 刷新權限
FLUSH PRIVILEGES;

-- 顯示創建的資料庫
SHOW DATABASES;

-- 顯示用戶權限
SELECT User, Host FROM mysql.user WHERE User IN ('stream_user', 'stream_reader');

-- 顯示初始化完成訊息
SELECT 'MySQL initialization completed successfully!' AS status; 