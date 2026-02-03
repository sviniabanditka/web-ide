<template>
  <div class="projects-page">
    <header class="header">
      <h1>Projects</h1>
      <button @click="handleLogout" class="logout-btn">Logout</button>
    </header>
    
    <div class="projects-content">
      <div v-if="projectsStore.loading" class="loading">Loading...</div>
      
      <div v-else-if="projectsStore.error" class="error">{{ projectsStore.error }}</div>
      
      <div v-else-if="projectsStore.projects.length === 0" class="empty">
        No projects found. Add projects to {{ projectsDir }} directory.
      </div>
      
      <div v-else class="projects-grid">
        <div 
          v-for="project in projectsStore.projects" 
          :key="project.id" 
          class="project-card"
          @click="openProject(project.id)"
        >
          <div class="project-icon">📁</div>
          <div class="project-info">
            <h3>{{ project.name }}</h3>
            <p>{{ project.root_path }}</p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { useProjectsStore } from '../stores/projects'

const router = useRouter()
const authStore = useAuthStore()
const projectsStore = useProjectsStore()

const projectsDir = '/projects'

async function handleLogout() {
  await authStore.logout()
  router.push('/login')
}

function openProject(id: string) {
  router.push(`/projects/${id}`)
}

onMounted(async () => {
  await authStore.checkAuth()
  await projectsStore.fetchProjects()
})
</script>

<style scoped>
.projects-page {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 24px;
  background: #252526;
  border-bottom: 1px solid #3c3c3c;
}

h1 {
  font-size: 18px;
  font-weight: 500;
}

.logout-btn {
  padding: 8px 16px;
  background: transparent;
  border: 1px solid #3c3c3c;
  border-radius: 4px;
  color: #ccc;
  font-size: 13px;
}

.logout-btn:hover {
  background: #3c3c3c;
}

.projects-content {
  flex: 1;
  padding: 24px;
  overflow: auto;
}

.loading, .error, .empty {
  text-align: center;
  padding: 40px;
  color: #888;
}

.error {
  color: #f44336;
}

.projects-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 16px;
}

.project-card {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 20px;
  background: #252526;
  border: 1px solid #3c3c3c;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;
}

.project-card:hover {
  border-color: #0e639c;
  background: #2d2d30;
}

.project-icon {
  font-size: 32px;
}

.project-info h3 {
  font-size: 16px;
  font-weight: 500;
  margin-bottom: 4px;
}

.project-info p {
  font-size: 12px;
  color: #888;
}
</style>
