<template>
  <div class="public-stream-list">
    <div class="container mx-auto px-4 py-8">
      <!-- 頁面標題 -->
      <div class="mb-8">
        <h1 class="text-3xl font-bold text-gray-900 mb-2">公開直播</h1>
        <p class="text-gray-600">觀看免費的公開直播流</p>
      </div>

      <!-- 分類篩選 -->
      <div class="mb-6">
        <div class="flex flex-wrap gap-2">
          <button
            v-for="category in categories"
            :key="category.value"
            @click="selectedCategory = category.value"
            :class="[
              'px-4 py-2 rounded-lg text-sm font-medium transition-colors',
              selectedCategory === category.value
                ? 'bg-blue-600 text-white'
                : 'bg-gray-200 text-gray-700 hover:bg-gray-300'
            ]"
          >
            {{ category.label }}
          </button>
        </div>
      </div>

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
          @click="loadStreams"
          class="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
        >
          重新載入
        </button>
      </div>

      <!-- 流列表 -->
      <div v-else-if="streams.length > 0" class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        <div
          v-for="stream in filteredStreams"
          :key="stream.name"
          class="bg-white rounded-lg shadow-md overflow-hidden hover:shadow-lg transition-shadow"
        >
          <!-- 流卡片 -->
          <div class="relative">
            <!-- 預覽圖 -->
            <div class="aspect-video bg-gray-200 flex items-center justify-center">
              <div class="text-center">
                <svg class="w-16 h-16 mx-auto text-gray-400 mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z"></path>
                </svg>
                <p class="text-sm text-gray-500">直播預覽</p>
              </div>
            </div>

            <!-- 狀態標籤 -->
            <div class="absolute top-2 right-2">
              <span
                :class="[
                  'px-2 py-1 text-xs font-medium rounded-full',
                  stream.status === 'active'
                    ? 'bg-green-100 text-green-800'
                    : 'bg-red-100 text-red-800'
                ]"
              >
                {{ stream.status === 'active' ? '直播中' : '離線' }}
              </span>
            </div>

            <!-- 觀看者數量 -->
            <div class="absolute bottom-2 left-2">
              <span class="px-2 py-1 bg-black bg-opacity-50 text-white text-xs rounded-full">
                👥 {{ stream.viewer_count }}
              </span>
            </div>
          </div>

          <!-- 流資訊 -->
          <div class="p-4">
            <h3 class="text-lg font-semibold text-gray-900 mb-2">{{ stream.title }}</h3>
            <p class="text-gray-600 text-sm mb-3 line-clamp-2">{{ stream.description }}</p>
            
            <!-- 分類標籤 -->
            <div class="mb-4">
              <span class="inline-block px-2 py-1 bg-blue-100 text-blue-800 text-xs rounded-full">
                {{ getCategoryLabel(stream.category) }}
              </span>
            </div>

            <!-- 最後更新時間 -->
            <p class="text-xs text-gray-500 mb-4">
              最後更新: {{ formatTime(stream.last_update) }}
            </p>

            <!-- 操作按鈕 -->
            <div class="flex gap-2">
              <button
                @click="watchStream(stream)"
                :disabled="stream.status !== 'active'"
                :class="[
                  'flex-1 px-4 py-2 rounded-lg text-sm font-medium transition-colors',
                  stream.status === 'active'
                    ? 'bg-blue-600 text-white hover:bg-blue-700'
                    : 'bg-gray-300 text-gray-500 cursor-not-allowed'
                ]"
              >
                {{ stream.status === 'active' ? '觀看直播' : '離線' }}
              </button>
              
              <button
                @click="viewStreamInfo(stream)"
                class="px-4 py-2 bg-gray-200 text-gray-700 rounded-lg hover:bg-gray-300 transition-colors"
              >
                詳情
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- 空狀態 -->
      <div v-else class="text-center py-12">
        <div class="text-gray-400 mb-4">
          <svg class="w-16 h-16 mx-auto mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z"></path>
          </svg>
          <p class="text-lg font-medium">暫無可用的直播流</p>
          <p class="text-sm">請稍後再試</p>
        </div>
      </div>
    </div>

    <!-- 流詳情模態框 -->
    <Modal v-model:show="showStreamModal" title="流詳情">
      <div v-if="selectedStream" class="space-y-4">
        <div>
          <h3 class="text-lg font-semibold">{{ selectedStream.title }}</h3>
          <p class="text-gray-600">{{ selectedStream.description }}</p>
        </div>
        
        <div class="grid grid-cols-2 gap-4 text-sm">
          <div>
            <span class="font-medium">狀態:</span>
            <span :class="selectedStream.status === 'active' ? 'text-green-600' : 'text-red-600'">
              {{ selectedStream.status === 'active' ? '直播中' : '離線' }}
            </span>
          </div>
          <div>
            <span class="font-medium">觀看者:</span>
            <span>{{ selectedStream.viewer_count }}</span>
          </div>
          <div>
            <span class="font-medium">分類:</span>
            <span>{{ getCategoryLabel(selectedStream.category) }}</span>
          </div>
          <div>
            <span class="font-medium">最後更新:</span>
            <span>{{ formatTime(selectedStream.last_update) }}</span>
          </div>
        </div>
      </div>
    </Modal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { publicStreamApi } from '@/api/public-stream'
import type { PublicStreamInfo } from '@/types/public-stream'
import Modal from '@/components/common/Modal.vue'

const router = useRouter()

// 響應式數據
const streams = ref<PublicStreamInfo[]>([])
const loading = ref(false)
const error = ref('')
const selectedCategory = ref('all')
const showStreamModal = ref(false)
const selectedStream = ref<PublicStreamInfo | null>(null)

// 分類選項
const categories = [
  { value: 'all', label: '全部' },
  { value: 'test', label: '測試' },
  { value: 'space', label: '太空' },
  { value: 'news', label: '新聞' },
  { value: 'sports', label: '體育' }
]

// 計算屬性
const filteredStreams = computed(() => {
  if (selectedCategory.value === 'all') {
    return streams.value
  }
  return streams.value.filter(stream => stream.category === selectedCategory.value)
})

// 方法
const loadStreams = async () => {
  loading.value = true
  error.value = ''
  
  try {
    const response = await publicStreamApi.getAvailableStreams()
    streams.value = response.data.streams
  } catch (err) {
    console.error('載入流列表失敗:', err)
    error.value = '載入流列表失敗，請稍後再試'
  } finally {
    loading.value = false
  }
}

const watchStream = (stream: PublicStreamInfo) => {
  if (stream.status === 'active') {
    router.push(`/public-streams/${stream.name}`)
  }
}

const viewStreamInfo = (stream: PublicStreamInfo) => {
  selectedStream.value = stream
  showStreamModal.value = true
}

const getCategoryLabel = (category: string) => {
  const found = categories.find(c => c.value === category)
  return found ? found.label : category
}

const formatTime = (timeString: string) => {
  const date = new Date(timeString)
  return date.toLocaleString('zh-TW')
}

// 生命週期
onMounted(() => {
  loadStreams()
})
</script>

<style scoped>
.line-clamp-2 {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
</style> 