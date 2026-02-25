<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import {
  applyWebhook,
  cancelBooking,
  getMe,
  healthz,
  initPayment,
  listMyBookings,
  patchMe,
} from '../api'
import { clearToken } from '../lib/session'

const router = useRouter()
const backendStatus = ref('checking')
const toast = ref('')
const user = ref(null)

const meForm = reactive({
  full_name: '',
  phone: '',
  email: '',
})

const bookings = ref([])
const busy = ref(false)

const stats = computed(() => {
  const totalBookings = bookings.value.length
  const activeBookings = bookings.value.filter((b) => ['pending_payment', 'reserved', 'confirmed'].includes(b.status)).length
  const paidBookings = bookings.value.filter((b) => b.status === 'confirmed').length
  const pendingPayment = bookings.value.filter((b) => b.status === 'pending_payment').length

  return { totalBookings, activeBookings, paidBookings, pendingPayment }
})

function formatDateTime(v) {
  return new Date(v).toLocaleString('ru-RU', {
    day: '2-digit',
    month: 'short',
    hour: '2-digit',
    minute: '2-digit',
  })
}

function setToast(message) {
  toast.value = message
  window.setTimeout(() => {
    if (toast.value === message) {
      toast.value = ''
    }
  }, 2800)
}

function bookingBadgeClass(status) {
  if (status === 'confirmed' || status === 'attended') return 'ok'
  if (status === 'pending_payment' || status === 'reserved') return 'warn'
  if (status === 'cancelled' || status === 'expired' || status === 'failed') return 'bad'
  return ''
}

async function checkBackend() {
  try {
    const r = await healthz()
    backendStatus.value = r?.status === 'ok' ? 'online' : 'degraded'
  } catch {
    backendStatus.value = 'offline'
  }
}

async function loadMe() {
  const me = await getMe()
  user.value = me
  meForm.full_name = me.full_name || ''
  meForm.phone = me.phone || ''
  meForm.email = me.email || ''
}

async function loadBookings() {
  bookings.value = await listMyBookings()
}

async function updateMe() {
  busy.value = true
  try {
    const payload = {
      full_name: meForm.full_name || null,
      phone: meForm.phone || null,
      email: meForm.email || null,
    }
    user.value = await patchMe(payload)
    setToast('Профиль обновлен')
  } catch (e) {
    setToast(e.message)
  } finally {
    busy.value = false
  }
}

async function cancel(id) {
  busy.value = true
  try {
    await cancelBooking(id)
    await loadBookings()
    setToast('Бронирование отменено')
  } catch (e) {
    setToast(e.message)
  } finally {
    busy.value = false
  }
}

async function payAndConfirm(bookingId) {
  busy.value = true
  try {
    const p = await initPayment(bookingId)
    await applyWebhook(p.external_id, 'paid')
    await loadBookings()
    setToast('Оплата подтверждена (mock)')
  } catch (e) {
    setToast(e.message)
  } finally {
    busy.value = false
  }
}

async function hydrate() {
  await loadMe()
  await loadBookings()
}

async function logout() {
  clearToken()
  await router.push('/login')
}

onMounted(async () => {
  await checkBackend()
  try {
    await hydrate()
  } catch {
    await logout()
  }
})
</script>

<template>
  <section class="dashboard-head card reveal-a">
    <div>
      <h2>Личный кабинет</h2>
      <p class="subtitle">
        API:
        <span class="status" :class="`status-${backendStatus}`">{{ backendStatus }}</span>
        <span class="subtitle-sep">|</span>
        {{ user?.full_name || 'Клиент' }}
      </p>
    </div>
    <button class="btn danger" @click="logout">Выйти</button>
  </section>

  <section class="stats-grid reveal-b">
    <article class="card stat-card">
      <p>Активные брони</p>
      <strong>{{ stats.activeBookings }}</strong>
    </article>
    <article class="card stat-card">
      <p>Всего броней</p>
      <strong>{{ stats.totalBookings }}</strong>
    </article>
    <article class="card stat-card">
      <p>Подтверждено</p>
      <strong>{{ stats.paidBookings }}</strong>
    </article>
    <article class="card stat-card">
      <p>Ожидают оплаты</p>
      <strong>{{ stats.pendingPayment }}</strong>
    </article>
  </section>

  <section class="card reveal-b">
    <div class="section-head">
      <h2>Профиль</h2>
      <span class="role-chip">{{ user?.role }}</span>
    </div>
    <div class="form-grid three">
      <label>ФИО<input v-model="meForm.full_name" /></label>
      <label>Телефон<input v-model="meForm.phone" /></label>
      <label>Email<input v-model="meForm.email" /></label>
    </div>
    <button class="btn primary" :disabled="busy" @click="updateMe">Сохранить профиль</button>
  </section>

  <section class="card reveal-c">
    <div class="section-head">
      <h2>Мои брони</h2>
      <button class="btn ghost" :disabled="busy" @click="loadBookings">Обновить</button>
    </div>

    <div class="table-wrap">
      <table>
        <thead>
          <tr>
            <th>ID</th>
            <th>Бассейн</th>
            <th>Старт</th>
            <th>Статус</th>
            <th>Цена</th>
            <th></th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="booking in bookings" :key="booking.id">
            <td>#{{ booking.id }}</td>
            <td>{{ booking.pool_name }}</td>
            <td>{{ formatDateTime(booking.starts_at) }}</td>
            <td><span class="badge" :class="bookingBadgeClass(booking.status)">{{ booking.status }}</span></td>
            <td>{{ booking.price }} ₽</td>
            <td class="actions">
              <button
                class="btn small"
                :disabled="busy || booking.status !== 'pending_payment'"
                @click="payAndConfirm(booking.id)"
              >
                Оплатить
              </button>
              <button
                class="btn small danger"
                :disabled="busy || !['pending_payment', 'reserved', 'confirmed'].includes(booking.status)"
                @click="cancel(booking.id)"
              >
                Отменить
              </button>
            </td>
          </tr>
          <tr v-if="bookings.length === 0">
            <td colspan="6" class="empty">Бронирований пока нет. Перейди в раздел "Расписание" в шапке.</td>
          </tr>
        </tbody>
      </table>
    </div>
  </section>

  <p v-if="toast" class="toast">{{ toast }}</p>
</template>
