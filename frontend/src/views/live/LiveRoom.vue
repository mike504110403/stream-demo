<template>
  <div class="live-room">
    <div class="page-header">
      <h1>{{ roomInfo?.title || '直播間' }}</h1>
      <div class="header-actions">
        
        <!-- 主播專用按鈕 -->
        <template v-if="isCreator">
          <el-button 
            v-if="roomInfo?.status === 'created' || roomInfo?.status === 'ended'" 
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
          <el-button 
            type="primary" 
            @click="showStreamInfo = true"
          >
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
          <el-button 
            type="info" 
            @click="handleLeaveRoom"
          >
            離開直播間
          </el-button>
        </template>
      </div>
    </div>

    <div v-if="loading" class="loading-container">
      <el-skeleton :rows="3" animated />
    </div>

    <div v-else-if="error" class="error-container">
      <el-result
        icon="error"
        :title="error"
        sub-title="無法載入直播間資訊"
      >
        <template #extra>
          <el-button type="primary" @click="loadRoomInfo">
            重新載入
          </el-button>
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
              <video 
                ref="videoPlayer"
                controls 
                autoplay 
                muted
                class="video-player"
              >
                <source :src="streamUrl" type="application/x-mpegURL">
                您的瀏覽器不支援影片播放
              </video>
            </div>
            <div v-else class="offline-message">
              <div class="offline-icon">📺</div>
              <div class="offline-text">
                {{ roomInfo.status === 'created' ? '直播尚未開始' : '直播已結束' }}
              </div>
              <div v-if="isCreator && (roomInfo.status === 'created' || roomInfo.status === 'ended')" class="offline-action">
                <el-button type="primary" @click="handleStartLive">
                  {{ roomInfo.status === 'ended' ? '重新開始直播' : '開始直播' }}
                </el-button>
              </div>
            </div>
          </div>
        </div>

        <!-- 右側：聊天室 -->
        <div class="chat-section">
          <div class="chat-container">
            <div class="chat-header">
              <h3>聊天室</h3>
              <div class="chat-status">
                <span class="viewer-count">{{ roomInfo.viewer_count }} 觀眾</span>
                <el-tag 
                  :type="isConnected ? 'success' : 'danger'" 
                  size="small"
                >
                  {{ isConnected ? '已連接' : '未連接' }}
                </el-tag>
              </div>
            </div>
            <div class="chat-messages" ref="chatMessages">
              <div class="message-list">
                <div v-for="message in messages" :key="message.id" class="message">
                  <span class="username" :class="{ 'creator': message.role === 'creator' }">
                    {{ message.username }}
                    <el-tag v-if="message.role === 'creator'" size="small" type="warning">主播</el-tag>:
                  </span>
                  <span class="content">{{ message.content }}</span>
                  <span class="timestamp">{{ formatTime(message.timestamp) }}</span>
                </div>
              </div>
            </div>
            <div class="chat-input">
              <el-input 
                v-model="newMessage" 
                placeholder="輸入訊息..."
                @keyup.enter="sendMessage"
                :disabled="!isConnected"
              >
                <template #append>
                  <el-button @click="sendMessage" :disabled="!newMessage.trim() || !isConnected">
                    發送
                  </el-button>
                </template>
              </el-input>
            </div>
          </div>
        </div>
      </div>

      <!-- 直播詳情 -->
      <div class="live-details">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>直播間詳情</span>
              <div class="live-status">
                <el-tag :type="getStatusType(roomInfo.status)">
                  {{ getStatusText(roomInfo.status) }}
                </el-tag>
                <span class="viewer-count">
                  {{ roomInfo.viewer_count }}/{{ roomInfo.max_viewers }} 觀眾
                </span>
              </div>
            </div>
          </template>
          
          <div class="details-content">
            <h2>{{ roomInfo.title }}</h2>
            <p v-if="roomInfo.description" class="description">
              {{ roomInfo.description }}
            </p>
            
            <div class="meta-info">
              <div class="meta-item">
                <span class="label">創建者：</span>
                <span class="value">用戶 ID: {{ roomInfo.creator_id }}</span>
              </div>
              <div class="meta-item">
                <span class="label">創建時間：</span>
                <span class="value">{{ formatDate(roomInfo.created_at) }}</span>
              </div>
              <div class="meta-item" v-if="roomInfo.started_at">
                <span class="label">開始時間：</span>
                <span class="value">{{ formatDate(roomInfo.started_at) }}</span>
              </div>
              
              <!-- 主播專用資訊 -->
              <template v-if="isCreator">
                <div class="meta-item">
                  <span class="label">串流金鑰：</span>
                  <span class="value">{{ roomInfo.stream_key }}</span>
                </div>
              </template>
            </div>
          </div>
        </el-card>
      </div>
    </div>

    <!-- 串流資訊對話框 -->
    <el-dialog 
      v-model="showStreamInfo" 
      title="串流資訊" 
      width="600px"
    >
      <div class="stream-info">
        <div class="info-item">
          <label>串流金鑰：</label>
          <div class="info-content">
            <el-input :value="roomInfo?.stream_key" readonly />
            <el-button @click="copyStreamKey" size="small">複製</el-button>
          </div>
        </div>
        
        <div class="info-item">
          <label>RTMP 推流地址：</label>
          <div class="info-content">
            <el-input :value="rtmpUrl" readonly />
            <el-button @click="copyRtmpUrl" size="small">複製</el-button>
          </div>
        </div>
        
        <div class="info-item">
          <label>HLS 播放地址：</label>
          <div class="info-content">
            <el-input :value="hlsUrl" readonly />
            <el-button @click="copyHlsUrl" size="small">複製</el-button>
          </div>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, nextTick, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getRoomById, joinRoom, leaveRoom, startLive as startLiveAPI, endLive as endLiveAPI, closeRoom, getUserRole as getUserRoleAPI } from '@/api/live-room'
