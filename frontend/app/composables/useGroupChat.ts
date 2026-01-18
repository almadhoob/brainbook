import { extractErrorMessage, formatDate, toDataUrl, initialsFromName, buildFullName } from './useGroupHelpers'
import type { ReceiveGroupMessageEventPayload } from '~/types'
import type { GroupMember } from './useGroupMembers'

export interface ApiGroupMessage {
  id?: number
  user_id?: number
  f_name?: string | null
  l_name?: string | null
  avatar?: string | null
  content?: string | null
  created_at?: string | null
}

export interface GroupChatMessage {
  id: number | string
  senderId: number
  senderName: string
  senderInitials: string
  avatarSrc?: string
  content: string
  createdAtRaw: string
  createdAtFormatted: string
}

export function useGroupChat(apiBase: string, groupId: Ref<number | null>, currentUserId: Ref<number | null>) {
  const toast = useToast()
  const sessionToken = useCookie<string | null>('session_token', { watch: false })
  const { sendGroupMessage, isUserOnline } = useRealtime()

  const groupMessages = reactive<Record<number, GroupChatMessage[]>>({})
  const groupMessagesLoading = reactive<Record<number, boolean>>({})
  const groupMessageDrafts = reactive<Record<number, string>>({})
  const groupMessageSending = reactive<Record<number, boolean>>({})
  const members = ref<GroupMember[]>([])

  const activeGroupMessages = computed(() =>
    groupId.value ? groupMessages[groupId.value] ?? [] : []
  )

  async function loadMessages() {
    if (!groupId.value) return

    const gid = groupId.value
    groupMessagesLoading[gid] = true
    try {
      const response = await $fetch<{ messages: ApiGroupMessage[] }>(
        `${apiBase}/protected/v1/groups/${gid}/messages?limit=50`,
        { credentials: 'include' }
      )
      groupMessages[gid] = normalizeGroupMessages(response.messages)
    } catch (error) {
      toast.add({
        title: 'Unable to load chat',
        description: extractErrorMessage(error) || 'Please try again later.',
        color: 'error'
      })
    } finally {
      groupMessagesLoading[gid] = false
    }
  }

  async function sendMessage() {
    const gid = groupId.value
    if (!gid) return

    const draft = groupMessageDrafts[gid]?.trim()
    if (!draft) {
      toast.add({ title: 'Message required', color: 'error' })
      return
    }

    if (!sessionToken.value) {
      toast.add({
        title: 'Missing session',
        description: 'Please sign in again.',
        color: 'error'
      })
      return
    }

    groupMessageSending[gid] = true
    const optimisticMessage: GroupChatMessage = {
      id: `local-${Date.now()}`,
      senderId: currentUserId.value ?? -1,
      senderName: 'You',
      senderInitials: 'You',
      avatarSrc: undefined,
      content: draft,
      createdAtRaw: new Date().toISOString(),
      createdAtFormatted: formatDate(new Date().toISOString())
    }
    appendGroupMessage(gid, optimisticMessage)
    try {
      sendGroupMessage({
        message: draft,
        group_id: gid,
        session_token: sessionToken.value
      })
      groupMessageDrafts[gid] = ''
    } catch (error) {
      toast.add({
        title: 'Message not sent',
        description: extractErrorMessage(error) || 'Try again shortly.',
        color: 'error'
      })
      await loadMessages()
    } finally {
      groupMessageSending[gid] = false
    }
  }

  function handleIncomingMessage(event: ReceiveGroupMessageEventPayload, groupMembers?: GroupMember[]) {
    const gid = event.group_id
    if (!gid) return

    const member = groupMembers?.find(m => m.id === event.sender_id)
    const message: GroupChatMessage = {
      id: `${gid}-${event.sent_at}-${event.sender_id}`,
      senderId: event.sender_id,
      senderName: member?.fullName || `User ${event.sender_id}`,
      senderInitials: member?.initials || initialsFromName(member?.fullName || ''),
      avatarSrc: member?.avatarSrc,
      content: event.message,
      createdAtRaw: event.sent_at,
      createdAtFormatted: formatDate(event.sent_at)
    }

    // Remove optimistic echo if it matches this incoming message.
    const existing = groupMessages[gid] ?? []
    const trimmed = event.message.trim()
    const cleaned = existing.filter(m =>
      !(typeof m.id === 'string'
        && m.id.startsWith('local-')
        && m.senderId === event.sender_id
        && m.content.trim() === trimmed)
    )
    if (cleaned.length !== existing.length) {
      groupMessages[gid] = cleaned
    }
    appendGroupMessage(gid, message)
  }

  function appendGroupMessage(gid: number, message: GroupChatMessage) {
    const list = groupMessages[gid] ?? []
    const exists = list.some(
      m => m.senderId === message.senderId
        && m.content === message.content
        && m.createdAtRaw === message.createdAtRaw
    )
    if (exists) return

    groupMessages[gid] = [...list, message]
  }

  function normalizeGroupMessages(messages?: ApiGroupMessage[]): GroupChatMessage[] {
    if (!Array.isArray(messages)) return []

    return messages.map((message, index) => {
      const senderName = buildFullName(message.f_name, message.l_name)
        || `User ${message.user_id ?? ''}`.trim()
      return {
        id: typeof message.id === 'number' ? message.id : `msg-${index}`,
        senderId: typeof message.user_id === 'number' ? message.user_id : -1,
        senderName,
        senderInitials: initialsFromName(senderName),
        avatarSrc: toDataUrl(message.avatar),
        content: (message.content ?? '').trim(),
        createdAtRaw: message.created_at ?? '',
        createdAtFormatted: formatDate(message.created_at)
      }
    })
  }

  return {
    groupMessages,
    groupMessagesLoading,
    groupMessageDrafts,
    groupMessageSending,
    activeGroupMessages,
    members,
    loadMessages,
    sendMessage,
    handleIncomingMessage,
    isUserOnline
  }
}
