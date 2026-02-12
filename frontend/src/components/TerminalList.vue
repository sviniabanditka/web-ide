<template>
  <div class="h-full flex flex-col">
    <div class="flex items-center h-11 bg-card border-b">
      <div class="px-2">
        <Button size="sm" @click="createNewTerminal">+ Terminal</Button>
      </div>
      <div class="flex-1 flex overflow-x-auto">
        <div
          v-for="term in terminalsStore.terminals"
          :key="term.id"
          class="flex items-center gap-2 px-4 py-2 text-sm cursor-pointer border-r border-background hover:bg-accent"
          :class="{ 'bg-background border-t-2 border-t-primary': terminalsStore.currentTerminal?.id === term.id }"
          @click="selectTerminal(term)"
        >
          <span class="truncate">{{ term.title || 'Terminal' }}</span>
          <button class="text-muted-foreground hover:text-foreground" @click.stop="closeTerminal(term.id)">×</button>
        </div>
      </div>
    </div>

    <div class="flex-1 overflow-hidden">
      <TerminalPane
        v-if="terminalsStore.currentTerminal"
        :key="terminalsStore.currentTerminal.id"
        :project-id="project.id"
        :terminal-id="terminalsStore.currentTerminal.id"
        :height="contentHeight"
      />
      <div v-else class="h-full flex flex-col items-center justify-center text-muted-foreground">
        <p class="mb-2">No terminal open</p>
        <Button @click="createNewTerminal">Create Terminal</Button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted } from 'vue'
import { useTerminalsStore } from '../stores/terminals'
import TerminalPane from '../components/TerminalPane.vue'
import Button from '@/components/ui/Button.vue'

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
