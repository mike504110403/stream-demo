// 動態配置工具
export const getServerConfig = () => {
  // 檢查是否在 ngrok 環境中
  const isNgrok = window.location.hostname.includes('ngrok') || 
                  window.location.hostname.includes('tcp.jp.ngrok.io') ||
                  window.location.hostname.includes('tcp.ngrok.io')
  
  if (isNgrok) {
    // 在 ngrok 環境中，使用相對路徑或動態獲取當前域名
    const protocol = window.location.protocol
    const hostname = window.location.hostname
    // const port = window.location.port  // 暫時註釋掉未使用的變數
    
    // 根據當前端口推斷其他服務的端口
    // const currentPort = parseInt(port) || 5173  // 暫時註釋掉未使用的變數
    
    // 假設後端 API 在 8080，HLS 在 8082，WebSocket 在 8080
    // 這些需要根據實際的 ngrok 配置調整
    const apiPort = 8080
    const hlsPort = 8082
    const wsPort = 8080
    
    return {
      apiBaseUrl: `${protocol}//${hostname}:${apiPort}/api`,
      hlsBaseUrl: `${protocol}//${hostname}:${hlsPort}`,
      wsBaseUrl: `${protocol === 'https:' ? 'wss:' : 'ws:'}//${hostname}:${wsPort}`,
      streamPullerBaseUrl: `${protocol}//${hostname}:8083`
    }
  } else {
    // 本地開發環境 - 統一使用相對路徑
    return {
      apiBaseUrl: '/api',
      hlsBaseUrl: '/hls',
      wsBaseUrl: '/ws',  // 使用相對路徑
      streamPullerBaseUrl: '/stream-puller'
    }
  }
}

// 獲取當前配置
export const config = getServerConfig() 