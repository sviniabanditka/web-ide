import { defineStore } from 'pinia'
import { ref } from 'vue'
import { api } from '../api'

interface GitStatus {
  is_git_repo: boolean
  status: {
    changed: FileStatus[]
    staged: FileStatus[]
    untracked: string[]
  }
  branches: string[]
  current_branch?: string
}

interface FileStatus {
  path: string
  status: string
}

interface GitLogEntry {
  hash: string
  author: string
  email: string
  date: string
  subject: string
}

export const useGitStore = defineStore('git', () => {
  const status = ref<GitStatus | null>(null)
  const diff = ref<string>('')
  const log = ref<GitLogEntry[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)
  const currentBranch = ref<string>('')

  async function fetchStatus(projectId: string) {
    loading.value = true
    error.value = null
    try {
      const response = await api.get(`/api/v1/projects/${projectId}/git/status`)
      status.value = response.data
      currentBranch.value = extractCurrentBranch(response.data.branches)
    } catch (e: any) {
      error.value = e.response?.data?.error || 'Failed to fetch git status'
    } finally {
      loading.value = false
    }
  }

  async function fetchDiff(projectId: string, cached = false) {
    console.log('Fetching diff, cached:', cached)
    try {
      const url = `/api/v1/projects/${projectId}/git/diff?cached=${cached ? 1 : 0}`
      console.log('Fetching diff from:', url)
      const response = await api.get(url)
      console.log('Diff response:', response.data.substring(0, 200))
      diff.value = response.data
    } catch (e: any) {
      console.error('Failed to fetch diff:', e)
    }
  }

  async function stageFiles(projectId: string, paths: string[]) {
    try {
      await api.post(`/api/v1/projects/${projectId}/git/stage`, { paths })
      await fetchStatus(projectId)
    } catch (e: any) {
      error.value = e.response?.data?.error || 'Failed to stage files'
    }
  }

  async function unstageFiles(projectId: string, paths: string[]) {
    try {
      await api.post(`/api/v1/projects/${projectId}/git/unstage`, { paths })
      await fetchStatus(projectId)
    } catch (e: any) {
      error.value = e.response?.data?.error || 'Failed to unstage files'
    }
  }

  async function commit(projectId: string, message: string) {
    try {
      await api.post(`/api/v1/projects/${projectId}/git/commit`, { message })
      await fetchStatus(projectId)
      await fetchLog(projectId)
    } catch (e: any) {
      error.value = e.response?.data?.error || 'Failed to commit'
      throw e
    }
  }

  async function push(projectId: string, remote = '', branch = '') {
    try {
      await api.post(`/api/v1/projects/${projectId}/git/push`, { remote, branch })
    } catch (e: any) {
      error.value = e.response?.data?.error || 'Failed to push'
      throw e
    }
  }

  async function fetchLog(projectId: string, limit = 20) {
    try {
      const response = await api.get(`/api/v1/projects/${projectId}/git/log`, {
        params: { limit }
      })
      log.value = response.data.log || []
    } catch (e: any) {
      console.error('Failed to fetch log:', e)
    }
  }

  async function fetchBranches(projectId: string) {
    try {
      const response = await api.get(`/api/v1/projects/${projectId}/git/branches`)
      return response.data.branches || []
    } catch (e: any) {
      console.error('Failed to fetch branches:', e)
      return []
    }
  }

  function extractCurrentBranch(branches: string[]): string {
    const current = branches.find(b => !b.startsWith('remotes/') && !b.startsWith('HEAD'))
    return current || 'main'
  }

  function stageAll(projectId: string) {
    const allPaths = [
      ...(status.value?.status.changed.map(f => f.path) || []),
      ...(status.value?.status.untracked || [])
    ]
    if (allPaths.length > 0) {
      stageFiles(projectId, allPaths)
    }
  }

  function unstageAll(projectId: string) {
    const stagedPaths = status.value?.status.staged.map(f => f.path) || []
    if (stagedPaths.length > 0) {
      unstageFiles(projectId, stagedPaths)
    }
  }

  return {
    status,
    diff,
    log,
    loading,
    error,
    currentBranch,
    fetchStatus,
    fetchDiff,
    stageFiles,
    unstageFiles,
    commit,
    push,
    fetchLog,
    fetchBranches,
    stageAll,
    unstageAll
  }
})
