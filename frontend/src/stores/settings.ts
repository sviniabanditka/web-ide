import { defineStore } from 'pinia'
import { ref } from 'vue'
import { api } from '../api'

export interface Theme {
  id: string
  name: string
  colors?: Record<string, string>
}

export interface UserSettings {
  id: string
  user_id: string
  ai_provider: string
  ai_base_url: string
  ai_api_key: string
  ai_model: string
  ui_theme_id: string
  editor_theme_id: string
  terminal_theme_id: string
  custom_theme_json: string
  created_at: string
  updated_at: string
}

export const useSettingsStore = defineStore('settings', () => {
  const settings = ref<UserSettings | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  const uiThemes: Theme[] = [
    { id: 'dark-plus', name: 'Dark+' },
    { id: 'light-plus', name: 'Light+' },
    { id: 'monokai', name: 'Monokai' },
    { id: 'nord', name: 'Nord' },
    { id: 'dracula', name: 'Dracula' }
  ]

  const editorThemes: Theme[] = [
    { id: 'vs-dark', name: 'Dark (VS Code)' },
    { id: 'vs', name: 'Light (VS Code)' },
    { id: 'hc-black', name: 'High Contrast' }
  ]

  const terminalThemes: Theme[] = [
    { id: 'monokai', name: 'Monokai', colors: { background: '#272822', foreground: '#f8f8f2' } },
    { id: 'nord', name: 'Nord', colors: { background: '#2e3440', foreground: '#d8dee9' } },
    { id: 'dracula', name: 'Dracula', colors: { background: '#282a36', foreground: '#f8f8f2' } },
    { id: 'github-dark', name: 'GitHub Dark', colors: { background: '#0d1117', foreground: '#c9d1d9' } },
    { id: 'github-light', name: 'GitHub Light', colors: { background: '#ffffff', foreground: '#24292f' } },
    { id: 'one-dark', name: 'One Dark', colors: { background: '#282c34', foreground: '#abb2bf' } }
  ]

  const defaultDarkUITheme: Record<string, string> = {
    background: '#1e1e1e',
    foreground: '#d4d4d4',
    muted: '#2d2d30',
    'muted-foreground': '#858585',
    border: '#454545',
    card: '#252526',
    'card-foreground': '#cccccc',
    primary: '#007acc',
    'primary-foreground': '#ffffff',
    secondary: '#3c3c3c',
    'secondary-foreground': '#cccccc',
    accent: '#3c3c3c',
    'accent-foreground': '#cccccc',
    destructive: '#d13438',
    'destructive-foreground': '#ffffff',
    popover: '#1e1e1e',
    'popover-foreground': '#d4d4d4',
    input: '#454545',
    ring: '#007acc'
  }

  function applyUITheme(colors: Record<string, string>) {
    const root = document.documentElement
    Object.entries(colors).forEach(([key, value]) => {
      root.style.setProperty(`--${key}`, value)
    })
  }

  applyUITheme(defaultDarkUITheme)

  function getEditorThemeById(id: string): string {
    return editorThemes.find(t => t.id === id)?.id || 'vs-dark'
  }

  function getTerminalThemeColors(id: string): { background: string; foreground: string } {
    const theme = terminalThemes.find(t => t.id === id)
    if (theme?.colors) {
      return { background: theme.colors.background, foreground: theme.colors.foreground }
    }
    return { background: '#0d1117', foreground: '#c9d1d9' }
  }

  async function fetchSettings() {
    loading.value = true
    error.value = null
    try {
      const response = await api.get('/api/v1/settings')
      settings.value = response.data

      if (response.data.ui_theme_id) {
        const uiTheme = uiThemes.find(t => t.id === response.data.ui_theme_id)
        if (uiTheme) {
          const fullTheme = getFullUITheme(uiTheme.id)
          applyUITheme(fullTheme)
        }
      }
    } catch (e: any) {
      error.value = e.response?.data?.error || 'Failed to fetch settings'
    } finally {
      loading.value = false
    }
  }

  async function saveSettings(newSettings: Partial<UserSettings>) {
    loading.value = true
    error.value = null
    try {
      const currentSettings = settings.value
      const mergedSettings = {
        ai_provider: currentSettings?.ai_provider || 'anthropic',
        ai_base_url: currentSettings?.ai_base_url || '',
        ai_api_key: currentSettings?.ai_api_key || '',
        ai_model: currentSettings?.ai_model || 'claude-sonnet-4-20250514',
        ui_theme_id: currentSettings?.ui_theme_id || 'dark-plus',
        editor_theme_id: currentSettings?.editor_theme_id || 'vs-dark',
        terminal_theme_id: currentSettings?.terminal_theme_id || 'monokai',
        custom_theme_json: currentSettings?.custom_theme_json || '{}',
      }

      if (newSettings.ai_provider) mergedSettings.ai_provider = newSettings.ai_provider
      if (newSettings.ai_base_url !== undefined) mergedSettings.ai_base_url = newSettings.ai_base_url
      if (newSettings.ai_api_key !== undefined) mergedSettings.ai_api_key = newSettings.ai_api_key
      if (newSettings.ai_model) mergedSettings.ai_model = newSettings.ai_model
      if (newSettings.ui_theme_id) mergedSettings.ui_theme_id = newSettings.ui_theme_id
      if (newSettings.editor_theme_id) mergedSettings.editor_theme_id = newSettings.editor_theme_id
      if (newSettings.terminal_theme_id) mergedSettings.terminal_theme_id = newSettings.terminal_theme_id
      if (newSettings.custom_theme_json !== undefined) mergedSettings.custom_theme_json = newSettings.custom_theme_json

      const response = await api.put('/api/v1/settings', mergedSettings)
      if (response.data.status === 'saved') {
        await fetchSettings()
      }
      return true
    } catch (e: any) {
      error.value = e.response?.data?.error || 'Failed to save settings'
      return false
    } finally {
      loading.value = false
    }
  }

  function getFullUITheme(themeId: string): Record<string, string> {
    const themes: Record<string, Record<string, string>> = {
      'dark-plus': {
        background: '#1e1e1e',
        foreground: '#d4d4d4',
        muted: '#2d2d30',
        'muted-foreground': '#858585',
        border: '#454545',
        card: '#252526',
        'card-foreground': '#cccccc',
        primary: '#007acc',
        'primary-foreground': '#ffffff',
        secondary: '#3c3c3c',
        'secondary-foreground': '#cccccc',
        accent: '#3c3c3c',
        'accent-foreground': '#cccccc',
        destructive: '#d13438',
        'destructive-foreground': '#ffffff',
        popover: '#1e1e1e',
        'popover-foreground': '#d4d4d4',
        input: '#454545',
        ring: '#007acc'
      },
      'light-plus': {
        background: '#ffffff',
        foreground: '#333333',
        muted: '#f3f3f3',
        'muted-foreground': '#666666',
        border: '#e0e0e0',
        card: '#ffffff',
        'card-foreground': '#333333',
        primary: '#007acc',
        'primary-foreground': '#ffffff',
        secondary: '#f3f3f3',
        'secondary-foreground': '#333333',
        accent: '#f3f3f3',
        'accent-foreground': '#333333',
        destructive: '#d13438',
        'destructive-foreground': '#ffffff',
        popover: '#ffffff',
        'popover-foreground': '#333333',
        input: '#e0e0e0',
        ring: '#007acc'
      },
      'monokai': {
        background: '#272822',
        foreground: '#f8f8f2',
        muted: '#3e3d32',
        'muted-foreground': '#a6a6a6',
        border: '#49483e',
        card: '#2d2d2a',
        'card-foreground': '#f8f8f2',
        primary: '#a6e22e',
        'primary-foreground': '#272822',
        secondary: '#3e3d32',
        'secondary-foreground': '#f8f8f2',
        accent: '#3e3d32',
        'accent-foreground': '#f8f8f2',
        destructive: '#f92672',
        'destructive-foreground': '#f8f8f2',
        popover: '#272822',
        'popover-foreground': '#f8f8f2',
        input: '#49483e',
        ring: '#a6e22e'
      },
      'nord': {
        background: '#2e3440',
        foreground: '#d8dee9',
        muted: '#3b4252',
        'muted-foreground': '#81a1c1',
        border: '#4c566a',
        card: '#3b4252',
        'card-foreground': '#d8dee9',
        primary: '#88c0d0',
        'primary-foreground': '#2e3440',
        secondary: '#434c5e',
        'secondary-foreground': '#d8dee9',
        accent: '#434c5e',
        'accent-foreground': '#d8dee9',
        destructive: '#bf616a',
        'destructive-foreground': '#2e3440',
        popover: '#2e3440',
        'popover-foreground': '#d8dee9',
        input: '#4c566a',
        ring: '#88c0d0'
      },
      'dracula': {
        background: '#282a36',
        foreground: '#f8f8f2',
        muted: '#44475a',
        'muted-foreground': '#6272a4',
        border: '#44475a',
        card: '#21222c',
        'card-foreground': '#f8f8f2',
        primary: '#bd93f9',
        'primary-foreground': '#282a36',
        secondary: '#44475a',
        'secondary-foreground': '#f8f8f2',
        accent: '#44475a',
        'accent-foreground': '#f8f8f2',
        destructive: '#ff5555',
        'destructive-foreground': '#f8f8f2',
        popover: '#282a36',
        'popover-foreground': '#f8f8f2',
        input: '#44475a',
        ring: '#bd93f9'
      }
    }
    return themes[themeId] || themes['dark-plus']
  }

  return {
    settings,
    loading,
    error,
    uiThemes,
    editorThemes,
    terminalThemes,
    fetchSettings,
    saveSettings,
    getEditorThemeById,
    getTerminalThemeColors,
    getFullUITheme
  }
})
