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
  theme?: string
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void
  (e: 'save'): void
}>()

const editorContainer = ref<HTMLElement | null>(null)
const editor = shallowRef<monaco.editor.IStandaloneCodeEditor | null>(null)

onMounted(() => {
  if (!editorContainer.value) return

  const options: monaco.editor.IStandaloneEditorConstructionOptions = {
    value: props.modelValue || '',
    language: props.language,
    theme: props.theme || 'vs-dark',
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

:deep(.monaco-editor .current-line) {
  border: none !important;
  background-color: rgba(255, 255, 255, 0.03) !important;
}

:deep(.monaco-editor .selected-text) {
  background-color: rgba(55, 148, 255, 0.3) !important;
}

:deep(.monaco-editor .margin) {
  background-color: transparent !important;
}

:deep(.monaco-editor .line-numbers) {
  color: #6e7681;
}

:deep(.monaco-editor) {
  background-color: #0d1117;
}

:deep(.monaco-editor-background) {
  background-color: #0d1117;
}

:deep(.monaco-editor .current-line-line-number) {
  color: #e6edf3 !important;
  font-weight: 600 !important;
}

:deep(.monaco-editor .view-overlays .current-line) {
  border: none !important;
}

:deep(.monaco-editor.focused) {
  outline: none !important;
}

:deep(.monaco-editor .focused .selected-text) {
  background-color: rgba(55, 148, 255, 0.4) !important;
}
</style>
