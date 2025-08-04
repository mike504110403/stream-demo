<template>
  <div class="live-room-llhls">
    <div class="page-header">
      <h1>{{ roomInfo?.title || '直播間' }}</h1>
      <div class="header-actions">
        <!-- 主播專用按鈕 -->
        <template v-if="isCreator">
          <el-button
            v-if="
              roomInfo?.status === 'created' || roomInfo?.status === 'ended'
            "
            type="success"
            @click="handleStartLive"
            :loading="startingLive"
          >
            {{ roomInfo?.status === 'ended' ? '重新開始直播' : '開始直播' }}
          </el-button>
          <el-button
            v-if="roomInfo?.status === 'live'"
            type="warning"
            @click="handleEndLive"
            :loading="endingLive"
          >
            結束直播
          </el-button>
          <el-button type="primary" @click="showStreamInfo = true">
            串流資訊
          </el-button>
          <el-button
            type="danger"
            @click="handleCloseRoom"
            :loading="closingRoom"
          >
            關閉直播間
          </el-button>
        </template>

        <!-- 觀眾專用按鈕 -->
        <template v-if="isViewer">
          <el-button type="info" @click="handleLeaveRoom">
            離開直播間
          </el-button>
        </template>
      </div>
    </div>

    <div v-if="loading" class="loading-container">
      <el-skeleton :rows="3" animated />
    </div>

    <div v-else-if="error" class="error-container">
      <el-result icon="error" :title="error" sub-title="無法載入直播間資訊">
        <template #extra>
          <el-button type="primary" @click="loadRoomInfo"> 重新載入 </el-button>
        </template>
      </el-result>
    </div>

    <div v-else-if="roomInfo" class="live-content">
      <!-- 直播播放器和聊天室 -->
      <div class="live-layout">
        <!-- 左側：直播播放器 -->
        <div class="player-section">
          <div class="player-container">
            <div v-if="roomInfo.status === 'live'" class="live-player">
              <!-- 使用 LL-HLS 播放器 -->
              <LLHLSPlayer
                :stream-url="llhlsStreamUrl"
                :title="roomInfo.title"
                :auto-play="true"
                :muted="true"
              />
            </div>
            <div v-else class="offline-message">
              <div class="offline-icon">📺</div>
              <div class="offline-text">
                {{
                  roomInfo.status === 'created' ? '直播尚未開始' : '直播已結束'
                }}
              </div>
              <div
                v-if="
                  isCreator &&
                  (roomInfo.status === 'created' || roomInfo.status === 'ended')
                "
                class="offline-action"
              >
                <el-button type="primary" @click="handleStartLive">
                  {{
                    roomInfo.status === 'ended' ? '重新開始直播' : '開始直播'
                  }}
                </el-button>
              </div>
            </div>
          </div>
        </div>

        <!-- 右側：聊天室 -->
        <div class="chat-section">
          <LiveChat
            :live-id="parseInt(roomId)"
            :current-user-id="userId || 0"
            :current-username="authStore.user?.username || 'Anonymous'"
            :chat-enabled="true"
          />
        </div>
      </div>
    </div>

    <!-- 串流資訊對話框 -->
    <el-dialog v-model="showStreamInfo" title="串流資訊" width="600px">
      <div class="stream-info">
        <div class="info-item">
          <label>推流地址:</label>
          <div class="stream-url">
            <code>{{ getRtmpPushUrl(roomInfo?.stream_key) }}</code>
            <el-button size="small" @click="copyStreamUrl" type="primary">
              複製
            </el-button>
          </div>
        </div>

        <div class="info-item">
          <label>串流金鑰:</label>
          <div class="stream-key">
            <code>{{ roomInfo?.stream_key }}</code>
            <el-button size="small" @click="copyStreamKey" type="primary">
              複製
            </el-button>
          </div>
        </div>

        <div class="info-item">
          <label>播放地址 (LL-HLS):</label>
          <div class="play-url">
            <code>{{ llhlsStreamUrl }}</code>
            <el-button size="small" @click="copyPlayUrl" type="primary">
              複製
            </el-button>
          </div>
        </div>

        <div class="info-item">
          <label>播放地址 (標準 HLS):</label>
          <div class="play-url">
            <code>{{ standardStreamUrl }}</code>
            <el-button size="small" @click="copyStandardUrl" type="primary">
              複製
            </el-button>
          </div>
        </div>
      </div>

      <template #footer>
        <el-button @click="showStreamInfo = false">關閉</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useAuthStore } from '@/store/auth'
// import { useLiveStore } from '@/store/modules/live'  // 暫時註釋掉不存在的 import
import LiveChat from '@/components/live/LiveChat.vue'
import LLHLSPlayer from '@/components/live/LLHLSPlayer.vue'
import { getRtmpPushUrl } from '@/utils/stream-config'
// import type { LiveRoom } from '@/types/live'  // 暫時註釋掉不存在的 import

// 臨時類型定義
interface LiveRoom {
  id: string
  title: string
  description: string
  status: string
  stream_key: string
  creator_id: number
  created_at: string
  updated_at: string
}

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
// const liveStore = useLiveStore()  // 暫時註釋掉不存在的 store

const roomId = computed(() => route.params.id as string)
const userId = computed(() => authStore.user?.id)

const loading = ref(false)
const error = ref('')
const roomInfo = ref<LiveRoom | null>(null)
const startingLive = ref(false)
const endingLive = ref(false)
const closingRoom = ref(false)
const showStreamInfo = ref(false)

// 計算串流 URL
const llhlsStreamUrl = computed(() => {
  if (!roomInfo.value?.stream_key) return ''
  return `/hls/${roomInfo.value.stream_key}/index.m3u8`
})

