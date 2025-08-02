/// <reference types="vite/client" />

declare module '*.vue' {
  import type { DefineComponent } from 'vue'
  const component: DefineComponent<{}, {}, any>
  export default component
}

interface ImportMetaEnv {
  // 應用程式基本配置
  readonly VITE_APP_TITLE: string
  readonly VITE_APP_VERSION: string
  readonly VITE_APP_ENV: string

  // API 配置
  readonly VITE_API_BASE_URL: string

  // 串流服務配置
  readonly VITE_HLS_BASE_URL: string
  readonly VITE_WS_BASE_URL: string
  readonly VITE_STREAM_PULLER_BASE_URL: string

  // 第三方服務配置
  readonly VITE_AGORA_APP_ID?: string
  readonly VITE_AGORA_TOKEN?: string

  // 功能開關
  readonly VITE_DEBUG_MODE: string
  readonly VITE_ENABLE_DEV_TOOLS: string
  readonly VITE_ENABLE_ERROR_TRACKING: string

  // 開發環境配置
  readonly VITE_DEV_SERVER_PORT: string
  readonly VITE_DEV_HTTPS: string

  // 生產環境配置
  readonly VITE_ENABLE_PWA: string
  readonly VITE_ENABLE_SW: string

  // 監控和分析
  readonly VITE_GA_ID?: string
  readonly VITE_SENTRY_DSN?: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}
