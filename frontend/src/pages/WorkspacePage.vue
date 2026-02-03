<template>
  <div class="workspace-page">
    <header class="header">
      <div class="header-left">
        <router-link to="/projects" class="back-btn">← Projects</router-link>
        <span class="project-name">{{ project?.name }}</span>
      </div>
      <div class="header-right">
        <button @click="handleLogout" class="logout-btn">Logout</button>
      </div>
    </header>

    <div class="workspace-tabs">
      <button
        class="tab-btn"
        :class="{ active: activeTab === 'editor' }"
        @click="activeTab = 'editor'"
      >
        Editor
      </button>
      <button
        class="tab-btn"
        :class="{ active: activeTab === 'terminal' }"
        @click="activeTab = 'terminal'"
      >
        Terminal
      </button>
      <button
        class="tab-btn"
        :class="{ active: activeTab === 'ai' }"
        @click="activeTab = 'ai'"
      >
        AI
      </button>
      <button
        class="tab-btn"
        :class="{ active: activeTab === 'git' }"
        @click="activeTab = 'git'"
      >
        Git
      </button>
    </div>

    <div v-if="loading" class="loading">Loading...</div>
    <div v-else-if="error" class="error">{{ error }}</div>
    <template v-else-if="project">
      <div class="workspace-content">
        <EditorPane v-if="activeTab === 'editor'" :project="project" />
        
        <TerminalWorkspace v-if="activeTab === 'terminal'" :project="project" />
        
        <AIPane v-if="activeTab === 'ai'" :project="project" />
        
        <GitPage v-if="activeTab === 'git'" />
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed, watch } from 'vue'
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

<style scoped>
.workspace-page {
  height: 100%;
  display: flex;
  flex-direction: column;
  background: #1e1e1e;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 16px;
  height: 52px;
  background: #252526;
  border-bottom: 1px solid #3c3c3c;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 16px;
}

.back-btn {
  color: #4fc3f7;
  text-decoration: none;
  font-size: 13px;
}

.back-btn:hover {
  text-decoration: underline;
}

.project-name {
  font-size: 14px;
  font-weight: 500;
}

.logout-btn {
  padding: 6px 12px;
  background: transparent;
  border: 1px solid #3c3c3c;
  border-radius: 4px;
  color: #ccc;
  font-size: 12px;
}

.logout-btn:hover {
  background: #3c3c3c;
}

.workspace-tabs {
  display: flex;
  background: #252526;
  border-bottom: 1px solid #3c3c3c;
}

.tab-btn {
  padding: 12px 24px;
  background: none;
  border: none;
  border-bottom: 2px solid transparent;
  color: #888;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s;
}

.tab-btn:hover {
  color: #ccc;
}

.tab-btn.active {
  color: #fff;
  border-bottom-color: #0e639c;
}

.workspace-content {
  flex: 1;
  overflow: hidden;
}

.loading, .error {
  padding: 40px;
  text-align: center;
  color: #888;
}

.error {
  color: #f44336;
}
</style>
