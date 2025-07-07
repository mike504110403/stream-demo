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
          >
            <el-icon class="el-icon--upload"><upload-filled /></el-icon>
            <div class="el-upload__text">
              將檔案拖到此處，或<em>點擊上傳</em>
            </div>
            <template #tip>
              <div class="el-upload__tip">
                支援 mp4, avi, mov 等格式，檔案大小不超過 100MB
              </div>
            </template>
          </el-upload>
        </el-form-item>

        <el-form-item>
          <el-button type="primary" @click="handleUpload" :loading="uploading">
            上傳影片
          </el-button>
          <el-button @click="resetForm">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, type FormInstance, type FormRules, type UploadFile } from 'element-plus'
import { UploadFilled } from '@element-plus/icons-vue'
import { uploadVideo } from '@/api/video'
import type { UploadVideoRequest } from '@/types'

const router = useRouter()

const uploading = ref(false)
const uploadFormRef = ref<FormInstance>()
const selectedFile = ref<File | null>(null)

const uploadForm = reactive<UploadVideoRequest>({
  title: '',
  description: ''
})

const uploadRules: FormRules = {
  title: [
    { required: true, message: '請輸入標題', trigger: 'blur' },
    { min: 1, max: 100, message: '標題長度在 1 到 100 個字符', trigger: 'blur' }
  ]
}

const handleFileChange = (file: UploadFile) => {
  selectedFile.value = file.raw || null
}

const resetForm = () => {
  uploadForm.title = ''
  uploadForm.description = ''
  selectedFile.value = null
  uploadFormRef.value?.resetFields()
}

const handleUpload = async () => {
  if (!uploadFormRef.value) return

  await uploadFormRef.value.validate(async (valid) => {
    if (valid) {
      if (!selectedFile.value) {
        ElMessage.warning('請選擇要上傳的影片檔案')
        return
      }

      uploading.value = true
      try {
        // 注意：這裡只是模擬上傳，實際需要處理檔案上傳
        await uploadVideo(uploadForm)
        ElMessage.success('影片上傳成功！正在處理中...')
        router.push('/videos')
      } catch (error) {
        console.error('上傳失敗:', error)
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
</style> 