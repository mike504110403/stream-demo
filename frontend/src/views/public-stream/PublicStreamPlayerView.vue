<template>
  <div class="live-stream-page">
    <!-- 頂部導航 -->
    <div class="top-nav">
      <div class="nav-left">
        <el-button @click="goBack" class="back-btn">
          <svg fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7"></path>
          </svg>
          返回
        </el-button>
      </div>
      <div class="stream-info">
        <h1 class="stream-title">{{ streamInfo?.title }}</h1>
      </div>
      <div class="nav-right">
        <div class="viewer-count">
          👥 {{ streamInfo?.viewer_count || 0 }}
        </div>
      </div>
    </div>

    <!-- 載入狀態 -->
    <div v-if="loading" class="loading-overlay">
      <div class="loading-content">
        <div class="loading-spinner">
          <div class="spinner-ring blue"></div>
          <div class="spinner-ring purple"></div>
        </div>
        <p>正在載入直播...</p>
      </div>
    </div>

    <!-- 錯誤狀態 -->
    <div v-else-if="error" class="error-overlay">
      <div class="error-content">
        <div class="error-icon">⚠️</div>
        <h3>載入失敗</h3>
        <p>{{ error }}</p>
        <el-button @click="loadStreamInfo" type="primary">重新載入</el-button>
      </div>
    </div>

    <!-- 主要內容區域 -->
    <div v-else-if="streamInfo" class="main-content">
      <!-- 左側播放器區域 -->
      <div class="player-section">
        <!-- 播放器容器 -->
        <div class="player-container" :class="{ 'fullscreen': isFullscreen }">
          <!-- 視頻播放器 -->
          <video
            v-if="streamInfo.status === 'active'"
            ref="videoPlayer"
            class="video-player"
            autoplay
            :muted="isMuted"
            crossorigin="anonymous"
            @click="toggleFullscreen"
          >
            您的瀏覽器不支援 HLS 播放。
          </video>

          <!-- 播放器控制層 -->
          <div class="player-controls" v-show="showControls">
                         <!-- 頂部控制 -->
             <div class="top-controls">
               <div class="live-badge">
                 <span class="live-dot"></span>
                 LIVE
               </div>
             </div>

            <!-- 底部控制 -->
            <div class="bottom-controls">
              <div class="control-left">
                <button @click="toggleMute" class="control-btn">
                  <svg v-if="isMuted" fill="currentColor" viewBox="0 0 24 24">
                    <path d="M16.5 12c0-1.77-1.02-3.29-2.5-4.03v2.21l2.45 2.45c.03-.2.05-.41.05-.63zm2.5 0c0 .94-.2 1.82-.54 2.64l1.51 1.51C20.63 14.91 21 13.5 21 12c0-4.28-2.99-7.86-7-8.77v2.06c2.89.86 5 3.54 5 6.71zM4.27 3L3 4.27 7.73 9H3v6h4l5 5v-6.73l4.25 4.25c-.67.52-1.42.93-2.25 1.18v2.06c1.38-.31 2.63-.95 3.69-1.81L19.73 21 21 19.73l-9-9L4.27 3zM12 4L9.91 6.09 12 8.18V4z"/>
                  </svg>
                  <svg v-else fill="currentColor" viewBox="0 0 24 24">
                    <path d="M3 9v6h4l5 5V4L7 9H3zm13.5 3c0-1.77-1.02-3.29-2.5-4.03v8.05c1.48-.73 2.5-2.25 2.5-4.02zM14 3.23v2.06c2.89.86 5 3.54 5 6.71s-2.11 5.85-5 6.71v2.06c4.01-.91 7-4.49 7-8.77s-2.99-7.86-7-8.77z"/>
                  </svg>
                </button>
                <div class="volume-slider">
                  <el-slider v-model="volume" :min="0" :max="100" @change="changeVolume" />
                </div>
              </div>
              <div class="control-right">
                <button @click="toggleFullscreen" class="control-btn">
                  <svg fill="currentColor" viewBox="0 0 24 24">
                    <path d="M7 14H5v5h5v-2H7v-3zm-2-4h2V7h3V5H5v5zm12 7h-3v2h5v-5h-2v3zM14 5v2h3v3h2V5h-5z"/>
                  </svg>
                </button>
                <button @click="rotateScreen" class="control-btn">
                  <svg fill="currentColor" viewBox="0 0 24 24">
                    <path d="M16.48 2.52c3.27 1.55 5.61 4.72 5.97 8.48h1.5C23.44 4.84 18.29 0 12 0l-.66.03 3.81 3.81 1.33-1.32zm-6.25-.77c-.59-.59-1.54-.59-2.12 0L1.75 8.11c-.59.59-.59 1.54 0 2.12l12.02 12.02c.59.59 1.54.59 2.12 0l6.36-6.36c.59-.59.59-1.54 0-2.12L10.23 1.75zm4.6 19.44L2.81 9.17l6.36-6.36 12.02 12.02-6.36 6.36z"/>
                  </svg>
                </button>
              </div>
            </div>
          </div>

          <!-- 載入狀態 -->
          <div v-if="loadingPlaybackUrl" class="loading-overlay">
            <div class="loading-content">
              <div class="loading-spinner">
                <div class="spinner-ring blue"></div>
                <div class="spinner-ring purple"></div>
              </div>
              <p>正在載入直播流...</p>
            </div>
          </div>
          
          <!-- 緩衝提示 - 更柔和的顯示 -->
          <div v-if="hlsLoading" class="buffering-indicator">
            <div class="buffering-dots">
              <span></span>
              <span></span>
              <span></span>
            </div>
            <p class="buffering-text">正在緩衝...</p>
          </div>
          
          <!-- 直播狀態指示器 - 顯示正在載入新內容 -->
          <div v-if="isLiveStreaming && !hlsLoading && !loadingPlaybackUrl" class="live-status-indicator">
            <div class="live-dot"></div>
            <p class="live-text">直播中 - 正在更新內容</p>
          </div>
          
          <!-- 播放按鈕 - 當影片暫停時顯示 -->
          <div v-if="videoPlayer?.paused && !loadingPlaybackUrl" class="play-button-overlay">
            <button @click="playVideo" class="play-button">
              <svg fill="currentColor" viewBox="0 0 24 24" width="48" height="48">
                <path d="M8 5v14l11-7z"/>
              </svg>
            </button>
            <p class="play-text">點擊播放直播</p>
          </div>
          
          <!-- 調試信息 -->
          <div v-if="true" class="debug-info" style="position: absolute; top: 10px; right: 10px; background: rgba(0,0,0,0.8); color: white; padding: 10px; border-radius: 5px; font-size: 12px; z-index: 1000;">
            <div>loadingPlaybackUrl: {{ loadingPlaybackUrl }}</div>
            <div>hlsLoading: {{ hlsLoading }}</div>
            <div>isLiveStreaming: {{ isLiveStreaming }}</div>
            <div>videoReadyState: {{ videoPlayer?.readyState }}</div>
            <div>videoPaused: {{ videoPlayer?.paused }}</div>
            <div>videoCurrentTime: {{ videoPlayer?.currentTime }}</div>
            <button @click="hlsLoading = !hlsLoading" style="margin-top: 5px; padding: 5px; background: #3b82f6; color: white; border: none; border-radius: 3px; cursor: pointer;">
              切換 Loading
            </button>
            <button @click="isLiveStreaming = !isLiveStreaming" style="margin-top: 5px; margin-left: 5px; padding: 5px; background: #ef4444; color: white; border: none; border-radius: 3px; cursor: pointer;">
              切換 Live
            </button>
          </div>
        </div>

        <!-- 流資訊卡片 -->
        <div class="stream-card">
          <div class="stream-details">
            <h3>{{ streamInfo.title }}</h3>
            <p>{{ streamInfo.description }}</p>
            <div class="stream-meta">
              <span class="category">{{ getCategoryLabel(streamInfo.category) }}</span>
              <span class="update-time">{{ formatTime(streamInfo.last_update) }}</span>
            </div>
          </div>
        </div>
      </div>

            <!-- 聊天室浮動按鈕 -->
      <div class="chat-toggle" @click="toggleChat">
        <svg fill="currentColor" viewBox="0 0 24 24">
          <path d="M20 2H4c-1.1 0-2 .9-2 2v12c0 1.1.9 2 2 2h4l4 4 4-4h4c1.1 0 2-.9 2-2V4c0-1.1-.9-2-2-2z"/>
        </svg>
        <span class="chat-badge" v-if="unreadCount > 0">{{ unreadCount }}</span>
      </div>

      <!-- 浮動聊天室面板 -->
      <div class="chat-panel" :class="{ 'chat-open': isChatOpen }">
        <div class="chat-header">
          <h3>聊天室</h3>
          <div class="chat-controls">
            <span class="online-count">在線 {{ streamInfo.viewer_count || 0 }}</span>
            <button @click="toggleChat" class="close-btn">
              <svg fill="currentColor" viewBox="0 0 24 24">
                <path d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z"/>
              </svg>
            </button>
          </div>
        </div>
        
        <div class="chat-messages" ref="chatMessagesRef">
          <div v-for="(message, index) in chatMessages" :key="index" class="message">
            <div class="message-avatar">
              <span>{{ message.username.charAt(0) }}</span>
            </div>
            <div class="message-content">
              <div class="message-header">
                <span class="username">{{ message.username }}</span>
                <span class="time">{{ formatTime(message.timestamp) }}</span>
              </div>
              <p class="message-text">{{ message.text }}</p>
            </div>
          </div>
        </div>

        <div class="chat-input">
          <el-input
            v-model="newMessage"
            placeholder="輸入訊息..."
            @keyup.enter="sendMessage"
            :disabled="!isLoggedIn"
          >
            <template #append>
              <el-button @click="sendMessage" :disabled="!newMessage.trim() || !isLoggedIn">
                發送
              </el-button>
            </template>
          </el-input>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.live-stream-page {
  height: 100vh;
  display: flex;
  flex-direction: column;
  background: #0f0f0f;
  overflow: hidden;
}

