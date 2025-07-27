<template>
  <div class="public-stream-player">
    <div class="container mx-auto px-4 py-8">
      <!-- 載入狀態 -->
      <div v-if="loading" class="flex justify-center items-center py-12">
        <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
      </div>

      <!-- 錯誤狀態 -->
      <div v-else-if="error" class="text-center py-12">
        <div class="text-red-600 mb-4">
          <svg class="w-16 h-16 mx-auto mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L3.732 16.5c-.77.833.192 2.5 1.732 2.5z"></path>
          </svg>
          <p class="text-lg font-medium">{{ error }}</p>
        </div>
        <button
          @click="loadStreamInfo"
          class="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
        >
          重新載入
        </button>
      </div>

      <!-- 播放器 -->
      <div v-else-if="streamInfo" class="space-y-6">
        <!-- 返回按鈕 -->
        <div class="flex items-center gap-4">
          <button
            @click="goBack"
            class="flex items-center gap-2 px-4 py-2 text-gray-600 hover:text-gray-900 transition-colors"
          >
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7"></path>
            </svg>
            返回列表
          </button>
        </div>

        <!-- 流資訊 -->
        <div class="bg-white rounded-lg shadow-md p-6">
          <div class="flex items-start justify-between mb-4">
            <div>
              <h1 class="text-2xl font-bold text-gray-900 mb-2">{{ streamInfo.title }}</h1>
              <p class="text-gray-600">{{ streamInfo.description }}</p>
            </div>
            <div class="flex items-center gap-4">
              <!-- 狀態標籤 -->
              <span
                :class="[
                  'px-3 py-1 text-sm font-medium rounded-full',
                  streamInfo.status === 'active'
                    ? 'bg-green-100 text-green-800'
                    : 'bg-red-100 text-red-800'
                ]"
              >
                {{ streamInfo.status === 'active' ? '直播中' : '離線' }}
              </span>
              
              <!-- 觀看者數量 -->
              <span class="flex items-center gap-1 text-sm text-gray-600">
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z"></path>
                </svg>
                {{ streamInfo.viewer_count }}
              </span>
            </div>
          </div>

          <!-- 分類和時間 -->
          <div class="flex items-center gap-4 text-sm text-gray-500">
            <span class="inline-block px-2 py-1 bg-blue-100 text-blue-800 rounded-full">
              {{ getCategoryLabel(streamInfo.category) }}
            </span>
            <span>最後更新: {{ formatTime(streamInfo.last_update) }}</span>
          </div>
        </div>

        <!-- 視頻播放器 -->
        <div class="bg-black rounded-lg overflow-hidden">
          <div class="aspect-video relative">
            <!-- HLS 播放器 -->
            <video
              v-if="streamInfo && streamInfo.status === 'active'"
              ref="videoPlayer"
              class="w-full h-full"
              controls
              autoplay
              muted
              crossorigin="anonymous"
            >
              您的瀏覽器不支援 HLS 播放。
            </video>
            
            <!-- 調試信息 -->
            <div v-if="streamInfo" class="text-sm text-gray-500 mt-2">
              <p>流狀態: {{ streamInfo.status }}</p>
              <p>播放 URL: {{ playbackUrl || '未載入' }}</p>
              <p>videoPlayer 元素: {{ videoPlayer ? '已準備' : '未準備' }}</p>
            </div>

            <!-- 離線狀態 -->
            <div v-else class="w-full h-full flex items-center justify-center bg-gray-900">
              <div class="text-center text-white">
                <svg class="w-16 h-16 mx-auto mb-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z"></path>
                </svg>
                <p class="text-lg font-medium mb-2">直播已離線</p>
                <p class="text-sm text-gray-400">此直播目前無法觀看</p>
              </div>
            </div>

            <!-- 載入狀態 -->
            <div v-if="loadingPlaybackUrl" class="absolute inset-0 flex items-center justify-center bg-black bg-opacity-50">
              <div class="text-center text-white">
                <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-white mx-auto mb-2"></div>
                <p class="text-sm">載入播放器...</p>
              </div>
            </div>
          </div>
        </div>

        <!-- 播放模式選擇 -->
        <div class="bg-white rounded-lg shadow-md p-4">
          <div class="flex items-center justify-between mb-4">
            <h3 class="text-lg font-medium text-gray-900">播放模式</h3>
            <div class="flex items-center gap-2 text-sm text-gray-500">
              <span>品質: 自動</span>
            </div>
          </div>
          
          <div class="flex items-center gap-4">
            <button
              @click="switchToHLS"
              :class="[
                'flex items-center gap-2 px-4 py-2 rounded-lg transition-colors',
                playbackMode === 'hls' 
                  ? 'bg-blue-600 text-white' 
                  : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
              ]"
            >
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z"></path>
              </svg>
              HLS (相容性好，延遲 2-5 秒)
            </button>
            
            <button
              @click="switchToRTMP"
              :class="[
                'flex items-center gap-2 px-4 py-2 rounded-lg transition-colors',
                playbackMode === 'rtmp' 
                  ? 'bg-green-600 text-white' 
                  : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
              ]"
            >
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z"></path>
              </svg>
              RTMP (低延遲，需要外部播放器)
            </button>
          </div>
          
          <div class="mt-4 flex items-center gap-4">
            <button
              @click="toggleMute"
              class="flex items-center gap-2 px-3 py-2 text-gray-600 hover:text-gray-900 transition-colors"
            >
              <svg v-if="!isMuted" class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15.536 8.464a5 5 0 010 7.072m2.828-9.9a9 9 0 010 12.728M5.586 15H4a1 1 0 01-1-1v-4a1 1 0 011-1h1.586l4.707-4.707C10.923 3.663 12 4.109 12 5v14c0 .891-1.077 1.337-1.707.707L5.586 15z"></path>
              </svg>
              <svg v-else class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5.586 15H4a1 1 0 01-1-1v-4a1 1 0 011-1h1.586l4.707-4.707C10.923 3.663 12 4.109 12 5v14c0 .891-1.077 1.337-1.707.707L5.586 15z"></path>
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2"></path>
              </svg>
              {{ isMuted ? '取消靜音' : '靜音' }}
            </button>

            <button
              @click="toggleFullscreen"
              class="flex items-center gap-2 px-3 py-2 text-gray-600 hover:text-gray-900 transition-colors"
            >
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 8V4m0 0h4M4 4l5 5m11-1V4m0 0h-4m4 0l-5 5M4 16v4m0 0h4m-4 0l5-5m11 5l-5-5m5 5v-4m0 4h-4"></path>
              </svg>
              全螢幕
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

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
const error = ref('')
const isMuted = ref(false)
const videoPlayer = ref<HTMLVideoElement>()
const hls = ref<Hls | null>(null)
const flvPlayer = ref<flvjs.Player | null>(null)
const playbackMode = ref<'hls' | 'rtmp'>('hls')
const streamURLs = ref<{ hls: string; rtmp: string } | null>(null)

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
    streamInfo.value = response.data
    
    console.log('流資訊載入成功:', response.data)
    
    // 如果流是活躍的，獲取播放 URL
    if (response.data.status === 'active') {
      console.log('流狀態為 active，開始載入播放 URL')
      // 延遲一下再載入播放 URL，確保 DOM 已更新
      setTimeout(() => {
        loadPlaybackUrl(streamName)
      }, 500)
    } else {
      console.log('流狀態不是 active:', response.data.status)
    }
  } catch (err) {
    console.error('載入流資訊失敗:', err)
    error.value = '載入流資訊失敗，請稍後再試'
  } finally {
    loading.value = false
  }
}

