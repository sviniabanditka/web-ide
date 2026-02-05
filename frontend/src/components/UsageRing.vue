<template>
  <div class="usage-ring-container" v-if="usageData" @mouseenter="showTooltip = true" @mouseleave="showTooltip = false">
    <svg class="usage-ring" width="32" height="32" viewBox="0 0 32 32">
      <circle class="ring-bg" cx="16" cy="16" r="14" fill="none" stroke="#3c3c3c" stroke-width="4"/>
      <circle class="ring-progress" cx="16" cy="16" r="14" fill="none" :stroke="progressColor" stroke-width="4"
        :stroke-dasharray="dashArray" :stroke-dashoffset="dashOffset" stroke-linecap="round" transform="rotate(-90 16 16)"/>
    </svg>
    <div class="tooltip" v-show="showTooltip">
      <div class="tooltip-title">{{ usageData.model_name || 'Model Usage' }}</div>
      <div class="tooltip-row">
        <span>Used:</span>
        <span>{{ formatCount(usageData.usage_count) }}</span>
      </div>
      <div class="tooltip-row">
        <span>Remaining:</span>
        <span>{{ formatCount(usageData.remaining_credits) }}</span>
      </div>
      <div class="tooltip-row">
        <span>Total:</span>
        <span>{{ formatCount(usageData.total_count) }}</span>
      </div>
      <div class="tooltip-row">
        <span>Progress:</span>
        <span>{{ usageData.percent_used.toFixed(1) }}%</span>
      </div>
      <div class="tooltip-time" v-if="timeRange">
        {{ timeRange }}
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
  if (!usageData.value) return '#0e639c'
  const percent = usageData.value.percent_used
  if (percent >= 90) return '#c62828'
  if (percent >= 70) return '#f57c00'
  return '#2e7d32'
})

const timeRange = computed(() => {
  if (!usageData.value || !usageData.value.start_time || !usageData.value.end_time) return null
  const start = new Date(usageData.value.start_time).toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit', timeZone: 'UTC' })
  const end = new Date(usageData.value.end_time).toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit', timeZone: 'UTC' })
  return `${start}-${end} (UTC)`
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

<style scoped>
.usage-ring-container {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
}

.usage-ring {
  display: block;
}

.ring-progress {
  transition: stroke-dashoffset 0.5s ease, stroke 0.3s ease;
}

.tooltip {
  position: absolute;
  bottom: 100%;
  left: 50%;
  transform: translateX(-50%);
  background: #2d2d30;
  border: 1px solid #3c3c3c;
  border-radius: 6px;
  padding: 8px 12px;
  min-width: 140px;
  margin-bottom: 8px;
  z-index: 1000;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
}

.tooltip::after {
  content: '';
  position: absolute;
  top: 100%;
  left: 50%;
  transform: translateX(-50%);
  border: 6px solid transparent;
  border-top-color: #2d2d30;
}

.tooltip-title {
  font-size: 12px;
  font-weight: 600;
  color: #fff;
  margin-bottom: 6px;
  border-bottom: 1px solid #3c3c3c;
  padding-bottom: 4px;
}

.tooltip-row {
  display: flex;
  justify-content: space-between;
  font-size: 11px;
  color: #ccc;
  margin: 2px 0;
}

.tooltip-row span:last-child {
  font-weight: 500;
}

.tooltip-time {
  font-size: 10px;
  color: #888;
  margin-top: 6px;
  padding-top: 6px;
  border-top: 1px solid #3c3c3c;
}
</style>