.top-nav {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 24px;
  background: rgba(0, 0, 0, 0.8);
  backdrop-filter: blur(10px);
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  z-index: 100;
}

.nav-left {
  flex-shrink: 0;
  width: 120px;
}

.nav-right {
  flex-shrink: 0;
  width: 120px;
  display: flex;
  justify-content: flex-end;
}

.back-btn {
  background: rgba(255, 255, 255, 0.1);
  border: 1px solid rgba(255, 255, 255, 0.2);
  color: white;
  z-index: 101;
  position: relative;
  min-width: 70px;
}

.stream-info {
  display: flex;
  align-items: center;
  gap: 16px;
  flex: 1;
  justify-content: center;
  padding: 0 20px;
}

.stream-title {
  color: white;
  font-size: 1.2rem;
  font-weight: 600;
  margin: 0;
}



.viewer-count {
  color: rgba(255, 255, 255, 0.8);
  font-size: 14px;
}

.main-content {
  flex: 1;
  display: flex;
  overflow: hidden;
}

.player-section {
  flex: 1;
  display: flex;
  flex-direction: column;
  padding: 20px;
}

.player-container {
  position: relative;
  background: #000;
  border-radius: 12px;
  overflow: hidden;
  width: 100%;
  height: 0;
  padding-bottom: 56.25%; /* 16:9 比例 */
  margin-bottom: 16px;
}

.player-container.fullscreen {
  position: fixed;
  top: 0;
  left: 0;
  width: 100vw;
  height: 100vh;
  z-index: 9999;
  border-radius: 0;
  padding-bottom: 0;
}

.video-player {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  object-fit: contain;
  cursor: pointer;
}

