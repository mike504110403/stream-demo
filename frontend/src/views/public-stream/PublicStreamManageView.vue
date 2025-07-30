<template>
  <div class="public-stream-manage">
    <!-- 頁面標題 -->
    <div class="page-header">
      <h1 class="page-title">公開直播源管理</h1>
      <p class="page-subtitle">管理外部直播源的配置和狀態</p>
    </div>

    <!-- 操作按鈕 -->
    <div class="action-bar">
      <el-button @click="showAddDialog = true" type="primary" class="add-btn">
        <svg fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"></path>
        </svg>
        新增直播源
      </el-button>
      <el-button @click="refreshStreams" type="info" class="refresh-btn">
        <svg fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"></path>
        </svg>
        重新載入
      </el-button>
    </div>

    <!-- 載入狀態 -->
    <div v-if="loading" class="loading-container">
      <el-loading-component />
      <p>正在載入配置...</p>
    </div>

    <!-- 錯誤狀態 -->
    <div v-else-if="error" class="error-container">
      <el-alert
        :title="error"
        type="error"
        :closable="false"
        show-icon
      />
      <el-button @click="loadStreams" type="primary" class="retry-button">
        重新載入
      </el-button>
    </div>

    <!-- 流列表 -->
    <div v-else class="stream-table-container">
      <el-table :data="streams" style="width: 100%" class="stream-table">
        <el-table-column prop="name" label="名稱" width="150">
          <template #default="{ row }">
            <div class="stream-name">
              <span class="name-text">{{ row.name }}</span>
              <el-tag v-if="row.enabled" type="success" size="small">啟用</el-tag>
              <el-tag v-else type="info" size="small">停用</el-tag>
            </div>
          </template>
        </el-table-column>
        
        <el-table-column prop="title" label="標題" width="200">
          <template #default="{ row }">
            <div class="stream-title">
              <h4>{{ row.title }}</h4>
              <p class="description">{{ row.description }}</p>
            </div>
          </template>
        </el-table-column>
        
        <el-table-column prop="url" label="URL" width="300">
          <template #default="{ row }">
            <div class="stream-url">
              <el-input v-model="row.url" size="small" readonly>
                <template #append>
                  <el-button @click="copyToClipboard(row.url)" size="small">
                    複製
                  </el-button>
                </template>
              </el-input>
            </div>
          </template>
        </el-table-column>
        
        <el-table-column prop="type" label="類型" width="100">
          <template #default="{ row }">
            <el-tag :type="row.type === 'rtmp' ? 'danger' : 'primary'" size="small">
              {{ row.type.toUpperCase() }}
            </el-tag>
          </template>
        </el-table-column>
        
        <el-table-column prop="category" label="分類" width="120">
          <template #default="{ row }">
            <el-tag type="warning" size="small">{{ getCategoryLabel(row.category) }}</el-tag>
          </template>
        </el-table-column>
        
        <el-table-column prop="status" label="狀態" width="120">
          <template #default="{ row }">
            <div class="status-indicator">
              <span class="status-dot" :class="{ 'active': row.status === 'active' }">●</span>
              {{ row.status === 'active' ? '運行中' : '已停止' }}
            </div>
          </template>
        </el-table-column>
        
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <div class="action-buttons">
              <el-button @click="editStream(row)" size="small" type="primary">
                編輯
              </el-button>
              <el-button @click="toggleStream(row)" size="small" :type="row.enabled ? 'warning' : 'success'">
                {{ row.enabled ? '停用' : '啟用' }}
              </el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- 新增/編輯對話框 -->
    <el-dialog
      v-model="showAddDialog"
      :title="editingStream ? '編輯直播源' : '新增直播源'"
      width="600px"
      @close="resetForm"
    >
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-form-item label="名稱" prop="name">
          <el-input v-model="form.name" placeholder="輸入唯一的名稱" :disabled="!!editingStream" />
        </el-form-item>
        
        <el-form-item label="標題" prop="title">
          <el-input v-model="form.title" placeholder="輸入顯示標題" />
        </el-form-item>
        
        <el-form-item label="描述" prop="description">
          <el-input v-model="form.description" type="textarea" placeholder="輸入描述" />
        </el-form-item>
        
        <el-form-item label="URL" prop="url">
          <el-input v-model="form.url" placeholder="輸入直播源 URL" />
        </el-form-item>
        
        <el-form-item label="類型" prop="type">
          <el-select v-model="form.type" placeholder="選擇流類型">
            <el-option label="HLS" value="hls" />
            <el-option label="RTMP" value="rtmp" />
          </el-select>
        </el-form-item>
        
        <el-form-item label="分類" prop="category">
          <el-select v-model="form.category" placeholder="選擇分類">
            <el-option label="測試" value="test" />
            <el-option label="太空" value="space" />
            <el-option label="新聞" value="news" />
            <el-option label="體育" value="sports" />
            <el-option label="娛樂" value="entertainment" />
          </el-select>
        </el-form-item>
        
        <el-form-item label="啟用狀態" prop="enabled">
          <el-switch v-model="form.enabled" />
        </el-form-item>
      </el-form>
      
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="showAddDialog = false">取消</el-button>
          <el-button @click="saveStream" type="primary" :loading="saving">
            {{ editingStream ? '更新' : '新增' }}
          </el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
