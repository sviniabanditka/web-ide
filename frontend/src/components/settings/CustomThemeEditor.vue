<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { useSettingsStore, type CustomTheme } from '@/stores/settings'
import Button from '@/components/ui/Button.vue'
import Label from '@/components/ui/Label.vue'
import ColorPicker from '@/components/ui/ColorPicker.vue'
import { Trash2, Save } from 'lucide-vue-next'

interface Props {
  themeType: 'ui' | 'editor' | 'terminal'
  currentThemeId: string
}

const props = defineProps<Props>()
const emit = defineEmits<{
  'update:themeId': [value: string]
}>()

const settingsStore = useSettingsStore()

const customThemes = computed(() => settingsStore.getCustomThemesByType(props.themeType))
const predefinedThemes = computed(() => {
  switch (props.themeType) {
    case 'ui': return settingsStore.uiThemes
    case 'editor': return settingsStore.editorThemes
    case 'terminal': return settingsStore.terminalThemes
    default: return []
  }
})

const allThemes = computed(() => {
  return [
    ...predefinedThemes.value.map(t => ({ ...t, isCustom: false })),
    ...customThemes.value.map(t => ({ id: t.id, name: t.name, isCustom: true, colors: t.colors }))
  ]
})

const isCreating = ref(false)
const editingTheme = ref<CustomTheme | null>(null)
const newThemeName = ref('')
const newThemeColors = ref<Record<string, string>>({})
const isSaving = ref(false)

const colorDefinitions = computed(() => {
  switch (props.themeType) {
    case 'ui': return settingsStore.uiColorDefinitions
    case 'editor': return settingsStore.editorColorDefinitions
    case 'terminal': return settingsStore.terminalColorDefinitions
    default: return []
  }
})

const getDefaultColors = (): Record<string, string> => {
  switch (props.themeType) {
    case 'ui':
      return {
        background: '#1e1e1e', foreground: '#d4d4d4', card: '#252526', 'card-foreground': '#cccccc',
        popover: '#1e1e1e', 'popover-foreground': '#d4d4d4', primary: '#007acc', 'primary-foreground': '#ffffff',
        secondary: '#3c3c3c', 'secondary-foreground': '#cccccc', accent: '#3c3c3c', 'accent-foreground': '#cccccc',
        muted: '#2d2d30', 'muted-foreground': '#858585', destructive: '#d13438', 'destructive-foreground': '#ffffff',
        border: '#454545', input: '#454545', ring: '#007acc'
      }
    case 'editor':
      return {
        'editor.background': '#1e1e1e', 'editor.foreground': '#d4d4d4',
        'editor.selectionBackground': '#264f78', 'editor.selectionForeground': '#ffffff',
        'editor.lineHighlightBackground': '#2d2d30', 'editorCursor.foreground': '#aeafad',
        'editorLineNumber.foreground': '#858585', 'editorLineNumber.activeForeground': '#c6c6c6'
      }
    case 'terminal':
      return {
        background: '#0d1117', foreground: '#c9d1d9', cursor: '#c9d1d9', selectionBackground: '#264f78',
        black: '#484f58', red: '#ff7b72', green: '#3fb950', yellow: '#d29922',
        blue: '#58a6ff', magenta: '#bc8cff', cyan: '#39c5cf', white: '#c9d1d9',
        brightBlack: '#6e7681', brightRed: '#ffa39e', brightGreen: '#7ee787', brightYellow: '#d2a8ff',
        brightBlue: '#79c0ff', brightMagenta: '#ff7b72', brightCyan: '#56d4db', brightWhite: '#ffffff'
      }
    default:
      return {}
  }
}

function startCreating() {
  isCreating.value = true
  editingTheme.value = null
  newThemeName.value = ''
  newThemeColors.value = getDefaultColors()
}

function cancelCreating() {
  isCreating.value = false
  editingTheme.value = null
  newThemeName.value = ''
  newThemeColors.value = {}
}

function selectTheme(themeId: string) {
  if (editingTheme.value?.id !== themeId) {
    editingTheme.value = null
  }
  emit('update:themeId', themeId)
}

async function saveNewTheme() {
  if (!newThemeName.value.trim()) return

  isSaving.value = true
  const result = await settingsStore.createCustomTheme(props.themeType, newThemeName.value.trim(), newThemeColors.value)
  isSaving.value = false

  if (result) {
    isCreating.value = false
    newThemeName.value = ''
    newThemeColors.value = {}
    emit('update:themeId', result.id)
  }
}

