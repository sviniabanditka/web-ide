<template>
  <div class="h-full flex flex-col bg-background">
    <div class="flex items-center justify-between px-4 py-3 bg-card border-b">
      <div class="flex items-center gap-2">
        <span class="text-lg">⑂</span>
        <span class="font-medium">{{ gitStore.currentBranch || 'main' }}</span>
      </div>
      <Button size="sm" @click="handlePush" :disabled="!hasChanges">↑ Push</Button>
    </div>

    <div class="flex-1 flex overflow-hidden">
      <aside class="w-[280px] bg-card border-r flex flex-col overflow-hidden">
        <ScrollArea class="flex-1 p-3">
          <div class="space-y-4">
            <div>
              <div class="flex items-center justify-between px-1 py-2">
                <span class="text-xs font-medium text-muted-foreground uppercase">Staged</span>
                <Badge variant="secondary">{{ stagedFiles.length }}</Badge>
              </div>
              <div class="space-y-1">
                <div
                  v-for="file in stagedFiles"
                  :key="file.path"
                  class="flex items-center gap-2 px-2 py-1 text-sm rounded hover:bg-accent cursor-pointer"
                >
                  <span class="text-cyan-500 font-mono text-xs w-4 text-center">{{ getFileIcon(file.status) }}</span>
                  <span class="truncate flex-1">{{ file.path }}</span>
                  <button @click="unstageFile(file.path)" class="text-muted-foreground hover:text-foreground" title="Unstage">−</button>
                </div>
                <div v-if="stagedFiles.length === 0" class="px-2 py-2 text-sm text-muted-foreground text-center">No staged files</div>
              </div>
            </div>

            <div>
              <div class="flex items-center justify-between px-1 py-2">
                <span class="text-xs font-medium text-muted-foreground uppercase">Changed</span>
                <Badge variant="secondary">{{ changedFiles.length }}</Badge>
              </div>
              <div class="space-y-1">
                <div
                  v-for="file in changedFiles"
                  :key="file.path"
                  class="flex items-center gap-2 px-2 py-1 text-sm rounded hover:bg-accent cursor-pointer"
                >
                  <span class="text-yellow-500 font-mono text-xs w-4 text-center">{{ getFileIcon(file.status) }}</span>
                  <span class="truncate flex-1">{{ file.path }}</span>
                  <button @click="stageFile(file.path)" class="text-muted-foreground hover:text-foreground" title="Stage">+</button>
                </div>
                <div v-if="changedFiles.length === 0" class="px-2 py-2 text-sm text-muted-foreground text-center">No changes</div>
              </div>
            </div>

            <div>
              <div class="flex items-center justify-between px-1 py-2">
                <span class="text-xs font-medium text-muted-foreground uppercase">Untracked</span>
                <Badge variant="secondary">{{ untrackedFiles.length }}</Badge>
              </div>
              <div class="space-y-1">
                <div
                  v-for="file in untrackedFiles"
                  :key="file"
                  class="flex items-center gap-2 px-2 py-1 text-sm rounded hover:bg-accent cursor-pointer"
                >
                  <span class="text-orange-500 font-mono text-xs w-4 text-center">?</span>
                  <span class="truncate flex-1">{{ file }}</span>
                  <button @click="stageUntracked(file)" class="text-muted-foreground hover:text-foreground" title="Stage">+</button>
                </div>
                <div v-if="untrackedFiles.length === 0" class="px-2 py-2 text-sm text-muted-foreground text-center">No untracked</div>
              </div>
            </div>
          </div>
        </ScrollArea>
        <div class="p-3 border-t flex gap-2">
          <Button variant="secondary" size="sm" class="flex-1" @click="stageAll" :disabled="!hasChanges">Stage All</Button>
          <Button variant="secondary" size="sm" class="flex-1" @click="unstageAll" :disabled="stagedFiles.length === 0">Unstage All</Button>
        </div>
      </aside>

      <main class="flex-1 flex flex-col overflow-hidden">
        <div class="p-4 border-b space-y-2">
          <Label class="text-xs font-medium text-muted-foreground">Commit Changes</Label>
          <Textarea
            v-model="commitMessage"
            placeholder="Enter commit message..."
            rows="3"
            class="font-mono text-sm"
          />
          <div class="flex justify-end">
            <Button @click="handleCommit" :disabled="stagedFiles.length === 0 || !commitMessage.trim()">Commit</Button>
          </div>
        </div>

        <div class="flex-1 flex flex-col overflow-hidden border-b">
          <div class="flex items-center justify-between px-4 py-2 border-b bg-muted/30">
            <span class="text-xs font-medium text-muted-foreground">Diff</span>
            <div class="flex gap-1">
              <Button
                size="sm"
                :variant="!showCached ? 'secondary' : 'ghost'"
                @click="showCached = false"
              >Working Tree</Button>
              <Button
                size="sm"
                :variant="showCached ? 'secondary' : 'ghost'"
                @click="showCached = true"
              >Staged</Button>
            </div>
          </div>
          <ScrollArea class="flex-1">
            <pre v-if="diffContent" class="p-4 font-mono text-sm whitespace-pre-wrap">{{ diffContent }}</pre>
            <div v-else class="p-4 text-center text-muted-foreground text-sm">No changes to show</div>
          </ScrollArea>
        </div>

        <div class="h-[200px] flex flex-col overflow-hidden">
          <div class="flex items-center justify-between px-4 py-2 border-b bg-muted/30">
            <span class="text-xs font-medium text-muted-foreground">Recent Commits</span>
            <Button variant="ghost" size="sm" @click="refreshLog">↻</Button>
          </div>
          <ScrollArea class="flex-1">
            <div
              v-for="entry in gitStore.log"
              :key="entry.hash"
              class="px-4 py-2 border-b"
            >
              <div class="font-mono text-sm text-primary">{{ entry.hash.substring(0, 7) }}</div>
              <div class="text-sm">{{ entry.subject }}</div>
              <div class="text-xs text-muted-foreground">{{ entry.author }} · {{ formatDate(entry.date) }}</div>
            </div>
            <div v-if="gitStore.log.length === 0" class="p-4 text-center text-muted-foreground text-sm">No commits yet</div>
          </ScrollArea>
        </div>
      </main>
    </div>

    <div v-if="gitStore.error" class="fixed bottom-4 right-4 bg-destructive text-destructive-foreground px-4 py-2 rounded-md text-sm flex items-center gap-2">
      {{ gitStore.error }}
      <button @click="gitStore.error = null">×</button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useGitStore } from '../stores/git'
