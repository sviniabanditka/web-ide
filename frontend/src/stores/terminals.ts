import { defineStore } from 'pinia'
import { ref } from 'vue'
import { api } from '../api'

interface Terminal {
  id: string
  project_id: string
  title: string
  cwd: string
  shell: string
  status: string
  created_at: string
}

interface WSConnection {
  ws: WebSocket | null
  callbacks: Map<string, ((data: any) => void)>
  terminalId: string
  backlog: string
}

function getSessionToken(): string {
  const value = `; ${document.cookie}`
  const parts = value.split(`; session_token=`)
  if (parts.length === 2) return parts.pop()?.split(';').shift() || ''
  return ''
}

export const useTerminalsStore = defineStore('terminals', () => {
  const terminals = ref<Terminal[]>([])
  const currentTerminal = ref<Terminal | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)
  const wsConnections = ref<Map<string, WSConnection>>(new Map())

  function getOrCreateWSConnection(terminalId: string): WebSocket {
    const existing = wsConnections.value.get(terminalId)
    console.log('[WS] getOrCreateWSConnection:', terminalId, 'existing:', existing?.ws?.readyState)
    if (existing?.ws && existing.ws.readyState === WebSocket.OPEN) {
      console.log('[WS] reusing existing connection for:', terminalId)
      return existing.ws
    }

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const wsUrl = `${protocol}//${window.location.host}/api/v1/terminals/${terminalId}/ws?token=${getSessionToken()}`
    console.log('[WS] creating new connection for:', terminalId, 'url:', wsUrl)

    const ws = new WebSocket(wsUrl)

    wsConnections.value.set(terminalId, {
      ws,
      callbacks: new Map(),
      terminalId,
      backlog: ''
    })

    ws.onopen = () => {
      console.log('[WS] connected:', terminalId)
    }

    ws.onmessage = (event) => {
      const conn = wsConnections.value.get(terminalId)
      if (conn) {
        conn.backlog += event.data
        conn.callbacks.forEach((callback) => {
          callback(event.data)
        })
      }
    }

    ws.onclose = () => {
      console.log('[WS] closed:', terminalId)
      const conn = wsConnections.value.get(terminalId)
      if (conn) {
        conn.ws = null
      }
    }

    ws.onerror = () => {
      console.log('[WS] error:', terminalId)
      const conn = wsConnections.value.get(terminalId)
      if (conn) {
        conn.ws = null
      }
    }

    return ws
  }

  function onWSMessage(terminalId: string, key: string, callback: (data: any) => void): () => void {
    let conn = wsConnections.value.get(terminalId)
    if (!conn) {
      console.log('[WS] no connection for', terminalId, ', creating...')
      getOrCreateWSConnection(terminalId)
      conn = wsConnections.value.get(terminalId)!
    }
    console.log('[WS] subscribing to', terminalId, ', readyState:', conn.ws?.readyState)
    if (conn.backlog) {
      console.log('[WS] sending backlog:', conn.backlog.length, 'chars')
      callback(conn.backlog)
    }
    conn.callbacks.set(key, callback)
    return () => conn?.callbacks.delete(key)
  }

  function sendToTerminal(terminalId: string, data: string) {
    const conn = wsConnections.value.get(terminalId)
    if (conn?.ws && conn.ws.readyState === WebSocket.OPEN) {
      conn.ws.send(data)
    }
  }

  function closeWSConnection(terminalId: string) {
    const conn = wsConnections.value.get(terminalId)
    if (conn?.ws) {
      conn.ws.close()
      conn.ws = null
      conn.backlog = ''
    }
    wsConnections.value.delete(terminalId)
  }

  function closeAllWSConnections() {
    wsConnections.value.forEach((conn) => {
      if (conn.ws) {
        conn.ws.close()
      }
    })
    wsConnections.value.clear()
  }

  async function loadTerminalsFromState(projectId: string, terminalIds: string[]) {
    await fetchTerminals(projectId)

    const openTerminals = terminals.value.filter(t => terminalIds.includes(t.id))
    if (openTerminals.length > 0) {
      currentTerminal.value = openTerminals[0]
    } else if (terminals.value.length > 0) {
      currentTerminal.value = terminals.value[0]
    }
  }

  async function fetchTerminals(projectId: string) {
    loading.value = true
    error.value = null
    try {
      console.log('Fetching terminals for project:', projectId)
      const response = await api.get(`/api/v1/projects/${projectId}/terminals`)
      console.log('Fetched terminals:', response.data)
      terminals.value = response.data || []
    } catch (e: any) {
      error.value = e.response?.data?.error || 'Failed to fetch terminals'
      console.error('fetchTerminals error:', e)
      terminals.value = []
    } finally {
      loading.value = false
    }
  }

  async function createTerminal(projectId: string, title?: string, cwd?: string) {
    loading.value = true
    error.value = null
    try {
      console.log('Creating terminal for project:', projectId)
      const response = await api.post(`/api/v1/projects/${projectId}/terminals`, {
        title,
        cwd
      })
      console.log('Terminal created:', response.data)
      if (!terminals.value) {
        terminals.value = []
      }
      terminals.value.push(response.data)
      currentTerminal.value = response.data
      return response.data
    } catch (e: any) {
      error.value = e.response?.data?.error || 'Failed to create terminal'
      console.error('createTerminal error:', e)
      return null
    } finally {
      loading.value = false
    }
  }

  async function closeTerminal(terminalId: string) {
    console.log('closeTerminal called:', terminalId)
    try {
      await api.post(`/api/v1/terminals/${terminalId}/close`)
      console.log('Terminal closed on server')
      closeWSConnection(terminalId)
      terminals.value = terminals.value.filter(t => t.id !== terminalId)
      console.log('Terminal removed from list, remaining:', terminals.value.length)
      if (currentTerminal.value?.id === terminalId) {
        currentTerminal.value = terminals.value.length > 0 ? terminals.value[0] : null
        console.log('Current terminal cleared')
      }
    } catch (e: any) {
      error.value = e.response?.data?.error || 'Failed to close terminal'
      console.error('closeTerminal error:', e)
    }
  }

  async function resizeTerminal(terminalId: string, cols: number, rows: number) {
    try {
      await api.post(`/api/v1/terminals/${terminalId}/resize`, { cols, rows })
    } catch (e: any) {
      console.error('Failed to resize terminal:', e)
    }
  }

  function setCurrentTerminal(terminal: Terminal | null) {
    console.log('setCurrentTerminal:', terminal?.id)
    currentTerminal.value = terminal
  }

  function getOpenTerminalIds(): string[] {
    return terminals.value.map(t => t.id)
  }

  return {
    terminals,
    currentTerminal,
    loading,
    error,
    loadTerminalsFromState,
    fetchTerminals,
    createTerminal,
    closeTerminal,
    resizeTerminal,
    setCurrentTerminal,
    getOpenTerminalIds,
    getOrCreateWSConnection,
    onWSMessage,
    sendToTerminal,
    closeWSConnection,
    closeAllWSConnections
  }
})
