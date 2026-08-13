<script setup>
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'

import { useAuthStore } from '../stores/auth'

const router = useRouter()
const auth = useAuthStore()

const username = ref('')
const password = ref('')
const confirmPassword = ref('')

const registerError = ref('')
const isLoading = ref(false)

const passwordLength = computed(() => {
  return Array.from(password.value).length
})

const hasValidLength = computed(() => {
  return passwordLength.value >= 12 && passwordLength.value <= 64
})

const hasUppercase = computed(() => /\p{Lu}/u.test(password.value))
const hasLowercase = computed(() => /\p{Ll}/u.test(password.value))
const hasNumber = computed(() => /\p{Nd}/u.test(password.value))
const hasSymbol = computed(() => /[\p{P}\p{S}]/u.test(password.value))

const passwordsMatch = computed(() => {
  return confirmPassword.value.length > 0 && password.value === confirmPassword.value
})

const passwordIsValid = computed(() => {
  return (
    hasValidLength.value &&
    hasUppercase.value &&
    hasLowercase.value &&
    hasNumber.value &&
    hasSymbol.value
  )
})

const canSubmit = computed(() => {
  return (
    username.value.trim().length > 0 &&
    passwordIsValid.value &&
    passwordsMatch.value &&
    !isLoading.value
  )
})

async function handleRegister() {
  registerError.value = ''

  if (!passwordIsValid.value) {
    registerError.value = 'Password does not meet the requirements.'
    return
  }

  if (!passwordsMatch.value) {
    registerError.value = 'Passwords do not match.'
    return
  }

  isLoading.value = true

  try {
    await auth.register(username.value, password.value)

    await router.replace({
      name: 'login',
      query: {
        registered: '1',
      },
    })
  } catch (error) {
    registerError.value = error instanceof Error ? error.message : 'Registration failed'
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
          <h2>Create your account</h2>

          <p>Create a VetNest account to set up and manage your clinic workspace.</p>
        </div>

        <form class="form" @submit.prevent="handleRegister">
          <label>
            Username

            <input
              v-model.trim="username"
              type="text"
              autocomplete="username"
              placeholder="Choose a username"
              required
            />
          </label>

          <label>
            Password

            <input
              v-model="password"
              type="password"
              autocomplete="new-password"
              placeholder="Create a password"
              required
            />
          </label>

          <div class="password-requirements">
            <p>Password requirements</p>

            <ul>
              <li :class="{ met: hasValidLength }">12–64 characters</li>

              <li :class="{ met: hasUppercase }">One uppercase letter</li>

              <li :class="{ met: hasLowercase }">One lowercase letter</li>

              <li :class="{ met: hasNumber }">One number</li>

              <li :class="{ met: hasSymbol }">One special character</li>
            </ul>
          </div>

          <label>
            Confirm password

            <input
              v-model="confirmPassword"
              type="password"
              autocomplete="new-password"
              placeholder="Repeat your password"
              required
            />
          </label>

          <p v-if="confirmPassword && !passwordsMatch" class="login-error">
            Passwords do not match.
          </p>

          <p v-if="registerError" class="login-error">
            {{ registerError }}
          </p>

          <button type="submit" :disabled="!canSubmit">
            {{ isLoading ? 'Creating account...' : 'Create account' }}
          </button>
        </form>

        <p class="auth-switch">
          Already have an account?
          <RouterLink to="/login">Sign in</RouterLink>
        </p>
      </section>
    </section>
  </main>
</template>
