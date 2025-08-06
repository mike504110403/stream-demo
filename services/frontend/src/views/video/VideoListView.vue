<template>
  <div class="video-list">
    <div class="page-header">
      <h1>影片管理</h1>
      <el-button type="primary" @click="$router.push('/videos/upload')">
        上傳影片
      </el-button>
    </div>

    <!-- 搜尋和篩選 -->
    <div class="search-section">
      <el-row :gutter="20">
        <el-col :span="16">
          <el-input
            v-model="searchQuery"
            placeholder="搜尋影片標題..."
            @keyup.enter="handleSearch"
            clearable
          >
            <template #append>
              <el-button @click="handleSearch">搜尋</el-button>
            </template>
          </el-input>
        </el-col>
        <el-col :span="8">
          <el-select
            v-model="statusFilter"
            placeholder="篩選狀態"
            @change="loadVideos"
          >
            <el-option label="全部" value="" />
            <el-option label="上傳中" value="uploading" />
            <el-option label="轉碼中" value="transcoding" />
            <el-option label="處理中" value="processing" />
            <el-option label="已完成" value="ready" />
            <el-option label="失敗" value="failed" />
          </el-select>
        </el-col>
      </el-row>

      <!-- 轉碼狀態提示 -->
      <el-alert
        v-if="hasProcessingVideos"
        title="轉碼提示"
        type="info"
        :closable="false"
        show-icon
        style="margin-top: 16px"
      >
        <template #default>
          有影片正在轉碼中，請定期刷新頁面查看最新狀態。轉碼完成後影片將自動出現在列表中。
        </template>
      </el-alert>
    </div>

    <!-- 影片列表 -->
    <div class="video-grid" v-loading="loading">
      <div v-if="videos.length === 0 && !loading" class="empty-state">
        <div class="empty-icon">🎬</div>
        <div class="empty-text">暫無影片</div>
        <el-button type="primary" @click="$router.push('/videos/upload')">
          上傳第一個影片
        </el-button>
      </div>

      <el-row :gutter="20" v-else>
        <el-col :span="6" v-for="video in videos" :key="video.id">
          <el-card class="video-card" @click="viewVideo(Number(video.id))">
            <div class="video-thumbnail">
              <img
                v-if="video.thumbnail_url"
                :src="video.thumbnail_url"
                :alt="video.title"
              />
              <div v-else class="placeholder-thumbnail">
                <div class="placeholder-icon">🎬</div>
              </div>
              <div class="video-status">
                <el-tag :type="getStatusType(video.status)" size="small">
                  {{ getStatusText(video.status) }}
                </el-tag>
              </div>
            </div>

            <div class="video-info">
              <h3 class="video-title">{{ video.title }}</h3>
              <p class="video-description">
                {{ video.description || "暫無描述" }}
              </p>

              <div class="video-stats">
                <span class="stat">
                  <el-icon><View /></el-icon>
                  {{ video.views }}
                </span>
                <span class="stat">
                  <el-icon><Star /></el-icon>
                  {{ video.likes }}
                </span>
              </div>

              <div class="video-date">
                {{ formatDate(video.created_at) }}
              </div>
            </div>

            <div class="video-actions" @click.stop>
              <el-button size="small" @click="editVideo(video)">編輯</el-button>
              <el-button
                size="small"
                type="danger"
                @click="deleteVideo(video.id)"
                >刪除</el-button
              >
            </div>
          </el-card>
        </el-col>
      </el-row>
    </div>

    <!-- 編輯對話框 -->
    <el-dialog v-model="editDialogVisible" title="編輯影片" width="500px">
      <el-form :model="editForm" :rules="editRules" ref="editFormRef">
        <el-form-item label="標題" prop="title">
          <el-input v-model="editForm.title" />
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input v-model="editForm.description" type="textarea" :rows="3" />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="editDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleUpdate" :loading="updating">
          確定
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive, computed } from "vue";
import { useRouter } from "vue-router";
import {
  ElMessage,
  ElMessageBox,
  type FormInstance,
  type FormRules,
} from "element-plus";
import { View, Star } from "@element-plus/icons-vue";
import {
  getVideos,
  searchVideos,
  updateVideo,
  deleteVideo as deleteVideoApi,
} from "@/api/video";
import type { Video, UpdateVideoRequest } from "@/types";

const router = useRouter();

const loading = ref(false);
const updating = ref(false);
const videos = ref<Video[]>([]);
const searchQuery = ref("");
const statusFilter = ref("");

const editDialogVisible = ref(false);
const editFormRef = ref<FormInstance>();
const editForm = reactive<UpdateVideoRequest & { id?: number }>({
  title: "",
  description: "",
});

const editRules: FormRules = {
  title: [
    { required: true, message: "請輸入標題", trigger: "blur" },
    {
      min: 1,
      max: 100,
      message: "標題長度在 1 到 100 個字符",
      trigger: "blur",
    },
  ],
};

// 計算是否有正在處理的影片
const hasProcessingVideos = computed(() => {
  return videos.value.some((video) =>
    ["uploading", "transcoding", "processing"].includes(video.status),
  );
});

