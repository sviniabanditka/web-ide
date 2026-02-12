<script setup lang="ts">
import { cn } from '@/lib/utils'
interface Props {}
defineProps<Props>()
</script>
<template>
  <div :class="cn('rounded-lg border bg-card text-card-foreground shadow-sm', $attrs.class as string)">
    <slot />
  </div>
</template>
