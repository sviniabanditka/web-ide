<script setup lang="ts">
import { ref, computed } from 'vue'

interface Props {
  modelValue: string
  label: string
}

const props = defineProps<Props>()
const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const inputRef = ref<HTMLInputElement | null>(null)

const colorValue = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value)
})

function handleInput(e: Event) {
  const target = e.target as HTMLInputElement
  colorValue.value = target.value
}

function triggerPicker() {
  inputRef.value?.click()
}
</script>

<template>
  <div class="flex items-center gap-2">
    <button
      type="button"
      class="w-8 h-8 rounded border cursor-pointer shrink-0"
      :style="{ backgroundColor: colorValue }"
      @click="triggerPicker"
    />
    <input
      ref="inputRef"
      type="color"
      :value="colorValue"
      class="sr-only"
      @input="handleInput"
    />
    <div class="flex-1 min-w-0">
      <label class="text-sm text-muted-foreground">{{ label }}</label>
      <input
        type="text"
        :value="colorValue"
        class="w-full text-sm bg-transparent border-0 p-0 focus:ring-0 font-mono"
        @input="handleInput"
      />
    </div>
  </div>
</template>
