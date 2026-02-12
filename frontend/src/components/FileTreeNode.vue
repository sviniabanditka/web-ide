<template>
  <div class="select-none">
    <div
      class="flex items-center gap-1.5 px-3 py-1 cursor-pointer text-sm text-muted-foreground hover:text-foreground hover:bg-accent"
      :class="{ 'bg-accent text-foreground': isSelected }"
      @click="handleClick"
      @contextmenu.prevent="handleRightClick"
    >
      <span class="text-base">{{ node.type === 'directory' ? (isOpen ? '📂' : '📁') : '📄' }}</span>
      <span class="truncate">{{ node.name }}</span>
    </div>

    <div v-if="node.type === 'directory' && isOpen" class="pl-3">
      <FileTreeNode
        v-for="child in node.children"
        :key="child.id"
        :node="child"
        :project-id="projectId"
        @select="emit('select', $event)"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, inject } from 'vue'
import { useEditorStore } from '../stores/editor'

interface FileNode {
  id: string
  name: string
  path: string
  type: 'file' | 'directory'
  children?: FileNode[]
}

interface ContextMenuHandler {
  (event: MouseEvent, node: FileNode): void
}

const props = defineProps<{
  node: FileNode
  projectId: string
}>()

const emit = defineEmits<{
  (e: 'select', path: string): void
}>()

const editorStore = useEditorStore()

const onContextMenu = inject<ContextMenuHandler | null>('onContextMenu', null)

const isOpen = computed(() => editorStore.isDirExpanded(props.node.path))
const isSelected = computed(() => editorStore.activeFile?.path === props.node.path)

function handleClick() {
  if (props.node.type === 'directory') {
    editorStore.toggleExpandedDir(props.node.path)
  } else {
    emit('select', props.node.path)
  }
}

function handleRightClick(event: MouseEvent) {
  if (onContextMenu) {
    onContextMenu(event, props.node)
  }
}
</script>
