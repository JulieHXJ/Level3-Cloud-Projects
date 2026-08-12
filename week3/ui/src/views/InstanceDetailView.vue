<script setup>
import { useRoute } from 'vue-router'

const route = useRoute()
const instanceId = route.params.id
</script>

<template>
  <main class="page">
    <header class="page-header">
      <div>
        <h1>Instance Detail</h1>
        <p class="muted">Instance ID: {{ instanceId }}</p>
      </div>

      <RouterLink to="/instances">Back to instances</RouterLink>
    </header>

    <section class="panel">
      <h2>{{ instanceId }}</h2>
      <p>The instance details and connection information will appear here.</p>
    </section>
  </main>
</template>
