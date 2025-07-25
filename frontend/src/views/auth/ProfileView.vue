<template>
  <div class="profile">
    <div class="profile-header">
      <h1>個人資料</h1>
    </div>

    <el-card class="profile-card">
      <el-form
        ref="profileFormRef"
        :model="profileForm"
        :rules="profileRules"
        label-width="100px"
      >
        <el-form-item label="頭像">
          <div class="avatar-section">
            <el-avatar :size="80" :src="profileForm.avatar">
              {{ profileForm.username?.charAt(0).toUpperCase() }}
            </el-avatar>
            <div class="avatar-info">
              <el-input
                v-model="profileForm.avatar"
                placeholder="請輸入頭像 URL"
                style="margin-top: 12px;"
              />
            </div>
          </div>
        </el-form-item>

        <el-form-item label="用戶名" prop="username">
          <el-input v-model="profileForm.username" placeholder="請輸入用戶名" />
        </el-form-item>

        <el-form-item label="郵箱" prop="email">
          <el-input v-model="profileForm.email" type="email" placeholder="請輸入郵箱" />
        </el-form-item>

        <el-form-item label="個人簡介" prop="bio">
          <el-input
            v-model="profileForm.bio"
            type="textarea"
            :rows="4"
            placeholder="請輸入個人簡介"
          />
        </el-form-item>

        <el-form-item>
          <el-button type="primary" @click="handleUpdate" :loading="loading">
            更新資料
          </el-button>
          <el-button @click="resetForm">
            重置
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 帳號設定 -->
    <el-card class="profile-card" style="margin-top: 20px;">
      <template #header>
        <span>帳號設定</span>
      </template>
      
      <div class="account-info">
        <div class="info-item">
          <span class="info-label">註冊時間：</span>
          <span class="info-value">{{ formatDate(authStore.user?.created_at) }}</span>
        </div>
        <div class="info-item">
          <span class="info-label">最後更新：</span>
          <span class="info-value">{{ formatDate(authStore.user?.updated_at) }}</span>
        </div>
      </div>

      <div class="danger-zone">
        <h3>危險操作</h3>
        <el-button type="danger" @click="handleDeleteAccount">
          刪除帳號
        </el-button>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import { useAuthStore } from '@/store/auth'
import { updateUser, deleteUser } from '@/api/user'
import type { UpdateUserRequest } from '@/types'

const router = useRouter()
const authStore = useAuthStore()

const loading = ref(false)
const profileFormRef = ref<FormInstance>()

const profileForm = reactive<UpdateUserRequest & { id?: number }>({
  username: '',
  email: '',
  avatar: '',
  bio: ''
})

const profileRules: FormRules = {
  username: [
    { required: true, message: '請輸入用戶名', trigger: 'blur' },
    { min: 3, max: 32, message: '用戶名長度在 3 到 32 個字符', trigger: 'blur' }
  ],
  email: [
    { required: true, message: '請輸入郵箱', trigger: 'blur' },
    { type: 'email', message: '請輸入正確的郵箱格式', trigger: 'blur' }
  ],
  bio: [
    { max: 500, message: '個人簡介不能超過 500 個字符', trigger: 'blur' }
  ]
}

const initForm = () => {
  if (authStore.user) {
    profileForm.id = authStore.user.id
    profileForm.username = authStore.user.username
    profileForm.email = authStore.user.email
    profileForm.avatar = authStore.user.avatar || ''
    profileForm.bio = authStore.user.bio || ''
  }
}

const resetForm = () => {
  initForm()
}

const handleUpdate = async () => {
  if (!profileFormRef.value || !profileForm.id) return

  await profileFormRef.value.validate(async (valid) => {
    if (valid) {
      loading.value = true
      try {
        const updatedUser = await updateUser(profileForm.id!, {
          username: profileForm.username,
          email: profileForm.email,
          avatar: profileForm.avatar,
          bio: profileForm.bio
        })
        
        // 更新本地用戶資料
        const userData = updatedUser
        authStore.updateUser(userData)
        
        ElMessage.success('更新成功')
      } catch (error) {
        console.error('更新失敗:', error)
      } finally {
        loading.value = false
      }
    }
  })
}

const handleDeleteAccount = async () => {
  try {
    await ElMessageBox.confirm(
      '確定要刪除您的帳號嗎？此操作不可恢復！',
      '確認刪除帳號',
      {
        confirmButtonText: '確定刪除',
        cancelButtonText: '取消',
        type: 'error'
      }
    )

    if (authStore.user?.id) {
      await deleteUser(authStore.user.id)
      ElMessage.success('帳號已刪除')
      authStore.logout()
      router.push('/')
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error('刪除帳號失敗:', error)
    }
  }
}

const formatDate = (dateString?: string) => {
  if (!dateString) return '未知'
  return new Date(dateString).toLocaleDateString('zh-TW')
}

onMounted(() => {
  initForm()
})
</script>

<style scoped>
.profile {
  max-width: 800px;
  margin: 0 auto;
}

.profile-header {
  margin-bottom: 24px;
}

.profile-header h1 {
  margin: 0;
  color: #333;
}

.profile-card {
  margin-bottom: 20px;
}

.avatar-section {
  display: flex;
  align-items: flex-start;
  gap: 20px;
}

.avatar-info {
  flex: 1;
}

.account-info {
  margin-bottom: 32px;
}

.info-item {
  display: flex;
  margin-bottom: 12px;
}

.info-label {
  font-weight: bold;
  color: #666;
  width: 100px;
}

.info-value {
  color: #333;
}

.danger-zone {
  border-top: 1px solid #f0f0f0;
  padding-top: 24px;
}

.danger-zone h3 {
  color: #f56c6c;
  margin-bottom: 16px;
}
</style> 