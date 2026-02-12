<script setup lang="ts">
import { cn } from '@/lib/utils'
interface Props {}
defineProps<Props>()
</script>
<template>
  <div :class="cn('relative overflow-auto', $attrs.class as string)">
    <slot />
  </div>
</template>
