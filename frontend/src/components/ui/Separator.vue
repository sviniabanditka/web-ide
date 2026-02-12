<script setup lang="ts">
import { cn } from '@/lib/utils'
interface Props {
  orientation?: 'horizontal' | 'vertical'
}
defineProps<Props>()
</script>
<template>
  <hr :class="cn('shrink-0 bg-border', orientation === 'vertical' ? 'h-full w-[1px]' : 'h-[1px] w-full', $attrs.class as string)" />
</template>
