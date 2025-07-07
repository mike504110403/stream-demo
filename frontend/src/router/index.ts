import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/store/auth'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      name: 'Home',
      component: () => import('@/views/home/HomeView.vue'),
      meta: { requiresAuth: false }
    },
    {
      path: '/login',
      name: 'Login',
      component: () => import('@/views/auth/LoginView.vue'),
      meta: { requiresAuth: false }
    },
    {
      path: '/register',
      name: 'Register',
      component: () => import('@/views/auth/RegisterView.vue'),
      meta: { requiresAuth: false }
    },
    {
      path: '/profile',
      name: 'Profile',
      component: () => import('@/views/auth/ProfileView.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/videos',
      name: 'Videos',
      component: () => import('@/views/video/VideoListView.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/videos/upload',
      name: 'UploadVideo',
      component: () => import('@/views/video/UploadVideoView.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/videos/:id',
      name: 'VideoDetail',
      component: () => import('@/views/video/VideoDetailView.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/lives',
      name: 'Lives',
      component: () => import('@/views/live/LiveListView.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/lives/create',
      name: 'CreateLive',
      component: () => import('@/views/live/CreateLiveView.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/lives/:id',
      name: 'LiveDetail',
      component: () => import('@/views/live/LiveDetailView.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/lives/:id/stream',
      name: 'LiveStream',
      component: () => import('@/views/live/LiveStreamView.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/payments',
      name: 'Payments',
      component: () => import('@/views/payment/PaymentListView.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/payments/create',
      name: 'CreatePayment',
      component: () => import('@/views/payment/CreatePaymentView.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/dashboard',
      name: 'Dashboard',
      component: () => import('@/views/DashboardView.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/:pathMatch(.*)*',
      name: 'NotFound',
      component: () => import('@/views/NotFoundView.vue')
    }
  ]
})

// 路由守衛
router.beforeEach((to, _from, next) => {
  const authStore = useAuthStore()
  
  if (to.meta.requiresAuth && !authStore.isAuthenticated) {
    next('/login')
  } else if ((to.name === 'Login' || to.name === 'Register') && authStore.isAuthenticated) {
    next('/dashboard')
  } else {
    next()
  }
})

export default router
