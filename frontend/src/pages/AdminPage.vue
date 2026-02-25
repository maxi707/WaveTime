<script setup>
import { onMounted, reactive, ref } from 'vue'
import {
  createAdminSlot,
  getMe,
  healthz,
  listAdminBookings,
  listPools,
  listSchedule,
  updateAdminBooking,
  updateAdminSlot,
} from '../api'

const backendStatus = ref('checking')
const toast = ref('')
const busy = ref(false)
const user = ref(null)

const pools = ref([])
const selectedPool = ref('')
const scheduleRange = reactive({
  from: todayISO(),
  to: plusDaysISO(7),
})
const schedule = ref([])

const bookings = ref([])
const bookingStatusFilter = ref('')

const createSlotForm = reactive({
  pool_id: '',
  training_type_id: '1',
  starts_at: '',
  capacity: 10,
  price: '800',
})

function todayISO() {
  return new Date().toISOString().slice(0, 10)
}

function plusDaysISO(days) {
  const d = new Date()
  d.setDate(d.getDate() + days)
  return d.toISOString().slice(0, 10)
}

function setToast(message) {
  toast.value = message
  window.setTimeout(() => {
    if (toast.value === message) toast.value = ''
  }, 2800)
}

function formatDateTime(v) {
  return new Date(v).toLocaleString('ru-RU', {
    day: '2-digit',
    month: 'short',
    hour: '2-digit',
    minute: '2-digit',
  })
}

function toInputDateTime(startsAt) {
  const date = new Date(startsAt)
  const local = new Date(date.getTime() - date.getTimezoneOffset() * 60000)
  return local.toISOString().slice(0, 16)
}

function toRFC3339(value) {
  if (!value) return ''
  const d = new Date(value)
  return d.toISOString()
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
  user.value = await getMe()
}

async function loadPools() {
  pools.value = await listPools()
  if (!selectedPool.value && pools.value.length > 0) {
    selectedPool.value = String(pools.value[0].id)
  }
  if (!createSlotForm.pool_id && pools.value.length > 0) {
    createSlotForm.pool_id = String(pools.value[0].id)
  }
}

async function loadSchedule() {
  if (!selectedPool.value) return
  schedule.value = await listSchedule(
    Number(selectedPool.value),
    scheduleRange.from,
    scheduleRange.to,
  )
}

async function loadAdminBookings() {
  bookings.value = await listAdminBookings(bookingStatusFilter.value, 200)
}

async function saveSlot(slot) {
  busy.value = true
  try {
    await updateAdminSlot(slot.id, {
      capacity: Number(slot.capacity),
      status: slot.status,
      price: String(slot.price),
    })
    setToast(`Слот #${slot.id} обновлен`)
  } catch (e) {
    setToast(e.message)
  } finally {
    busy.value = false
  }
}

async function createSlot() {
  busy.value = true
  try {
    const starts = toRFC3339(createSlotForm.starts_at)
    if (!starts) {
      setToast('Укажи starts_at')
      return
    }

    await createAdminSlot({
      pool_id: Number(createSlotForm.pool_id),
      training_type_id: Number(createSlotForm.training_type_id),
      starts_at: starts,
      capacity: Number(createSlotForm.capacity),
      price: String(createSlotForm.price),
    })

    setToast('Новый слот создан')
    await loadSchedule()
  } catch (e) {
    setToast(e.message)
  } finally {
    busy.value = false
  }
}

async function updateBookingStatus(bookingId, nextStatus) {
  busy.value = true
  try {
    await updateAdminBooking(bookingId, nextStatus, 'updated_by_admin_ui')
    await loadAdminBookings()
    setToast(`Бронь #${bookingId} -> ${nextStatus}`)
  } catch (e) {
    setToast(e.message)
  } finally {
    busy.value = false
  }
}

onMounted(async () => {
  await checkBackend()
  try {
    await loadMe()
    await loadPools()
    await loadSchedule()
    await loadAdminBookings()
  } catch (e) {
    setToast(e.message || 'Ошибка загрузки админ-данных')
  }
})
</script>

