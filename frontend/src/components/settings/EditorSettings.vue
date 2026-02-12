<script setup lang="ts">
import { computed } from 'vue'
import Label from '@/components/ui/Label.vue'
import { useSettingsStore } from '@/stores/settings'
import CustomThemeEditor from './CustomThemeEditor.vue'

const settingsStore = useSettingsStore()

const currentEditorThemeId = computed(() => settingsStore.settings?.editor_theme_id || 'vs-dark')

function updateEditorTheme(themeId: string) {
  settingsStore.saveSettings({ editor_theme_id: themeId })
}
</script>

<template>
  <div class="space-y-6">
    <CustomThemeEditor
      theme-type="editor"
      :current-theme-id="currentEditorThemeId"
      @update:themeId="updateEditorTheme"
    />

    <div class="pt-4 border-t">
      <Label class="text-base">Editor Theme Preview</Label>
      <p class="text-sm text-muted-foreground mt-1 mb-3">
        Selected: <span class="font-medium">{{ settingsStore.editorThemes.find(t => t.id === currentEditorThemeId)?.name }}</span>
      </p>

      <div
        class="rounded-lg border overflow-hidden"
        :style="{ backgroundColor: '#1e1e1e', borderColor: '#454545' }"
      >
        <div class="flex items-center gap-1 p-2 border-b" :style="{ borderColor: '#454545', backgroundColor: '#252526' }">
          <div class="w-3 h-3 rounded-full" :style="{ backgroundColor: '#ff5f56' }" />
          <div class="w-3 h-3 rounded-full" :style="{ backgroundColor: '#ffbd2e' }" />
          <div class="w-3 h-3 rounded-full" :style="{ backgroundColor: '#27c93f' }" />
          <span class="text-xs ml-2" :style="{ color: '#858585' }">editor.js</span>
        </div>
        <div class="p-4 font-mono text-sm">
          <div class="flex">
            <span class="w-6 text-right mr-3 select-none" :style="{ color: '#858585' }">1</span>
            <span :style="{ color: '#d4d4d4' }"><span :style="{ color: '#569cd6' }">function</span> <span :style="{ color: '#dcdcaa' }">hello</span>() {'{'}</span>
          </div>
          <div class="flex">
            <span class="w-6 text-right mr-3 select-none" :style="{ color: '#858585' }">2</span>
            <span class="pl-4" :style="{ color: '#d4d4d4' }"><span :style="{ color: '#c586c0' }">return</span> <span :style="{ color: '#ce9178' }">'Hello!'</span>;</span>
          </div>
          <div class="flex">
            <span class="w-6 text-right mr-3 select-none" :style="{ color: '#858585' }">3</span>
            <span :style="{ color: '#d4d4d4' }">{'}'}</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
