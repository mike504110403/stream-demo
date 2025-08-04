<template>
  <div class="upload-video">
    <div class="page-header">
      <h1>上傳影片</h1>
      <el-button @click="$router.back()">返回</el-button>
    </div>

    <el-card class="upload-card">
      <el-form
        ref="uploadFormRef"
        :model="uploadForm"
        :rules="uploadRules"
        label-width="100px"
      >
        <el-form-item label="標題" prop="title">
          <el-input v-model="uploadForm.title" placeholder="請輸入影片標題" />
        </el-form-item>

        <el-form-item label="描述" prop="description">
          <el-input
            v-model="uploadForm.description"
            type="textarea"
            :rows="4"
            placeholder="請輸入影片描述"
          />
        </el-form-item>

        <el-form-item label="影片檔案">
          <el-upload
            class="upload-demo"
            drag
            action="#"
            :auto-upload="false"
            :on-change="handleFileChange"
            :limit="1"
            accept="video/*"
            :disabled="uploading"
          >
            <el-icon class="el-icon--upload"><upload-filled /></el-icon>
            <div class="el-upload__text">
              將檔案拖到此處，或<em>點擊上傳</em>
            </div>
            <template #tip>
              <div class="el-upload__tip">
                支援 mp4, avi, mov 等格式，檔案大小不超過 500MB
              </div>
            </template>
          </el-upload>
        </el-form-item>

        <!-- 上傳進度顯示 -->
        <el-form-item v-if="uploadProgress.percent > 0" label="上傳進度">
          <div class="progress-container">
            <el-progress
              :percentage="uploadProgress.percent"
              :status="
                uploadProgress.status === 'error'
                  ? 'exception'
                  : uploadProgress.status === 'success'
                    ? 'success'
                    : undefined
              "
            />
            <div class="progress-message">{{ uploadProgress.message }}</div>
          </div>
        </el-form-item>

        <el-form-item>
          <el-button
            type="primary"
            @click="handleUpload"
            :loading="uploading"
            :disabled="!selectedFile || uploading"
          >
            {{ uploading ? '上傳中...' : '開始上傳' }}
          </el-button>
          <el-button @click="resetForm" :disabled="uploading">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import {
  ElMessage,
  type FormInstance,
  type FormRules,
  type UploadFile,
} from 'element-plus'
import { UploadFilled } from '@element-plus/icons-vue'
import { uploadVideoSeparately, type UploadProgress } from '@/utils/upload'

const router = useRouter()

const uploading = ref(false)
const uploadFormRef = ref<FormInstance>()
const selectedFile = ref<File | null>(null)

const uploadForm = reactive({
  title: '',
  description: '',
})

const uploadProgress = reactive<UploadProgress>({
  percent: 0,
  message: '',
  status: 'pending',
})

const uploadRules: FormRules = {
  title: [
    { required: true, message: '請輸入標題', trigger: 'blur' },
    {
      min: 1,
      max: 100,
      message: '標題長度在 1 到 100 個字符',
      trigger: 'blur',
    },
  ],
}

const handleFileChange = (file: UploadFile) => {
  selectedFile.value = file.raw || null

  // 檔案大小檢查（500MB）
  if (selectedFile.value && selectedFile.value.size > 500 * 1024 * 1024) {
    ElMessage.error('檔案大小不能超過 500MB')
    selectedFile.value = null
    return
  }

  console.log(
    '選擇檔案:',
    selectedFile.value?.name,
    '大小:',
    selectedFile.value?.size
  )
}

const resetForm = () => {
  uploadForm.title = ''
  uploadForm.description = ''
  selectedFile.value = null
  uploadProgress.percent = 0
  uploadProgress.message = ''
  uploadProgress.status = 'pending'
  uploadFormRef.value?.resetFields()
}

const handleUpload = async () => {
  if (!uploadFormRef.value) return

  await uploadFormRef.value.validate(async valid => {
    if (valid) {
      if (!selectedFile.value) {
        ElMessage.warning('請選擇要上傳的影片檔案')
        return
      }

      uploading.value = true

      try {
        console.log('開始分離式上傳...', {
          title: uploadForm.title,
          description: uploadForm.description,
          filename: selectedFile.value.name,
          fileSize: selectedFile.value.size,
        })

        // 使用分離式上傳
        const result = await uploadVideoSeparately(
          selectedFile.value,
          {
            title: uploadForm.title,
            description: uploadForm.description,
          },
          progress => {
            // 更新進度
            Object.assign(uploadProgress, progress)
            console.log('上傳進度:', progress)
          }
        )

        console.log('上傳成功:', result)
        ElMessage.success('影片上傳成功！轉碼處理已開始，請稍後查看影片列表')

        // 延遲跳轉，讓用戶看到成功訊息
        setTimeout(() => {
          router.push('/videos')
        }, 3000)
      } catch (error) {
        console.error('上傳失敗:', error)
        ElMessage.error(uploadProgress.message || '上傳失敗，請重試')
      } finally {
        uploading.value = false
      }
    }
  })
}
</script>

<style scoped>
.upload-video {
  max-width: 800px;
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

.upload-card {
  padding: 20px;
}

.upload-demo {
  width: 100%;
}

.progress-container {
  width: 100%;
}

.progress-message {
  margin-top: 8px;
  font-size: 14px;
  color: #666;
}
</style>
