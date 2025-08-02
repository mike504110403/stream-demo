<template>
  <div class="live-chat">
    <!-- 聊天標題 -->
    <div class="chat-header">
      <h3>直播聊天室</h3>
      <div class="chat-status">
        <span class="online-count">{{ onlineUsers }} 人在線</span>
        <el-switch
          :model-value="chatEnabled"
          @update:model-value="toggleChat"
          active-text="聊天開啟"
          inactive-text="聊天關閉"
        />
      </div>
    </div>

    <!-- 聊天訊息區域 -->
    <div class="chat-messages" ref="messagesContainer">
      <div v-if="loading" class="loading-messages">
        <el-icon class="loading-icon" :size="20">
          <Loading />
        </el-icon>
        <span>載入聊天記錄...</span>
      </div>

      <div v-else-if="messages.length === 0" class="empty-messages">
        <div class="empty-icon">💬</div>
        <p>還沒有聊天訊息</p>
        <p>開始第一個對話吧！</p>
      </div>

      <div v-else class="messages-list">
        <div
          v-for="message in messages"
          :key="message.id"
          class="message-item"
          :class="{ 'message-own': message.user_id === currentUserId }"
        >
          <!-- 系統訊息 -->
          <div v-if="message.type === 'system'" class="system-message">
            <el-icon><InfoFilled /></el-icon>
            <span>{{ message.content }}</span>
          </div>

          <!-- 用戶訊息 -->
          <div v-else class="user-message">
            <div class="message-avatar">
              <el-avatar :size="32">
                {{ message.username?.charAt(0)?.toUpperCase() }}
              </el-avatar>
            </div>
            <div class="message-content">
              <div class="message-header">
                <span class="username">{{ message.username }}</span>
                <span class="timestamp">{{ formatTime(message.created_at) }}</span>
              </div>
              <div class="message-text">{{ message.content }}</div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 聊天輸入區域 -->
    <div class="chat-input" v-if="chatEnabled">
      <div class="input-container">
        <el-input
          v-model="messageText"
          placeholder="輸入訊息..."
          :disabled="!connected"
          @keyup.enter="sendMessage"
          @keyup.ctrl.enter="sendMessage"
        >
          <template #append>
            <el-button
              type="primary"
              :disabled="!messageText.trim() || !connected"
              @click="sendMessage"
            >
              發送
            </el-button>
          </template>
        </el-input>
      </div>
      <div class="input-tips">
        <span>按 Enter 發送，Ctrl+Enter 換行</span>
        <span v-if="!connected" class="connection-status">
          <el-icon><Warning /></el-icon>
          連接中...
        </span>
      </div>
    </div>

    <!-- 聊天已關閉提示 -->
    <div v-else class="chat-disabled">
      <el-icon><ChatDotRound /></el-icon>
      <p>聊天功能已關閉</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, nextTick, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { Loading, InfoFilled, Warning, ChatDotRound } from '@element-plus/icons-vue'
import type { ChatMessage } from '@/types'

interface Props {
  liveId: number
  currentUserId: number
  currentUsername: string
  chatEnabled?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  chatEnabled: true
})

// 響應式數據
const messages = ref<ChatMessage[]>([])
const messageText = ref('')
const loading = ref(false)
const connected = ref(false)
const onlineUsers = ref(0)
const messagesContainer = ref<HTMLElement>()

// WebSocket 相關
let ws: WebSocket | null = null
let reconnectTimer: number | null = null
let heartbeatTimer: number | null = null

// 初始化 WebSocket 連接
const initWebSocket = () => {
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const wsUrl = `${protocol}//${window.location.host}/ws/live/${props.liveId}`
  
  console.log('連接 WebSocket:', wsUrl)
  
  ws = new WebSocket(wsUrl)
  
  ws.onopen = () => {
    console.log('WebSocket 連接成功')
    connected.value = true
    startHeartbeat()
  }
  
  ws.onmessage = (event) => {
    try {
      const data = JSON.parse(event.data)
      handleWebSocketMessage(data)
    } catch (error) {
      console.error('解析 WebSocket 訊息失敗:', error)
    }
  }
  
  ws.onclose = () => {
    console.log('WebSocket 連接關閉')
    connected.value = false
    stopHeartbeat()
    scheduleReconnect()
  }
  
  ws.onerror = (error) => {
    console.error('WebSocket 錯誤:', error)
    connected.value = false
  }
}

// 處理 WebSocket 訊息
const handleWebSocketMessage = (data: any) => {
  switch (data.type) {
    case 'chat_message':
      addMessage(data.message)
      break
    case 'system_message':
      addSystemMessage(data.content)
      break
    case 'user_count':
      onlineUsers.value = data.count
      break
    case 'chat_toggle':
      // 聊天開關狀態更新
      break
    case 'pong':
      // 心跳回應
      break
    default:
      console.log('未知訊息類型:', data.type)
  }
}

// 發送訊息
const sendMessage = () => {
  if (!messageText.value.trim() || !connected.value) {
    return
  }
  
  const message = {
    type: 'chat_message',
    live_id: props.liveId,
    user_id: props.currentUserId,
    username: props.currentUsername,
    content: messageText.value.trim()
  }
  
  if (ws && ws.readyState === WebSocket.OPEN) {
    ws.send(JSON.stringify(message))
    messageText.value = ''
  } else {
    ElMessage.error('連接已斷開，請重新載入頁面')
  }
}

// 添加訊息到列表
const addMessage = (message: ChatMessage) => {
  messages.value.push(message)
  scrollToBottom()
}

// 添加系統訊息
const addSystemMessage = (content: string) => {
  const systemMessage: ChatMessage = {
    id: Date.now(),
    type: 'system',
    live_id: props.liveId,
    user_id: 0,
    username: '系統',
    content,
    created_at: new Date().toISOString()
  }
  
  messages.value.push(systemMessage)
  scrollToBottom()
}

