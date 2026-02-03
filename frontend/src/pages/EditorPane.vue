<template>
  <div class="editor-pane">
    <div class="editor-sidebar">
      <div class="sidebar-header">Files</div>
      <div class="file-tree">
        <FileTreeNode
          v-if="editorStore.fileTree"
          :node="editorStore.fileTree"
          :project-id="project.id"
          @select="handleSelect"
        />
        <div v-else class="empty-state">Loading...</div>
      </div>
    </div>

    <div class="editor-main">
      <div class="editor-tabs" v-if="editorStore.openFiles.length > 0">
        <div
          v-for="file in editorStore.openFiles"
          :key="file.path"
          class="editor-tab"
          :class="{ active: editorStore.activeFile?.path === file.path }"
          @click="editorStore.setActiveFile(file.path)"
        >
          <span class="tab-name">{{ file.name }}</span>
          <button class="close-tab" @click.stop="closeFile(file.path)">×</button>
        </div>
      </div>

        <div class="editor-content">
          <template v-if="editorStore.activeFile">
            <MonacoEditor
              :modelValue="editorStore.activeFile.content"
              :language="editorStore.activeFile.language || 'plaintext'"
              :path="editorStore.activeFile.path"
              theme="vs-dark"
              @update:modelValue="handleContentChange"
              @save="handleSave"
              class="monaco-editor"
            />
          </template>
        <div v-else class="no-file">
          <div class="no-file-content">
            <h3>No file open</h3>
            <p>Select a file from the tree to edit</p>
          </div>
        </div>
      </div>
    </div>

    <ContextMenu
      :visible="contextMenuVisible"
      :x="contextMenuX"
      :y="contextMenuY"
      :path="contextMenuPath"
      :type="contextMenuType"
      @close="contextMenuVisible = false"
      @action="handleContextAction"
    />

    <FileOperationModal
      :visible="modalVisible"
      :mode="modalMode"
      :type="modalType"
      :current-name="modalCurrentName"
      :path="modalPath"
      @close="modalVisible = false"
      @submit="handleModalSubmit"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, provide } from 'vue'
import MonacoEditor from '../components/MonacoEditor.vue'
import { useEditorStore } from '../stores/editor'
import FileTreeNode from '../components/FileTreeNode.vue'
import ContextMenu from '../components/ContextMenu.vue'
import FileOperationModal from '../components/FileOperationModal.vue'

interface Project {
  id: string
  name: string
  root_path: string
}

const props = defineProps<{
  project: Project
}>()

const editorStore = useEditorStore()

const contextMenuVisible = ref(false)
const contextMenuX = ref(0)
const contextMenuY = ref(0)
const contextMenuPath = ref('')
const contextMenuType = ref<'file' | 'directory'>('file')

const modalVisible = ref(false)
const modalMode = ref<'create' | 'rename'>('create')
const modalType = ref<'file' | 'folder'>('file')
const modalPath = ref('')
const modalCurrentName = ref('')

function handleContextMenu(event: MouseEvent, node: any) {
  console.log('Context menu triggered:', { nodePath: node.path, nodeType: node.type, nodeName: node.name })
  contextMenuX.value = event.clientX
  contextMenuY.value = event.clientY
  contextMenuPath.value = node.path
  contextMenuType.value = node.type
  contextMenuVisible.value = true
}

// Provide context menu handler to all nested FileTreeNode components
provide('onContextMenu', handleContextMenu)

async function handleSelect(path: string) {
  await editorStore.openFile(props.project.id, path)
}

async function handleContentChange(content: string) {
  if (editorStore.activeFile) {
    editorStore.activeFile.content = content
  }
}

async function handleSave() {
  if (editorStore.activeFile) {
    try {
      await editorStore.saveFile(
        props.project.id,
        editorStore.activeFile.path,
        editorStore.activeFile.content,
        editorStore.activeFile.etag
      )
    } catch (e: any) {
      console.error('Save failed:', e.message)
      alert('Save failed: ' + e.message)
    }
  }
}

