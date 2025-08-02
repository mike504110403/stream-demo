<template>
  <div class="video-detail">
    <div class="page-header">
      <h1>影片詳情</h1>
      <el-button @click="$router.back()">返回</el-button>
    </div>

    <div v-if="loading" class="loading">
      <el-skeleton :rows="5" animated />
    </div>

    <div v-else-if="video" class="video-content">
      <el-card class="video-card">
        <div class="video-player">
          <!-- 優先使用 HLS，然後是原始影片 -->
          <div v-if="getVideoURL()" class="video-container">
            <!-- 品質選擇器 -->
            <div
              v-if="video.qualities && video.qualities.length > 0"
              class="quality-selector"
            >
              <el-select
                v-model="selectedQuality"
                @change="changeQuality"
                size="small"
              >
                <el-option label="自動 (最佳品質)" :value="0" />
                <el-option
                  v-for="quality in video.qualities"
                  :key="quality.id"
                  :label="`${quality.quality} (${quality.width}x${quality.height})`"
                  :value="quality.id"
                />
              </el-select>
              <!-- 自動品質切換狀態 -->
              <div
                v-if="selectedQuality === 0 && autoQualityInfo.show"
                class="auto-quality-info"
              >
                <el-tag
                  :type="autoQualityInfo.type"
                  size="small"
                  @click="showAutoQualityDetails = !showAutoQualityDetails"
                >
                  {{ autoQualityInfo.message }}
                </el-tag>
              </div>
            </div>

            <video
              ref="videoElement"
              controls
              width="100%"
              height="auto"
              @loadstart="handleVideoLoad"
              @error="handleVideoError"
              @canplay="handleVideoCanPlay"
              @loadedmetadata="handleVideoMetadata"
            >
              您的瀏覽器不支援影片播放
            </video>
            <div v-if="showDebug" class="debug-info">
              <h4>🔧 調試資訊</h4>
              <p>原始 URL: {{ video.original_url || '無' }}</p>
              <p>MP4 URL: {{ video.mp4_url || '無' }}</p>
              <p>HLS URL: {{ video.hls_master_url || '無' }}</p>
              <p>縮圖 URL: {{ video.thumbnail_url || '無' }}</p>
              <p>狀態: {{ video.status }}</p>
              <p>處理進度: {{ video.processing_progress }}%</p>
              <p>目前播放 URL: {{ getVideoURL() || '無' }}</p>
              <p>格式: {{ getVideoFormat() }}</p>
              <p v-if="videoError" class="error">錯誤: {{ videoError }}</p>
            </div>
            <div v-if="videoError" class="video-error">
              <p>⚠️ 播放錯誤: {{ videoError }}</p>
              <p>影片格式: {{ getVideoFormat() }}</p>
              <div v-if="getVideoFormat() === 'MOV'" class="format-suggestion">
                <p>
                  💡 <strong>建議</strong>: MOV 格式在網頁瀏覽器中的支援有限
                </p>
                <p>🔧 <strong>解決方案</strong>:</p>
                <ul>
                  <li>使用 MP4 格式上傳 (建議)</li>
                  <li>
                    使用 FFmpeg 轉換:
                    <code>ffmpeg -i input.mov output.mp4</code>
                  </li>
                  <li>或使用線上轉換工具</li>
                </ul>
              </div>
              <el-button size="small" @click="showDebug = !showDebug">
                {{ showDebug ? '隱藏' : '顯示' }}調試信息
              </el-button>
              <el-button
                size="small"
                type="primary"
                @click="downloadVideo"
                v-if="getVideoURL()"
              >
                下載影片
              </el-button>
            </div>
          </div>
          <div v-else class="placeholder-player">
            <div class="placeholder-icon">🎬</div>
            <p v-if="video.status === 'processing'">影片處理中，請稍後...</p>
            <p v-else-if="video.status === 'failed'">影片處理失敗</p>
            <p v-else>影片暫時無法播放</p>
            <div
              class="debug-info"
              style="margin-top: 16px; font-size: 12px; color: #999"
            >
              <p>狀態: {{ video.status }}</p>
              <p>原始 URL: {{ video.original_url || '無' }}</p>
              <p>HLS URL: {{ video.hls_master_url || '無' }}</p>
            </div>
          </div>
        </div>

        <div class="video-info">
          <h2>{{ video.title }}</h2>
          <p class="video-description">{{ video.description || '暫無描述' }}</p>

          <div class="video-meta">
            <div class="meta-item">
              <span class="meta-label">狀態：</span>
              <el-tag :type="getStatusType(video.status)">
                {{ getStatusText(video.status) }}
              </el-tag>
            </div>
            <div class="meta-item">
              <span class="meta-label">觀看數：</span>
              <span>{{ video.views }}</span>
            </div>
            <div class="meta-item">
              <span class="meta-label">按讚數：</span>
              <span>{{ video.likes }}</span>
            </div>
            <div class="meta-item">
              <span class="meta-label">上傳時間：</span>
              <span>{{ formatDate(video.created_at) }}</span>
            </div>
          </div>

          <div class="video-actions">
            <el-button type="primary" @click="handleLike" :loading="liking">
              👍 按讚 ({{ video.likes }})
            </el-button>
            <el-button @click="$router.push(`/videos/${video.id}/edit`)">
              編輯
            </el-button>
          </div>
        </div>
      </el-card>
    </div>

    <div v-else class="error-state">
      <div class="error-icon">❌</div>
      <h3>影片不存在</h3>
      <p>找不到指定的影片，可能已被刪除</p>
      <el-button type="primary" @click="$router.push('/videos')">
        返回影片列表
      </el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, nextTick } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { getVideo, likeVideo } from '@/api/video'