import { useAuthStore } from '@/store/auth'
import type { LiveRoomInfo } from '@/types'
import { LiveRoomWebSocket, type LiveRoomMessage } from '@/utils/websocket'
import Hls from 'hls.js'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

// 響應式數據
const loading = ref(true)
const error = ref('')
const roomInfo = ref<LiveRoomInfo | null>(null)
const showStreamInfo = ref(false)
const startingLive = ref(false)
const endingLive = ref(false)
const closingRoom = ref(false)
const userRole = ref<string>('') // 添加用戶角色狀態

// 聊天相關
const messages = ref<Array<{id: string, username: string, content: string, role?: string, timestamp: number}>>([])
const newMessage = ref('')
const chatMessages = ref<HTMLElement>()

// WebSocket 相關
const wsClient = ref<LiveRoomWebSocket | null>(null)
const isConnected = ref(false)

// HLS 播放器相關
const videoPlayer = ref<HTMLVideoElement>()
const hls = ref<Hls | null>(null)

// 計算屬性
const roomId = computed(() => route.params.id as string)
const currentUserId = computed(() => authStore.user?.id || 0)
const currentUsername = computed(() => authStore.user?.username || '')

// 用戶角色相關
const isCreator = computed(() => {
  // 優先檢查用戶角色，然後檢查創建者ID
  const result = userRole.value === 'creator' || roomInfo.value?.creator_id === currentUserId.value
  console.log('角色判斷:', {
    userRole: userRole.value,
    creator_id: roomInfo.value?.creator_id,
    currentUserId: currentUserId.value,
    isCreator: result
  })
  return result
})
const isViewer = computed(() => !isCreator.value)

// 串流 URL
const streamUrl = computed(() => {
  if (!roomInfo.value || roomInfo.value.status !== 'live') return ''
  return `http://localhost:8083/${roomInfo.value.stream_key}/index.m3u8`
})

