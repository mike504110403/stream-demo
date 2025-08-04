<template>
  <div class="llhls-player-container">
    <div class="player-header">
      <h3>{{ title }}</h3>
      <div class="player-status">
        <span :class="['status-indicator', status]">{{ statusText }}</span>
        <span v-if="latency" class="latency">延遲: {{ latency }}ms</span>
      </div>
    </div>

    <div class="video-container">
      <video
        ref="videoRef"
        class="video-player"
        controls
        autoplay
        muted
        @loadstart="onLoadStart"
        @loadedmetadata="onLoadedMetadata"
        @canplay="onCanPlay"
        @playing="onPlaying"
        @waiting="onWaiting"
        @error="onError"
      ></video>

      <div v-if="loading" class="loading-overlay">
        <div class="loading-spinner"></div>
        <p>正在載入直播流...</p>
      </div>

      <div v-if="error" class="error-overlay">
        <div class="error-icon">⚠️</div>
        <p>{{ error }}</p>
        <button @click="retry" class="retry-btn">重試</button>
      </div>
    </div>

    <div class="player-info">
      <div class="info-row">
        <span>解析度:</span>
        <span>{{ videoInfo.resolution }}</span>
      </div>
      <div class="info-row">
        <span>比特率:</span>
        <span>{{ videoInfo.bitrate }}</span>
      </div>
      <div class="info-row">
        <span>FPS:</span>
        <span>{{ videoInfo.fps }}</span>
      </div>
      <div class="info-row">
        <span>緩衝:</span>
        <span>{{ videoInfo.buffered }}</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed, watch } from 'vue'
import Hls from 'hls.js'

interface Props {
  streamUrl: string
  title?: string
  autoPlay?: boolean
  muted?: boolean
}

interface VideoInfo {
  resolution: string
  bitrate: string
  fps: string
  buffered: string
}

const props = withDefaults(defineProps<Props>(), {
  title: '直播',
  autoPlay: true,
  muted: true,
})

const videoRef = ref<HTMLVideoElement>()
const hls = ref<Hls | null>(null)
const loading = ref(false)
const error = ref('')
const status = ref<'idle' | 'loading' | 'playing' | 'error'>('idle')
const latency = ref(0)
const videoInfo = ref<VideoInfo>({
  resolution: '未知',
  bitrate: '未知',
  fps: '未知',
  buffered: '0s',
})

const statusText = computed(() => {
  switch (status.value) {
    case 'idle':
      return '待機'
    case 'loading':
      return '載入中'
    case 'playing':
      return '播放中'
    case 'error':
      return '錯誤'
    default:
      return '未知'
  }
})

let latencyTimer: number | null = null
let infoTimer: number | null = null

// 初始化 HLS
const initHLS = () => {
  if (!videoRef.value) return

  // 檢查瀏覽器支援
  if (!Hls.isSupported()) {
    error.value = '瀏覽器不支援 HLS'
    status.value = 'error'
    return
  }

  // 創建 HLS 實例
  hls.value = new Hls({
    // LL-HLS 配置
    lowLatencyMode: true,
    liveSyncDurationCount: 1, // 1個片段同步
    liveMaxLatencyDurationCount: 3, // 最大3個片段延遲
    liveDurationInfinity: true,
    enableSoftwareAES: true,

    // LL-HLS 特定配置
    enableDateRangeMetadataCues: true,
    enableEmsgMetadataCues: true,
    enableID3MetadataCues: true,
    enableWebVTT: true,
    enableIMSC1: true,
    enableCEA708Captions: true,

    // 性能優化
    maxBufferLength: 3,
    maxMaxBufferLength: 5,
    maxBufferSize: 2 * 1000 * 1000, // 2MB
    maxBufferHole: 0.1,

    // 錯誤處理
    maxFragLookUpTolerance: 0.25,
    maxStarvationDelay: 4,
    maxLoadingDelay: 4,

    // 自適應
    abrEwmaFastLive: 3,
    abrEwmaSlowLive: 9,
    abrEwmaFastVoD: 3,
    abrEwmaSlowVoD: 9,

    // 重試配置
    // maxRetries: 3,  // 暫時註釋掉不支援的屬性
    // retryDelayMs: 1000,  // 暫時註釋掉不支援的屬性
    // maxRetryDelayMs: 5000,  // 暫時註釋掉不支援的屬性
  })

  // 綁定事件
  hls.value.on(Hls.Events.MEDIA_ATTACHED, () => {
    console.log('HLS 媒體已附加')
    hls.value?.loadSource(props.streamUrl)
  })

  hls.value.on(Hls.Events.MANIFEST_LOADED, () => {
    console.log('HLS 清單已載入')
    loading.value = false
    status.value = 'playing'
  })

  hls.value.on(Hls.Events.LEVEL_LOADED, (_event, data) => {
    console.log('HLS 品質等級已載入:', data.level)
    updateVideoInfo()
  })

  hls.value.on(Hls.Events.ERROR, (_event, data) => {
    console.error('HLS 錯誤:', data)
    if (data.fatal) {
      error.value = `播放錯誤: ${data.details}`
      status.value = 'error'
    }
  })

  hls.value.on(Hls.Events.FRAG_LOADED, () => {
    // 計算延遲
    if (videoRef.value) {
      const currentTime = videoRef.value.currentTime
      const buffered = videoRef.value.buffered
      if (buffered.length > 0) {
        const bufferEnd = buffered.end(buffered.length - 1)
        latency.value = Math.round((bufferEnd - currentTime) * 1000)
      }
    }
  })

  // 附加到視頻元素
  hls.value.attachMedia(videoRef.value)
}

