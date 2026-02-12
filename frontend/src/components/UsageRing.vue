<template>
  <div v-if="usageData" class="relative" @mouseenter="showTooltip = true" @mouseleave="showTooltip = false">
    <svg class="w-8 h-8" viewBox="0 0 32 32">
      <circle class="stroke-muted" cx="16" cy="16" r="14" fill="none" stroke-width="4"/>
      <circle
        class="stroke-primary transition-all duration-500 ease-out"
        cx="16" cy="16" r="14"
        fill="none"
        :stroke="progressColor"
        stroke-width="4"
        :stroke-dasharray="dashArray"
        :stroke-dashoffset="dashOffset"
        stroke-linecap="round"
        transform="rotate(-90 16 16)"
      />
    </svg>
    <div
      v-show="showTooltip"
      class="absolute bottom-full left-1/2 transform -translate-x-1/2 mb-2 bg-popover border rounded-md shadow-lg p-3 min-w-[140px] z-50"
    >
      <div class="text-sm font-semibold mb-2 border-b pb-2">{{ usageData.model_name || 'Model Usage' }}</div>
      <div class="flex justify-between text-xs mb-1">
        <span class="text-muted-foreground">Used:</span>
        <span>{{ formatCount(usageData.usage_count) }}</span>
      </div>
      <div class="flex justify-between text-xs mb-1">
        <span class="text-muted-foreground">Remaining:</span>
        <span>{{ formatCount(usageData.remaining_credits) }}</span>
      </div>
      <div class="flex justify-between text-xs mb-1">
        <span class="text-muted-foreground">Total:</span>
        <span>{{ formatCount(usageData.total_count) }}</span>
      </div>
      <div class="flex justify-between text-xs mb-2">
        <span class="text-muted-foreground">Progress:</span>
        <span>{{ usageData.percent_used.toFixed(1) }}%</span>
      </div>
      <div v-if="timeRange" class="text-xs text-muted-foreground pt-2 border-t">
        {{ timeRange }} (UTC)
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, onMounted } from 'vue'
import { useAIStore } from '../stores/ai'

const aiStore = useAIStore()
const showTooltip = ref(false)

const usageData = computed(() => aiStore.usage)

const circumference = 2 * Math.PI * 14
const dashArray = circumference

const dashOffset = computed(() => {
  if (!usageData.value) return dashArray
  const percentUsed = usageData.value.percent_used / 100
  return dashArray * (1 - percentUsed)
})

const progressColor = computed(() => {
  if (!usageData.value) return 'hsl(var(--primary))'
  const percent = usageData.value.percent_used
  if (percent >= 90) return 'hsl(var(--destructive))'
  if (percent >= 70) return '#f57c00'
  return 'hsl(var(--primary))'
})

const timeRange = computed(() => {
  if (!usageData.value || !usageData.value.start_time || !usageData.value.end_time) return null
  const start = new Date(usageData.value.start_time).toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit', timeZone: 'UTC' })
  const end = new Date(usageData.value.end_time).toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit', timeZone: 'UTC' })
  return `${start}-${end}`
})

function formatCount(value: number): string {
  if (value >= 1000000) {
    return (value / 1000000).toFixed(2) + 'M'
  }
  if (value >= 1000) {
    return (value / 1000).toFixed(1) + 'K'
  }
  return value.toFixed(0)
}

onMounted(() => {
  aiStore.fetchUsage()
})
</script>
