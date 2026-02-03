import { defineStore } from 'pinia'
import { ref } from 'vue'
import { api } from '../api'

interface Project {
  id: string
  name: string
  root_path: string
  created_at: string
  last_opened_at: string
}

export const useProjectsStore = defineStore('projects', () => {
  const projects = ref<Project[]>([])
  const currentProject = ref<Project | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  async function fetchProjects() {
    loading.value = true
    error.value = null
    try {
      const response = await api.get('/api/v1/projects')
      projects.value = response.data
    } catch (e: any) {
      error.value = e.response?.data?.error || 'Failed to fetch projects'
    } finally {
      loading.value = false
    }
  }

  async function fetchProject(id: string) {
    loading.value = true
    error.value = null
    try {
      const response = await api.get(`/api/v1/projects/${id}`)
      currentProject.value = response.data
      return response.data
    } catch (e: any) {
      error.value = e.response?.data?.error || 'Failed to fetch project'
      return null
    } finally {
      loading.value = false
    }
  }

  return { projects, currentProject, loading, error, fetchProjects, fetchProject }
})
