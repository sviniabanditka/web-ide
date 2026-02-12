<script setup lang="ts">
import { cn } from '@/lib/utils'
interface Props {
  open?: boolean
}
const props = defineProps<Props>()
const emits = defineEmits<{
  'update:open': [value: boolean]
}>()
function toggle() {
  emits('update:open', !props.open)
}
function close() {
  emits('update:open', false)
}
</script>
<template>
  <div :class="cn('relative inline-block text-left', $attrs.class as string)">
    <slot name="trigger" :toggle="toggle" :open="open" />
    <div
      v-if="open"
      class="absolute right-0 z-50 mt-2 w-56 origin-top-right rounded-md border bg-popover p-1 text-popover-foreground shadow-lg ring-1 ring-black ring-opacity-5 focus:outline-none"
      role="menu"
    >
      <slot :close="close" />
    </div>
  </div>
</template>
