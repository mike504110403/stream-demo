<template>
  <div class="public-stream-list">
    <!-- 頁面標題 -->
    <div class="page-header">
      <h1 class="page-title">公開直播</h1>
      <p class="page-subtitle">觀看來自全球的精彩直播內容</p>
    </div>

    <!-- 載入狀態 -->
    <div v-if="loading" class="loading-container">
      <el-loading-component />
      <p>正在載入直播列表...</p>
    </div>

    <!-- 錯誤狀態 -->
    <div v-else-if="error" class="error-container">
      <el-alert
        :title="error"
        type="error"
        :closable="false"
        show-icon
      />
      <el-button @click="loadStreams" type="primary" class="retry-button">
        重新載入
      </el-button>
    </div>

    <!-- 流列表 -->
    <div v-else-if="streams.length > 0" class="stream-grid">
      <div
        v-for="stream in filteredStreams"
        :key="stream.name"
        class="stream-card"
        @click="watchStream(stream)"
      >
        <!-- 預覽圖 -->
        <div class="preview-container">
          <div class="preview-image">
            <div class="preview-icon">
              <svg fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 002 2z"></path>
              </svg>
            </div>
            <p class="preview-text">直播中</p>
          </div>
          
          <!-- 播放按鈕覆蓋層 -->
          <div class="play-overlay">
            <div class="play-button">
              <svg fill="currentColor" viewBox="0 0 24 24">
                <path d="M8 5v14l11-7z"/>
              </svg>
            </div>
          </div>

          <!-- 狀態標籤 -->
          <div class="status-badge" :class="stream.status === 'active' ? 'active' : 'inactive'">
            <span class="status-dot" :class="{ 'pulse': stream.status === 'active' }">●</span>
            {{ stream.status === 'active' ? '直播中' : '離線' }}
          </div>

          <!-- 觀看者數量 -->
          <div class="viewer-count">
            👥 {{ stream.viewer_count }}
          </div>
        </div>

        <!-- 流資訊 -->
        <div class="stream-info">
          <h3 class="stream-title">{{ stream.title }}</h3>
          <p class="stream-description">{{ stream.description }}</p>
          
          <!-- 分類標籤 -->
          <div class="category-tag">
            {{ getCategoryLabel(stream.category) }}
          </div>

          <!-- 最後更新時間 -->
          <p class="update-time">
            <svg fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"></path>
            </svg>
            {{ formatTime(stream.last_update) }}
          </p>

          <!-- 操作按鈕 -->
          <el-button
            @click.stop="watchStream(stream)"
            :disabled="stream.status !== 'active'"
            :type="stream.status === 'active' ? 'primary' : 'info'"
            class="watch-button"
          >
            <svg fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.828 14.828a4 4 0 01-5.656 0M9 10h1m4 0h1m-6 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>
            </svg>
            {{ stream.status === 'active' ? '立即觀看' : '離線' }}
          </el-button>
        </div>
      </div>
    </div>

    <!-- 空狀態 -->
    <div v-else class="empty-container">
      <el-empty description="暫無可用的直播流" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { publicStreamApi } from '@/api/public-stream'
import type { PublicStreamInfo } from '@/types/public-stream'

const router = useRouter()

// 響應式數據
const streams = ref<PublicStreamInfo[]>([])
const loading = ref(false)
const error = ref('')

// 分類選項
const categories = [
  { value: 'all', label: '全部', icon: '🌐' },
  { value: 'test', label: '測試', icon: '🧪' },
  { value: 'space', label: '太空', icon: '🚀' },
  { value: 'news', label: '新聞', icon: '📰' },
  { value: 'sports', label: '體育', icon: '⚽' }
]

// 計算屬性
const filteredStreams = computed(() => {
  return streams.value
})

// 方法
const loadStreams = async () => {
  loading.value = true
  error.value = ''
  
  try {
    const response = await publicStreamApi.getAvailableStreams()
    streams.value = response.streams
  } catch (err) {
    console.error('載入流列表失敗:', err)
    error.value = '載入流列表失敗，請稍後再試'
  } finally {
    loading.value = false
  }
}

const watchStream = (stream: PublicStreamInfo) => {
  if (stream.status === 'active') {
    router.push(`/public-streams/${stream.name}`)
  }
}

const getCategoryLabel = (category: string) => {
  const found = categories.find(c => c.value === category)
  return found ? found.label : category
}

const formatTime = (timeString: string) => {
  const date = new Date(timeString)
  return date.toLocaleString('zh-TW')
}

