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
  <div class="tool-approval-card">
    <div class="approval-header">
      <span class="warning-icon">⚠️</span>
      <span class="title">Approval Required</span>
    </div>
    
    <div class="tool-info">
      <div class="tool-name-row">
        <span class="tool-icon">{{ getToolIcon(tool.name) }}</span>
        <span class="tool-name">{{ tool.name }}</span>
      </div>
      
      <div v-if="tool.summary" class="tool-summary">
        {{ tool.summary }}
      </div>
      
      <div class="tool-arguments">
        <div class="arguments-label">Arguments:</div>
        <pre>{{ formatArguments(tool.arguments) }}</pre>
      </div>
    </div>
    
    <div class="approval-actions">
      <button class="approve-btn" @click="onApprove">
        ✓ Approve
      </button>
      <button class="reject-btn" @click="onReject">
        ✕ Reject
      </button>
    </div>
  </div>
</template>

<style scoped>
.tool-approval-card {
  background: linear-gradient(135deg, #2a1a1a 0%, #1a1a2e 100%);
  border: 2px solid #f59e0b;
  border-radius: 12px;
  padding: 16px;
  margin: 12px 0;
}

.approval-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
  padding-bottom: 12px;
  border-bottom: 1px solid #f59e0b40;
}

.warning-icon {
  font-size: 20px;
}

.title {
  font-weight: 600;
  color: #f59e0b;
  font-size: 14px;
  text-transform: uppercase;
  letter-spacing: 1px;
}

.tool-info {
  margin-bottom: 16px;
}

.tool-name-row {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.tool-name {
  font-weight: 600;
  color: #e0e0e0;
  font-size: 16px;
}

.tool-summary {
  color: #a0a0c0;
  font-size: 13px;
  margin-bottom: 12px;
  font-style: italic;
}

.tool-arguments {
  background: #0f0f1a;
  border-radius: 6px;
  padding: 12px;
}

.arguments-label {
  font-size: 11px;
  color: #666;
  text-transform: uppercase;
  margin-bottom: 8px;
}

.tool-arguments pre {
  margin: 0;
  font-family: 'Monaco', 'Menlo', monospace;
  font-size: 12px;
  color: #a0a0c0;
  white-space: pre-wrap;
  word-break: break-all;
}

.approval-actions {
  display: flex;
  gap: 12px;
}

.approve-btn,
.reject-btn {
  flex: 1;
  padding: 12px 16px;
  border: none;
  border-radius: 8px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.approve-btn {
  background: #10b981;
  color: white;
}

.approve-btn:hover {
  background: #059669;
  transform: translateY(-1px);
}

.reject-btn {
  background: #3a3a5e;
  color: #e0e0e0;
}

.reject-btn:hover {
  background: #4a4a6e;
  transform: translateY(-1px);
}
</style>
