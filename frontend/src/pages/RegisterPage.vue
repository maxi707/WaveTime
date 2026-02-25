<script setup>
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { registerUser } from '../api'
import { setToken } from '../lib/session'

const router = useRouter()
const busy = ref(false)
const error = ref('')

const form = reactive({
  full_name: '',
  phone: '',
  email: '',
  password: '',
})

async function submit() {
  busy.value = true
  error.value = ''
  try {
    const payload = {
      full_name: form.full_name,
      phone: form.phone || null,
      email: form.email || null,
      password: form.password,
    }
    const r = await registerUser(payload)
    setToken(r.token)
    await router.push('/app')
  } catch (e) {
    error.value = e.message
  } finally {
    busy.value = false
  }
}
</script>

<template>
  <section class="auth-shell reveal-a">
    <aside class="card auth-side">
      <p class="auth-side-mark">New Account</p>
      <h1>Регистрация</h1>
      <p class="subtitle">Заполни профиль и сразу получи доступ к бронированию тренировок.</p>
      <ul class="auth-benefits">
        <li>Подбор тренировок по дате и времени</li>
        <li>Единый профиль для сети бассейнов</li>
        <li>История броней и оплат</li>
      </ul>
    </aside>

    <article class="auth-page card">
      <h2>Создание аккаунта</h2>
      <p class="subtitle">Минимум данных, быстрый вход в систему</p>

      <div class="form-grid">
        <label>
          Full Name
          <input v-model="form.full_name" placeholder="Иван Петров" />
        </label>
        <label>
          Phone
          <input v-model="form.phone" placeholder="+79990001122" />
        </label>
        <label>
          Email
          <input v-model="form.email" placeholder="mail@example.com" />
        </label>
        <label>
          Password
          <input v-model="form.password" type="password" placeholder="******" />
        </label>
      </div>

      <p v-if="error" class="error-text">{{ error }}</p>

      <div class="auth-actions">
        <button class="btn primary" :disabled="busy" @click="submit">Создать аккаунт</button>
        <RouterLink class="btn ghost" to="/login">Уже есть аккаунт</RouterLink>
      </div>
    </article>
  </section>
</template>
