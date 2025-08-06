<template>
  <div class="home">
    <div class="hero-section">
      <div class="hero-content">
        <h1 class="hero-title">🎬 歡迎來到串流平台</h1>
        <p class="hero-subtitle">分享您的精彩內容，與世界連接</p>
        <div class="hero-actions">
          <el-button
            v-if="!authStore.isAuthenticated"
            type="primary"
            size="large"
            @click="$router.push('/register')"
          >
            立即註冊
          </el-button>
          <el-button
            v-if="!authStore.isAuthenticated"
            size="large"
            @click="$router.push('/login')"
          >
            登入
          </el-button>
          <el-button
            v-if="authStore.isAuthenticated"
            type="primary"
            size="large"
            @click="$router.push('/dashboard')"
          >
            進入儀表板
          </el-button>
        </div>
      </div>
    </div>

    <div class="features-section">
      <div class="container">
        <h2 class="section-title">平台特色</h2>
        <el-row :gutter="32">
          <el-col :span="8">
            <div class="feature-card">
              <div class="feature-icon">🎥</div>
              <h3>影片上傳</h3>
              <p>輕鬆上傳您的影片內容，支援多種格式，快速處理</p>
            </div>
          </el-col>
          <el-col :span="8">
            <div class="feature-card">
              <div class="feature-icon">📺</div>
              <h3>直播功能</h3>
              <p>創建直播間，與觀眾實時互動，分享您的精彩時刻</p>
            </div>
          </el-col>
          <el-col :span="8">
            <div class="feature-card">
              <div class="feature-icon">💰</div>
              <h3>支付系統</h3>
              <p>安全便捷的支付功能，支援多種支付方式</p>
            </div>
          </el-col>
        </el-row>
      </div>
    </div>

    <div class="stats-section">
      <div class="container">
        <h2 class="section-title">平台數據</h2>
        <el-row :gutter="32">
          <el-col :span="6">
            <div class="stat-item">
              <div class="stat-number">1000+</div>
              <div class="stat-label">註冊用戶</div>
            </div>
          </el-col>
          <el-col :span="6">
            <div class="stat-item">
              <div class="stat-number">5000+</div>
              <div class="stat-label">影片數量</div>
            </div>
          </el-col>
          <el-col :span="6">
            <div class="stat-item">
              <div class="stat-number">100+</div>
              <div class="stat-label">直播場次</div>
            </div>
          </el-col>
          <el-col :span="6">
            <div class="stat-item">
              <div class="stat-number">24/7</div>
              <div class="stat-label">服務時間</div>
            </div>
          </el-col>
        </el-row>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useAuthStore } from "@/store/auth";

const authStore = useAuthStore();
</script>

<style scoped>
.home {
  min-height: calc(100vh - 60px);
}

.hero-section {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  padding: 120px 0;
  text-align: center;
}

.hero-title {
  font-size: 48px;
  font-weight: bold;
  margin-bottom: 16px;
}

.hero-subtitle {
  font-size: 20px;
  margin-bottom: 32px;
  opacity: 0.9;
}

.hero-actions {
  display: flex;
  gap: 16px;
  justify-content: center;
}

.features-section,
.stats-section {
  padding: 80px 0;
}

.features-section {
  background-color: #f8f9fa;
}

.container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 20px;
}

.section-title {
  text-align: center;
  font-size: 36px;
  margin-bottom: 48px;
  color: #333;
}

.feature-card {
  text-align: center;
  padding: 40px 20px;
  background: white;
  border-radius: 12px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  transition: transform 0.2s;
}

.feature-card:hover {
  transform: translateY(-4px);
}

.feature-icon {
  font-size: 64px;
  margin-bottom: 24px;
}

.feature-card h3 {
  font-size: 24px;
  margin-bottom: 16px;
  color: #333;
}

.feature-card p {
  color: #666;
  line-height: 1.6;
}

.stat-item {
  text-align: center;
}

.stat-number {
  font-size: 48px;
  font-weight: bold;
  color: #409eff;
  margin-bottom: 8px;
}

.stat-label {
  font-size: 18px;
  color: #666;
}
</style>
