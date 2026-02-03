<template>
  <div class="git-page">
    <div class="git-header">
      <div class="branch-info">
        <span class="branch-icon">⑂</span>
        <span class="branch-name">{{ gitStore.currentBranch || 'main' }}</span>
      </div>
      <div class="git-actions">
        <button @click="handlePush" :disabled="!hasChanges" class="action-btn push-btn">
          ↑ Push
        </button>
      </div>
    </div>

    <div class="git-content">
      <div class="git-sidebar">
        <div class="section">
          <div class="section-header">
            <span>Staged</span>
            <span class="count">{{ stagedFiles.length }}</span>
          </div>
          <div class="file-list">
            <div
              v-for="file in stagedFiles"
              :key="file.path"
              class="file-item staged"
            >
              <span class="status-icon">{{ getFileIcon(file.status) }}</span>
              <span class="file-path">{{ file.path }}</span>
              <button @click="unstageFile(file.path)" class="file-btn" title="Unstage">−</button>
            </div>
            <div v-if="stagedFiles.length === 0" class="empty-state">No staged files</div>
          </div>
        </div>

        <div class="section">
          <div class="section-header">
            <span>Changed</span>
            <span class="count">{{ changedFiles.length }}</span>
          </div>
          <div class="file-list">
            <div
              v-for="file in changedFiles"
              :key="file.path"
              class="file-item changed"
            >
              <span class="status-icon">{{ getFileIcon(file.status) }}</span>
              <span class="file-path">{{ file.path }}</span>
              <button @click="stageFile(file.path)" class="file-btn" title="Stage">+</button>
            </div>
            <div v-if="changedFiles.length === 0" class="empty-state">No changes</div>
          </div>
        </div>

        <div class="section">
          <div class="section-header">
            <span>Untracked</span>
            <span class="count">{{ untrackedFiles.length }}</span>
          </div>
          <div class="file-list">
            <div
              v-for="file in untrackedFiles"
              :key="file"
              class="file-item untracked"
            >
              <span class="status-icon">?</span>
              <span class="file-path">{{ file }}</span>
              <button @click="stageUntracked(file)" class="file-btn" title="Stage">+</button>
            </div>
            <div v-if="untrackedFiles.length === 0" class="empty-state">No untracked</div>
          </div>
        </div>

        <div class="section-actions">
          <button @click="stageAll" :disabled="!hasChanges" class="btn">Stage All</button>
          <button @click="unstageAll" :disabled="stagedFiles.length === 0" class="btn">Unstage All</button>
        </div>
      </div>

      <div class="git-main">
        <div class="commit-section">
          <div class="commit-header">Commit Changes</div>
          <textarea
            v-model="commitMessage"
            placeholder="Enter commit message..."
            class="commit-message"
            rows="3"
          ></textarea>
          <button
            @click="handleCommit"
            :disabled="stagedFiles.length === 0 || !commitMessage.trim()"
            class="commit-btn"
          >
            Commit
          </button>
        </div>

        <div class="diff-section">
          <div class="section-header">
            <span>Diff</span>
            <div class="diff-tabs">
              <button
                :class="{ active: !showCached }"
                @click="showCached = false"
              >Working Tree</button>
              <button
                :class="{ active: showCached }"
                @click="showCached = true"
              >Staged</button>
            </div>
          </div>
          <div class="diff-content">
            <pre v-if="diffContent" class="diff-view">{{ diffContent }}</pre>
            <div v-else class="empty-state">No changes to show</div>
          </div>
        </div>

        <div class="log-section">
          <div class="section-header">
            <span>Recent Commits</span>
            <button @click="refreshLog" class="refresh-btn">↻</button>
          </div>
          <div class="log-list">
            <div
              v-for="entry in gitStore.log"
              :key="entry.hash"
              class="log-entry"
            >
              <div class="log-hash">{{ entry.hash.substring(0, 7) }}</div>
              <div class="log-message">{{ entry.subject }}</div>
              <div class="log-meta">{{ entry.author }} · {{ formatDate(entry.date) }}</div>
            </div>
            <div v-if="gitStore.log.length === 0" class="empty-state">No commits yet</div>
          </div>
        </div>
      </div>
    </div>

    <div v-if="gitStore.error" class="error-toast">
      {{ gitStore.error }}
      <button @click="gitStore.error = null">×</button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useGitStore } from '../stores/git'

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

<style scoped>
.git-page {
  height: 100%;
  display: flex;
  flex-direction: column;
  background: #1e1e1e;
  color: #d4d4d4;
}

.git-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  background: #252526;
  border-bottom: 1px solid #3c3c3c;
}

.branch-info {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
}

.branch-icon {
  color: #f44747;
}

.branch-name {
  font-weight: 500;
}

