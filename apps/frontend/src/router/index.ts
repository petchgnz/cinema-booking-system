import { createRouter, createWebHistory } from 'vue-router';
import { useAuthStore } from '../stores/auth';
import LoginView from '../views/LoginView.vue';
import HomeView from '../views/HomeView.vue';

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/login',
      component: LoginView,
    },
    {
      path: '/',
      component: HomeView,
      meta: { requiresAuth: true },
    },
  ],
});


router.beforeEach((destination) => {
  const authStore = useAuthStore();

  if (destination.meta.requiresAuth && !authStore.isLoggedIn) {
    return '/login'
  }
})

export default router
