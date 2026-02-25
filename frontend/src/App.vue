<script setup>
import { computed, onMounted, ref } from 'vue'
import { RouterLink, RouterView, useRoute } from 'vue-router'
import { isAdmin, isAuthed } from './lib/session'

const route = useRoute()
const theme = ref(localStorage.getItem('wavetime_theme') || 'ocean-dark')
const authed = computed(() => isAuthed())
const admin = computed(() => isAdmin())

const pageTitle = computed(() => {
  if (route.path === '/app') return 'Кабинет клиента'
  if (route.path === '/schedule') return 'Расписание и бронь'
  if (route.path === '/admin') return 'Администрирование'
  if (route.path === '/login') return 'Вход в систему'
  if (route.path === '/register') return 'Регистрация'
  return 'Бронирование бассейнов'
})
const currentYear = new Date().getFullYear()

function applyTheme(nextTheme) {
  theme.value = nextTheme
  document.documentElement.setAttribute('data-theme', nextTheme)
  localStorage.setItem('wavetime_theme', nextTheme)
}

function toggleTheme() {
  applyTheme(theme.value === 'ocean-dark' ? 'ocean-light' : 'ocean-dark')
}

onMounted(() => {
  applyTheme(theme.value)
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
