<script setup lang="ts">
import { computed } from 'vue'
import { cn } from '@/lib/utils'
interface Props {
  modelValue?: number
}
const props = withDefaults(defineProps<Props>(), {
  modelValue: 0
})
const modelValue = computed({
  get: () => props.modelValue,
  set: (value) => emits('update:modelValue', value)
})
const emits = defineEmits<{
  'update:modelValue': [value: number]
}>()
</script>
<template>
  <div :class="cn('relative h-4 w-full overflow-hidden rounded-full bg-secondary', $attrs.class as string)">
    <div
      class="h-full w-full flex-1 bg-primary transition-all"
      :style="{ transform: `translateX(-${100 - (modelValue || 0)}%)` }"
    />
  </div>
</template>
