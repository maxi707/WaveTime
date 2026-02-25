import { createRouter, createWebHistory } from 'vue-router'
import LandingPage from './pages/LandingPage.vue'
import LoginPage from './pages/LoginPage.vue'
import RegisterPage from './pages/RegisterPage.vue'
import DashboardPage from './pages/DashboardPage.vue'
import SchedulePage from './pages/SchedulePage.vue'
import AdminPage from './pages/AdminPage.vue'
import { isAdmin, isAuthed } from './lib/session'

const routes = [
  { path: '/', name: 'home', component: LandingPage },
  { path: '/login', name: 'login', component: LoginPage, meta: { guestOnly: true } },
  { path: '/register', name: 'register', component: RegisterPage, meta: { guestOnly: true } },
  { path: '/schedule', name: 'schedule', component: SchedulePage, meta: { requiresAuth: true } },
  { path: '/admin', name: 'admin', component: AdminPage, meta: { requiresAuth: true, requiresAdmin: true } },
  { path: '/app', name: 'app', component: DashboardPage, meta: { requiresAuth: true } },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

router.beforeEach((to) => {
  const authed = isAuthed()

  if (to.meta.requiresAuth && !authed) {
    return { name: 'login' }
  }

  if (to.meta.guestOnly && authed) {
    return { name: 'app' }
  }

  if (to.meta.requiresAdmin && !isAdmin()) {
    return { name: 'app' }
  }

  return true
})

export default router
