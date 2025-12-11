import { createSharedComposable, useEventBus, useWebSocket } from '@vueuse/core'
import type { WebSocketStatus } from '@vueuse/core'
import type { ReceiveMessageEventPayload, WebsocketNotificationPayload } from '~/types'

interface UserStatusInfo {
  id: number
  full_name: string
  status: number
  last_message_time?: string | null
}

interface UserStatusUpdatePayload {
  online_users?: UserStatusInfo[]
  offline_user_ids?: number[]
}

const buildWsUrl = (baseURL: string) => {
  try {
    const url = new URL(baseURL)
    url.protocol = url.protocol === 'https:' ? 'wss:' : 'ws:'
    url.pathname = '/protected/ws'
    url.search = ''
    url.hash = ''
    return url.toString()
  } catch (err) {
    console.error('Invalid API base URL for websocket:', baseURL, err)
    return ''
  }
}

const _useRealtime = () => {
  const runtimeConfig = useRuntimeConfig()
  const apiBase = runtimeConfig.public?.apiBase || 'http://localhost:8080'

  const onlineUsers = ref<Record<number, UserStatusInfo>>({})
  const pendingQueue: string[] = []

  const directMessageBus = useEventBus<ReceiveMessageEventPayload>('ws:direct-message')
  const notificationBus = useEventBus<WebsocketNotificationPayload>('ws:notification')

  const wsEndpoint = buildWsUrl(apiBase)
  let socket: ReturnType<typeof useWebSocket> | null = null

  if (import.meta.client && wsEndpoint) {
    socket = useWebSocket(wsEndpoint, {
      immediate: false,
      autoReconnect: {
        retries: 10,
        delay: 2000
      },
      onMessage: (_ws, event) => {
        handleIncoming(event.data)
      },
      onConnected: () => {
        flushQueue()
      },
      onDisconnected: () => {
        // no-op, autoReconnect will handle
      }
    })
  }

  const getStatus = (): WebSocketStatus => (socket ? socket.status.value : 'CLOSED')
  const status = computed<WebSocketStatus>(() => getStatus())

  const connect = () => {
    if (!socket || !import.meta.client) return
    const currentStatus = getStatus()
    if (currentStatus === 'OPEN' || currentStatus === 'CONNECTING') {
      return
    }
    socket.open()
  }

  const flushQueue = () => {
    if (!socket) return
    while (pendingQueue.length) {
      const next = pendingQueue.shift()
      if (!next) break
      socket.send(next)
    }
  }

  const sendEvent = (type: string, payload: Record<string, unknown>) => {
    if (!socket || !import.meta.client) return false
    const message = JSON.stringify({ type, payload })
    if (getStatus() !== 'OPEN') {
      pendingQueue.push(message)
      connect()
      return true
    }
    return socket.send(message)
  }

  const handleIncoming = (raw: string) => {
    try {
      const parsed = JSON.parse(raw) as { type: string, payload: unknown }
      switch (parsed.type) {
        case 'receive_message':
          directMessageBus.emit(parsed.payload as ReceiveMessageEventPayload)
          break
        case 'notification':
          notificationBus.emit(parsed.payload as WebsocketNotificationPayload)
          break
        case 'user_status_update':
          applyStatusUpdate(parsed.payload as UserStatusUpdatePayload)
          break
        default:
          break
      }
    } catch (err) {
      console.error('Failed to parse websocket payload', err)
    }
  }

  const applyStatusUpdate = (update: UserStatusUpdatePayload) => {
    let next: Record<number, UserStatusInfo> = { ...onlineUsers.value }
    for (const user of update.online_users ?? []) {
      next[user.id] = user
    }
    for (const id of update.offline_user_ids ?? []) {
      const { [id]: _removed, ...rest } = next
      next = rest
    }
    onlineUsers.value = next
  }

  const isUserOnline = (userId: number | null | undefined) => {
    if (typeof userId !== 'number') return false
    return Boolean(onlineUsers.value[userId])
  }

  const sendDirectMessage = (payload: Record<string, unknown>) => sendEvent('send_message', payload)
  const sendTypingEvent = (payload: Record<string, unknown>) => sendEvent('send_typing', payload)

  return {
    status,
    connect,
    sendEvent,
    sendDirectMessage,
    sendTypingEvent,
    directMessageBus,
    notificationBus,
    onlineUsers: readonly(onlineUsers),
    isUserOnline
  }
}

export const useRealtime = createSharedComposable(_useRealtime)
