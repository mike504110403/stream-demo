<template>
  <div class="not-found">
    <div class="not-found-content">
      <div class="not-found-icon">🔍</div>
      <h1>404</h1>
      <h2>頁面不存在</h2>
      <p>抱歉，您訪問的頁面不存在或已被刪除</p>
      <div class="not-found-actions">
        <el-button type="primary" @click="$router.push('/')">
          返回首頁
        </el-button>
        <el-button @click="$router.back()">
          返回上頁
        </el-button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.not-found {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: calc(100vh - 60px);
  background-color: #f5f5f5;
}

.not-found-content {
  text-align: center;
  padding: 60px;
  background: white;
  border-radius: 12px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.not-found-icon {
  font-size: 80px;
  margin-bottom: 24px;
}

.not-found-content h1 {
  font-size: 72px;
  font-weight: bold;
  color: #409eff;
  margin: 0 0 16px 0;
}

.not-found-content h2 {
  font-size: 32px;
  color: #333;
  margin: 0 0 16px 0;
}

.not-found-content p {
  font-size: 16px;
  color: #666;
  margin: 0 0 32px 0;
}

.not-found-actions {
  display: flex;
  gap: 16px;
  justify-content: center;
}
</style> 