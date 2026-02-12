import { defineStore } from 'pinia'
import { ref } from 'vue'
import { api } from '../api'

export interface Theme {
  id: string
  name: string
  colors?: Record<string, string>
  isCustom?: boolean
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

export interface CustomTheme {
  id: string
  type: 'ui' | 'editor' | 'terminal'
  name: string
  colors: Record<string, string>
  created_at: string
  updated_at: string
}

export const useSettingsStore = defineStore('settings', () => {
  const settings = ref<UserSettings | null>(null)
  const customThemes = ref<CustomTheme[]>([])
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
    { id: 'monokai', name: 'Monokai' },
    { id: 'nord', name: 'Nord' },
    { id: 'dracula', name: 'Dracula' },
    { id: 'github-dark', name: 'GitHub Dark' },
    { id: 'github-light', name: 'GitHub Light' },
    { id: 'one-dark', name: 'One Dark' }
  ]

  const uiColorDefinitions = [
    { key: 'background', label: 'Background' },
    { key: 'foreground', label: 'Foreground' },
    { key: 'card', label: 'Card' },
    { key: 'card-foreground', label: 'Card Foreground' },
    { key: 'popover', label: 'Popover' },
    { key: 'popover-foreground', label: 'Popover Foreground' },
    { key: 'primary', label: 'Primary' },
    { key: 'primary-foreground', label: 'Primary Foreground' },
    { key: 'secondary', label: 'Secondary' },
    { key: 'secondary-foreground', label: 'Secondary Foreground' },
    { key: 'accent', label: 'Accent' },
    { key: 'accent-foreground', label: 'Accent Foreground' },
    { key: 'muted', label: 'Muted' },
    { key: 'muted-foreground', label: 'Muted Foreground' },
    { key: 'destructive', label: 'Destructive' },
    { key: 'destructive-foreground', label: 'Destructive Foreground' },
    { key: 'border', label: 'Border' },
    { key: 'input', label: 'Input' },
    { key: 'ring', label: 'Ring' }
  ]

  const editorColorDefinitions = [
    { key: 'editor.background', label: 'Background', monacoToken: 'background' },
    { key: 'editor.foreground', label: 'Foreground', monacoToken: 'foreground' },
    { key: 'editor.selectionBackground', label: 'Selection Background', monacoToken: 'selection' },
    { key: 'editor.selectionForeground', label: 'Selection Foreground', monacoToken: 'selectionForeground' },
    { key: 'editor.lineHighlightBackground', label: 'Line Highlight', monacoToken: 'lineHighlight' },
    { key: 'editorCursor.foreground', label: 'Cursor', monacoToken: 'cursor' },
    { key: 'editorLineNumber.foreground', label: 'Line Numbers', monacoToken: 'lineNumber' },
    { key: 'editorLineNumber.activeForeground', label: 'Active Line Number', monacoToken: 'activeLineNumber' }
  ]

  const terminalColorDefinitions = [
    { key: 'background', label: 'Background' },
    { key: 'foreground', label: 'Foreground' },
    { key: 'cursor', label: 'Cursor' },
    { key: 'selectionBackground', label: 'Selection Background' },
    { key: 'black', label: 'Black' },
    { key: 'red', label: 'Red' },
    { key: 'green', label: 'Green' },
    { key: 'yellow', label: 'Yellow' },
    { key: 'blue', label: 'Blue' },
    { key: 'magenta', label: 'Magenta' },
    { key: 'cyan', label: 'Cyan' },
    { key: 'white', label: 'White' },
    { key: 'brightBlack', label: 'Bright Black' },
    { key: 'brightRed', label: 'Bright Red' },
    { key: 'brightGreen', label: 'Bright Green' },
    { key: 'brightYellow', label: 'Bright Yellow' },
    { key: 'brightBlue', label: 'Bright Blue' },
    { key: 'brightMagenta', label: 'Bright Magenta' },
    { key: 'brightCyan', label: 'Bright Cyan' },
    { key: 'brightWhite', label: 'Bright White' }
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
    const predefined: Record<string, { background: string; foreground: string }> = {
      'monokai': { background: '#272822', foreground: '#f8f8f2' },
      'nord': { background: '#2e3440', foreground: '#d8dee9' },
      'dracula': { background: '#282a36', foreground: '#f8f8f2' },
      'github-dark': { background: '#0d1117', foreground: '#c9d1d9' },
      'github-light': { background: '#ffffff', foreground: '#24292f' },
      'one-dark': { background: '#282c34', foreground: '#abb2bf' }
    }
    return predefined[id] || { background: '#0d1117', foreground: '#c9d1d9' }
  }

  function getTerminalFullColors(id: string): Record<string, string> {
    const predefined: Record<string, Record<string, string>> = {
      'monokai': {
        background: '#272822', foreground: '#f8f8f2', cursor: '#f8f8f2', selectionBackground: '#49483e',
        black: '#484f58', red: '#ff7b72', green: '#a6e22e', yellow: '#e6db74',
        blue: '#66d9ef', magenta: '#f92672', cyan: '#a1efe4', white: '#f8f8f2',
        brightBlack: '#6e7681', brightRed: '#ff7b72', brightGreen: '#a6e22e', brightYellow: '#e6db74',
        brightBlue: '#66d9ef', brightMagenta: '#f92672', brightCyan: '#a1efe4', brightWhite: '#ffffff'
      },
      'nord': {
        background: '#2e3440', foreground: '#d8dee9', cursor: '#d8dee9', selectionBackground: '#434c5e',
        black: '#3b4252', red: '#bf616a', green: '#a3be8c', yellow: '#ebcb8b',
        blue: '#81a1c1', magenta: '#b48ead', cyan: '#88c0d0', white: '#e5e9f0',
        brightBlack: '#4c566a', brightRed: '#bf616a', brightGreen: '#a3be8c', brightYellow: '#ebcb8b',
        brightBlue: '#81a1c1', brightMagenta: '#b48ead', brightCyan: '#8fbcbb', brightWhite: '#eceff4'
      },
      'dracula': {
        background: '#282a36', foreground: '#f8f8f2', cursor: '#f8f8f2', selectionBackground: '#44475a',
        black: '#44475a', red: '#ff5555', green: '#50fa7b', yellow: '#f1fa8c',
        blue: '#bd93f9', magenta: '#ff79c6', cyan: '#8be9fd', white: '#f8f8f2',
        brightBlack: '#6272a4', brightRed: '#ff6e6e', brightGreen: '#69ff94', brightYellow: '#ffffa5',
        brightBlue: '#d6acff', brightMagenta: '#ff92df', brightCyan: '#a4ffff', brightWhite: '#ffffff'
      },
      'github-dark': {
        background: '#0d1117', foreground: '#c9d1d9', cursor: '#c9d1d9', selectionBackground: '#264f78',
        black: '#484f58', red: '#ff7b72', green: '#3fb950', yellow: '#d29922',
        blue: '#58a6ff', magenta: '#bc8cff', cyan: '#39c5cf', white: '#c9d1d9',
        brightBlack: '#6e7681', brightRed: '#ffa39e', brightGreen: '#7ee787', brightYellow: '#d2a8ff',
        brightBlue: '#79c0ff', brightMagenta: '#d2a8ff', brightCyan: '#56d4db', brightWhite: '#ffffff'
      },
      'github-light': {
        background: '#ffffff', foreground: '#24292f', cursor: '#24292f', selectionBackground: '#0969da',
        black: '#24292f', red: '#cf222e', green: '#1a7f37', yellow: '#9a6700',
        blue: '#0969da', magenta: '#8250df', cyan: '#0550ae', white: '#24292f',
        brightBlack: '#57606a', brightRed: '#ff818266', brightGreen: '#1a7f37', brightYellow: '#9a6700',
        brightBlue: '#0969da', brightMagenta: '#8250df', brightCyan: '#0550ae', brightWhite: '#24292f'
      },
      'one-dark': {
        background: '#282c34', foreground: '#abb2bf', cursor: '#528bff', selectionBackground: '#3e4451',
        black: '#5c6370', red: '#e06c75', green: '#98c379', yellow: '#e5c07b',
        blue: '#61afef', magenta: '#c678dd', cyan: '#56b6c2', white: '#abb2bf',
        brightBlack: '#5c6370', brightRed: '#e06c75', brightGreen: '#98c379', brightYellow: '#e5c07b',
        brightBlue: '#61afef', brightMagenta: '#c678dd', brightCyan: '#56b6c2', brightWhite: '#ffffff'
      }
    }
    return predefined[id] || predefined['github-dark']
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

      await fetchCustomThemes()
    } catch (e: any) {
      error.value = e.response?.data?.error || 'Failed to fetch settings'
    } finally {
      loading.value = false
    }
  }

