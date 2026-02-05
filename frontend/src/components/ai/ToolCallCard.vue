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
  <div class="tool-call-card">
    <div class="tool-header">
      <span class="tool-icon">{{ getToolIcon(tool.name) }}</span>
      <span class="tool-name">{{ tool.name }}</span>
      <span v-if="tool.status !== 'pending'" class="tool-status" :class="tool.status">
        {{ tool.status }}
      </span>
    </div>
    <div class="tool-arguments">
      <pre>{{ formatArguments(tool.arguments) }}</pre>
    </div>
  </div>
</template>

<style scoped>
.tool-call-card {
  background: #1a1a2e;
  border: 1px solid #3a3a5e;
  border-radius: 8px;
  padding: 12px;
  margin: 8px 0;
}

.tool-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.tool-icon {
  font-size: 18px;
}

.tool-name {
  font-weight: 600;
  color: #e0e0e0;
}

.tool-status {
  font-size: 11px;
  padding: 2px 8px;
  border-radius: 12px;
  margin-left: auto;
  text-transform: uppercase;
}

.tool-status.pending {
  background: #f59e0b20;
  color: #f59e0b;
}

.tool-status.approved {
  background: #10b98120;
  color: #10b981;
}

.tool-status.rejected {
  background: #ef444420;
  color: #ef4444;
}

.tool-status.executing {
  background: #3b82f620;
  color: #3b82f6;
}

.tool-status.completed {
  background: #10b98120;
  color: #10b981;
}

.tool-status.error {
  background: #ef444420;
  color: #ef4444;
}

.tool-arguments {
  background: #0f0f1a;
  border-radius: 4px;
  padding: 8px;
  overflow-x: auto;
}

.tool-arguments pre {
  margin: 0;
  font-family: 'Monaco', 'Menlo', monospace;
  font-size: 12px;
  color: #a0a0c0;
}
</style>
