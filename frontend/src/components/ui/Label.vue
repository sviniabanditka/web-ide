<script setup lang="ts">
import { cn } from '@/lib/utils'
interface Props {
  for?: string
}
defineProps<Props>()
</script>
<template>
  <label :class="cn('text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70', $attrs.class as string)" :for="for">
    <slot />
  </label>
</template>
