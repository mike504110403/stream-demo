import { createApp } from 'vue'
import { createPinia } from 'pinia'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import App from './App.vue'
import router from './router'
import { useAuthStore } from './store/auth'

// 創建應用實例
const app = createApp(App)

// 使用插件
const pinia = createPinia()
app.use(pinia)
app.use(router)
app.use(ElementPlus)

// 初始化認證狀態
const authStore = useAuthStore()
console.log('應用啟動 - 認證狀態:', authStore.isAuthenticated) // 調試用

// 掛載應用
app.mount('#app')
