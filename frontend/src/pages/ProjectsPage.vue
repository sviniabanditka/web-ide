<template>
  <div class="h-screen flex flex-col bg-background">
    <header class="flex items-center justify-between px-6 h-[52px] border-b bg-card">
      <div class="flex items-center gap-4">
        <span class="text-sm font-medium">Projects</span>
      </div>
      <Button variant="outline" size="sm" @click="handleLogout">Logout</Button>
    </header>

    <main class="flex-1 overflow-auto">
      <div v-if="projectsStore.loading" class="flex items-center justify-center h-full text-muted-foreground">
        Loading...
      </div>

      <div v-else-if="projectsStore.error" class="flex items-center justify-center h-full text-destructive">
        {{ projectsStore.error }}
      </div>

      <div v-else-if="projectsStore.projects.length === 0" class="flex items-center justify-center h-full text-muted-foreground">
        No projects found. Add projects to {{ projectsDir }} directory.
      </div>

      <div v-else class="flex flex-col">
        <div
          v-for="project in projectsStore.projects"
          :key="project.id"
          class="group flex items-center gap-4 px-4 py-3 border-b cursor-pointer hover:bg-accent/50 transition-colors"
          @click="openProject(project.id)"
        >
          <div class="w-10 h-10 rounded-lg bg-primary/10 flex items-center justify-center shrink-0">
            <FolderIcon class="w-5 h-5 text-primary" />
          </div>
          <div class="flex-1 min-w-0">
            <div class="flex items-center gap-2">
              <h3 class="font-medium truncate">{{ project.name }}</h3>
              <span class="text-xs px-2 py-0.5 rounded-full bg-muted text-muted-foreground">Active</span>
            </div>
            <p class="text-sm text-muted-foreground truncate">{{ project.root_path }}</p>
          </div>
          <div class="flex items-center gap-2 text-xs text-muted-foreground">
            <ClockIcon class="w-3 h-3" />
            <span>2 hours ago</span>
          </div>
          <ChevronRightIcon class="w-4 h-4 text-muted-foreground opacity-0 group-hover:opacity-100 transition-opacity" />
        </div>
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
import { Folder, ChevronRight, Clock } from 'lucide-vue-next'

const router = useRouter()
const authStore = useAuthStore()
const projectsStore = useProjectsStore()

const projectsDir = '/projects'

const FolderIcon = Folder
const ChevronRightIcon = ChevronRight
const ClockIcon = Clock

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
