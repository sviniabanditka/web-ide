<template>
  <Teleport to="body">
    <div v-if="visible" class="modal-overlay" @click.self="handleClose">
      <div class="modal">
        <div class="modal-header">
          <h3>{{ title }}</h3>
          <button class="close-btn" @click="handleClose">×</button>
        </div>
        <div class="modal-body">
          <div v-if="error" class="error-message">{{ error }}</div>
          <div class="form-group">
            <label v-if="showTypeSelect">Type</label>
            <select v-if="showTypeSelect" v-model="fileType" class="input">
              <option value="file">File</option>
              <option value="folder">Folder</option>
            </select>
          </div>
          <div class="form-group">
            <label>{{ label }}</label>
            <input
              v-model="name"
              ref="nameInput"
              type="text"
              class="input"
              :placeholder="placeholder"
              @keydown.enter="handleSubmit"
              @keydown.escape="handleClose"
            />
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn secondary" @click="handleClose">Cancel</button>
          <button class="btn primary" @click="handleSubmit" :disabled="!isValid || loading">
            {{ loading ? '...' : action }}
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, computed, watch, nextTick } from 'vue'

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

<style scoped>
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.6);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1001;
}

.modal {
  background: #252526;
  border: 1px solid #3c3c3c;
  border-radius: 8px;
  width: 400px;
  max-width: 90vw;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.5);
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  border-bottom: 1px solid #3c3c3c;
}

.modal-header h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 500;
  color: #fff;
}

.close-btn {
  background: none;
  border: none;
  font-size: 24px;
  color: #888;
  cursor: pointer;
  padding: 0;
  line-height: 1;
}

.close-btn:hover {
  color: #fff;
}

.modal-body {
  padding: 20px;
}

.form-group {
  margin-bottom: 16px;
}

.form-group:last-child {
  margin-bottom: 0;
}

.form-group label {
  display: block;
  margin-bottom: 6px;
  font-size: 13px;
  color: #888;
}

.input {
  width: 100%;
  padding: 10px 12px;
  background: #1e1e1e;
  border: 1px solid #3c3c3c;
  border-radius: 4px;
  color: #fff;
  font-size: 14px;
}

.input:focus {
  outline: none;
  border-color: #0e639c;
}

select.input {
  cursor: pointer;
}

.error-message {
  background: rgba(244, 67, 54, 0.1);
  border: 1px solid #f44336;
  color: #f44336;
  padding: 10px 12px;
  border-radius: 4px;
  margin-bottom: 16px;
  font-size: 13px;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  padding: 16px 20px;
  border-top: 1px solid #3c3c3c;
}

.btn {
  padding: 8px 16px;
  border: none;
  border-radius: 4px;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.15s;
}

.btn.secondary {
  background: #3c3c3c;
  color: #ccc;
}

.btn.secondary:hover {
  background: #4c4c4c;
}

.btn.primary {
  background: #0e639c;
  color: #fff;
}

.btn.primary:hover:not(:disabled) {
  background: #1177bb;
}

.btn.primary:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
</style>
