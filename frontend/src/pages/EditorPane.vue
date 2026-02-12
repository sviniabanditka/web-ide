<template>
  <div class="h-full flex">
    <aside class="w-[280px] bg-card border-r flex flex-col">
      <div v-if="editorStore.openFiles.length > 0" class="flex flex-col border-b bg-secondary/50">
        <div class="px-3 py-2 text-xs font-medium text-muted-foreground uppercase">Open Editors</div>
        <div class="flex flex-col">
          <div
            v-for="file in editorStore.openFiles"
            :key="file.path"
            class="group flex items-center gap-2 px-3 py-1.5 text-sm cursor-pointer border-l-2 border-transparent hover:bg-accent/50"
            :class="{ 'bg-accent border-l-primary': editorStore.activeFile?.path === file.path }"
            @click="editorStore.setActiveFile(file.path)"
          >
            <span class="truncate flex-1">{{ file.name }}</span>
            <button
              class="text-muted-foreground hover:text-foreground opacity-0 group-hover:opacity-100 p-0.5 rounded"
              @click.stop="closeFile(file.path)"
            >
              <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M18 6 6 18"/>
                <path d="m6 6 12 12"/>
              </svg>
            </button>
          </div>
        </div>
      </div>

      <div class="px-4 py-3 text-xs font-medium text-muted-foreground uppercase border-b">Files</div>
      <ScrollArea class="flex-1">
        <FileTreeNode
          v-if="editorStore.fileTree"
          :node="editorStore.fileTree"
          :project-id="project.id"
          @select="handleSelect"
        />
        <div v-else class="p-4 text-center text-muted-foreground text-sm">
          Loading...
        </div>
      </ScrollArea>
    </aside>

    <div class="flex-1 flex flex-col overflow-hidden">
      <div class="flex-1 overflow-hidden">
        <MonacoEditor
          v-if="editorStore.activeFile"
          :modelValue="editorStore.activeFile.content"
          :language="editorStore.activeFile.language || 'plaintext'"
          :path="editorStore.activeFile.path"
          theme="vs-dark"
          @update:modelValue="handleContentChange"
          @save="handleSave"
          class="w-full h-full"
        />
        <div v-else class="h-full flex items-center justify-center bg-background">
          <div class="text-center text-muted-foreground">
            <h3 class="text-lg mb-2">No file open</h3>
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
import ScrollArea from '@/components/ui/ScrollArea.vue'

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
  contextMenuX.value = event.clientX
  contextMenuY.value = event.clientY
  contextMenuPath.value = node.path
  contextMenuType.value = node.type
  contextMenuVisible.value = true
}

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

  const normalizedPath = path.startsWith('/') ? path : '/' + path
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
  if (modalMode.value === 'create') {
    void await editorStore.createFile(props.project.id, modalPath.value, name, type === 'folder')
  } else {
    await editorStore.renameFile(props.project.id, modalPath.value, name)
  }
  modalVisible.value = false
}
</script>
