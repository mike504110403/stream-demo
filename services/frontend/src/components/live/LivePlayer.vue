<template>
  <div class="live-player">
    <div class="player-container">
      <!-- 載入狀態 -->
      <div v-if="loading" class="loading-overlay">
        <div class="loading-content">
          <el-icon class="loading-icon" :size="32">
            <Loading />
          </el-icon>
          <p>正在載入直播流...</p>
        </div>
      </div>

      <!-- 錯誤狀態 -->
      <div v-if="error" class="error-overlay">
        <div class="error-content">
          <el-icon class="error-icon" :size="32">
            <Warning />
          </el-icon>
          <h3>直播載入失敗</h3>
          <p>{{ error }}</p>
          <el-button type="primary" @click="retryLoad"> 重新載入 </el-button>
        </div>
      </div>

      <!-- 直播狀態指示器 -->
      <div v-if="isLive" class="live-indicator">
        <div class="live-dot"></div>
        <span>直播中</span>
        <span v-if="viewerCount > 0" class="viewer-count">
          {{ formatViewerCount(viewerCount) }} 觀看
        </span>
      </div>

      <!-- 影片播放器 -->
      <video
        ref="videoElement"
        class="video-element"
        controls
        autoplay
        muted
        @loadstart="handleLoadStart"
        @canplay="handleCanPlay"
        @error="handleError"
        @waiting="handleWaiting"
        @playing="handlePlaying"
        @pause="handlePause"
        @ended="handleEnded"
      >
        您的瀏覽器不支援影片播放
      </video>

      <!-- 播放控制欄 -->
      <div class="player-controls">
        <div class="control-left">
          <el-button
            :icon="isPlaying ? 'Pause' : 'VideoPlay'"
            circle
            @click="togglePlay"
            :disabled="!canPlay"
          />
          <span class="time-display">
            {{ formatTime(currentTime) }} / {{ formatTime(duration) }}
          </span>
        </div>

        <div class="control-right">
          <el-button
            icon="Refresh"
            circle
            @click="retryLoad"
            :disabled="loading"
          />
          <el-button icon="FullScreen" circle @click="toggleFullscreen" />
        </div>
      </div>
    </div>

    <!-- 直播資訊 -->
    <div v-if="liveInfo" class="live-info">
      <h2>{{ liveInfo.title }}</h2>
      <p v-if="liveInfo.description">{{ liveInfo.description }}</p>
      <div class="live-meta">
        <span class="meta-item">
          <el-icon><User /></el-icon>
          {{ liveInfo.user?.username || "未知用戶" }}
        </span>
        <span class="meta-item">
          <el-icon><Clock /></el-icon>
          {{ formatStartTime(liveInfo.start_time) }}
        </span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch } from "vue";
import { ElMessage } from "element-plus";
import { Loading, Warning, User, Clock } from "@element-plus/icons-vue";
import Hls from "hls.js";
import type { Live } from "@/types";

interface Props {
  streamUrl?: string;
  liveInfo?: Live;
  autoPlay?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  autoPlay: true,
});

// 響應式數據
const videoElement = ref<HTMLVideoElement>();
const hls = ref<Hls | null>(null);
const loading = ref(false);
const error = ref<string>("");
const isLive = ref(false);
const isPlaying = ref(false);
const canPlay = ref(false);
const currentTime = ref(0);
const duration = ref(0);
const viewerCount = ref(0);

// 定時器
const timeUpdateTimer = ref<number | null>(null);
const retryTimer = ref<number | null>(null);

// 初始化播放器
const initPlayer = () => {
  if (!videoElement.value || !props.streamUrl) {
    return;
  }

  loading.value = true;
  error.value = "";

  console.log("初始化直播播放器:", props.streamUrl);

  // 清理之前的 HLS 實例
  if (hls.value) {
    hls.value.destroy();
    hls.value = null;
  }

  // 檢查是否為 HLS 串流
  if (props.streamUrl.includes(".m3u8")) {
    setupHLSPlayer();
  } else {
    setupDirectPlayer();
  }
};