import type { Video } from '@/types'
import Hls from 'hls.js'

const route = useRoute()

const loading = ref(true)
const liking = ref(false)
const video = ref<Video | null>(null)
const showDebug = ref(false) // 關閉調試模式
const videoError = ref<string>('') // 影片錯誤信息
const videoElement = ref<HTMLVideoElement>()
const hls = ref<Hls | null>(null)
const selectedQuality = ref<number>(0)

// 自動品質切換相關變數
const autoQualityInfo = ref({
  show: false,
  type: 'info' as 'info' | 'warning' | 'success',
  message: '',
  currentQuality: '',
  reason: '',
})
const showAutoQualityDetails = ref(false)
const autoQualityTimer = ref<ReturnType<typeof setInterval> | null>(null)
const bufferingCount = ref(0)

const loadVideo = async () => {
  const videoId = Number(route.params.id)
  if (!videoId) return

  loading.value = true
  try {
    const response = await getVideo(videoId)
    console.log('影片 API 響應:', response) // 調試用
    // request.ts 攔截器已經提取了 data，所以 response 就是實際數據
    video.value = response
    console.log('影片數據:', video.value) // 調試用

    // 等待 DOM 更新後設置影片源
    await nextTick()

    // 確保 videoElement 準備好後再設置影片源
    const setupVideoWithRetry = () => {
      if (videoElement.value) {
        console.log('videoElement 已準備好，設置影片源')
        setupVideoSource()
      } else {
        console.log('videoElement 尚未準備好，等待...')
        setTimeout(setupVideoWithRetry, 100)
      }
    }

    setupVideoWithRetry()
  } catch (error) {
    console.error('載入影片失敗:', error)
    ElMessage.error('載入影片失敗')
  } finally {
    loading.value = false
  }
}

// 獲取播放 URL（優先使用轉碼後的 MP4，然後是 HLS，最後是原始 URL）
const getVideoURL = () => {
  if (!video.value) return null

  // 優先使用轉碼後的 MP4（最佳相容性）
  if (video.value.mp4_url) {
    return video.value.mp4_url
  }

  // 然後使用 HLS（適合串流）
  if (video.value.hls_master_url) {
    return video.value.hls_master_url
  }

  // 最後使用原始影片 URL
  if (video.value.original_url) {
    return video.value.original_url
  }

  return null
}

// 設置影片源（支援 HLS）
const setupVideoSource = () => {
  if (!videoElement.value || !video.value) {
    console.log('影片元素或數據未準備好')
    return
  }

  const url = getVideoURL()
  if (!url) {
    console.log('無法獲取影片 URL')
    return
  }

  console.log('設置影片源:', url)

  // 設置影片源
  if (url.includes('.m3u8')) {
    setupHLSPlayer(url)
  } else {
    setupMP4Player(url)
  }

  // 如果選擇自動品質，啟動監控
  if (selectedQuality.value === 0) {
    startAutoQualityMonitoring()
  }
}

