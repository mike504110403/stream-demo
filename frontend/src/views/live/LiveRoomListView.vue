<template>
  <div class="live-room-list">
    <div class="page-header">
      <h1>直播間列表</h1>
      <div class="header-actions">
        <el-button type="primary" @click="$router.push('/live-rooms/create')">
          創建直播間
        </el-button>
      </div>
    </div>

    <div class="room-grid" v-loading="loading">
      <div v-if="rooms.length === 0 && !loading" class="empty-state">
        <div class="empty-icon">🏠</div>
        <div class="empty-text">暫無直播間</div>
        <div class="empty-subtext">包括活躍和已結束的直播間</div>
        <el-button type="primary" @click="$router.push('/live-rooms/create')">
          創建第一個直播間
        </el-button>
      </div>

      <el-row :gutter="20" v-else>
        <el-col :span="8" v-for="room in rooms" :key="room.id">
          <el-card class="room-card" shadow="hover">
            <div class="room-info">
              <h3 class="room-title">{{ room.title }}</h3>
              <p class="room-description">
                {{ room.description || '暫無描述' }}
              </p>

              <div class="room-meta">
                <div class="meta-item">
                  <span class="meta-label">狀態：</span>
                  <el-tag :type="getStatusType(room.status)">
                    {{ getStatusText(room.status) }}
                  </el-tag>
                </div>
                <div class="meta-item">
                  <span class="meta-label">觀眾：</span>
                  <span>{{ room.viewer_count }}/{{ room.max_viewers }}</span>
                </div>
                <div class="meta-item">
                  <span class="meta-label">創建時間：</span>
                  <span>{{ formatDate(room.created_at) }}</span>
                </div>
                <div v-if="room.started_at" class="meta-item">
                  <span class="meta-label">開始時間：</span>
                  <span>{{ formatDate(room.started_at) }}</span>
                </div>
              </div>
            </div>

            <div class="room-actions">
              <el-button
                v-if="room.status === 'cancelled'"
                size="small"
                type="info"
                @click="viewRoom(room.id)"
              >
                查看回放
              </el-button>
              <el-button v-else size="small" @click="viewRoom(room.id)">
                進入直播間
              </el-button>
              <el-button
                v-if="
                  room.creator_id === currentUserId &&
                  (room.status === 'created' || room.status === 'ended')
                "
                size="small"
                type="success"
                @click="startLive(room.id)"
              >
                {{ room.status === 'ended' ? '重新開始直播' : '開始直播' }}
              </el-button>
              <el-button
                v-if="
                  room.creator_id === currentUserId && room.status === 'live'
                "
                size="small"
                type="warning"
                @click="endLive(room.id)"
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
  getAllRooms,
  startLive as startLiveAPI,
  endLive as endLiveAPI,
} from '@/api/live-room'
import { useAuthStore } from '@/store/auth'
import type { LiveRoomInfo } from '@/types'

const router = useRouter()
const authStore = useAuthStore()

const loading = ref(false)
const rooms = ref<LiveRoomInfo[]>([])
const currentUserId = authStore.user?.id

const loadRooms = async () => {
  loading.value = true
  try {
    console.log('開始載入直播間...')
    const response = await getAllRooms({ limit: 50 })
    console.log('API 響應:', response)
    rooms.value = response || []
    console.log('設置的直播間數據:', rooms.value)
  } catch (error) {
    console.error('載入直播間失敗:', error)
    ElMessage.error('載入直播間失敗')
    // 確保在錯誤時清空數據
    rooms.value = []
  } finally {
    loading.value = false
  }
}

const viewRoom = (roomId: string) => {
  router.push(`/live-rooms/${roomId}`)
}

const startLive = async (roomId: string) => {
  try {
    await startLiveAPI(roomId)
    ElMessage.success('直播開始成功')
    await loadRooms() // 重新載入列表
  } catch (error) {
    console.error('開始直播失敗:', error)
    ElMessage.error('開始直播失敗')
  }
}

const endLive = async (roomId: string) => {
  try {
    await endLiveAPI(roomId)
    ElMessage.success('直播結束成功')
    await loadRooms() // 重新載入列表
  } catch (error) {
    console.error('結束直播失敗:', error)
    ElMessage.error('結束直播失敗')
  }
}

const getStatusType = (status: string) => {
  switch (status) {
    case 'created':
      return 'info'
    case 'live':
      return 'success'
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
    case 'created':
      return '已創建'
    case 'live':
      return '直播中'
    case 'ended':
      return '已結束'
    case 'cancelled':
      return '已取消'
    default:
      return status
  }
}

const formatDate = (dateString: string) => {
  if (!dateString) return '未知'
  return new Date(dateString).toLocaleString('zh-TW')
}

onMounted(() => {
  loadRooms()
})
</script>

<style scoped>
.live-room-list {
  padding: 20px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 30px;
}

.page-header h1 {
  margin: 0;
  color: #333;
}

.room-grid {
  min-height: 400px;
}

.empty-state {
  text-align: center;
  padding: 60px 20px;
}

.empty-icon {
  font-size: 64px;
  margin-bottom: 20px;
}

.empty-text {
  font-size: 18px;
  color: #666;
  margin-bottom: 20px;
}

.room-card {
  margin-bottom: 20px;
  transition: transform 0.2s;
}

.room-card:hover {
  transform: translateY(-2px);
}

.room-info {
  margin-bottom: 15px;
}

.room-title {
  margin: 0 0 10px 0;
  font-size: 18px;
  color: #333;
}

.room-description {
  color: #666;
  margin: 0 0 15px 0;
  line-height: 1.5;
}

.room-meta {
  margin-bottom: 15px;
}

.meta-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
  font-size: 14px;
}

.meta-label {
  color: #666;
  font-weight: 500;
}

.room-actions {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
}
</style>
