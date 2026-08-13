import { computed, ref } from 'vue'
import { defineStore } from 'pinia'

export const useAuthStore = defineStore('auth', () => {
  const token = ref(sessionStorage.getItem('vetnest-token') || '')
  const tokenType = ref(sessionStorage.getItem('vetnest-token-type') || 'Bearer')
  const role = ref(sessionStorage.getItem('vetnest-role') || '')
  const username = ref(sessionStorage.getItem('vetnest-username') || '')

  const isAuthenticated = computed(() => {
    return Boolean(token.value && (role.value === 'admin' || role.value === 'user'))
  })

  const isAdmin = computed(() => role.value === 'admin')
  const isUser = computed(() => role.value === 'user')

  const dashboardPath = computed(() => {
    if (role.value === 'admin') {
      return '/admin/dashboard'
    }

    if (role.value === 'user') {
      return '/user/dashboard'
    }

    return '/login'
  })

  const authorizationHeader = computed(() => {
    if (!token.value) {
      return ''
    }

    return `${tokenType.value} ${token.value}`
  })

  async function register(registerUsername, password) {
    const response = await fetch('/api/v1/auth/register', {
      method: 'POST',

      headers: {
        'Content-Type': 'application/json',
      },

      body: JSON.stringify({
        username: registerUsername.trim(),
        password,
      }),
    })

    const body = await response.json().catch(() => ({}))

    if (!response.ok) {
      throw new Error(body.message || body.error || `Registration failed: ${response.status}`)
    }

    return body
  }

  async function login(loginUsername, password) {
    const response = await fetch('/api/v1/auth/login', {
      method: 'POST',

      headers: {
        'Content-Type': 'application/json',
      },

      body: JSON.stringify({
        username: loginUsername.trim(),
        password,
      }),
    })

    const body = await response.json().catch(() => ({}))

    if (!response.ok) {
      throw new Error(body.message || body.error || `Login failed: ${response.status}`)
    }

    if (!body.token) {
      throw new Error('Login response does not contain a token')
    }

    if (body.role !== 'admin' && body.role !== 'user') {
      throw new Error('Login response contains an unsupported role')
    }

    token.value = body.token
    tokenType.value = body.tokenType || 'Bearer'
    role.value = body.role
    username.value = loginUsername.trim()

    sessionStorage.setItem('vetnest-token', token.value)
    sessionStorage.setItem('vetnest-token-type', tokenType.value)
    sessionStorage.setItem('vetnest-role', role.value)
    sessionStorage.setItem('vetnest-username', username.value)
  }

  function logout() {
    token.value = ''
    tokenType.value = 'Bearer'
    role.value = ''
    username.value = ''

    sessionStorage.removeItem('vetnest-token')
    sessionStorage.removeItem('vetnest-token-type')
    sessionStorage.removeItem('vetnest-role')
    sessionStorage.removeItem('vetnest-username')
  }

  return {
    token,
    tokenType,
    role,
    username,

    isAuthenticated,
    isAdmin,
    isUser,

    dashboardPath,
    authorizationHeader,

    register,
    login,
    logout,
  }
})