.action-btn {
  padding: 6px 12px;
  background: #0e639c;
  border: none;
  border-radius: 4px;
  color: #fff;
  font-size: 12px;
}

.action-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.git-content {
  flex: 1;
  display: flex;
  overflow: hidden;
}

.git-sidebar {
  width: 280px;
  background: #252526;
  border-right: 1px solid #3c3c3c;
  overflow-y: auto;
  padding: 8px;
}

.section {
  margin-bottom: 16px;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 4px;
  font-size: 12px;
  font-weight: 500;
  color: #888;
  text-transform: uppercase;
}

.count {
  background: #3c3c3c;
  padding: 2px 6px;
  border-radius: 10px;
  font-size: 11px;
}

.file-list {
  display: flex;
  flex-direction: column;
}

.file-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 8px;
  font-size: 13px;
  cursor: pointer;
  border-radius: 4px;
}

.file-item:hover {
  background: #2d2d30;
}

.status-icon {
  width: 16px;
  text-align: center;
  font-size: 11px;
  font-weight: bold;
}

.file-item.staged .status-icon { color: #4ec9b0; }
.file-item.changed .status-icon { color: #dcdcaa; }
.file-item.untracked .status-icon { color: #ce9178; }

.file-path {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.file-btn {
  background: none;
  border: none;
  color: #888;
  font-size: 16px;
  padding: 0 4px;
  opacity: 0;
  transition: opacity 0.2s;
}

.file-item:hover .file-btn {
  opacity: 1;
}

.file-btn:hover {
  color: #fff;
}

.section-actions {
  display: flex;
  gap: 8px;
  padding: 8px 4px;
}

.section-actions .btn {
  flex: 1;
  padding: 6px;
  background: #3c3c3c;
  border: none;
  border-radius: 4px;
  color: #ccc;
  font-size: 12px;
}

.section-actions .btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.git-main {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.commit-section {
  padding: 16px;
  border-bottom: 1px solid #3c3c3c;
}

.commit-header {
  font-size: 12px;
  font-weight: 500;
  color: #888;
  margin-bottom: 8px;
}

.commit-message {
  width: 100%;
  padding: 8px;
  background: #1e1e1e;
  border: 1px solid #3c3c3c;
  border-radius: 4px;
  color: #d4d4d4;
  font-size: 13px;
  resize: none;
}

.commit-message:focus {
  outline: none;
  border-color: #0e639c;
}

.commit-btn {
  margin-top: 8px;
  padding: 8px 16px;
  background: #0e639c;
  border: none;
  border-radius: 4px;
  color: #fff;
  font-size: 13px;
}

.commit-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.diff-section {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  border-bottom: 1px solid #3c3c3c;
}

.diff-section .section-header {
  padding: 12px 16px;
}

.diff-tabs {
  display: flex;
  gap: 4px;
}

.diff-tabs button {
  padding: 4px 12px;
  background: transparent;
  border: none;
  border-radius: 4px;
  color: #888;
  font-size: 12px;
}

.diff-tabs button.active {
  background: #3c3c3c;
  color: #fff;
}

.diff-content {
  flex: 1;
  overflow: auto;
  padding: 0 16px 16px;
}

.diff-view {
  font-family: 'Menlo', 'Monaco', monospace;
  font-size: 12px;
  white-space: pre-wrap;
  color: #d4d4d4;
}

.diff-view .add { color: #4ec9b0; }
.diff-view .del { color: #ce9178; }

.log-section {
  height: 200px;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.log-section .section-header {
  padding: 12px 16px;
}

.refresh-btn {
  background: none;
  border: none;
  color: #888;
  font-size: 14px;
  padding: 0 4px;
}

.refresh-btn:hover {
  color: #fff;
}

.log-list {
  flex: 1;
  overflow-y: auto;
  padding: 0 16px 16px;
}

.log-entry {
  padding: 8px 0;
  border-bottom: 1px solid #2d2d30;
}

.log-hash {
  font-family: 'Menlo', 'Monaco', monospace;
  font-size: 12px;
  color: #4fc3f7;
  margin-bottom: 2px;
}

.log-message {
  font-size: 13px;
  margin-bottom: 2px;
}

.log-meta {
  font-size: 11px;
  color: #888;
}

.empty-state {
  padding: 16px;
  text-align: center;
  color: #666;
  font-size: 13px;
}

.error-toast {
  position: fixed;
  bottom: 20px;
  right: 20px;
  padding: 12px 16px;
  background: #f44336;
  color: #fff;
  border-radius: 4px;
  font-size: 13px;
  display: flex;
  align-items: center;
  gap: 12px;
}

.error-toast button {
  background: none;
  border: none;
  color: #fff;
  font-size: 16px;
  padding: 0;
}
</style>
