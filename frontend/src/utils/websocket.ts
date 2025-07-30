import { ElMessage } from 'element-plus'

export interface LiveRoomMessage {
  type: string
  room_id?: string
  user_id?: number
  username?: string
  role?: string
  content?: string
  data?: any
  timestamp: number
}

export class LiveRoomWebSocket {
  private ws: WebSocket | null = null
  private roomId: string
  private token: string
  private reconnectAttempts = 0
  private maxReconnectAttempts = 5
  private reconnectInterval = 3000
  private heartbeatInterval: number | null = null
  private messageHandlers: Map<string, (message: LiveRoomMessage) => void> = new Map()

  constructor(roomId: string, token: string) {
    this.roomId = roomId
    this.token = token
  }

  // 連接 WebSocket
  connect(): Promise<void> {
    return new Promise((resolve, reject) => {
      try {
        // 使用相對路徑，由 nginx 反向代理處理
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
        const wsUrl = `${protocol}//${window.location.host}/ws/live-room/${this.roomId}?token=${this.token}`
        this.ws = new WebSocket(wsUrl)

        this.ws.onopen = () => {
          console.log('WebSocket 連接成功')
          this.reconnectAttempts = 0
          this.startHeartbeat()
          resolve()
        }

        this.ws.onmessage = (event) => {
          try {
            const message: LiveRoomMessage = JSON.parse(event.data)
            this.handleMessage(message)
          } catch (error) {
            console.error('解析 WebSocket 消息失敗:', error)
          }
        }

        this.ws.onclose = (event) => {
          console.log('WebSocket 連接關閉:', event.code, event.reason)
          this.stopHeartbeat()
          this.handleReconnect()
        }

        this.ws.onerror = (error) => {
          console.error('WebSocket 錯誤:', error)
          reject(error)
        }
      } catch (error) {
        reject(error)
      }
    })
  }

  // 斷開連接
  disconnect(): void {
    if (this.ws) {
      this.ws.close(1000, '用戶主動斷開')
      this.ws = null
    }
    this.stopHeartbeat()
  }

  // 發送消息
  sendMessage(type: string, content?: string, data?: any): void {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
      console.warn('WebSocket 未連接')
      return
    }

    const message: LiveRoomMessage = {
      type,
      room_id: this.roomId,
      content,
      data,
      timestamp: Date.now()
    }

    this.ws.send(JSON.stringify(message))
  }

  // 發送聊天消息
  sendChatMessage(content: string): void {
    this.sendMessage('chat', content)
  }

  // 發送 ping
  sendPing(): void {
    this.sendMessage('ping')
  }

  // 註冊消息處理器
  on(type: string, handler: (message: LiveRoomMessage) => void): void {
    this.messageHandlers.set(type, handler)
  }

  // 移除消息處理器
  off(type: string): void {
    this.messageHandlers.delete(type)
  }

  // 處理接收到的消息
  private handleMessage(message: LiveRoomMessage): void {
    console.log('收到 WebSocket 消息:', message)

    const handler = this.messageHandlers.get(message.type)
    if (handler) {
      handler(message)
    }

    // 特殊處理某些消息類型
    switch (message.type) {
      case 'welcome':
        ElMessage.success('歡迎加入直播間！')
        break
      case 'user_joined':
        if (message.username) {
          ElMessage.info(`${message.username} 加入了直播間`)
        }
        break
      case 'user_left':
        if (message.username) {
          ElMessage.info(`${message.username} 離開了直播間`)
        }
        break
      case 'chat':
        // 聊天消息由具體的處理器處理
        break
      case 'pong':
        // 心跳回應，不需要特殊處理
        break
      case 'room_closed':
        ElMessage.warning('直播間已關閉')
        // 跳轉回直播間列表
        if (typeof window !== 'undefined') {
          window.location.href = '/live-rooms'
        }
        break
      case 'live_started':
        ElMessage.success('直播已開始！')
        break
      case 'live_ended':
        ElMessage.info('直播已結束')
        break
      case 'viewer_count_update':
        // 觀眾數量更新，由具體的處理器處理
        break
      default:
        console.log('未處理的消息類型:', message.type)
    }
  }

  // 處理重連
  private handleReconnect(): void {
    if (this.reconnectAttempts < this.maxReconnectAttempts) {
      this.reconnectAttempts++
      console.log(`嘗試重連 (${this.reconnectAttempts}/${this.maxReconnectAttempts})...`)
      
      setTimeout(() => {
        this.connect().catch(error => {
          console.error('重連失敗:', error)
        })
      }, this.reconnectInterval)
    } else {
      console.error('重連次數已達上限')
      ElMessage.error('WebSocket 連接失敗，請刷新頁面重試')
    }
  }

  // 開始心跳
  private startHeartbeat(): void {
    this.heartbeatInterval = setInterval(() => {
      this.sendPing()
    }, 30000) // 每30秒發送一次 ping
  }

  // 停止心跳
  private stopHeartbeat(): void {
    if (this.heartbeatInterval) {
      clearInterval(this.heartbeatInterval)
      this.heartbeatInterval = null
    }
  }

  // 獲取連接狀態
  getConnectionState(): number {
    return this.ws ? this.ws.readyState : WebSocket.CLOSED
  }

  // 檢查是否已連接
  isConnected(): boolean {
    return this.ws ? this.ws.readyState === WebSocket.OPEN : false
  }
} 