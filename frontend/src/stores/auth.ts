import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import axios from 'axios'

// Module-level API instance - NOT returned through Pinia to avoid reactive proxy wrapping
export const api = axios.create({ baseURL: '/api' })

export const useAuthStore = defineStore('auth', () => {
  const token = ref(localStorage.getItem('token') || '')
  const userId = ref(parseInt(localStorage.getItem('userId') || '0'))
  const username = ref(localStorage.getItem('username') || '')

  const isLoggedIn = computed(() => !!token.value)

  // Set auth header for all requests
  api.interceptors.request.use((config) => {
    if (token.value) {
      config.headers.Authorization = `Bearer ${token.value}`
    }
    return config
  })

  async function register(user: string, pass: string) {
    const res = await api.post('/auth/register', { username: user, password: pass })
    setAuth(res.data)
  }

  async function login(user: string, pass: string) {
    const res = await api.post('/auth/login', { username: user, password: pass })
    setAuth(res.data)
  }

  function setAuth(data: { token: string; user_id: number; username: string }) {
    token.value = data.token
    userId.value = data.user_id
    username.value = data.username
    localStorage.setItem('token', data.token)
    localStorage.setItem('userId', data.user_id.toString())
    localStorage.setItem('username', data.username)
  }

  function logout() {
    token.value = ''
    userId.value = 0
    username.value = ''
    localStorage.removeItem('token')
    localStorage.removeItem('userId')
    localStorage.removeItem('username')
  }

  return { token, userId, username, isLoggedIn, register, login, logout }
})
