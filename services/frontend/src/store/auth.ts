import { defineStore } from "pinia";
import { ref, computed } from "vue";
import type { User } from "@/types";

export const useAuthStore = defineStore("auth", () => {
  // 狀態
  const token = ref<string | null>(localStorage.getItem("token"));
  const user = ref<User | null>(null);
  const isLoading = ref(false);

  // 計算屬性
  const isAuthenticated = computed(() => !!token.value);

  // 動作
  const setAuth = (authToken: string, userData: User) => {
    console.log("設置認證信息:", {
      token: !!authToken,
      user: userData?.username,
    }); // 調試用

    token.value = authToken;
    user.value = userData;
    localStorage.setItem("token", authToken);
    localStorage.setItem("user", JSON.stringify(userData));

    // 驗證存儲是否成功
    const storedToken = localStorage.getItem("token");
    const storedUser = localStorage.getItem("user");
    console.log("認證信息存儲驗證:", {
      tokenStored: storedToken === authToken,
      userStored: !!storedUser,
    }); // 調試用
  };

  const updateUser = (userData: User) => {
    user.value = userData;
    localStorage.setItem("user", JSON.stringify(userData));
  };

  const logout = () => {
    token.value = null;
    user.value = null;
    localStorage.removeItem("token");
    localStorage.removeItem("user");
  };

  const initAuth = () => {
    const savedToken = localStorage.getItem("token");
    const savedUser = localStorage.getItem("user");

    console.log("初始化認證狀態:", {
      savedToken: !!savedToken,
      savedUser: !!savedUser,
    }); // 調試用

    if (savedToken && savedUser) {
      token.value = savedToken;
      try {
        user.value = JSON.parse(savedUser);
        console.log("認證狀態恢復成功:", user.value?.username); // 調試用
      } catch (error) {
        console.error("解析用戶資料失敗:", error);
        logout();
      }
    } else if (savedToken || savedUser) {
      // 如果只有一個存在，清理所有數據以確保一致性
      console.warn("認證數據不完整，清理所有認證信息");
      logout();
    }
  };

  // 初始化
  initAuth();

  return {
    token,
    user,
    isLoading,
    isAuthenticated,
    setAuth,
    updateUser,
    logout,
    initAuth,
  };
});
