import { defineStore } from 'pinia'
import { ref } from 'vue'
import { api } from '../api'

export interface Job {
  id: string
  project_id: string
  type: string
  status: string
  payload_json: string
  result_json: string
  error_text: string
  created_at: string
  started_at: string | null
  finished_at: string | null
}

export interface ChangeSet {
  id: string
  project_id: string
  job_id: string | null
  title: string
  base_ref: string
  target_ref: string | null
  apply_mode: string
  status: string
  summary_text: string
  created_at: string
  updated_at: string
}

export interface ReviewThread {
  id: string
  changeset_id: string
  file_path: string
  anchor_json: string
  status: string
  created_at: string
}

export interface ChatMessage {
  id: string
  chat_id: string
  role: 'user' | 'assistant' | 'system'
  content: string
  created_at: string
}

export interface Chat {
  id: string
  project_id: string
  title: string
  status: string
  created_at: string
  updated_at: string
}

export interface ChatChangeSet {
  id: string
  chat_id: string
  job_id: string | null
  title: string
  diff: string
  status: string
  summary_text: string
  created_at: string
}

export const useAIStore = defineStore('ai', () => {
  const jobs = ref<Job[]>([])
  const changeSets = ref<ChangeSet[]>([])
  const activeJob = ref<Job | null>(null)
  const activeChangeSet = ref<ChangeSet | null>(null)
  const messages = ref<ChatMessage[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)
  const wsConnected = ref(false)
  const pendingSubscribe = ref<string | null>(null)
  let ws: WebSocket | null = null

  let currentProjectId: string | null = null

  const chats = ref<Chat[]>([])
  const activeChat = ref<Chat | null>(null)
  const chatMessages = ref<ChatMessage[]>([])
  const chatWs = ref<WebSocket | null>(null)
  const streamingMessageId = ref<string | null>(null)
  const streamingContent = ref('')
  const isStreaming = ref(false)

  function connectWebSocket() {
    if (ws) {
      ws.close()
    }

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const wsUrl = `${protocol}//${window.location.host}/ws/ai`

    ws = new WebSocket(wsUrl)

    ws.onopen = () => {
      wsConnected.value = true
      if (pendingSubscribe.value) {
        ws?.send(JSON.stringify({ type: 'subscribe', payload: pendingSubscribe.value }))
        pendingSubscribe.value = null
      } else if (currentProjectId) {
        ws?.send(JSON.stringify({ type: 'subscribe', payload: currentProjectId }))
      }
      if (currentProjectId) {
        fetchJobs(currentProjectId)
        fetchChangeSets(currentProjectId)
      }
    }

    ws.onclose = () => {
      wsConnected.value = false
      pendingSubscribe.value = currentProjectId
      setTimeout(connectWebSocket, 3000)
    }

    ws.onerror = () => {
      wsConnected.value = false
    }

    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data)
        handleWSMessage(data)
      } catch (e) {
        console.error('Failed to parse WebSocket message:', e)
      }
    }
  }

  function handleWSMessage(data: any) {
    if (data.type === 'job_update') {
      const payload = data.payload
      const jobIndex = jobs.value.findIndex(j => j.id === payload.job_id)
      if (jobIndex !== -1) {
        jobs.value[jobIndex].status = payload.status
        if (payload.error) {
          jobs.value[jobIndex].error_text = payload.error
        }
        if (payload.result) {
          jobs.value[jobIndex].result_json = JSON.stringify(payload.result)
        }
      }
      if (activeJob.value?.id === payload.job_id && activeJob.value) {
        activeJob.value.status = payload.status
        if (payload.error) {
          activeJob.value.error_text = payload.error
        }
        if (payload.result) {
          activeJob.value.result_json = JSON.stringify(payload.result)
        }
      }
    } else if (data.type === 'changeset_created') {
      const payload = data.payload
      changeSets.value.unshift({
        id: payload.changeset_id,
        project_id: currentProjectId || '',
        job_id: null,
        title: payload.title,
        base_ref: '',
        target_ref: null,
        apply_mode: 'working_tree',
        status: payload.status,
        summary_text: payload.summary || '',
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString()
      })
    }
  }

  async function fetchJobs(projectId: string) {
    currentProjectId = projectId
    loading.value = true
    error.value = null
    jobs.value = []
    try {
      const response = await api.get(`/api/v1/projects/${projectId}/jobs`)
      jobs.value = response.data || []
    } catch (e: any) {
      error.value = e.response?.data?.error || 'Failed to fetch jobs'
      jobs.value = []
    } finally {
      loading.value = false
    }
  }

  async function fetchChangeSets(projectId: string) {
    currentProjectId = projectId
    loading.value = true
    error.value = null
    changeSets.value = []
    try {
      const response = await api.get(`/api/v1/projects/${projectId}/changesets`)
      changeSets.value = response.data || []
    } catch (e: any) {
      error.value = e.response?.data?.error || 'Failed to fetch changesets'
      changeSets.value = []
    } finally {
      loading.value = false
    }
  }

  function setProjectId(projectId: string) {
    currentProjectId = projectId
    pendingSubscribe.value = projectId
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify({ type: 'subscribe', payload: projectId }))
      pendingSubscribe.value = null
    }
  }

  async function createAITask(projectId: string, prompt: string, context?: Record<string, any>) {
    loading.value = true
    error.value = null
    try {
      const response = await api.post(`/api/v1/projects/${projectId}/ai/tasks`, {
        prompt,
        context: context || {},
        mode: 'patch_to_working_tree',
        max_files: 20
      })
      return response.data.job_id
    } catch (e: any) {
      error.value = e.response?.data?.error || 'Failed to create AI task'
      return null
    } finally {
      loading.value = false
    }
  }

  async function getJob(jobId: string) {
    loading.value = true
    error.value = null
    try {
      const response = await api.get(`/api/v1/jobs/${jobId}`)
      activeJob.value = response.data
      return response.data
    } catch (e: any) {
      error.value = e.response?.data?.error || 'Failed to fetch job'
      return null
    } finally {
      loading.value = false
    }
  }

  async function getChangeSet(csId: string) {
    loading.value = true
    error.value = null
    try {
      const response = await api.get(`/api/v1/changesets/${csId}`)
      activeChangeSet.value = response.data.changeset
      return response.data
    } catch (e: any) {
      error.value = e.response?.data?.error || 'Failed to fetch changeset'
      return null
    } finally {
      loading.value = false
    }
  }

  async function approveChangeSet(csId: string) {
    try {
      await api.post(`/api/v1/changesets/${csId}/approve`)
      await fetchChangeSets(currentProjectId || '')
    } catch (e: any) {
      error.value = e.response?.data?.error || 'Failed to approve'
    }
  }

  async function requestChanges(csId: string, comment: string) {
    try {
      await api.post(`/api/v1/changesets/${csId}/request-changes`, { comment })
      await fetchChangeSets(currentProjectId || '')
    } catch (e: any) {
      error.value = e.response?.data?.error || 'Failed to request changes'
    }
  }

  async function revertChangeSet(csId: string) {
    try {
      await api.post(`/api/v1/changesets/${csId}/revert`)
      await fetchChangeSets(currentProjectId || '')
    } catch (e: any) {
      error.value = e.response?.data?.error || 'Failed to revert'
    }
  }

  async function createThread(csId: string, filePath: string, anchor: string, body: string) {
    try {
      await api.post(`/api/v1/changesets/${csId}/threads`, {
        filePath,
        anchor,
        body
      })
    } catch (e: any) {
      error.value = e.response?.data?.error || 'Failed to create thread'
    }
  }

  async function addComment(threadId: string, body: string) {
    try {
      await api.post(`/api/v1/threads/${threadId}/comments`, { body })
    } catch (e: any) {
      error.value = e.response?.data?.error || 'Failed to add comment'
    }
  }

  async function resolveThread(threadId: string) {
    try {
      await api.post(`/api/v1/threads/${threadId}/resolve`)
    } catch (e: any) {
      error.value = e.response?.data?.error || 'Failed to resolve thread'
    }
  }

  function addMessage(role: ChatMessage['role'], content: string) {
    messages.value.push({
      id: crypto.randomUUID(),
      chat_id: '',
      role,
      content,
      created_at: new Date().toISOString()
    })
  }

  function clearMessages() {
    messages.value = []
  }

  async function fetchChats(projectId: string) {
    currentProjectId = projectId
    loading.value = true
    error.value = null
    try {
      const response = await api.get(`/api/v1/projects/${projectId}/ai/chats`)
      chats.value = response.data || []
    } catch (e: any) {
      error.value = e.response?.data?.error || 'Failed to fetch chats'
      chats.value = []
    } finally {
      loading.value = false
    }
  }

  async function createChat(projectId: string, title: string) {
    loading.value = true
    error.value = null
    try {
      const response = await api.post(`/api/v1/projects/${projectId}/ai/chats`, { title })
      const chat = response.data
      chats.value.unshift(chat)
      return chat
    } catch (e: any) {
      error.value = e.response?.data?.error || 'Failed to create chat'
      return null
    } finally {
      loading.value = false
    }
  }

  async function deleteChat(chatId: string) {
    loading.value = true
    error.value = null
    try {
      await api.delete(`/api/v1/projects/${currentProjectId}/ai/chats/${chatId}`)
      chats.value = chats.value.filter(c => c.id !== chatId)
      if (activeChat.value?.id === chatId) {
        activeChat.value = null
        chatMessages.value = []
      }
    } catch (e: any) {
      error.value = e.response?.data?.error || 'Failed to delete chat'
    } finally {
      loading.value = false
    }
  }

  async function fetchChatMessages(chatId: string) {
    loading.value = true
    error.value = null
    try {
      const response = await api.get(`/api/v1/projects/${currentProjectId}/ai/chats/${chatId}/messages`)
      chatMessages.value = response.data || []
    } catch (e: any) {
      error.value = e.response?.data?.error || 'Failed to fetch messages'
      chatMessages.value = []
    } finally {
      loading.value = false
    }
  }

  async function selectChat(chat: Chat) {
    activeChat.value = chat
    await fetchChatMessages(chat.id)
  }

  async function connectChatWS(chatId: string) {
    if (chatWs.value) {
      chatWs.value.close()
    }

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const wsUrl = `${protocol}//${window.location.host}/ws/ai/chats/${chatId}`

    chatWs.value = new WebSocket(wsUrl)

    chatWs.value.onopen = () => {
      wsConnected.value = true
    }

    chatWs.value.onclose = () => {
      wsConnected.value = false
      chatWs.value = null
    }

    chatWs.value.onerror = () => {
      wsConnected.value = false
    }

    chatWs.value.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data)
        handleChatWSMessage(data)
      } catch (e) {
        console.error('Failed to parse chat WebSocket message:', e)
      }
    }
  }

  function handleChatWSMessage(data: any) {
    if (data.type === 'chunk') {
      const payload = data.payload
      if (!payload.done) {
        if (streamingMessageId.value !== payload.message_id) {
          streamingMessageId.value = payload.message_id
          streamingContent.value = payload.content
          chatMessages.value.push({
            id: payload.message_id,
            chat_id: activeChat.value?.id || '',
            role: 'assistant',
            content: payload.content,
            created_at: new Date().toISOString()
          })
        } else {
          const msgIndex = chatMessages.value.findIndex(m => m.id === payload.message_id)
          if (msgIndex !== -1) {
            chatMessages.value[msgIndex].content += payload.content
          }
          streamingContent.value = chatMessages.value[msgIndex].content
        }
      } else {
        streamingMessageId.value = null
        streamingContent.value = ''
        isStreaming.value = false
      }
    } else if (data.type === 'message_created') {
      const payload = data.payload
      chatMessages.value.push({
        id: payload.id,
        chat_id: payload.chat_id,
        role: payload.role,
        content: payload.content,
        created_at: payload.created_at
      })
    }
  }

  function sendChatMessage(content: string) {
    if (!chatWs.value || chatWs.value.readyState !== WebSocket.OPEN) return

    isStreaming.value = true
    chatWs.value.send(JSON.stringify({
      type: 'send_message',
      payload: { content }
    }))
  }

  function stopStreaming() {
    if (chatWs.value) {
      chatWs.value.send(JSON.stringify({ type: 'stop' }))
    }
    isStreaming.value = false
    streamingMessageId.value = null
    streamingContent.value = ''
  }

  async function fetchChatChangeSets(chatId: string) {
    loading.value = true
    error.value = null
    try {
      const response = await api.get(`/api/v1/projects/${currentProjectId}/ai/chats/${chatId}/changesets`)
      return response.data || []
    } catch (e: any) {
      error.value = e.response?.data?.error || 'Failed to fetch chat changesets'
      return []
    } finally {
      loading.value = false
    }
  }

  return {
    jobs,
    changeSets,
    activeJob,
    activeChangeSet,
    messages,
    loading,
    error,
    wsConnected,
    connectWebSocket,
    setProjectId,
    fetchJobs,
    fetchChangeSets,
    createAITask,
    getJob,
    getChangeSet,
    approveChangeSet,
    requestChanges,
    revertChangeSet,
    createThread,
    addComment,
    resolveThread,
    addMessage,
    clearMessages,
    chats,
    activeChat,
    chatMessages,
    chatWs,
    fetchChats,
    createChat,
    deleteChat,
    selectChat,
    fetchChatMessages,
    connectChatWS,
    sendChatMessage,
    stopStreaming,
    isStreaming,
    streamingContent,
    fetchChatChangeSets
  }
})
