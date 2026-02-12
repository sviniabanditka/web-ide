<script setup lang="ts">
interface ToolResult {
  id: string
  name: string
  ok: boolean
  result?: Record<string, unknown>
  error?: {
    code: string
    message: string
  }
}

interface ToolCall {
  id: string
  name: string
  arguments: Record<string, unknown>
  status: 'pending' | 'approved' | 'rejected' | 'executing' | 'completed' | 'error'
}

const props = defineProps<{
  tool: ToolCall
  result?: ToolResult
}>()

function formatArguments(args: Record<string, unknown>): string {
  try {
    return JSON.stringify(args, null, 2)
  } catch {
    return String(args)
  }
}

function formatResult(result: Record<string, unknown> | undefined): string {
  if (!result) return ''
  try {
    return JSON.stringify(result, null, 2)
  } catch {
    return String(result)
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

function getStatusClass(): string {
  if (props.result) {
    return props.result.ok ? 'completed' : 'error'
  }
  return props.tool.status
}

function getStatusText(): string {
  if (props.result) {
    return props.result.ok ? 'completed' : 'error'
  }
  return props.tool.status
}
</script>

<template>
  <div class="tool-block rounded-lg p-3 mr-auto max-w-[80%] mt-2 mb-2" :class="{
    'bg-blue-950/50 border-blue-500': tool.status === 'executing',
    'bg-green-950/50 border-green-500': getStatusClass() === 'completed',
    'bg-red-950/50 border-red-500': getStatusClass() === 'error'
  }">
    <div class="flex items-center gap-2 mb-2">
      <span class="text-lg">{{ getToolIcon(tool.name) }}</span>
      <span class="font-semibold">{{ tool.name }}</span>
      <Badge
        variant="outline"
        class="ml-auto text-xs"
        :class="{
          'border-blue-500 text-blue-500': tool.status === 'executing',
          'border-green-500 text-green-500': getStatusClass() === 'completed',
          'border-red-500 text-red-500': getStatusClass() === 'error'
        }"
      >
        {{ getStatusText() }}
      </Badge>
    </div>
    <div
      class="rounded p-2 overflow-x-auto"
      :class="{
        'bg-blue-950/30': tool.status === 'executing',
        'bg-green-950/30': getStatusClass() === 'completed',
        'bg-red-950/30': getStatusClass() === 'error'
      }"
    >
      <pre class="font-mono text-xs text-muted-foreground">{{ formatArguments(tool.arguments) }}</pre>
    </div>
    <div
      v-if="result"
      class="mt-2 rounded p-2 overflow-x-auto"
      :class="result.ok ? 'bg-green-950/20' : 'bg-red-950/20'"
    >
      <pre class="font-mono text-xs" :class="result.ok ? 'text-green-400' : 'text-red-400'">
{{ result.ok ? formatResult(result.result) : result.error?.message }}
      </pre>
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
