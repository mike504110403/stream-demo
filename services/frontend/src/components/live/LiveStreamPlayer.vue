<template>
  <div class="live-stream-player">
    <div class="player-container">
      <!-- 播放器 -->
      <video
        ref="videoPlayer"
        class="video-player"
        controls
        autoplay
        muted
        @loadstart="onLoadStart"
        @loadeddata="onLoadedData"
        @error="onError"
        @waiting="onWaiting"
        @playing="onPlaying"
        @pause="onPause"
        @ended="onEnded"
      >
        <source :src="streamUrl" type="application/x-mpegURL" />
        您的瀏覽器不支援 HLS 播放
      </video>

      <!-- 播放器控制欄 -->
      <div class="player-controls" v-if="showControls">
        <div class="control-left">
          <button @click="togglePlay" class="control-btn">
            <i :class="isPlaying ? 'fas fa-pause' : 'fas fa-play'"></i>
          </button>
          <div class="time-display">
            {{ formatTime(currentTime) }} / {{ formatTime(duration) }}
          </div>
        </div>

        <div class="control-right">
          <button @click="toggleMute" class="control-btn">
            <i :class="isMuted ? 'fas fa-volume-mute' : 'fas fa-volume-up'"></i>
          </button>
          <button @click="toggleFullscreen" class="control-btn">
            <i class="fas fa-expand"></i>
          </button>
        </div>
      </div>

      <!-- 載入狀態 -->
      <div v-if="loading" class="loading-overlay">
        <div class="loading-spinner"></div>
        <p>正在載入直播流...</p>
      </div>

      <!-- 錯誤狀態 -->
      <div v-if="error" class="error-overlay">
        <div class="error-content">
          <i class="fas fa-exclamation-triangle"></i>
          <h3>播放失敗</h3>
          <p>{{ errorMessage }}</p>
          <button @click="retry" class="retry-btn">重試</button>
        </div>
      </div>
    </div>

    <!-- 流資訊 -->
    <div class="stream-info">
      <h3>{{ streamInfo.title }}</h3>
      <p>{{ streamInfo.description }}</p>
      <div class="stream-stats">
        <span class="stat">
          <i class="fas fa-eye"></i>
          {{ viewerCount }} 觀眾
        </span>
        <span class="stat">
          <i class="fas fa-signal"></i>
          {{ streamStatus }}
        </span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch } from "vue";
import { publicStreamApi } from "@/api/public-stream";
import type { PublicStreamInfo } from "@/types/public-stream";

interface Props {
  streamName: string;
  autoPlay?: boolean;
  showControls?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  autoPlay: true,
  showControls: true,
});

// 響應式數據
const videoPlayer = ref<HTMLVideoElement>();
const streamUrl = ref("");
const streamInfo = ref<PublicStreamInfo>({} as PublicStreamInfo);
const loading = ref(true);
const error = ref(false);
const errorMessage = ref("");
const isPlaying = ref(false);
const isMuted = ref(false);
const currentTime = ref(0);
const duration = ref(0);
const viewerCount = ref(0);
const streamStatus = ref("載入中...");

// 定時器
let timeUpdateTimer: number | null = null;
let statsTimer: number | null = null;

// 獲取流資訊
const fetchStreamInfo = async () => {
  try {
    const response = await publicStreamApi.getStreamInfo(props.streamName);
    streamInfo.value = response;
    streamStatus.value = response.status;
  } catch (err) {
    console.error("獲取流資訊失敗:", err);
  }
};

// 獲取播放 URL
const fetchStreamURL = async () => {
  try {
    loading.value = true;
    error.value = false;

    const response = await publicStreamApi.getStreamURLs(props.streamName);
    streamUrl.value = response.urls.hls;

    // 更新觀眾數（從流資訊中獲取）
    viewerCount.value = streamInfo.value.viewer_count || 0;
  } catch (err: any) {
    error.value = true;
    errorMessage.value = err.message || "獲取播放地址失敗";
    console.error("獲取播放 URL 失敗:", err);
  } finally {
    loading.value = false;
  }
};

// 播放器事件處理
const onLoadStart = () => {
  loading.value = true;
  error.value = false;
};

const onLoadedData = () => {
  loading.value = false;
  if (props.autoPlay && videoPlayer.value) {
    videoPlayer.value.play().catch((err) => {
      console.error("自動播放失敗:", err);
    });
  }
};

const onError = (event: Event) => {
  loading.value = false;
  error.value = true;
  errorMessage.value = "視頻載入失敗，請檢查網路連接";
  console.error("視頻播放錯誤:", event);
};

const onWaiting = () => {
  loading.value = true;
};

