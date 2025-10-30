# 🚀 即時通訊功能開發路徑

## 📋 目標概述

將現有的串流平台擴展為具備 **LINE 級別即時通訊功能** 的綜合平台，包括：
- 一對一私聊
- 群組聊天  
- 音視訊通話
- 檔案分享（圖片、影片、文件）
- 好友系統管理
- 訊息持久化與搜尋

## 🏗️ 技術架構策略

### 核心原則
- ✅ **最大化現有投資複用**：基於現有 Go + Vue + PostgreSQL + Redis 架構
- ✅ **開源免費方案優先**：使用 NATS + Coturn 等開源服務
- ✅ **漸進式開發**：分階段實施，確保每階段都能獨立運行
- ✅ **技術棧一致性**：保持 Go 後端技術棧統一

### 新增技術組件
- **NATS JetStream**: 訊息隊列和持久化
- **Coturn**: STUN/TURN 服務器（WebRTC 支援）
- **WebRTC**: 瀏覽器原生音視訊通話
- **現有系統擴展**: WebSocket Hub、PostgreSQL、MinIO

## 🎯 階段性開發計劃

## 階段一：基礎聊天功能 (2-3週)

### 🗄️ 資料庫結構擴展

#### 新增資料表
```sql
-- 好友關係表
CREATE TABLE friendships (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    friend_id INTEGER REFERENCES users(id),
    status VARCHAR(20) DEFAULT 'pending', -- pending, accepted, blocked
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(user_id, friend_id)
);

-- 聊天會話表 (支援一對一和群組)
CREATE TABLE chat_sessions (
    id SERIAL PRIMARY KEY,
    type VARCHAR(20) NOT NULL, -- private, group
    name VARCHAR(255), -- 群組名稱，私聊可為空
    avatar_url VARCHAR(500), -- 群組頭像
    created_by INTEGER REFERENCES users(id),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 會話成員表
CREATE TABLE chat_session_members (
    id SERIAL PRIMARY KEY,
    session_id INTEGER REFERENCES chat_sessions(id),
    user_id INTEGER REFERENCES users(id),
    role VARCHAR(20) DEFAULT 'member', -- admin, member
    joined_at TIMESTAMP DEFAULT NOW(),
    last_read_at TIMESTAMP,
    UNIQUE(session_id, user_id)
);

-- 聊天訊息表
CREATE TABLE chat_messages (
    id SERIAL PRIMARY KEY,
    session_id INTEGER REFERENCES chat_sessions(id),
    sender_id INTEGER REFERENCES users(id),
    message_type VARCHAR(20) DEFAULT 'text', -- text, image, file, audio, video
    content TEXT, -- 文字內容
    file_url VARCHAR(500), -- 檔案 URL
    file_name VARCHAR(255), -- 原始檔案名
    file_size BIGINT, -- 檔案大小
    metadata JSONB, -- 額外元數據（如圖片尺寸）
    reply_to INTEGER REFERENCES chat_messages(id), -- 回覆訊息
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 訊息狀態表（已讀回條）
CREATE TABLE message_status (
    id SERIAL PRIMARY KEY,
    message_id INTEGER REFERENCES chat_messages(id),
    user_id INTEGER REFERENCES users(id),
    status VARCHAR(20) DEFAULT 'sent', -- sent, delivered, read
    timestamp TIMESTAMP DEFAULT NOW(),
    UNIQUE(message_id, user_id)
);
```

#### 索引優化
```sql
CREATE INDEX idx_friendships_user_id ON friendships(user_id);
CREATE INDEX idx_friendships_friend_id ON friendships(friend_id);
CREATE INDEX idx_chat_messages_session_id ON chat_messages(session_id);
CREATE INDEX idx_chat_messages_created_at ON chat_messages(created_at);
CREATE INDEX idx_message_status_message_id ON message_status(message_id);
CREATE INDEX idx_message_status_user_id ON message_status(user_id);
```

### 🔧 Go 後端開發

