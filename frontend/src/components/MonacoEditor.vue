<template>
  <div ref="editorContainer" class="monaco-container"></div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch, shallowRef } from 'vue'
import * as monaco from 'monaco-editor'
import editorWorker from 'monaco-editor/esm/vs/editor/editor.worker?worker'
import jsonWorker from 'monaco-editor/esm/vs/language/json/json.worker?worker'
import cssWorker from 'monaco-editor/esm/vs/language/css/css.worker?worker'
import htmlWorker from 'monaco-editor/esm/vs/language/html/html.worker?worker'
import tsWorker from 'monaco-editor/esm/vs/language/typescript/ts.worker?worker'
import { useSettingsStore } from '@/stores/settings'

self.MonacoEnvironment = {
  getWorker: function (_workerId: string, label: string): Worker {
    switch (label) {
      case 'json':
        return new jsonWorker()
      case 'css':
      case 'scss':
      case 'less':
        return new cssWorker()
      case 'html':
      case 'handlebars':
      case 'razor':
        return new htmlWorker()
      case 'typescript':
      case 'javascript':
        return new tsWorker()
      default:
        return new editorWorker()
    }
  }
}

const props = defineProps<{
  modelValue?: string
  language: string
  path?: string
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void
  (e: 'save'): void
}>()

const settingsStore = useSettingsStore()
const editorContainer = ref<HTMLElement | null>(null)
const editor = shallowRef<monaco.editor.IStandaloneCodeEditor | null>(null)

function getCurrentTheme(): string {
  if (settingsStore.settings?.editor_theme_id) {
    return settingsStore.settings.editor_theme_id
  }
  return 'vs-dark'
}

onMounted(() => {
  if (!editorContainer.value) return

  const options: monaco.editor.IStandaloneEditorConstructionOptions = {
    value: props.modelValue || '',
    language: props.language,
    theme: getCurrentTheme(),
    automaticLayout: true,
    minimap: { enabled: true },
    fontSize: 14,
    fontFamily: "'Menlo', 'Monaco', 'Courier New', monospace",
    scrollBeyondLastLine: false,
    renderWhitespace: 'selection',
    tabSize: 2,
    wordWrap: 'on',
    renderLineHighlight: 'line',
    overviewRulerBorder: false
  }

  if (props.path) {
    const uri = monaco.Uri.parse(`file://${props.path}`)
    const model = monaco.editor.getModel(uri)
    if (model) {
      options.model = model
    } else {
      options.model = monaco.editor.createModel(props.modelValue || '', props.language, uri)
    }
  }

  editor.value = monaco.editor.create(editorContainer.value, options)

  editor.value.onDidChangeModelContent(() => {
    const value = editor.value?.getValue() || ''
    emit('update:modelValue', value)
  })

  editor.value.addCommand(monaco.KeyMod.CtrlCmd | monaco.KeyCode.KeyS, () => {
    emit('save')
  })
})

watch(() => props.modelValue, (newValue) => {
  if (editor.value && newValue !== undefined && editor.value.getValue() !== newValue) {
    editor.value.setValue(newValue)
  }
})

watch(() => props.language, (newLang) => {
  if (editor.value) {
    const model = editor.value.getModel()
    if (model) {
      monaco.editor.setModelLanguage(model, newLang)
    }
  }
})

watch(() => settingsStore.settings?.editor_theme_id, (newTheme) => {
  if (editor.value && newTheme) {
    monaco.editor.setTheme(newTheme)
  }
})

onUnmounted(() => {
  if (editor.value) {
    editor.value.dispose()
  }
})
</script>

<style scoped>
.monaco-container {
  width: 100%;
  height: 100%;
}
</style>
