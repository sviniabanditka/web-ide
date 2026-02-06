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
</script>

<template>
  <div class="tool-block" :class="getStatusClass()">
    <div class="tool-header">
      <span class="tool-icon">{{ getToolIcon(tool.name) }}</span>
      <span class="tool-name">{{ tool.name }}</span>
      <span class="tool-status" :class="getStatusClass()">
        {{ result ? (result.ok ? 'completed' : 'error') : tool.status }}
      </span>
    </div>
    <div class="tool-arguments">
      <pre>{{ formatArguments(tool.arguments) }}</pre>
    </div>
    <div v-if="result" class="tool-result" :class="{ error: !result.ok }">
      <pre>{{ result.ok ? formatResult(result.result) : result.error?.message }}</pre>
    </div>
  </div>
</template>

<style scoped>
.tool-block {
  background: #1a1a2e;
  border: 1px solid #3a3a5e;
  border-left: 3px solid #3b82f6;
  border-radius: 8px;
  padding: 12px;
  margin: 8px 0;
  transition: border-color 0.2s;
}

.tool-block.completed {
  border-left-color: #10b981;
}

.tool-block.error {
  border-left-color: #ef4444;
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

.tool-result {
  margin-top: 8px;
  background: #0f1a1a;
  border-radius: 4px;
  padding: 8px;
  overflow-x: auto;
}

.tool-result.error {
  background: #1a0f0f;
}

.tool-result pre {
  margin: 0;
  font-family: 'Monaco', 'Menlo', monospace;
  font-size: 12px;
  color: #a0c0a0;
}

.tool-result.error pre {
  color: #c0a0a0;
}
</style>