#### 新增資料模型
```go
// services/api/database/models/chat.go
type ChatSession struct {
    ID        uint      `gorm:"primaryKey"`
    Type      string    `gorm:"not null"` // private, group
    Name      string    `gorm:"size:255"`
    AvatarURL string    `gorm:"size:500"`
    CreatedBy uint      `gorm:"not null"`
    CreatedAt time.Time
    UpdatedAt time.Time
    
    // 關聯
    Members []ChatSessionMember `gorm:"foreignKey:SessionID"`
    Messages []ChatMessage     `gorm:"foreignKey:SessionID"`
}

type ChatSessionMember struct {
    ID         uint      `gorm:"primaryKey"`
    SessionID  uint      `gorm:"not null"`
    UserID     uint      `gorm:"not null"`
    Role       string    `gorm:"default:member"`
    JoinedAt   time.Time `gorm:"default:CURRENT_TIMESTAMP"`
    LastReadAt *time.Time
    
    // 關聯
    Session *ChatSession `gorm:"foreignKey:SessionID"`
    User    *User        `gorm:"foreignKey:UserID"`
}

type ChatMessage struct {
    ID          uint      `gorm:"primaryKey"`
    SessionID   uint      `gorm:"not null"`
    SenderID    uint      `gorm:"not null"`
    MessageType string    `gorm:"default:text"`
    Content     string    `gorm:"type:text"`
    FileURL     string    `gorm:"size:500"`
    FileName    string    `gorm:"size:255"`
    FileSize    int64
    Metadata    string    `gorm:"type:jsonb"`
    ReplyTo     *uint
    CreatedAt   time.Time
    UpdatedAt   time.Time
    
    // 關聯
    Session   *ChatSession `gorm:"foreignKey:SessionID"`
    Sender    *User        `gorm:"foreignKey:SenderID"`
    ReplyToMsg *ChatMessage `gorm:"foreignKey:ReplyTo"`
    Status    []MessageStatus `gorm:"foreignKey:MessageID"`
}

type Friendship struct {
    ID        uint      `gorm:"primaryKey"`
    UserID    uint      `gorm:"not null"`
    FriendID  uint      `gorm:"not null"`
    Status    string    `gorm:"default:pending"`
    CreatedAt time.Time
    UpdatedAt time.Time
    
    // 關聯
    User   *User `gorm:"foreignKey:UserID"`
    Friend *User `gorm:"foreignKey:FriendID"`
}
```

#### 新增服務層
```go
// services/api/services/chat.go
type ChatService struct {
    Config      *config.Config
    Repo        *postgresqlRepo.PostgreSQLRepo
    NATSClient  *nats.Conn
    RedisClient *redis.Client
    FileService *FileService
}

func NewChatService(config *config.Config) *ChatService {
    // 連接 NATS
    nc, err := nats.Connect("nats://localhost:4222")
    if err != nil {
        log.Fatal(err)
    }
    
    return &ChatService{
        Config:     config,
        NATSClient: nc,
        // ... 其他初始化
    }
}

func (cs *ChatService) CreatePrivateChat(userID1, userID2 uint) (*ChatSession, error) {
    // 檢查是否已存在私聊
    existingSession := cs.findExistingPrivateChat(userID1, userID2)
    if existingSession != nil {
        return existingSession, nil
    }
    
    // 創建新的私聊會話
    session := &ChatSession{
        Type:      "private",
        CreatedBy: userID1,
    }
    
    // 保存到資料庫
    if err := cs.Repo.CreateChatSession(session); err != nil {
        return nil, err
    }
    
    // 添加成員
    members := []ChatSessionMember{
        {SessionID: session.ID, UserID: userID1, Role: "member"},
        {SessionID: session.ID, UserID: userID2, Role: "member"},
    }
    
    for _, member := range members {
        cs.Repo.AddSessionMember(&member)
    }
    
    return session, nil
}

func (cs *ChatService) SendMessage(sessionID, senderID uint, content, messageType string) (*ChatMessage, error) {
    // 驗證用戶是否為會話成員
    if !cs.isSessionMember(sessionID, senderID) {
        return nil, errors.New("用戶不是會話成員")
    }
    
    // 創建訊息
    message := &ChatMessage{
        SessionID:   sessionID,
        SenderID:    senderID,
        Content:     content,
        MessageType: messageType,
    }
    
    // 保存到資料庫
    if err := cs.Repo.CreateMessage(message); err != nil {
        return nil, err
    }
    
    // 發布到 NATS
    if err := cs.publishMessage(message); err != nil {
        log.Printf("Failed to publish message to NATS: %v", err)
    }
    
    return message, nil
}

func (cs *ChatService) publishMessage(message *ChatMessage) error {
    // 序列化訊息
    msgData, err := json.Marshal(message)
    if err != nil {
        return err
    }
    
    // 發布到 NATS 主題
    subject := fmt.Sprintf("chat.session.%d", message.SessionID)
    return cs.NATSClient.Publish(subject, msgData)
}
```

### 🌐 前端開發