// 設置 HLS 播放器
const setupHLSPlayer = () => {
  if (!videoElement.value || !props.streamUrl) return;

  if (Hls.isSupported()) {
    hls.value = new Hls({
      debug: false,
      enableWorker: true,
      lowLatencyMode: true,
      backBufferLength: 90,
      maxBufferLength: 30,
      maxMaxBufferLength: 600,
      maxBufferSize: 60 * 1000 * 1000,
      maxBufferHole: 0.5,
      highBufferWatchdogPeriod: 2,
      nudgeOffset: 0.2,
      nudgeMaxRetry: 5,
      maxFragLookUpTolerance: 0.25,
      liveSyncDurationCount: 3,
      liveMaxLatencyDurationCount: 10,
    });

    hls.value.loadSource(props.streamUrl);
    hls.value.attachMedia(videoElement.value);

    hls.value.on(Hls.Events.MANIFEST_PARSED, () => {
      console.log("HLS 清單解析完成");
      loading.value = false;
      isLive.value = true;

      if (props.autoPlay) {
        videoElement.value?.play().catch((e) => {
          console.error("自動播放失敗:", e);
        });
      }
    });

    hls.value.on(Hls.Events.ERROR, (_event, data) => {
      console.error("HLS 錯誤:", data);
      handleHlsError(data);
    });

    hls.value.on(Hls.Events.FRAG_LOADED, () => {
      // 直播流正常載入
      isLive.value = true;
    });
  } else {
    // 瀏覽器原生支援 HLS
    console.log("使用瀏覽器原生 HLS 支援");
    videoElement.value.src = props.streamUrl;
    loading.value = false;
    isLive.value = true;
  }
};

// 設置直接播放器（非 HLS）
const setupDirectPlayer = () => {
  if (!videoElement.value || !props.streamUrl) return;

  console.log("設置直接播放器");
  videoElement.value.src = props.streamUrl;
  loading.value = false;
  isLive.value = true;
};

// 處理 HLS 錯誤
const handleHlsError = (data: any) => {
  if (data.fatal) {
    switch (data.type) {
      case Hls.ErrorTypes.NETWORK_ERROR:
        error.value = "網路錯誤，無法載入直播流";
        break;
      case Hls.ErrorTypes.MEDIA_ERROR:
        error.value = "媒體錯誤，直播流格式不支援";
        break;
      default:
        error.value = "播放器錯誤，請重新載入";
    }
    loading.value = false;
  }
};

// 事件處理
const handleLoadStart = () => {
  console.log("開始載入直播流");
  loading.value = true;
  error.value = "";
};

const handleCanPlay = () => {
  console.log("直播流可以播放");
  loading.value = false;
  canPlay.value = true;
  isLive.value = true;
};

const handleError = (event: Event) => {
  const video = event.target as HTMLVideoElement;
  const videoError = video.error;

  console.error("影片播放錯誤:", videoError);

  if (videoError) {
    switch (videoError.code) {
      case videoError.MEDIA_ERR_ABORTED:
        error.value = "播放被中止";
        break;
      case videoError.MEDIA_ERR_NETWORK:
        error.value = "網路錯誤，無法載入直播流";
        break;
      case videoError.MEDIA_ERR_DECODE:
        error.value = "解碼錯誤，直播流格式不支援";
        break;
      case videoError.MEDIA_ERR_SRC_NOT_SUPPORTED:
        error.value = "不支援的直播流格式";
        break;
      default:
        error.value = "播放器錯誤";
    }
  } else {
    error.value = "直播流載入失敗";
  }

  loading.value = false;
};

const handleWaiting = () => {
  console.log("等待直播流載入");
  loading.value = true;
};

const handlePlaying = () => {
  console.log("直播開始播放");
  isPlaying.value = true;
  loading.value = false;
  startTimeUpdate();
};

const handlePause = () => {
  console.log("直播暫停");
  isPlaying.value = false;
  stopTimeUpdate();
};

const handleEnded = () => {
  console.log("直播結束");
  isPlaying.value = false;
  stopTimeUpdate();
};

// 播放控制
const togglePlay = () => {
  if (!videoElement.value) return;

  if (isPlaying.value) {
    videoElement.value.pause();
  } else {
    videoElement.value.play().catch((e) => {
      console.error("播放失敗:", e);
      ElMessage.error("播放失敗");
    });
  }
};

const retryLoad = () => {
  console.log("重新載入直播流");
  initPlayer();
};

const toggleFullscreen = () => {
  if (!videoElement.value) return;

  if (document.fullscreenElement) {
    document.exitFullscreen();
  } else {
    videoElement.value.requestFullscreen().catch((e) => {
      console.error("全螢幕切換失敗:", e);
    });
  }
};

