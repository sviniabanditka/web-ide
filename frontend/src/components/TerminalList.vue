<template>
  <div class="terminal-list">
    <div class="tabs-bar">
      <div class="tab-actions">
        <button class="tab-btn" @click="createNewTerminal">
          + Terminal
        </button>
      </div>
      <div class="terminal-tabs">
        <div
          v-for="term in terminalsStore.terminals"
          :key="term.id"
          class="terminal-tab"
          :class="{ active: terminalsStore.currentTerminal?.id === term.id }"
          @click="selectTerminal(term)"
        >
          <span class="tab-title">{{ term.title || 'Terminal' }}</span>
          <button class="close-tab" @click.stop="closeTerminal(term.id)">×</button>
        </div>
      </div>
    </div>

    <div class="content-area">
      <template v-if="terminalsStore.currentTerminal">
        <TerminalPane
          :key="terminalsStore.currentTerminal.id"
          :project-id="project.id"
          :terminal-id="terminalsStore.currentTerminal.id"
          :height="contentHeight"
        />
      </template>
      <div v-else class="no-terminal">
        <p>No terminal open</p>
        <button @click="createNewTerminal" class="create-btn">Create Terminal</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted } from 'vue'
import { useTerminalsStore } from '../stores/terminals'
import TerminalPane from '../components/TerminalPane.vue'

interface Project {
  id: string
  name: string
  root_path: string
}

const props = defineProps<{
  project: Project
}>()

const terminalsStore = useTerminalsStore()

const contentHeight = computed(() => {
  return window.innerHeight - 52 - 44 - 1
})

function selectTerminal(term: any) {
  terminalsStore.setCurrentTerminal(term)
}

async function createNewTerminal() {
  const term = await terminalsStore.createTerminal(props.project.id, 'Terminal', props.project.root_path)
  if (term) {
    terminalsStore.setCurrentTerminal(term)
  }
}

async function closeTerminal(termId: string) {
  const wasCurrent = terminalsStore.currentTerminal?.id === termId
  await terminalsStore.closeTerminal(termId)
  
  if (wasCurrent && terminalsStore.terminals.length > 0) {
    const lastTerminal = terminalsStore.terminals[terminalsStore.terminals.length - 1]
    terminalsStore.setCurrentTerminal(lastTerminal)
  }
}

onMounted(async () => {
  await terminalsStore.fetchTerminals(props.project.id)
  if (terminalsStore.terminals.length > 0 && !terminalsStore.currentTerminal) {
    terminalsStore.setCurrentTerminal(terminalsStore.terminals[0])
  }
})

onUnmounted(() => {
  terminalsStore.setCurrentTerminal(null)
})
</script>

<style scoped>
.terminal-list {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.tabs-bar {
  display: flex;
  align-items: center;
  height: 44px;
  background: #252526;
  border-bottom: 1px solid #3c3c3c;
}

.tab-actions {
  padding: 0 8px;
}

.tab-btn {
  padding: 4px 12px;
  background: #0e639c;
  border: none;
  border-radius: 4px;
  color: #fff;
  font-size: 12px;
}

.tab-btn:hover {
  background: #1177bb;
}

.terminal-tabs {
  display: flex;
  flex: 1;
  overflow-x: auto;
}

.terminal-tab {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 0 16px;
  height: 44px;
  background: #2d2d30;
  border-right: 1px solid #1e1e1e;
  cursor: pointer;
  font-size: 13px;
}

.terminal-tab:hover {
  background: #3c3c3c;
}

.terminal-tab.active {
  background: #1e1e1e;
  border-top: 2px solid #0e639c;
}

.close-tab {
  background: none;
  border: none;
  color: #888;
  font-size: 16px;
  padding: 0;
  line-height: 1;
}

.close-tab:hover {
  color: #fff;
}

.content-area {
  flex: 1;
  overflow: hidden;
}

.no-terminal {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  color: #888;
}

.create-btn {
  margin-top: 16px;
  padding: 8px 16px;
  background: #0e639c;
  border: none;
  border-radius: 4px;
  color: #fff;
  font-size: 14px;
}

.create-btn:hover {
  background: #1177bb;
}
</style>
