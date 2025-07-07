<template>
  <div class="video-detail">
    <div class="page-header">
      <h1>影片詳情</h1>
      <el-button @click="$router.back()">返回</el-button>
    </div>

    <div v-if="loading" class="loading">
      <el-skeleton :rows="5" animated />
    </div>

    <div v-else-if="video" class="video-content">
      <el-card class="video-card">
        <div class="video-player">
          <div v-if="video.video_url" class="video-container">
            <video :src="video.video_url" controls width="100%">
              您的瀏覽器不支援影片播放
            </video>
          </div>
          <div v-else class="placeholder-player">
            <div class="placeholder-icon">🎬</div>
            <p>影片處理中，請稍後...</p>
          </div>
        </div>

        <div class="video-info">
          <h2>{{ video.title }}</h2>
          <p class="video-description">{{ video.description || '暫無描述' }}</p>
          
          <div class="video-meta">
            <div class="meta-item">
              <span class="meta-label">狀態：</span>
              <el-tag :type="getStatusType(video.status)">
                {{ getStatusText(video.status) }}
              </el-tag>
            </div>
            <div class="meta-item">
              <span class="meta-label">觀看數：</span>
              <span>{{ video.views }}</span>
            </div>
            <div class="meta-item">
              <span class="meta-label">按讚數：</span>
              <span>{{ video.likes }}</span>
            </div>
            <div class="meta-item">
              <span class="meta-label">上傳時間：</span>
              <span>{{ formatDate(video.created_at) }}</span>
            </div>
          </div>

          <div class="video-actions">
            <el-button type="primary" @click="handleLike" :loading="liking">
              👍 按讚 ({{ video.likes }})
            </el-button>
            <el-button @click="$router.push(`/videos/${video.id}/edit`)">
              編輯
            </el-button>
          </div>
        </div>
      </el-card>
    </div>

    <div v-else class="error-state">
      <div class="error-icon">❌</div>
      <h3>影片不存在</h3>
      <p>找不到指定的影片，可能已被刪除</p>
      <el-button type="primary" @click="$router.push('/videos')">
        返回影片列表
      </el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { getVideo, likeVideo } from '@/api/video'
import type { Video } from '@/types'

const route = useRoute()

const loading = ref(true)
const liking = ref(false)
const video = ref<Video | null>(null)

const loadVideo = async () => {
  const videoId = Number(route.params.id)
  if (!videoId) return

  loading.value = true
  try {
    const response = await getVideo(videoId)
    video.value = response.data || response
  } catch (error) {
    console.error('載入影片失敗:', error)
  } finally {
    loading.value = false
  }
}

const handleLike = async () => {
  if (!video.value) return

  liking.value = true
  try {
    await likeVideo(video.value.id)
    video.value.likes += 1
    ElMessage.success('按讚成功！')
  } catch (error) {
    console.error('按讚失敗:', error)
  } finally {
    liking.value = false
  }
}

const getStatusType = (status: string) => {
  switch (status) {
    case 'ready': return 'success'
    case 'processing': return 'warning'
    case 'failed': return 'danger'
    default: return 'info'
  }
}

const getStatusText = (status: string) => {
  switch (status) {
    case 'ready': return '已完成'
    case 'processing': return '處理中'
    case 'failed': return '失敗'
    default: return status
  }
}

const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleDateString('zh-TW')
}

onMounted(() => {
  loadVideo()
})
</script>

<style scoped>
.video-detail {
  max-width: 1000px;
  margin: 0 auto;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.page-header h1 {
  margin: 0;
  color: #333;
}

.loading {
  padding: 40px;
}

.video-card {
  overflow: hidden;
}

.video-player {
  margin-bottom: 24px;
}

.video-container {
  position: relative;
  width: 100%;
  background: #000;
  border-radius: 8px;
  overflow: hidden;
}

.placeholder-player {
  height: 400px;
  background: #f5f5f5;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  border-radius: 8px;
}

.placeholder-icon {
  font-size: 64px;
  margin-bottom: 16px;
  color: #ccc;
}

.video-info h2 {
  margin: 0 0 16px 0;
  color: #333;
  font-size: 24px;
}

.video-description {
  color: #666;
  line-height: 1.6;
  margin-bottom: 24px;
}

.video-meta {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 16px;
  margin-bottom: 24px;
  padding: 20px;
  background: #f8f9fa;
  border-radius: 8px;
}

.meta-item {
  display: flex;
  align-items: center;
}

.meta-label {
  font-weight: bold;
  color: #666;
  margin-right: 8px;
}

.video-actions {
  display: flex;
  gap: 16px;
}

.error-state {
  text-align: center;
  padding: 80px 20px;
}

.error-icon {
  font-size: 64px;
  margin-bottom: 16px;
}

.error-state h3 {
  color: #333;
  margin-bottom: 8px;
}

.error-state p {
  color: #666;
  margin-bottom: 24px;
}
</style> 