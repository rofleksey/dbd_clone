import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      redirect: '/lobbies',
    },
    {
      path: '/login',
      name: 'login',
      component: () => import('../views/LoginView.vue'),
    },
    {
      path: '/register',
      name: 'register',
      component: () => import('../views/RegisterView.vue'),
    },
    {
      path: '/lobbies',
      name: 'lobbies',
      component: () => import('../views/LobbyListView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/lobby/:id',
      name: 'lobby',
      component: () => import('../views/LobbyView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/game/:id',
      name: 'game',
      component: () => import('../views/GameView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/leaderboard',
      name: 'leaderboard',
      component: () => import('../views/LeaderboardView.vue'),
    },
    {
      path: '/games',
      name: 'games',
      component: () => import('../views/GamesListView.vue'),
    },
  ],
})

router.beforeEach((to, from, next) => {
  const authStore = useAuthStore()
  if (to.meta.requiresAuth && !authStore.isLoggedIn) {
    next('/login')
  } else {
    next()
  }
})

export default router
