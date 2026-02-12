<template>
  <div class="terminal-pane" :style="{ height: height + 'px' }">
    <div ref="terminalRef" class="terminal-container"></div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch, nextTick } from 'vue'
import { Terminal } from 'xterm'
import { FitAddon } from 'xterm-addon-fit'
import { useDebounceFn } from '@vueuse/core'
import { useTerminalsStore } from '../stores/terminals'
import { useSettingsStore } from '@/stores/settings'

const props = defineProps<{
  projectId: string
  terminalId: string
  height: number
}>()

const emit = defineEmits<{
  (e: 'close'): void
}>()

const terminalsStore = useTerminalsStore()
const settingsStore = useSettingsStore()
const terminalRef = ref<HTMLElement | null>(null)
let terminal: Terminal | null = null
let fitAddon: FitAddon | null = null
let unsubscribe: (() => void) | null = null

const handleResize = useDebounceFn(() => {
  if (fitAddon && terminal) {
    fitAddon.fit()
    const cols = terminal.cols
    const rows = terminal.rows
    terminalsStore.resizeTerminal(props.terminalId, cols, rows)
    terminalsStore.sendToTerminal(props.terminalId, JSON.stringify({ type: 'resize', cols, rows }))
  }
}, 100)

function getTerminalTheme() {
  const themeId = settingsStore.settings?.terminal_theme_id || 'github-dark'
  return settingsStore.getTerminalThemeColors(themeId)
}

function initTerminal() {
  console.log('[TERM] initTerminal:', props.terminalId)
  if (!terminalRef.value) {
    console.log('[TERM] no terminalRef')
    return
  }

  console.log('[TERM] creating xterm for:', props.terminalId)

  const theme = getTerminalTheme()

  terminal = new Terminal({
    cursorBlink: true,
    cursorStyle: 'block',
    fontSize: 14,
    fontFamily: 'Menlo, Monaco, "Courier New", monospace',
    theme: {
      background: theme.background,
      foreground: theme.foreground,
      cursor: theme.foreground,
      selectionBackground: '#264f78',
      black: '#484f58',
      red: '#ff7b72',
      green: '#3fb950',
      yellow: '#d29922',
      blue: '#58a6ff',
      magenta: '#bc8cff',
      cyan: '#39c5cf',
      white: theme.foreground,
      brightBlack: '#6e7681',
      brightRed: '#ff7b72',
      brightGreen: '#3fb950',
      brightYellow: '#d29922',
      brightBlue: '#58a6ff',
      brightMagenta: '#bc8cff',
      brightCyan: '#39c5cf',
      brightWhite: '#ffffff'
    },
    convertEol: true
  })

  fitAddon = new FitAddon()
  terminal.loadAddon(fitAddon)
  terminal.open(terminalRef.value)
  fitAddon.fit()

  terminal.onData((data) => {
    terminalsStore.sendToTerminal(props.terminalId, data)
  })

  terminal.onResize((size) => {
    terminalsStore.resizeTerminal(props.terminalId, size.cols, size.rows)
    terminalsStore.sendToTerminal(props.terminalId, JSON.stringify({ type: 'resize', cols: size.cols, rows: size.rows }))
  })

  window.addEventListener('resize', handleResize)

  console.log('[TERM] subscribing to WS for:', props.terminalId)
  unsubscribe = terminalsStore.onWSMessage(props.terminalId, 'data', (data: string) => {
    console.log('[TERM] received data for', props.terminalId, ':', data.substring(0, 50))
    if (terminal) {
      terminal.write(data)
    }
  })
}

function cleanup() {
  console.log('[TERM] cleanup:', props.terminalId)
  window.removeEventListener('resize', handleResize)

  if (unsubscribe) {
    console.log('[TERM] unsubscribing')
    unsubscribe()
    unsubscribe = null
  }

  if (terminal) {
    console.log('[TERM] disposing terminal')
    terminal.dispose()
    terminal = null
  }

  if (fitAddon) {
    fitAddon = null
  }
}

function refreshTerminal() {
  cleanup()
  setTimeout(() => {
    nextTick(() => {
      if (props.terminalId) {
        initTerminal()
      }
    })
  }, 50)
}

onMounted(() => {
  nextTick(() => {
    initTerminal()
  })
})

onUnmounted(() => {
  cleanup()
})

watch(() => props.terminalId, () => {
  refreshTerminal()
})

watch(() => props.height, () => {
  nextTick(() => {
    fitAddon?.fit()
  })
})

watch(() => settingsStore.settings?.terminal_theme_id, () => {
  if (props.terminalId) {
    refreshTerminal()
  }
})
</script>

<style scoped>
.terminal-pane {
  width: 100%;
  overflow: hidden;
}

.terminal-container {
  width: 100%;
  height: 100%;
  padding: 8px;
}
</style>
