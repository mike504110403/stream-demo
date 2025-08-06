// import request from '@/utils/request';  // 暫時註釋掉未使用的 import
import type {
  PublicStreamResponse,
  PublicStreamData,
  PublicStreamURLResponse,
  PublicStreamStats,
  PublicStreamInfo,
} from "@/types/public-stream";

// 公開流 API 服務
export const publicStreamApi = {
  // 獲取所有可用的公開流
  getAvailableStreams(): Promise<PublicStreamData> {
    // 調用後端 API 服務
    return fetch("/api/public-streams")
      .then((response) => response.json())
      .then((data) => {
        // 處理後端 API 的響應格式
        if (data.success && data.data && data.data.streams) {
          return {
            streams: data.data.streams.map((stream: any) => ({
              ...stream,
              status: stream.enabled ? "active" : "inactive",
            })),
            total: data.data.total,
          };
        } else {
          throw new Error("響應格式不正確");
        }
      });
  },

  // 獲取特定流的詳細資訊
  getStreamInfo(streamName: string): Promise<PublicStreamInfo> {
    // 從後端 API 獲取流資訊
    return fetch("/api/public-streams")
      .then((response) => response.json())
      .then((data) => {
        if (data.success && data.data && data.data.streams) {
          const stream = data.data.streams.find(
            (s: any) => s.name === streamName,
          );
          if (stream) {
            return {
              ...stream,
              status: stream.enabled ? "active" : "inactive",
            };
          } else {
            throw new Error("流不存在");
          }
        } else {
          throw new Error("響應格式不正確");
        }
      });
  },

  // 獲取流的播放 URL
  getStreamURL(streamName: string): Promise<PublicStreamURLResponse> {
    // 直接返回 Stream-Puller 的 HLS 播放 URL
    const hlsUrl = `/stream-puller/hls/${streamName}/index.m3u8`;
    return Promise.resolve({
      success: true,
      data: {
        stream_name: streamName,
        urls: { hls: hlsUrl },
      },
    });
  },

  // 獲取播放 URL (HLS)
  getStreamURLs(
    streamName: string,
  ): Promise<{ stream_name: string; urls: { hls: string } }> {
    // 直接從 Stream-Puller 獲取 HLS 播放 URL
    const hlsUrl = `/stream-puller/hls/${streamName}/index.m3u8`;
    return Promise.resolve({
      stream_name: streamName,
      urls: { hls: hlsUrl },
    });
  },

  // 獲取流的統計資訊
  getStreamStats(
    streamName: string,
  ): Promise<{ success: boolean; data: PublicStreamStats }> {
    // 返回模擬的統計資訊，因為 Stream-Puller 沒有統計 API
    return Promise.resolve({
      success: true,
      data: {
        viewer_count: Math.floor(Math.random() * 100) + 1,
        stream_name: streamName,
        status: "active",
        uptime: Math.floor(Math.random() * 3600) + 1,
        title: streamName,
        last_update: new Date().toISOString(),
        category: "default",
      },
    });
  },

  // 按分類獲取流
  getStreamsByCategory(category: string): Promise<PublicStreamResponse> {
    // 從後端 API 獲取流並按分類過濾
    return fetch("/api/public-streams")
      .then((response) => response.json())
      .then((data) => {
        if (data.success && data.data && data.data.streams) {
          const filteredStreams = data.data.streams.filter(
            (s: any) => s.category === category,
          );
          return {
            success: true,
            data: {
              streams: filteredStreams.map((stream: any) => ({
                ...stream,
                status: stream.enabled ? "active" : "inactive",
              })),
              total: filteredStreams.length,
            },
          };
        } else {
          throw new Error("響應格式不正確");
        }
      });
  },
};
