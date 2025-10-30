import { defineConfig, loadEnv } from 'vite'
import { resolve } from 'path'

// https://vitejs.dev/config/
export default defineConfig(async ({ mode }) => {
  // 載入環境變數
  const env = loadEnv(mode, process.cwd(), '')
  
  // 動態導入 vue plugin
  const { default: vue } = await import('@vitejs/plugin-vue')
  
  return {
    plugins: [vue()],
    resolve: {
      alias: {
        '@': resolve(__dirname, 'src')
      }
    },
    server: {
      host: '0.0.0.0',
      port: parseInt(env.VITE_DEV_SERVER_PORT) || 5173,
      https: env.VITE_DEV_HTTPS === 'true' ? {} : false,
      proxy: {
        '/api': {
          target: 'http://localhost:8084',
          changeOrigin: true,
          secure: false,
          ws: true, // 支援 WebSocket
          configure: (proxy, _options) => {
            proxy.on('error', (err, _req, _res) => {
              console.log('proxy error', err);
            });
            proxy.on('proxyReq', (proxyReq, req, _res) => {
              console.log('Sending Request to the Target:', req.method, req.url);
            });
            proxy.on('proxyRes', (proxyRes, req, _res) => {
              console.log('Received Response from the Target:', proxyRes.statusCode, req.url);
            });
          },
        },
        '/stream-puller': {
          target: env.VITE_STREAM_PULLER_BASE_URL || 'http://localhost:8083',
          changeOrigin: true,
          secure: false,
          rewrite: (path) => path.replace(/^\/stream-puller/, ''),
          configure: (proxy, _options) => {
            proxy.on('error', (err, _req, _res) => {
              console.log('stream-puller proxy error', err);
            });
          },
        },
        // 注意：HLS 代理已移除，現在直接訪問 live-cdn 服務
        '/ws': {
          target: env.VITE_WS_BASE_URL || 'http://localhost:8084',
          changeOrigin: true,
          secure: false,
          ws: true, // 支援 WebSocket
          configure: (proxy, _options) => {
            proxy.on('error', (err, _req, _res) => {
              console.log('ws proxy error', err);
            });
          },
        }
      }
    },
    define: {
      // 在客戶端暴露環境變數
      __APP_VERSION__: JSON.stringify(env.VITE_APP_VERSION || '1.0.0'),
      __APP_ENV__: JSON.stringify(env.VITE_APP_ENV || 'development'),
      __DEBUG_MODE__: JSON.stringify(env.VITE_DEBUG_MODE === 'true'),
    }
  }
})
