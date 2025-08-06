<template>
  <div class="login-container">
    <div class="login-card">
      <div class="login-header">
        <h1>🎬 串流平台</h1>
        <p>歡迎回來！請登入您的帳號</p>
      </div>

      <el-form
        ref="loginFormRef"
        :model="loginForm"
        :rules="loginRules"
        class="login-form"
        @submit.prevent
      >
        <el-form-item prop="username">
          <el-input
            v-model="loginForm.username"
            type="text"
            placeholder="請輸入用戶名"
            size="large"
            prefix-icon="User"
            @keyup.enter="handleLogin"
          />
        </el-form-item>

        <el-form-item prop="password">
          <el-input
            v-model="loginForm.password"
            type="password"
            placeholder="請輸入密碼"
            size="large"
            prefix-icon="Lock"
            show-password
            @keyup.enter="handleLogin"
          />
        </el-form-item>

        <el-form-item>
          <el-button
            type="primary"
            size="large"
            :loading="loading"
            @click.prevent="handleLogin"
            class="login-button"
          >
            登入
          </el-button>
        </el-form-item>
      </el-form>

      <div class="login-footer">
        <p>
          還沒有帳號？
          <router-link to="/register" class="register-link">
            立即註冊
          </router-link>
        </p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from "vue";
import { useRouter } from "vue-router";
import { ElMessage, type FormInstance, type FormRules } from "element-plus";
import { useAuthStore } from "@/store/auth";
import { login } from "@/api/user";
import type { LoginRequest } from "@/types";

const router = useRouter();
const authStore = useAuthStore();

const loginFormRef = ref<FormInstance>();
const loading = ref(false);

const loginForm = reactive<LoginRequest>({
  username: "",
  password: "",
});

const loginRules: FormRules = {
  username: [
    { required: true, message: "請輸入帳號", trigger: "blur" },
    { min: 6, message: "帳號長度不能少於6位", trigger: "blur" },
  ],
  password: [
    { required: true, message: "請輸入密碼", trigger: "blur" },
    { min: 6, message: "密碼長度不能少於6位", trigger: "blur" },
  ],
};

const handleLogin = async () => {
  if (!loginFormRef.value) return;

  try {
    const valid = await loginFormRef.value.validate();
    if (!valid) {
      console.log("表單驗證失敗");
      return;
    }

    loading.value = true;
    console.log("開始登入請求，數據:", loginForm);

    try {
      // 響應攔截器已經處理了統一格式，直接獲取 data 內容
      const response = await login(loginForm);
      const data = response as any; // 響應攔截器已經提取了 data

      console.log("登入成功，收到數據:", data);

      // 檢查必要的字段
      if (!data || !data.token || !data.user) {
        console.error("登入響應缺少必要字段:", {
          data,
          hasToken: !!data?.token,
          hasUser: !!data?.user,
        });
        ElMessage.error("登入響應格式錯誤");
        return;
      }

      // 存儲認證信息
      console.log("存儲認證信息:", {
        token: !!data.token,
        user: data.user.username,
      });
      authStore.setAuth(data.token, data.user);

      ElMessage.success("登入成功！");

      // 處理重定向邏輯
      const redirect = router.currentRoute.value.query.redirect as string;
      const targetRoute = redirect || "/dashboard";

      console.log("重定向到:", targetRoute);
      await router.replace(targetRoute);
    } catch (apiError) {
      console.error("API 請求錯誤:", apiError);
      // 錯誤已經由響應攔截器處理，這裡記錄即可
    }
  } catch (validateError) {
    console.error("表單驗證錯誤:", validateError);
    // 表單驗證錯誤通常由 Element Plus 自動顯示
  } finally {
    loading.value = false;
  }
};
</script>

<style scoped>
.login-container {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.login-card {
  width: 400px;
  padding: 40px;
  background: white;
  border-radius: 12px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1);
}

.login-header {
  text-align: center;
  margin-bottom: 32px;
}

.login-header h1 {
  margin: 0 0 8px 0;
  color: #333;
  font-size: 28px;
  font-weight: bold;
}

.login-header p {
  margin: 0;
  color: #666;
  font-size: 14px;
}

.login-form .el-form-item {
  margin-bottom: 24px;
}

.login-button {
  width: 100%;
  height: 48px;
  font-size: 16px;
  font-weight: bold;
}

.login-footer {
  text-align: center;
  margin-top: 24px;
}

.login-footer p {
  margin: 0;
  color: #666;
  font-size: 14px;
}

.register-link {
  color: #409eff;
  text-decoration: none;
  font-weight: bold;
}

.register-link:hover {
  text-decoration: underline;
}
</style>
