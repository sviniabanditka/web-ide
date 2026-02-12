<template>
  <div class="h-full flex">
    <aside class="w-[220px] bg-card border-r flex flex-col">
      <div class="p-3 border-b">
        <Button class="w-full" @click="createNewChat">New Chat</Button>
      </div>
      <ScrollArea class="flex-1">
        <div v-if="aiStore.chats.length === 0" class="p-5 text-center text-sm text-muted-foreground">
          No chats yet
        </div>
        <div
          v-for="chat in aiStore.chats"
          :key="chat.id"
          class="flex items-center justify-between px-3 py-2.5 cursor-pointer hover:bg-accent rounded-sm mx-2 mt-1"
          :class="{ 'bg-accent': aiStore.activeChat?.id === chat.id }"
          @click="selectChat(chat)"
        >
          <span class="text-sm truncate">{{ chat.title || 'Untitled' }}</span>
          <button class="text-muted-foreground hover:text-destructive" @click.stop="deleteChat(chat.id)">×</button>
        </div>
      </ScrollArea>
    </aside>

    <div class="flex-1 flex flex-col overflow-hidden" v-if="aiStore.activeChat">
      <div class="flex-shrink-0 px-4 py-3 border-b">
        <h3 class="font-medium">{{ aiStore.activeChat.title }}</h3>
      </div>
      <ScrollArea class="flex-1">
        <div class="p-4 space-y-4">
          <template v-for="msg in sortedMessages" :key="msg.id">
            <div v-if="msg.role === 'thinking' && msg.content" class="thinking-message">
              <ThinkingBlock :thinking="msg.content" />
            </div>
            <div v-else-if="msg.role === 'tool' && msg.tool_results?.length" class="message-tool-results">
              <ToolBlock
                v-for="tool in msg.tool_results"
                :key="tool.id"
                :tool="{
                  id: tool.id,
                  name: tool.name,
                  arguments: {},
                  status: tool.ok ? 'completed' : 'error'
                }"
                :result="tool"
              />
            </div>
            <div v-else-if="msg.role === 'tool_block' && msg.tool_calls?.length" class="message-tool-calls">
              <ToolBlock
                v-for="tool in msg.tool_calls"
                :key="tool.id"
                :tool="tool"
                :result="msg.tool_results?.find(r => r.id === tool.id)"
              />
            </div>
            <div
              v-else-if="msg.role === 'assistant' && msg.content"
              class="flex gap-3 max-w-[80%] mr-auto"
            >
              <div>
                <div class="text-xs text-muted-foreground mb-1">AI</div>
                <div
                  class="px-3.5 py-2 rounded-lg text-sm bg-muted"
                >
                  <div class="markdown-content" v-html="msg.parsedContent || ''"></div>
                  <span v-if="aiStore.streamingMessageId === msg.id && aiStore.isStreaming" class="cursor">|</span>
                </div>
              </div>
            </div>
            <div
              v-else-if="msg.role === 'user'"
              class="flex gap-3 max-w-[80%] ml-auto flex-row-reverse"
            >
              <div>
                <div class="text-xs text-muted-foreground mb-1 text-right">You</div>
                <div
                  class="px-3.5 py-2 rounded-lg text-sm bg-primary text-primary-foreground"
                >
                  <span>{{ msg.content }}</span>
                </div>
              </div>
            </div>
          </template>
        </div>
      </ScrollArea>
      <div class="flex-shrink-0 p-4 border-t bg-card space-y-2">
        <Textarea
          v-model="userMessage"
          placeholder="Describe what you want to do..."
          :disabled="aiStore.isStreaming"
          class="min-h-[80px] resize-none"
          @keydown.ctrl.enter="sendMessage"
        />
        <div class="flex items-center justify-between">
          <div v-if="aiStore.modelStatus !== 'idle'" class="flex items-center gap-2 text-sm">
            <span class="w-2 h-2 rounded-full animate-pulse" :class="{
              'bg-amber-500': aiStore.modelStatus === 'thinking',
              'bg-blue-500': aiStore.modelStatus === 'using_tool',
              'bg-green-500': aiStore.modelStatus === 'editing',
              'bg-purple-500': aiStore.modelStatus === 'planning'
            }"></span>
            <span class="text-muted-foreground">{{ getStatusText(aiStore.modelStatus) }}</span>
          </div>
          <div v-else></div>
          <div class="flex items-center gap-2">
            <UsageRing />
            <Button v-if="aiStore.isStreaming" variant="destructive" size="sm" @click="stopStreaming">Stop</Button>
            <Button @click="sendMessage" :disabled="aiStore.isStreaming || !userMessage.trim()">Send</Button>
          </div>
        </div>
      </div>
    </div>

    <div v-else class="flex-1 flex items-center justify-center">
      <div class="text-center text-muted-foreground">
        <h3 class="text-xl mb-2">AI Chat</h3>
        <p class="mb-4">Select a chat or create a new one to start coding with AI</p>
        <Button @click="createNewChat">New Chat</Button>
      </div>
    </div>

    <aside v-if="aiStore.activeChat" class="w-[280px] border-l bg-card flex flex-col">
      <div class="px-4 py-3 border-b">
        <h4 class="font-medium text-sm">Changes</h4>
      </div>
      <ScrollArea class="flex-1">
        <div
          v-for="cs in chatChangeSets"
          :key="cs.id"
          class="flex items-center gap-3 p-3 cursor-pointer hover:bg-accent rounded-sm mx-2 mt-1"
          @click="selectChangeSet(cs)"
        >
          <Badge :variant="getStatusVariant(cs.status)" class="text-xs">{{ cs.status }}</Badge>
          <div class="flex-1 min-w-0">
            <div class="text-sm truncate">{{ cs.title || 'No title' }}</div>
            <div class="text-xs text-muted-foreground">{{ formatDate(cs.created_at) }}</div>
          </div>
        </div>
        <div v-if="chatChangeSets.length === 0" class="p-5 text-center text-sm text-muted-foreground">
          No changes yet
        </div>
      </ScrollArea>
    </aside>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch, nextTick, computed } from 'vue'
