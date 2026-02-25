<script setup>
import { nextTick, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { clearToken, getAuthChangedEventName, isAuthed } from '../lib/session'

const router = useRouter()
const authed = ref(isAuthed())

function syncScrollMode() {
  const scrolling = document.scrollingElement || document.documentElement
  const overflowPx = Math.ceil(scrolling.scrollHeight - scrolling.clientHeight)
  const fitsViewport = overflowPx <= 12
  const mode = fitsViewport ? 'hidden' : 'auto'
  document.documentElement.style.overflowY = mode
  document.body.style.overflowY = mode
}

function scheduleSync() {
  window.requestAnimationFrame(() => {
    window.requestAnimationFrame(syncScrollMode)
  })
}

onMounted(async () => {
  await nextTick()
  authed.value = isAuthed()
  scheduleSync()
  window.addEventListener('resize', scheduleSync)
  window.addEventListener(getAuthChangedEventName(), handleAuthChanged)
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', scheduleSync)
  window.removeEventListener(getAuthChangedEventName(), handleAuthChanged)
  document.documentElement.style.overflowY = 'auto'
  document.body.style.overflowY = 'auto'
})

function handleAuthChanged() {
  authed.value = isAuthed()
}

async function logout() {
  clearToken()
  authed.value = false
  await router.replace('/')
}
</script>

<template>
  <section class="landing-stage">
    <section class="feature-grid reveal-b">
      <article class="feature card">
        <h3>Умное расписание</h3>
        <p>Отбор по филиалу и диапазону дат, мгновенная видимость доступных мест в каждом слоте.</p>
      </article>
      <article class="feature card">
        <h3>Личный кабинет</h3>
        <p>Профиль клиента, история броней, статусы оплат и быстрые действия без лишних экранов.</p>
      </article>
      <article class="feature card">
        <h3>Контроль для админа</h3>
        <p>Корректировка записей, управление слотами и полный audit trail действий персонала.</p>
      </article>
    </section>

    <div class="landing-center">
      <article class="hero card hero-card reveal-a">
        <p class="hero-mark">Network Pool Booking</p>
        <h1 class="hero-logo">WaveTime</h1>
        <p class="hero-subtitle">
          Платформа сети бассейнов: бронируйте тренировки, управляйте слотами и подтверждайте оплату за минуты.
        </p>

        <div class="hero-actions">
          <RouterLink v-if="!authed" class="btn primary hero-btn" to="/register">Создать аккаунт</RouterLink>
          <RouterLink v-if="!authed" class="btn ghost hero-btn" to="/login">Войти в кабинет</RouterLink>
          <button v-if="authed" class="btn danger hero-btn" @click="logout">Выйти</button>
        </div>

        <div class="hero-points">
          <div class="point-card">
            <span>07:00-23:00</span>
            <p>Ежедневные слоты</p>
          </div>
          <div class="point-card">
            <span>60 мин</span>
            <p>Фиксированная сессия</p>
          </div>
          <div class="point-card">
            <span>Online Pay</span>
            <p>Подтверждение оплаты</p>
          </div>
        </div>
      </article>

      <div class="hero-wave" aria-hidden="true"></div>
    </div>
  </section>
</template>