// 智能自動品質切換邏輯
const startAutoQualityMonitoring = () => {
  if (selectedQuality.value !== 0 || !video.value?.qualities) {
    return
  }

  console.log('開始自動品質監控')

  // 清理之前的定時器
  if (autoQualityTimer.value) {
    clearInterval(autoQualityTimer.value)
  }

  // 每3秒檢查一次播放狀態
  autoQualityTimer.value = setInterval(() => {
    if (!videoElement.value || selectedQuality.value !== 0) {
      return
    }

    const video = videoElement.value
    const currentTime = video.currentTime
    const buffered = video.buffered

    // 檢查緩衝區狀態
    let bufferedEnd = 0
    if (buffered.length > 0) {
      bufferedEnd = buffered.end(buffered.length - 1)
    }

    const bufferAhead = bufferedEnd - currentTime
    const isBuffering = video.readyState < 3 // HAVE_FUTURE_DATA

    console.log('自動品質檢查:', {
      currentTime,
      bufferedEnd,
      bufferAhead,
      isBuffering,
      readyState: video.readyState,
    })

    // 如果緩衝區不足或正在緩衝，考慮降低品質
    if (bufferAhead < 5 || isBuffering) {
      bufferingCount.value++

      if (bufferingCount.value >= 2) {
        // 連續2次檢測到問題
        console.log('檢測到播放問題，考慮降低品質')
        autoSwitchToLowerQuality()
        bufferingCount.value = 0
      }
    } else {
      // 播放正常，重置計數器
      bufferingCount.value = 0
    }
  }, 3000)
}

// 自動切換到較低品質
const autoSwitchToLowerQuality = () => {
  if (!video.value?.qualities || selectedQuality.value !== 0) {
    return
  }

  const qualities = [...video.value.qualities].sort((a, b) => {
    // 按解析度排序（從高到低）
    const aHeight = parseInt(a.quality.replace(/\D/g, ''))
    const bHeight = parseInt(b.quality.replace(/\D/g, ''))
    return bHeight - aHeight
  })

  // 找到當前播放的品質
  const currentURL = getVideoURL()
  let currentQualityIndex = -1

  for (let i = 0; i < qualities.length; i++) {
    if (qualities[i].file_url === currentURL) {
      currentQualityIndex = i
      break
    }
  }

  // 如果找不到當前品質或已經是最低品質，嘗試使用 MP4
  if (
    currentQualityIndex === -1 ||
    currentQualityIndex === qualities.length - 1
  ) {
    if (video.value.mp4_url && video.value.mp4_url !== currentURL) {
      console.log('自動切換到 MP4 格式')
      switchToQuality('mp4', '網路較慢，已切換到 MP4 格式')
      return
    }
  }

  // 切換到較低品質
  if (currentQualityIndex > 0) {
    const lowerQuality = qualities[currentQualityIndex - 1]
    console.log(`自動切換到較低品質: ${lowerQuality.quality}`)
    switchToQuality(
      lowerQuality.id,
      `網路較慢，已自動切換到 ${lowerQuality.quality}`
    )
  }
}

// 切換到指定品質
const switchToQuality = (qualityId: number | string, reason: string) => {
  if (!video.value) return

  let targetURL = ''
  let qualityName = ''

  if (qualityId === 'mp4') {
    targetURL = video.value.mp4_url || ''
    qualityName = 'MP4'
  } else {
    const quality = video.value.qualities?.find(q => q.id === qualityId)
    if (quality) {
      targetURL = quality.file_url
      qualityName = quality.quality
    }
  }

  if (!targetURL) {
    console.log('無法找到目標品質的 URL')
    return
  }

  // 更新自動品質信息
  autoQualityInfo.value = {
    show: true,
    type: 'warning',
    message: `已切換到 ${qualityName}`,
    currentQuality: qualityName,
    reason: reason,
  }

  // 3秒後隱藏通知
  setTimeout(() => {
    autoQualityInfo.value.show = false
  }, 3000)

  // 設置新的影片源
  if (targetURL.includes('.m3u8')) {
    setupHLSPlayer(targetURL)
  } else {
    setupMP4Player(targetURL)
  }
}

