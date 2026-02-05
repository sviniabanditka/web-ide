import { defineStore } from 'pinia'
import { ref } from 'vue'

export interface ToolCall {
  id: string
  name: string
  arguments: Record<string, unknown>
  status: 'pending' | 'approved' | 'rejected' | 'executing' | 'completed' | 'error'
  result?: ToolResult
}

export interface ToolResult {
  ok: boolean
  data?: unknown
  error?: {
    code: string
    message: string
  }
  meta?: {
    duration_ms: number
    truncated?: boolean
  }
}

export interface PendingApproval {
  toolCall: ToolCall
  timestamp: number
}

export interface CommandOutputLine {
  stream: 'stdout' | 'stderr' | 'system'
  text: string
  ts: number
}

export const useAIToolsStore = defineStore('ai-tools', () => {
  const toolCalls = ref<Map<string, ToolCall>>(new Map())
  const pendingApprovals = ref<PendingApproval[]>([])
  const commandOutputs = ref<Map<string, CommandOutputLine[]>>(new Map())

  function handleToolCall(data: any) {
    const toolCall: ToolCall = {
      id: data.id || data.toolCallId,
      name: data.name,
      arguments: data.arguments || {},
      status: 'pending'
    }
    toolCalls.value.set(toolCall.id, toolCall)
  }

  function handleApprovalRequired(data: any) {
    const toolCall = toolCalls.value.get(data.id || data.toolCallId)
    if (toolCall) {
      toolCall.status = 'pending'
      
      pendingApprovals.value.push({
        toolCall,
        timestamp: Date.now()
      })
    }
  }

  function handleToolResult(data: any) {
    const id = data.id || data.toolCallId
    const toolCall = toolCalls.value.get(id)
    if (toolCall) {
      toolCall.status = data.ok ? 'completed' : 'error'
      toolCall.result = {
        ok: data.ok,
        data: data.result,
        error: data.error,
        meta: data.meta ? {
          duration_ms: data.duration || data.meta.duration_ms,
          truncated: data.meta.truncated
        } : undefined
      }

      pendingApprovals.value = pendingApprovals.value.filter(
        p => p.toolCall.id !== id
      )
    }
  }

  function handleToolError(data: any) {
    const id = data.id
    const toolCall = toolCalls.value.get(id)
    if (toolCall) {
      toolCall.status = 'error'
      toolCall.result = {
        ok: false,
        error: data.error || {
          code: 'TOOL_ERROR',
          message: 'Tool execution failed'
        }
      }
    }
  }

  function handleCommandOutput(data: any) {
    const handle = data.handle
    if (!commandOutputs.value.has(handle)) {
      commandOutputs.value.set(handle, [])
    }
    const outputs = commandOutputs.value.get(handle)!
    outputs.push({
      stream: data.stream,
      text: data.text,
      ts: data.ts || Date.now()
    })
  }

  function handleCommandDone(data: any) {
    const handle = data.handle
    const outputs = commandOutputs.value.get(handle)
    if (outputs) {
      outputs.push({
        stream: 'system',
        text: `[Command exited with code ${data.exit_code}]`,
        ts: Date.now()
      })
    }
  }

  function approveTool(toolCallId: string) {
    const ws = (window as unknown as { chatWs: WebSocket | null }).chatWs
    if (!ws || ws.readyState !== WebSocket.OPEN) {
      console.error('[TOOLS] WebSocket not available')
      return
    }

    const message = JSON.stringify({
      type: 'tool.approve',
      payload: { tool_call_id: toolCallId }
    })
    ws.send(message)

    const toolCall = toolCalls.value.get(toolCallId)
    if (toolCall) {
      toolCall.status = 'executing'
    }
  }

  function rejectTool(toolCallId: string, reason?: string) {
    const ws = (window as unknown as { chatWs: WebSocket | null }).chatWs
    if (!ws || ws.readyState !== WebSocket.OPEN) {
      console.error('[TOOLS] WebSocket not available')
      return
    }

    const message = JSON.stringify({
      type: 'tool.reject',
      payload: { tool_call_id: toolCallId, reason }
    })
    ws.send(message)

    const toolCall = toolCalls.value.get(toolCallId)
    if (toolCall) {
      toolCall.status = 'rejected'
    }
  }

  function getCommandOutput(handle: string): CommandOutputLine[] {
    return commandOutputs.value.get(handle) || []
  }

  function clearToolState() {
    toolCalls.value.clear()
    pendingApprovals.value = []
    commandOutputs.value.clear()
  }

  return {
    toolCalls,
    pendingApprovals,
    commandOutputs,
    handleToolCall,
    handleApprovalRequired,
    handleToolResult,
    handleToolError,
    handleCommandOutput,
    handleCommandDone,
    approveTool,
    rejectTool,
    getCommandOutput,
    clearToolState
  }
})

declare global {
  interface Window {
    chatWs: WebSocket | null
  }
}
