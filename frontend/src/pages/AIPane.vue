<template>
  <div class="ai-pane">
    <div class="ai-sidebar">
      <div class="chat-list-header">
        <button class="new-chat-btn" @click="createNewChat">New Chat</button>
      </div>
      <div class="chat-list">
        <div
          v-for="chat in aiStore.chats"
          :key="chat.id"
          class="chat-item"
          :class="{ active: aiStore.activeChat?.id === chat.id }"
          @click="selectChat(chat)"
        >
          <span class="chat-title">{{ chat.title || 'Untitled' }}</span>
          <button class="delete-chat-btn" @click.stop="deleteChat(chat.id)">×</button>
        </div>
        <div v-if="aiStore.chats.length === 0" class="empty-chat-list">
          No chats yet
        </div>
      </div>
    </div>

    <div class="ai-main">
      <div v-if="aiStore.activeChat" class="chat-container">
        <div class="chat-header">
          <h3>{{ aiStore.activeChat.title }}</h3>
        </div>
        <div class="chat-messages" ref="chatContainer">
          <template v-for="msg in sortedMessages" :key="msg.id">
            <div v-if="msg.role === 'thinking' && msg.content" class="thinking-message">
              <ThinkingBlock :thinking="msg.content" />
            </div>
            <div
              v-else-if="msg.role === 'assistant' && (msg.content || msg.tool_calls?.length)"
              class="chat-message assistant"
              :class="{ streaming: aiStore.streamingMessageId === msg.id && aiStore.isStreaming }"
            >
              <div v-if="msg.content" class="message-role">AI</div>
              <div v-if="msg.content" class="message-content markdown-content">
                <div v-html="msg.parsedContent || ''"></div>
                <span v-if="aiStore.streamingMessageId === msg.id && aiStore.isStreaming" class="cursor"></span>
              </div>
              <div v-if="msg.tool_calls?.length" class="message-tool-calls">
                <ToolBlock
                  v-for="tool in msg.tool_calls"
                  :key="tool.id"
                  :tool="tool"
                  :result="msg.tool_results?.find(r => r.id === tool.id)"
                />
              </div>
            </div>
            <div
              v-else-if="msg.role === 'user'"
              class="chat-message user"
            >
              <div class="message-role">You</div>
              <div class="message-content">{{ msg.content }}</div>
            </div>
          </template>
        </div>
        <div class="chat-input">
          <textarea
            v-model="userMessage"
            placeholder="Describe what you want to do..."
            @keydown.ctrl.enter="sendMessage"
            :disabled="aiStore.isStreaming"
          ></textarea>
          <div class="input-actions">
            <div v-if="aiStore.modelStatus !== 'idle'" class="model-status" :class="aiStore.modelStatus">
              <span class="status-indicator"></span>
              <span class="status-text">{{ getStatusText(aiStore.modelStatus) }}</span>
            </div>
            <UsageRing />
            <button @click="stopStreaming" v-if="aiStore.isStreaming" class="stop-btn">
              Stop
            </button>
            <button @click="sendMessage" :disabled="aiStore.isStreaming || !userMessage.trim()" class="send-btn">
              Send
            </button>
          </div>
        </div>
      </div>

      <div v-else class="no-chat-selected">
        <div class="no-chat-content">
          <h3>AI Chat</h3>
          <p>Select a chat or create a new one to start coding with AI</p>
          <button class="new-chat-btn-large" @click="createNewChat">New Chat</button>
        </div>
      </div>
    </div>

    <div class="ai-changes-panel" v-if="aiStore.activeChat">
      <div class="panel-header">
        <h4>Changes</h4>
      </div>
      <div class="changes-list">
        <div
          v-for="cs in chatChangeSets"
          :key="cs.id"
          class="change-item"
          @click="selectChangeSet(cs)"
        >
          <div class="change-status" :class="cs.status">{{ cs.status }}</div>
          <div class="change-info">
            <div class="change-title">{{ cs.title || 'No title' }}</div>
            <div class="change-date">{{ formatDate(cs.created_at) }}</div>
          </div>
        </div>
        <div v-if="chatChangeSets.length === 0" class="empty-changes">
          No changes yet
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch, nextTick, computed } from 'vue'
import { useAIStore, type Chat, type ChatChangeSet } from '../stores/ai'
import UsageRing from '../components/UsageRing.vue'
import ToolBlock from '../components/ai/ToolBlock.vue'
import ThinkingBlock from '../components/ai/ThinkingBlock.vue'

