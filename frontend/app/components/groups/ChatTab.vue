<script setup lang="ts">
import type { GroupChatMessage } from '~/composables/useGroupChat'

interface Props {
  messages: GroupChatMessage[]
  messagesLoading: boolean
  messageDraft: string
  messageSending: boolean
  currentUserId: number | null
  isOnline: boolean
}

interface Emits {
  (e: 'send-message'): void
  (e: 'update:messageDraft', value: string): void
}

const MAX_MESSAGE_LENGTH = 200

defineProps<Props>()
const emit = defineEmits<Emits>()

function handleKeydown(event: KeyboardEvent) {
  if (event.key === 'Enter') {
    event.preventDefault()
    emit('send-message')
  }
}

function handleInput(value: string) {
  const cleaned = value.replace(/\n/g, '').slice(0, MAX_MESSAGE_LENGTH)
  emit('update:messageDraft', cleaned)
}
</script>

<template>
  <div class="space-y-6">
    <section>
      <div class="flex items-center justify-between">
        <h3 class="text-lg font-semibold">
          Group chat
        </h3>
        <UBadge :color="isOnline ? 'primary' : 'neutral'" variant="subtle">
          {{ isOnline ? 'Online' : 'Offline' }}
        </UBadge>
      </div>
      <div class="mt-4 rounded-2xl border border-default/60 p-4 space-y-3 bg-elevated/20">
        <div class="h-64 overflow-y-auto rounded-xl border border-default/60 bg-white/60 dark:bg-elevated/80 p-3 space-y-3">
          <div v-if="messagesLoading" class="text-sm text-muted">
            Loading chat...
          </div>
          <div v-else-if="!messages.length" class="text-sm text-muted">
            No messages yet. Start the conversation!
          </div>
          <div v-else class="space-y-3">
            <div
              v-for="message in messages"
              :key="message.id"
              class="flex"
              :class="message.senderId === currentUserId ? 'justify-end' : 'justify-start'"
            >
              <div class="flex items-start gap-2 max-w-[80%]">
                <UAvatar
                  v-if="message.senderId !== currentUserId"
                  :src="message.avatarSrc"
                  :text="message.senderInitials"
                  size="xs"
                />
                <div
                  class="rounded-2xl px-3 py-2 text-sm border border-default/60"
                  :class="message.senderId === currentUserId ? 'bg-primary text-white border-primary/70' : 'bg-white dark:bg-elevated/70 text-highlighted'"
                >
                  <p class="font-medium text-xs">
                    {{ message.senderId === currentUserId ? 'You' : message.senderName }}
                  </p>
                  <p class="whitespace-pre-line break-words">
                    {{ message.content }}
                  </p>
                  <p class="mt-1 text-[11px]" :class="message.senderId === currentUserId ? 'text-white/70' : 'text-muted'">
                    {{ message.createdAtFormatted }}
                  </p>
                </div>
                <UAvatar
                  v-if="message.senderId === currentUserId"
                  :src="message.avatarSrc"
                  :text="message.senderInitials"
                  size="xs"
                />
              </div>
            </div>
          </div>
        </div>
        <div class="flex flex-col gap-2">
          <div class="relative">
            <UTextarea
              :model-value="messageDraft"
              placeholder="Send a message to the group (Press Enter to send)"
              :maxlength="MAX_MESSAGE_LENGTH"
              autoresize
              :rows="2"
              class="w-full"
              @update:model-value="handleInput"
              @keydown="handleKeydown"
            />
            <div class="absolute bottom-2 right-3 text-xs" :class="messageDraft.length > MAX_MESSAGE_LENGTH - 20 ? 'text-error' : 'text-muted'">
              {{ messageDraft.length }}/{{ MAX_MESSAGE_LENGTH }}
            </div>
          </div>
          <div class="flex justify-end">
            <UButton
              color="primary"
              :loading="messageSending"
              icon="i-lucide-send"
              :disabled="!messageDraft.trim() || messageDraft.length > MAX_MESSAGE_LENGTH"
              @click="emit('send-message')"
            >
              Send
            </UButton>
          </div>
        </div>
      </div>
    </section>
  </div>
</template>