const rtmpUrl = computed(() => {
  if (!roomInfo.value) return ''
  return `rtmp://localhost:1935/live/${roomInfo.value.stream_key}`
})

const hlsUrl = computed(() => {
  if (!roomInfo.value) return ''
  return `http://localhost:8083/${roomInfo.value.stream_key}/index.m3u8`
})

// 初始化 HLS 播放器
const initHLSPlayer = async () => {
  if (!videoPlayer.value || !streamUrl.value) return
  
  console.log('初始化 HLS 播放器:', streamUrl.value)
  
  // 清理現有的 HLS 實例
  if (hls.value) {
    hls.value.destroy()
    hls.value = null
  }
  
  // 檢查瀏覽器是否支援 HLS
  if (Hls.isSupported()) {
    hls.value = new Hls({
      debug: false,
      enableWorker: true,
      lowLatencyMode: true,
      // 低延遲直播優化
      maxBufferLength: 4,           // 最大緩衝 4 秒
      maxMaxBufferLength: 8,        // 絕對最大緩衝 8 秒
      maxBufferSize: 8 * 1000 * 1000, // 8MB 緩衝
      maxBufferHole: 0.1,           // 允許的緩衝空洞
      highBufferWatchdogPeriod: 1,  // 高緩衝監控週期
      nudgeOffset: 0.1,             // 調整偏移
      nudgeMaxRetry: 3,             // 最大重試次數
      maxFragLookUpTolerance: 0.1,  // 片段查找容差
      liveSyncDurationCount: 1,     // 直播同步片段數
      liveMaxLatencyDurationCount: 3, // 最大延遲片段數
      liveDurationInfinity: true,   // 無限直播
      enableSoftwareAES: true,      // 啟用軟體 AES
      abrEwmaFastLive: 3,           // 快速 ABR
      abrEwmaSlowLive: 9,           // 慢速 ABR
    })
    
    hls.value.loadSource(streamUrl.value)
    hls.value.attachMedia(videoPlayer.value)
    
    hls.value.on(Hls.Events.MANIFEST_PARSED, () => {
      console.log('HLS 播放列表已解析，開始播放')
      if (videoPlayer.value) {
        videoPlayer.value.play().catch(err => {
          console.error('自動播放失敗:', err)
        })
      }
    })
    
    hls.value.on(Hls.Events.ERROR, (event, data) => {
      console.error('HLS 錯誤:', data)
      if (data.fatal) {
        switch (data.type) {
          case Hls.ErrorTypes.NETWORK_ERROR:
            console.log('網絡錯誤，嘗試恢復...')
            hls.value?.startLoad()
            break
          case Hls.ErrorTypes.MEDIA_ERROR:
            console.log('媒體錯誤，嘗試恢復...')
            hls.value?.recoverMediaError()
            break
          default:
            console.error('致命錯誤，無法恢復')
            break
        }
      }
    })
  } else if (videoPlayer.value.canPlayType('application/vnd.apple.mpegurl')) {
    // Safari 原生支援 HLS
    console.log('使用 Safari 原生 HLS 播放')
    videoPlayer.value.src = streamUrl.value
    videoPlayer.value.addEventListener('loadedmetadata', () => {
      videoPlayer.value?.play().catch(err => {
        console.error('Safari 自動播放失敗:', err)
      })
    })
  } else {
    console.error('瀏覽器不支援 HLS 播放')
  }
}

// 清理 HLS 播放器
const cleanupHLSPlayer = () => {
  if (hls.value) {
    hls.value.destroy()
    hls.value = null
  }
  if (videoPlayer.value) {
    videoPlayer.value.src = ''
  }
}

