<script setup lang="ts">
import { cn } from '@/lib/utils'
interface Props {}
defineProps<Props>()
</script>
<template>
  <div :class="cn('-mx-1 my-1 h-px bg-muted', $attrs.class as string)" role="separator" />
</template>
