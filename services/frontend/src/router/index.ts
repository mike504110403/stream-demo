import { createRouter, createWebHistory } from "vue-router";
import { useAuthStore } from "@/store/auth";

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: "/",
      name: "Home",
      component: () => import("@/views/home/HomeView.vue"),
      meta: { requiresAuth: false },
    },
    {
      path: "/login",
      name: "Login",
      component: () => import("@/views/auth/LoginView.vue"),
      meta: { requiresAuth: false },
    },
    {
      path: "/register",
      name: "Register",
      component: () => import("@/views/auth/RegisterView.vue"),
      meta: { requiresAuth: false },
    },
    {
      path: "/profile",
      name: "Profile",
      component: () => import("@/views/auth/ProfileView.vue"),
      meta: { requiresAuth: true },
    },
    {
      path: "/videos",
      name: "Videos",
      component: () => import("@/views/video/VideoListView.vue"),
      meta: { requiresAuth: true },
    },
    {
      path: "/videos/upload",
      name: "UploadVideo",
      component: () => import("@/views/video/UploadVideoView.vue"),
      meta: { requiresAuth: true },
    },
    {
      path: "/videos/:id",
      name: "VideoDetail",
      component: () => import("@/views/video/VideoDetailView.vue"),
      meta: { requiresAuth: true },
    },
    // 直播間路由
    {
      path: "/live-rooms",
      name: "LiveRooms",
      component: () => import("@/views/live/LiveRoomListView.vue"),
      meta: { requiresAuth: true },
    },
    {
      path: "/live-rooms/create",
      name: "CreateLiveRoom",
      component: () => import("@/views/live/LiveRoomCreate.vue"),
      meta: { requiresAuth: true },
    },
    {
      path: "/live-rooms/:id",
      name: "LiveRoomDetail",
      component: () => import("@/views/live/LiveRoom.vue"),
      meta: { requiresAuth: true },
    },
    {
      path: "/public-streams",
      name: "PublicStreams",
      component: () => import("@/views/public-stream/PublicStreamListView.vue"),
      meta: { requiresAuth: false },
    },
    {
      path: "/debug",
      name: "Debug",
      component: () => import("@/views/DebugView.vue"),
      meta: { requiresAuth: false },
    },
    {
      path: "/public-streams/:name",
      name: "PublicStreamPlayer",
      component: () =>
        import("@/views/public-stream/PublicStreamPlayerView.vue"),
      meta: { requiresAuth: false },
    },
    {
      path: "/public-streams/manage",
      name: "PublicStreamManage",
      component: () =>
        import("@/views/public-stream/PublicStreamManageView.vue"),
      meta: { requiresAuth: true },
    },
    {
      path: "/payments",
      name: "Payments",
      component: () => import("@/views/DashboardView.vue"), // 暫時使用 Dashboard
      meta: { requiresAuth: true },
    },
    {
      path: "/payments/create",
      name: "CreatePayment",
      component: () => import("@/views/DashboardView.vue"), // 暫時使用 Dashboard
      meta: { requiresAuth: true },
    },
    {
      path: "/dashboard",
      name: "Dashboard",
      component: () => import("@/views/DashboardView.vue"),
      meta: { requiresAuth: true },
    },
    {
      path: "/debug",
      name: "Debug",
      component: () => import("@/views/DebugView.vue"),
      meta: { requiresAuth: false },
    },
    {
      path: "/:pathMatch(.*)*",
      name: "NotFound",
      component: () => import("@/views/NotFoundView.vue"),
    },
  ],
});

// 路由守衛
router.beforeEach((to, _from, next) => {
  const authStore = useAuthStore();

  if (to.meta.requiresAuth && !authStore.isAuthenticated) {
    // 保存用戶原本要訪問的路由
    const redirect = to.fullPath;
    next({
      path: "/login",
      query: { redirect },
    });
  } else if (
    (to.name === "Login" || to.name === "Register") &&
    authStore.isAuthenticated
  ) {
    // 如果已登入，檢查是否有重定向參數
    const redirect = to.query.redirect as string;
    next(redirect || "/dashboard");
  } else {
    next();
  }
});

export default router;
