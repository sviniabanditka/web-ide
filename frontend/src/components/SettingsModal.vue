<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { Settings as SettingsIcon, Bot, Palette, Code, Terminal } from 'lucide-vue-next'
import Button from '@/components/ui/Button.vue'
import { useSettingsStore } from '@/stores/settings'
import AIAgentSettings from './settings/AIAgentSettings.vue'
import AppearanceSettings from './settings/AppearanceSettings.vue'
import EditorSettings from './settings/EditorSettings.vue'
import TerminalSettings from './settings/TerminalSettings.vue'

interface Props {
  open?: boolean
}

const props = defineProps<Props>()
const emit = defineEmits<{
  'update:open': [value: boolean]
}>()

const settingsStore = useSettingsStore()

const activeCategory = ref('ai-agent')
const activeSubcategory = ref<string | null>(null)

const categories = [
  { id: 'ai-agent', name: 'AI Agent', icon: Bot, description: 'Configure AI provider settings' },
  { id: 'appearance', name: 'Appearance', icon: Palette, description: 'Customize look and feel' }
]

const subcategories = [
  { id: 'appearance-general', parentId: 'appearance', name: 'General', icon: Palette },
  { id: 'appearance-editor', parentId: 'appearance', name: 'Editor', icon: Code },
  { id: 'appearance-terminal', parentId: 'appearance', name: 'Terminal', icon: Terminal }
]

const isSubcategoryActive = computed(() => {
  const cat = categories.find(c => c.id === activeCategory.value)
  return cat?.id === 'appearance'
})

const currentSubcategories = computed(() => {
  return subcategories.filter(s => s.parentId === activeCategory.value)
})

function selectCategory(catId: string) {
  activeCategory.value = catId
  if (catId !== 'appearance') {
    activeSubcategory.value = null
  } else if (currentSubcategories.value.length > 0) {
    activeSubcategory.value = currentSubcategories.value[0].id
  }
}

function selectSubcategory(subId: string) {
  activeSubcategory.value = subId
}

function close() {
  emit('update:open', false)
}

onMounted(() => {
  settingsStore.fetchSettings()
})
</script>

<template>
  <Teleport to="body">
    <div v-if="open" class="fixed inset-0 z-50 bg-black/50 flex items-center justify-center" @click.self="close">
      <div class="bg-background rounded-lg shadow-lg w-[900px] h-[600px] flex overflow-hidden" role="dialog">
        <aside class="w-64 bg-card border-r flex flex-col">
          <div class="p-4 border-b">
            <h2 class="text-lg font-semibold flex items-center gap-2">
              <SettingsIcon class="w-5 h-5" />
              Settings
            </h2>
          </div>
          <nav class="flex-1 p-2 space-y-1 overflow-y-auto">
            <button
              v-for="cat in categories"
              :key="cat.id"
              @click="selectCategory(cat.id)"
              :class="[
                'w-full flex items-center gap-3 px-3 py-2 rounded-md text-sm transition-colors',
                activeCategory === cat.id
                  ? 'bg-primary text-primary-foreground'
                  : 'hover:bg-accent hover:text-accent-foreground'
              ]"
            >
              <component :is="cat.icon" class="w-4 h-4" />
              <div class="text-left">
                <div class="font-medium">{{ cat.name }}</div>
                <div v-if="cat.description" class="text-xs opacity-70">{{ cat.description }}</div>
              </div>
            </button>
            <div v-if="isSubcategoryActive && currentSubcategories.length > 0" class="ml-4 mt-2 space-y-1 border-l-2 border-border pl-2">
              <button
                v-for="sub in currentSubcategories"
                :key="sub.id"
                @click="selectSubcategory(sub.id)"
                :class="[
                  'w-full flex items-center gap-2 px-2 py-1.5 rounded text-sm transition-colors',
                  activeSubcategory === sub.id
                    ? 'bg-accent text-accent-foreground'
                    : 'hover:bg-accent/50 hover:text-accent-foreground'
                ]"
              >
                <component :is="sub.icon" class="w-3 h-3" />
                {{ sub.name }}
              </button>
            </div>
          </nav>
        </aside>
        <main class="flex-1 flex flex-col overflow-hidden">
          <div class="p-4 border-b flex items-center justify-between">
            <h3 class="text-lg font-medium">
              <template v-if="activeSubcategory">
                {{ subcategories.find(s => s.id === activeSubcategory)?.name }}
              </template>
              <template v-else>
                {{ categories.find(c => c.id === activeCategory)?.name }}
              </template>
            </h3>
            <Button variant="ghost" size="sm" @click="close">Close</Button>
          </div>
          <div class="flex-1 overflow-y-auto p-6">
            <div v-if="settingsStore.loading" class="flex items-center justify-center h-full text-muted-foreground">
              Loading...
            </div>
            <div v-else-if="settingsStore.error" class="flex items-center justify-center h-full text-destructive">
              {{ settingsStore.error }}
            </div>
            <template v-else>
              <AIAgentSettings v-if="activeCategory === 'ai-agent'" />
              <AppearanceSettings v-else-if="activeCategory === 'appearance' && (!activeSubcategory || activeSubcategory === 'appearance-general')" />
              <EditorSettings v-else-if="activeCategory === 'appearance' && activeSubcategory === 'appearance-editor'" />
              <TerminalSettings v-else-if="activeCategory === 'appearance' && activeSubcategory === 'appearance-terminal'" />
            </template>
          </div>
        </main>
      </div>
    </div>
  </Teleport>
</template>