// 滾動到底部
const scrollToBottom = () => {
  nextTick(() => {
    if (messagesContainer.value) {
      messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight
    }
  })
}

// 切換聊天功能
const toggleChat = (enabled: string | number | boolean) => {
  // 這裡可以調用 API 來切換聊天狀態
  console.log('切換聊天狀態:', enabled)
}

// 心跳機制
const startHeartbeat = () => {
  heartbeatTimer = window.setInterval(() => {
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify({ type: 'ping' }))
    }
  }, 30000) // 每30秒發送一次心跳
}

const stopHeartbeat = () => {
  if (heartbeatTimer) {
    clearInterval(heartbeatTimer)
    heartbeatTimer = null
  }
}

// 重連機制
const scheduleReconnect = () => {
  if (reconnectTimer) {
    clearTimeout(reconnectTimer)
  }
  
  reconnectTimer = window.setTimeout(() => {
    console.log('嘗試重新連接...')
    initWebSocket()
  }, 3000) // 3秒後重連
}

// 載入歷史訊息
const loadHistoryMessages = async () => {
  loading.value = true
  try {
    // 這裡應該調用 API 來載入歷史訊息
    // const response = await getChatMessages(props.liveId)
    // messages.value = response.data
    
    // 暫時使用模擬數據
    setTimeout(() => {
      messages.value = [
        {
          id: 1,
          type: 'system',
          live_id: props.liveId,
          user_id: 0,
          username: '系統',
          content: '歡迎來到直播間！',
          created_at: new Date(Date.now() - 60000).toISOString()
        }
      ]
      loading.value = false
      scrollToBottom()
    }, 1000)
  } catch (error) {
    console.error('載入聊天記錄失敗:', error)
    loading.value = false
  }
}

// 工具函數
const formatTime = (dateString: string): string => {
  const date = new Date(dateString)
  const now = new Date()
  const diff = now.getTime() - date.getTime()
  
  if (diff < 60000) { // 1分鐘內
    return '剛剛'
  } else if (diff < 3600000) { // 1小時內
    return `${Math.floor(diff / 60000)}分鐘前`
  } else {
    return date.toLocaleTimeString('zh-TW', { 
      hour: '2-digit', 
      minute: '2-digit' 
    })
  }
}

// 監聽聊天開關狀態
watch(() => props.chatEnabled, (enabled) => {
  if (enabled && !connected.value) {
    initWebSocket()
  }
})

// 生命週期
onMounted(() => {
  loadHistoryMessages()
  
  if (props.chatEnabled) {
    initWebSocket()
  }
})

onUnmounted(() => {
  if (ws) {
    ws.close()
    ws = null
  }
  
  if (reconnectTimer) {
    clearTimeout(reconnectTimer)
  }
  
  if (heartbeatTimer) {
    clearInterval(heartbeatTimer)
  }
})
</script>

<style scoped>
.live-chat {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.chat-header {
  padding: 16px;
  border-bottom: 1px solid #f0f0f0;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.chat-header h3 {
  margin: 0;
  color: #333;
  font-size: 16px;
}

.chat-status {
  display: flex;
  align-items: center;
  gap: 12px;
}

.online-count {
  color: #666;
  font-size: 14px;
}

.chat-messages {
  flex: 1;
  overflow-y: auto;
  padding: 16px;
  min-height: 300px;
  max-height: 500px;
}

.loading-messages,
.empty-messages {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 200px;
  color: #999;
}

.loading-icon {
  animation: spin 1s linear infinite;
  margin-bottom: 8px;
}

.empty-icon {
  font-size: 48px;
  margin-bottom: 16px;
}

.empty-messages p {
  margin: 4px 0;
  font-size: 14px;
}

.messages-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.message-item {
  display: flex;
  flex-direction: column;
}

.system-message {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  background: #f5f5f5;
  border-radius: 16px;
  color: #666;
  font-size: 12px;
  align-self: center;
  max-width: 80%;
}

.user-message {
  display: flex;
  gap: 8px;
  max-width: 100%;
}

.message-own {
  flex-direction: row-reverse;
}

.message-avatar {
  flex-shrink: 0;
}

.message-content {
  flex: 1;
  max-width: calc(100% - 40px);
}

.message-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 4px;
}

.username {
  font-weight: bold;
  color: #333;
  font-size: 14px;
}

.timestamp {
  color: #999;
  font-size: 12px;
}

.message-text {
  background: #f0f0f0;
  padding: 8px 12px;
  border-radius: 12px;
  color: #333;
  font-size: 14px;
  line-height: 1.4;
  word-wrap: break-word;
}

.message-own .message-text {
  background: #409eff;
  color: white;
}

.message-own .message-header {
  flex-direction: row-reverse;
}

.chat-input {
  padding: 16px;
  border-top: 1px solid #f0f0f0;
}

.input-container {
  margin-bottom: 8px;
}

.input-tips {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 12px;
  color: #999;
}

.connection-status {
  display: flex;
  align-items: center;
  gap: 4px;
  color: #f56c6c;
}

.chat-disabled {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 200px;
  color: #999;
}

.chat-disabled .el-icon {
  font-size: 48px;
  margin-bottom: 16px;
}

.chat-disabled p {
  margin: 0;
  font-size: 14px;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

/* 響應式設計 */
@media (max-width: 768px) {
  .chat-header {
    padding: 12px;
  }
  
  .chat-messages {
    padding: 12px;
    min-height: 250px;
    max-height: 400px;
  }
  
  .chat-input {
    padding: 12px;
  }
  
  .message-text {
    font-size: 13px;
    padding: 6px 10px;
  }
}
</style>