import { useAIStore, type Chat, type ChatChangeSet } from '../stores/ai'
import UsageRing from '../components/UsageRing.vue'
import ToolBlock from '../components/ai/ToolBlock.vue'
import ThinkingBlock from '../components/ai/ThinkingBlock.vue'
import Button from '@/components/ui/Button.vue'
import Textarea from '@/components/ui/Textarea.vue'
import ScrollArea from '@/components/ui/ScrollArea.vue'
import Badge from '@/components/ui/Badge.vue'

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
const chatChangeSets = ref<ChatChangeSet[]>([])

const sortedMessages = computed(() => {
  return [...aiStore.chatMessages]
})

function scrollToBottom() {
  nextTick(() => {
    const container = document.querySelector('.scroll-area-content')
    if (container) {
      container.scrollTop = container.scrollHeight
    }
  })
}

watch(() => aiStore.chatMessages.length, scrollToBottom)
watch(() => aiStore.chatMessages, scrollToBottom, { deep: true })
watch(() => aiStore.streamingContent, scrollToBottom)

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

function getStatusVariant(status: string): 'default' | 'secondary' | 'destructive' | 'outline' {
  switch (status) {
    case 'draft': return 'secondary'
    case 'needs_review': return 'outline'
    case 'approved': return 'default'
    case 'merged': return 'secondary'
    default: return 'secondary'
  }
}

onMounted(async () => {
  await aiStore.fetchChats(props.project.id)
})

onUnmounted(() => {
  if (aiStore.chatWs) {
    aiStore.chatWs.close()
  }
})
</script>

<style>
.cursor {
  display: inline-block;
  animation: blink 1s infinite;
}

@keyframes blink {
  0%, 50% { opacity: 1; }
  51%, 100% { opacity: 0; }
}

.markdown-content {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
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
  border: 1px solid hsl(var(--border));
  padding: 8px 12px;
  text-align: left;
}

.markdown-content th {
  background: hsl(var(--muted));
  font-weight: 600;
}

.markdown-content tr:nth-child(even) {
  background: hsl(var(--card));
}

.markdown-content a {
  color: hsl(var(--primary));
  text-decoration: none;
}

.markdown-content a:hover {
  text-decoration: underline;
}

.markdown-content blockquote {
  border-left: 3px solid hsl(var(--primary));
  margin: 8px 0;
  padding-left: 16px;
  color: hsl(var(--muted-foreground));
}
</style>