interface Project {
  id: string
  name: string
  root_path: string
}

const props = defineProps<{
  project: Project
}>()

const aiStore = useAIStore()
const userMessage = ref('')
const chatContainer = ref<HTMLElement | null>(null)
const chatChangeSets = ref<ChatChangeSet[]>([])

const sortedMessages = computed(() => {
  return [...aiStore.chatMessages]
})

function scrollToBottom() {
  nextTick(() => {
    if (chatContainer.value) {
      chatContainer.value.scrollTop = chatContainer.value.scrollHeight
    }
  })
}

watch(() => aiStore.chatMessages.length, scrollToBottom)
watch(() => aiStore.chatMessages, scrollToBottom, { deep: true })

async function createNewChat() {
  const chat = await aiStore.createChat(props.project.id, 'New Chat')
  if (chat) {
    await aiStore.selectChat(chat)
    aiStore.connectChatWS(chat.id)
  }
}

async function selectChat(chat: Chat) {
  await aiStore.selectChat(chat)
  aiStore.connectChatWS(chat.id)
  chatChangeSets.value = await aiStore.fetchChatChangeSets(chat.id)
}

async function deleteChat(chatId: string) {
  if (confirm('Delete this chat?')) {
    await aiStore.deleteChat(chatId)
  }
}

async function sendMessage() {
  if (!userMessage.value.trim() || aiStore.isStreaming) return

  const content = userMessage.value
  userMessage.value = ''

  await aiStore.sendChatMessage(content)
}

function stopStreaming() {
  aiStore.stopStreaming()
}

async function selectChangeSet(cs: ChatChangeSet) {
  console.log('Selected changeset:', cs)
}

function formatDate(dateStr: string): string {
  if (!dateStr) return ''
  const date = new Date(dateStr)
  return date.toLocaleString()
}

function getStatusText(status: string): string {
  const statusMap: Record<string, string> = {
    thinking: 'Thinking...',
    using_tool: 'Using tool...',
    editing: 'Making edits...',
    planning: 'Planning...'
  }
  return statusMap[status] || status
}

watch(() => aiStore.chatMessages.length, scrollToBottom)
watch(() => aiStore.chatMessages, scrollToBottom, { deep: true })
watch(() => aiStore.streamingContent, scrollToBottom)

onMounted(async () => {
  await aiStore.fetchChats(props.project.id)
})

onUnmounted(() => {
  if (aiStore.chatWs) {
    aiStore.chatWs.close()
  }
})
</script>

<style scoped>
.ai-pane {
  height: 100%;
  display: flex;
  background: #1e1e1e;
}

.ai-sidebar {
  width: 220px;
  background: #252526;
  border-right: 1px solid #3c3c3c;
  display: flex;
  flex-direction: column;
}

.chat-list-header {
  padding: 12px;
  border-bottom: 1px solid #3c3c3c;
}

.new-chat-btn {
  width: 100%;
  padding: 8px 12px;
  background: #0e639c;
  border: none;
  border-radius: 6px;
  color: #fff;
  font-size: 13px;
  cursor: pointer;
}

.new-chat-btn:hover {
  background: #1177bb;
}

.chat-list {
  flex: 1;
  overflow-y: auto;
  padding: 8px;
}

.chat-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 12px;
  border-radius: 6px;
  cursor: pointer;
  transition: background 0.2s;
  margin-bottom: 4px;
}

.chat-item:hover {
  background: #2d2d30;
}

.chat-item.active {
  background: #37373d;
}

.chat-title {
  font-size: 13px;
  color: #ccc;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.delete-chat-btn {
  background: none;
  border: none;
  color: #888;
  font-size: 16px;
  cursor: pointer;
  opacity: 0;
  transition: opacity 0.2s;
}

.chat-item:hover .delete-chat-btn {
  opacity: 1;
}

.delete-chat-btn:hover {
  color: #c62828;
}

.empty-chat-list {
  padding: 20px;
  text-align: center;
  color: #666;
  font-size: 13px;
}

