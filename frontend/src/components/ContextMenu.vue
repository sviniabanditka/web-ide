<template>
  <Teleport to="body">
    <div
      v-if="visible"
      class="context-menu"
      :style="{ top: `${y}px`, left: `${x}px` }"
      @click.stop
    >
      <div class="menu-item" @click="handleAction('open')">
        <span class="icon">📝</span> Open
      </div>
      <div v-if="isDirectory" class="menu-item" @click="handleAction('create')">
        <span class="icon">➕</span> New File
      </div>
      <div v-if="isDirectory" class="menu-item" @click="handleAction('createFolder')">
        <span class="icon">📁</span> New Folder
      </div>
      <div class="menu-divider"></div>
      <div class="menu-item" @click="handleAction('rename')">
        <span class="icon">✏️</span> Rename
      </div>
      <div class="menu-item danger" @click="handleAction('delete')">
        <span class="icon">🗑️</span> Delete
      </div>
    </div>
    <div v-if="visible" class="context-menu-backdrop" @click="close"></div>
  </Teleport>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted } from 'vue'

interface Props {
  visible: boolean
  x: number
  y: number
  path: string
  type: 'file' | 'directory'
}

const props = defineProps<Props>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'action', action: string, path: string, type: 'file' | 'directory'): void
}>()

const isDirectory = computed(() => props.type === 'directory')

function handleAction(action: string) {
  emit('action', action, props.path, props.type)
  close()
}

function close() {
  emit('close')
}

function handleClickOutside(e: MouseEvent) {
  const target = e.target as HTMLElement
  if (!target.closest('.context-menu')) {
    close()
  }
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
})
</script>

<style scoped>
.context-menu {
  position: fixed;
  background: #252526;
  border: 1px solid #3c3c3c;
  border-radius: 6px;
  padding: 6px 0;
  min-width: 160px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.4);
  z-index: 10000;
}

.context-menu-backdrop {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 999;
}

.menu-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  font-size: 13px;
  color: #ccc;
  cursor: pointer;
  transition: background 0.15s;
}

.menu-item:hover {
  background: #37373d;
  color: #fff;
}

.menu-item.danger:hover {
  background: #8b0000;
}

.menu-item .icon {
  font-size: 14px;
}

.menu-divider {
  height: 1px;
  background: #3c3c3c;
  margin: 6px 0;
}
</style>
