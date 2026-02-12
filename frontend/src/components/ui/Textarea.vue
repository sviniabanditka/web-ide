<script setup lang="ts">
import { computed } from 'vue'
import { cn } from '@/lib/utils'
interface Props {
  modelValue?: string | number
}
const props = withDefaults(defineProps<Props>(), {
  modelValue: ''
})
const modelValue = computed({
  get: () => props.modelValue,
  set: (value) => emits('update:modelValue', value)
})
const emits = defineEmits<{
  'update:modelValue': [value: string | number]
}>()
</script>
<template>
  <textarea
    :class="cn(
      'flex min-h-[80px] w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50',
      $attrs.class as string
    )"
    :value="modelValue"
    @input="modelValue = ($event.target as HTMLTextAreaElement).value"
  />
</template>
