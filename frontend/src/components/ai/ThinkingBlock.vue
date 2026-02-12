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
  <div class="thinking-block rounded-lg overflow-hidden border bg-card mt-2 mb-2" :class="{ 'border-muted': isCollapsed }">
    <div class="flex items-center gap-2 px-3 py-2 bg-muted/30 cursor-pointer hover:bg-muted/50" @click="toggle">
      <span class="text-sm">💭</span>
      <span class="text-xs font-medium text-muted-foreground uppercase tracking-wide">Thinking</span>
      <span class="ml-auto text-xs text-muted-foreground">{{ isCollapsed ? '▶' : '▼' }}</span>
    </div>
    <div v-if="!isCollapsed" class="p-3 bg-muted/10 border-t">
      <pre class="font-sans text-sm text-muted-foreground whitespace-pre-wrap break-words leading-relaxed">{{ displayText }}</pre>
    </div>
  </div>
</template>
