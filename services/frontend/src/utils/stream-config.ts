// 串流配置管理
export const getStreamConfig = () => {
  // 檢查是否在 ngrok 環境中
  const isNgrok =
    window.location.hostname.includes("ngrok") ||
    window.location.hostname.includes("tcp.jp.ngrok.io") ||
    window.location.hostname.includes("tcp.ngrok.io");

  if (isNgrok) {
    // 在 ngrok 環境中，使用動態獲取當前域名
    const protocol = window.location.protocol;
    const hostname = window.location.hostname;

    return {
      // RTMP 推流地址 (通過 ngrok 暴露的端口)
      rtmpPushUrl: `rtmp://${hostname}:1935/live`,
      // HLS 播放地址 (通過 nginx 反向代理)
      hlsPlayUrl: `${protocol}//${hostname}/hls`,
      // WebSocket 地址
      wsUrl: `${protocol === "https:" ? "wss:" : "ws:"}//${hostname}/ws`,
    };
  } else {
    // 本地開發環境
    return {
      // RTMP 推流地址 (直接連接到 nginx-rtmp)
      rtmpPushUrl: "rtmp://localhost:1935/live",
      // HLS 播放地址 (通過 nginx 反向代理)
      hlsPlayUrl: "/hls",
      // WebSocket 地址
      wsUrl: "/ws",
    };
  }
};

// 獲取當前配置
export const streamConfig = getStreamConfig();

// 生成 RTMP 推流地址
export const getRtmpPushUrl = (streamKey: string | undefined) => {
  if (!streamKey) return "";
  return `${streamConfig.rtmpPushUrl}/${streamKey}`;
};

// 生成 HLS 播放地址
export const getHlsPlayUrl = (streamKey: string | undefined) => {
  if (!streamKey) return "";
  return `${streamConfig.hlsPlayUrl}/${streamKey}/index.m3u8`;
};
