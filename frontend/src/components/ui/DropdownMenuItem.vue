<script setup lang="ts">
import { cn } from '@/lib/utils'
interface Props {}
defineProps<Props>()
</script>
<template>
  <div
    :class="cn('relative flex cursor-default select-none items-center rounded-sm px-2 py-1.5 text-sm outline-none transition-colors focus:bg-accent focus:text-accent-foreground data-[disabled]:pointer-events-none data-[disabled]:opacity-50 hover:bg-accent hover:text-accent-foreground', $attrs.class as string)"
    role="menuitem"
    tabindex="-1"
  >
    <slot />
  </div>
</template>
