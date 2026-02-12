<script setup lang="ts">
import { computed } from 'vue'
import Label from '@/components/ui/Label.vue'
import { useSettingsStore } from '@/stores/settings'
import CustomThemeEditor from './CustomThemeEditor.vue'

const settingsStore = useSettingsStore()

const currentTerminalThemeId = computed(() => settingsStore.settings?.terminal_theme_id || 'github-dark')

const currentThemeColors = computed(() => {
  return settingsStore.getTerminalFullColors(currentTerminalThemeId.value)
})

function updateTerminalTheme(themeId: string) {
  settingsStore.saveSettings({ terminal_theme_id: themeId })
}
</script>

<template>
  <div class="space-y-6">
    <CustomThemeEditor
      theme-type="terminal"
      :current-theme-id="currentTerminalThemeId"
      @update:themeId="updateTerminalTheme"
    />

    <div class="pt-4 border-t">
      <Label class="text-base">Terminal Preview</Label>
      <div
        class="mt-3 rounded-lg border overflow-hidden font-mono text-sm"
        :style="{
          backgroundColor: currentThemeColors.background,
          color: currentThemeColors.foreground,
          borderColor: currentThemeColors.background
        }"
      >
        <div class="flex items-center gap-2 px-3 py-2 border-b" :style="{ borderColor: currentThemeColors.background }">
          <div class="flex gap-1">
            <div class="w-3 h-3 rounded-full opacity-50" :style="{ backgroundColor: currentThemeColors.foreground }" />
            <div class="w-3 h-3 rounded-full opacity-50" :style="{ backgroundColor: currentThemeColors.foreground }" />
            <div class="w-3 h-3 rounded-full opacity-50" :style="{ backgroundColor: currentThemeColors.foreground }" />
          </div>
          <span class="text-xs opacity-50">bash</span>
        </div>
        <div class="p-4 space-y-1">
          <div>
            <span :style="{ color: currentThemeColors.green }">user@webide</span>:
            <span :style="{ color: currentThemeColors.yellow }">~</span>$ npm run dev
          </div>
          <div :style="{ color: currentThemeColors.green }">> dev</div>
          <div :style="{ color: currentThemeColors.yellow }">VITE v5.x.x ready in 300 ms</div>
          <div>
            <span :style="{ color: currentThemeColors.blue }">➜</span>
            <span class="ml-1">Local:</span> http://localhost:5173
          </div>
          <div class="mt-2">
            <span :style="{ color: currentThemeColors.blue }">➜</span>
            <span class="ml-1 animate-pulse">_</span>
          </div>
        </div>
      </div>
    </div>

    <div class="pt-4 border-t">
      <Label class="text-sm mb-3 block">Terminal Colors</Label>
      <div class="grid grid-cols-2 md:grid-cols-4 gap-3">
        <div class="flex items-center gap-2">
          <div class="w-6 h-6 rounded border" :style="{ backgroundColor: currentThemeColors.background }" />
          <span class="text-xs">Background</span>
        </div>
        <div class="flex items-center gap-2">
          <div class="w-6 h-6 rounded border" :style="{ backgroundColor: currentThemeColors.foreground }" />
          <span class="text-xs">Foreground</span>
        </div>
        <div class="flex items-center gap-2">
          <div class="w-6 h-6 rounded border" :style="{ backgroundColor: currentThemeColors.red }" />
          <span class="text-xs">Red</span>
        </div>
        <div class="flex items-center gap-2">
          <div class="w-6 h-6 rounded border" :style="{ backgroundColor: currentThemeColors.green }" />
          <span class="text-xs">Green</span>
        </div>
        <div class="flex items-center gap-2">
          <div class="w-6 h-6 rounded border" :style="{ backgroundColor: currentThemeColors.yellow }" />
          <span class="text-xs">Yellow</span>
        </div>
        <div class="flex items-center gap-2">
          <div class="w-6 h-6 rounded border" :style="{ backgroundColor: currentThemeColors.blue }" />
          <span class="text-xs">Blue</span>
        </div>
        <div class="flex items-center gap-2">
          <div class="w-6 h-6 rounded border" :style="{ backgroundColor: currentThemeColors.magenta }" />
          <span class="text-xs">Magenta</span>
        </div>
        <div class="flex items-center gap-2">
          <div class="w-6 h-6 rounded border" :style="{ backgroundColor: currentThemeColors.cyan }" />
          <span class="text-xs">Cyan</span>
        </div>
      </div>
    </div>
  </div>
</template>
