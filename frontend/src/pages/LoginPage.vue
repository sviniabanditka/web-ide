<template>
  <div class="login-page">
    <div class="login-container">
      <h1>WebIDE</h1>
      <p class="subtitle">Self-hosted Browser IDE</p>
      
      <form @submit.prevent="handleLogin">
        <div class="form-group">
          <label for="email">Email</label>
          <input 
            id="email"
            v-model="email" 
            type="email" 
            placeholder="test@example.com"
            required
          />
        </div>
        
        <div class="form-group">
          <label for="password">Password</label>
          <input 
            id="password"
            v-model="password" 
            type="password" 
            placeholder="Password"
            required
          />
        </div>
        
        <button type="submit" :disabled="authStore.loading">
          {{ authStore.loading ? 'Signing in...' : 'Sign In' }}
        </button>
        
        <p v-if="authStore.error" class="error">{{ authStore.error }}</p>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { useProjectsStore } from '../stores/projects'

const router = useRouter()
const authStore = useAuthStore()
const projectsStore = useProjectsStore()

const email = ref('')
const password = ref('')

async function handleLogin() {
  const success = await authStore.login(email.value, password.value)
  if (success) {
    await projectsStore.fetchProjects()
    router.push('/projects')
  }
}

onMounted(async () => {
  if (authStore.token || localStorage.getItem('session_token')) {
    const isAuth = await authStore.checkAuth()
    if (isAuth) {
      await projectsStore.fetchProjects()
      router.push('/projects')
    }
  }
})
</script>

<style scoped>
.login-page {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
  background: linear-gradient(135deg, #1a1a2e 0%, #16213e 100%);
}

.login-container {
  background: #252526;
  padding: 40px;
  border-radius: 8px;
  width: 100%;
  max-width: 400px;
  box-shadow: 0 10px 40px rgba(0,0,0,0.4);
}

h1 {
  color: #4fc3f7;
  margin-bottom: 8px;
}

.subtitle {
  color: #888;
  margin-bottom: 32px;
}

.form-group {
  margin-bottom: 20px;
}

label {
  display: block;
  margin-bottom: 8px;
  color: #ccc;
}

input {
  width: 100%;
  padding: 12px;
  background: #1e1e1e;
  border: 1px solid #3c3c3c;
  border-radius: 4px;
  color: #fff;
  font-size: 14px;
}

input:focus {
  outline: none;
  border-color: #4fc3f7;
}

button {
  width: 100%;
  padding: 12px;
  background: #0e639c;
  border: none;
  border-radius: 4px;
  color: #fff;
  font-size: 14px;
  font-weight: 500;
}

button:hover:not(:disabled) {
  background: #1177bb;
}

button:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.error {
  color: #f44336;
  margin-top: 16px;
  text-align: center;
}
</style>
