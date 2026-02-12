<script setup lang="ts">
import { ref } from 'vue'

const props = defineProps<{
  text: string
  position?: 'top' | 'bottom' | 'left' | 'right'
}>()

const show = ref(false)
</script>

<template>
  <div
    class="tooltip-wrapper"
    @mouseenter="show = true"
    @mouseleave="show = false"
  >
    <slot />
    <Transition name="tooltip-fade">
      <div
        v-if="show"
        class="tooltip-content"
        :class="[`tooltip-${position || 'top'}`]"
      >
        {{ text }}
      </div>
    </Transition>
  </div>
</template>

<style scoped>
.tooltip-wrapper {
  position: relative;
  display: inline-flex;
}

.tooltip-content {
  position: absolute;
  z-index: 50;
  padding: 6px 10px;
  font-size: 12px;
  background: hsl(var(--popover));
  color: hsl(var(--popover-foreground));
  border: 1px solid hsl(var(--border));
  border-radius: 6px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  white-space: nowrap;
  pointer-events: none;
}

.tooltip-top {
  bottom: 100%;
  left: 50%;
  transform: translateX(-50%);
  margin-bottom: 8px;
}

.tooltip-bottom {
  top: 100%;
  left: 50%;
  transform: translateX(-50%);
  margin-top: 8px;
}

.tooltip-left {
  right: 100%;
  top: 50%;
  transform: translateY(-50%);
  margin-right: 8px;
}

.tooltip-right {
  left: 100%;
  top: 50%;
  transform: translateY(-50%);
  margin-left: 8px;
}

.tooltip-fade-enter-active,
.tooltip-fade-leave-active {
  transition: opacity 0.15s ease, transform 0.15s ease;
}

.tooltip-fade-enter-from,
.tooltip-fade-leave-to {
  opacity: 0;
}

.tooltip-top.tooltip-fade-enter-from,
.tooltip-top.tooltip-fade-leave-to {
  transform: translateX(-50%) translateY(4px);
}

.tooltip-bottom.tooltip-fade-enter-from,
.tooltip-bottom.tooltip-fade-leave-to {
  transform: translateX(-50%) translateY(-4px);
}

.tooltip-left.tooltip-fade-enter-from,
.tooltip-left.tooltip-fade-leave-to {
  transform: translateY(-50%) translateX(4px);
}

.tooltip-right.tooltip-fade-enter-from,
.tooltip-right.tooltip-fade-leave-to {
  transform: translateY(-50%) translateX(-4px);
}
</style>