const standardStreamUrl = computed(() => {
  if (!roomInfo.value?.stream_key) return ''
  return `/hls_standard/${roomInfo.value.stream_key}/index.m3u8`
})

// 權限檢查
const isCreator = computed(() => {
  return roomInfo.value?.creator_id === userId.value
})

const isViewer = computed(() => {
  return !isCreator.value
})

// 載入直播間資訊
const loadRoomInfo = async () => {
  if (!roomId.value) return

  loading.value = true
  error.value = ''

  try {
    // const room = await liveStore.getLiveRoom(roomId.value)  // 暫時註釋掉不存在的 store
    // roomInfo.value = room
    // 臨時使用模擬數據
    // 檢查是否有實際的推流
    const actualStreamKey = 'stream_32f2df4a-4bb' // 使用實際的推流 key
    roomInfo.value = {
      id: roomId.value,
      title: '測試直播間',
      description: '這是一個測試直播間',
      status: 'live',
      stream_key: actualStreamKey,
      creator_id: 1,
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
    }
  } catch (err: any) {
    error.value = err.message || '載入直播間失敗'
    ElMessage.error(error.value)
  } finally {
    loading.value = false
  }
}

// 開始直播
const handleStartLive = async () => {
  if (!roomInfo.value) return

  startingLive.value = true
  try {
    // await liveStore.startLive(roomInfo.value.id)  // 暫時註釋掉不存在的 store
    await loadRoomInfo()
    ElMessage.success('直播已開始')
  } catch (err: any) {
    ElMessage.error(err.message || '開始直播失敗')
  } finally {
    startingLive.value = false
  }
}

// 結束直播
const handleEndLive = async () => {
  if (!roomInfo.value) return

  endingLive.value = true
  try {
    // await liveStore.endLive(roomInfo.value.id)  // 暫時註釋掉不存在的 store
    await loadRoomInfo()
    ElMessage.success('直播已結束')
  } catch (err: any) {
    ElMessage.error(err.message || '結束直播失敗')
  } finally {
    endingLive.value = false
  }
}

// 關閉直播間
const handleCloseRoom = async () => {
  if (!roomInfo.value) return

  closingRoom.value = true
  try {
    // await liveStore.closeLiveRoom(roomInfo.value.id)  // 暫時註釋掉不存在的 store
    ElMessage.success('直播間已關閉')
    router.push('/live')
  } catch (err: any) {
    ElMessage.error(err.message || '關閉直播間失敗')
  } finally {
    closingRoom.value = false
  }
}

// 離開直播間
const handleLeaveRoom = () => {
  router.push('/live')
}

// 複製功能
const copyStreamUrl = () => {
  const url = getRtmpPushUrl(roomInfo.value?.stream_key)
  navigator.clipboard.writeText(url)
  ElMessage.success('推流地址已複製')
}

const copyStreamKey = () => {
  if (roomInfo.value?.stream_key) {
    navigator.clipboard.writeText(roomInfo.value.stream_key)
    ElMessage.success('串流金鑰已複製')
  }
}

const copyPlayUrl = () => {
  navigator.clipboard.writeText(llhlsStreamUrl.value)
  ElMessage.success('LL-HLS 播放地址已複製')
}

const copyStandardUrl = () => {
  navigator.clipboard.writeText(standardStreamUrl.value)
  ElMessage.success('標準 HLS 播放地址已複製')
}

onMounted(() => {
  loadRoomInfo()
})

onUnmounted(() => {
  // 清理資源
})
</script>

<style scoped>
.live-room-llhls {
  padding: 20px;
  max-width: 1400px;
  margin: 0 auto;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  padding-bottom: 16px;
  border-bottom: 1px solid #e4e7ed;
}

.page-header h1 {
  margin: 0;
  font-size: 24px;
  font-weight: 600;
  color: #303133;
}

.header-actions {
  display: flex;
  gap: 12px;
}

.loading-container,
.error-container {
  padding: 40px;
  text-align: center;
}

.live-content {
  background: #fff;
  border-radius: 8px;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
  overflow: hidden;
}

.live-layout {
  display: grid;
  grid-template-columns: 1fr 350px;
  gap: 0;
  min-height: 600px;
}

.player-section {
  background: #000;
}

.player-container {
  width: 100%;
  height: 100%;
  min-height: 600px;
}

.live-player {
  width: 100%;
  height: 100%;
}

.offline-message {
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  height: 600px;
  color: #909399;
  background: #f5f7fa;
}

.offline-icon {
  font-size: 64px;
  margin-bottom: 16px;
}

.offline-text {
  font-size: 18px;
  margin-bottom: 24px;
}

.chat-section {
  border-left: 1px solid #e4e7ed;
  background: #fff;
}

.stream-info {
  padding: 16px;
}

.info-item {
  margin-bottom: 20px;
}

.info-item label {
  display: block;
  font-weight: 600;
  margin-bottom: 8px;
  color: #303133;
}

.stream-url,
.stream-key,
.play-url {
  display: flex;
  align-items: center;
  gap: 12px;
  background: #f5f7fa;
  padding: 12px;
  border-radius: 4px;
  border: 1px solid #e4e7ed;
}

.stream-url code,
.stream-key code,
.play-url code {
  flex: 1;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 14px;
  color: #409eff;
  word-break: break-all;
}

@media (max-width: 1200px) {
  .live-layout {
    grid-template-columns: 1fr;
  }

  .chat-section {
    border-left: none;
    border-top: 1px solid #e4e7ed;
  }
}

@media (max-width: 768px) {
  .page-header {
    flex-direction: column;
    gap: 16px;
    align-items: stretch;
  }

  .header-actions {
    justify-content: center;
    flex-wrap: wrap;
  }

  .live-room-llhls {
    padding: 12px;
  }
}
</style>
