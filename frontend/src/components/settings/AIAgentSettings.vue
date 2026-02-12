<script setup lang="ts">
import { ref, watch } from 'vue'
import Input from '@/components/ui/Input.vue'
import Label from '@/components/ui/Label.vue'
import Button from '@/components/ui/Button.vue'
import { useSettingsStore } from '@/stores/settings'

const settingsStore = useSettingsStore()

const form = ref({
  ai_provider: 'anthropic',
  ai_base_url: '',
  ai_api_key: '',
  ai_model: 'claude-sonnet-4-20250514'
})

watch(
  () => settingsStore.settings,
  (settings) => {
    if (settings) {
      form.value.ai_provider = settings.ai_provider
      form.value.ai_base_url = settings.ai_base_url
      form.value.ai_api_key = settings.ai_api_key
      form.value.ai_model = settings.ai_model
    }
  },
  { immediate: true }
)

const providers = [
  { id: 'anthropic', name: 'Anthropic' },
  { id: 'openai', name: 'OpenAI' }
]

const models = [
  { id: 'claude-sonnet-4-20250514', name: 'Claude Sonnet 4 (Latest)' },
  { id: 'claude-opus-4-20250514', name: 'Claude Opus 4 (Latest)' },
  { id: 'claude-haiku-4-20250514', name: 'Claude Haiku 4 (Latest)' },
  { id: 'gpt-4o', name: 'GPT-4o' },
  { id: 'gpt-4-turbo', name: 'GPT-4 Turbo' },
  { id: 'gpt-3.5-turbo', name: 'GPT-3.5 Turbo' }
]

const saving = ref(false)

async function save() {
  saving.value = true
  await settingsStore.saveSettings(form.value)
  saving.value = false
}
</script>

<template>
  <div class="space-y-6 max-w-xl">
    <div class="space-y-4">
      <div>
        <Label for="provider">AI Provider</Label>
        <select
          id="provider"
          v-model="form.ai_provider"
          class="flex h-10 w-full items-center justify-between rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50 mt-1"
        >
          <option v-for="provider in providers" :key="provider.id" :value="provider.id">
            {{ provider.name }}
          </option>
        </select>
      </div>

      <div>
        <Label for="baseUrl">Base URL</Label>
        <Input
          id="baseUrl"
          v-model="form.ai_base_url"
          placeholder="https://api.anthropic.com"
          class="mt-1"
        />
        <p class="text-xs text-muted-foreground mt-1">
          Leave empty to use default API endpoint
        </p>
      </div>

      <div>
        <Label for="apiKey">API Key</Label>
        <Input
          id="apiKey"
          v-model="form.ai_api_key"
          type="password"
          placeholder="sk-ant-api03-..."
          class="mt-1"
        />
      </div>

      <div>
        <Label for="model">Model</Label>
        <select
          id="model"
          v-model="form.ai_model"
          class="flex h-10 w-full items-center justify-between rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50 mt-1"
        >
          <option v-for="model in models" :key="model.id" :value="model.id">
            {{ model.name }}
          </option>
        </select>
      </div>
    </div>

    <div class="flex justify-end pt-4 border-t">
      <Button @click="save" :disabled="saving">
        {{ saving ? 'Saving...' : 'Save Changes' }}
      </Button>
    </div>
  </div>
</template>