// 設置 HLS 播放器
const setupHLSPlayer = (url: string) => {
  if (!videoElement.value) return

  console.log('設置 HLS 播放器:', url)

  // 清理之前的 HLS 實例
  if (hls.value) {
    hls.value.destroy()
    hls.value = null
  }

  if (Hls.isSupported()) {
    hls.value = new Hls({
      debug: true,
      enableWorker: true,
      lowLatencyMode: true,
    })

    hls.value.loadSource(url)
    hls.value.attachMedia(videoElement.value)

    hls.value.on(Hls.Events.MANIFEST_PARSED, () => {
      console.log('HLS 清單解析完成，開始播放')
      videoElement.value?.play().catch(e => {
        console.error('HLS 自動播放失敗:', e)
      })
    })

    hls.value.on(Hls.Events.ERROR, (_event, data) => {
      console.error('HLS 錯誤:', data)
      if (data.fatal) {
        videoError.value = `HLS 播放錯誤: ${data.details}`
      }
    })

    // 添加備用的自動播放邏輯
    videoElement.value.oncanplay = () => {
      console.log('HLS 可以播放')
      if (videoElement.value?.paused) {
        videoElement.value.play().catch(e => {
          console.error('HLS canplay 自動播放失敗:', e)
        })
      }
    }
  } else {
    console.log('瀏覽器原生支援 HLS')
    videoElement.value.src = url

    // 為原生 HLS 添加自動播放
    videoElement.value.oncanplay = () => {
      console.log('原生 HLS 可以播放')
      if (videoElement.value?.paused) {
        videoElement.value.play().catch(e => {
          console.error('原生 HLS 自動播放失敗:', e)
        })
      }
    }
  }
}

// 設置 MP4 播放器
const setupMP4Player = (url: string) => {
  if (!videoElement.value) return

  console.log('設置 MP4 播放器:', url)

  // 清理之前的事件監聽器
  videoElement.value.onloadeddata = null
  videoElement.value.oncanplay = null
  videoElement.value.onloadedmetadata = null

  videoElement.value.src = url

  // 添加載入完成事件
  videoElement.value.onloadeddata = () => {
    console.log('MP4 載入完成，開始播放')
    videoElement.value?.play().catch(e => {
      console.error('MP4 自動播放失敗:', e)
    })
  }

  // 添加可以播放事件（備用）
  videoElement.value.oncanplay = () => {
    console.log('MP4 可以播放')
    if (videoElement.value?.paused) {
      videoElement.value.play().catch(e => {
        console.error('MP4 canplay 自動播放失敗:', e)
      })
    }
  }

  // 添加元數據載入事件（備用）
  videoElement.value.onloadedmetadata = () => {
    console.log('MP4 元數據載入完成')
    if (videoElement.value?.paused) {
      videoElement.value.play().catch(e => {
        console.error('MP4 metadata 自動播放失敗:', e)
      })
    }
  }
}

// 切換影片品質
const changeQuality = () => {
  console.log('切換品質到:', selectedQuality.value)

  // 停止自動品質監控
  if (autoQualityTimer.value) {
    clearInterval(autoQualityTimer.value)
    autoQualityTimer.value = null
  }

  // 隱藏自動品質信息
  autoQualityInfo.value.show = false

  if (!video.value?.qualities || selectedQuality.value === 0) {
    console.log('使用預設品質')
    // 使用預設品質（MP4 或 HLS 主播放列表）
    setupVideoSource()
    return
  }

  const quality = video.value.qualities.find(
    q => q.id === selectedQuality.value
  )
  if (!quality) {
    console.log('找不到指定品質')
    return
  }

  console.log('切換到品質:', quality.quality, 'URL:', quality.file_url)

  // 設置新的品質 URL
  if (quality.file_url.includes('.m3u8')) {
    setupHLSPlayer(quality.file_url)
  } else {
    setupMP4Player(quality.file_url)
  }
}

// 獲取影片格式
const getVideoFormat = () => {
  if (!video.value?.original_url) return '未知'
  const extension = video.value.original_url.split('.').pop()?.toUpperCase()
  return extension || '未知'
}

// 處理影片載入
const handleVideoLoad = () => {
  console.log('影片開始載入:', getVideoURL())
  videoError.value = ''
}

// 處理影片錯誤
const handleVideoError = (event: Event) => {
  const videoElement = event.target as HTMLVideoElement
  const error = videoElement.error
  let errorMessage = '未知錯誤'

  if (error) {
    switch (error.code) {
      case error.MEDIA_ERR_ABORTED:
        errorMessage = '播放被中止'
        break
      case error.MEDIA_ERR_NETWORK:
        errorMessage = '網路錯誤'
        break
      case error.MEDIA_ERR_DECODE:
        errorMessage = '解碼錯誤 - 可能是不支援的影片格式'
        break
      case error.MEDIA_ERR_SRC_NOT_SUPPORTED:
        errorMessage = '不支援的影片格式或來源'
        break
      default:
        errorMessage = `錯誤代碼: ${error.code}`
    }
  }

  videoError.value = errorMessage
  console.error('影片播放錯誤:', errorMessage, error)
}

// 處理影片可以播放
const handleVideoCanPlay = () => {
  console.log('影片可以播放')
  videoError.value = ''
}