#### 新增 API 服務
```typescript
// services/frontend/src/api/chat.ts
export interface ChatSession {
  id: number;
  type: 'private' | 'group';
  name?: string;
  avatarUrl?: string;
  members: User[];
  lastMessage?: ChatMessage;
}

export interface ChatMessage {
  id: number;
  sessionId: number;
  senderId: number;
  messageType: 'text' | 'image' | 'file' | 'audio' | 'video';
  content: string;
  fileUrl?: string;
  fileName?: string;
  createdAt: string;
  sender: User;
}

class ChatAPI {
  async getChatSessions(): Promise<ChatSession[]> {
    const response = await request.get('/api/chat/sessions');
    return response.data;
  }
  
  async getSessionMessages(sessionId: number, page = 1): Promise<ChatMessage[]> {
    const response = await request.get(`/api/chat/sessions/${sessionId}/messages`, {
      params: { page }
    });
    return response.data;
  }
  
  async sendMessage(sessionId: number, content: string, type = 'text'): Promise<ChatMessage> {
    const response = await request.post(`/api/chat/sessions/${sessionId}/messages`, {
      content,
      messageType: type
    });
    return response.data;
  }
  
  async createPrivateChat(userId: number): Promise<ChatSession> {
    const response = await request.post('/api/chat/private', { userId });
    return response.data;
  }
}

export const chatAPI = new ChatAPI();
```

#### 聊天組件
```vue
<!-- services/frontend/src/components/chat/ChatWindow.vue -->
<template>
  <div class="chat-window">
    <div class="chat-header">
      <h3>{{ session.name || getFriendName(session) }}</h3>
      <el-button @click="startVideoCall" type="primary" size="small">
        視訊通話
      </el-button>
    </div>
    
    <div class="messages-container" ref="messagesContainer">
      <div v-for="message in messages" :key="message.id" class="message-item">
        <div :class="['message', { 'own-message': message.senderId === currentUser.id }]">
          <div class="message-content">{{ message.content }}</div>
          <div class="message-time">{{ formatTime(message.createdAt) }}</div>
        </div>
      </div>
    </div>
    
    <div class="message-input">
      <el-input
        v-model="newMessage"
        @keyup.enter="sendMessage"
        placeholder="輸入訊息..."
      >
        <template #append>
          <el-button @click="sendMessage" type="primary">發送</el-button>
        </template>
      </el-input>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, nextTick } from 'vue';
import { chatAPI } from '@/api/chat';
import { useWebSocketStore } from '@/store/websocket';
import type { ChatSession, ChatMessage } from '@/api/chat';

const props = defineProps<{
  session: ChatSession;
}>();

const messages = ref<ChatMessage[]>([]);
const newMessage = ref('');
const messagesContainer = ref<HTMLElement>();
const wsStore = useWebSocketStore();

onMounted(async () => {
  // 載入歷史訊息
  await loadMessages();
  
  // 訂閱 WebSocket 訊息
  wsStore.subscribeToSession(props.session.id, handleNewMessage);
});

const loadMessages = async () => {
  try {
    messages.value = await chatAPI.getSessionMessages(props.session.id);
    scrollToBottom();
  } catch (error) {
    console.error('載入訊息失敗:', error);
  }
};

const sendMessage = async () => {
  if (!newMessage.value.trim()) return;
  
  try {
    const message = await chatAPI.sendMessage(props.session.id, newMessage.value);
    messages.value.push(message);
    newMessage.value = '';
    scrollToBottom();
  } catch (error) {
    console.error('發送訊息失敗:', error);
  }
};

const handleNewMessage = (message: ChatMessage) => {
  messages.value.push(message);
  scrollToBottom();
};

const scrollToBottom = () => {
  nextTick(() => {
    if (messagesContainer.value) {
      messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight;
    }
  });
};
</script>
```

### 🔗 WebSocket 系統增強

