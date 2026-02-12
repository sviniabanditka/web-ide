<script setup lang="ts">
import { cn } from '@/lib/utils'
interface Props {}
defineProps<Props>()
</script>
<template>
  <p :class="cn('text-sm text-muted-foreground', $attrs.class as string)">
    <slot />
  </p>
</template>
