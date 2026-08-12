<template>
  <main class="page">
    <header class="page-header">
      <div>
        <h1>PostgreSQL Instances</h1>
        <p class="muted">Instance list will be loaded from the Go API.</p>
      </div>

      <RouterLink to="/dashboard">Dashboard</RouterLink>
    </header>

    <section class="panel">
      <h2>Instances</h2>

      <p>API data has not been connected yet.</p>

      <RouterLink class="button-link" to="/instances/example-instance">
        Open example instance
      </RouterLink>
    </section>
  </main>
</template>

<!-- GET /api/v1/instances
POST /api/v1/instances
PATCH /api/v1/instances/{id}
DELETE /api/v1/instances/{id} -->