const onPlaying = () => {
  loading.value = false;
  isPlaying.value = true;
};

const onPause = () => {
  isPlaying.value = false;
};

const onEnded = () => {
  isPlaying.value = false;
};

// 控制功能
const togglePlay = () => {
  if (videoPlayer.value) {
    if (isPlaying.value) {
      videoPlayer.value.pause();
    } else {
      videoPlayer.value.play();
    }
  }
};

const toggleMute = () => {
  if (videoPlayer.value) {
    videoPlayer.value.muted = !videoPlayer.value.muted;
    isMuted.value = videoPlayer.value.muted;
  }
};

const toggleFullscreen = () => {
  if (videoPlayer.value) {
    if (document.fullscreenElement) {
      document.exitFullscreen();
    } else {
      videoPlayer.value.requestFullscreen();
    }
  }
};

const retry = () => {
  fetchStreamURL();
};

// 時間格式化
const formatTime = (seconds: number): string => {
  const mins = Math.floor(seconds / 60);
  const secs = Math.floor(seconds % 60);
  return `${mins.toString().padStart(2, "0")}:${secs.toString().padStart(2, "0")}`;
};

// 更新時間
const updateTime = () => {
  if (videoPlayer.value) {
    currentTime.value = videoPlayer.value.currentTime;
    duration.value = videoPlayer.value.duration;
  }
};

// 更新統計資訊
const updateStats = async () => {
  try {
    const response = await publicStreamApi.getStreamStats(props.streamName);
    viewerCount.value = response.data.viewer_count || 0;
  } catch (err) {
    console.error("更新統計資訊失敗:", err);
  }
};

// 生命週期
onMounted(async () => {
  await fetchStreamInfo();
  await fetchStreamURL();

  // 啟動定時器
  timeUpdateTimer = window.setInterval(updateTime, 1000);
  statsTimer = window.setInterval(updateStats, 5000);
});

onUnmounted(() => {
  if (timeUpdateTimer) {
    clearInterval(timeUpdateTimer);
  }
  if (statsTimer) {
    clearInterval(statsTimer);
  }
});

// 監聽流名稱變化
watch(
  () => props.streamName,
  async () => {
    await fetchStreamInfo();
    await fetchStreamURL();
  },
);
</script>

<style scoped>
.live-stream-player {
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
}

.video-player {
  width: 100%;
  height: auto;
  min-height: 400px;
  display: block;
}

.player-controls {
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  background: linear-gradient(transparent, rgba(0, 0, 0, 0.7));
  padding: 20px;
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
  gap: 10px;
}

.control-btn {
  background: rgba(255, 255, 255, 0.2);
  border: none;
  color: white;
  padding: 8px 12px;
  border-radius: 4px;
  cursor: pointer;
  transition: background 0.2s;
}

.control-btn:hover {
  background: rgba(255, 255, 255, 0.3);
}

.time-display {
  color: white;
  font-size: 14px;
  font-family: monospace;
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
  flex-direction: column;
  justify-content: center;
  align-items: center;
  color: white;
}

.loading-spinner {
  width: 40px;
  height: 40px;
  border: 4px solid rgba(255, 255, 255, 0.3);
  border-top: 4px solid white;
  border-radius: 50%;
  animation: spin 1s linear infinite;
  margin-bottom: 16px;
}

@keyframes spin {
  0% {
    transform: rotate(0deg);
  }
  100% {
    transform: rotate(360deg);
  }
}

.error-content {
  text-align: center;
}

.error-content i {
  font-size: 48px;
  color: #ff6b6b;
  margin-bottom: 16px;
}

.error-content h3 {
  margin: 0 0 8px 0;
  font-size: 18px;
}

.error-content p {
  margin: 0 0 16px 0;
  opacity: 0.8;
}

.retry-btn {
  background: #007bff;
  color: white;
  border: none;
  padding: 8px 16px;
  border-radius: 4px;
  cursor: pointer;
  transition: background 0.2s;
}

.retry-btn:hover {
  background: #0056b3;
}

.stream-info {
  margin-top: 16px;
  padding: 16px;
  background: #f8f9fa;
  border-radius: 8px;
}

.stream-info h3 {
  margin: 0 0 8px 0;
  font-size: 18px;
  color: #333;
}

.stream-info p {
  margin: 0 0 12px 0;
  color: #666;
  line-height: 1.5;
}

.stream-stats {
  display: flex;
  gap: 16px;
}

.stat {
  display: flex;
  align-items: center;
  gap: 4px;
  color: #666;
  font-size: 14px;
}

.stat i {
  color: #007bff;
}
</style>
