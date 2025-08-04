<template>
  <div class="debug-container">
    <h1>🔧 調試頁面</h1>

    <div class="debug-section">
      <h2>認證狀態</h2>
      <div class="debug-info">
        <p>
          <strong>Token:</strong> {{ authStore.token ? '已設置' : '未設置' }}
        </p>
        <p>
          <strong>用戶:</strong>
          {{
            authStore.user ? JSON.stringify(authStore.user, null, 2) : '未設置'
          }}
        </p>
        <p>
          <strong>已認證:</strong> {{ authStore.isAuthenticated ? '是' : '否' }}
        </p>
      </div>

      <div class="debug-actions">
        <el-button @click="refreshAuth">刷新認證狀態</el-button>
        <el-button @click="clearAuth" type="danger">清除認證</el-button>
        <el-button @click="testLogin">測試登入</el-button>
      </div>
    </div>

    <div class="debug-section">
      <h2>LocalStorage 內容</h2>
      <div class="debug-info">
        <p><strong>Token:</strong> {{ localStorageToken }}</p>
        <p><strong>User:</strong> {{ localStorageUser }}</p>
      </div>
    </div>

    <div class="debug-section">
      <h2>角色測試</h2>
      <div class="debug-info">
        <p><strong>當前用戶ID:</strong> {{ currentUserId }}</p>
        <p><strong>測試房間創建者ID:</strong> {{ testCreatorId }}</p>
        <p><strong>是否為創建者:</strong> {{ isCreatorTest ? '是' : '否' }}</p>
      </div>

      <div class="debug-actions">
        <el-input
          v-model="testCreatorId"
          placeholder="輸入測試房間創建者ID"
          style="width: 200px; margin-right: 10px"
        />
        <el-button @click="testRole">測試角色判斷</el-button>
      </div>
    </div>

    <div class="debug-section">
      <h2>API 測試</h2>
      <div class="debug-actions">
        <el-button @click="testGetUserInfo">測試獲取用戶信息</el-button>
        <el-button @click="testCreateRoom">測試創建直播間</el-button>
      </div>

      <div v-if="apiResult" class="debug-info">
        <h3>API 結果:</h3>
        <pre>{{ apiResult }}</pre>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useAuthStore } from '@/store/auth'
import { getUserInfo } from '@/api/user'
import { createRoom } from '@/api/live-room'
import { ElMessage } from 'element-plus'

const authStore = useAuthStore()

// 響應式數據
const localStorageToken = ref('')
const localStorageUser = ref('')
const testCreatorId = ref('21')
const apiResult = ref('')

// 計算屬性
const currentUserId = computed(() => authStore.user?.id || 0)
const isCreatorTest = computed(() => {
  const creatorId = parseInt(testCreatorId.value) || 0
  return creatorId === currentUserId.value
})

// 方法
const refreshAuth = () => {
  authStore.initAuth()
  loadLocalStorage()
  ElMessage.success('認證狀態已刷新')
}

const clearAuth = () => {
  authStore.logout()
  loadLocalStorage()
  ElMessage.success('認證已清除')
}

const testLogin = async () => {
  try {
    // 模擬登入過程
    const testUser = {
      id: 21,
      username: 'test_user',
      email: 'test@example.com',
      avatar: '',
      bio: '',
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
    }

    authStore.setAuth('test_token', testUser)
    loadLocalStorage()
    ElMessage.success('測試登入成功')
  } catch (error) {
    ElMessage.error('測試登入失敗')
  }
}

const testRole = () => {
  const creatorId = parseInt(testCreatorId.value) || 0
  const result = creatorId === currentUserId.value
  ElMessage.info(`角色測試結果: ${result ? '是創建者' : '不是創建者'}`)
}

const testGetUserInfo = async () => {
  if (!authStore.user?.id) {
    ElMessage.error('請先登入')
    return
  }

  try {
    const response = await getUserInfo(authStore.user.id)
    apiResult.value = JSON.stringify(response, null, 2)
    ElMessage.success('獲取用戶信息成功')
  } catch (error: any) {
    apiResult.value = `錯誤: ${error.message}`
    ElMessage.error('獲取用戶信息失敗')
  }
}

const testCreateRoom = async () => {
  if (!authStore.token) {
    ElMessage.error('請先登入')
    return
  }

  try {
    const response = await createRoom({
      title: `測試直播間 ${Date.now()}`,
      description: '這是調試用的測試直播間',
    })
    apiResult.value = JSON.stringify(response, null, 2)
    ElMessage.success('創建直播間成功')
  } catch (error: any) {
    apiResult.value = `錯誤: ${error.message}`
    ElMessage.error('創建直播間失敗')
  }
}

const loadLocalStorage = () => {
  localStorageToken.value = localStorage.getItem('token') || '無'
  const userStr = localStorage.getItem('user') || '無'
  try {
    localStorageUser.value = JSON.stringify(JSON.parse(userStr), null, 2)
  } catch {
    localStorageUser.value = userStr
  }
}

onMounted(() => {
  loadLocalStorage()
})
</script>

<style scoped>
.debug-container {
  max-width: 800px;
  margin: 0 auto;
  padding: 20px;
}

.debug-section {
  margin-bottom: 30px;
  padding: 20px;
  border: 1px solid #ddd;
  border-radius: 8px;
  background-color: #f9f9f9;
}

.debug-section h2 {
  margin-top: 0;
  color: #333;
  border-bottom: 2px solid #409eff;
  padding-bottom: 10px;
}

.debug-info {
  background-color: white;
  padding: 15px;
  border-radius: 4px;
  margin-bottom: 15px;
}

.debug-info p {
  margin: 5px 0;
}

.debug-info pre {
  background-color: #f5f5f5;
  padding: 10px;
  border-radius: 4px;
  overflow-x: auto;
  white-space: pre-wrap;
}

.debug-actions {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
  align-items: center;
}
</style>
