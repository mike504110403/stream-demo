<template>
  <div class="live-list">
    <div class="page-header">
      <h1>直播管理</h1>
      <el-button type="primary" @click="$router.push('/lives/create')">
        創建直播
      </el-button>
    </div>

    <div class="live-grid" v-loading="loading">
      <div v-if="lives.length === 0 && !loading" class="empty-state">
        <div class="empty-icon">📺</div>
        <div class="empty-text">暫無直播</div>
        <el-button type="primary" @click="$router.push('/lives/create')">
          創建第一個直播
        </el-button>
      </div>

      <el-row :gutter="20" v-else>
        <el-col :span="8" v-for="live in lives" :key="live.id">
          <el-card class="live-card">
            <div class="live-info">
              <h3 class="live-title">{{ live.title }}</h3>
              <p class="live-description">
                {{ live.description || '暫無描述' }}
              </p>

              <div class="live-meta">
                <div class="meta-item">
                  <span class="meta-label">狀態：</span>
                  <el-tag :type="getStatusType(live.status)">
                    {{ getStatusText(live.status) }}
                  </el-tag>
                </div>
                <div class="meta-item">
                  <span class="meta-label">觀看人數：</span>
                  <span>{{ live.viewer_count }}</span>
                </div>
                <div class="meta-item">
                  <span class="meta-label">開始時間：</span>
                  <span>{{ formatDate(live.started_at) }}</span>
                </div>
              </div>
            </div>

            <div class="live-actions">
              <el-button size="small" @click="viewLive(live.id)"
                >詳情</el-button
              >
              <el-button
                v-if="live.status === 'created'"
                size="small"
                type="success"
                @click="startLive(live.id)"
              >
                開始直播
              </el-button>
              <el-button
                v-if="live.status === 'live'"
                size="small"
                type="warning"
                @click="endLive(live.id)"
              >
                結束直播
              </el-button>
            </div>
          </el-card>
        </el-col>
      </el-row>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import {
  getActiveRooms,
  startLive as startLiveAPI,
  endLive as endLiveAPI,
} from '@/api/live-room'
import type { LiveRoomInfo } from '@/types'

const router = useRouter()

const loading = ref(false)
const lives = ref<LiveRoomInfo[]>([])

const loadLives = async () => {
  loading.value = true
  try {
    const response = await getActiveRooms()
    lives.value = response || []
  } catch (error) {
    console.error('載入直播間失敗:', error)
  } finally {
    loading.value = false
  }
}

const viewLive = (id: string) => {
  router.push(`/live-rooms/${id}`)
}

const startLive = async (id: string) => {
  try {
    await startLiveAPI(id)
    ElMessage.success('直播已開始！')
    loadLives()
  } catch (error) {
    console.error('開始直播失敗:', error)
  }
}

const endLive = async (id: string) => {
  try {
    await endLiveAPI(id)
    ElMessage.success('直播已結束！')
    loadLives()
  } catch (error) {
    console.error('結束直播失敗:', error)
  }
}

const getStatusType = (status: string) => {
  switch (status) {
    case 'live':
      return 'success'
    case 'created':
      return 'info'
    case 'ended':
      return 'danger'
    case 'cancelled':
      return 'warning'
    default:
      return 'info'
  }
}

const getStatusText = (status: string) => {
  switch (status) {
    case 'live':
      return '直播中'
    case 'created':
      return '已創建'
    case 'ended':
      return '已結束'
    case 'cancelled':
      return '已取消'
    default:
      return status
  }
}

const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleString('zh-TW')
}

onMounted(() => {
  loadLives()
})
</script>

<style scoped>
.live-list {
  max-width: 1200px;
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

.live-grid {
  min-height: 400px;
}

.empty-state {
  text-align: center;
  padding: 80px 0;
}

.empty-icon {
  font-size: 64px;
  margin-bottom: 16px;
}

.empty-text {
  font-size: 18px;
  color: #666;
  margin-bottom: 24px;
}

.live-card {
  margin-bottom: 20px;
  height: 280px;
  display: flex;
  flex-direction: column;
}

.live-info {
  flex: 1;
}

.live-title {
  margin: 0 0 8px 0;
  font-size: 16px;
  font-weight: bold;
  color: #333;
}

.live-description {
  margin: 0 0 16px 0;
  color: #666;
  font-size: 14px;
}

.live-meta {
  margin-bottom: 16px;
}

.meta-item {
  display: flex;
  margin-bottom: 8px;
}

.meta-label {
  font-weight: bold;
  color: #666;
  width: 80px;
}

.live-actions {
  display: flex;
  gap: 8px;
  justify-content: flex-end;
}
</style>