// 更新視頻信息
const updateVideoInfo = () => {
  if (!videoRef.value) return

  const video = videoRef.value

  // 解析度
  if (video.videoWidth && video.videoHeight) {
    videoInfo.value.resolution = `${video.videoWidth}x${video.videoHeight}`
  }

  // 比特率 (如果 HLS 提供)
  if (hls.value && hls.value.levels.length > 0) {
    const currentLevel = hls.value.levels[hls.value.currentLevel]
    if (currentLevel) {
      videoInfo.value.bitrate = `${Math.round(currentLevel.bitrate / 1000)}kbps`
    }
  }

  // FPS (估算)
  if ((video as any).webkitVideoDecodedByteCount !== undefined) {
    // 這是一個粗略的估算
    videoInfo.value.fps = '30'
  }

  // 緩衝狀態
  if (video.buffered.length > 0) {
    const buffered =
      video.buffered.end(video.buffered.length - 1) - video.currentTime
    videoInfo.value.buffered = `${buffered.toFixed(1)}s`
  }
}

// 重試播放
const retry = () => {
  error.value = ''
  status.value = 'idle'
  loading.value = true

  if (hls.value) {
    hls.value.destroy()
    hls.value = null
  }

  setTimeout(() => {
    initHLS()
  }, 1000)
}

// 視頻事件處理
const onLoadStart = () => {
  loading.value = true
  status.value = 'loading'
}

const onLoadedMetadata = () => {
  updateVideoInfo()
}

const onCanPlay = () => {
  loading.value = false
  status.value = 'playing'
}

const onPlaying = () => {
  status.value = 'playing'
  startLatencyMonitoring()
  startInfoMonitoring()
}

const onWaiting = () => {
  status.value = 'loading'
}

const onError = (e: Event) => {
  console.error('視頻錯誤:', e)
  error.value = '視頻播放錯誤'
  status.value = 'error'
}

// 開始延遲監控
const startLatencyMonitoring = () => {
  if (latencyTimer) clearInterval(latencyTimer)

  latencyTimer = window.setInterval(() => {
    if (videoRef.value && hls.value) {
      const currentTime = videoRef.value.currentTime
      const buffered = videoRef.value.buffered
      if (buffered.length > 0) {
        const bufferEnd = buffered.end(buffered.length - 1)
        latency.value = Math.round((bufferEnd - currentTime) * 1000)
      }
    }
  }, 1000)
}

// 開始信息監控
const startInfoMonitoring = () => {
  if (infoTimer) clearInterval(infoTimer)

  infoTimer = window.setInterval(() => {
    updateVideoInfo()
  }, 2000)
}

// 清理
const cleanup = () => {
  if (latencyTimer) {
    clearInterval(latencyTimer)
    latencyTimer = null
  }

  if (infoTimer) {
    clearInterval(infoTimer)
    infoTimer = null
  }

  if (hls.value) {
    hls.value.destroy()
    hls.value = null
  }
}

// 監聽 URL 變化
watch(
  () => props.streamUrl,
  newUrl => {
    if (newUrl && hls.value) {
      hls.value.loadSource(newUrl)
    }
  }
)

onMounted(() => {
  initHLS()
})

onUnmounted(() => {
  cleanup()
})
</script>

<style scoped>
.llhls-player-container {
  width: 100%;
  max-width: 1200px;
  margin: 0 auto;
  background: #000;
  border-radius: 8px;
  overflow: hidden;
}

.player-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  background: rgba(0, 0, 0, 0.8);
  color: white;
}

.player-header h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
}

.player-status {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 14px;
}

.status-indicator {
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;
}

.status-indicator.idle {
  background: #6c757d;
}

.status-indicator.loading {
  background: #ffc107;
  color: #000;
}

.status-indicator.playing {
  background: #28a745;
}

.status-indicator.error {
  background: #dc3545;
}

.latency {
  color: #17a2b8;
  font-family: monospace;
}

.video-container {
  position: relative;
  width: 100%;
  background: #000;
}

.video-player {
  width: 100%;
  height: auto;
  display: block;
}

.loading-overlay,
.error-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  background: rgba(0, 0, 0, 0.8);
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

.error-icon {
  font-size: 48px;
  margin-bottom: 16px;
}

.retry-btn {
  margin-top: 16px;
  padding: 8px 16px;
  background: #007bff;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
}

.retry-btn:hover {
  background: #0056b3;
}

.player-info {
  padding: 12px 16px;
  background: rgba(0, 0, 0, 0.8);
  color: white;
  font-size: 14px;
}

.info-row {
  display: flex;
  justify-content: space-between;
  margin-bottom: 4px;
}

.info-row:last-child {
  margin-bottom: 0;
}

.info-row span:first-child {
  color: #adb5bd;
}

.info-row span:last-child {
  font-family: monospace;
  color: #17a2b8;
}
</style>
