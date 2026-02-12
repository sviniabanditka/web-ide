<template>
  <Teleport to="body">
    <div v-if="visible" class="fixed inset-0 z-50 bg-black/50 flex items-center justify-center" @click.self="handleClose">
      <div class="bg-popover rounded-lg border shadow-lg w-full max-w-md">
        <div class="flex items-center justify-between px-6 py-4 border-b">
          <h3 class="font-semibold">{{ title }}</h3>
          <button class="text-muted-foreground hover:text-foreground text-xl" @click="handleClose">×</button>
        </div>
        <div class="p-6">
          <div v-if="error" class="bg-destructive/10 border border-destructive text-destructive px-3 py-2 rounded-md text-sm mb-4">
            {{ error }}
          </div>
          <div v-if="showTypeSelect" class="mb-4">
            <Label class="mb-2">Type</Label>
            <Select v-model="fileType">
              <option value="file">File</option>
              <option value="folder">Folder</option>
            </Select>
          </div>
          <div>
            <Label class="mb-2">{{ label }}</Label>
            <Input
              v-model="name"
              ref="nameInput"
              type="text"
              :placeholder="placeholder"
              @keydown.enter="handleSubmit"
              @keydown.escape="handleClose"
            />
          </div>
        </div>
        <div class="flex justify-end gap-3 px-6 py-4 border-t">
          <Button variant="secondary" @click="handleClose">Cancel</Button>
          <Button @click="handleSubmit" :disabled="!isValid || loading">
            {{ loading ? '...' : action }}
          </Button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, computed, watch, nextTick } from 'vue'
import Button from '@/components/ui/Button.vue'
import Input from '@/components/ui/Input.vue'
import Label from '@/components/ui/Label.vue'
import Select from '@/components/ui/Select.vue'

interface Props {
  visible: boolean
  mode: 'create' | 'rename'
  type: 'file' | 'folder'
  currentName?: string
  path?: string
}

const props = withDefaults(defineProps<Props>(), {
  currentName: '',
  path: ''
})

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'submit', name: string, type: 'file' | 'folder'): void
}>()

const name = ref('')
const fileType = ref<'file' | 'folder'>('file')
const error = ref<string | null>(null)
const loading = ref(false)
const nameInput = ref<HTMLInputElement | null>(null)

const title = computed(() => {
  if (props.mode === 'create') {
    return fileType.value === 'folder' ? 'New Folder' : 'New File'
  }
  return 'Rename'
})

const action = computed(() => {
  if (props.mode === 'create') {
    return 'Create'
  }
  return 'Rename'
})

const label = computed(() => {
  return props.mode === 'create' ? 'Name' : 'New name'
})

const placeholder = computed(() => {
  if (props.mode === 'create') {
    return fileType.value === 'folder' ? 'folder-name' : 'file-name.txt'
  }
  return props.currentName
})

const showTypeSelect = computed(() => {
  return props.mode === 'create'
})

const isValid = computed(() => {
  const trimmed = name.value.trim()
  if (!trimmed) return false
  if (trimmed.includes('/') || trimmed.includes('\\')) return false
  if (trimmed.startsWith('.') && props.mode === 'create') return false
  if (props.mode === 'rename') {
    return trimmed !== props.currentName
  }
  return true
})

function handleClose() {
  emit('close')
}

async function handleSubmit() {
  if (!isValid.value || loading.value) return

  loading.value = true
  error.value = null

  try {
    emit('submit', name.value.trim(), fileType.value)
  } catch (e: any) {
    error.value = e.message || 'Operation failed'
    loading.value = false
  }
}

watch(() => props.visible, (visible) => {
  if (visible) {
    name.value = props.mode === 'rename' ? props.currentName : ''
    fileType.value = props.type
    error.value = null
    loading.value = false
    nextTick(() => {
      nameInput.value?.focus()
    })
  }
})
</script>
