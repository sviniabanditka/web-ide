<script setup lang="ts">
import { computed } from 'vue'
import { cn } from '@/lib/utils'
interface Props {
  modelValue?: boolean
  id?: string
}
const props = withDefaults(defineProps<Props>(), {
  modelValue: false
})
const modelValue = computed({
  get: () => props.modelValue,
  set: (value) => emits('update:modelValue', value)
})
const emits = defineEmits<{
  'update:modelValue': [value: boolean]
}>()
</script>
<template>
  <input
    :id="id"
    type="checkbox"
    :checked="modelValue"
    :class="cn(
      'peer h-4 w-4 shrink-0 rounded-sm border border-primary ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50',
      $attrs.class as string
    )"
    @change="modelValue = ($event.target as HTMLInputElement).checked"
  />
</template>
