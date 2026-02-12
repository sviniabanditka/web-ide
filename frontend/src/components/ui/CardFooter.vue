<script setup lang="ts">
import { cn } from '@/lib/utils'
interface Props {}
defineProps<Props>()
</script>
<template>
  <div :class="cn('flex items-center p-6 pt-0', $attrs.class as string)">
    <slot />
  </div>
</template>
