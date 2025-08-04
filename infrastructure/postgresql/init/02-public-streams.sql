-- 創建公開流配置表
CREATE TABLE IF NOT EXISTS public_streams (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) UNIQUE NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    url TEXT NOT NULL,
    category VARCHAR(100),
    type VARCHAR(50) DEFAULT 'hls',
    enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 插入測試資料
INSERT INTO public_streams (name, title, description, url, category, type, enabled) VALUES
('tears_of_steel', 'Tears of Steel', 'Unified Streaming 測試影片 - 科幻短片', 'https://demo.unified-streaming.com/k8s/features/stable/video/tears-of-steel/tears-of-steel.ism/.m3u8', 'demo', 'hls', true),
('mux_test', 'Mux 測試流', 'Mux 提供的測試 HLS 流', 'https://test-streams.mux.dev/x36xhzz/x36xhzz.m3u8', 'demo', 'hls', true)
ON CONFLICT (name) DO NOTHING; 