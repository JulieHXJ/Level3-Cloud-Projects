<script setup>
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { useAuthStore } from '../stores/auth'

const router = useRouter()
const route = useRoute()
const auth = useAuthStore()

const username = ref('')
const password = ref('')
const loginError = ref('')
const isLoading = ref(false)

const registrationSuccess = computed(() => {
  return route.query.registered === '1'
})

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
        <div>
          <h1>VetNest</h1>
          <p>Veterinary Health Platform</p>
        </div>
      </div>

      <section class="login-card">
        <div class="login-heading">
          <h2>Welcome back</h2>

          <p>Sign in to access your clinic workspace and manage your veterinary data.</p>
        </div>

        <p v-if="registrationSuccess" class="login-success">
          Account created successfully. You can now sign in.
        </p>

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

        <p class="auth-switch">
          New to VetNest?
          <RouterLink to="/register">Create an account</RouterLink>
        </p>
      </section>
    </section>
  </main>
</template>
