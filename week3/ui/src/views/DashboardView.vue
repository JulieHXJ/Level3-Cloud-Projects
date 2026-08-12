<script setup>
import { useRouter } from 'vue-router'

import { useAuthStore } from '../stores/auth'

const router = useRouter()
const auth = useAuthStore()

function handleLogout() {
  auth.logout()
  router.push('/login')
}
</script>

<template>
  <main class="page">
    <header class="page-header">
      <div>
        <h1>Cloud3 Dashboard</h1>

        <p class="muted">
          Signed in as {{ auth.username }}
          <span class="role-badge">{{ auth.role }}</span>
        </p>
      </div>

      <button class="secondary-button" @click="handleLogout">
        Logout
      </button>
    </header>

    <section class="panel">
      <div class="dashboard-title">
        <div>
          <h2>PostgreSQL Instances</h2>
          <p class="muted">
            {{ auth.isAdmin ? 'Instance administration' : 'Read-only access' }}
          </p>
        </div>

        <button v-if="auth.isAdmin" type="button">
          Create instance
        </button>
      </div>

      <p>下一步将在这里加入 Instance 总数和可滚动列表。</p>

      <div v-if="auth.isAdmin" class="actions">
        <button type="button">Update instance</button>
        <button type="button" class="danger-button">Delete instance</button>
      </div>

      <p v-else class="notice">
        Viewer 可以查看 Instance 和详情，但不能创建、更新或删除。
      </p>
    </section>
  </main>
</template>