<script setup>
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { useAuthStore } from '../stores/auth'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

const instanceId = route.params.id

const instance = ref(null)
const connection = ref(null)

const isLoading = ref(false)
const loadError = ref('')

async function authorizedFetch(url) {
  const response = await fetch(url, {
    headers: {
      Authorization: auth.authorizationHeader,
    },
  })

  if (response.status === 401) {
    auth.logout()
    await router.replace('/login')
    return null
  }

  return response
}

async function loadInstance() {
  const response = await authorizedFetch(
    `/api/v1/instances/${encodeURIComponent(instanceId)}`,
  )

  if (!response) {
    return
  }

  const body = await response.json().catch(() => ({}))

  if (!response.ok) {
    throw new Error(
      body.message ||
        body.error ||
        `Failed to load instance: ${response.status}`,
    )
  }

  instance.value = body
}

async function loadConnection() {
  const response = await authorizedFetch(
    `/api/v1/instances/${encodeURIComponent(instanceId)}/connection`,
  )

  if (!response) {
    return
  }

  const body = await response.json().catch(() => ({}))

  /*
   * Connection data may not be available immediately while
   * the PostgreSQL cluster is still provisioning.
   */
  if (response.status === 409 || response.status === 404) {
    connection.value = null
    return
  }

  if (!response.ok) {
    throw new Error(
      body.message ||
        body.error ||
        `Failed to load connection information: ${response.status}`,
    )
  }

  connection.value = body
}

async function loadDetail() {
  isLoading.value = true
  loadError.value = ''

  try {
    await loadInstance()

    /*
     * Do not block the complete detail page just because
     * connection information is temporarily unavailable.
     */
    try {
      await loadConnection()
    } catch (error) {
      console.error(error)
      connection.value = null
    }
  } catch (error) {
    loadError.value =
      error instanceof Error
        ? error.message
        : 'Failed to load instance details'
  } finally {
    isLoading.value = false
  }
}

function goBack() {
  router.push(auth.dashboardPath)
}

function formatDate(value) {
  if (!value) {
    return '—'
  }

  const date = new Date(value)

  if (Number.isNaN(date.getTime())) {
    return value
  }

  return date.toLocaleString()
}

async function copyToClipboard(value) {
  if (!value) {
    return
  }

  await navigator.clipboard.writeText(String(value))
}

onMounted(() => {
  loadDetail()
})
</script>

<template>
  <main class="dashboard-page">
    <header class="detail-header">
      <div>
        <button
          type="button"
          class="back-link"
          @click="goBack"
        >
          ← Back to dashboard
        </button>

        <div class="detail-title-row">
            <div>
                <div class="instance-title-line">
                <h1>
                    {{ instance?.name || instanceId }}
                </h1>

                <span
                    v-if="instance?.status"
                    class="status-badge"
                >
                    Healthy
                </span>
                </div>

                <p class="dashboard-subtitle">
                Instance ID: {{ instanceId }}
                </p>
            </div>
        </div>
      </div>
    </header>

    <p v-if="isLoading" class="state-message">
      Loading instance details...
    </p>

    <p
      v-else-if="loadError"
      class="state-message error-message"
    >
      {{ loadError }}
    </p>

    <template v-else-if="instance">
      <section class="metric-grid">
        <article class="metric-card">
            <span class="metric-label">
            CPU Usage
            </span>

            <strong class="metric-value">
            —
            </strong>

            <span class="metric-secondary">
            Metrics unavailable
            </span>
        </article>

        <article class="metric-card">
            <span class="metric-label">
            Memory Usage
            </span>

            <strong class="metric-value">
            —
            </strong>

            <span class="metric-secondary">
            Metrics unavailable
            </span>
        </article>

        <article class="metric-card">
            <span class="metric-label">
            Replicas
            </span>

            <strong class="metric-value">
            {{ instance.instances }}
            </strong>

            <span class="metric-secondary">
            PostgreSQL instances
            </span>
        </article>

        <article class="metric-card">
            <span class="metric-label">
            Storage
            </span>

            <strong class="metric-value">
            {{ instance.storage || '—' }}
            </strong>

            <span class="metric-secondary">
            Requested capacity
            </span>
        </article>

        <article class="metric-card">
            <span class="metric-label">
            CPU Request
            </span>

            <strong class="metric-value">
            {{ instance.cpu || '—' }}
            </strong>

            <span class="metric-secondary">
            Requested CPU
            </span>
        </article>
      </section>

      

      <section class="dashboard-panel detail-section">
        <div class="section-heading">
          <div>
            <h2>Connection information</h2>

            <p>
              Connection details for applications using this PostgreSQL instance.
            </p>
          </div>
        </div>

        <div
          v-if="connection"
          class="detail-table"
        >
          <div
            v-if="connection.host"
            class="detail-row"
          >
            <span class="detail-label">
              Host
            </span>

            <div class="connection-value">
              <code>{{ connection.host }}</code>

              <button
                type="button"
                class="copy-button"
                @click="copyToClipboard(connection.host)"
              >
                Copy
              </button>
            </div>
          </div>

          <div
            v-if="connection.port"
            class="detail-row"
          >
            <span class="detail-label">
              Port
            </span>

            <div class="connection-value">
              <code>{{ connection.port }}</code>

              <button
                type="button"
                class="copy-button"
                @click="copyToClipboard(connection.port)"
              >
                Copy
              </button>
            </div>
          </div>

          <div
            v-if="connection.database"
            class="detail-row"
          >
            <span class="detail-label">
              Database
            </span>

            <div class="connection-value">
              <code>{{ connection.database }}</code>

              <button
                type="button"
                class="copy-button"
                @click="copyToClipboard(connection.database)"
              >
                Copy
              </button>
            </div>
          </div>

          <div
            v-if="connection.username"
            class="detail-row"
          >
            <span class="detail-label">
              Username
            </span>

            <div class="connection-value">
              <code>{{ connection.username }}</code>

              <button
                type="button"
                class="copy-button"
                @click="copyToClipboard(connection.username)"
              >
                Copy
              </button>
            </div>
          </div>

          <div
            v-if="connection.password"
            class="detail-row"
          >
            <span class="detail-label">
              Password
            </span>

            <div class="connection-value">
              <code>••••••••••••</code>

              <button
                type="button"
                class="copy-button"
                @click="copyToClipboard(connection.password)"
              >
                Copy
              </button>
            </div>
          </div>

          <div
            v-if="connection.uri"
            class="detail-row"
          >
            <span class="detail-label">
              Connection URI
            </span>

            <div class="connection-value connection-uri">
              <code>{{ connection.uri }}</code>

              <button
                type="button"
                class="copy-button"
                @click="copyToClipboard(connection.uri)"
              >
                Copy
              </button>
            </div>
          </div>
        </div>

        <div
          v-else
          class="connection-unavailable"
        >
          <strong>
            Connection information is not available yet.
          </strong>

          <p>
            The PostgreSQL cluster may still be provisioning or the connection
            secret is not ready.
          </p>
        </div>
      </section>
    </template>
  </main>
</template>

<!-- 
using endpoints：
GET /api/v1/instances/db-pxwt6
GET /api/v1/instances/db-pxwt6/connection -->