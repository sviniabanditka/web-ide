<template>
  <div class="terminal-workspace">
    <TerminalList :project="project" />
  </div>
</template>

<script setup lang="ts">
import TerminalList from '../components/TerminalList.vue'

interface Project {
  id: string
  name: string
  root_path: string
}

defineProps<{
  project: Project
}>()
</script>

<style scoped>
.terminal-workspace {
  height: 100%;
  display: flex;
  flex-direction: column;
}
</style>
