<template>
  <div class="file-tree-node">
    <div
      class="node-row"
      :class="{ active: isSelected }"
      @click="handleClick"
      @contextmenu.prevent="handleRightClick"
    >
      <span class="node-icon">{{ node.type === 'directory' ? (isOpen ? '📂' : '📁') : '📄' }}</span>
      <span class="node-name">{{ node.name }}</span>
    </div>

    <div v-if="node.type === 'directory' && isOpen" class="node-children">
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

<style scoped>
.file-tree-node {
  user-select: none;
}

.node-row {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 4px 12px 4px 24px;
  cursor: pointer;
  font-size: 13px;
  color: #ccc;
}

.node-row:hover {
  background: #2d2d30;
}

.node-row.active {
  background: #37373d;
  color: #fff;
}

.node-icon {
  font-size: 14px;
  flex-shrink: 0;
}

.node-name {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.node-children {
  padding-left: 12px;
}
</style>