// 載入直播間資訊
const loadRoomInfo = async () => {
  if (!roomId.value) {
    error.value = '無效的直播間 ID'
    loading.value = false
    return
  }

  loading.value = true
  error.value = ''

  try {
    const response = await getRoomById(roomId.value)
    roomInfo.value = response
    
    // 調試：檢查認證狀態
    console.log('載入房間信息時的認證狀態:', {
      token: !!authStore.token,
      user: authStore.user,
      currentUserId: currentUserId.value,
      roomCreatorId: roomInfo.value?.creator_id
    })
    
    // 加入直播間
    await joinRoom(roomId.value)
    
    // 獲取用戶在房間中的角色
    await getUserRole()
    
    // 初始化空的聊天消息列表
    messages.value = []
  } catch (err: any) {
    console.error('載入直播間資訊失敗:', err)
    error.value = err.message || '載入直播間資訊失敗'
  } finally {
    loading.value = false
  }
}

// 獲取用戶在房間中的角色
const getUserRole = async () => {
  try {
    const response = await getUserRoleAPI(roomId.value)
    userRole.value = response.role
    console.log('用戶角色設置:', userRole.value)
  } catch (err: any) {
    console.error('獲取用戶角色失敗:', err)
    // 如果 API 失敗，使用創建者ID來判斷
    if (roomInfo.value?.creator_id === currentUserId.value) {
      userRole.value = 'creator'
    } else {
      userRole.value = 'viewer'
    }
    console.log('使用備用角色判斷:', userRole.value)
  }
}

// 開始直播
const handleStartLive = async () => {
  if (!roomId.value) return
  
  startingLive.value = true
  try {
    await startLiveAPI(roomId.value)
    // 狀態會通過 WebSocket 實時更新，不需要重新載入
  } catch (err: any) {
    console.error('開始直播失敗:', err)
    ElMessage.error(err.message || '開始直播失敗')
  } finally {
    startingLive.value = false
  }
}

// 結束直播
const handleEndLive = async () => {
  if (!roomId.value) return
  
  endingLive.value = true
  try {
    await endLiveAPI(roomId.value)
    // 狀態會通過 WebSocket 實時更新，不需要重新載入
  } catch (err: any) {
    console.error('結束直播失敗:', err)
    ElMessage.error(err.message || '結束直播失敗')
  } finally {
    endingLive.value = false
  }
}

// 關閉直播間
const handleCloseRoom = async () => {
  if (!roomId.value) return
  
  // 確認對話框
  try {
    await ElMessageBox.confirm(
      '確定要關閉這個直播間嗎？關閉後將無法恢復。',
      '確認關閉',
      {
        confirmButtonText: '確定關閉',
        cancelButtonText: '取消',
        type: 'warning',
      }
    )
  } catch {
    return // 用戶取消
  }
  
  closingRoom.value = true
  try {
    await closeRoom(roomId.value)
    // 跳轉會通過 WebSocket 的 room_closed 消息處理
  } catch (err: any) {
    console.error('關閉直播間失敗:', err)
    ElMessage.error(err.message || '關閉直播間失敗')
  } finally {
    closingRoom.value = false
  }
}

