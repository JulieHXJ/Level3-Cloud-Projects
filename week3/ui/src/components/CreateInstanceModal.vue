<script setup>
import { ref } from 'vue'

import BaseModal from './BaseModal.vue'
import { useAuthStore } from '../stores/auth'

const emit = defineEmits(['close', 'created'])

const auth = useAuthStore()

const name = ref('')
const instances = ref(1)
const storage = ref('1Gi')
const cpu = ref('250m')

const isSubmitting = ref(false)
const submitError = ref('')

async function handleSubmit() {
  submitError.value = ''
  isSubmitting.value = true

  try {
    const response = await fetch('/api/v1/instances', {
      method: 'POST',

      headers: {
        'Content-Type': 'application/json',
        Authorization: auth.authorizationHeader,
      },

      body: JSON.stringify({
        name: name.value,
        instances: Number(instances.value),
        storage: storage.value,
        cpu: cpu.value,
      }),
    })

    const body = await response.json().catch(() => ({}))

    if (!response.ok) {
      throw new Error(body.message || body.error || `Failed to create instance: ${response.status}`)
    }

    emit('created', body)
    emit('close')
  } catch (error) {
    submitError.value = error instanceof Error ? error.message : 'Failed to create instance'
  } finally {
    isSubmitting.value = false
  }
}
</script>

<template>
  <BaseModal title="Create PostgreSQL instance" @close="emit('close')">
    <p class="modal-description">
      Configure the PostgreSQL instance that will be provisioned on Cloud3.
    </p>

    <form id="create-instance-form" class="form" @submit.prevent="handleSubmit">
      <label>
        Instance name

        <input v-model.trim="name" type="text" placeholder="my-postgres" required />
      </label>

      <label>
        Instances

        <input v-model.number="instances" type="number" min="1" required />
      </label>

      <label>
        Storage

        <input v-model.trim="storage" type="text" placeholder="1Gi" required />
      </label>

      <label>
        CPU

        <input v-model.trim="cpu" type="text" placeholder="250m" required />
      </label>

      <p v-if="submitError" class="login-error">
        {{ submitError }}
      </p>
    </form>

    <template #footer>
      <button
        type="button"
        class="secondary-button"
        :disabled="isSubmitting"
        @click="emit('close')"
      >
        Cancel
      </button>

      <button
        type="submit"
        form="create-instance-form"
        class="primary-button"
        :disabled="isSubmitting"
      >
        {{ isSubmitting ? 'Creating...' : 'Create instance' }}
      </button>
    </template>
  </BaseModal>
</template>