const loadVideos = async () => {
  loading.value = true;
  try {
    const response = await getVideos();
    console.log("API 響應:", response); // 調試用

    // 處理後端 ListResponse 結構: {total: number, items: Video[]}
    // request.ts 攔截器已經提取了 data，所以 response 就是實際數據
    const result = response;
    let filteredVideos: Video[] = [];

    if (result && typeof result === "object") {
      // 如果有 items 字段，說明是 ListResponse 結構
      if ("items" in result && Array.isArray(result.items)) {
        filteredVideos = result.items;
      }
      // 如果直接是數組
      else if (Array.isArray(result)) {
        filteredVideos = result;
      }
    }

    // 狀態篩選
    if (statusFilter.value) {
      filteredVideos = filteredVideos.filter(
        (video: Video) => video.status === statusFilter.value,
      );
    }

    videos.value = filteredVideos;
    console.log("處理後的影片列表:", filteredVideos); // 調試用
  } catch (error) {
    console.error("載入影片失敗:", error);
    ElMessage.error("載入影片失敗");
  } finally {
    loading.value = false;
  }
};

const handleSearch = async () => {
  if (!searchQuery.value.trim()) {
    loadVideos();
    return;
  }

  loading.value = true;
  try {
    const response = await searchVideos({ q: searchQuery.value });
    console.log("搜尋 API 響應:", response); // 調試用

    // 處理搜尋結果
    // request.ts 攔截器已經提取了 data，所以 response 就是實際數據
    const result = response;
    let searchResults: Video[] = [];

    if (result && typeof result === "object") {
      // 如果有 items 字段，說明是 ListResponse 結構
      if ("items" in result && Array.isArray(result.items)) {
        searchResults = result.items;
      }
      // 如果直接是數組
      else if (Array.isArray(result)) {
        searchResults = result;
      }
    }

    videos.value = searchResults;
  } catch (error) {
    console.error("搜尋影片失敗:", error);
    ElMessage.error("搜尋影片失敗");
  } finally {
    loading.value = false;
  }
};

const viewVideo = (id: number) => {
  router.push(`/videos/${id.toString()}`);
};

const editVideo = (video: Video) => {
  editForm.id = video.id;
  editForm.title = video.title;
  editForm.description = video.description || "";
  editDialogVisible.value = true;
};

const handleUpdate = async () => {
  if (!editFormRef.value || !editForm.id) return;

  await editFormRef.value.validate(async (valid) => {
    if (valid) {
      updating.value = true;
      try {
        await updateVideo(editForm.id!, {
          title: editForm.title,
          description: editForm.description,
        });
        ElMessage.success("更新成功");
        editDialogVisible.value = false;
        loadVideos();
      } catch (error) {
        console.error("更新影片失敗:", error);
      } finally {
        updating.value = false;
      }
    }
  });
};

const deleteVideo = async (id: number) => {
  try {
    await ElMessageBox.confirm("確定要刪除這個影片嗎？", "確認刪除", {
      confirmButtonText: "確定",
      cancelButtonText: "取消",
      type: "warning",
    });

    await deleteVideoApi(id);
    ElMessage.success("刪除成功");
    loadVideos();
  } catch (error) {
    if (error !== "cancel") {
      console.error("刪除影片失敗:", error);
    }
  }
};

const getStatusType = (status: string) => {
  switch (status) {
    case "ready":
      return "success";
    case "uploading":
      return "info";
    case "transcoding":
      return "warning";
    case "processing":
      return "warning";
    case "failed":
      return "danger";
    default:
      return "info";
  }
};

const getStatusText = (status: string) => {
  switch (status) {
    case "ready":
      return "已完成";
    case "uploading":
      return "上傳中";
    case "transcoding":
      return "轉碼中";
    case "processing":
      return "處理中";
    case "failed":
      return "失敗";
    default:
      return status;
  }
};

const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleDateString("zh-TW");
};

onMounted(() => {
  loadVideos();
});
</script>

<style scoped>
.video-list {
  max-width: 1200px;
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

.search-section {
  margin-bottom: 24px;
}

.video-grid {
  min-height: 400px;
}

.empty-state {
  text-align: center;
  padding: 80px 0;
}

.empty-icon {
  font-size: 64px;
  margin-bottom: 16px;
}

.empty-text {
  font-size: 18px;
  color: #666;
  margin-bottom: 24px;
}

.video-card {
  margin-bottom: 20px;
  cursor: pointer;
  transition: all 0.2s;
}

.video-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.video-thumbnail {
  position: relative;
  height: 160px;
  overflow: hidden;
  border-radius: 4px;
  margin-bottom: 12px;
}

.video-thumbnail img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.placeholder-thumbnail {
  width: 100%;
  height: 100%;
  background-color: #f5f5f5;
  display: flex;
  align-items: center;
  justify-content: center;
}

.placeholder-icon {
  font-size: 48px;
  color: #ccc;
}

.video-status {
  position: absolute;
  top: 8px;
  right: 8px;
}

.video-info {
  margin-bottom: 12px;
}

.video-title {
  margin: 0 0 8px 0;
  font-size: 16px;
  font-weight: bold;
  color: #333;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.video-description {
  margin: 0 0 8px 0;
  color: #666;
  font-size: 14px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.video-stats {
  display: flex;
  gap: 16px;
  margin-bottom: 8px;
}

.stat {
  display: flex;
  align-items: center;
  gap: 4px;
  color: #666;
  font-size: 14px;
}

.video-date {
  color: #999;
  font-size: 12px;
}

.video-actions {
  display: flex;
  gap: 8px;
}
</style>
