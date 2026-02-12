<template>
  <div class="h-screen flex bg-background">
    <aside class="w-[52px] bg-card border-r flex flex-col items-center py-2 gap-1">
      <Tooltip text="Editor" position="right">
        <Button
          :variant="activeTab === 'editor' ? 'secondary' : 'ghost'"
          size="icon"
          @click="activeTab = 'editor'"
          class="w-10 h-10"
        >
          <FileCodeIcon class="w-5 h-5" />
        </Button>
      </Tooltip>

      <Tooltip text="Terminal" position="right">
        <Button
          :variant="activeTab === 'terminal' ? 'secondary' : 'ghost'"
          size="icon"
          @click="activeTab = 'terminal'"
          class="w-10 h-10"
        >
          <TerminalIcon class="w-5 h-5" />
        </Button>
      </Tooltip>

      <Tooltip text="AI" position="right">
        <Button
          :variant="activeTab === 'ai' ? 'secondary' : 'ghost'"
          size="icon"
          @click="activeTab = 'ai'"
          class="w-10 h-10"
        >
          <BotIcon class="w-5 h-5" />
        </Button>
      </Tooltip>

      <Tooltip text="Git" position="right">
        <Button
          :variant="activeTab === 'git' ? 'secondary' : 'ghost'"
          size="icon"
          @click="activeTab = 'git'"
          class="w-10 h-10"
        >
          <GitBranchIcon class="w-5 h-5" />
        </Button>
      </Tooltip>

      <div class="flex-1"></div>

      <Tooltip text="Projects" position="right">
        <Button
          variant="ghost"
          size="icon"
          @click="goToProjects"
          class="w-10 h-10"
        >
          <FolderIcon class="w-5 h-5" />
        </Button>
      </Tooltip>

      <Tooltip text="Logout" position="right">
        <Button
          variant="ghost"
          size="icon"
          @click="handleLogout"
          class="w-10 h-10"
        >
          <LogOutIcon class="w-5 h-5" />
        </Button>
      </Tooltip>
    </aside>

    <div class="flex-1 flex flex-col overflow-hidden">
      <header class="flex items-center justify-between px-4 h-[52px] border-b bg-card">
        <div class="flex items-center gap-2">
          <span class="text-sm font-medium">{{ project?.name }}</span>
        </div>
      </header>

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
import Tooltip from '@/components/ui/Tooltip.vue'
import { FileCode, Terminal, Bot, GitBranch, LogOut, Folder } from 'lucide-vue-next'

const FileCodeIcon = FileCode
const TerminalIcon = Terminal
const BotIcon = Bot
const GitBranchIcon = GitBranch
const LogOutIcon = LogOut
const FolderIcon = Folder

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

function goToProjects() {
  router.push('/projects')
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