  async function fetchCustomThemes(themeType?: string) {
    try {
      let url = '/api/v1/settings/themes/custom'
      if (themeType) {
        url += `?type=${themeType}`
      }
      const response = await api.get(url)
      customThemes.value = response.data
    } catch (e: any) {
      console.warn('Failed to fetch custom themes:', e)
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

  async function createCustomTheme(type: 'ui' | 'editor' | 'terminal', name: string, colors: Record<string, string>) {
    loading.value = true
    error.value = null
    try {
      const response = await api.post('/api/v1/settings/themes/custom', {
        type,
        name,
        colors
      })
      await fetchCustomThemes(type)
      return response.data
    } catch (e: any) {
      error.value = e.response?.data?.error || 'Failed to create custom theme'
      return null
    } finally {
      loading.value = false
    }
  }

  async function updateCustomTheme(themeId: string, colors: Record<string, string>) {
    loading.value = true
    error.value = null
    try {
      const response = await api.put(`/api/v1/settings/themes/custom/${themeId}`, {
        colors
      })
      if (response.data.status === 'saved') {
        await fetchCustomThemes()
      }
      return true
    } catch (e: any) {
      error.value = e.response?.data?.error || 'Failed to update custom theme'
      return false
    } finally {
      loading.value = false
    }
  }

  async function deleteCustomTheme(themeId: string, type?: string) {
    loading.value = true
    error.value = null
    try {
      const response = await api.delete(`/api/v1/settings/themes/custom/${themeId}`)
      if (response.data.status === 'deleted') {
        await fetchCustomThemes(type)
      }
      return true
    } catch (e: any) {
      error.value = e.response?.data?.error || 'Failed to delete custom theme'
      return false
    } finally {
      loading.value = false
    }
  }

  function getCustomThemesByType(type: 'ui' | 'editor' | 'terminal'): CustomTheme[] {
    return customThemes.value.filter(t => t.type === type)
  }

  function getFullUITheme(themeId: string): Record<string, string> {
    const customTheme = customThemes.value.find(t => t.id === themeId && t.type === 'ui')
    if (customTheme) {
      return customTheme.colors
    }

    const themes: Record<string, Record<string, string>> = {
      'dark-plus': {
        background: '#1e1e1e', foreground: '#d4d4d4', muted: '#2d2d30', 'muted-foreground': '#858585',
        border: '#454545', card: '#252526', 'card-foreground': '#cccccc', primary: '#007acc',
        'primary-foreground': '#ffffff', secondary: '#3c3c3c', 'secondary-foreground': '#cccccc',
        accent: '#3c3c3c', 'accent-foreground': '#cccccc', destructive: '#d13438',
        'destructive-foreground': '#ffffff', popover: '#1e1e1e', 'popover-foreground': '#d4d4d4',
        input: '#454545', ring: '#007acc'
      },
      'light-plus': {
        background: '#ffffff', foreground: '#333333', muted: '#f3f3f3', 'muted-foreground': '#666666',
        border: '#e0e0e0', card: '#ffffff', 'card-foreground': '#333333', primary: '#007acc',
        'primary-foreground': '#ffffff', secondary: '#f3f3f3', 'secondary-foreground': '#333333',
        accent: '#f3f3f3', 'accent-foreground': '#333333', destructive: '#d13438',
        'destructive-foreground': '#ffffff', popover: '#ffffff', 'popover-foreground': '#333333',
        input: '#e0e0e0', ring: '#007acc'
      },
      'monokai': {
        background: '#272822', foreground: '#f8f8f2', muted: '#3e3d32', 'muted-foreground': '#a6a6a6',
        border: '#49483e', card: '#2d2d2a', 'card-foreground': '#f8f8f2', primary: '#a6e22e',
        'primary-foreground': '#272822', secondary: '#3e3d32', 'secondary-foreground': '#f8f8f2',
        accent: '#3e3d32', 'accent-foreground': '#f8f8f2', destructive: '#f92672',
        'destructive-foreground': '#f8f8f2', popover: '#272822', 'popover-foreground': '#f8f8f2',
        input: '#49483e', ring: '#a6e22e'
      },
      'nord': {
        background: '#2e3440', foreground: '#d8dee9', muted: '#3b4252', 'muted-foreground': '#81a1c1',
        border: '#4c566a', card: '#3b4252', 'card-foreground': '#d8dee9', primary: '#88c0d0',
        'primary-foreground': '#2e3440', secondary: '#434c5e', 'secondary-foreground': '#d8dee9',
        accent: '#434c5e', 'accent-foreground': '#d8dee9', destructive: '#bf616a',
        'destructive-foreground': '#2e3440', popover: '#2e3440', 'popover-foreground': '#d8dee9',
        input: '#4c566a', ring: '#88c0d0'
      },
      'dracula': {
        background: '#282a36', foreground: '#f8f8f2', muted: '#44475a', 'muted-foreground': '#6272a4',
        border: '#44475a', card: '#21222c', 'card-foreground': '#f8f8f2', primary: '#bd93f9',
        'primary-foreground': '#282a36', secondary: '#44475a', 'secondary-foreground': '#f8f8f2',
        accent: '#44475a', 'accent-foreground': '#f8f8f2', destructive: '#ff5555',
        'destructive-foreground': '#f8f8f2', popover: '#282a36', 'popover-foreground': '#f8f8f2',
        input: '#44475a', ring: '#bd93f9'
      }
    }
    return themes[themeId] || themes['dark-plus']
  }

  return {
    settings,
    customThemes,
    loading,
    error,
    uiThemes,
    editorThemes,
    terminalThemes,
    uiColorDefinitions,
    editorColorDefinitions,
    terminalColorDefinitions,
    fetchSettings,
    saveSettings,
    fetchCustomThemes,
    createCustomTheme,
    updateCustomTheme,
    deleteCustomTheme,
    getCustomThemesByType,
    getEditorThemeById,
    getTerminalThemeColors,
    getTerminalFullColors,
    getFullUITheme
  }
})