async function saveThemeChanges() {
  if (!editingTheme.value) return

  isSaving.value = true
  await settingsStore.updateCustomTheme(editingTheme.value.id, newThemeColors.value)
  isSaving.value = false
}

async function deleteTheme() {
  if (!editingTheme.value) return

  if (confirm(`Delete theme "${editingTheme.value.name}"?`)) {
    await settingsStore.deleteCustomTheme(editingTheme.value.id, props.themeType)
    editingTheme.value = null
    emit('update:themeId', predefinedThemes.value[0]?.id || '')
  }
}

function updateColor(key: string, value: string) {
  newThemeColors.value = { ...newThemeColors.value, [key]: value }
}

watch(() => props.currentThemeId, (newId) => {
  if (!isCreating.value && !editingTheme.value) {
    const customTheme = customThemes.value.find(t => t.id === newId)
    if (customTheme) {
      editingTheme.value = customTheme
      newThemeColors.value = { ...customTheme.colors }
    }
  }
})

onMounted(() => {
  const customTheme = customThemes.value.find(t => t.id === props.currentThemeId)
  if (customTheme) {
    editingTheme.value = customTheme
    newThemeColors.value = { ...customTheme.colors }
  }
})
</script>

<template>
  <div class="space-y-4">
    <div v-if="!isCreating && !editingTheme">
      <Label class="text-base mb-3 block">Select Theme</Label>
      <div class="grid grid-cols-2 md:grid-cols-3 gap-2">
        <button
          v-for="theme in allThemes"
          :key="theme.id"
          @click="selectTheme(theme.id)"
          :class="[
            'relative p-3 rounded-lg border-2 text-left transition-all',
            currentThemeId === theme.id
              ? 'border-primary ring-2 ring-primary/20'
              : 'border-border hover:border-primary/50'
          ]"
        >
          <div class="font-medium text-sm">{{ theme.name }}</div>
          <div v-if="theme.isCustom" class="text-xs text-muted-foreground">Custom</div>
          <div
            v-if="currentThemeId === theme.id"
            class="absolute top-2 right-2 w-2 h-2 rounded-full bg-primary"
          />
        </button>

        <button
          @click="startCreating"
          class="p-3 rounded-lg border-2 border-dashed border-border hover:border-primary/50 text-left transition-all flex items-center justify-center min-h-[60px]"
        >
          <span class="text-sm text-muted-foreground">+ Create Theme</span>
        </button>
      </div>
    </div>

    <div v-if="isCreating || editingTheme" class="space-y-4">
      <div class="flex items-center justify-between">
        <Label class="text-base">
          {{ isCreating ? 'Create New Theme' : `Edit: ${editingTheme?.name}` }}
        </Label>
        <div class="flex gap-2">
          <Button
            v-if="editingTheme"
            variant="ghost"
            size="sm"
            @click="deleteTheme"
            class="text-destructive hover:text-destructive"
          >
            <Trash2 class="w-4 h-4 mr-1" />
            Delete
          </Button>
          <Button
            v-if="isCreating"
            variant="ghost"
            size="sm"
            @click="cancelCreating"
          >
            Cancel
          </Button>
          <Button
            size="sm"
            @click="isCreating ? saveNewTheme() : saveThemeChanges()"
            :disabled="isSaving || (isCreating && !newThemeName.trim())"
          >
            <Save class="w-4 h-4 mr-1" />
            {{ isSaving ? 'Saving...' : 'Save' }}
          </Button>
        </div>
      </div>

      <div v-if="isCreating" class="space-y-2">
        <Label for="themeName">Theme Name</Label>
        <input
          id="themeName"
          v-model="newThemeName"
          type="text"
          class="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
          placeholder="My Custom Theme"
        />
      </div>

      <div class="border-t pt-4">
        <Label class="text-sm mb-3 block">Colors</Label>
        <div class="grid grid-cols-1 md:grid-cols-2 gap-3 max-h-[400px] overflow-y-auto pr-2">
          <ColorPicker
            v-for="def in colorDefinitions"
            :key="def.key"
            :model-value="newThemeColors[def.key] || '#000000'"
            :label="def.label"
            @update:modelValue="updateColor(def.key, $event)"
          />
        </div>
      </div>
    </div>
  </div>
</template>
