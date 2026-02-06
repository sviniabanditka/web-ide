<script setup lang="ts">
import { ref, computed } from 'vue'

interface Props {
  thinking: string
  visible?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  visible: true
})

const emit = defineEmits<{
  toggle: [collapsed: boolean]
}>()

const isCollapsed = ref(!props.visible)

const displayText = computed(() => {
  if (!props.thinking) return ''
  const lines = props.thinking.split('\n')
  if (lines.length <= 3 || !isCollapsed.value) return props.thinking
  return lines.slice(0, 3).join('\n') + '\n...'
})

function toggle() {
  isCollapsed.value = !isCollapsed.value
  emit('toggle', isCollapsed.value)
}
</script>

<template>
  <div class="thinking-block" :class="{ collapsed: isCollapsed }">
    <div class="thinking-header" @click="toggle">
      <span class="thinking-icon">💭</span>
      <span class="thinking-label">Thinking</span>
      <span class="thinking-toggle">{{ isCollapsed ? '▶' : '▼' }}</span>
    </div>
    <div v-if="!isCollapsed" class="thinking-content">
      <pre>{{ displayText }}</pre>
    </div>
  </div>
</template>

<style scoped>
.thinking-block {
  background: #2a2a35;
  border: 1px solid #3a3a4a;
  border-radius: 8px;
  margin: 8px 0;
  overflow: hidden;
}

.thinking-block.collapsed {
  border-color: #3a3a4a;
}

.thinking-header {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  background: #252530;
  cursor: pointer;
  user-select: none;
}

.thinking-header:hover {
  background: #2d2d38;
}

.thinking-icon {
  font-size: 14px;
}

.thinking-label {
  font-size: 12px;
  font-weight: 500;
  color: #888;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.thinking-toggle {
  font-size: 10px;
  color: #666;
  margin-left: auto;
}

.thinking-content {
  padding: 12px;
  background: #1e1e25;
  border-top: 1px solid #3a3a4a;
}

.thinking-content pre {
  margin: 0;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
  font-size: 13px;
  color: #999;
  white-space: pre-wrap;
  word-break: break-word;
  line-height: 1.6;
}
</style>