// 處理影片元數據載入
const handleVideoMetadata = (event: Event) => {
  const videoElement = event.target as HTMLVideoElement
  console.log('影片元數據:', {
    duration: videoElement.duration,
    videoWidth: videoElement.videoWidth,
    videoHeight: videoElement.videoHeight,
  })
}

// 下載影片
const downloadVideo = () => {
  const url = getVideoURL()
  if (url) {
    const link = document.createElement('a')
    link.href = url
    link.download = `${video.value?.title || 'video'}.${getVideoFormat().toLowerCase()}`
    link.target = '_blank'
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
  }
}

const handleLike = async () => {
  if (!video.value) return

  liking.value = true
  try {
    await likeVideo(video.value.id)
    video.value.likes += 1
    ElMessage.success('按讚成功！')
  } catch (error) {
    console.error('按讚失敗:', error)
  } finally {
    liking.value = false
  }
}

const getStatusType = (status: string) => {
  switch (status) {
    case 'ready':
      return 'success'
    case 'processing':
      return 'warning'
    case 'failed':
      return 'danger'
    default:
      return 'info'
  }
}

const getStatusText = (status: string) => {
  switch (status) {
    case 'ready':
      return '已完成'
    case 'processing':
      return '處理中'
    case 'failed':
      return '失敗'
    default:
      return status
  }
}

const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleDateString('zh-TW')
}

onMounted(() => {
  loadVideo()
})

onUnmounted(() => {
  // 清理 HLS 實例
  if (hls.value) {
    hls.value.destroy()
    hls.value = null
  }

  // 清理自動品質監控定時器
  if (autoQualityTimer.value) {
    clearInterval(autoQualityTimer.value)
    autoQualityTimer.value = null
  }
})
</script>

<style scoped>
.video-detail {
  max-width: 1000px;
  margin: 0 auto;
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

.loading {
  padding: 40px;
}

.video-card {
  overflow: hidden;
}

.video-player {
  margin-bottom: 24px;
}

.video-container {
  position: relative;
  width: 100%;
  background: #000;
  border-radius: 8px;
  overflow: hidden;
}

.video-container video {
  width: 100%;
  height: auto;
  display: block;
  object-fit: contain;
  max-height: 80vh;
}

.quality-selector {
  position: absolute;
  top: 10px;
  right: 10px;
  z-index: 10;
  background: rgba(0, 0, 0, 0.7);
  border-radius: 4px;
  padding: 4px;
}

.auto-quality-info {
  position: absolute;
  top: 50px;
  right: 10px;
  z-index: 10;
  animation: slideIn 0.3s ease-out;
}

@keyframes slideIn {
  from {
    opacity: 0;
    transform: translateX(20px);
  }
  to {
    opacity: 1;
    transform: translateX(0);
  }
}

.placeholder-player {
  height: 400px;
  background: #f5f5f5;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  border-radius: 8px;
}

.placeholder-icon {
  font-size: 64px;
  margin-bottom: 16px;
  color: #ccc;
}

.video-info h2 {
  margin: 0 0 16px 0;
  color: #333;
  font-size: 24px;
}

.video-description {
  color: #666;
  line-height: 1.6;
  margin-bottom: 24px;
}

.video-meta {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 16px;
  margin-bottom: 24px;
  padding: 20px;
  background: #f8f9fa;
  border-radius: 8px;
}

.meta-item {
  display: flex;
  align-items: center;
}

.meta-label {
  font-weight: bold;
  color: #666;
  margin-right: 8px;
}

.video-actions {
  display: flex;
  gap: 16px;
}

.video-error {
  margin-top: 16px;
  padding: 16px;
  background-color: #fef2f2;
  border: 1px solid #fecaca;
  border-radius: 8px;
  color: #dc2626;
}

.format-suggestion {
  margin-top: 12px;
  padding: 12px;
  background-color: #fffbeb;
  border: 1px solid #fed7aa;
  border-radius: 6px;
  color: #d97706;
}

.format-suggestion ul {
  margin: 8px 0 0 16px;
  padding: 0;
}

.format-suggestion li {
  margin: 4px 0;
}

.format-suggestion code {
  background-color: #374151;
  color: #f3f4f6;
  padding: 2px 6px;
  border-radius: 4px;
  font-family: 'Courier New', monospace;
  font-size: 12px;
}

.error-state {
  text-align: center;
  padding: 80px 20px;
}

.error-icon {
  font-size: 64px;
  margin-bottom: 16px;
}

.error-state h3 {
  color: #333;
  margin-bottom: 8px;
}

.error-state p {
  color: #666;
  margin-bottom: 24px;
}
</style>
