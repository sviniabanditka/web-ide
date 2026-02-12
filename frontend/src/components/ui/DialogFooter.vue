<script setup lang="ts">
import { cn } from '@/lib/utils'
interface Props {}
defineProps<Props>()
</script>
<template>
  <div :class="cn('flex flex-col-reverse sm:flex-row sm:justify-end sm:space-x-2', $attrs.class as string)">
    <slot />
  </div>
</template>
