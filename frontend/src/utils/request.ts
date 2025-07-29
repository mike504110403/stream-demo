import axios from 'axios'
import type { AxiosInstance, AxiosResponse, AxiosError } from 'axios'
import { ElMessage } from 'element-plus'

// 自定義 request 接口，返回提取後的數據
interface CustomAxiosInstance {
  get<T = any>(url: string, config?: any): Promise<T>
  post<T = any>(url: string, data?: any, config?: any): Promise<T>
  put<T = any>(url: string, data?: any, config?: any): Promise<T>
  delete<T = any>(url: string, config?: any): Promise<T>
}

// 創建 axios 實例
const axiosInstance: AxiosInstance = axios.create({
  baseURL: import.meta.env.DEV ? '/api' : (import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api'),
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// 請求攔截器
axiosInstance.interceptors.request.use(
  async (config) => {
    // 動態導入 useAuthStore 避免循環依賴
    const { useAuthStore } = await import('@/store/auth')
    const authStore = useAuthStore()
    if (authStore.token) {
      config.headers.Authorization = `Bearer ${authStore.token}`
    }
    return config
  },
  (error: AxiosError) => {
    console.error('Request error:', error)
    return Promise.reject(error)
  }
)

// 響應攔截器
axiosInstance.interceptors.response.use(
  (response: AxiosResponse) => {
    const res = response.data
    console.log('API 成功響應:', response.status, res) // 調試用
    
    // 檢查是否為統一的響應格式 {code, message, data} 或 {success, data} 或 {message, data}
    if (res && typeof res === 'object') {
      if ('code' in res) {
        // 格式: {code, message, data}
        if (res.code === 200) {
          return res.data || {}
        } else {
          const message = res.message || '請求失敗'
          ElMessage.error(message)
          return Promise.reject(new Error(message))
        }
      } else if ('success' in res) {
        // 格式: {success, data}
        if (res.success) {
          return res.data
        } else {
          const message = res.message || '請求失敗'
          ElMessage.error(message)
          return Promise.reject(new Error(message))
        }
      } else if ('data' in res && 'message' in res) {
        // 格式: {message, data} - 直播間 API 使用的格式
        return res.data
      } else if ('data' in res && 'code' in res) {
        // 格式: {code, message, data} - 用戶 API 使用的格式
        return res.data
      } else if ('message' in res && Object.keys(res).length === 2) {
        // 格式: {message, xxx} - 其他 API 使用的格式（如 {message, room_id}）
        // 返回除了 message 之外的其他字段
        const { message, ...otherFields } = res
        return otherFields
      }
    }
    
    // 如果不是統一格式，直接返回原始數據
    return res
  },
  async (error: AxiosError) => {
    console.error('API 錯誤響應:', error) // 調試用
    
    // 處理不同類型的錯誤
    if (error.response) {
      // 服務器返回了錯誤狀態碼
      const status = error.response.status
      const data = error.response.data as any
      
      console.error('錯誤狀態碼:', status)
      console.error('錯誤數據:', data)
      
      // 直接顯示後端返回的 message
      const message = data?.message || `請求失敗 (${status})`
      
      ElMessage.error(message)
      
      return Promise.reject(error)
    } else if (error.request) {
      // 請求已發送但沒有收到響應
      console.error('網路錯誤 - 沒有收到響應:', error.request)
      ElMessage.error('網路連接失敗，請檢查網路狀態')
      return Promise.reject(new Error('網路連接失敗'))
    } else {
      // 其他錯誤
      console.error('請求配置錯誤:', error.message)
      ElMessage.error('請求失敗: ' + error.message)
      return Promise.reject(error)
    }
  }
)

// 創建自定義 request 對象，具有正確的類型
const request: CustomAxiosInstance = axiosInstance as any

export default request
