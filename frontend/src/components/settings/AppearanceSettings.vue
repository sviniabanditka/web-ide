<script setup lang="ts">
import { computed } from 'vue'
import { useSettingsStore } from '@/stores/settings'
import CustomThemeEditor from './CustomThemeEditor.vue'
import Label from '@/components/ui/Label.vue'

const settingsStore = useSettingsStore()

const currentUIThemeId = computed(() => settingsStore.settings?.ui_theme_id || 'dark-plus')

function updateUITheme(themeId: string) {
  settingsStore.saveSettings({ ui_theme_id: themeId })
}
</script>

<template>
  <div class="space-y-6">
    <CustomThemeEditor
      theme-type="ui"
      :current-theme-id="currentUIThemeId"
      @update:themeId="updateUITheme"
    />

    <div class="pt-4 border-t">
      <Label class="text-base">UI Color Preview</Label>
      <div class="mt-3 p-4 rounded-lg border space-y-3" :style="{ backgroundColor: 'var(--background)', color: 'var(--foreground)' }">
        <div class="flex items-center gap-2">
          <span class="px-2 py-1 rounded text-xs font-medium" :style="{ backgroundColor: 'var(--primary)', color: 'var(--primary-foreground)' }">Primary</span>
          <span class="px-2 py-1 rounded text-xs font-medium" :style="{ backgroundColor: 'var(--secondary)', color: 'var(--secondary-foreground)' }">Secondary</span>
          <span class="px-2 py-1 rounded text-xs font-medium" :style="{ backgroundColor: 'var(--accent)', color: 'var(--accent-foreground)' }">Accent</span>
        </div>
        <div class="flex items-center gap-2">
          <span class="text-sm" :style="{ color: 'var(--muted-foreground)' }">Muted text</span>
          <span class="text-sm border px-2 py-0.5 rounded" :style="{ borderColor: 'var(--border)' }">Border</span>
          <span class="text-sm px-2 py-0.5 rounded" :style="{ backgroundColor: 'var(--card)', color: 'var(--card-foreground)' }">Card</span>
        </div>
        <div class="flex items-center gap-2">
          <span class="text-sm px-2 py-0.5 rounded bg-destructive/10" :style="{ color: 'var(--destructive)' }">Destructive</span>
        </div>
      </div>
    </div>
  </div>
</template>