// 生命週期
onMounted(() => {
  loadStreams()
})
</script>

<style scoped>
.public-stream-list {
  max-width: 1200px;
  margin: 0 auto;
  padding: 20px;
}

.page-header {
  text-align: center;
  margin-bottom: 40px;
}

.page-title {
  font-size: 2.5rem;
  font-weight: bold;
  color: #2c3e50;
  margin-bottom: 10px;
}

.page-subtitle {
  font-size: 1.1rem;
  color: #7f8c8d;
}

.loading-container,
.error-container,
.empty-container {
  text-align: center;
  padding: 60px 20px;
}

.retry-button {
  margin-top: 20px;
}

.stream-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 24px;
  margin-top: 20px;
}

.stream-card {
  background: white;
  border-radius: 16px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.1);
  overflow: hidden;
  cursor: pointer;
  transition: all 0.3s ease;
  border: 1px solid #e1e8ed;
}

.stream-card:hover {
  transform: translateY(-8px);
  box-shadow: 0 12px 40px rgba(0, 0, 0, 0.15);
}

.preview-container {
  position: relative;
  height: 200px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
}

.preview-image {
  text-align: center;
  color: white;
  z-index: 2;
}

.preview-icon {
  width: 64px;
  height: 64px;
  background: rgba(255, 255, 255, 0.2);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto 16px;
  backdrop-filter: blur(10px);
}

.preview-icon svg {
  width: 32px;
  height: 32px;
}

.preview-text {
  font-size: 14px;
  font-weight: 500;
  margin: 0;
}

.play-overlay {
  position: absolute;
  inset: 0;
  background: rgba(0, 0, 0, 0.3);
  display: flex;
  align-items: center;
  justify-content: center;
  opacity: 0;
  transition: opacity 0.3s ease;
}

.stream-card:hover .play-overlay {
  opacity: 1;
}

.play-button {
  width: 48px;
  height: 48px;
  background: rgba(255, 255, 255, 0.9);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  backdrop-filter: blur(10px);
}

.play-button svg {
  width: 24px;
  height: 24px;
  color: #667eea;
}

.status-badge {
  position: absolute;
  top: 12px;
  left: 12px;
  padding: 6px 12px;
  border-radius: 20px;
  font-size: 12px;
  font-weight: bold;
  color: white;
  display: flex;
  align-items: center;
  gap: 4px;
  backdrop-filter: blur(10px);
}

.status-badge.active {
  background: linear-gradient(135deg, #4ade80, #22c55e);
}

.status-badge.inactive {
  background: linear-gradient(135deg, #f87171, #ef4444);
}

.status-dot {
  font-size: 8px;
}

.status-dot.pulse {
  animation: pulse 2s infinite;
}

.viewer-count {
  position: absolute;
  top: 12px;
  right: 12px;
  padding: 6px 12px;
  background: rgba(0, 0, 0, 0.7);
  color: white;
  border-radius: 20px;
  font-size: 12px;
  font-weight: 500;
  backdrop-filter: blur(10px);
}

.stream-info {
  padding: 24px;
}

.stream-title {
  font-size: 18px;
  font-weight: bold;
  color: #2c3e50;
  margin: 0 0 12px 0;
  line-height: 1.4;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.stream-description {
  font-size: 14px;
  color: #7f8c8d;
  margin: 0 0 16px 0;
  line-height: 1.5;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.category-tag {
  display: inline-block;
  padding: 6px 12px;
  background: linear-gradient(135deg, #e0f2fe, #b3e5fc);
  color: #0277bd;
  border-radius: 20px;
  font-size: 12px;
  font-weight: 500;
  margin-bottom: 16px;
}

.update-time {
  font-size: 12px;
  color: #95a5a6;
  margin: 0 0 20px 0;
  display: flex;
  align-items: center;
  gap: 6px;
}

.update-time svg {
  width: 14px;
  height: 14px;
}

.watch-button {
  width: 100%;
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  font-weight: 500;
}

.watch-button svg {
  width: 16px;
  height: 16px;
}

@keyframes pulse {
  0%, 100% {
    opacity: 1;
  }
  50% {
    opacity: 0.5;
  }
}

/* 響應式設計 */
@media (max-width: 768px) {
  .stream-grid {
    grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
    gap: 16px;
  }
  
  .page-title {
    font-size: 2rem;
  }
  
  .public-stream-list {
    padding: 16px;
  }
}

@media (max-width: 480px) {
  .stream-grid {
    grid-template-columns: 1fr;
  }
  
  .page-title {
    font-size: 1.8rem;
  }
}
</style> 