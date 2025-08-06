<template>
  <div class="live-create">
    <div class="page-header">
      <h1>創建直播</h1>
      <el-button @click="$router.back()">返回</el-button>
    </div>

    <div class="create-content">
      <el-card class="create-form">
        <template #header>
          <div class="card-header">
            <span>直播資訊</span>
          </div>
        </template>

        <el-form
          ref="formRef"
          :model="formData"
          :rules="formRules"
          label-width="100px"
          @submit.prevent="handleSubmit"
        >
          <el-form-item label="直播標題" prop="title">
            <el-input
              v-model="formData.title"
              placeholder="請輸入直播標題"
              maxlength="100"
              show-word-limit
            />
          </el-form-item>

          <el-form-item label="直播描述" prop="description">
            <el-input
              v-model="formData.description"
              type="textarea"
              :rows="4"
              placeholder="請輸入直播描述（可選）"
              maxlength="500"
              show-word-limit
            />
          </el-form-item>

          <el-form-item label="開始時間" prop="start_time">
            <el-date-picker
              v-model="formData.start_time"
              type="datetime"
              placeholder="選擇直播開始時間"
              format="YYYY-MM-DD HH:mm:ss"
              value-format="YYYY-MM-DD HH:mm:ss"
              :disabled-date="disabledDate"
              :disabled-time="disabledTime"
              style="width: 100%"
            />
          </el-form-item>

          <el-form-item label="聊天功能">
            <el-switch
              v-model="formData.chat_enabled"
              active-text="開啟"
              inactive-text="關閉"
            />
            <div class="form-tip">開啟後觀眾可以在直播間聊天</div>
          </el-form-item>

          <el-form-item>
            <el-button
              type="primary"
              :loading="submitting"
              @click="handleSubmit"
            >
              創建直播
            </el-button>
            <el-button @click="resetForm"> 重置 </el-button>
          </el-form-item>
        </el-form>
      </el-card>

      <!-- 創建指南 -->
      <el-card class="create-guide">
        <template #header>
          <div class="card-header">
            <span>創建指南</span>
          </div>
        </template>

        <div class="guide-content">
          <h3>📺 直播創建步驟</h3>
          <ol>
            <li>填寫直播標題和描述</li>
            <li>設置直播開始時間</li>
            <li>選擇是否開啟聊天功能</li>
            <li>點擊創建直播</li>
            <li>獲取串流金鑰和推流地址</li>
            <li>使用 OBS 等軟體開始推流</li>
          </ol>

          <h3>🔧 推流軟體設置</h3>
          <div class="software-guide">
            <div class="software-item">
              <h4>OBS Studio</h4>
              <p>1. 打開 OBS Studio</p>
              <p>2. 設置 → 串流</p>
              <p>3. 服務選擇「自訂」</p>
              <p>4. 伺服器填入 RTMP 地址</p>
              <p>5. 串流金鑰填入獲得的金鑰</p>
            </div>

            <div class="software-item">
              <h4>手機推流</h4>
              <p>1. 下載 Larix Broadcaster</p>
              <p>2. 添加新的串流</p>
              <p>3. 填入 RTMP 地址和金鑰</p>
              <p>4. 開始推流</p>
            </div>
          </div>

          <h3>⚠️ 注意事項</h3>
          <ul>
            <li>串流金鑰請妥善保管，不要外洩</li>
            <li>建議提前 10 分鐘開始推流測試</li>
            <li>確保網路穩定，建議使用有線網路</li>
            <li>直播開始後可以隨時結束直播</li>
          </ul>
        </div>
      </el-card>
    </div>

    <!-- 創建成功對話框 -->
    <el-dialog
      v-model="showSuccessDialog"
      title="直播創建成功"
      width="600px"
      :close-on-click-modal="false"
    >
      <div class="success-content">
        <el-result
          icon="success"
          title="直播創建成功！"
          sub-title="請保存以下串流資訊"
        >
          <template #extra>
            <div class="stream-info">
              <div class="info-item">
                <label>串流金鑰：</label>
                <div class="key-display">
                  <el-input
                    :model-value="createdLive?.stream_key || ''"
                    readonly
                    size="small"
                  />
                  <el-button type="primary" size="small" @click="copyStreamKey">
                    複製
                  </el-button>
                </div>
              </div>

              <div class="info-item">
                <label>RTMP 推流地址：</label>
                <div class="key-display">
                  <el-input v-model="rtmpUrl" readonly size="small" />
                  <el-button type="primary" size="small" @click="copyRtmpUrl">
                    複製
                  </el-button>
                </div>
              </div>

              <div class="info-item">
                <label>直播間地址：</label>
                <div class="key-display">
                  <el-input v-model="liveRoomUrl" readonly size="small" />
                  <el-button
                    type="primary"
                    size="small"
                    @click="copyLiveRoomUrl"
                  >
                    複製
                  </el-button>
                </div>
              </div>
            </div>
          </template>
        </el-result>
      </div>

      <template #footer>
        <div class="dialog-footer">
          <el-button @click="showSuccessDialog = false"> 關閉 </el-button>
          <el-button type="primary" @click="goToLiveRoom">
            進入直播間
          </el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from "vue";
import { useRouter } from "vue-router";
import { ElMessage } from "element-plus";
import type { FormInstance, FormRules } from "element-plus";
import { createLive } from "@/api/live";
import type { Live, CreateLiveRequest } from "@/types";
import { getRtmpPushUrl } from "@/utils/stream-config";

const router = useRouter();