#### 擴展 WebSocket Hub
```go
// services/api/ws/chat_hub.go
type ChatHub struct {
    // 現有欄位
    rooms       map[uint]*Room
    mu          sync.RWMutex
    messaging   *utils.RedisMessaging
    
    // 新增聊天相關
    chatSessions map[uint]*ChatSession // sessionID -> ChatSession
    userSessions map[uint][]uint       // userID -> []sessionID
    natsConn     *nats.Conn
}

type ChatSession struct {
    ID      uint
    clients map[*Client]bool
    mu      sync.RWMutex
}

func (h *ChatHub) JoinChatSession(userID, sessionID uint, conn *websocket.Conn) {
    h.mu.Lock()
    defer h.mu.Unlock()
    
    // 獲取或創建聊天會話
    chatSession, exists := h.chatSessions[sessionID]
    if !exists {
        chatSession = &ChatSession{
            ID:      sessionID,
            clients: make(map[*Client]bool),
        }
        h.chatSessions[sessionID] = chatSession
    }
    
    // 創建客戶端
    client := &Client{
        conn:      conn,
        userID:    userID,
        sessionID: sessionID,
    }
    
    // 添加到會話
    chatSession.mu.Lock()
    chatSession.clients[client] = true
    chatSession.mu.Unlock()
    
    // 啟動客戶端處理
    go client.writePump()
    go client.readPump()
}

func (h *ChatHub) BroadcastToSession(sessionID uint, message *ChatMessage) {
    h.mu.RLock()
    chatSession, exists := h.chatSessions[sessionID]
    h.mu.RUnlock()
    
    if !exists {
        return
    }
    
    // 序列化訊息
    msgData, err := json.Marshal(message)
    if err != nil {
        return
    }
    
    // 廣播給會話中的所有客戶端
    chatSession.mu.RLock()
    for client := range chatSession.clients {
        select {
        case client.send <- msgData:
        default:
            // 客戶端緩衝區滿，移除客戶端
            delete(chatSession.clients, client)
            close(client.send)
        }
    }
    chatSession.mu.RUnlock()
}
```

### 📡 NATS 訊息隊列整合

#### NATS 配置
```yaml
# infrastructure/nats/nats.conf
port: 4222
http_port: 8222

jetstream: enabled

jetstream {
    store_dir: "/data"
    max_memory: 256M
    max_file: 1G
}
```

#### Docker 配置
```yaml
# 添加到 docker-compose.dev.yml
services:
  nats:
    image: nats:alpine
    container_name: stream-demo-nats
    restart: unless-stopped
    ports:
      - "4222:4222"
      - "8222:8222"
    command: [
      "--jetstream",
      "--store_dir=/data",
      "--max_memory=256MB",
      "--max_file=1GB"
    ]
    volumes:
      - nats_data:/data
    networks:
      - stream-demo-network
    healthcheck:
      test: ["CMD", "nats", "server", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5

volumes:
  nats_data:
    driver: local
```

## 階段二：檔案分享功能 (1-2週)

### 📁 檔案上傳整合

#### 擴展現有檔案服務
```go
// services/api/services/file.go (擴展現有)
func (fs *FileService) UploadChatFile(file multipart.File, header *multipart.FileHeader, userID uint) (*ChatFile, error) {
    // 檔案類型檢查
    if !fs.isAllowedChatFileType(header.Filename) {
        return nil, errors.New("不支援的檔案類型")
    }
    
    // 檔案大小檢查 (50MB 限制)
    if header.Size > 50*1024*1024 {
        return nil, errors.New("檔案大小超過限制")
    }
    
    // 生成檔案路徑
    ext := filepath.Ext(header.Filename)
    filename := fmt.Sprintf("chat/%d/%s%s", userID, uuid.New().String(), ext)
    
    // 上傳到 MinIO
    url, err := fs.S3Storage.UploadChatFile(file, filename, header.Size)
    if err != nil {
        return nil, err
    }
    
    return &ChatFile{
        URL:      url,
        Filename: header.Filename,
        Size:     header.Size,
        Type:     fs.getFileType(ext),
    }, nil
}
```

#### 前端檔案上傳組件
```vue
<!-- services/frontend/src/components/chat/FileUpload.vue -->
<template>
  <div class="file-upload">
    <el-upload
      :action="uploadAction"
      :headers="uploadHeaders"
      :on-success="handleUploadSuccess"
      :on-error="handleUploadError"
      :show-file-list="false"
      accept=".jpg,.jpeg,.png,.gif,.mp4,.pdf,.doc,.docx"
    >
      <el-button type="primary" size="small">
        <el-icon><Upload /></el-icon>
        上傳檔案
      </el-button>
    </el-upload>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { useAuthStore } from '@/store/auth';
import { ElMessage } from 'element-plus';

const props = defineProps<{
  sessionId: number;
}>();

const emit = defineEmits<{
  fileUploaded: [fileInfo: any];
}>();

const authStore = useAuthStore();

const uploadAction = computed(() => `/api/chat/sessions/${props.sessionId}/files`);
const uploadHeaders = computed(() => ({
  Authorization: `Bearer ${authStore.token}`
}));

const handleUploadSuccess = (response: any) => {
  ElMessage.success('檔案上傳成功');
  emit('fileUploaded', response.data);
};

const handleUploadError = () => {
  ElMessage.error('檔案上傳失敗');
};
</script>
```

