<template>
  <div class="live-room">
    <div class="page-header">
      <h1>直播間</h1>
      <div class="header-actions">
        <el-button @click="$router.back()">返回</el-button>
        <el-button 
          v-if="liveInfo?.user_id === currentUserId" 
          type="primary" 
          @click="showStreamInfo = true"
        >
          串流資訊
        </el-button>
      </div>
    </div>

    <div v-if="loading" class="loading-container">
      <el-skeleton :rows="3" animated />
    </div>

    <div v-else-if="error" class="error-container">
      <el-result
        icon="error"
        :title="error"
        sub-title="無法載入直播資訊"
      >
        <template #extra>
          <el-button type="primary" @click="loadLiveInfo">
            重新載入
          </el-button>
        </template>
      </el-result>
    </div>

    <div v-else-if="liveInfo" class="live-content">
      <!-- 直播播放器和聊天室 -->
      <div class="live-layout">
        <!-- 左側：直播播放器 -->
        <div class="player-section">
          <LivePlayer
            :stream-url="streamUrl"
            :live-info="liveInfo"
            :auto-play="true"
          />
        </div>

        <!-- 右側：聊天室 -->
        <div class="chat-section">
          <LiveChat
            :live-id="liveId"
            :current-user-id="currentUserId"
            :current-username="currentUsername"
            :chat-enabled="liveInfo.chat_enabled"
          />
        </div>
      </div>

      <!-- 直播詳情 -->
      <div class="live-details">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>直播詳情</span>
              <div class="live-status">
                <el-tag :type="getStatusType(liveInfo.status)">
                  {{ getStatusText(liveInfo.status) }}
                </el-tag>
                <span class="viewer-count">
                  {{ liveInfo.viewer_count }} 觀看
                </span>
              </div>
            </div>
          </template>
          
          <div class="details-content">
            <h2>{{ liveInfo.title }}</h2>
            <p v-if="liveInfo.description" class="description">
              {{ liveInfo.description }}
            </p>
            
            <div class="meta-info">
              <div class="meta-item">
                <span class="label">主播：</span>
                <span class="value">{{ liveInfo.user?.username || '未知用戶' }}</span>
              </div>
              <div class="meta-item">
                <span class="label">開始時間：</span>
                <span class="value">{{ formatStartTime(liveInfo.start_time) }}</span>
              </div>
              <div class="meta-item" v-if="liveInfo.end_time">
                <span class="label">結束時間：</span>
                <span class="value">{{ formatStartTime(liveInfo.end_time) }}</span>
              </div>
            </div>
          </div>
        </el-card>
      </div>
    </div>

    <!-- 串流資訊對話框 -->
    <el-dialog
      v-model="showStreamInfo"
      title="串流資訊"
      width="500px"
    >
      <div class="stream-info">
        <div class="info-item">
          <label>串流金鑰：</label>
          <div class="key-display">
            <el-input
              v-model="liveInfo?.stream_key"
              readonly
              size="small"
            />
            <el-button
              type="primary"
              size="small"
              @click="copyStreamKey"
            >
              複製
            </el-button>
          </div>
        </div>
        
        <div class="info-item">
          <label>RTMP 推流地址：</label>
          <div class="key-display">
            <el-input
              v-model="rtmpUrl"
              readonly
              size="small"
            />
            <el-button
              type="primary"
              size="small"
              @click="copyRtmpUrl"
            >
              複製
            </el-button>
          </div>
        </div>
        
        <div class="info-item">
          <label>HLS 播放地址：</label>
          <div class="key-display">
            <el-input
              v-model="hlsUrl"
              readonly
              size="small"
            />
            <el-button
              type="primary"
              size="small"
              @click="copyHlsUrl"
            >
              複製
            </el-button>
          </div>
        </div>
        
        <el-alert
          title="推流說明"
          type="info"
          :closable="false"
          show-icon
        >
          <p>1. 使用 OBS 或其他推流軟體</p>
          <p>2. 設置推流地址為上述 RTMP 地址</p>
          <p>3. 串流金鑰請妥善保管，不要外洩</p>
        </el-alert>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { getLive } from '@/api/live'
import { useAuthStore } from '@/store/auth'
import LivePlayer from '@/components/live/LivePlayer.vue'
import LiveChat from '@/components/live/LiveChat.vue'
import type { Live } from '@/types'

const route = useRoute()
const authStore = useAuthStore()

// 響應式數據
const loading = ref(true)
const error = ref('')
const liveInfo = ref<Live | null>(null)
const showStreamInfo = ref(false)

