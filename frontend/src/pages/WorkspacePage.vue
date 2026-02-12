<template>
  <div class="h-screen flex flex-col bg-background">
    <header class="flex items-center justify-between px-4 h-[52px] border-b bg-card">
      <div class="flex items-center gap-4">
        <router-link to="/projects" class="text-sm text-primary hover:underline">← Projects</router-link>
        <span class="text-sm font-medium">{{ project?.name }}</span>
      </div>
      <Button variant="outline" size="sm" @click="handleLogout">Logout</Button>
    </header>

    <div class="flex bg-card border-b">
      <Button
        :variant="activeTab === 'editor' ? 'secondary' : 'ghost'"
        @click="activeTab = 'editor'"
        class="rounded-none border-b-2 border-transparent data-[state=active]:border-primary"
      >
        Editor
      </Button>
      <Button
        :variant="activeTab === 'terminal' ? 'secondary' : 'ghost'"
        @click="activeTab = 'terminal'"
        class="rounded-none border-b-2 border-transparent data-[state=active]:border-primary"
      >
        Terminal
      </Button>
      <Button
        :variant="activeTab === 'ai' ? 'secondary' : 'ghost'"
        @click="activeTab = 'ai'"
        class="rounded-none border-b-2 border-transparent data-[state=active]:border-primary"
      >
        AI
      </Button>
      <Button
        :variant="activeTab === 'git' ? 'secondary' : 'ghost'"
        @click="activeTab = 'git'"
        class="rounded-none border-b-2 border-transparent data-[state=active]:border-primary"
      >
        Git
      </Button>
    </div>

    <div v-if="loading" class="flex items-center justify-center flex-1 text-muted-foreground">
      Loading...
    </div>
    <div v-else-if="error" class="flex items-center justify-center flex-1 text-destructive">
      {{ error }}
    </div>
    <template v-else-if="project">
      <div class="flex-1 overflow-hidden">
        <EditorPane v-if="activeTab === 'editor'" :project="project" />
        <TerminalWorkspace v-if="activeTab === 'terminal'" :project="project" />
        <AIPane v-if="activeTab === 'ai'" :project="project" />
        <GitPage v-if="activeTab === 'git'" />
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { useProjectsStore } from '../stores/projects'
import { useEditorStore } from '../stores/editor'
import { useTerminalsStore } from '../stores/terminals'
import { api } from '../api'
import EditorPane from './EditorPane.vue'
import TerminalWorkspace from './TerminalWorkspace.vue'
import AIPane from './AIPane.vue'
import GitPage from './GitPage.vue'
import Button from '@/components/ui/Button.vue'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const projectsStore = useProjectsStore()
const editorStore = useEditorStore()
const terminalsStore = useTerminalsStore()

const project = computed(() => projectsStore.currentProject)
const activeTab = ref('terminal')
const loading = ref(true)
const error = ref<string | null>(null)

async function handleLogout() {
  await authStore.logout()
  router.push('/login')
}

async function saveActiveTab() {
  if (!project.value) return
  try {
    await api.put(`/api/v1/projects/${project.value.id}/workspace`, {
      open_files: editorStore.openFiles.map(f => f.path),
      expanded_dirs: Array.from(editorStore.expandedDirs),
      active_file: editorStore.activeFile?.path || null,
      active_tab: activeTab.value,
      open_terminals: terminalsStore.getOpenTerminalIds()
    })
  } catch (e) {
    console.warn('Failed to save active tab:', e)
  }
}

watch(activeTab, () => {
  saveActiveTab()
})

onUnmounted(() => {
  saveActiveTab()
})

onMounted(async () => {
  const projectId = route.params.id as string
  const result = await projectsStore.fetchProject(projectId)
  if (!result) {
    error.value = projectsStore.error || 'Project not found'
    router.push('/projects')
  } else {
    try {
      const savedTab = await editorStore.loadWorkspaceState(projectId)
      if (savedTab) {
        activeTab.value = savedTab
      }
      await terminalsStore.loadTerminalsFromState(projectId, [])
      await editorStore.fetchFileTree(projectId)
    } catch (e: any) {
      console.warn('Failed to load workspace state:', e)
    }
    loading.value = false
  }
})
</script>
