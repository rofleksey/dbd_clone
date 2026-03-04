<template>
  <div class="auth-page">
    <div class="auth-card">
      <h1 class="title">DEAD BY DAYLIGHT</h1>
      <h2 class="subtitle">CLONE</h2>
      <form @submit.prevent="handleLogin" class="auth-form">
        <div class="form-group">
          <label>Username</label>
          <input v-model="user" type="text" placeholder="Enter username" required minlength="3" maxlength="32" />
        </div>
        <div class="form-group">
          <label>Password</label>
          <input v-model="pass" type="password" placeholder="Enter password" required minlength="4" />
        </div>
        <p v-if="error" class="error">{{ error }}</p>
        <button type="submit" class="btn-primary full-width" :disabled="loading">
          {{ loading ? 'Logging in...' : 'Login' }}
        </button>
      </form>
      <p class="switch-link">
        Don't have an account? <router-link to="/register">Register</router-link>
      </p>
      <div class="nav-links">
        <router-link to="/leaderboard">Leaderboard</router-link>
        <router-link to="/games">Live Games</router-link>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const auth = useAuthStore()

const user = ref('')
const pass = ref('')
const error = ref('')
const loading = ref(false)

async function handleLogin() {
  loading.value = true
  error.value = ''
  try {
    await auth.login(user.value, pass.value)
    router.push('/lobbies')
  } catch (e: any) {
    error.value = e.response?.data?.error || 'Login failed'
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.auth-page {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100vh;
  background: linear-gradient(135deg, #0a0a0a 0%, #1a0a0a 50%, #0a0a0a 100%);
}

.auth-card {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: 8px;
  padding: 40px;
  width: 400px;
  text-align: center;
}

.title {
  font-size: 28px;
  color: var(--primary);
  letter-spacing: 4px;
  margin-bottom: 4px;
}

.subtitle {
  font-size: 14px;
  color: var(--text-dim);
  letter-spacing: 8px;
  margin-bottom: 32px;
}

.auth-form {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
  text-align: left;
}

.form-group label {
  font-size: 12px;
  color: var(--text-dim);
  text-transform: uppercase;
  letter-spacing: 1px;
}

.full-width {
  width: 100%;
  padding: 12px;
}

.error {
  color: var(--danger);
  font-size: 13px;
}

.switch-link {
  margin-top: 20px;
  font-size: 13px;
  color: var(--text-dim);
}

.nav-links {
  margin-top: 16px;
  display: flex;
  gap: 20px;
  justify-content: center;
  font-size: 13px;
}
</style>
