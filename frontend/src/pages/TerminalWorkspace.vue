<template>
  <div class="h-full flex">
    <aside class="w-[220px] bg-card border-r flex flex-col">
      <div class="px-3 py-2 border-b">
        <Button size="sm" @click="createNewTerminal" class="w-full">
          <PlusIcon class="w-4 h-4 mr-2" />
          New Terminal
        </Button>
      </div>
      <div class="flex-1 overflow-y-auto">
        <div
          v-for="term in terminalsStore.terminals"
          :key="term.id"
          class="group flex items-center gap-2 px-3 py-1.5 text-sm cursor-pointer border-l-2 border-transparent hover:bg-accent/50"
          :class="{ 'bg-accent border-l-primary': terminalsStore.currentTerminal?.id === term.id }"
          @click="selectTerminal(term)"
        >
          <TerminalIcon class="w-4 h-4 text-muted-foreground shrink-0" />
          <span class="truncate flex-1">{{ term.title || 'Terminal' }}</span>
          <button
            class="text-muted-foreground hover:text-foreground opacity-0 group-hover:opacity-100 p-0.5 rounded"
            @click.stop="closeTerminal(term.id)"
          >
            <XIcon class="w-3 h-3" />
          </button>
        </div>
        <div v-if="terminalsStore.terminals.length === 0" class="p-4 text-center text-sm text-muted-foreground">
          No terminals
        </div>
      </div>
    </aside>

    <div class="flex-1 flex flex-col overflow-hidden">
      <div class="flex-1 overflow-hidden">
        <TerminalPane
          v-if="terminalsStore.currentTerminal"
          :key="terminalsStore.currentTerminal.id"
          :project-id="project.id"
          :terminal-id="terminalsStore.currentTerminal.id"
          :height="contentHeight"
        />
        <div v-else class="h-full flex flex-col items-center justify-center text-muted-foreground">
          <TerminalIcon class="w-12 h-12 mb-2 opacity-50" />
          <p>No terminal open</p>
          <Button @click="createNewTerminal" class="mt-2">Create Terminal</Button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted } from 'vue'
import { useTerminalsStore } from '../stores/terminals'
import TerminalPane from '../components/TerminalPane.vue'
import Button from '@/components/ui/Button.vue'
import { Terminal, Plus, X } from 'lucide-vue-next'

interface Project {
  id: string
  name: string
  root_path: string
}

const props = defineProps<{
  project: Project
}>()

const terminalsStore = useTerminalsStore()

const TerminalIcon = Terminal
const PlusIcon = Plus
const XIcon = X

const contentHeight = computed(() => {
  return window.innerHeight - 52
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
