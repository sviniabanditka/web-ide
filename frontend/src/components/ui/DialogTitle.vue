<script setup lang="ts">
import { cn } from '@/lib/utils'
interface Props {}
defineProps<Props>()
</script>
<template>
  <h2 :class="cn('text-lg font-semibold leading-none tracking-tight', $attrs.class as string)">
    <slot />
  </h2>
</template>