import Button from '@/components/ui/Button.vue'
import Textarea from '@/components/ui/Textarea.vue'
import Label from '@/components/ui/Label.vue'
import Badge from '@/components/ui/Badge.vue'
import ScrollArea from '@/components/ui/ScrollArea.vue'

const route = useRoute()
const gitStore = useGitStore()

const commitMessage = ref('')
const showCached = ref(false)

const stagedFiles = computed(() => gitStore.status?.status?.staged || [])
const changedFiles = computed(() => gitStore.status?.status?.changed || [])
const untrackedFiles = computed(() => gitStore.status?.status?.untracked || [])
const hasChanges = computed(() =>
  stagedFiles.value.length > 0 ||
  changedFiles.value.length > 0 ||
  untrackedFiles.value.length > 0
)
const diffContent = computed(() => gitStore.diff)

function getFileIcon(status: string): string {
  switch (status) {
    case 'M': return 'M'
    case 'A': return '+'
    case 'D': return '−'
    case '?': return '?'
    default: return status
  }
}

async function stageFile(path: string) {
  const projectId = route.params.id as string
  await gitStore.stageFiles(projectId, [path])
  await gitStore.fetchDiff(projectId, showCached.value)
}

async function unstageFile(path: string) {
  const projectId = route.params.id as string
  await gitStore.unstageFiles(projectId, [path])
  await gitStore.fetchDiff(projectId, showCached.value)
}

async function stageUntracked(path: string) {
  const projectId = route.params.id as string
  await gitStore.stageFiles(projectId, [path])
}

function stageAll() {
  const projectId = route.params.id as string
  gitStore.stageAll(projectId)
}

function unstageAll() {
  const projectId = route.params.id as string
  gitStore.unstageAll(projectId)
}

async function handleCommit() {
  const projectId = route.params.id as string
  try {
    await gitStore.commit(projectId, commitMessage.value)
    commitMessage.value = ''
  } catch (e) {
    // Error shown in toast
  }
}

async function handlePush() {
  const projectId = route.params.id as string
  try {
    await gitStore.push(projectId)
  } catch (e) {
    // Error shown in toast
  }
}

function refreshLog() {
  const projectId = route.params.id as string
  gitStore.fetchLog(projectId)
}

function formatDate(dateStr: string): string {
  try {
    const date = new Date(dateStr)
    return date.toLocaleDateString() + ' ' + date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
  } catch {
    return dateStr
  }
}

async function loadData() {
  const projectId = route.params.id as string
  await Promise.all([
    gitStore.fetchStatus(projectId),
    gitStore.fetchDiff(projectId, showCached.value),
    gitStore.fetchLog(projectId)
  ])
}

watch(showCached, async (cached) => {
  const projectId = route.params.id as string
  await gitStore.fetchDiff(projectId, cached)
})

onMounted(() => {
  loadData()
})
</script>
