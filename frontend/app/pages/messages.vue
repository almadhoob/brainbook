<script setup lang="ts">
import { formatDistanceToNow } from 'date-fns'
import type { ApiConversationMessage, ApiUserListItem, ReceiveMessageEventPayload } from '~/types'
import { normalizeAvatar } from '~/utils'

const runtimeConfig = useRuntimeConfig()
const apiBase = runtimeConfig.public?.apiBase || 'http://localhost:8080'
const toast = useToast()
const route = useRoute()
const router = useRouter()

const { session, hydrate } = useSession()
if (import.meta.client) {
  await hydrate()
}

const { status: realtimeStatus, connect, directMessageBus, sendDirectMessage, isUserOnline } = useRealtime()
const sessionToken = useCookie<string | null>('session_token', { watch: false })

type ChatPartner = {
  id: number
  name: string
  avatar?: string
  lastMessageTime?: string | null
  lastMessageSnippet?: string
  hasUnread?: boolean
}

type ConversationMessage = {
  id: string
  senderId: number
  content: string
  createdAt: string
  isMine: boolean
}

const chatPartners = ref<ChatPartner[]>([])
const partnerSearch = ref('')
const loadingPartners = ref(true)
const loadingConversation = ref(false)
const selectedPartnerId = ref<number | null>(null)
const newMessage = ref('')
const sendingMessage = ref(false)
const messagesByUser = reactive<Record<number, ConversationMessage[]>>({})
const conversationRef = ref<HTMLElement | null>(null)

const currentUserId = computed(() => session.value.user_id)
const filteredPartners = computed(() => {
  const query = partnerSearch.value.toLowerCase().trim()
  if (!query) return chatPartners.value
  return chatPartners.value.filter(partner => partner.name.toLowerCase().includes(query))
})
const selectedPartner = computed(() => chatPartners.value.find(partner => partner.id === selectedPartnerId.value) || null)
const activeMessages = computed(() => (selectedPartnerId.value ? messagesByUser[selectedPartnerId.value] ?? [] : []))
const isRealtimeConnected = computed(() => realtimeStatus.value === 'OPEN')

const fetchPartners = async () => {
  loadingPartners.value = true
  try {
    const response = await $fetch<{ users: ApiUserListItem[] }>(`${apiBase}/protected/v1/user-list`, {
      credentials: 'include'
    })
    chatPartners.value = (response.users ?? []).map(user => ({
      id: user.user_id,
      name: user.user_full_name,
      avatar: normalizeAvatar(user.user_avatar ?? null),
      lastMessageTime: user.last_message_time ?? null,
      lastMessageSnippet: undefined,
      hasUnread: false
    }))

    const requested = Number(route.query.user)
    if (!Number.isNaN(requested) && chatPartners.value.some(partner => partner.id === requested)) {
      await selectPartner(requested)
    } else if (!selectedPartnerId.value) {
      const firstPartner = chatPartners.value[0]
      if (firstPartner) {
        selectedPartnerId.value = firstPartner.id
        await loadConversation(firstPartner.id)
      }
    }
  } catch (err) {
    toast.add({ title: 'Unable to load people', description: 'Please try again later', color: 'error' })
    console.error(err)
  } finally {
    loadingPartners.value = false
  }
}

const loadConversation = async (partnerId: number) => {
  if (messagesByUser[partnerId]) return
  loadingConversation.value = true
  try {
    const response = await $fetch<{ messages: ApiConversationMessage[] }>(`${apiBase}/protected/v1/private-messages/user/${partnerId}?limit=50`, {
      credentials: 'include'
    })
    const history = response.messages ?? []
    messagesByUser[partnerId] = history.map(message => ({
      id: `${partnerId}-${message.created_at}-${message.sender_id}-${Math.random()}`,
      senderId: message.sender_id,
      content: message.content,
      createdAt: message.created_at,
      isMine: determineIsMine(message.sender_id, partnerId)
    }))
    const lastMessage = history.at(-1)
    if (lastMessage) {
      updatePartnerPreview(partnerId, lastMessage.content, lastMessage.created_at, false)
    }
    nextTick(scrollConversationToBottom)
  } catch (err) {
    toast.add({ title: 'Unable to load conversation', description: 'Please try again later', color: 'error' })
    console.error(err)
  } finally {
    loadingConversation.value = false
  }
}

const determineIsMine = (senderId: number, partnerId: number) => {
  if (currentUserId.value) {
    return senderId === currentUserId.value
  }
  return senderId !== partnerId
}

const selectPartner = async (partnerId: number) => {
  if (selectedPartnerId.value === partnerId) return
  selectedPartnerId.value = partnerId
  await loadConversation(partnerId)
  markPartnerAsRead(partnerId)
  nextTick(scrollConversationToBottom)
  if (route.query.user !== String(partnerId)) {
    router.replace({ query: { ...route.query, user: partnerId } })
  }
}

