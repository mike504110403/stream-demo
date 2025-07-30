<template>
  <div class="dashboard">
    <div class="dashboard-header">
      <h1>歡迎回來，{{ authStore.user?.username }}！</h1>
      <p>這是您的個人儀表板</p>
    </div>

    <!-- 快速操作 -->
    <div class="quick-actions">
      <h2>快速操作</h2>
      <el-row :gutter="20">
        <el-col :span="6">
          <el-card class="action-card" @click="$router.push('/videos')">
            <div class="action-content">
              <div class="action-icon">🎥</div>
              <div class="action-title">影片管理</div>
              <div class="action-desc">查看和管理您的影片</div>
            </div>
          </el-card>
        </el-col>
        
        <el-col :span="6">
          <el-card class="action-card" @click="$router.push('/live-rooms')">
            <div class="action-content">
              <div class="action-icon">📺</div>
              <div class="action-title">直播間管理</div>
              <div class="action-desc">管理您的直播間</div>
            </div>
          </el-card>
        </el-col>
        
        <el-col :span="6">
          <el-card class="action-card" @click="$router.push('/payments')">
            <div class="action-content">
              <div class="action-icon">💰</div>
              <div class="action-title">支付記錄</div>
              <div class="action-desc">查看支付交易記錄</div>
            </div>
          </el-card>
        </el-col>
        
        <el-col :span="6">
          <el-card class="action-card" @click="$router.push('/profile')">
            <div class="action-content">
              <div class="action-icon">👤</div>
              <div class="action-title">個人資料</div>
              <div class="action-desc">編輯個人資料</div>
            </div>
          </el-card>
        </el-col>
      </el-row>
    </div>

    <!-- 功能介紹 -->
    <div class="features">
      <h2>平台功能</h2>
      <el-row :gutter="20">
        <el-col :span="8">
          <el-card>
            <h3>🎬 影片上傳</h3>
            <p>上傳和分享您的影片內容，支援多種格式</p>
            <el-button type="primary" @click="$router.push('/videos/upload')">
              立即上傳
            </el-button>
          </el-card>
        </el-col>
        
        <el-col :span="8">
          <el-card>
            <h3>🔴 直播間功能</h3>
            <p>創建直播間，與觀眾實時互動</p>
            <el-button type="primary" @click="$router.push('/live-rooms/create')">
              創建直播間
            </el-button>
          </el-card>
        </el-col>
        
        <el-col :span="8">
          <el-card>
            <h3>💳 支付系統</h3>
            <p>安全便捷的支付功能，支援多種支付方式</p>
            <el-button type="primary" @click="$router.push('/payments/create')">
              創建支付
            </el-button>
          </el-card>
        </el-col>
      </el-row>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useAuthStore } from '@/store/auth'

const authStore = useAuthStore()
</script>

<style scoped>
.dashboard {
  max-width: 1200px;
  margin: 0 auto;
}

.dashboard-header {
  margin-bottom: 32px;
  text-align: center;
}

.dashboard-header h1 {
  margin: 0 0 8px 0;
  font-size: 32px;
  color: #333;
}

.dashboard-header p {
  margin: 0;
  color: #666;
  font-size: 16px;
}

.quick-actions, .features {
  margin-bottom: 32px;
}

.quick-actions h2, .features h2 {
  margin-bottom: 16px;
  color: #333;
}

.action-card {
  cursor: pointer;
  transition: all 0.2s;
  text-align: center;
  height: 100%;
}

.action-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 8px 16px rgba(0, 0, 0, 0.1);
}

.action-content {
  padding: 24px;
}

.action-icon {
  font-size: 48px;
  margin-bottom: 16px;
}

.action-title {
  font-size: 18px;
  font-weight: bold;
  color: #333;
  margin-bottom: 8px;
}

.action-desc {
  color: #666;
  font-size: 14px;
}

.features .el-card {
  text-align: center;
  height: 100%;
}

.features h3 {
  margin-top: 0;
  color: #333;
}

.features p {
  color: #666;
  margin-bottom: 20px;
}
</style>