import { createApp } from 'vue'
import { createPinia } from 'pinia'
import { createRouter, createWebHistory } from 'vue-router'
import App from './App.vue'

import LoginPage from './pages/LoginPage.vue'
import ProjectsPage from './pages/ProjectsPage.vue'
import WorkspacePage from './pages/WorkspacePage.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/login', component: LoginPage },
    { path: '/projects', component: ProjectsPage },
    { path: '/projects/:id', component: WorkspacePage },
    { path: '/', redirect: '/login' }
  ]
})

const app = createApp(App)
app.use(createPinia())
app.use(router)
app.mount('#app')