<template>
  <section class="dashboard-head card reveal-a">
    <div>
      <h2>Админ-панель</h2>
      <p class="subtitle">
        Управление публичным графиком и статусами записей
        <span class="subtitle-sep">|</span>
        API: <span class="status" :class="`status-${backendStatus}`">{{ backendStatus }}</span>
      </p>
    </div>
    <span class="role-chip">{{ user?.role }}</span>
  </section>

  <section class="card reveal-b">
    <div class="section-head">
      <h3>Фильтр расписания</h3>
      <button class="btn ghost" :disabled="busy" @click="loadSchedule">Обновить</button>
    </div>

    <div class="form-grid four">
      <label>
        Бассейн
        <select v-model="selectedPool">
          <option v-for="pool in pools" :key="pool.id" :value="String(pool.id)">
            {{ pool.name }}
          </option>
        </select>
      </label>
      <label>Дата от<input v-model="scheduleRange.from" type="date" /></label>
      <label>Дата до<input v-model="scheduleRange.to" type="date" /></label>
    </div>
  </section>

  <section class="card reveal-b">
    <div class="section-head">
      <h3>Создать новый слот</h3>
      <button class="btn primary" :disabled="busy" @click="createSlot">Создать</button>
    </div>

    <div class="form-grid four">
      <label>
        Бассейн
        <select v-model="createSlotForm.pool_id">
          <option v-for="pool in pools" :key="pool.id" :value="String(pool.id)">
            {{ pool.name }}
          </option>
        </select>
      </label>
      <label>
        Training Type ID
        <input v-model="createSlotForm.training_type_id" type="number" min="1" />
      </label>
      <label>
        starts_at
        <input v-model="createSlotForm.starts_at" type="datetime-local" />
      </label>
      <label>
        Capacity
        <input v-model="createSlotForm.capacity" type="number" min="1" />
      </label>
      <label>
        Price
        <input v-model="createSlotForm.price" type="number" min="0" step="0.01" />
      </label>
    </div>
  </section>

  <section class="card reveal-c">
    <div class="section-head">
      <h3>Публичный график (редактирование)</h3>
      <span class="subtitle">{{ schedule.length }} слотов</span>
    </div>

    <div class="table-wrap">
      <table>
        <thead>
          <tr>
            <th>ID</th>
            <th>Старт</th>
            <th>Тренировка</th>
            <th>Price</th>
            <th>Capacity</th>
            <th>Status</th>
            <th>Free</th>
            <th></th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="slot in schedule" :key="slot.id">
            <td>#{{ slot.id }}</td>
            <td>{{ formatDateTime(slot.starts_at) }}</td>
            <td>{{ slot.training_type }}</td>
            <td><input v-model="slot.price" type="number" min="0" step="0.01" /></td>
            <td><input v-model.number="slot.capacity" type="number" min="1" /></td>
            <td>
              <select v-model="slot.status">
                <option value="open">open</option>
                <option value="closed">closed</option>
                <option value="cancelled">cancelled</option>
              </select>
            </td>
            <td>{{ slot.free_count }} / {{ slot.capacity }}</td>
            <td>
              <button class="btn small" :disabled="busy" @click="saveSlot(slot)">Сохранить</button>
            </td>
          </tr>
          <tr v-if="schedule.length === 0">
            <td colspan="8" class="empty">Нет слотов в выбранном диапазоне</td>
          </tr>
        </tbody>
      </table>
    </div>
  </section>

  <section class="card reveal-c">
    <div class="section-head">
      <h3>Управление бронированиями</h3>
      <button class="btn ghost" :disabled="busy" @click="loadAdminBookings">Обновить</button>
    </div>

    <div class="form-grid four">
      <label>
        Фильтр статуса
        <select v-model="bookingStatusFilter" @change="loadAdminBookings">
          <option value="">all</option>
          <option value="pending_payment">pending_payment</option>
          <option value="reserved">reserved</option>
          <option value="confirmed">confirmed</option>
          <option value="cancelled">cancelled</option>
          <option value="expired">expired</option>
          <option value="attended">attended</option>
          <option value="no_show">no_show</option>
        </select>
      </label>
    </div>

    <div class="table-wrap">
      <table>
        <thead>
          <tr>
            <th>ID</th>
            <th>User</th>
            <th>Pool</th>
            <th>Start</th>
            <th>Status</th>
            <th>Цена</th>
            <th></th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="booking in bookings" :key="booking.id">
            <td>#{{ booking.id }}</td>
            <td>{{ booking.user_id }}</td>
            <td>{{ booking.pool_name }}</td>
            <td>{{ formatDateTime(booking.starts_at) }}</td>
            <td>{{ booking.status }}</td>
            <td>{{ booking.price }} ₽</td>
            <td class="actions">
              <button class="btn small" :disabled="busy" @click="updateBookingStatus(booking.id, 'confirmed')">confirm</button>
              <button class="btn small" :disabled="busy" @click="updateBookingStatus(booking.id, 'cancelled')">cancel</button>
              <button class="btn small" :disabled="busy" @click="updateBookingStatus(booking.id, 'attended')">attended</button>
            </td>
          </tr>
          <tr v-if="bookings.length === 0">
            <td colspan="7" class="empty">Нет бронирований</td>
          </tr>
        </tbody>
      </table>
    </div>
  </section>

  <p v-if="toast" class="toast">{{ toast }}</p>
</template>