.ai-main {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.chat-container {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0;
}

.chat-header {
  flex-shrink: 0;
  padding: 12px 16px;
  border-bottom: 1px solid #3c3c3c;
}

.chat-header h3 {
  font-size: 14px;
  font-weight: 500;
  color: #fff;
  margin: 0;
}

.chat-messages {
  flex: 1;
  overflow-y: auto;
  padding: 16px;
  min-height: 0;
}

.chat-message {
  margin-bottom: 16px;
  max-width: 80%;
}

.chat-message.user {
  margin-left: auto;
}

.chat-message.assistant {
  margin-right: auto;
}

.message-role {
  font-size: 11px;
  color: #888;
  margin-bottom: 4px;
}

.chat-message.tool .message-role {
  color: #10b981;
}

.message-content {
  padding: 10px 14px;
  border-radius: 8px;
  font-size: 13px;
  line-height: 1.5;
  white-space: pre-wrap;
}

.chat-message.user .message-content {
  background: #0e639c;
  color: #fff;
}

.chat-message.assistant .message-content {
  background: #2d2d30;
  color: #ccc;
}

.chat-message.tool .message-content {
  background: #1a2e1a;
  border: 1px solid #10b981;
  color: #a0c0a0;
}

.chat-message.streaming .message-content {
  background: #2d2d30;
}

.cursor {
  display: inline-block;
  width: 2px;
  height: 14px;
  background: #888;
  margin-left: 2px;
  vertical-align: middle;
  animation: blink 1s infinite;
}

@keyframes blink {
  0%, 50% { opacity: 1; }
  51%, 100% { opacity: 0; }
}

.markdown-content {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, sans-serif;
}

.markdown-content code {
  background: #1a1a1a;
  padding: 2px 6px;
  border-radius: 4px;
  font-family: 'SF Mono', Monaco, Consolas, monospace;
  font-size: 12px;
}

.markdown-content pre {
  background: #1a1a1a;
  padding: 12px;
  border-radius: 6px;
  overflow-x: auto;
  margin: 8px 0;
}

.markdown-content pre code {
  background: none;
  padding: 0;
}

.markdown-content p {
  margin: 8px 0;
}

.markdown-content p:first-child {
  margin-top: 0;
}

.markdown-content p:last-child {
  margin-bottom: 0;
}

.markdown-content ul,
.markdown-content ol {
  margin: 8px 0;
  padding-left: 32px;
}

.markdown-content li {
  margin: 4px 0;
}

.markdown-content table {
  width: 100%;
  border-collapse: collapse;
  margin: 8px 0;
}

.markdown-content th,
.markdown-content td {
  border: 1px solid #3c3c3c;
  padding: 8px 12px;
  text-align: left;
}

.markdown-content th {
  background: #2d2d30;
  font-weight: 600;
}

.markdown-content tr:nth-child(even) {
  background: #252526;
}

.markdown-content a {
  color: #6bb8ff;
  text-decoration: none;
}

.markdown-content a:hover {
  text-decoration: underline;
}

.markdown-content blockquote {
  border-left: 3px solid #0e639c;
  margin: 8px 0;
  padding-left: 16px;
  color: #888;
}

.markdown-content blockquote {
  border-left: 3px solid #0e639c;
  margin: 8px 0;
  padding-left: 16px;
  color: #888;
}

.chat-input {
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 12px 16px;
  border-top: 1px solid #3c3c3c;
  background: #252526;
}

.chat-input textarea {
  width: 100%;
  padding: 10px 12px;
  background: #1e1e1e;
  border: 1px solid #3c3c3c;
  border-radius: 6px;
  color: #ccc;
  font-size: 13px;
  resize: none;
  height: 80px;
  font-family: inherit;
}

.chat-input textarea:focus {
  outline: none;
  border-color: #0e639c;
}

.input-actions {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 8px;
}

.send-btn,
.stop-btn {
  padding: 8px 16px;
  border: none;
  border-radius: 6px;
  font-size: 13px;
  cursor: pointer;
}

.send-btn {
  background: #0e639c;
  color: #fff;
}

.send-btn:disabled {
  background: #3c3c3c;
  color: #888;
  cursor: not-allowed;
}

.stop-btn {
  background: #c62828;
  color: #fff;
}

.stop-btn:hover {
  background: #d32f2f;
}

.model-status {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  border-radius: 6px;
  font-size: 12px;
  margin-right: auto;
}

.status-indicator {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  animation: pulse 1.5s infinite;
}

.model-status.thinking .status-indicator {
  background: #f59e0b;
}

.model-status.using_tool .status-indicator {
  background: #3b82f6;
}

.model-status.editing .status-indicator {
  background: #10b981;
}

.model-status.planning .status-indicator {
  background: #8b5cf6;
}

.status-text {
  color: #888;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

.message-tool-calls,
.message-tool-results {
  margin-top: 8px;
}

.thinking-message {
  margin-right: auto;
  max-width: 80%;
  margin-top: 8px;
  margin-bottom: 8px;
}

.no-chat-selected {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
}

.no-chat-content {
  text-align: center;
  color: #888;
}

.no-chat-content h3 {
  font-size: 24px;
  color: #fff;
  margin-bottom: 8px;
}

.no-chat-content p {
  font-size: 14px;
  margin-bottom: 20px;
}

.new-chat-btn-large {
  padding: 12px 24px;
  background: #0e639c;
  border: none;
  border-radius: 6px;
  color: #fff;
  font-size: 14px;
  cursor: pointer;
}

.new-chat-btn-large:hover {
  background: #1177bb;
}

.ai-changes-panel {
  width: 280px;
  background: #252526;
  border-left: 1px solid #3c3c3c;
  display: flex;
  flex-direction: column;
}

.panel-header {
  padding: 12px 16px;
  border-bottom: 1px solid #3c3c3c;
}

.panel-header h4 {
  font-size: 14px;
  font-weight: 500;
  color: #ccc;
  margin: 0;
}

.changes-list {
  flex: 1;
  overflow-y: auto;
  padding: 8px;
}

.change-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 12px;
  border-radius: 6px;
  cursor: pointer;
  transition: background 0.2s;
  margin-bottom: 4px;
}

.change-item:hover {
  background: #2d2d30;
}

.change-status {
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 10px;
  text-transform: uppercase;
  font-weight: 500;
}

.change-status.draft { background: #3c3c3c; color: #888; }
.change-status.needs_review { background: #f57c00; color: #fff; }
.change-status.approved { background: #2e7d32; color: #fff; }
.change-status.merged { background: #7b1fa2; color: #fff; }

.change-info {
  flex: 1;
  min-width: 0;
}

.change-title {
  font-size: 13px;
  color: #ccc;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.change-date {
  font-size: 11px;
  color: #888;
}

.empty-changes {
  padding: 20px;
  text-align: center;
  color: #666;
  font-size: 13px;
}
</style>

<style>
.ai-pane .markdown-content {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, sans-serif;
}

.ai-pane .markdown-content code {
  background: #1a1a1a;
  padding: 2px 6px;
  border-radius: 4px;
  font-family: 'SF Mono', Monaco, Consolas, monospace;
  font-size: 12px;
}

.ai-pane .markdown-content pre {
  background: #1a1a1a;
  padding: 12px;
  border-radius: 6px;
  overflow-x: auto;
  margin: 8px 0;
}

.ai-pane .markdown-content pre code {
  background: none;
  padding: 0;
}

.ai-pane .markdown-content p {
  margin: 8px 0;
}

.ai-pane .markdown-content p:first-child {
  margin-top: 0;
}

.ai-pane .markdown-content p:last-child {
  margin-bottom: 0;
}

.ai-pane .markdown-content ul,
.ai-pane .markdown-content ol {
  margin: 8px 0;
  padding-left: 32px;
}

.ai-pane .markdown-content li {
  margin: 4px 0;
}

.ai-pane .markdown-content table {
  width: 100%;
  border-collapse: collapse;
  margin: 8px 0;
}

.ai-pane .markdown-content th,
.ai-pane .markdown-content td {
  border: 1px solid #3c3c3c;
  padding: 8px 12px;
  text-align: left;
}

.ai-pane .markdown-content th {
  background: #2d2d30;
  font-weight: 600;
}

.ai-pane .markdown-content tr:nth-child(even) {
  background: #252526;
}

.ai-pane .markdown-content a {
  color: #6bb8ff;
  text-decoration: none;
}

.ai-pane .markdown-content a:hover {
  text-decoration: underline;
}

.ai-pane .markdown-content blockquote {
  border-left: 3px solid #0e639c;
  margin: 8px 0;
  padding-left: 16px;
  color: #888;
}
</style>