## 階段三：音視訊通話 (3-4週)

### 🌐 Coturn 服務部署

#### Coturn 配置
```bash
# infrastructure/coturn/turnserver.conf
listening-port=3478
tls-listening-port=5349

listening-ip=0.0.0.0
relay-ip=YOUR_SERVER_IP
external-ip=YOUR_SERVER_IP

min-port=49152
max-port=65535

user=turnuser:turnpass
realm=your-domain.com

no-stun
```

#### Docker 配置
```yaml
# 添加到 docker-compose.dev.yml
services:
  coturn:
    image: coturn/coturn:latest
    container_name: stream-demo-coturn
    restart: unless-stopped
    network_mode: host  # 需要 host 模式處理 NAT
    volumes:
      - ./infrastructure/coturn/turnserver.conf:/etc/turnserver.conf:ro
    environment:
      - DETECT_EXTERNAL_IP=yes
    healthcheck:
      test: ["CMD", "turnutils_uclient", "-t", "-u", "turnuser", "-w", "turnpass", "127.0.0.1"]
      interval: 30s
      timeout: 10s
      retries: 3
```

### 📞 WebRTC 信令服務

#### Go 信令服務
```go
// services/api/services/webrtc.go
type WebRTCService struct {
    Config     *config.Config
    Hub        *ws.Hub
    NATSClient *nats.Conn
}

type CallSession struct {
    ID           string
    CallerID     uint
    CalleeID     uint
    Type         string // audio, video
    Status       string // ringing, connected, ended
    StartTime    time.Time
    EndTime      *time.Time
}

func (w *WebRTCService) InitiateCall(callerID, calleeID uint, callType string) (*CallSession, error) {
    callID := uuid.New().String()
    
    session := &CallSession{
        ID:        callID,
        CallerID:  callerID,
        CalleeID:  calleeID,
        Type:      callType,
        Status:    "ringing",
        StartTime: time.Now(),
    }
    
    // 通知被叫方
    notification := map[string]interface{}{
        "type":     "incoming_call",
        "callId":   callID,
        "callerId": callerID,
        "callType": callType,
    }
    
    w.sendSignalingMessage(calleeID, notification)
    
    return session, nil
}

func (w *WebRTCService) HandleSignalingMessage(userID uint, message map[string]interface{}) error {
    msgType := message["type"].(string)
    
    switch msgType {
    case "offer":
        return w.handleOffer(userID, message)
    case "answer":
        return w.handleAnswer(userID, message)
    case "ice_candidate":
        return w.handleIceCandidate(userID, message)
    case "call_end":
        return w.handleCallEnd(userID, message)
    }
    
    return nil
}

func (w *WebRTCService) sendSignalingMessage(userID uint, message map[string]interface{}) error {
    // 通過 WebSocket 發送信令訊息
    return w.Hub.SendToUser(userID, message)
}
```

