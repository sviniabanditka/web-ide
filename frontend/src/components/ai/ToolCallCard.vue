<script setup lang="ts">
interface ToolCall {
  id: string
  name: string
  arguments: Record<string, unknown>
  status: 'pending' | 'approved' | 'rejected' | 'executing' | 'completed' | 'error'
}

const props = defineProps<{
  tool: ToolCall
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
</script>

<template>
  <div class="rounded-lg border bg-card p-3 mt-2">
    <div class="flex items-center gap-2 mb-2">
      <span class="text-lg">{{ getToolIcon(tool.name) }}</span>
      <span class="font-semibold">{{ tool.name }}</span>
      <Badge v-if="tool.status !== 'pending'" variant="outline" class="ml-auto text-xs">
        {{ tool.status }}
      </Badge>
    </div>
    <div class="bg-muted/50 rounded p-2 overflow-x-auto">
      <pre class="font-mono text-xs text-muted-foreground">{{ formatArguments(tool.arguments) }}</pre>
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
