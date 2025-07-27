import request from '@/utils/request';
import type {
  PublicStreamResponse,
  PublicStreamDetailResponse,
  PublicStreamURLResponse,
  PublicStreamStats
} from '@/types/public-stream';

// 公開流 API 服務
export const publicStreamApi = {
  // 獲取所有可用的公開流
  getAvailableStreams(): Promise<PublicStreamResponse> {
    return request.get('/public-streams');
  },

  // 獲取特定流的詳細資訊
  getStreamInfo(streamName: string): Promise<PublicStreamDetailResponse> {
    return request.get(`/public-streams/${streamName}`);
  },

  // 獲取流的播放 URL
  getStreamURL(streamName: string): Promise<PublicStreamURLResponse> {
    return request.get(`/public-streams/${streamName}/url`);
  },

  // 獲取所有播放 URL (HLS + RTMP)
  getStreamURLs(streamName: string): Promise<{ success: boolean; data: { stream_name: string; urls: { hls: string; rtmp: string } } }> {
    return request.get(`/public-streams/${streamName}/urls`);
  },

  // 獲取流的統計資訊
  getStreamStats(streamName: string): Promise<{ success: boolean; data: PublicStreamStats }> {
    return request.get(`/public-streams/${streamName}/stats`);
  },

  // 按分類獲取流
  getStreamsByCategory(category: string): Promise<PublicStreamResponse> {
    return request.get(`/public-streams/category/${category}`);
  }
}; 