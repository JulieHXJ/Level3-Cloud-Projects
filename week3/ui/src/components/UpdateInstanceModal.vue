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

const emit = defineEmits(['close', 'updated'])

const router = useRouter()
const auth = useAuthStore()

const name = ref(props.instance.name)
const instances = ref(props.instance.instances)
const storage = ref(props.instance.storage)
const cpu = ref(props.instance.cpu)

const isSubmitting = ref(false)
const errorMessage = ref('')

async function submitUpdate() {
  errorMessage.value = ''

  const patch = {}

  if (name.value.trim() !== props.instance.name) {
    patch.name = name.value.trim()
  }

  if (Number(instances.value) !== props.instance.instances) {
    patch.instances = Number(instances.value)
  }

  if (storage.value.trim() !== props.instance.storage) {
    patch.storage = storage.value.trim()
  }

  if (cpu.value.trim() !== props.instance.cpu) {
    patch.cpu = cpu.value.trim()
  }

  if (Object.keys(patch).length === 0) {
    errorMessage.value = 'No changes were made.'
    return
  }

  isSubmitting.value = true

  try {
    const response = await fetch(`/api/v1/instances/${encodeURIComponent(props.instance.id)}`, {
      method: 'PATCH',

      headers: {
        'Content-Type': 'application/json',
        Authorization: auth.authorizationHeader,
      },

      body: JSON.stringify(patch),
    })

    const body = await response.json().catch(() => ({}))

    if (response.status === 401) {
      auth.logout()
      await router.replace('/login')
      return
    }

    if (!response.ok) {
      throw new Error(body.message || body.error || `Failed to update instance: ${response.status}`)
    }

    emit('updated', body)
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : 'Failed to update instance'
  } finally {
    isSubmitting.value = false
  }
}
</script>

<template>
  <div class="modal-backdrop" @click.self="emit('close')">
    <section class="instance-modal">
      <header class="modal-header">
        <div>
          <h2>Update instance</h2>
          <p>{{ instance.id }}</p>
        </div>

        <button type="button" class="modal-close" @click="emit('close')">×</button>
      </header>

      <form class="modal-form" @submit.prevent="submitUpdate">
        <label>
          Name
          <input v-model="name" type="text" required />
        </label>

        <label>
          Instances
          <input v-model.number="instances" type="number" min="1" required />
        </label>

        <label>
          Storage
          <input v-model="storage" type="text" required />
        </label>

        <label>
          CPU
          <input v-model="cpu" type="text" required />
        </label>

        <p v-if="errorMessage" class="error-message">
          {{ errorMessage }}
        </p>

        <div class="modal-actions">
          <button
            type="button"
            class="secondary-button"
            :disabled="isSubmitting"
            @click="emit('close')"
          >
            Cancel
          </button>

          <button type="submit" class="primary-button" :disabled="isSubmitting">
            {{ isSubmitting ? 'Updating...' : 'Update instance' }}
          </button>
        </div>
      </form>
    </section>
  </div>
</template>
