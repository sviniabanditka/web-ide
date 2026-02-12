<script setup lang="ts">
import { cn } from '@/lib/utils'
interface Props {}
defineProps<Props>()
</script>
<template>
  <thead :class="cn('[&_tr]:border-b', $attrs.class as string)">
    <slot />
  </thead>
</template>
