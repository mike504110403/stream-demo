// 動態配置工具
export const getServerConfig = () => {
  // 檢查是否在 ngrok 環境中
  const isNgrok =
    window.location.hostname.includes("ngrok") ||
    window.location.hostname.includes("tcp.jp.ngrok.io") ||
    window.location.hostname.includes("tcp.ngrok.io");

  if (isNgrok) {
    // 在 ngrok 環境中，使用相對路徑或動態獲取當前域名
    const protocol = window.location.protocol;
    const hostname = window.location.hostname;

    // 根據當前端口推斷其他服務的端口
    const apiPort = 8080;
    const hlsPort = 8082;
    const wsPort = 8080;

    return {
      apiBaseUrl: `${protocol}//${hostname}:${apiPort}/api`,
      hlsBaseUrl: `${protocol}//${hostname}:${hlsPort}`,
      wsBaseUrl: `${protocol === "https:" ? "wss:" : "ws:"}//${hostname}:${wsPort}`,
      streamPullerBaseUrl: `${protocol}//${hostname}:8083`,
    };
  } else {
    // 使用環境變數配置，如果沒有則使用預設值
    const apiBaseUrl = import.meta.env.VITE_API_BASE_URL || "/api";
    const hlsBaseUrl = import.meta.env.VITE_HLS_BASE_URL || "/hls";
    const wsBaseUrl = import.meta.env.VITE_WS_BASE_URL || "/ws";
    const streamPullerBaseUrl =
      import.meta.env.VITE_STREAM_PULLER_BASE_URL || "/stream-puller";

    return {
      apiBaseUrl,
      hlsBaseUrl,
      wsBaseUrl,
      streamPullerBaseUrl,
    };
  }
};

// 獲取應用程式基本配置
export const getAppConfig = () => {
  return {
    title: import.meta.env.VITE_APP_TITLE || "Stream Demo Platform",
    version: import.meta.env.VITE_APP_VERSION || "1.0.0",
    env: import.meta.env.VITE_APP_ENV || "development",
    debugMode: import.meta.env.VITE_DEBUG_MODE === "true",
    enableDevTools: import.meta.env.VITE_ENABLE_DEV_TOOLS === "true",
    enableErrorTracking: import.meta.env.VITE_ENABLE_ERROR_TRACKING === "true",
  };
};

// 獲取第三方服務配置
export const getThirdPartyConfig = () => {
  return {
    // 如果有第三方串流服務配置
    agora: {
      appId: import.meta.env.VITE_AGORA_APP_ID,
      token: import.meta.env.VITE_AGORA_TOKEN,
    },
    // 監控和分析
    analytics: {
      gaId: import.meta.env.VITE_GA_ID,
      sentryDsn: import.meta.env.VITE_SENTRY_DSN,
    },
  };
};

// 獲取當前配置
export const config = getServerConfig();
export const appConfig = getAppConfig();
export const thirdPartyConfig = getThirdPartyConfig();
