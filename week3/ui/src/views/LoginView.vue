<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'

import { useAuthStore } from '../stores/auth'

const router = useRouter()
const auth = useAuthStore()

const username = ref('')
const password = ref('')
const loginError = ref('')
const isLoading = ref(false)

async function handleLogin() {
  loginError.value = ''
  isLoading.value = true

  try {
    await auth.login(username.value, password.value)
    await router.replace(auth.dashboardPath)
  } catch (error) {
    loginError.value =
      error instanceof Error ? error.message : 'Login failed'
  } finally {
    isLoading.value = false
  }
}
</script>

<template>
  <main class="page narrow-page">
    <section class="panel">
      <h1>Cloud3 Login</h1>

      <p class="muted">
        Sign in to manage PostgreSQL instances.
      </p>

      <form class="form" @submit.prevent="handleLogin">
        <label>
          Username
          <input
            v-model.trim="username"
            type="text"
            autocomplete="username"
            required
          />
        </label>

        <label>
          Password
          <input
            v-model="password"
            type="password"
            autocomplete="current-password"
            required
          />
        </label>

        <p v-if="loginError" class="error-message">
          {{ loginError }}
        </p>

        <button type="submit" :disabled="isLoading">
          {{ isLoading ? 'Signing in...' : 'Login' }}
        </button>
      </form>
    </section>
  </main>
</template>