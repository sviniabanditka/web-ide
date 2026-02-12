<template>
  <div class="h-full flex flex-col">
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
