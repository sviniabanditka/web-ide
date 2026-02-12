<script setup lang="ts">
interface ToolCall {
  id: string
  name: string
  arguments: Record<string, unknown>
  summary?: string
}

const props = defineProps<{
  tool: ToolCall
}>()

const emit = defineEmits<{
  approve: []
  reject: [reason?: string]
}>()

function formatArguments(args: Record<string, unknown>): string {
  try {
    return JSON.stringify(args, null, 2)
  } catch {
    return String(args)
  }
}

function getToolIcon(name: string): string {
  const icons: Record<string, string> = {
    read_file: '📄',
    list_dir: '📁',
    search_in_files: '🔍',
    apply_patch: '✏️',
    run_command: '⚡',
    get_command_output: '📊',
    cancel_command: '🛑',
  }
  return icons[name] || '🔧'
}

function onApprove() {
  emit('approve')
}

function onReject() {
  const reason = prompt('Reason for rejection (optional):') || undefined
  emit('reject', reason)
}
</script>

<template>
  <div class="rounded-lg border-2 border-amber-500 p-4 bg-gradient-to-br from-amber-950/30 to-background">
    <div class="flex items-center gap-2 mb-3 pb-3 border-b border-amber-500/30">
      <span class="text-xl">⚠️</span>
      <span class="font-semibold text-amber-500 uppercase tracking-wide text-sm">Approval Required</span>
    </div>
    
    <div class="space-y-3 mb-4">
      <div class="flex items-center gap-2">
        <span class="text-lg">{{ getToolIcon(tool.name) }}</span>
        <span class="font-semibold">{{ tool.name }}</span>
      </div>
      
      <div v-if="tool.summary" class="text-sm text-muted-foreground italic">
        {{ tool.summary }}
      </div>
      
      <div class="bg-muted/50 rounded p-2">
        <div class="text-xs text-muted-foreground uppercase mb-1">Arguments:</div>
        <pre class="font-mono text-xs text-muted-foreground whitespace-pre-wrap break-all">{{ formatArguments(tool.arguments) }}</pre>
      </div>
    </div>
    
    <div class="flex gap-3">
      <Button class="flex-1 bg-green-600 hover:bg-green-700" @click="onApprove">
        ✓ Approve
      </Button>
      <Button variant="secondary" class="flex-1" @click="onReject">
        ✕ Reject
      </Button>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue'
import Button from '@/components/ui/Button.vue'

export default defineComponent({
  components: { Button }
})
</script>