// import { ElMessageBox } from 'element-plus'  // 暫時註釋掉未使用的 import

// 響應式數據
const streams = ref<any[]>([])
const loading = ref(false)
const saving = ref(false)
const error = ref('')
const showAddDialog = ref(false)
const editingStream = ref<any>(null)
const formRef = ref()

// 表單數據
const form = ref({
  name: '',
  title: '',
  description: '',
  url: '',
  type: 'hls',
  category: 'test',
  enabled: true
})

// 表單驗證規則
const rules: any = {
  name: [
    { required: true, message: '請輸入名稱', trigger: 'blur' },
    { min: 2, max: 50, message: '名稱長度在 2 到 50 個字符', trigger: 'blur' }
  ],
  title: [
    { required: true, message: '請輸入標題', trigger: 'blur' },
    { min: 2, max: 100, message: '標題長度在 2 到 100 個字符', trigger: 'blur' }
  ],
  url: [
    { required: true, message: '請輸入 URL', trigger: 'blur' },
    { type: 'url', message: '請輸入有效的 URL', trigger: 'blur' }
  ],
  type: [
    { required: true, message: '請選擇類型', trigger: 'change' }
  ],
  category: [
    { required: true, message: '請選擇分類', trigger: 'change' }
  ]
}

// 分類標籤映射
const categoryLabels: Record<string, string> = {
  test: '測試',
  space: '太空',
  news: '新聞',
  sports: '體育',
  entertainment: '娛樂'
}

// 方法
const loadStreams = async () => {
  loading.value = true
  error.value = ''
  
  try {
    // 通過 Vite 代理調用 Stream-Puller 的 API，獲取資料庫中的所有資料
    const response = await fetch('/stream-puller/api/public-streams')
    const data = await response.json()
    
    console.log('Stream-Puller API 響應:', data) // 調試用
    
    if (data.success && data.data && data.data.streams) {
      streams.value = data.data.streams.map((stream: any) => ({
        ...stream,
        status: stream.enabled ? 'active' : 'inactive'
      }))
    } else {
      throw new Error('響應格式不正確')
    }
  } catch (err) {
    console.error('載入流配置失敗:', err)
    error.value = '載入流配置失敗，請稍後再試'
  } finally {
    loading.value = false
  }
}

const refreshStreams = () => {
  loadStreams()
}

const editStream = (stream: any) => {
  editingStream.value = stream
  form.value = { ...stream }
  showAddDialog.value = true
}

const toggleStream = async (stream: any) => {
  try {
    console.log('切換流狀態:', stream) // 調試用
    
    if (!stream.name) {
      ElMessage.error('流名稱不能為空')
      return
    }
    
    const newEnabled = !stream.enabled
    const formData = new URLSearchParams()
    formData.append('name', stream.name)
    formData.append('enabled', newEnabled.toString())
    
    console.log('發送資料:', { name: stream.name, enabled: newEnabled }) // 調試用
    
    const response = await fetch(`/stream-puller/api/public-streams`, {
      method: 'PUT',
      body: formData,
      headers: {
        'Content-Type': 'application/x-www-form-urlencoded'
      }
    })
    
    if (response.ok) {
      stream.enabled = newEnabled
      stream.status = newEnabled ? 'active' : 'inactive'
      ElMessage.success(`已${newEnabled ? '啟用' : '停用'}直播源`)
    } else {
      const errorText = await response.text()
      console.error('API 錯誤響應:', errorText) // 調試用
      throw new Error(`操作失敗: ${errorText}`)
    }
  } catch (err) {
    console.error('切換狀態失敗:', err)
    ElMessage.error('操作失敗，請稍後再試')
  }
}

