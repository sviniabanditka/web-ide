<script setup lang="ts">
import { cn } from '@/lib/utils'
interface Props {
  open?: boolean
}
const props = defineProps<Props>()
const emits = defineEmits<{
  'update:open': [value: boolean]
}>()
function close() {
  emits('update:open', false)
}
</script>
<template>
  <div :class="cn('relative', $attrs.class as string)">
    <slot name="trigger" :open="() => emits('update:open', true)" />
    <div v-if="open" class="fixed inset-0 z-50 bg-black/50 flex items-center justify-center" @click.self="close">
      <div class="bg-background rounded-lg shadow-lg max-w-lg w-full p-6" role="dialog">
        <slot :close="close" />
      </div>
    </div>
  </div>
</template>
