<script setup lang="ts">
import { computed } from 'vue'
import { cn } from '@/lib/utils'
interface Props {
  size?: 'sm' | 'md' | 'lg'
}
const props = withDefaults(defineProps<Props>(), {
  size: 'md'
})
const sizeClasses = computed(() => {
  switch (props.size) {
    case 'sm':
      return 'h-8 w-8'
    case 'lg':
      return 'h-12 w-12'
    default:
      return 'h-10 w-10'
  }
})
</script>
<template>
  <div :class="cn('relative flex shrink-0 overflow-hidden rounded-full', sizeClasses, $attrs.class as string)">
    <slot />
  </div>
</template>
