<template>
  <div class="webrtc-player">
    <video
      ref="videoElement"
      class="w-full h-full"
      controls
      autoplay
      muted
      playsinline
    >
      您的瀏覽器不支援 WebRTC 播放。
    </video>

    <!-- 連接狀態 -->
    <div v-if="connectionStatus" class="absolute top-2 right-2">
      <span
        :class="[
          'px-2 py-1 text-xs font-bold rounded-full',
          connectionStatus === 'connected'
            ? 'bg-green-500 text-white'
            : connectionStatus === 'connecting'
              ? 'bg-yellow-500 text-white'
              : 'bg-red-500 text-white',
        ]"
      >
        {{ getStatusText(connectionStatus) }}
      </span>
    </div>

    <!-- 錯誤信息 -->
    <div
      v-if="error"
      class="absolute inset-0 bg-black/80 flex items-center justify-center"
    >
      <div class="text-center text-white">
        <div class="text-red-400 mb-2">
          <svg
            class="w-12 h-12 mx-auto"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L3.732 16.5c-.77.833.192 2.5 1.732 2.5z"
            ></path>
          </svg>
        </div>
        <h3 class="text-lg font-bold mb-2">WebRTC 連接失敗</h3>
        <p class="text-gray-300 mb-4">{{ error }}</p>
        <button
          @click="connect"
          class="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600 transition-colors"
        >
          重新連接
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from "vue";

interface Props {
  streamUrl: string;
}

const props = defineProps<Props>();

const videoElement = ref<HTMLVideoElement>();
const connectionStatus = ref<"disconnected" | "connecting" | "connected">(
  "disconnected",
);
const error = ref<string>("");

let peerConnection: RTCPeerConnection | null = null;
let wsConnection: WebSocket | null = null;

const getStatusText = (status: string) => {
  switch (status) {
    case "connected":
      return "已連接";
    case "connecting":
      return "連接中...";
    case "disconnected":
      return "未連接";
    default:
      return "未知狀態";
  }
};

const connect = async () => {
  if (!videoElement.value) return;

  try {
    connectionStatus.value = "connecting";
    error.value = "";

    // 創建 WebRTC 連接
    peerConnection = new RTCPeerConnection({
      iceServers: [{ urls: "stun:stun.l.google.com:19302" }],
    });

    // 處理遠程流
    peerConnection.ontrack = (event) => {
      if (videoElement.value) {
        videoElement.value.srcObject = event.streams[0];
        connectionStatus.value = "connected";
      }
    };

    // 處理連接狀態變化
    peerConnection.onconnectionstatechange = () => {
      console.log("WebRTC 連接狀態:", peerConnection?.connectionState);
      if (peerConnection?.connectionState === "connected") {
        connectionStatus.value = "connected";
      } else if (peerConnection?.connectionState === "failed") {
        error.value = "WebRTC 連接失敗";
        connectionStatus.value = "disconnected";
      }
    };

    // 連接到信令服務器
    await connectToSignalingServer();
  } catch (err) {
    console.error("WebRTC 連接錯誤:", err);
    error.value = err instanceof Error ? err.message : "連接失敗";
    connectionStatus.value = "disconnected";
  }
};

const connectToSignalingServer = async () => {
  // 這裡需要連接到後端的 WebRTC 信令服務器
  // 暫時使用模擬連接
  console.log("連接到 WebRTC 信令服務器:", props.streamUrl);

  // 模擬連接成功
  setTimeout(() => {
    if (peerConnection) {
      connectionStatus.value = "connected";
    }
  }, 1000);
};

const disconnect = () => {
  if (peerConnection) {
    peerConnection.close();
    peerConnection = null;
  }

  if (wsConnection) {
    wsConnection.close();
    wsConnection = null;
  }

  connectionStatus.value = "disconnected";
};

onMounted(() => {
  connect();
});

onUnmounted(() => {
  disconnect();
});
</script>

<style scoped>
.webrtc-player {
  position: relative;
  width: 100%;
  height: 100%;
}
</style>
