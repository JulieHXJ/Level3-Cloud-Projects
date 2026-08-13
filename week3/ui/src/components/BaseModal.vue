<script setup>
defineProps({
  title: {
    type: String,
    required: true,
  },
})

const emit = defineEmits(['close'])
</script>

<template>
  <Teleport to="body">
    <div class="modal-overlay" @click.self="emit('close')">
      <section class="modal-card">
        <header class="modal-header">
          <h2>{{ title }}</h2>

          <button
            type="button"
            class="modal-close-button"
            aria-label="Close"
            @click="emit('close')"
          >
            ×
          </button>
        </header>

        <div class="modal-body">
          <slot />
        </div>

        <footer v-if="$slots.footer" class="modal-footer">
          <slot name="footer" />
        </footer>
      </section>
    </div>
  </Teleport>
</template>
