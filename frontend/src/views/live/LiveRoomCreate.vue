<template>
  <div class="live-room-create">
    <div class="page-header">
      <h1>創建直播間</h1>
    </div>

    <div class="create-form">
      <el-card>
        <el-form 
          ref="formRef" 
          :model="form" 
          :rules="rules" 
          label-width="120px"
          @submit.prevent="handleSubmit"
        >
          <el-form-item label="直播間標題" prop="title">
            <el-input 
              v-model="form.title" 
              placeholder="請輸入直播間標題"
              maxlength="50"
              show-word-limit
            />
          </el-form-item>

          <el-form-item label="直播間描述" prop="description">
            <el-input 
              v-model="form.description" 
              type="textarea" 
              :rows="4"
              placeholder="請輸入直播間描述（可選）"
              maxlength="200"
              show-word-limit
            />
          </el-form-item>

          <el-form-item>
            <el-button 
              type="primary" 
              @click="handleSubmit" 
              :loading="loading"
              size="large"
            >
              創建直播間
            </el-button>
            <el-button @click="$router.back()" size="large">
              取消
            </el-button>
          </el-form-item>
        </el-form>
      </el-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { createRoom } from '@/api/live-room'
import type { CreateRoomRequest } from '@/types'

const router = useRouter()
const formRef = ref<FormInstance>()
const loading = ref(false)

const form = ref<CreateRoomRequest>({
  title: '',
  description: ''
})

const rules: FormRules = {
  title: [
    { required: true, message: '請輸入直播間標題', trigger: 'blur' },
    { min: 2, max: 50, message: '標題長度應在 2 到 50 個字符之間', trigger: 'blur' }
  ],
  description: [
    { max: 200, message: '描述不能超過 200 個字符', trigger: 'blur' }
  ]
}

const handleSubmit = async () => {
  if (!formRef.value) return
  
  // 防止重複提交
  if (loading.value) return

  try {
    await formRef.value.validate()
    loading.value = true

    const response = await createRoom(form.value)
    console.log('創建直播間回應:', response) // 調試用
    
    // 檢查回應是否包含 id
    if (!response || !response.id) {
      console.error('無法取得直播間ID，回應格式:', response)
      throw new Error('無法取得直播間ID')
    }
    
    const roomId = response.id
    
    ElMessage.success('直播間創建成功！')
    
    // 跳轉到直播間詳情頁面
    router.push(`/live-rooms/${roomId}`)
  } catch (error) {
    console.error('創建直播間失敗:', error)
    if (error instanceof Error) {
      ElMessage.error(error.message)
    } else {
      ElMessage.error('創建直播間失敗')
    }
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.live-room-create {
  padding: 20px;
  max-width: 800px;
  margin: 0 auto;
}

.page-header {
  margin-bottom: 30px;
}

.page-header h1 {
  margin: 0;
  color: #333;
}

.create-form {
  max-width: 600px;
}

.el-form-item {
  margin-bottom: 24px;
}

.el-button {
  margin-right: 12px;
}
</style> 