-- 初始化公開直播源配置
-- 創建 public_streams 表（如果不存在）
CREATE TABLE IF NOT EXISTS public_streams (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    title VARCHAR(200) NOT NULL,
    description TEXT,
    url TEXT NOT NULL,
    category VARCHAR(50) DEFAULT 'test',
    type VARCHAR(20) DEFAULT 'hls',
    enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 插入預設的公開直播源
INSERT INTO public_streams (name, title, description, url, category, type, enabled) VALUES
(
    'tears_of_steel',
    'Tears of Steel',
    'Unified Streaming 測試影片 - 科幻短片，展示 HLS 流媒體技術',
    'https://demo.unified-streaming.com/k8s/features/stable/video/tears-of-steel/tears-of-steel.ism/.m3u8',
    'test',
    'hls',
    true
),
(
    'mux_test',
    'Mux 測試流',
    'Mux 提供的測試 HLS 流，用於開發和測試',
    'https://test-streams.mux.dev/x36xhzz/x36xhzz.m3u8',
    'test',
    'hls',
    true
),
(
    'big_buck_bunny',
    'Big Buck Bunny',
    'Blender 開源動畫短片，經典的測試影片',
    'https://commondatastorage.googleapis.com/gtv-videos-bucket/sample/BigBuckBunny.mp4',
    'entertainment',
    'hls',
    true
),
(
    'elephants_dream',
    'Elephants Dream',
    'Blender 基金會的第一部開源動畫電影',
    'https://commondatastorage.googleapis.com/gtv-videos-bucket/sample/ElephantsDream.mp4',
    'entertainment',
    'hls',
    true
),
(
    'sintel',
    'Sintel',
    'Blender 基金會的開源動畫短片，講述一個女孩尋找龍的故事',
    'https://commondatastorage.googleapis.com/gtv-videos-bucket/sample/Sintel.mp4',
    'entertainment',
    'hls',
    true
),
(
    'space_station',
    '國際太空站直播',
    'NASA 提供的國際太空站實時直播流',
    'https://www.nasa.gov/multimedia/nasatv/iss_ustream.m3u8',
    'space',
    'hls',
    false
),
(
    'earth_view',
    '地球視角直播',
    '從太空看地球的實時視角',
    'https://www.nasa.gov/multimedia/nasatv/earth_ustream.m3u8',
    'space',
    'hls',
    false
)
ON CONFLICT (name) DO UPDATE SET
    title = EXCLUDED.title,
    description = EXCLUDED.description,
    url = EXCLUDED.url,
    category = EXCLUDED.category,
    type = EXCLUDED.type,
    enabled = EXCLUDED.enabled,
    updated_at = CURRENT_TIMESTAMP;

-- 創建索引以提高查詢性能
CREATE INDEX IF NOT EXISTS idx_public_streams_enabled ON public_streams(enabled);
CREATE INDEX IF NOT EXISTS idx_public_streams_category ON public_streams(category);
CREATE INDEX IF NOT EXISTS idx_public_streams_type ON public_streams(type);

-- 添加註釋
COMMENT ON TABLE public_streams IS '公開直播源配置表';
COMMENT ON COLUMN public_streams.name IS '流的唯一名稱';
COMMENT ON COLUMN public_streams.title IS '顯示標題';
COMMENT ON COLUMN public_streams.description IS '流描述';
COMMENT ON COLUMN public_streams.url IS '直播源 URL';
COMMENT ON COLUMN public_streams.category IS '分類 (test, space, news, sports, entertainment)';
COMMENT ON COLUMN public_streams.type IS '流類型 (hls, rtmp)';
COMMENT ON COLUMN public_streams.enabled IS '是否啟用'; 