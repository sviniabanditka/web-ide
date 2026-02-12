<script setup lang="ts">
import { computed } from 'vue'
import { cn } from '@/lib/utils'
interface Props {
  modelValue?: string
}
const props = withDefaults(defineProps<Props>(), {
  modelValue: ''
})
const modelValue = computed({
  get: () => props.modelValue,
  set: (value) => emits('update:modelValue', value)
})
const emits = defineEmits<{
  'update:modelValue': [value: string]
}>()
</script>
<template>
  <select
    :class="cn(
      'flex h-10 w-full items-center justify-between rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50',
      $attrs.class as string
    )"
    :value="modelValue"
    @change="modelValue = ($event.target as HTMLSelectElement).value"
  >
    <slot />
  </select>
</template>
