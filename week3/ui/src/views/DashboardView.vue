<script setup>
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import CreateInstanceModal from '../components/CreateInstanceModal.vue'

const router = useRouter()
const auth = useAuthStore()

const instances = ref([])
const isLoading = ref(false)
const loadError = ref('')
const showCreateModal = ref(false)

async function loadInstances() {
  isLoading.value = true
  loadError.value = ''

  try {
    const response = await fetch('/api/v1/instances', {
      headers: {
        Authorization: auth.authorizationHeader,
      },
    })

    const body = await response.json().catch(() => ({}))

    if (response.status === 401) {
      auth.logout()
      await router.replace('/login')
      return
    }

    if (!response.ok) {
      throw new Error(
        body.message ||
          body.error ||
          `Failed to load instances: ${response.status}`,
      )
    }

    if (!Array.isArray(body)) {
      throw new Error('Invalid instance list response')
    }

    instances.value = body
  } catch (error) {
    loadError.value =
      error instanceof Error ? error.message : 'Failed to load instances'
  } finally {
    isLoading.value = false
  }
}

function openInstance(instanceId) {
  router.push(`/instances/${instanceId}`)
}

function handleLogout() {
  auth.logout()
  router.push('/login')
}

onMounted(() => {
  loadInstances()
})
</script>

<template>
  <main class="dashboard-page">
    <header class="dashboard-header">
      <div>
        <h1>Cloud3 Dashboard</h1>
        <p class="dashboard-subtitle">
          Signed in as {{ auth.username }}
          <span class="role-badge">{{ auth.role }}</span>
        </p>
      </div>

      <button class="logout-button" @click="handleLogout">
        Logout
      </button>
    </header>

    <section class="dashboard-panel">
      <div class="panel-header">
        <div>
          <h2>PostgreSQL Instances</h2>
          <p class="panel-subtitle">
            {{ auth.isAdmin ? 'Instance administration' : 'Read-only access' }}
          </p>
        </div>

        <button
            v-if="auth.isAdmin"
            class="primary-button"
            type="button"
            @click="showCreateModal = true"
        >
            Create instance
        </button>
      </div>

      <div class="toolbar-cards">
        <div class="toolbar-card stat-card">
          <span class="card-label">Total instances</span>
          <strong class="card-value">{{ instances.length }}</strong>
        </div>

        <button
          v-if="auth.isAdmin"
          type="button"
          class="toolbar-card action-card"
        >
          Update instance
        </button>

        <button
          v-if="auth.isAdmin"
          type="button"
          class="toolbar-card action-card danger-card"
        >
          Delete instance
        </button>
      </div>

      <p v-if="isLoading" class="state-message">Loading instances...</p>

      <p v-else-if="loadError" class="state-message error-message">
        {{ loadError }}
      </p>

      <p v-else-if="instances.length === 0" class="state-message empty-message">
        No PostgreSQL instances found.
      </p>

      <div v-else class="table-wrapper">
        <table class="instance-table">
          <thead>
            <tr>
              <th>ID</th>
              <th>Name</th>
              <th>Instances</th>
              <th>Storage</th>
              <th>CPU</th>
              <th>Status</th>
              <th>Created</th>
            </tr>
          </thead>

          <tbody>
            <tr v-for="instance in instances" :key="instance.id">
              <td>
                <button
                  class="table-link"
                  type="button"
                  @click="openInstance(instance.id)"
                >
                  {{ instance.id }}
                </button>
              </td>

              <td>
                <button
                  class="table-link"
                  type="button"
                  @click="openInstance(instance.id)"
                >
                  {{ instance.name }}
                </button>
              </td>

              <td>{{ instance.instances }}</td>
              <td>{{ instance.storage }}</td>
              <td>{{ instance.cpu }}</td>
              <td>{{ instance.status }}</td>
              <td>{{ new Date(instance.createdAt).toLocaleString() }}</td>
            </tr>
          </tbody>
        </table>
      </div>

      <p v-if="auth.isViewer" class="viewer-note">
        Viewer can view instances and details, but cannot create, update, or delete.
      </p>
    </section>
    <CreateInstanceModal
        v-if="showCreateModal"
        @close="showCreateModal = false"
        @created="loadInstances"
    />
  </main>
</template>

