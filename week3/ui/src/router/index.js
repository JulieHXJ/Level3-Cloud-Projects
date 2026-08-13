import { createRouter, createWebHistory } from 'vue-router'

import LoginView from '../views/LoginView.vue'
import DashboardView from '../views/DashboardView.vue'
import InstanceDetailView from '../views/InstanceDetailView.vue'
import { useAuthStore } from '../stores/auth'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),

  routes: [
    {
      path: '/',
      redirect: '/login',
    },
    {
      path: '/login',
      name: 'login',
      component: LoginView,
    },
    {
      path: '/dashboard',
      redirect: () => {
        const auth = useAuthStore()

        if (!auth.isAuthenticated) {
          return '/login'
        }

        return auth.dashboardPath
      },
    },
    {
      path: '/admin/dashboard',
      name: 'admin-dashboard',
      component: DashboardView,
      meta: {
        requiresAuth: true,
        roles: ['admin'],
      },
    },
    {
      path: '/user/dashboard',
      name: 'user-dashboard',
      component: DashboardView,
      meta: {
        requiresAuth: true,
        roles: ['user'],
      },
    },
    {
      path: '/instances',
      redirect: '/dashboard',
    },
    {
      path: '/instances/:id',
      name: 'instance-detail',
      component: InstanceDetailView,
      meta: {
        requiresAuth: true,
        roles: ['admin', 'user'],
      },
    },
  ],
})

router.beforeEach((to) => {
  const auth = useAuthStore()

  if (to.name === 'login' && auth.isAuthenticated) {
    return auth.dashboardPath
  }

  if (to.meta.requiresAuth && !auth.isAuthenticated) {
    return {
      name: 'login',
      query: {
        redirect: to.fullPath,
      },
    }
  }

  if (Array.isArray(to.meta.roles) && !to.meta.roles.includes(auth.role)) {
    return auth.dashboardPath
  }
})

export default router