// 發送消息
const sendMessage = () => {
  if (!newMessage.value.trim() || !roomInfo.value) return
  
  // 通過 WebSocket 發送聊天消息
  if (wsClient.value && isConnected.value) {
    wsClient.value.sendChatMessage(newMessage.value)
    newMessage.value = ''
  } else {
    // 如果 WebSocket 未連接，使用本地消息（僅用於測試）
    const message = {
      id: Date.now().toString(),
      username: currentUsername.value,
      content: newMessage.value,
      timestamp: Date.now()
    }
    
    messages.value.push(message)
    newMessage.value = ''
    
    // 滾動到底部
    setTimeout(() => {
      if (chatMessages.value) {
        chatMessages.value.scrollTop = chatMessages.value.scrollHeight
      }
    }, 100)
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
  if (roomInfo.value?.stream_key) {
    copyToClipboard(roomInfo.value.stream_key, '串流金鑰')
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
    case 'created': return 'info'
    case 'ended': return 'danger'
    case 'cancelled': return 'warning'
    default: return 'info'
  }
}

const getStatusText = (status: string) => {
  switch (status) {
    case 'live': return '直播中'
    case 'created': return '已創建'
    case 'ended': return '已結束'
    case 'cancelled': return '已取消'
    default: return status
  }
}

const formatDate = (dateString: string) => {
  if (!dateString) return '未知'
  return new Date(dateString).toLocaleString('zh-TW')
}

const formatTime = (timestamp: number) => {
  return new Date(timestamp).toLocaleTimeString('zh-TW', { 
    hour: '2-digit', 
    minute: '2-digit' 
  })
}

// WebSocket 連接
const connectWebSocket = async () => {
  if (!roomId.value || !authStore.token) return
  
  try {
    wsClient.value = new LiveRoomWebSocket(roomId.value, authStore.token)
    
    // 註冊消息處理器
    wsClient.value.on('chat', (message: LiveRoomMessage) => {
      const chatMessage = {
        id: message.timestamp.toString(),
        username: message.username || `user_${message.user_id}`,
        content: message.content || '',
        role: message.role,
        timestamp: message.timestamp
      }
      messages.value.push(chatMessage)
      
      // 滾動到底部
      nextTick(() => {
        if (chatMessages.value) {
          chatMessages.value.scrollTop = chatMessages.value.scrollHeight
        }
      })
    })
    
    wsClient.value.on('user_joined', (message: LiveRoomMessage) => {
      if (message.data?.viewer_count !== undefined && roomInfo.value) {
        roomInfo.value.viewer_count = message.data.viewer_count
        console.log('觀眾數量更新 (加入):', message.data.viewer_count)
      }
      // 只有主播能看到加入消息
      if (isCreator.value && message.username) {
        ElMessage.info(`${message.username} 加入了直播間`)
      }
    })
    
    wsClient.value.on('user_left', (message: LiveRoomMessage) => {
      if (message.data?.viewer_count !== undefined && roomInfo.value) {
        roomInfo.value.viewer_count = message.data.viewer_count
        console.log('觀眾數量更新 (離開):', message.data.viewer_count)
      }
      // 只有主播能看到離開消息
      if (isCreator.value && message.username) {
        ElMessage.info(`${message.username} 離開了直播間`)
      }
    })

    // 定期更新觀眾數量（備用方案）
    wsClient.value.on('viewer_count_update', (message: LiveRoomMessage) => {
      if (message.data?.viewer_count !== undefined && roomInfo.value) {
        roomInfo.value.viewer_count = message.data.viewer_count
        console.log('觀眾數量定期更新:', message.data.viewer_count)
      }
    })

    // 處理直播開始通知
    wsClient.value.on('live_started', (message: LiveRoomMessage) => {
      if (roomInfo.value) {
        roomInfo.value.status = 'live'
        console.log('直播狀態更新: 已開始')
        // 初始化 HLS 播放器
        nextTick(() => {
          initHLSPlayer()
        })
      }
    })

    // 處理直播結束通知
    wsClient.value.on('live_ended', (message: LiveRoomMessage) => {
      if (roomInfo.value) {
        roomInfo.value.status = 'ended'
        console.log('直播狀態更新: 已結束')
      }
    })

    // 處理直播間關閉通知
    wsClient.value.on('room_closed', (message: LiveRoomMessage) => {
      ElMessage.warning('直播間已關閉')
      router.push('/live-rooms')
    })
    
    // 連接 WebSocket
    await wsClient.value.connect()
    isConnected.value = true
    console.log('WebSocket 連接成功')
    
  } catch (error) {
    console.error('WebSocket 連接失敗:', error)
    ElMessage.warning('WebSocket 連接失敗，聊天功能可能無法正常使用')
  }
}

// 斷開 WebSocket 連接
const disconnectWebSocket = () => {
  if (wsClient.value) {
    wsClient.value.disconnect()
    wsClient.value = null
    isConnected.value = false
  }
}

// 離開直播間
const handleLeaveRoom = async () => {
  if (roomId.value) {
    try {
      await leaveRoom(roomId.value)
      ElMessage.success('已離開直播間')
      router.push('/live-rooms')
    } catch (err) {
      console.error('離開直播間失敗:', err)
      ElMessage.error('離開直播間失敗')
    }
  }
}

// 監聽 streamUrl 變化
watch(streamUrl, (newUrl) => {
  if (newUrl && roomInfo.value?.status === 'live') {
    console.log('串流 URL 變化，重新初始化播放器:', newUrl)
    nextTick(() => {
      initHLSPlayer()
    })
  }
})

onMounted(async () => {
  await loadRoomInfo()
  await connectWebSocket()
})

onUnmounted(() => {
  cleanupHLSPlayer()
  disconnectWebSocket()
  handleLeaveRoom()
})
</script>

<style scoped>
.live-room {
  padding: 20px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.page-header h1 {
  margin: 0;
  color: #333;
}

.header-actions {
  display: flex;
  gap: 10px;
}

.loading-container,
.error-container {
  padding: 40px;
  text-align: center;
}

.live-content {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.live-layout {
  display: grid;
  grid-template-columns: 2fr 1fr;
  gap: 20px;
  min-height: 500px;
}

.player-section {
  background: #000;
  border-radius: 8px;
  overflow: hidden;
}

.player-container {
  width: 100%;
  height: 100%;
  min-height: 400px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.live-player {
  width: 100%;
  height: 100%;
}

.video-player {
  width: 100%;
  height: 100%;
  object-fit: contain;
}

.offline-message {
  text-align: center;
  color: #fff;
}

.offline-icon {
  font-size: 64px;
  margin-bottom: 20px;
}

.offline-text {
  font-size: 18px;
  margin-bottom: 20px;
}

.chat-section {
  background: #f5f5f5;
  border-radius: 8px;
  display: flex;
  flex-direction: column;
}

.chat-container {
  display: flex;
  flex-direction: column;
  height: 100%;
}

.chat-header {
  padding: 15px;
  border-bottom: 1px solid #e0e0e0;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.chat-header h3 {
  margin: 0;
  color: #333;
}

.chat-status {
  display: flex;
  align-items: center;
  gap: 10px;
}

.viewer-count {
  color: #666;
  font-size: 14px;
}

.chat-messages {
  flex: 1;
  padding: 15px;
  overflow-y: auto;
  max-height: 300px;
}

.message-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.message {
  padding: 8px 12px;
  background: #fff;
  border-radius: 8px;
  font-size: 14px;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.username {
  font-weight: bold;
  color: #333;
  display: flex;
  align-items: center;
  gap: 8px;
}

.username.creator {
  color: #e6a23c;
}

.content {
  color: #666;
  margin-left: 0;
}

.timestamp {
  color: #999;
  font-size: 12px;
  align-self: flex-end;
}

.chat-input {
  padding: 15px;
  border-top: 1px solid #e0e0e0;
}

.live-details {
  margin-top: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.live-status {
  display: flex;
  align-items: center;
  gap: 10px;
}

.details-content h2 {
  margin: 0 0 10px 0;
  color: #333;
}

.description {
  color: #666;
  margin: 0 0 20px 0;
  line-height: 1.5;
}

.meta-info {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 15px;
}

.meta-item {
  display: flex;
  flex-direction: column;
  gap: 5px;
}

.label {
  font-weight: bold;
  color: #333;
  font-size: 14px;
}

.value {
  color: #666;
  font-size: 14px;
}

.stream-info {
  display: flex;
  flex-direction: column;
  gap: 20px;
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

.info-content {
  display: flex;
  gap: 10px;
}

.info-content .el-input {
  flex: 1;
}

@media (max-width: 768px) {
  .live-layout {
    grid-template-columns: 1fr;
  }
  
  .meta-info {
    grid-template-columns: 1fr;
  }
}
</style>
