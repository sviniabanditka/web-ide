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
  <div class="tool-result-card" :class="{ error: !result.ok }">
    <div class="result-header">
      <span class="tool-icon">{{ getToolIcon(toolName) }}</span>
      <span class="tool-name">{{ toolName }}</span>
      <span class="result-badge" :class="result.ok ? 'success' : 'error'">
        {{ result.ok ? '✓ Success' : '✕ Error' }}
      </span>
      <span v-if="result.meta?.duration_ms" class="duration">
        {{ formatDuration(result.meta.duration_ms) }}
      </span>
    </div>
    
    <div v-if="result.ok && result.data !== undefined" class="result-data">
      <pre>{{ formatResult(result.data) }}</pre>
      <div v-if="result.meta?.truncated" class="truncated-badge">
        Truncated
      </div>
    </div>
    
    <div v-if="!result.ok && result.error" class="error-details">
      <div class="error-code">{{ result.error.code }}</div>
      <div class="error-message">{{ result.error.message }}</div>
    </div>
  </div>
</template>

<style scoped>
.tool-result-card {
  background: #1a2e1a;
  border: 1px solid #10b981;
  border-radius: 8px;
  padding: 12px;
  margin: 8px 0;
}

.tool-result-card.error {
  background: #2e1a1a;
  border-color: #ef4444;
}

.result-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
  flex-wrap: wrap;
}

.tool-icon {
  font-size: 16px;
}

.tool-name {
  font-weight: 600;
  color: #e0e0e0;
}

.result-badge {
  font-size: 11px;
  padding: 2px 8px;
  border-radius: 12px;
  font-weight: 600;
}

.result-badge.success {
  background: #10b98120;
  color: #10b981;
}

.result-badge.error {
  background: #ef444420;
  color: #ef4444;
}

.duration {
  margin-left: auto;
  font-size: 11px;
  color: #666;
}

.result-data {
  background: #0f1a0f;
  border-radius: 4px;
  padding: 8px;
  position: relative;
}

.result-data pre {
  margin: 0;
  font-family: 'Monaco', 'Menlo', monospace;
  font-size: 12px;
  color: #a0c0a0;
  white-space: pre-wrap;
  word-break: break-all;
  max-height: 300px;
  overflow-y: auto;
}

.truncated-badge {
  position: absolute;
  bottom: 4px;
  right: 4px;
  font-size: 10px;
  background: #f59e0b;
  color: #000;
  padding: 2px 6px;
  border-radius: 4px;
}

.error-details {
  background: #1a0f0f;
  border-radius: 4px;
  padding: 8px;
}

.error-code {
  font-family: 'Monaco', 'Menlo', monospace;
  font-size: 11px;
  color: #ef4444;
  margin-bottom: 4px;
}

.error-message {
  font-size: 13px;
  color: #e0e0e0;
}
</style>
