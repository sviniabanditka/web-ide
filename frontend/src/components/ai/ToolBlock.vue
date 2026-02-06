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
  <div class="tool-block" :class="getStatusClass()">
    <div class="tool-header">
      <span class="tool-icon">{{ getToolIcon(tool.name) }}</span>
      <span class="tool-name">{{ tool.name }}</span>
      <span class="tool-status-badge">{{ getStatusText() }}</span>
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
  border-radius: 8px;
  padding: 12px;
  margin: 8px 0;
  transition: all 0.2s;
}

.tool-block.executing {
  background: #1a2a3a;
  border: 1px solid #3b82f6;
}

.tool-block.completed {
  background: #1a2e1a;
  border: 1px solid #10b981;
}

.tool-block.error {
  background: #2e1a1a;
  border: 1px solid #ef4444;
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

.tool-status-badge {
  font-size: 11px;
  padding: 2px 8px;
  border-radius: 12px;
  margin-left: auto;
  text-transform: uppercase;
}

.tool-block.executing .tool-status-badge {
  background: #3b82f620;
  color: #3b82f6;
}

.tool-block.completed .tool-status-badge {
  background: #10b98120;
  color: #10b981;
}

.tool-block.error .tool-status-badge {
  background: #ef444420;
  color: #ef4444;
}

.tool-arguments {
  background: rgba(0, 0, 0, 0.3);
  border-radius: 4px;
  padding: 8px;
  overflow-x: auto;
}

.tool-block.executing .tool-arguments {
  background: rgba(59, 130, 246, 0.1);
}

.tool-block.completed .tool-arguments {
  background: rgba(16, 185, 129, 0.1);
}

.tool-block.error .tool-arguments {
  background: rgba(239, 68, 68, 0.1);
}

.tool-arguments pre {
  margin: 0;
  font-family: 'Monaco', 'Menlo', monospace;
  font-size: 12px;
  color: #a0a0c0;
}

.tool-result {
  margin-top: 8px;
  background: rgba(0, 0, 0, 0.2);
  border-radius: 4px;
  padding: 8px;
  overflow-x: auto;
}

.tool-block.completed .tool-result {
  background: rgba(16, 185, 129, 0.1);
}

.tool-block.error .tool-result {
  background: rgba(239, 68, 68, 0.1);
}

.tool-result pre {
  margin: 0;
  font-family: 'Monaco', 'Menlo', monospace;
  font-size: 12px;
  color: #a0c0a0;
}

.tool-block.error .tool-result pre {
  color: #c0a0a0;
}
</style>

