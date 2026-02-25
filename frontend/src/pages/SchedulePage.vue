<script setup>
import { computed, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import { createBooking, healthz, listPools, listSchedule } from '../api'

const backendStatus = ref('checking')
const toast = ref('')
const busy = ref(false)

const pools = ref([])
const selectedPool = ref('')
const scheduleRange = reactive({
  from: todayISO(),
  to: plusDaysISO(7),
})
const schedule = ref([])
const selectedSlotIds = ref(new Set())
let filterTimer = null

const groupedSchedule = computed(() => {
  const groups = new Map()
  for (const slot of schedule.value) {
    const date = new Date(slot.starts_at).toLocaleDateString('ru-RU', {
      weekday: 'short',
      day: '2-digit',
      month: 'short',
    })
    if (!groups.has(date)) groups.set(date, [])
    groups.get(date).push(slot)
  }
  return Array.from(groups.entries()).map(([date, slots]) => ({ date, slots }))
})

const scheduleStats = computed(() => {
  const totalSlots = schedule.value.length
  const totalFree = schedule.value.reduce((acc, s) => acc + Math.max(0, s.free_count), 0)
  const availableSlots = schedule.value.filter((s) => s.free_count > 0).length
  return { totalSlots, totalFree, availableSlots }
})
const selectedCount = computed(() => selectedSlotIds.value.size)

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

function formatTime(v) {
  return new Date(v).toLocaleTimeString('ru-RU', {
    hour: '2-digit',
    minute: '2-digit',
  })
}

async function checkBackend() {
  try {
    const r = await healthz()
    backendStatus.value = r?.status === 'ok' ? 'online' : 'degraded'
  } catch {
    backendStatus.value = 'offline'
  }
}

async function loadPools() {
  pools.value = await listPools()
  if (!selectedPool.value && pools.value.length > 0) {
    selectedPool.value = String(pools.value[0].id)
  }
}

async function loadSchedule() {
  if (!selectedPool.value) {
    setToast('Выбери бассейн для загрузки расписания')
    return
  }

  schedule.value = await listSchedule(
    Number(selectedPool.value),
    scheduleRange.from,
    scheduleRange.to,
  )
  selectedSlotIds.value = new Set()
}

function scheduleAutoReload() {
  if (filterTimer) {
    window.clearTimeout(filterTimer)
  }
  filterTimer = window.setTimeout(() => {
    loadSchedule().catch((e) => setToast(e.message || 'Не удалось обновить расписание'))
  }, 250)
}

async function bookSlot(slotId) {
  busy.value = true
  try {
    await createBooking(slotId)
    await loadSchedule()
    setToast('Бронирование создано')
  } catch (e) {
    setToast(e.message)
  } finally {
    busy.value = false
  }
}

function toggleSlotSelection(slotId) {
  const next = new Set(selectedSlotIds.value)
  if (next.has(slotId)) {
    next.delete(slotId)
  } else {
    next.add(slotId)
  }
  selectedSlotIds.value = next
}

async function bookSelectedSlots() {
  if (selectedSlotIds.value.size === 0) {
    setToast('Выбери минимум один слот')
    return
  }

  busy.value = true
  try {
    const ids = Array.from(selectedSlotIds.value)
    let success = 0
    for (const slotID of ids) {
      try {
        await createBooking(slotID)
        success += 1
      } catch {
        // continue, aggregated result below
      }
    }
    await loadSchedule()
    if (success === ids.length) {
      setToast(`Забронировано слотов: ${success}`)
    } else if (success > 0) {
      setToast(`Забронировано ${success} из ${ids.length}. Часть слотов уже недоступна.`)
    } else {
      setToast('Не удалось забронировать выбранные слоты')
    }
  } finally {
    busy.value = false
  }
}

onMounted(async () => {
  await checkBackend()
  try {
    await loadPools()
    await loadSchedule()
  } catch (e) {
    setToast(e.message || 'Не удалось загрузить расписание')
  }
})

onBeforeUnmount(() => {
  if (filterTimer) {
    window.clearTimeout(filterTimer)
  }
})

watch(selectedPool, () => {
  if (!selectedPool.value) return
  scheduleAutoReload()
})

watch(
  () => [scheduleRange.from, scheduleRange.to],
  () => {
    if (!selectedPool.value) return
    scheduleAutoReload()
  },
)
</script>

<template>
  <section class="dashboard-head card reveal-a">
    <div>
      <h2>График тренировок и слотов</h2>
      <p class="subtitle">
        Подбор слотов для записи
        <span class="subtitle-sep">|</span>
        API: <span class="status" :class="`status-${backendStatus}`">{{ backendStatus }}</span>
      </p>
    </div>
    <button class="btn ghost" :disabled="busy" @click="loadSchedule">Обновить</button>
  </section>

  <section class="stats-grid reveal-b">
    <article class="card stat-card">
      <p>Слотов в диапазоне</p>
      <strong>{{ scheduleStats.totalSlots }}</strong>
    </article>
    <article class="card stat-card">
      <p>Доступных слотов</p>
      <strong>{{ scheduleStats.availableSlots }}</strong>
    </article>
    <article class="card stat-card">
      <p>Свободных мест</p>
      <strong>{{ scheduleStats.totalFree }}</strong>
    </article>
    <article class="card stat-card">
      <p>Интервал</p>
      <strong>{{ scheduleRange.from }} - {{ scheduleRange.to }}</strong>
    </article>
  </section>

  <section class="card reveal-b">
    <div class="section-head">
      <h3>Фильтры</h3>
      <button class="btn primary" :disabled="busy || selectedCount === 0" @click="bookSelectedSlots">
        Забронировать выбранные ({{ selectedCount }})
      </button>
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

    <div class="legend-row">
      <span class="badge ok">Свободно</span>
      <span class="badge warn">Мало мест / нет мест</span>
    </div>
  </section>

  <section class="reveal-c schedule-days">
    <article v-for="day in groupedSchedule" :key="day.date" class="card">
      <div class="section-head">
        <h3>{{ day.date }}</h3>
        <span class="subtitle">{{ day.slots.length }} слотов</span>
      </div>

      <div class="table-wrap">
        <table>
          <thead>
            <tr>
              <th>Выбор</th>
              <th>Время</th>
              <th>Тренировка</th>
              <th>Цена</th>
              <th>Свободно</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="slot in day.slots" :key="slot.id">
              <td>
                <input
                  type="checkbox"
                  :checked="selectedSlotIds.has(slot.id)"
                  :disabled="busy || slot.free_count <= 0"
                  @change="toggleSlotSelection(slot.id)"
                />
              </td>
              <td>{{ formatTime(slot.starts_at) }}</td>
              <td>{{ slot.training_type }}</td>
              <td>{{ slot.price }} ₽</td>
              <td>
                <span class="badge" :class="slot.free_count > 0 ? 'ok' : 'warn'">
                  {{ slot.free_count }} / {{ slot.capacity }}
                </span>
              </td>
              <td>
                <button class="btn small" :disabled="busy || slot.free_count <= 0" @click="bookSlot(slot.id)">
                  Забронировать
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </article>

    <article v-if="groupedSchedule.length === 0" class="card">
      <p class="empty">Нет слотов в выбранном диапазоне</p>
    </article>
  </section>

  <p v-if="toast" class="toast">{{ toast }}</p>
</template>