const markPartnerAsRead = (partnerId: number) => {
  chatPartners.value = chatPartners.value.map(partner => partner.id === partnerId
    ? { ...partner, hasUnread: false }
    : partner)
}

const updatePartnerPreview = (partnerId: number, snippet: string, timestamp: string, markUnread: boolean) => {
  chatPartners.value = chatPartners.value.map((partner) => {
    if (partner.id !== partnerId) return partner
    return {
      ...partner,
      lastMessageSnippet: snippet,
      lastMessageTime: timestamp,
      hasUnread: markUnread ? true : partner.hasUnread
    }
  })
}

const appendMessage = (partnerId: number, message: ConversationMessage) => {
  const existing = messagesByUser[partnerId] ?? []
  messagesByUser[partnerId] = [...existing, message]
  if (selectedPartnerId.value === partnerId) {
    nextTick(scrollConversationToBottom)
  }
}

const handleIncomingMessage = (event: ReceiveMessageEventPayload) => {
  const partnerId = resolvePartnerId(event)
  const message: ConversationMessage = {
    id: `${partnerId}-${event.sent_at}-${event.sender_id}-${globalThis.crypto?.randomUUID?.() ?? Math.random()}`,
    senderId: event.sender_id,
    content: event.message,
    createdAt: event.sent_at,
    isMine: determineIsMine(event.sender_id, partnerId)
  }

  appendMessage(partnerId, message)
  const isIncoming = currentUserId.value
    ? event.sender_id !== currentUserId.value
    : event.sender_id === partnerId
  updatePartnerPreview(partnerId, event.message, event.sent_at, isIncoming && partnerId !== selectedPartnerId.value)

  if (isIncoming && partnerId !== selectedPartnerId.value) {
    chatPartners.value = chatPartners.value.map(partner => partner.id === partnerId
      ? { ...partner, hasUnread: true }
      : partner)
  } else {
    markPartnerAsRead(partnerId)
  }
}

const resolvePartnerId = (event: ReceiveMessageEventPayload) => {
  if (currentUserId.value && event.sender_id === currentUserId.value) {
    return event.receiver_id
  }
  if (currentUserId.value && event.receiver_id === currentUserId.value) {
    return event.sender_id
  }
  if (selectedPartnerId.value && event.sender_id === selectedPartnerId.value) {
    return event.receiver_id
  }
  return event.sender_id
}

const sendMessage = async () => {
  if (!selectedPartnerId.value) return
  const trimmed = newMessage.value.trim()
  if (!trimmed) return
  if (!sessionToken.value) {
    toast.add({ title: 'Missing session token', description: 'Please sign in again.', color: 'error' })
    return
  }

  sendingMessage.value = true
  const payload = {
    message: trimmed,
    receiver_id: selectedPartnerId.value,
    session_token: sessionToken.value
  }

  const optimisticMessage: ConversationMessage = {
    id: `local-${Date.now()}`,
    senderId: currentUserId.value ?? -1,
    content: trimmed,
    createdAt: new Date().toISOString(),
    isMine: true
  }

  appendMessage(selectedPartnerId.value, optimisticMessage)
  updatePartnerPreview(selectedPartnerId.value, trimmed, optimisticMessage.createdAt, false)
  newMessage.value = ''

  try {
    connect()
    const sent = sendDirectMessage(payload)
    if (!sent) {
      toast.add({ title: 'Reconnecting…', description: 'We will retry sending automatically.', color: 'neutral' })
    }
  } catch (err) {
    toast.add({ title: 'Message not sent', description: 'Please try again.', color: 'error' })
    console.error(err)
  } finally {
    sendingMessage.value = false
  }
}

const handleComposerKeydown = (event: KeyboardEvent) => {
  if (event.key === 'Enter' && !event.shiftKey) {
    event.preventDefault()
    sendMessage()
  }
}

const scrollConversationToBottom = () => {
  if (!conversationRef.value) return
  conversationRef.value.scrollTop = conversationRef.value.scrollHeight
}

const formatTimestamp = (timestamp?: string | null) => {
  if (!timestamp) return 'No activity'
  const date = new Date(timestamp)
  if (Number.isNaN(date.getTime())) return timestamp
  return formatDistanceToNow(date, { addSuffix: true })
}

watch(() => route.query.user, async (userId) => {
  const parsed = Number(userId)
  if (!Number.isNaN(parsed) && chatPartners.value.some(partner => partner.id === parsed)) {
    await selectPartner(parsed)
  }
}, { immediate: true })

watch(activeMessages, () => {
  if (selectedPartnerId.value) {
    markPartnerAsRead(selectedPartnerId.value)
  }
})

if (import.meta.client) {
  connect()
  fetchPartners()

  const stop = directMessageBus.on(handleIncomingMessage)
  onScopeDispose(stop)
}
</script>

