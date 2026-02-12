<template>
  <div class="min-h-screen flex items-center justify-center bg-gradient-to-br from-background to-muted p-4">
    <Card class="w-full max-w-md">
      <CardHeader class="text-center">
        <CardTitle class="text-3xl font-bold text-primary">WebIDE</CardTitle>
        <CardDescription class="text-muted-foreground">Self-hosted Browser IDE</CardDescription>
      </CardHeader>
      <CardContent>
        <form @submit.prevent="handleLogin" class="space-y-4">
          <div class="space-y-2">
            <Label for="email">Email</Label>
            <Input
              id="email"
              v-model="email"
              type="email"
              placeholder="test@example.com"
              required
            />
          </div>
          <div class="space-y-2">
            <Label for="password">Password</Label>
            <Input
              id="password"
              v-model="password"
              type="password"
              placeholder="Password"
              required
            />
          </div>
          <Button type="submit" class="w-full" :disabled="authStore.loading">
            {{ authStore.loading ? 'Signing in...' : 'Sign In' }}
          </Button>
          <p v-if="authStore.error" class="text-sm text-destructive text-center">
            {{ authStore.error }}
          </p>
        </form>
      </CardContent>
    </Card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { useProjectsStore } from '../stores/projects'
import Button from '@/components/ui/Button.vue'
import Card from '@/components/ui/Card.vue'
import CardHeader from '@/components/ui/Card.vue'
import CardTitle from '@/components/ui/Card.vue'
import CardDescription from '@/components/ui/Card.vue'
import CardContent from '@/components/ui/Card.vue'
import Input from '@/components/ui/Input.vue'
import Label from '@/components/ui/Label.vue'

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
