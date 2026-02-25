<script setup>
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { loginUser } from '../api'
import { setToken } from '../lib/session'

const router = useRouter()
const busy = ref(false)
const error = ref('')

const form = reactive({
  login: '',
  password: '',
})

async function submit() {
  busy.value = true
  error.value = ''
  try {
    const r = await loginUser(form)
    setToken(r.token)
    const target = r?.user?.role === 'admin' ? '/admin' : '/app'
    await router.push(target)
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
      <p class="auth-side-mark">Member Access</p>
      <h1>Вход в WaveTime</h1>
      <p class="subtitle">Используй email или телефон, чтобы перейти к расписанию и броням.</p>
      <ul class="auth-benefits">
        <li>Доступ к слотам всех филиалов</li>
        <li>Оплата и подтверждение в одном окне</li>
        <li>Управление бронированиями в личном кабинете</li>
      </ul>
    </aside>

    <article class="auth-page card">
      <h2>Авторизация</h2>
      <p class="subtitle">Вход по email или телефону</p>

      <div class="form-grid">
        <label>
          Login
          <input v-model="form.login" placeholder="mail@example.com или +79990001122" />
        </label>
        <label>
          Password
          <input v-model="form.password" type="password" placeholder="******" />
        </label>
      </div>

      <p v-if="error" class="error-text">{{ error }}</p>

      <div class="auth-actions">
        <button class="btn primary" :disabled="busy" @click="submit">Войти</button>
        <RouterLink class="btn ghost" to="/register">Нет аккаунта</RouterLink>
      </div>
    </article>
  </section>
</template>
