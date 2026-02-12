<script setup lang="ts">
import { cn } from '@/lib/utils'
interface Props {}
defineProps<Props>()
</script>
<template>
  <div :class="cn('w-full overflow-auto', $attrs.class as string)">
    <table :class="cn('w-full caption-bottom text-sm', $attrs.class as string)">
      <slot />
    </table>
  </div>
</template>
