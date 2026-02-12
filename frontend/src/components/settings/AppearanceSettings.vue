<script setup lang="ts">
import { ref, computed } from 'vue'
import Button from '@/components/ui/Button.vue'
import Label from '@/components/ui/Label.vue'
import { useSettingsStore } from '@/stores/settings'

const settingsStore = useSettingsStore()

const customThemeName = ref('')
const isCreatingCustom = ref(false)

const themes = computed(() => settingsStore.uiThemes)
const currentThemeId = computed(() => settingsStore.settings?.ui_theme_id || 'dark-plus')

function selectTheme(themeId: string) {
  settingsStore.saveSettings({ ui_theme_id: themeId })
}

function startCreatingCustom() {
  isCreatingCustom.value = true
  customThemeName.value = ''
}

function cancelCreatingCustom() {
  isCreatingCustom.value = false
}

async function createCustomTheme() {
  if (!customThemeName.value.trim()) return
  isCreatingCustom.value = false
}
</script>

<template>
  <div class="space-y-6">
    <div>
      <Label class="text-base">Interface Theme</Label>
      <p class="text-sm text-muted-foreground mt-1 mb-4">
        Choose a color theme for the user interface
      </p>
      <div class="grid grid-cols-2 md:grid-cols-3 gap-3">
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
            class="h-16 rounded-md mb-2 flex gap-0.5 overflow-hidden"
            :style="{ backgroundColor: theme.id === 'dark-plus' ? '#1e1e1e' : theme.id === 'light-plus' ? '#ffffff' : '#272822' }"
          >
            <div
              v-for="i in 5"
              :key="i"
              class="flex-1 flex flex-col gap-0.5 p-1"
            >
              <div
                v-for="j in 3"
                :key="j"
                class="h-2 rounded-sm"
                :style="{
                  backgroundColor: j === 1 ? '#007acc' : j === 2 ? '#3c3c3c' : '#858585'
                }"
              />
            </div>
          </div>
          <div class="font-medium text-sm">{{ theme.name }}</div>
          <div
            v-if="currentThemeId === theme.id"
            class="absolute top-2 right-2 w-2 h-2 rounded-full bg-primary"
          />
        </button>
      </div>
    </div>

    <div class="pt-4 border-t">
      <Label class="text-base">Custom Themes</Label>
      <p class="text-sm text-muted-foreground mt-1 mb-4">
        Create and manage custom color themes
      </p>

      <div v-if="!isCreatingCustom">
        <Button variant="outline" @click="startCreatingCustom">
          Create Custom Theme
        </Button>
      </div>

      <div v-else class="p-4 rounded-lg border bg-card space-y-3">
        <div class="font-medium text-sm">New Custom Theme</div>
        <div>
          <Label for="themeName">Theme Name</Label>
          <input
            id="themeName"
            v-model="customThemeName"
            class="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm mt-1"
            placeholder="My Custom Theme"
          />
        </div>
        <div class="flex gap-2">
          <Button size="sm" @click="createCustomTheme" :disabled="!customThemeName.trim()">
            Create
          </Button>
          <Button size="sm" variant="ghost" @click="cancelCreatingCustom">
            Cancel
          </Button>
        </div>
      </div>
    </div>
  </div>
</template>