// 響應式數據
const formRef = ref<FormInstance>();
const submitting = ref(false);
const showSuccessDialog = ref(false);
const createdLive = ref<Live | null>(null);

// 表單數據
const formData = ref<CreateLiveRequest & { chat_enabled: boolean }>({
  title: "",
  description: "",
  start_time: "",
  chat_enabled: true,
});

// 表單驗證規則
const formRules: FormRules = {
  title: [
    { required: true, message: "請輸入直播標題", trigger: "blur" },
    {
      min: 2,
      max: 100,
      message: "標題長度在 2 到 100 個字符",
      trigger: "blur",
    },
  ],
  start_time: [
    { required: true, message: "請選擇開始時間", trigger: "change" },
  ],
};

// 計算屬性
const rtmpUrl = computed(() => {
  if (!createdLive.value) return "";
  return getRtmpPushUrl(createdLive.value.stream_key);
});

const liveRoomUrl = computed(() => {
  if (!createdLive.value) return "";
  return `${window.location.origin}/lives/${createdLive.value.id}/stream`;
});

// 提交表單
const handleSubmit = async () => {
  if (!formRef.value) return;

  try {
    await formRef.value.validate();
  } catch (error) {
    return;
  }

  submitting.value = true;

  try {
    const response = await createLive({
      title: formData.value.title,
      description: formData.value.description,
      start_time: formData.value.start_time,
    });

    createdLive.value = response;
    showSuccessDialog.value = true;
    ElMessage.success("直播創建成功！");
  } catch (error: any) {
    console.error("創建直播失敗:", error);
    ElMessage.error(error.message || "創建直播失敗");
  } finally {
    submitting.value = false;
  }
};

// 重置表單
const resetForm = () => {
  if (formRef.value) {
    formRef.value.resetFields();
  }
  formData.value = {
    title: "",
    description: "",
    start_time: "",
    chat_enabled: true,
  };
};

// 複製功能
const copyToClipboard = async (text: string, label: string) => {
  try {
    await navigator.clipboard.writeText(text);
    ElMessage.success(`${label} 已複製到剪貼簿`);
  } catch (err) {
    console.error("複製失敗:", err);
    ElMessage.error("複製失敗");
  }
};

const copyStreamKey = () => {
  if (createdLive.value?.stream_key) {
    copyToClipboard(createdLive.value.stream_key, "串流金鑰");
  }
};

const copyRtmpUrl = () => {
  copyToClipboard(rtmpUrl.value, "RTMP 推流地址");
};

const copyLiveRoomUrl = () => {
  copyToClipboard(liveRoomUrl.value, "直播間地址");
};

// 進入直播間
const goToLiveRoom = () => {
  if (createdLive.value) {
    router.push(`/lives/${createdLive.value.id}/stream`);
  }
};

// 日期限制
const disabledDate = (time: Date) => {
  return time.getTime() < Date.now() - 8.64e7; // 不能選擇過去的日期
};

const disabledTime = (date: Date) => {
  if (date) {
    const now = new Date();
    const selectedDate = new Date(date);

    // 如果是今天，限制時間不能早於當前時間
    if (selectedDate.toDateString() === now.toDateString()) {
      return {
        disabledHours: () =>
          Array.from({ length: now.getHours() }, (_, i) => i),
        disabledMinutes: (hour: number) => {
          if (hour === now.getHours()) {
            return Array.from({ length: now.getMinutes() }, (_, i) => i);
          }
          return [];
        },
      };
    }
  }
  return {};
};
</script>

<style scoped>
.live-create {
  max-width: 1200px;
  margin: 0 auto;
  padding: 20px;
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

.create-content {
  display: grid;
  grid-template-columns: 1fr 400px;
  gap: 24px;
}

.create-form {
  min-height: 500px;
}

.card-header {
  font-weight: bold;
  color: #333;
}

.form-tip {
  margin-top: 8px;
  font-size: 12px;
  color: #999;
}

.create-guide {
  height: fit-content;
}

.guide-content h3 {
  margin: 20px 0 12px 0;
  color: #333;
  font-size: 16px;
}

.guide-content h3:first-child {
  margin-top: 0;
}

.guide-content ol,
.guide-content ul {
  margin: 0 0 16px 0;
  padding-left: 20px;
}

.guide-content li {
  margin-bottom: 8px;
  line-height: 1.5;
}

.software-guide {
  display: flex;
  flex-direction: column;
  gap: 16px;
  margin-bottom: 16px;
}

.software-item {
  background: #f5f5f5;
  padding: 12px;
  border-radius: 6px;
}

.software-item h4 {
  margin: 0 0 8px 0;
  color: #333;
  font-size: 14px;
}

.software-item p {
  margin: 4px 0;
  font-size: 12px;
  color: #666;
}

.success-content {
  padding: 20px 0;
}

.stream-info {
  display: flex;
  flex-direction: column;
  gap: 16px;
  margin-top: 20px;
}

.info-item {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.info-item label {
  font-weight: bold;
  color: #333;
  font-size: 14px;
}

.key-display {
  display: flex;
  gap: 8px;
  align-items: center;
}

.key-display .el-input {
  flex: 1;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

/* 響應式設計 */
@media (max-width: 1000px) {
  .create-content {
    grid-template-columns: 1fr;
    gap: 16px;
  }
}

@media (max-width: 768px) {
  .live-create {
    padding: 12px;
  }

  .page-header {
    flex-direction: column;
    gap: 12px;
    align-items: stretch;
  }

  .page-header h1 {
    text-align: center;
  }

  .create-form {
    min-height: auto;
  }
}
</style>