// 移除刪除功能

const saveStream = async () => {
  if (!formRef.value) return
  
  try {
    await formRef.value.validate()
    saving.value = true
    
    const formData = new URLSearchParams()
    Object.keys(form.value).forEach(key => {
      formData.append(key, form.value[key as keyof typeof form.value].toString())
    })
    
    const method = editingStream.value ? 'PUT' : 'POST'
    const response = await fetch(`/stream-puller/api/public-streams`, {
      method,
      body: formData,
      headers: {
        'Content-Type': 'application/x-www-form-urlencoded'
      }
    })
    
    if (response.ok) {
      ElMessage.success(editingStream.value ? '更新成功' : '新增成功')
      showAddDialog.value = false
      loadStreams()
    } else {
      throw new Error('保存失敗')
    }
  } catch (err) {
    console.error('保存失敗:', err)
    ElMessage.error('保存失敗，請稍後再試')
  } finally {
    saving.value = false
  }
}

const resetForm = () => {
  editingStream.value = null
  form.value = {
    name: '',
    title: '',
    description: '',
    url: '',
    type: 'hls',
    category: 'test',
    enabled: true
  }
  if (formRef.value) {
    formRef.value.resetFields()
  }
}

const copyToClipboard = async (text: string) => {
  try {
    await navigator.clipboard.writeText(text)
    ElMessage.success('已複製到剪貼板')
  } catch (err) {
    console.error('複製失敗:', err)
    ElMessage.error('複製失敗')
  }
}

const getCategoryLabel = (category: string) => {
  return categoryLabels[category] || category
}

// 生命週期
onMounted(() => {
  loadStreams()
})
</script>

<style scoped>
.public-stream-manage {
  max-width: 1400px;
  margin: 0 auto;
  padding: 20px;
}

.page-header {
  text-align: center;
  margin-bottom: 30px;
}

.page-title {
  font-size: 2.5rem;
  font-weight: bold;
  color: #2c3e50;
  margin-bottom: 10px;
}

.page-subtitle {
  font-size: 1.1rem;
  color: #7f8c8d;
}

.action-bar {
  display: flex;
  gap: 12px;
  margin-bottom: 24px;
}

.add-btn,
.refresh-btn {
  display: flex;
  align-items: center;
  gap: 8px;
}

.loading-container,
.error-container {
  text-align: center;
  padding: 60px 20px;
}

.retry-button {
  margin-top: 20px;
}

.stream-table-container {
  background: white;
  border-radius: 12px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.1);
  overflow: hidden;
}

.stream-table {
  border: none;
}

.stream-name {
  display: flex;
  align-items: center;
  gap: 8px;
}

.name-text {
  font-weight: 600;
  color: #2c3e50;
}

.stream-title h4 {
  margin: 0 0 4px 0;
  font-size: 14px;
  font-weight: 600;
  color: #2c3e50;
}

.stream-title .description {
  margin: 0;
  font-size: 12px;
  color: #7f8c8d;
  line-height: 1.4;
}

.stream-url {
  width: 100%;
}

.status-indicator {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
}

.status-dot {
  font-size: 8px;
  color: #dc3545;
}

.status-dot.active {
  color: #28a745;
  animation: pulse 2s infinite;
}

.action-buttons {
  display: flex;
  gap: 6px;
}

.action-buttons .el-button {
  padding: 4px 8px;
  font-size: 12px;
}

@keyframes pulse {
  0%, 100% {
    opacity: 1;
  }
  50% {
    opacity: 0.5;
  }
}

.dialog-footer {
  text-align: right;
}

/* 響應式設計 */
@media (max-width: 768px) {
  .public-stream-manage {
    padding: 16px;
  }
  
  .page-title {
    font-size: 2rem;
  }
  
  .action-bar {
    flex-direction: column;
  }
  
  .stream-table-container {
    overflow-x: auto;
  }
}
</style> 