<template>
  <Teleport to="body">
    <div
      v-if="visible"
      class="fixed bg-popover border rounded-md shadow-lg min-w-[160px] py-1 z-[10000]"
      :style="{ top: `${y}px`, left: `${x}px` }"
      @click.stop
    >
      <div
        class="flex items-center gap-2 px-4 py-1.5 text-sm cursor-pointer text-popover-foreground hover:bg-accent hover:text-accent-foreground"
        @click="handleAction('open')"
      >
        <span>📝</span> Open
      </div>
      <div
        v-if="isDirectory"
        class="flex items-center gap-2 px-4 py-1.5 text-sm cursor-pointer text-popover-foreground hover:bg-accent hover:text-accent-foreground"
        @click="handleAction('create')"
      >
        <span>➕</span> New File
      </div>
      <div
        v-if="isDirectory"
        class="flex items-center gap-2 px-4 py-1.5 text-sm cursor-pointer text-popover-foreground hover:bg-accent hover:text-accent-foreground"
        @click="handleAction('createFolder')"
      >
        <span>📁</span> New Folder
      </div>
      <div class="h-px bg-border my-1"></div>
      <div
        class="flex items-center gap-2 px-4 py-1.5 text-sm cursor-pointer text-popover-foreground hover:bg-accent hover:text-accent-foreground"
        @click="handleAction('rename')"
      >
        <span>✏️</span> Rename
      </div>
      <div
        class="flex items-center gap-2 px-4 py-1.5 text-sm cursor-pointer text-destructive hover:bg-destructive/10"
        @click="handleAction('delete')"
      >
        <span>🗑️</span> Delete
      </div>
    </div>
    <div v-if="visible" class="fixed inset-0 z-[9999]" @click="close"></div>
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
  if (!target.closest('.fixed')) {
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
