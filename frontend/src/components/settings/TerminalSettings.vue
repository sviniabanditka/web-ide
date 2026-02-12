<script setup lang="ts">
import { computed } from 'vue'
import Label from '@/components/ui/Label.vue'
import { useSettingsStore } from '@/stores/settings'

const settingsStore = useSettingsStore()

const themes = computed(() => settingsStore.terminalThemes)
const currentThemeId = computed(() => settingsStore.settings?.terminal_theme_id || 'github-dark')

const currentTheme = computed(() => {
  return settingsStore.getTerminalThemeColors(currentThemeId.value)
})

function selectTheme(themeId: string) {
  settingsStore.saveSettings({ terminal_theme_id: themeId })
}
</script>

<template>
  <div class="space-y-6">
    <div>
      <Label class="text-base">Terminal Color Scheme</Label>
      <p class="text-sm text-muted-foreground mt-1 mb-4">
        Choose a color theme for the terminal emulator (xterm.js)
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
          <div class="flex flex-col gap-1">
            <div
              class="h-12 rounded flex flex-col items-center justify-center gap-0.5"
              :style="{
                backgroundColor: theme.colors?.background || '#0d1117',
                color: theme.colors?.foreground || '#c9d1d9'
              }"
            >
              <div class="flex gap-0.5">
                <div
                  v-for="i in 4"
                  :key="i"
                  class="w-3 h-3 rounded-sm"
                  :style="{
                    backgroundColor: i === 1 ? '#007acc' : i === 2 ? '#3fb950' : i === 3 ? '#d29922' : '#bc8cff'
                  }"
                />
              </div>
              <div class="text-[8px] mt-1 opacity-75">$ echo "test"</div>
            </div>
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
      <Label class="text-base">Terminal Preview</Label>
      <p class="text-sm text-muted-foreground mt-1 mb-4">
        Preview of the selected terminal theme
      </p>

      <div
        class="p-4 rounded-lg border font-mono text-sm"
        :style="{
          backgroundColor: currentTheme.background,
          color: currentTheme.foreground
        }"
      >
        <div class="flex items-center gap-2 mb-3">
          <div class="flex gap-1">
            <div class="w-3 h-3 rounded-full bg-red-500 opacity-50" />
            <div class="w-3 h-3 rounded-full bg-yellow-500 opacity-50" />
            <div class="w-3 h-3 rounded-full bg-green-500 opacity-50" />
          </div>
          <span class="text-xs opacity-50">bash</span>
        </div>
        <div class="space-y-1">
          <div><span style="color: #007acc">user@webide</span>:<span style="color: #3fb950">~</span>$ npm run dev</div>
          <div style="color: #3fb950">&gt; dev</div>
          <div style="color: #d29922">VITE v5.x.x ready in 300 ms</div>
          <div><span style="color: #007acc">➜</span> <span>Local:</span> http://localhost:5173</div>
        </div>
      </div>
    </div>

    <div class="pt-4 border-t">
      <Label class="text-base">Theme Colors</Label>
      <div class="grid grid-cols-2 gap-4 mt-2">
        <div class="flex items-center gap-3">
          <div
            class="w-8 h-8 rounded border"
            :style="{ backgroundColor: currentTheme.background }"
          />
          <div>
            <div class="text-sm font-medium">Background</div>
            <div class="text-xs font-mono text-muted-foreground">{{ currentTheme.background }}</div>
          </div>
        </div>
        <div class="flex items-center gap-3">
          <div
            class="w-8 h-8 rounded border"
            :style="{ backgroundColor: currentTheme.foreground }"
          />
          <div>
            <div class="text-sm font-medium">Foreground</div>
            <div class="text-xs font-mono text-muted-foreground">{{ currentTheme.foreground }}</div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
