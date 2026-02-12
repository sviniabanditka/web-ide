<template>
  <div class="h-screen flex flex-col">
    <header class="flex items-center justify-between px-6 py-3 border-b bg-card">
      <h1 class="text-lg font-semibold">Projects</h1>
      <Button variant="outline" size="sm" @click="handleLogout">Logout</Button>
    </header>
    
    <main class="flex-1 overflow-auto p-6">
      <div v-if="projectsStore.loading" class="flex items-center justify-center h-full text-muted-foreground">
        Loading...
      </div>
      
      <div v-else-if="projectsStore.error" class="flex items-center justify-center h-full text-destructive">
        {{ projectsStore.error }}
      </div>
      
      <div v-else-if="projectsStore.projects.length === 0" class="flex items-center justify-center h-full text-muted-foreground">
        No projects found. Add projects to {{ projectsDir }} directory.
      </div>
      
      <div v-else class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
        <Card
          v-for="project in projectsStore.projects"
          :key="project.id"
          class="cursor-pointer hover:border-primary transition-colors"
          @click="openProject(project.id)"
        >
          <CardContent class="flex items-center gap-4 p-4">
            <div class="text-3xl">📁</div>
            <div class="flex-1 min-w-0">
              <h3 class="font-medium truncate">{{ project.name }}</h3>
              <p class="text-sm text-muted-foreground truncate">{{ project.root_path }}</p>
            </div>
          </CardContent>
        </Card>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { useProjectsStore } from '../stores/projects'
import Button from '@/components/ui/Button.vue'
import Card from '@/components/ui/Card.vue'
import CardContent from '@/components/ui/Card.vue'

const router = useRouter()
const authStore = useAuthStore()
const projectsStore = useProjectsStore()

const projectsDir = '/projects'

async function handleLogout() {
  await authStore.logout()
  router.push('/login')
}

function openProject(id: string) {
  router.push(`/projects/${id}`)
}

onMounted(async () => {
  await authStore.checkAuth()
  await projectsStore.fetchProjects()
})
</script>
