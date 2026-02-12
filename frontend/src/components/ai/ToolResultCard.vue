<script setup lang="ts">
interface ToolResult {
  ok: boolean
  data?: unknown
  error?: {
    code: string
    message: string
  }
  meta?: {
    duration_ms: number
    truncated?: boolean
  }
}

interface Props {
  toolName: string
  toolCallId: string
  result: ToolResult
}

const props = defineProps<Props>()

function formatResult(data: unknown): string {
  if (data === undefined || data === null) {
    return 'No data'
  }
  try {
    return JSON.stringify(data, null, 2)
  } catch {
    return String(data)
  }
}

function formatDuration(ms: number): string {
  if (ms < 1000) {
    return `${ms}ms`
  }
  return `${(ms / 1000).toFixed(2)}s`
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
</script>

<template>
  <div class="rounded-lg border p-3 mt-2" :class="result.ok ? 'bg-green-950/20 border-green-500' : 'bg-red-950/20 border-red-500'">
    <div class="flex items-center gap-2 mb-2 flex-wrap">
      <span class="text-base">{{ getToolIcon(toolName) }}</span>
      <span class="font-semibold">{{ toolName }}</span>
      <Badge
        variant="outline"
        class="text-xs"
        :class="result.ok ? 'border-green-500 text-green-500' : 'border-red-500 text-red-500'"
      >
        {{ result.ok ? '✓ Success' : '✕ Error' }}
      </Badge>
      <span v-if="result.meta?.duration_ms" class="ml-auto text-xs text-muted-foreground">
        {{ formatDuration(result.meta.duration_ms) }}
      </span>
    </div>
    
    <div v-if="result.ok && result.data !== undefined" class="bg-muted/30 rounded p-2 relative">
      <pre class="font-mono text-xs text-muted-foreground whitespace-pre-wrap break-all max-h-[300px] overflow-y-auto">{{ formatResult(result.data) }}</pre>
      <Badge v-if="result.meta?.truncated" variant="outline" class="absolute bottom-2 right-2 text-xs">
        Truncated
      </Badge>
    </div>
    
    <div v-if="!result.ok && result.error" class="bg-muted/30 rounded p-2">
      <div class="font-mono text-xs text-red-400 mb-1">{{ result.error.code }}</div>
      <div class="text-sm">{{ result.error.message }}</div>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue'
import Badge from '@/components/ui/Badge.vue'

export default defineComponent({
  components: { Badge }
})
</script>
