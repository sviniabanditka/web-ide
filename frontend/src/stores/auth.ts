import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { api } from '../api'

interface User {
  id: string
  email: string
  created_at: string
}

export const useAuthStore = defineStore('auth', () => {
  const user = ref<User | null>(null)
  const token = ref<string | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  const isAuthenticated = computed(() => !!token.value)

  async function login(email: string, password: string) {
    loading.value = true
    error.value = null
    
    localStorage.removeItem('session_token')
    token.value = null
    
    try {
      const response = await api.post('/api/v1/auth/login', { email, password })
      user.value = response.data.user
      
      const sessionToken = response.data.token
      if (sessionToken) {
        token.value = sessionToken
        localStorage.setItem('session_token', sessionToken)
      }
      return true
    } catch (e: any) {
      error.value = e.response?.data?.error || 'Login failed'
      return false
    } finally {
      loading.value = false
    }
  }

  async function logout() {
    try {
      await api.post('/api/v1/auth/logout')
    } catch (e) {
      // Ignore errors
    }
    user.value = null
    token.value = null
    localStorage.removeItem('session_token')
  }

  async function checkAuth() {
    const storedToken = localStorage.getItem('session_token')
    if (!storedToken) return false

    token.value = storedToken
    
    try {
      const response = await api.get('/api/v1/auth/me')
      user.value = response.data
      return true
    } catch {
      logout()
      return false
    }
  }

  return { user, token, loading, error, isAuthenticated, login, logout, checkAuth }
})