#### 前端 WebRTC 整合
```typescript
// services/frontend/src/utils/webrtc.ts
export class WebRTCManager {
  private peerConnection: RTCPeerConnection;
  private localStream: MediaStream | null = null;
  private remoteStream: MediaStream | null = null;
  private wsStore: any;

  constructor(wsStore: any) {
    this.wsStore = wsStore;
    this.peerConnection = new RTCPeerConnection({
      iceServers: [
        { urls: 'stun:stun.l.google.com:19302' },
        { 
          urls: 'turn:your-domain.com:3478',
          username: 'turnuser',
          credential: 'turnpass'
        }
      ]
    });

    this.setupPeerConnection();
  }

  private setupPeerConnection() {
    this.peerConnection.onicecandidate = (event) => {
      if (event.candidate) {
        this.wsStore.sendSignalingMessage({
          type: 'ice_candidate',
          candidate: event.candidate
        });
      }
    };

    this.peerConnection.ontrack = (event) => {
      this.remoteStream = event.streams[0];
      this.onRemoteStream?.(this.remoteStream);
    };
  }

  async startVideoCall(calleeId: number): Promise<void> {
    try {
      // 獲取本地媒體流
      this.localStream = await navigator.mediaDevices.getUserMedia({
        video: true,
        audio: true
      });

      // 添加本地流到 peer connection
      this.localStream.getTracks().forEach(track => {
        this.peerConnection.addTrack(track, this.localStream!);
      });

      // 創建 offer
      const offer = await this.peerConnection.createOffer();
      await this.peerConnection.setLocalDescription(offer);

      // 發送 offer 到對方
      this.wsStore.sendSignalingMessage({
        type: 'offer',
        calleeId,
        offer: offer
      });

      this.onLocalStream?.(this.localStream);
    } catch (error) {
      console.error('啟動視訊通話失敗:', error);
      throw error;
    }
  }

  async handleOffer(offer: RTCSessionDescriptionInit): Promise<void> {
    try {
      await this.peerConnection.setRemoteDescription(offer);

      // 獲取本地媒體流
      this.localStream = await navigator.mediaDevices.getUserMedia({
        video: true,
        audio: true
      });

      this.localStream.getTracks().forEach(track => {
        this.peerConnection.addTrack(track, this.localStream!);
      });

      // 創建 answer
      const answer = await this.peerConnection.createAnswer();
      await this.peerConnection.setLocalDescription(answer);

      // 發送 answer
      this.wsStore.sendSignalingMessage({
        type: 'answer',
        answer: answer
      });

      this.onLocalStream?.(this.localStream);
    } catch (error) {
      console.error('處理 offer 失敗:', error);
    }
  }

  async handleAnswer(answer: RTCSessionDescriptionInit): Promise<void> {
    try {
      await this.peerConnection.setRemoteDescription(answer);
    } catch (error) {
      console.error('處理 answer 失敗:', error);
    }
  }

  async handleIceCandidate(candidate: RTCIceCandidateInit): Promise<void> {
    try {
      await this.peerConnection.addIceCandidate(candidate);
    } catch (error) {
      console.error('添加 ICE candidate 失敗:', error);
    }
  }

  endCall(): void {
    if (this.localStream) {
      this.localStream.getTracks().forEach(track => track.stop());
      this.localStream = null;
    }

    this.peerConnection.close();
    this.onCallEnded?.();
  }

  // 回調函數
  onLocalStream?: (stream: MediaStream) => void;
  onRemoteStream?: (stream: MediaStream) => void;
  onCallEnded?: () => void;
}
```

## 階段四：進階功能 (2-3週)

### 📬 訊息已讀狀態
### 🔍 訊息搜尋功能  
### 👥 群組管理
### 📱 推送通知
### 🔄 多端同步

## 📊 開發時間估算

| 階段 | 功能 | 時間 | 人力 | 依賴 |
|-----|------|------|------|------|
| 階段一 | 基礎聊天 | 2-3週 | 1-2人 | 現有系統 |
| 階段二 | 檔案分享 | 1-2週 | 1人 | 階段一 |
| 階段三 | 音視訊通話 | 3-4週 | 1-2人 | 階段一 |
| 階段四 | 進階功能 | 2-3週 | 1人 | 階段一-三 |
| **總計** | **完整功能** | **8-12週** | **1-2人** | **循序漸進** |

## 🎯 成功指標

### 技術指標
- [ ] 支援 1000+ 同時在線用戶
- [ ] 訊息延遲 < 200ms
- [ ] 音視訊通話延遲 < 500ms
- [ ] 檔案上傳成功率 > 99%

### 功能指標
- [ ] 一對一聊天功能完整
- [ ] 群組聊天（最多 100 人）
- [ ] 音視訊通話穩定
- [ ] 檔案分享支援常見格式
- [ ] 好友系統管理完善

### 用戶體驗指標
- [ ] 界面響應時間 < 1s
- [ ] 離線訊息 100% 送達
- [ ] 多端同步即時性
- [ ] 通話品質優良

## 🚀 部署策略

### 開發環境
```bash
# 啟動新增服務
docker-compose -f docker-compose.dev.yml up -d nats coturn

# 檢查服務狀態
./scripts/docker-manage.sh status

# 初始化聊天系統
./scripts/docker-manage.sh init-chat
```

### 生產環境
```bash
# 完整部署
./scripts/deploy.sh --with-chat

# 監控服務
./scripts/monitor.sh chat-services
```

這個詳細的開發路徑確保了：
1. ✅ **階段性實施**：每個階段都能獨立運行和測試
2. ✅ **技術複用**：最大化利用現有架構和投資  
3. ✅ **開源免費**：所有新增組件都是開源免費方案
4. ✅ **可擴展性**：架構設計支援未來功能擴展
5. ✅ **風險控制**：漸進式開發降低技術風險