// 計算屬性
const liveId = computed(() => Number(route.params.id))
const currentUserId = computed(() => authStore.user?.id || 0)
const currentUsername = computed(() => authStore.user?.username || '')

// 串流 URL
const streamUrl = computed(() => {
  if (!liveInfo.value) return ''
  
  // 根據直播狀態返回不同的串流 URL
  if (liveInfo.value.status === 'live') {
    // 使用 HLS 直播流
    return `http://localhost:8081/${liveInfo.value.stream_key}/index.m3u8`
  }
  
  return ''
})

const rtmpUrl = computed(() => {
  if (!liveInfo.value) return ''
  return `rtmp://localhost:1935/live/${liveInfo.value.stream_key}`
})

const hlsUrl = computed(() => {
  if (!liveInfo.value) return ''
  return `http://localhost:8081/${liveInfo.value.stream_key}/index.m3u8`
})

// 載入直播資訊
const loadLiveInfo = async () => {
  if (!liveId.value) {
    error.value = '無效的直播 ID'
    loading.value = false
    return
  }

  loading.value = true
  error.value = ''

  try {
    const response = await getLive(liveId.value)
    liveInfo.value = response
  } catch (err: any) {
    console.error('載入直播資訊失敗:', err)
    error.value = err.message || '載入直播資訊失敗'
  } finally {
    loading.value = false
  }
}

// 複製功能
const copyToClipboard = async (text: string, label: string) => {
  try {
    await navigator.clipboard.writeText(text)
    ElMessage.success(`${label} 已複製到剪貼簿`)
  } catch (err) {
    console.error('複製失敗:', err)
    ElMessage.error('複製失敗')
  }
}

const copyStreamKey = () => {
  if (liveInfo.value?.stream_key) {
    copyToClipboard(liveInfo.value.stream_key, '串流金鑰')
  }
}

const copyRtmpUrl = () => {
  copyToClipboard(rtmpUrl.value, 'RTMP 推流地址')
}

const copyHlsUrl = () => {
  copyToClipboard(hlsUrl.value, 'HLS 播放地址')
}

// 工具函數
const getStatusType = (status: string) => {
  switch (status) {
    case 'live': return 'success'
    case 'scheduled': return 'warning'
    case 'ended': return 'info'
    default: return 'info'
  }
}

const getStatusText = (status: string) => {
  switch (status) {
    case 'live': return '直播中'
    case 'scheduled': return '已排程'
    case 'ended': return '已結束'
    default: return status
  }
}

const formatStartTime = (dateString: string) => {
  const date = new Date(dateString)
  return date.toLocaleString('zh-TW')
}

// 生命週期
onMounted(() => {
  loadLiveInfo()
})
</script>

<style scoped>
.live-room {
  max-width: 1400px;
  margin: 0 auto;
  padding: 20px;
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

.header-actions {
  display: flex;
  gap: 12px;
}

.loading-container,
.error-container {
  padding: 40px;
}

.live-content {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.live-layout {
  display: grid;
  grid-template-columns: 1fr 350px;
  gap: 24px;
  min-height: 600px;
}

.player-section {
  min-height: 500px;
}

.chat-section {
  height: 600px;
}

.live-details {
  margin-top: 24px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.live-status {
  display: flex;
  align-items: center;
  gap: 12px;
}

.viewer-count {
  color: #666;
  font-size: 14px;
}

.details-content h2 {
  margin: 0 0 12px 0;
  color: #333;
  font-size: 20px;
}

.description {
  margin: 0 0 16px 0;
  color: #666;
  line-height: 1.5;
}

.meta-info {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.meta-item {
  display: flex;
  align-items: center;
}

.meta-item .label {
  font-weight: bold;
  color: #666;
  width: 80px;
  flex-shrink: 0;
}

.meta-item .value {
  color: #333;
}

.stream-info {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.info-item {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.info-item label {
  font-weight: bold;
  color: #333;
}

.key-display {
  display: flex;
  gap: 8px;
  align-items: center;
}

.key-display .el-input {
  flex: 1;
}

/* 響應式設計 */
@media (max-width: 1200px) {
  .live-layout {
    grid-template-columns: 1fr 300px;
    gap: 16px;
  }
}

@media (max-width: 768px) {
  .live-room {
    padding: 12px;
  }
  
  .live-layout {
    grid-template-columns: 1fr;
    gap: 16px;
  }
  
  .chat-section {
    height: 400px;
  }
  
  .page-header {
    flex-direction: column;
    gap: 12px;
    align-items: stretch;
  }
  
  .header-actions {
    justify-content: center;
  }
}
</style>