// 時間更新
const startTimeUpdate = () => {
  if (timeUpdateTimer.value) {
    clearInterval(timeUpdateTimer.value);
  }

  timeUpdateTimer.value = window.setInterval(() => {
    if (videoElement.value) {
      currentTime.value = videoElement.value.currentTime;
      duration.value = videoElement.value.duration;
    }
  }, 1000);
};

const stopTimeUpdate = () => {
  if (timeUpdateTimer.value) {
    clearInterval(timeUpdateTimer.value);
    timeUpdateTimer.value = null;
  }
};

// 工具函數
const formatTime = (seconds: number): string => {
  if (!seconds || isNaN(seconds)) return "00:00";

  const hours = Math.floor(seconds / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);
  const secs = Math.floor(seconds % 60);

  if (hours > 0) {
    return `${hours.toString().padStart(2, "0")}:${minutes.toString().padStart(2, "0")}:${secs.toString().padStart(2, "0")}`;
  }
  return `${minutes.toString().padStart(2, "0")}:${secs.toString().padStart(2, "0")}`;
};

const formatViewerCount = (count: number): string => {
  if (count >= 10000) {
    return `${(count / 10000).toFixed(1)}萬`;
  }
  return count.toString();
};

const formatStartTime = (dateString: string): string => {
  const date = new Date(dateString);
  return date.toLocaleString("zh-TW");
};

// 監聽串流 URL 變化
watch(
  () => props.streamUrl,
  (newUrl) => {
    if (newUrl) {
      initPlayer();
    }
  },
);

// 生命週期
onMounted(() => {
  if (props.streamUrl) {
    initPlayer();
  }
});

onUnmounted(() => {
  // 清理資源
  if (hls.value) {
    hls.value.destroy();
    hls.value = null;
  }

  stopTimeUpdate();

  if (retryTimer.value) {
    clearTimeout(retryTimer.value);
  }
});
</script>

<style scoped>
.live-player {
  width: 100%;
  max-width: 1200px;
  margin: 0 auto;
}

.player-container {
  position: relative;
  width: 100%;
  background: #000;
  border-radius: 8px;
  overflow: hidden;
  aspect-ratio: 16 / 9;
}

.loading-overlay,
.error-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.8);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 10;
}

.loading-content,
.error-content {
  text-align: center;
  color: white;
}

.loading-icon {
  animation: spin 1s linear infinite;
  margin-bottom: 16px;
}

.error-icon {
  color: #f56c6c;
  margin-bottom: 16px;
}

.live-indicator {
  position: absolute;
  top: 16px;
  left: 16px;
  background: rgba(0, 0, 0, 0.7);
  color: white;
  padding: 8px 12px;
  border-radius: 20px;
  display: flex;
  align-items: center;
  gap: 8px;
  z-index: 5;
  font-size: 14px;
}

.live-dot {
  width: 8px;
  height: 8px;
  background: #f56c6c;
  border-radius: 50%;
  animation: pulse 2s infinite;
}

.viewer-count {
  color: #909399;
}

.video-element {
  width: 100%;
  height: 100%;
  object-fit: contain;
}

.player-controls {
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  background: linear-gradient(transparent, rgba(0, 0, 0, 0.8));
  padding: 20px 16px 16px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  opacity: 0;
  transition: opacity 0.3s;
}

.player-container:hover .player-controls {
  opacity: 1;
}

.control-left,
.control-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.time-display {
  color: white;
  font-size: 14px;
  font-family: monospace;
}

.live-info {
  margin-top: 16px;
  padding: 16px;
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.live-info h2 {
  margin: 0 0 8px 0;
  color: #333;
  font-size: 20px;
}

.live-info p {
  margin: 0 0 16px 0;
  color: #666;
  line-height: 1.5;
}

.live-meta {
  display: flex;
  gap: 16px;
}

.meta-item {
  display: flex;
  align-items: center;
  gap: 4px;
  color: #666;
  font-size: 14px;
}

@keyframes spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

@keyframes pulse {
  0%,
  100% {
    opacity: 1;
  }
  50% {
    opacity: 0.5;
  }
}

/* 響應式設計 */
@media (max-width: 768px) {
  .player-container {
    aspect-ratio: 16 / 10;
  }

  .live-indicator {
    top: 8px;
    left: 8px;
    padding: 6px 10px;
    font-size: 12px;
  }

  .player-controls {
    padding: 16px 12px 12px;
  }

  .control-left,
  .control-right {
    gap: 8px;
  }

  .time-display {
    font-size: 12px;
  }
}
</style>
