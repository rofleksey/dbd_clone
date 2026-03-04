<template>
  <div class="auth-page">
    <div class="auth-card">
      <h1 class="title">DEAD BY DAYLIGHT</h1>
      <h2 class="subtitle">CREATE ACCOUNT</h2>
      <form @submit.prevent="handleRegister" class="auth-form">
        <div class="form-group">
          <label>Username</label>
          <input v-model="user" type="text" placeholder="Choose a username" required minlength="3" maxlength="32" />
        </div>
        <div class="form-group">
          <label>Password</label>
          <input v-model="pass" type="password" placeholder="Choose a password" required minlength="4" />
        </div>
        <div class="form-group">
          <label>Confirm Password</label>
          <input v-model="passConfirm" type="password" placeholder="Confirm password" required />
        </div>
        <p v-if="error" class="error">{{ error }}</p>
        <button type="submit" class="btn-primary full-width" :disabled="loading">
          {{ loading ? 'Creating...' : 'Register' }}
        </button>
      </form>
      <p class="switch-link">
        Already have an account? <router-link to="/login">Login</router-link>
      </p>
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
const passConfirm = ref('')
const error = ref('')
const loading = ref(false)

async function handleRegister() {
  if (pass.value !== passConfirm.value) {
    error.value = 'Passwords do not match'
    return
  }
  loading.value = true
  error.value = ''
  try {
    await auth.register(user.value, pass.value)
    router.push('/lobbies')
  } catch (e: any) {
    error.value = e.response?.data?.error || 'Registration failed'
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
.title { font-size: 28px; color: var(--primary); letter-spacing: 4px; margin-bottom: 4px; }
.subtitle { font-size: 14px; color: var(--text-dim); letter-spacing: 4px; margin-bottom: 32px; }
.auth-form { display: flex; flex-direction: column; gap: 16px; }
.form-group { display: flex; flex-direction: column; gap: 6px; text-align: left; }
.form-group label { font-size: 12px; color: var(--text-dim); text-transform: uppercase; letter-spacing: 1px; }
.full-width { width: 100%; padding: 12px; }
.error { color: var(--danger); font-size: 13px; }
.switch-link { margin-top: 20px; font-size: 13px; color: var(--text-dim); }
</style>
