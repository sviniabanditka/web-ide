import { defineStore } from 'pinia'
import { ref } from 'vue'
import { api } from '../api'

interface FileNode {
  id: string
  name: string
  path: string
  type: 'file' | 'directory'
  children?: FileNode[]
}

interface OpenFile {
  path: string
  name: string
  content: string
  etag: string
  language?: string
}

interface WorkspaceState {
  open_files: string[]
  expanded_dirs: string[]
  active_file: string | null
  active_tab: string
  open_terminals: string[]
}

export const useEditorStore = defineStore('editor', () => {
  const fileTree = ref<FileNode | null>(null)
  const openFiles = ref<OpenFile[]>([])
  const activeFile = ref<OpenFile | null>(null)
  const expandedDirs = ref<Set<string>>(new Set())
  const loading = ref(false)
  const error = ref<string | null>(null)
  let currentProjectId: string | null = null

  async function loadWorkspaceState(projectId: string) {
    currentProjectId = projectId
    try {
      const response = await api.get(`/api/v1/projects/${projectId}/workspace`)
      const state: WorkspaceState = response.data

      state.expanded_dirs.forEach(path => expandedDirs.value.add(path))

      // Open files sequentially to preserve order
      for (const path of state.open_files) {
        await openFile(projectId, path)
      }

      // Set active file after all files are opened
      if (state.active_file) {
        const file = openFiles.value.find(f => f.path === state.active_file)
        if (file) {
          activeFile.value = file
        }
      }

      return state.active_tab || 'terminal'
    } catch (e: any) {
      console.warn('Failed to load workspace state:', e.message)
      return 'terminal'
    }
  }

  async function saveWorkspaceState() {
    if (!currentProjectId) return
    try {
      await api.put(`/api/v1/projects/${currentProjectId}/workspace`, {
        open_files: openFiles.value.map(f => f.path),
        expanded_dirs: Array.from(expandedDirs.value),
        active_file: activeFile.value?.path || null,
        active_tab: 'editor',
        open_terminals: []
      })
    } catch (e: any) {
      console.warn('Failed to save workspace state:', e.message)
    }
  }

  async function fetchFileTree(projectId: string, path = '/') {
    loading.value = true
    error.value = null
    try {
      const response = await api.get(`/api/v1/projects/${projectId}/fs/tree`, {
        params: { path }
      })
      fileTree.value = response.data
    } catch (e: any) {
      error.value = e.response?.data?.error || 'Failed to fetch file tree'
    } finally {
      loading.value = false
    }
  }

  async function createFile(projectId: string, parentPath: string, name: string, isFolder: boolean) {
    if (!name || !name.trim()) {
      error.value = 'Name is required'
      return false
    }
    
    const trimmedName = name.trim()
    loading.value = true
    error.value = null
    
    // Нормализуем путь - убедимся что он начинается с /
    const normalizedParentPath = parentPath.startsWith('/') ? parentPath : '/' + parentPath
    const fullPath = normalizedParentPath === '/' ? `/${trimmedName}` : `${normalizedParentPath}/${trimmedName}`
    
    console.log('Creating file:', { parentPath, normalizedParentPath, name: trimmedName, path: fullPath, isFolder })
    try {
      if (isFolder) {
        await api.post(`/api/v1/projects/${projectId}/fs/mkdir`, { path: fullPath })
      } else {
        await api.put(`/api/v1/projects/${projectId}/fs/file`, { content: '' }, {
          params: { path: fullPath }
        })
      }
      await fetchFileTree(projectId, '/')
      return true
    } catch (e: any) {
      error.value = e.response?.data?.error || 'Failed to create'
      console.error('Create file error:', error.value)
      return false
    } finally {
      loading.value = false
    }
  }

  async function renameFile(projectId: string, oldPath: string, newName: string) {
    loading.value = true
    error.value = null
    try {
      // Нормализуем путь - убедимся что он начинается с /
      const normalizedOldPath = oldPath.startsWith('/') ? oldPath : '/' + oldPath
      const lastSlash = normalizedOldPath.lastIndexOf('/')
      const parentPath = lastSlash > 0 ? normalizedOldPath.substring(0, lastSlash) : '/'
      const newPath = parentPath === '/' ? `/${newName}` : `${parentPath}/${newName}`

      console.log('Renaming:', { oldPath: normalizedOldPath, parentPath, newPath, newName })

      await api.post(`/api/v1/projects/${projectId}/fs/rename`, {
        from: normalizedOldPath,
        to: newPath
      })

      const openFileIndex = openFiles.value.findIndex(f => f.path === normalizedOldPath)
      if (openFileIndex !== -1) {
        openFiles.value[openFileIndex].path = newPath
        openFiles.value[openFileIndex].name = newName
      }
      if (activeFile.value?.path === normalizedOldPath) {
        activeFile.value.path = newPath
        activeFile.value.name = newName
      }

      await fetchFileTree(projectId, '/')
      return true
    } catch (e: any) {
      error.value = e.response?.data?.error || 'Failed to rename'
      return false
    } finally {
      loading.value = false
    }
  }

  async function deleteFile(projectId: string, path: string, isFolder: boolean) {
    loading.value = true
    error.value = null
    try {
      // Нормализуем путь
      const normalizedPath = path.startsWith('/') ? path : '/' + path
      const lastSlash = normalizedPath.lastIndexOf('/')
      const parentPath = lastSlash > 0 ? normalizedPath.substring(0, lastSlash) : '/'

      console.log('Deleting:', { path: normalizedPath, parentPath, isFolder })

      await api.delete(`/api/v1/projects/${projectId}/fs/remove`, {
        data: { path: normalizedPath, recursive: isFolder }
      })

      const openFileIndex = openFiles.value.findIndex(f => f.path === normalizedPath)
      if (openFileIndex !== -1) {
        openFiles.value.splice(openFileIndex, 1)
      }
      if (activeFile.value?.path === normalizedPath) {
        activeFile.value = openFiles.value[openFiles.value.length - 1] || null
      }

      await fetchFileTree(projectId, '/')
      return true
    } catch (e: any) {
      error.value = e.response?.data?.error || 'Failed to delete'
      return false
    } finally {
      loading.value = false
    }
  }

  async function openFile(projectId: string, path: string) {
    const normalizedPath = path.replace(/\/+/g, '/').replace(/^\//, '/')
    const existing = openFiles.value.find(f => f.path === normalizedPath)
    if (existing) {
      activeFile.value = existing
      return existing
    }

    loading.value = true
    error.value = null
    try {
      const response = await api.get(`/api/v1/projects/${projectId}/fs/file`, {
        params: { path: normalizedPath }
      })

      const file: OpenFile = {
        path: normalizedPath,
        name: normalizedPath.split('/').pop() || normalizedPath,
        content: response.data.content || '',
        etag: response.data.etag || '',
        language: getLanguage(normalizedPath)
      }

      openFiles.value.push(file)
      activeFile.value = file
      loading.value = false
      saveWorkspaceState()
      return file
    } catch (e: any) {
      error.value = e.response?.data?.error || 'Failed to open file'
      loading.value = false
      return null
    }
  }

  async function saveFile(projectId: string, path: string, content: string, etag: string) {
    try {
      await api.put(`/api/v1/projects/${projectId}/fs/file`, {
        content,
        expectedEtag: etag
      }, {
        params: { path }
      })

      const file = openFiles.value.find(f => f.path === path)
      if (file) {
        file.content = content
        file.etag = calculateEtag(content)
      }
    } catch (e: any) {
      if (e.response?.status === 409) {
        throw new Error('File was modified by someone else')
      }
      throw e
    }
  }

  function closeFile(path: string) {
    const index = openFiles.value.findIndex(f => f.path === path)
    if (index !== -1) {
      openFiles.value.splice(index, 1)
    }
    if (activeFile.value?.path === path) {
      activeFile.value = openFiles.value[openFiles.value.length - 1] || null
    }
    saveWorkspaceState()
  }

  function setActiveFile(path: string) {
    const normalizedPath = path.replace(/\/+/g, '/').replace(/^\//, '/')
    const file = openFiles.value.find(f => f.path === normalizedPath)
    if (file) {
      activeFile.value = file
      saveWorkspaceState()
    }
  }

  function toggleExpandedDir(path: string) {
    const normalizedPath = path.replace(/\/+/g, '/').replace(/^\//, '/')
    if (expandedDirs.value.has(normalizedPath)) {
      expandedDirs.value.delete(normalizedPath)
    } else {
      expandedDirs.value.add(normalizedPath)
    }
    saveWorkspaceState()
  }

  function isDirExpanded(path: string): boolean {
    const normalizedPath = path.replace(/\/+/g, '/').replace(/^\//, '/')
    return expandedDirs.value.has(normalizedPath)
  }

  function getLanguage(filename: string): string {
    const ext = filename.split('.').pop()?.toLowerCase()
    const languages: Record<string, string> = {
      'go': 'go',
      'ts': 'typescript',
      'tsx': 'typescript',
      'js': 'javascript',
      'jsx': 'javascript',
      'vue': 'vue',
      'py': 'python',
      'rs': 'rust',
      'java': 'java',
      'c': 'c',
      'cpp': 'cpp',
      'h': 'c',
      'hpp': 'cpp',
      'json': 'json',
      'yaml': 'yaml',
      'yml': 'yaml',
      'md': 'markdown',
      'html': 'html',
      'css': 'css',
      'scss': 'scss',
      'sql': 'sql',
      'sh': 'shell',
      'bash': 'shell',
      'xml': 'xml',
      'toml': 'toml',
      'dockerfile': 'dockerfile'
    }
    return languages[ext || ''] || 'plaintext'
  }

  function calculateEtag(content: string): string {
    let hash = 0
    for (let i = 0; i < content.length; i++) {
      hash = ((hash << 5) - hash) + content.charCodeAt(i)
      hash = hash & hash
    }
    return `"${hash}-${Date.now()}"`
  }

  return {
    fileTree,
    openFiles,
    activeFile,
    expandedDirs,
    loading,
    error,
    loadWorkspaceState,
    saveWorkspaceState,
    fetchFileTree,
    createFile,
    renameFile,
    deleteFile,
    openFile,
    saveFile,
    closeFile,
    setActiveFile,
    toggleExpandedDir,
    isDirExpanded
  }
})