function closeFile(path: string) {
  editorStore.closeFile(path)
}

async function handleContextAction(action: string, path: string, nodeType: 'file' | 'directory') {
  contextMenuVisible.value = false

  // Нормализуем путь - убедимся что он полный (с ведущим /)
  const normalizedPath = path.startsWith('/') ? path : '/' + path
  console.log('handleContextAction:', { action, originalPath: path, normalizedPath, nodeType })

  const type: 'file' | 'folder' = nodeType === 'directory' ? 'folder' : 'file'

  switch (action) {
    case 'open':
      if (nodeType === 'directory') {
        editorStore.toggleExpandedDir(normalizedPath)
      } else {
        await handleSelect(normalizedPath)
      }
      break
    case 'create':
      modalMode.value = 'create'
      modalType.value = 'file'
      modalPath.value = normalizedPath
      modalCurrentName.value = ''
      modalVisible.value = true
      break
    case 'createFolder':
      modalMode.value = 'create'
      modalType.value = 'folder'
      modalPath.value = normalizedPath
      modalCurrentName.value = ''
      modalVisible.value = true
      break
    case 'rename':
      modalMode.value = 'rename'
      modalType.value = type
      modalPath.value = normalizedPath
      modalCurrentName.value = path.split('/').pop() || ''
      modalVisible.value = true
      break
    case 'delete':
      const isFolder = nodeType === 'directory'
      if (confirm(`Delete ${isFolder ? 'folder' : 'file'} "${path.split('/').pop()}"?`)) {
        await editorStore.deleteFile(props.project.id, normalizedPath, isFolder)
      }
      break
  }
}

async function handleModalSubmit(name: string, type: 'file' | 'folder') {
  console.log('handleModalSubmit:', { mode: modalMode.value, type, name, path: modalPath.value })
  if (modalMode.value === 'create') {
    const result = await editorStore.createFile(props.project.id, modalPath.value, name, type === 'folder')
    console.log('createFile result:', result, 'error:', editorStore.error)
  } else {
    await editorStore.renameFile(props.project.id, modalPath.value, name)
  }
  modalVisible.value = false
}
</script>

<style scoped>
.editor-pane {
  height: 100%;
  display: flex;
  background: #1e1e1e;
}

.editor-sidebar {
  width: 250px;
  background: #252526;
  border-right: 1px solid #3c3c3c;
  display: flex;
  flex-direction: column;
}

.sidebar-header {
  padding: 12px 16px;
  font-size: 12px;
  font-weight: 500;
  color: #888;
  text-transform: uppercase;
  border-bottom: 1px solid #3c3c3c;
}

.file-tree {
  flex: 1;
  overflow-y: auto;
  padding: 8px 0;
}

.empty-state {
  padding: 16px;
  color: #666;
  font-size: 13px;
  text-align: center;
}

.editor-main {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.editor-tabs {
  display: flex;
  background: #252526;
  border-bottom: 1px solid #3c3c3c;
  overflow-x: auto;
}

.editor-tab {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  background: #2d2d30;
  border-right: 1px solid #1e1e1e;
  cursor: pointer;
  font-size: 13px;
  color: #ccc;
  white-space: nowrap;
}

.editor-tab:hover {
  background: #3c3c3c;
}

.editor-tab.active {
  background: #1e1e1e;
  color: #fff;
  border-top: 2px solid #0e639c;
}

.tab-name {
  max-width: 150px;
  overflow: hidden;
  text-overflow: ellipsis;
}

.close-tab {
  background: none;
  border: none;
  color: #888;
  font-size: 14px;
  padding: 0;
  line-height: 1;
}

.close-tab:hover {
  color: #fff;
}

.editor-content {
  flex: 1;
  overflow: hidden;
}

.monaco-editor {
  width: 100%;
  height: 100%;
}

.no-file {
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #1e1e1e;
}

.no-file-content {
  text-align: center;
  color: #666;
}

.no-file-content h3 {
  font-size: 18px;
  margin-bottom: 8px;
  color: #888;
}
</style>
