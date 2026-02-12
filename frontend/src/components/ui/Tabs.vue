<script setup lang="ts">
import { cn } from '@/lib/utils'
interface Props {
  modelValue?: string
}
const props = defineProps<Props>()
const emits = defineEmits<{
  'update:modelValue': [value: string]
}>()
function activateTab(value: string) {
  emits('update:modelValue', value)
}
</script>
<template>
  <div :class="cn('space-y-2', $attrs.class as string)">
    <div class="flex space-x-2 border-b" role="tablist">
      <slot name="trigger" :activate="activateTab" :active="modelValue" />
    </div>
    <div class="mt-2">
      <slot />
    </div>
  </div>
</template>