.video-player {
  width: 100%;
  height: 100%;
  object-fit: contain;
  cursor: pointer;
}

.player-controls {
  position: absolute;
  inset: 0;
  background: linear-gradient(to bottom, 
    rgba(0, 0, 0, 0.7) 0%, 
    transparent 20%, 
    transparent 80%, 
    rgba(0, 0, 0, 0.7) 100%);
  opacity: 0;
  transition: opacity 0.3s ease;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  padding: 16px;
}

.player-container:hover .player-controls {
  opacity: 1;
}

.top-controls {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.quality-selector {
  background: rgba(0, 0, 0, 0.8);
  border-radius: 8px;
}

.bottom-controls {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.control-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.control-btn {
  background: rgba(0, 0, 0, 0.8);
  border: none;
  color: white;
  padding: 8px;
  border-radius: 8px;
  cursor: pointer;
  transition: background 0.2s ease;
}

.control-btn:hover {
  background: rgba(255, 255, 255, 0.2);
}

.control-btn svg {
  width: 20px;
  height: 20px;
}

.volume-slider {
  width: 100px;
}

.control-right {
  display: flex;
  gap: 8px;
}

.stream-card {
  background: rgba(255, 255, 255, 0.05);
  border-radius: 12px;
  padding: 16px;
  border: 1px solid rgba(255, 255, 255, 0.1);
}

.stream-details h3 {
  color: white;
  margin: 0 0 8px 0;
  font-size: 1.1rem;
}

.stream-details p {
  color: rgba(255, 255, 255, 0.7);
  margin: 0 0 12px 0;
  font-size: 14px;
}

.stream-meta {
  display: flex;
  gap: 12px;
  font-size: 12px;
}

.category {
  background: rgba(59, 130, 246, 0.2);
  color: #3b82f6;
  padding: 4px 8px;
  border-radius: 8px;
}

.update-time {
  color: rgba(255, 255, 255, 0.5);
}

.chat-toggle {
  position: fixed;
  bottom: 20px;
  right: 20px;
  width: 56px;
  height: 56px;
  background: rgba(0, 0, 0, 0.8);
  border: 1px solid rgba(255, 255, 255, 0.2);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  color: white;
  z-index: 1000;
  transition: all 0.3s ease;
}

.chat-toggle:hover {
  background: rgba(255, 255, 255, 0.1);
  transform: scale(1.1);
}

.chat-toggle svg {
  width: 24px;
  height: 24px;
}

.chat-badge {
  position: absolute;
  top: -5px;
  right: -5px;
  background: #dc2626;
  color: white;
  border-radius: 50%;
  width: 20px;
  height: 20px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  font-weight: bold;
}

.chat-panel {
  position: fixed;
  top: 50%;
  right: -400px;
  transform: translateY(-50%);
  width: 350px;
  height: 500px;
  background: rgba(0, 0, 0, 0.9);
  border: 1px solid rgba(255, 255, 255, 0.2);
  border-radius: 12px;
  display: flex;
  flex-direction: column;
  z-index: 1001;
  transition: right 0.3s ease;
  backdrop-filter: blur(10px);
}

.chat-panel.chat-open {
  right: 20px;
}

.chat-header {
  padding: 16px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.chat-header h3 {
  color: white;
  margin: 0;
  font-size: 1rem;
}

.chat-controls {
  display: flex;
  align-items: center;
  gap: 12px;
}

.online-count {
  color: rgba(255, 255, 255, 0.6);
  font-size: 12px;
}

.close-btn {
  background: none;
  border: none;
  color: rgba(255, 255, 255, 0.6);
  cursor: pointer;
  padding: 4px;
  border-radius: 4px;
  transition: all 0.2s ease;
}

.close-btn:hover {
  background: rgba(255, 255, 255, 0.1);
  color: white;
}

.close-btn svg {
  width: 16px;
  height: 16px;
}

.chat-messages {
  flex: 1;
  padding: 16px;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.message {
  display: flex;
  gap: 8px;
}

.message-avatar {
  width: 32px;
  height: 32px;
  background: linear-gradient(135deg, #667eea, #764ba2);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-weight: bold;
  font-size: 12px;
  flex-shrink: 0;
}

.message-content {
  flex: 1;
}

.message-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 4px;
}

.username {
  color: white;
  font-weight: 600;
  font-size: 12px;
}

.time {
  color: rgba(255, 255, 255, 0.4);
  font-size: 10px;
}

.message-text {
  color: rgba(255, 255, 255, 0.8);
  font-size: 13px;
  margin: 0;
  line-height: 1.4;
}

.chat-input {
  padding: 16px;
  border-top: 1px solid rgba(255, 255, 255, 0.1);
}

.loading-overlay,
.error-overlay {
  position: absolute;
  inset: 0;
  background: rgba(0, 0, 0, 0.8);
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
}

.loading-content,
.error-content {
  text-align: center;
}

.loading-spinner {
  position: relative;
  margin-bottom: 16px;
}

.spinner-ring {
  position: absolute;
  inset: 0;
  border-radius: 50%;
  border: 4px solid transparent;
  animation: spin 1s linear infinite;
}

.spinner-ring.blue {
  border-top-color: #3b82f6;
}

.spinner-ring.purple {
  border-top-color: #8b5cf6;
  animation-delay: -0.5s;
}

@keyframes pulse {
  0%, 100% {
    opacity: 1;
  }
  50% {
    opacity: 0.5;
  }
}

@keyframes spin {
  0% {
    transform: rotate(0deg);
  }
  100% {
    transform: rotate(360deg);
  }
}

/* 緩衝指示器樣式 */
.buffering-indicator {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  background: rgba(0, 0, 0, 0.6);
  backdrop-filter: blur(8px);
  border-radius: 12px;
  padding: 16px 24px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  z-index: 10;
  border: 1px solid rgba(255, 255, 255, 0.1);
}

.buffering-dots {
  display: flex;
  gap: 6px;
}

.buffering-dots span {
  width: 8px;
  height: 8px;
  background: #3b82f6;
  border-radius: 50%;
  animation: buffering-pulse 1.4s ease-in-out infinite both;
}

.buffering-dots span:nth-child(1) {
  animation-delay: -0.32s;
}

.buffering-dots span:nth-child(2) {
  animation-delay: -0.16s;
}

.buffering-dots span:nth-child(3) {
  animation-delay: 0s;
}

.buffering-text {
  color: rgba(255, 255, 255, 0.9);
  font-size: 14px;
  font-weight: 500;
  margin: 0;
}

@keyframes buffering-pulse {
  0%, 80%, 100% {
    transform: scale(0.8);
    opacity: 0.5;
  }
  40% {
    transform: scale(1);
    opacity: 1;
  }
}

/* 直播狀態指示器樣式 */
.live-status-indicator {
  position: absolute;
  top: 20px;
  right: 20px;
  background: rgba(0, 0, 0, 0.7);
  backdrop-filter: blur(8px);
  border-radius: 8px;
  padding: 8px 12px;
  display: flex;
  align-items: center;
  gap: 8px;
  z-index: 10;
  border: 1px solid rgba(255, 255, 255, 0.1);
}

.live-dot {
  width: 8px;
  height: 8px;
  background: #ef4444;
  border-radius: 50%;
  animation: live-pulse 2s ease-in-out infinite;
}

.live-text {
  color: rgba(255, 255, 255, 0.9);
  font-size: 12px;
  font-weight: 500;
  margin: 0;
}

@keyframes live-pulse {
  0%, 100% {
    opacity: 1;
    transform: scale(1);
  }
  50% {
    opacity: 0.7;
    transform: scale(1.2);
  }
}

/* 播放按鈕樣式 */
.play-button-overlay {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
  z-index: 20;
}

.play-button {
  width: 80px;
  height: 80px;
  background: rgba(0, 0, 0, 0.8);
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-radius: 50%;
  color: white;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.3s ease;
  backdrop-filter: blur(8px);
}

.play-button:hover {
  background: rgba(0, 0, 0, 0.9);
  border-color: rgba(255, 255, 255, 0.5);
  transform: scale(1.1);
}

.play-text {
  color: rgba(255, 255, 255, 0.9);
  font-size: 14px;
  font-weight: 500;
  margin: 0;
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.8);
}

/* 響應式設計 */
@media (max-width: 768px) {
  .main-content {
    flex-direction: column;
  }
  
  .player-section {
    padding: 12px;
  }
  
  .top-nav {
    padding: 12px 16px;
  }
  
  .nav-left,
  .nav-right {
    width: 80px;
  }
  
  .stream-info {
    padding: 0 10px;
  }
  
  .stream-title {
    font-size: 1rem;
  }
  
  .chat-panel {
    width: calc(100vw - 40px);
    height: 400px;
    right: -100vw;
  }
  
  .chat-panel.chat-open {
    right: 20px;
  }
  
  .chat-toggle {
    bottom: 16px;
    right: 16px;
    width: 48px;
    height: 48px;
  }
  
  .chat-toggle svg {
    width: 20px;
    height: 20px;
  }
}

/* 超寬螢幕適配 */
@media (min-width: 1920px) {
  .player-container {
    max-width: 1600px;
    margin: 0 auto 16px auto;
  }
}

/* 高螢幕適配 */
@media (min-height: 1080px) {
  .player-section {
    padding: 24px;
  }
  
  .stream-card {
    padding: 20px;
  }
}

/* 低螢幕適配 */
@media (max-height: 600px) {
  .top-nav {
    padding: 8px 16px;
  }
  
  .stream-title {
    font-size: 0.9rem;
  }
  
  .player-section {
    padding: 8px;
  }
  
  .stream-card {
    padding: 12px;
  }
}

.page-header {
  text-align: center;
  margin-bottom: 30px;
}

.page-title {
  font-size: 2.5rem;
  font-weight: bold;
  color: white;
  margin-bottom: 8px;
  text-shadow: 0 2px 4px rgba(0, 0, 0, 0.3);
}

.page-subtitle {
  font-size: 1.1rem;
  color: rgba(255, 255, 255, 0.8);
}

.player-section {
  max-width: 1200px;
  margin: 0 auto;
}

.back-button {
  margin-bottom: 20px;
  background: rgba(255, 255, 255, 0.1);
  border: 1px solid rgba(255, 255, 255, 0.2);
  color: white;
  backdrop-filter: blur(10px);
}

.back-button:hover {
  background: rgba(255, 255, 255, 0.2);
}

.stream-info-card {
  background: rgba(255, 255, 255, 0.1);
  backdrop-filter: blur(10px);
  border-radius: 16px;
  padding: 24px;
  margin-bottom: 24px;
  border: 1px solid rgba(255, 255, 255, 0.2);
}

.stream-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 16px;
}

.stream-details {
  flex: 1;
}

.stream-title {
  font-size: 1.8rem;
  font-weight: bold;
  color: white;
  margin-bottom: 8px;
}

.stream-description {
  color: rgba(255, 255, 255, 0.8);
  line-height: 1.5;
}

.stream-status {
  display: flex;
  gap: 12px;
  align-items: center;
}

.status-badge {
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
  padding: 6px 12px;
  background: rgba(0, 0, 0, 0.7);
  color: white;
  border-radius: 20px;
  font-size: 12px;
  font-weight: 500;
  backdrop-filter: blur(10px);
}

.stream-meta {
  display: flex;
  gap: 16px;
  align-items: center;
}

.category-tag {
  padding: 6px 12px;
  background: linear-gradient(135deg, #e0f2fe, #b3e5fc);
  color: #0277bd;
  border-radius: 20px;
  font-size: 12px;
  font-weight: 500;
}

.update-time {
  color: rgba(255, 255, 255, 0.7);
  font-size: 12px;
  display: flex;
  align-items: center;
  gap: 6px;
}

.update-time svg {
  width: 14px;
  height: 14px;
}

.player-container {
  max-width: 90vw;
  margin: 0 auto;
}

.player-frame {
  background: #000;
  border-radius: 16px;
  overflow: hidden;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
  border: 4px solid rgba(255, 255, 255, 0.2);
  position: relative;
}

.player-decoration {
  position: absolute;
  inset: 0;
  background: linear-gradient(135deg, rgba(59, 130, 246, 0.1), rgba(147, 51, 234, 0.1), rgba(236, 72, 153, 0.1));
  border-radius: 16px;
  pointer-events: none;
}

.player-title-bar {
  background: linear-gradient(135deg, #374151, #1f2937);
  padding: 12px 24px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

.title-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.traffic-lights {
  display: flex;
  align-items: center;
  gap: 12px;
}

.traffic-light {
  width: 12px;
  height: 12px;
  border-radius: 50%;
}

.traffic-light.red { background: #ef4444; }
.traffic-light.yellow { background: #f59e0b; }
.traffic-light.green { background: #10b981; }

.stream-title {
  color: white;
  font-size: 14px;
  font-weight: 500;
  margin-left: 8px;
}

.stream-info {
  display: flex;
  align-items: center;
  gap: 8px;
}

.stream-type {
  color: #9ca3af;
  font-size: 12px;
}

.status-indicator {
  width: 8px;
  height: 8px;
  background: #10b981;
  border-radius: 50%;
  animation: pulse 2s infinite;
}

.video-container {
  aspect-ratio: 16/9;
  position: relative;
  background: #000;
}

.live-indicator {
  position: absolute;
  top: 16px;
  left: 16px;
  z-index: 10;
}

.live-badge {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 12px;
  background: #dc2626;
  color: white;
  border-radius: 20px;
  font-size: 12px;
  font-weight: bold;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
}

.live-dot {
  width: 8px;
  height: 8px;
  background: white;
  border-radius: 50%;
  animation: pulse 2s infinite;
}

.video-player {
  width: 100%;
  height: 100%;
  object-fit: contain;
}

.loading-overlay,
.error-overlay {
  position: absolute;
  inset: 0;
  background: rgba(0, 0, 0, 0.8);
  display: flex;
  align-items: center;
  justify-content: center;
}

.loading-content,
.error-content {
  text-align: center;
}

.loading-spinner {
  position: relative;
  margin-bottom: 16px;
}

.spinner-ring {
  position: absolute;
  inset: 0;
  border-radius: 50%;
  border: 4px solid transparent;
  animation: spin 1s linear infinite;
}

.spinner-ring.blue {
  border-top-color: #3b82f6;
}

.spinner-ring.purple {
  border-top-color: #8b5cf6;
  animation-delay: -0.5s;
}

.loading-text {
  color: white;
  font-size: 14px;
}

.error-icon {
  color: #f87171;
  margin-bottom: 16px;
}

.error-icon svg {
  width: 64px;
  height: 64px;
}

.error-title {
  font-size: 18px;
  font-weight: 600;
  color: white;
  margin-bottom: 8px;
}

.error-message {
  color: #d1d5db;
  margin-bottom: 16px;
  font-size: 14px;
}

.retry-button {
  margin-top: 16px;
}

.player-control-bar {
  background: linear-gradient(135deg, #374151, #1f2937);
  padding: 12px 24px;
  border-top: 1px solid rgba(255, 255, 255, 0.1);
}

.control-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 12px;
}

.control-info {
  display: flex;
  align-items: center;
  gap: 16px;
}

.info-item {
  color: #9ca3af;
}

.control-status {
  display: flex;
  align-items: center;
  gap: 8px;
}

.status-dot {
  color: #10b981;
  font-size: 8px;
}

.status-dot.pulse {
  animation: pulse 2s infinite;
}

.status-text {
  color: #9ca3af;
}

@keyframes pulse {
  0%, 100% {
    opacity: 1;
  }
  50% {
    opacity: 0.5;
  }
}

@keyframes spin {
  0% {
    transform: rotate(0deg);
  }
  100% {
    transform: rotate(360deg);
  }
}

/* 響應式設計 */
@media (max-width: 768px) {
  .player-container {
    max-width: 95vw;
  }
  
  .page-title {
    font-size: 1.5rem;
  }
  
  .public-stream-player {
    padding: 16px;
  }
  
  .control-info {
    gap: 8px;
  }
  
  .info-item {
    font-size: 10px;
  }
}

@media (max-width: 480px) {
  .player-container {
    max-width: 100vw;
  }
  
  .page-title {
    font-size: 1.3rem;
  }
  
  .traffic-lights {
    gap: 8px;
  }
  
  .traffic-light {
    width: 10px;
    height: 10px;
  }
}
</style>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { publicStreamApi } from '@/api/public-stream'
import type { PublicStreamInfo } from '@/types/public-stream'
import Hls from 'hls.js'
import flvjs from 'flv.js'

const route = useRoute()
const router = useRouter()

// 響應式數據
const streamInfo = ref<PublicStreamInfo | null>(null)
const playbackUrl = ref('')
const loading = ref(false)
const loadingPlaybackUrl = ref(false)
const hlsLoading = ref(false)
const error = ref('')
const videoPlayer = ref<HTMLVideoElement>()
const hls = ref<Hls | null>(null)
const flvPlayer = ref<flvjs.Player | null>(null)
const streamURLs = ref<{ hls: string } | null>(null)

// 播放器控制
const isFullscreen = ref(false)
const isMuted = ref(false)
const volume = ref(50)
const showControls = ref(true)

// 聊天室
const chatMessages = ref<Array<{
  username: string
  text: string
  timestamp: string
}>>([])
const newMessage = ref('')
const chatMessagesRef = ref<HTMLElement | null>(null)
const isLoggedIn = ref(true) // 簡化，實際應該從 auth store 獲取
const streamMonitorInterval = ref<number | null>(null)
const isChatOpen = ref(false)
const unreadCount = ref(0)
const isLiveStreaming = ref(false)

// 分類標籤映射
const categoryLabels: Record<string, string> = {
  test: '測試',
  space: '太空',
  news: '新聞',
  sports: '體育'
}

// 方法
const loadStreamInfo = async () => {
  const streamName = route.params.name as string
  if (!streamName) {
    error.value = '無效的流名稱'
    return
  }

  loading.value = true
  error.value = ''
  
  try {
    const response = await publicStreamApi.getStreamInfo(streamName)
    
    // 添加防護性檢查
    if (!response) {
      console.error('API 響應為空')
      error.value = '載入流資訊失敗：API 響應為空'
      return
    }
    
    console.log('流資訊載入成功:', response)
    streamInfo.value = response
    
    // 如果流是活躍的，獲取播放 URL
    if (response.status === 'active') {
      console.log('流狀態為 active，開始載入播放 URL')
      // 延遲一下再載入播放 URL，確保 DOM 已更新
      setTimeout(() => {
        loadPlaybackUrl(streamName)
      }, 500)
    } else {
      console.log('流狀態不是 active:', response.status)
    }
  } catch (err) {
    console.error('載入流資訊失敗:', err)
    error.value = '載入流資訊失敗，請稍後再試'
  } finally {
    loading.value = false
  }
}

const loadPlaybackUrl = async (streamName: string, mode: 'hls' = 'hls') => {
  loadingPlaybackUrl.value = true
  
  try {
    // 獲取播放 URL
    if (!streamURLs.value) {
      const response = await publicStreamApi.getStreamURLs(streamName)
      streamURLs.value = {
        hls: response.urls.hls || ''
      }
    }
    
    // 根據模式選擇 URL
    if (mode === 'hls') {
      playbackUrl.value = streamURLs.value!.hls
      console.log('HLS 播放 URL:', playbackUrl.value)
      
      // 等待 videoPlayer 元素準備好
      await nextTick()
      
      // 使用輪詢等待 videoPlayer 元素
      let attempts = 0
      while (!videoPlayer.value && attempts < 20) {
        await new Promise(resolve => setTimeout(resolve, 100))
        attempts++
        console.log(`等待 videoPlayer 元素... 嘗試 ${attempts}/20`)
      }
      
      if (!videoPlayer.value) {
        console.error('videoPlayer 元素未準備好')
        error.value = '播放器元素未準備好，請重新整理頁面'
        return
      }
      
      console.log('videoPlayer 元素已準備好:', !!videoPlayer.value)
      console.log('HLS.js 支援:', Hls.isSupported())
      
      // 使用 HLS.js 載入流
      if (Hls.isSupported()) {
        console.log('使用 HLS.js 載入流')
        
        // 清理之前的播放器
        if (hls.value) {
          hls.value.destroy()
          hls.value = null
        }
        if (flvPlayer.value) {
          flvPlayer.value.destroy()
          flvPlayer.value = null
        }
        
        // 移除禁止播放標記
        if (videoPlayer.value) {
          videoPlayer.value.removeAttribute('data-no-play')
        }
        
        // 顯示 HLS 載入狀態
        hlsLoading.value = true
        
        // 創建新的 HLS 實例
        hls.value = new Hls({
          debug: true, // 開啟調試模式
          enableWorker: true,
          lowLatencyMode: true,
          // 增加緩衝配置 - 目標 30 秒緩衝
          maxBufferLength: 30, // 最大緩衝長度 30 秒
          maxMaxBufferLength: 60, // 最大緩衝長度上限 60 秒
          maxBufferSize: 100 * 1000 * 1000, // 最大緩衝大小 100MB
          maxBufferHole: 0.5, // 最大緩衝空洞 0.5 秒
          highBufferWatchdogPeriod: 2, // 高緩衝監控週期 2 秒
          nudgeOffset: 0.2, // 推動偏移 0.2 秒
          nudgeMaxRetry: 5, // 最大重試次數 5 次
          maxFragLookUpTolerance: 0.25, // 最大片段查找容差 0.25 秒
          liveSyncDurationCount: 6, // 直播同步片段數量 6 個 (約 30 秒)
          liveMaxLatencyDurationCount: 12, // 最大延遲片段數量 12 個
          // 片段載入配置
          fragLoadingMaxRetry: 4, // 片段載入最大重試次數
          fragLoadingRetryDelay: 1000, // 片段載入重試延遲 1 秒
          fragLoadingMaxRetryTimeout: 64000, // 片段載入最大重試超時 64 秒
          // 播放列表配置
          manifestLoadingMaxRetry: 4, // 播放列表載入最大重試次數
          manifestLoadingRetryDelay: 1000, // 播放列表載入重試延遲 1 秒
          manifestLoadingMaxRetryTimeout: 64000, // 播放列表載入最大重試超時 64 秒
        })
        
        // 立即顯示載入狀態
        hlsLoading.value = true
        
                // 載入流
        hls.value.loadSource(playbackUrl.value)
        hls.value.attachMedia(videoPlayer.value)
        
        // 監聽事件
        hls.value.on(Hls.Events.MANIFEST_PARSED, () => {
          console.log('HLS 流載入成功')
          // 不要立即隱藏載入狀態，等待片段載入
          
          // 等待足夠的緩衝後再播放
          setTimeout(() => {
            if (videoPlayer.value && !videoPlayer.value.hasAttribute('data-no-play')) {
              console.log('嘗試 HLS 自動播放')
              videoPlayer.value.play().catch(err => {
                console.error('自動播放失敗:', err)
              })
            } else {
              console.log('跳過 HLS 自動播放，禁止播放:', videoPlayer.value?.hasAttribute('data-no-play'))
            }
          }, 4000) // 等待 4 秒確保有足夠緩衝
        })
        
        hls.value.on(Hls.Events.ERROR, (_event, data) => {
          console.error('HLS 錯誤:', data)
          hlsLoading.value = false // 隱藏載入狀態
          if (data.fatal) {
            error.value = '播放器載入失敗，請重新整理頁面'
          }
        })
        
        hls.value.on(Hls.Events.MEDIA_ATTACHED, () => {
          console.log('媒體元素已附加')
        })
        
        // 監聽片段載入狀態
        hls.value.on(Hls.Events.BUFFER_APPENDING, () => {
          console.log('正在追加緩衝')
          hlsLoading.value = true
        })
        
        hls.value.on(Hls.Events.BUFFER_APPENDED, () => {
          console.log('緩衝追加完成')
          hlsLoading.value = false
        })
        

        
        // 監聽播放列表更新
        hls.value.on(Hls.Events.MANIFEST_LOADING, () => {
          console.log('正在載入播放列表')
          hlsLoading.value = true
        })
        
        // 添加定時器監控流狀態
        let lastFragmentTime = Date.now()
        let fragmentCount = 0
        let manifestCount = 0
        let isBuffering = false
        
        // 監聽播放列表載入
        hls.value.on(Hls.Events.MANIFEST_LOADED, () => {
          manifestCount++
          console.log('m3u8 載入完成，計數:', manifestCount)
          // 不要立即隱藏 loading，等待片段載入
        })
        
        // 監聽片段載入開始
        hls.value.on(Hls.Events.FRAG_LOADING, () => {
          console.log('正在載入片段')
          isBuffering = true
          hlsLoading.value = true
        })
        
        // 監聽片段載入完成
        hls.value.on(Hls.Events.FRAG_LOADED, () => {
          console.log('片段載入完成')
          lastFragmentTime = Date.now()
          fragmentCount++
          console.log('片段載入完成，計數:', fragmentCount)
          
          // 延遲隱藏 loading，確保有足夠緩衝
          setTimeout(() => {
            if (!isBuffering) {
              hlsLoading.value = false
            }
          }, 2000) // 增加到 2 秒，確保有足夠緩衝
        })
        
        // 監控流是否卡住
        streamMonitorInterval.value = window.setInterval(() => {
          if (videoPlayer.value && hls.value) {
            const currentTime = Date.now()
            const timeSinceLastFragment = currentTime - lastFragmentTime
            
            // 如果超過 5 秒沒有新片段，但一直在載入 m3u8，顯示 loading
            if (timeSinceLastFragment > 5000 && manifestCount > fragmentCount) {
              console.log('有 m3u8 但沒有 .ts 片段，顯示載入狀態')
              hlsLoading.value = true
            }
            
            // 如果超過 8 秒沒有新片段，且影片正在等待，顯示 loading
            if (timeSinceLastFragment > 8000 && videoPlayer.value.readyState < 3) {
              console.log('流可能卡住，顯示載入狀態')
              hlsLoading.value = true
            }
            
            // 檢查是否正在直播（有持續的片段載入）
            if (fragmentCount > 0 && timeSinceLastFragment < 10000) {
              isLiveStreaming.value = true
            } else {
              isLiveStreaming.value = false
            }
            
            // 重置緩衝狀態
            isBuffering = false
          }
        }, 2000) // 每 2 秒檢查一次
        
        // 監聽影片播放狀態
        if (videoPlayer.value) {
          videoPlayer.value.addEventListener('waiting', () => {
            console.log('影片等待中，顯示載入狀態')
            hlsLoading.value = true
          })
          
          videoPlayer.value.addEventListener('canplay', () => {
            console.log('影片可以播放，隱藏載入狀態')
            setTimeout(() => {
              hlsLoading.value = false
            }, 1000)
          })
          
          videoPlayer.value.addEventListener('stalled', () => {
            console.log('影片停滯，顯示載入狀態')
            hlsLoading.value = true
          })
          
          videoPlayer.value.addEventListener('suspend', () => {
            console.log('影片暫停載入，顯示載入狀態')
            hlsLoading.value = true
          })
          
          videoPlayer.value.addEventListener('loadstart', () => {
            console.log('影片開始載入')
            hlsLoading.value = true
          })
          
          videoPlayer.value.addEventListener('loadeddata', () => {
            console.log('影片數據載入完成')
            setTimeout(() => {
              hlsLoading.value = false
            }, 1000)
          })
        }
        
        // 監聽播放狀態
        if (videoPlayer.value) {
          videoPlayer.value.addEventListener('waiting', () => {
            console.log('影片等待數據，顯示載入狀態')
            hlsLoading.value = true
          })
          
          videoPlayer.value.addEventListener('canplay', () => {
            console.log('影片可以播放，隱藏載入狀態')
            hlsLoading.value = false
          })
          
          videoPlayer.value.addEventListener('stalled', () => {
            console.log('影片停滯，顯示載入狀態')
            hlsLoading.value = true
          })
        }
        
      } else if (videoPlayer.value.canPlayType('application/vnd.apple.mpegurl')) {
        console.log('使用瀏覽器原生 HLS 支援')
        // Safari 原生支援 HLS
        videoPlayer.value.src = playbackUrl.value
        videoPlayer.value.addEventListener('loadedmetadata', () => {
          if (!videoPlayer.value?.hasAttribute('data-no-play')) {
            console.log('嘗試 Safari 原生 HLS 自動播放')
            videoPlayer.value?.play().catch(err => {
              console.error('自動播放失敗:', err)
            })
          } else {
            console.log('跳過 Safari 原生 HLS 自動播放，禁止播放:', videoPlayer.value?.hasAttribute('data-no-play'))
          }
        })
      } else {
        console.error('瀏覽器不支援 HLS')
        error.value = '您的瀏覽器不支援 HLS 播放'
      }

    }
  } catch (err) {
    console.error('獲取播放 URL 失敗:', err)
    error.value = '獲取播放 URL 失敗'
  } finally {
    loadingPlaybackUrl.value = false
  }
}

// 這些功能暫時未使用，保留以備將來擴展
// const toggleMute = () => {
//   if (videoPlayer.value) {
//     videoPlayer.value.muted = !videoPlayer.value.muted
//     isMuted.value = videoPlayer.value.muted
//   }
// }

// const toggleFullscreen = () => {
//   if (videoPlayer.value) {
//     if (document.fullscreenElement) {
//       document.exitFullscreen()
//     } else {
//       videoPlayer.value.requestFullscreen()
//     }
//   }
// }

const playVideo = async () => {
  if (videoPlayer.value) {
    try {
      console.log('手動播放影片')
      await videoPlayer.value.play()
    } catch (err) {
      console.error('播放失敗:', err)
      error.value = '播放失敗，請檢查瀏覽器設定'
    }
  }
}

const goBack = () => {
  router.push('/public-streams')
}

const getCategoryLabel = (category: string) => {
  return categoryLabels[category] || category
}

const formatTime = (timeString: string) => {
  const date = new Date(timeString)
  return date.toLocaleString('zh-TW')
}

// 播放器控制方法
const toggleMute = () => {
  if (videoPlayer.value) {
    videoPlayer.value.muted = !videoPlayer.value.muted
    isMuted.value = videoPlayer.value.muted
  }
}

const changeVolume = (value: number | number[]) => {
  const volumeValue = Array.isArray(value) ? value[0] : value
  if (videoPlayer.value) {
    videoPlayer.value.volume = volumeValue / 100
  }
}

const toggleFullscreen = () => {
  if (videoPlayer.value) {
    if (document.fullscreenElement) {
      document.exitFullscreen()
      isFullscreen.value = false
    } else {
      videoPlayer.value.requestFullscreen()
      isFullscreen.value = true
    }
  }
}

// 監聽全螢幕狀態變化
const handleFullscreenChange = () => {
  isFullscreen.value = !!document.fullscreenElement
}

const rotateScreen = () => {
  if (videoPlayer.value) {
    const currentRotation = videoPlayer.value.style.transform
    const newRotation = currentRotation.includes('rotate(90deg)') ? '' : 'rotate(90deg)'
    videoPlayer.value.style.transform = newRotation
  }
}

// 聊天室方法
const toggleChat = () => {
  isChatOpen.value = !isChatOpen.value
  if (isChatOpen.value) {
    unreadCount.value = 0 // 打開聊天室時清除未讀數
  }
}

const sendMessage = () => {
  if (newMessage.value.trim() && isLoggedIn.value) {
    chatMessages.value.push({
      username: '用戶',
      text: newMessage.value,
      timestamp: new Date().toISOString()
    })
    newMessage.value = ''
    
    // 滾動到底部
    nextTick(() => {
      if (chatMessagesRef.value) {
        chatMessagesRef.value.scrollTop = chatMessagesRef.value.scrollHeight
      }
    })
  }
}





// 生命週期
onMounted(async () => {
  await loadStreamInfo()
  
  // 添加全螢幕事件監聽器
  document.addEventListener('fullscreenchange', handleFullscreenChange)
  document.addEventListener('webkitfullscreenchange', handleFullscreenChange)
  document.addEventListener('mozfullscreenchange', handleFullscreenChange)
  document.addEventListener('MSFullscreenChange', handleFullscreenChange)
})

onUnmounted(() => {
  // 清理事件監聽器
  document.removeEventListener('fullscreenchange', handleFullscreenChange)
  document.removeEventListener('webkitfullscreenchange', handleFullscreenChange)
  document.removeEventListener('mozfullscreenchange', handleFullscreenChange)
  document.removeEventListener('MSFullscreenChange', handleFullscreenChange)
  
  // 清理 HLS 實例
  if (hls.value) {
    hls.value.destroy()
    hls.value = null
  }
  
  // 清理定時器
  if (streamMonitorInterval.value) {
    clearInterval(streamMonitorInterval.value)
    streamMonitorInterval.value = null
  }
})

onUnmounted(() => {
  // 清理播放器
  if (videoPlayer.value) {
    videoPlayer.value.pause()
    videoPlayer.value.src = ''
  }
  
  // 清理 HLS 實例
  if (hls.value) {
    hls.value.destroy()
    hls.value = null
  }
  
  // 清理 FLV 實例
  if (flvPlayer.value) {
    flvPlayer.value.destroy()
    flvPlayer.value = null
  }
})
</script>

<style scoped>
/* 自定義播放器樣式 */
video::-webkit-media-controls {
  background-color: rgba(0, 0, 0, 0.5);
}

video::-webkit-media-controls-panel {
  background-color: rgba(0, 0, 0, 0.5);
}

/* 自定義動畫 */
@keyframes pulse {
  0%, 100% {
    opacity: 1;
  }
  50% {
    opacity: 0.5;
  }
}

.animate-pulse {
  animation: pulse 2s cubic-bezier(0.4, 0, 0.6, 1) infinite;
}

/* 玻璃擬態效果 */
.backdrop-blur-sm {
  backdrop-filter: blur(8px);
  -webkit-backdrop-filter: blur(8px);
}

/* 漸變文字效果 */
.bg-clip-text {
  -webkit-background-clip: text;
  background-clip: text;
}

/* 自定義滾動條 */
::-webkit-scrollbar {
  width: 8px;
}

::-webkit-scrollbar-track {
  background: rgba(255, 255, 255, 0.1);
  border-radius: 4px;
}

::-webkit-scrollbar-thumb {
  background: linear-gradient(to bottom, #3b82f6, #8b5cf6);
  border-radius: 4px;
}

::-webkit-scrollbar-thumb:hover {
  background: linear-gradient(to bottom, #2563eb, #7c3aed);
}

/* 卡片懸停效果 */
.hover\:shadow-2xl:hover {
  box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.25);
}

/* 按鈕點擊效果 */
.transform:active {
  transform: scale(0.95);
}
</style> 