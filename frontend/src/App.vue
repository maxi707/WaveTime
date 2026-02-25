<script setup>
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { RouterLink, RouterView, useRoute, useRouter } from 'vue-router'
import { getMe } from './api'
import { clearToken, getAuthChangedEventName, isAdmin, isAuthed } from './lib/session'

const route = useRoute()
const router = useRouter()
const theme = ref(localStorage.getItem('wavetime_theme') || 'ocean-dark')
const authTick = ref(0)
const nickname = ref('')
const userRole = ref('')
const authed = computed(() => {
  authTick.value
  return isAuthed()
})
const admin = computed(() => {
  authTick.value
  return isAdmin()
})

const pageTitle = computed(() => {
  if (route.path === '/app') return 'Кабинет клиента'
  if (route.path === '/schedule') return 'Расписание и бронь'
  if (route.path === '/admin') return 'Администрирование'
  if (route.path === '/login') return 'Вход в систему'
  if (route.path === '/register') return 'Регистрация'
  return 'Бронирование бассейнов'
})
const currentYear = new Date().getFullYear()
const userInitials = computed(() => {
  const source = nickname.value.trim()
  if (!source) return 'WT'
  const parts = source.split(/\s+/).filter(Boolean)
  if (parts.length === 1) return parts[0].slice(0, 2).toUpperCase()
  return (parts[0][0] + parts[1][0]).toUpperCase()
})

function applyTheme(nextTheme) {
  theme.value = nextTheme
  document.documentElement.setAttribute('data-theme', nextTheme)
  localStorage.setItem('wavetime_theme', nextTheme)
}

function toggleTheme() {
  applyTheme(theme.value === 'ocean-dark' ? 'ocean-light' : 'ocean-dark')
}

function onAuthChanged() {
  authTick.value += 1
  void refreshIdentity()
}

async function refreshIdentity() {
  if (!isAuthed()) {
    nickname.value = ''
    userRole.value = ''
    return
  }
  try {
    const me = await getMe()
    nickname.value = me.full_name || me.phone || me.email || `User #${me.id}`
    userRole.value = me.role || ''
  } catch {
    clearToken()
    nickname.value = ''
    userRole.value = ''
    await router.replace('/login')
  }
}

async function logout() {
  clearToken()
  nickname.value = ''
  userRole.value = ''
  await router.replace('/')
}

onMounted(() => {
  applyTheme(theme.value)
  window.addEventListener(getAuthChangedEventName(), onAuthChanged)
  void refreshIdentity()
})

onBeforeUnmount(() => {
  window.removeEventListener(getAuthChangedEventName(), onAuthChanged)
})
</script>

<template>
  <div class="page-shell">
    <div class="sea-glow sea-glow-1"></div>
    <div class="sea-glow sea-glow-2"></div>
    <div class="sea-grid" aria-hidden="true"></div>

    <header class="topbar card">
      <div class="topbar-brand">
        <h1>WaveTime</h1>
        <p class="subtitle topbar-sub">{{ pageTitle }}</p>
      </div>

      <div class="topbar-actions">
        <RouterLink class="chip" to="/">Главная</RouterLink>
        <RouterLink v-if="!authed" class="chip" to="/login">Войти</RouterLink>
        <RouterLink v-if="!authed" class="chip" to="/register">Регистрация</RouterLink>
        <RouterLink v-if="authed" class="chip" to="/schedule">Расписание</RouterLink>
        <RouterLink v-if="authed" class="chip" to="/app">Кабинет</RouterLink>
        <RouterLink v-if="admin" class="chip" to="/admin">Админка</RouterLink>
        <div v-if="authed" class="user-pill">
          <span class="user-avatar">{{ userInitials }}</span>
          <span class="user-meta">
            <strong>{{ nickname || 'Пользователь' }}</strong>
            <small>{{ userRole }}</small>
          </span>
        </div>
        <button v-if="authed" class="btn danger" @click="logout">Выйти</button>
        <button class="btn ghost" @click="toggleTheme">
          Тема: {{ theme === 'ocean-dark' ? 'Ночь' : 'День' }}
        </button>
      </div>
    </header>

    <main class="layout">
      <RouterView />
    </main>

    <footer class="site-footer card">
      <div class="site-footer-left">
        <strong>WaveTime</strong>
        <p>© {{ currentYear }} WaveTime Network Pool Booking. Все права защищены.</p>
      </div>

      <div class="site-footer-right">
        <a href="mailto:support@wavetime.local">support@wavetime.local</a>
        <span>+7 (900) 000-00-00</span>
        <span>Ежедневно: 07:00-23:00</span>
      </div>
    </footer>
  </div>
</template>
