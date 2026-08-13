<script setup>
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'

import { useAuthStore } from '../stores/auth'

import CreateInstanceModal from '../components/CreateInstanceModal.vue'
import UpdateInstanceModal from '../components/UpdateInstanceModal.vue'
import DeleteInstanceModal from '../components/DeleteInstanceModal.vue'

const router = useRouter()
const auth = useAuthStore()

const instances = ref([])
const selectedInstance = ref(null)

const isLoading = ref(false)
const loadError = ref('')

const showCreateModal = ref(false)
const showUpdateModal = ref(false)
const showDeleteModal = ref(false)

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
        body.message || body.error || `Failed to load clinic workspaces: ${response.status}`,
      )
    }

    if (!Array.isArray(body)) {
      throw new Error('Invalid clinic workspace response')
    }

    instances.value = body

    if (
      selectedInstance.value &&
      !instances.value.some((instance) => instance.id === selectedInstance.value.id)
    ) {
      selectedInstance.value = null
    }
  } catch (error) {
    loadError.value = error instanceof Error ? error.message : 'Failed to load clinic workspaces'
  } finally {
    isLoading.value = false
  }
}

function selectInstance(instance) {
  selectedInstance.value = instance
}

function openInstance(instanceId) {
  router.push(`/instances/${instanceId}`)
}

function openUpdateModal() {
  if (!selectedInstance.value) {
    return
  }

  showUpdateModal.value = true
}

function openDeleteModal() {
  if (!selectedInstance.value) {
    return
  }

  showDeleteModal.value = true
}

async function handleCreated() {
  showCreateModal.value = false
  await loadInstances()
}

async function handleUpdated() {
  showUpdateModal.value = false
  selectedInstance.value = null
  await loadInstances()
}

async function handleDeleted() {
  showDeleteModal.value = false
  selectedInstance.value = null
  await loadInstances()
}

async function handleLogout() {
  auth.logout()
  await router.push('/login')
}

onMounted(() => {
  loadInstances()
})
</script>

<template>
  <main class="dashboard-page">
    <header class="dashboard-header">
      <div>
        <h1>VetNest</h1>

        <p class="dashboard-subtitle">
          Veterinary Health Platform
        </p>

        <p class="dashboard-subtitle dashboard-user">
          <span class="role-badge">
            {{ auth.isAdmin ? 'Platform Admin' : 'Clinic Admin' }}
          </span>
        </p>
      </div>

      <button
        class="logout-button"
        type="button"
        @click="handleLogout"
      >
        Logout
      </button>
    </header>

    <section class="dashboard-panel">
      <div class="panel-header">
        <div>
          <h2>
            {{ auth.isAdmin ? 'Clinic Infrastructure' : 'My Clinic Workspaces' }}
          </h2>

          <p class="panel-subtitle">
            {{
              auth.isAdmin
                ? 'Manage clinic workspaces and their database environments.'
                : 'Create and manage your veterinary clinic workspace.'
            }}
          </p>
        </div>

        <button class="primary-button" type="button" @click="showCreateModal = true">
          + Create clinic
        </button>
      </div>

      <div class="toolbar-cards">
        <div class="toolbar-card stat-card">
          <span class="card-label">
            {{ auth.isAdmin ? 'Total clinics' : 'My clinics' }}
          </span>

          <strong class="card-value">
            {{ instances.length }}
          </strong>
        </div>

        <button
          type="button"
          class="toolbar-card action-card"
          :disabled="!selectedInstance"
          @click="openUpdateModal"
        >
          Update clinic
        </button>

        <button
          type="button"
          class="toolbar-card action-card danger-card"
          :disabled="!selectedInstance"
          @click="openDeleteModal"
        >
          Delete clinic
        </button>
      </div>

      <p v-if="isLoading" class="state-message">Loading clinic workspaces...</p>

      <p v-else-if="loadError" class="state-message error-message">
        {{ loadError }}
      </p>

      <p v-else-if="instances.length === 0" class="state-message empty-message">
        {{
          auth.isAdmin
            ? 'No clinic workspaces have been created yet.'
            : 'You do not have a clinic workspace yet. Create one to get started.'
        }}
      </p>

      <div v-else class="table-wrapper">
        <table class="instance-table">
          <thead>
            <tr>
              <th>Select</th>
              <th>Clinic ID</th>
              <th>Clinic Name</th>
              <th>DB Nodes</th>
              <th>Storage</th>
              <th>CPU</th>
              <th>Status</th>
              <th>Created</th>
            </tr>
          </thead>

          <tbody>
            <tr
              v-for="instance in instances"
              :key="instance.id"
              :class="{
                'selected-row': selectedInstance?.id === instance.id,
              }"
            >
              <td>
                <input
                  type="radio"
                  name="selected-clinic"
                  :checked="selectedInstance?.id === instance.id"
                  :aria-label="`Select ${instance.name}`"
                  @change="selectInstance(instance)"
                />
              </td>

              <td>
                <button class="table-link" type="button" @click="openInstance(instance.id)">
                  {{ instance.id }}
                </button>
              </td>

              <td>
                <button class="table-link" type="button" @click="openInstance(instance.id)">
                  {{ instance.name }}
                </button>
              </td>

              <td>
                {{ instance.instances }}
              </td>

              <td>
                {{ instance.storage }}
              </td>

              <td>
                {{ instance.cpu }}
              </td>

              <td>
                {{ instance.status }}
              </td>

              <td>
                {{ new Date(instance.createdAt).toLocaleString() }}
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <p v-if="auth.isUser" class="user-note">
        Each clinic workspace is backed by an isolated PostgreSQL environment managed by VetNest.
      </p>

      <p v-if="selectedInstance" class="user-note">
        Selected clinic:
        <strong>{{ selectedInstance.name }}</strong>
      </p>
    </section>

    <CreateInstanceModal
      v-if="showCreateModal"
      @close="showCreateModal = false"
      @created="handleCreated"
    />

    <UpdateInstanceModal
      v-if="showUpdateModal && selectedInstance"
      :instance="selectedInstance"
      @close="showUpdateModal = false"
      @updated="handleUpdated"
    />

    <DeleteInstanceModal
      v-if="showDeleteModal && selectedInstance"
      :instance="selectedInstance"
      @close="showDeleteModal = false"
      @deleted="handleDeleted"
    />
  </main>
</template>
