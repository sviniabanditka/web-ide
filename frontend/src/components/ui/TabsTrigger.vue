<script setup lang="ts">
import { cn } from '@/lib/utils'
interface Props {
  value: string
  active?: boolean
}
const props = defineProps<Props>()
</script>
<template>
  <button
    :class="cn(
      'inline-flex items-center justify-center whitespace-nowrap px-3 py-1.5 text-sm font-medium ring-offset-background transition-all focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50',
      props.active
        ? 'border-b-2 border-primary text-primary'
        : 'text-muted-foreground hover:text-foreground',
      $attrs.class as string
    )"
    :aria-selected="props.active"
    role="tab"
  >
    <slot />
  </button>
</template>
