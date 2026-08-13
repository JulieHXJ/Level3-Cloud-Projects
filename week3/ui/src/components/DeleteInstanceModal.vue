<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const props = defineProps({
  instance: {
    type: Object,
    required: true,
  },
})

const emit = defineEmits(['close', 'deleted'])

const router = useRouter()
const auth = useAuthStore()

const isDeleting = ref(false)
const errorMessage = ref('')

async function deleteInstance() {
  errorMessage.value = ''
  isDeleting.value = true

  try {
    const response = await fetch(`/api/v1/instances/${encodeURIComponent(props.instance.id)}`, {
      method: 'DELETE',

      headers: {
        Authorization: auth.authorizationHeader,
      },
    })

    if (response.status === 401) {
      auth.logout()
      await router.replace('/login')
      return
    }

    if (!response.ok) {
      const body = await response.json().catch(() => ({}))

      throw new Error(body.message || body.error || `Failed to delete instance: ${response.status}`)
    }

    emit('deleted')
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : 'Failed to delete instance'
  } finally {
    isDeleting.value = false
  }
}
</script>

<template>
  <div class="modal-backdrop" @click.self="emit('close')">
    <section class="instance-modal">
      <header class="modal-header">
        <div>
          <h2>Delete instance</h2>
          <p>{{ instance.id }}</p>
        </div>

        <button type="button" class="modal-close" @click="emit('close')">×</button>
      </header>

      <div class="modal-form">
        <p>
          Are you sure you want to delete
          <strong>{{ instance.name }}</strong
          >?
        </p>

        <p>This will delete the managed PostgreSQL instance.</p>

        <p v-if="errorMessage" class="error-message">
          {{ errorMessage }}
        </p>

        <div class="modal-actions">
          <button
            type="button"
            class="secondary-button"
            :disabled="isDeleting"
            @click="emit('close')"
          >
            Cancel
          </button>

          <button
            type="button"
            class="danger-button"
            :disabled="isDeleting"
            @click="deleteInstance"
          >
            {{ isDeleting ? 'Deleting...' : 'Delete instance' }}
          </button>
        </div>
      </div>
    </section>
  </div>
</template>
