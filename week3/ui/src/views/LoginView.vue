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
    loginError.value = error instanceof Error ? error.message : 'Login failed'
  } finally {
    isLoading.value = false
  }
}
</script>

<template>
  <main class="login-page">
    <section class="login-container">
      <div class="login-brand">
        <!-- <div class="brand-mark">C3</div> -->

        <div>
          <h1>Cloud3</h1>
          <p>Managed PostgreSQL Platform</p>
        </div>
      </div>

      <section class="login-card">
        <div class="login-heading">
          <h2>Welcome back</h2>

          <p>Sign in to manage your instances.</p>
        </div>

        <form class="form" @submit.prevent="handleLogin">
          <label>
            Username

            <input
              v-model.trim="username"
              type="text"
              autocomplete="username"
              placeholder="Enter your username"
              required
            />
          </label>

          <label>
            Password

            <input
              v-model="password"
              type="password"
              autocomplete="current-password"
              placeholder="Enter your password"
              required
            />
          </label>

          <p v-if="loginError" class="login-error">
            {{ loginError }}
          </p>

          <button type="submit" :disabled="isLoading">
            {{ isLoading ? 'Signing in...' : 'Sign in' }}
          </button>
        </form>
      </section>
    </section>
  </main>
</template>
