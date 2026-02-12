<script setup lang="ts">
import { computed } from 'vue'
import Label from '@/components/ui/Label.vue'
import { useSettingsStore } from '@/stores/settings'

const settingsStore = useSettingsStore()

const themes = computed(() => settingsStore.editorThemes)
const currentThemeId = computed(() => settingsStore.settings?.editor_theme_id || 'vs-dark')

function selectTheme(themeId: string) {
  settingsStore.saveSettings({ editor_theme_id: themeId })
}
</script>

<template>
  <div class="space-y-6">
    <div>
      <Label class="text-base">Editor Color Scheme</Label>
      <p class="text-sm text-muted-foreground mt-1 mb-4">
        Choose a color theme for the Monaco Editor
      </p>
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
        <button
          v-for="theme in themes"
          :key="theme.id"
          @click="selectTheme(theme.id)"
          :class="[
            'relative p-3 rounded-lg border-2 transition-all text-left',
            currentThemeId === theme.id
              ? 'border-primary ring-2 ring-primary/20'
              : 'border-border hover:border-primary/50'
          ]"
        >
          <div
            class="h-12 rounded flex items-center justify-center text-xs font-medium"
            :style="{
              backgroundColor: theme.id === 'vs' ? '#ffffff' : theme.id === 'hc-black' ? '#000000' : '#1e1e1e',
              color: theme.id === 'vs' ? '#333333' : theme.id === 'hc-black' ? '#ffffff' : '#d4d4d4'
            }"
          >
            {{ theme.name }}
          </div>
          <div class="font-medium text-sm mt-2">{{ theme.name }}</div>
          <div
            v-if="currentThemeId === theme.id"
            class="absolute top-2 right-2 w-2 h-2 rounded-full bg-primary"
          />
        </button>
      </div>
    </div>

    <div class="pt-4 border-t">
      <Label class="text-base">Current Editor Theme</Label>
      <p class="text-sm text-muted-foreground mt-1">
        Selected: <span class="font-medium">{{ themes.find(t => t.id === currentThemeId)?.name }}</span>
      </p>
    </div>
  </div>
</template>