<template>
  <div class="flex h-full flex-col lg:flex-row lg:gap-6">
    <section class="lg:w-80 w-full lg:flex lg:flex-col">
      <UDashboardPanel id="messages-list" :resizable="false" class="h-full">
        <UDashboardNavbar title="Messages">
          <template #leading>
            <UDashboardSidebarCollapse />
          </template>
          <template #trailing>
            <UBadge :label="chatPartners.length" variant="subtle" />
          </template>
        </UDashboardNavbar>

        <div class="p-3 border-b border-default">
          <UInput
            v-model="partnerSearch"
            icon="i-lucide-search"
            placeholder="Search people"
          />
        </div>

        <div class="flex-1 overflow-y-auto">
          <div v-if="loadingPartners" class="py-8 text-center text-muted text-sm">
            Loading people…
          </div>

          <template v-else>
            <button
              v-for="partner in filteredPartners"
              :key="partner.id"
              type="button"
              class="w-full border-b border-default last:border-b-0 px-4 py-3 text-left hover:bg-primary/5 transition"
              :class="partner.id === selectedPartnerId && 'bg-primary/10'"
              @click="selectPartner(partner.id)"
            >
              <div class="flex items-center gap-3">
                <UAvatar
                  :src="partner.avatar"
                  :text="partner.name.split(' ').map(part => part[0]).join('').slice(0, 2)"
                />
                <div class="min-w-0 flex-1">
                  <div class="flex items-center gap-2">
                    <p class="font-medium text-highlighted truncate">
                      {{ partner.name }}
                    </p>
                    <span
                      class="size-2 rounded-full"
                      :class="isUserOnline(partner.id) ? 'bg-emerald-500' : 'bg-muted'"
                    />
                    <UBadge
                      v-if="partner.hasUnread"
                      label="new"
                      size="xs"
                      color="primary"
                      variant="subtle"
                    />
                  </div>
                  <p class="text-xs text-muted truncate">
                    {{ partner.lastMessageSnippet || 'Start a conversation' }}
                  </p>
                  <p class="text-[10px] text-muted">
                    {{ formatTimestamp(partner.lastMessageTime) }}
                  </p>
                </div>
              </div>
            </button>
          </template>
        </div>
      </UDashboardPanel>
    </section>

    <section class="flex-1 flex flex-col border border-default rounded-2xl overflow-hidden mt-6 lg:mt-0">
      <header class="border-b border-default px-4 py-3 flex items-center justify-between">
        <div class="min-w-0">
          <p class="font-semibold text-lg text-highlighted truncate">
            {{ selectedPartner?.name || 'Select a conversation' }}
          </p>
          <p class="text-sm text-muted">
            <span v-if="selectedPartnerId && isUserOnline(selectedPartnerId)" class="text-emerald-500">Online</span>
            <span v-else>Offline</span>
          </p>
        </div>
        <UBadge
          :label="isRealtimeConnected ? 'Connected' : 'Connecting'"
          :color="isRealtimeConnected ? 'primary' : 'neutral'"
          variant="subtle"
        />
      </header>

      <div ref="conversationRef" class="flex-1 overflow-y-auto bg-muted/10">
        <div v-if="!selectedPartnerId" class="h-full flex flex-col items-center justify-center text-muted gap-3">
          <UIcon name="i-lucide-inbox" class="size-10" />
          <p>
            Select someone to start chatting.
          </p>
        </div>

        <div v-else class="p-4 space-y-3">
          <div v-if="loadingConversation" class="text-center text-sm text-muted py-4">
            Loading conversation…
          </div>
          <template v-else>
            <div
              v-for="message in activeMessages"
              :key="message.id"
              class="flex"
              :class="message.isMine ? 'justify-end' : 'justify-start'"
            >
              <div
                class="max-w-[80%] rounded-2xl px-4 py-2 text-sm"
                :class="message.isMine ? 'bg-primary text-white' : 'bg-white dark:bg-elevated/70 text-highlighted border border-default'"
              >
                <p class="whitespace-pre-line wrap-break-word">
                  {{ message.content }}
                </p>
                <span class="block text-[11px] mt-1 text-right" :class="message.isMine ? 'text-white/70' : 'text-muted'">
                  {{ formatTimestamp(message.createdAt) }}
                </span>
              </div>
            </div>
          </template>
        </div>
      </div>

      <footer class="border-t border-default p-4">
        <form class="flex flex-col gap-2" @submit.prevent="sendMessage">
          <UTextarea
            v-model="newMessage"
            placeholder="Type a message"
            :disabled="!selectedPartnerId || sendingMessage"
            :rows="2"
            @keydown="handleComposerKeydown"
          />
          <div class="flex items-center justify-end gap-2">
            <UButton
              type="submit"
              color="primary"
              :disabled="!selectedPartnerId"
              :loading="sendingMessage"
              icon="i-lucide-send"
              label="Send"
            />
          </div>
        </form>
      </footer>
    </section>
  </div>
</template>
