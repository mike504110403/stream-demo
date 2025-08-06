<template>
  <div class="navbar">
    <div class="navbar-brand">
      <router-link to="/" class="brand-link">
        <h2>🎬 串流平台</h2>
      </router-link>
    </div>

    <div class="navbar-menu">
      <el-menu
        :default-active="activeIndex"
        mode="horizontal"
        @select="handleSelect"
        class="navbar-nav"
      >
        <el-menu-item index="/">首頁</el-menu-item>
        <el-menu-item index="/public-streams">公開直播</el-menu-item>
        <el-menu-item index="/videos">影片</el-menu-item>
        <el-menu-item index="/live-rooms">直播間</el-menu-item>
        <el-menu-item index="/payments">支付</el-menu-item>
        <el-menu-item index="/dashboard">儀表板</el-menu-item>
      </el-menu>
    </div>

    <div class="navbar-user">
      <el-dropdown @command="handleCommand">
        <span class="el-dropdown-link">
          <el-avatar :size="32" :src="authStore.user?.avatar">
            {{ authStore.user?.username?.charAt(0).toUpperCase() }}
          </el-avatar>
          <span class="username">{{ authStore.user?.username }}</span>
          <el-icon class="el-icon--right">
            <arrow-down />
          </el-icon>
        </span>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item command="profile">個人資料</el-dropdown-item>
            <el-dropdown-item divided command="logout">登出</el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useRouter, useRoute } from "vue-router";
import { useAuthStore } from "@/store/auth";
import { ArrowDown } from "@element-plus/icons-vue";

const router = useRouter();
const route = useRoute();
const authStore = useAuthStore();

const activeIndex = computed(() => route.path);

const handleSelect = (key: string) => {
  router.push(key);
};

const handleCommand = (command: string) => {
  switch (command) {
    case "profile":
      router.push("/profile");
      break;
    case "logout":
      authStore.logout();
      router.push("/login");
      break;
  }
};
</script>

<style scoped>
.navbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 60px;
  padding: 0 20px;
  background-color: #fff;
  border-bottom: 1px solid #e4e7ed;
}

.navbar-brand .brand-link {
  text-decoration: none;
  color: #409eff;
}

.navbar-brand h2 {
  margin: 0;
  font-size: 20px;
  font-weight: bold;
}

.navbar-nav {
  border-bottom: none;
}

.navbar-user {
  display: flex;
  align-items: center;
}

.el-dropdown-link {
  display: flex;
  align-items: center;
  cursor: pointer;
  color: #606266;
}

.username {
  margin-left: 8px;
  margin-right: 4px;
}
</style>
