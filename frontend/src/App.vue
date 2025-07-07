<template>
  <div id="app">
    <!-- 導航欄 -->
    <el-container>
      <el-header v-if="showNavbar">
        <NavBar />
      </el-header>
      
      <!-- 主要內容區域 -->
      <el-main>
        <router-view />
      </el-main>
    </el-container>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import NavBar from '@/components/NavBar.vue'

const route = useRoute()

// 在某些頁面不顯示導航欄
const showNavbar = computed(() => {
  const hideNavbarRoutes = ['Login', 'Register']
  return !hideNavbarRoutes.includes(route.name as string)
})
</script>

<style>
#app {
  font-family: Avenir, Helvetica, Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  min-height: 100vh;
}

.el-header {
  padding: 0;
  background-color: #fff;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.el-main {
  padding: 20px;
  background-color: #f5f5f5;
  min-height: calc(100vh - 60px);
}
</style>