const loadPlaybackUrl = async (streamName: string, mode: 'hls' | 'rtmp' = 'hls') => {
  loadingPlaybackUrl.value = true
  
  try {
    // 獲取所有播放 URL
    if (!streamURLs.value) {
      const response = await publicStreamApi.getStreamURLs(streamName)
      streamURLs.value = response.data.urls
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
        
        // 創建新的 HLS 實例
        hls.value = new Hls({
          debug: true, // 開啟調試模式
          enableWorker: true,
          lowLatencyMode: true,
        })
        
        // 載入流
        hls.value.loadSource(playbackUrl.value)
        hls.value.attachMedia(videoPlayer.value)
        
        // 監聽事件
        hls.value.on(Hls.Events.MANIFEST_PARSED, () => {
          console.log('HLS 流載入成功')
          // 只有在 HLS 模式下才嘗試播放
          if (videoPlayer.value && playbackMode.value === 'hls' && !videoPlayer.value.hasAttribute('data-no-play')) {
            console.log('嘗試 HLS 自動播放')
            videoPlayer.value.play().catch(err => {
              // 只有在 HLS 模式下才記錄錯誤
              if (playbackMode.value === 'hls') {
                console.error('自動播放失敗:', err)
              }
            })
          } else {
            console.log('跳過 HLS 自動播放，當前模式:', playbackMode.value, '禁止播放:', videoPlayer.value?.hasAttribute('data-no-play'))
          }
        })
        
        hls.value.on(Hls.Events.ERROR, (event, data) => {
          console.error('HLS 錯誤:', data)
          if (data.fatal) {
            error.value = '播放器載入失敗，請重新整理頁面'
          }
        })
        
        hls.value.on(Hls.Events.MEDIA_ATTACHED, () => {
          console.log('媒體元素已附加')
        })
        
      } else if (videoPlayer.value.canPlayType('application/vnd.apple.mpegurl')) {
        console.log('使用瀏覽器原生 HLS 支援')
        // Safari 原生支援 HLS
        videoPlayer.value.src = playbackUrl.value
        videoPlayer.value.addEventListener('loadedmetadata', () => {
          if (playbackMode.value === 'hls' && !videoPlayer.value?.hasAttribute('data-no-play')) {
            console.log('嘗試 Safari 原生 HLS 自動播放')
            videoPlayer.value?.play().catch(err => {
              // 只有在 HLS 模式下才記錄錯誤
              if (playbackMode.value === 'hls') {
                console.error('自動播放失敗:', err)
              }
            })
          } else {
            console.log('跳過 Safari 原生 HLS 自動播放，當前模式:', playbackMode.value, '禁止播放:', videoPlayer.value?.hasAttribute('data-no-play'))
          }
        })
      } else {
        console.error('瀏覽器不支援 HLS')
        error.value = '您的瀏覽器不支援 HLS 播放'
      }
    } else {
      // RTMP 模式 - 直接返回，不執行任何播放相關代碼
      playbackUrl.value = streamURLs.value!.rtmp
      console.log('RTMP URL:', playbackUrl.value)
      console.log('提示：RTMP 流需要使用外部播放器，如 VLC')
      
      // 清理所有播放器
      if (hls.value) {
        hls.value.destroy()
        hls.value = null
      }
      if (flvPlayer.value) {
        flvPlayer.value.destroy()
        flvPlayer.value = null
      }
      
      // 顯示 RTMP URL 供用戶複製
      error.value = `RTMP 播放需要外部播放器。請使用 VLC 或其他支援 RTMP 的播放器打開：${playbackUrl.value}`
      
      // 直接返回，不執行任何播放相關代碼
      return
    }
  } catch (err) {
    console.error('獲取播放 URL 失敗:', err)
    error.value = '獲取播放 URL 失敗'
  } finally {
    loadingPlaybackUrl.value = false
  }
}

const toggleMute = () => {
  if (videoPlayer.value) {
    videoPlayer.value.muted = !videoPlayer.value.muted
    isMuted.value = videoPlayer.value.muted
  }
}

const toggleFullscreen = () => {
  if (videoPlayer.value) {
    if (document.fullscreenElement) {
      document.exitFullscreen()
    } else {
      videoPlayer.value.requestFullscreen()
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

const switchToHLS = () => {
  playbackMode.value = 'hls'
  if (streamURLs.value) {
    loadPlaybackUrl(route.params.name as string, 'hls')
  }
}

const switchToRTMP = () => {
  playbackMode.value = 'rtmp'
  if (streamURLs.value) {
    loadPlaybackUrl(route.params.name as string, 'rtmp')
  }
}

// 生命週期
onMounted(async () => {
  await loadStreamInfo()
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
</style> 