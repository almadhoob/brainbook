<script setup lang="ts">
import { format, isToday } from 'date-fns'
import type { Message } from '~/types'

const props = defineProps<{
  messages: Message[]
}>()

const messagesRefs = ref<Element[]>([])
const selectedMessage = defineModel<Message | null>()

watch(selectedMessage, () => {
  if (!selectedMessage.value) {
    return
  }
  const ref = messagesRefs.value[selectedMessage.value.id]
  if (ref) {
    ref.scrollIntoView({ block: 'nearest' })
  }
})

defineShortcuts({
  arrowdown: () => {
    const index = props.messages.findIndex(message => message.id === selectedMessage.value?.id)

    if (index === -1) {
      selectedMessage.value = props.messages[0]
    } else if (index < props.messages.length - 1) {
      selectedMessage.value = props.messages[index + 1]
    }
  },
  arrowup: () => {
    const index = props.messages.findIndex(message => message.id === selectedMessage.value?.id)
    if (index === -1) {
      selectedMessage.value = props.messages[props.messages.length - 1]
    } else if (index > 0) {
      selectedMessage.value = props.messages[index - 1]
    }
  }
})
</script>

<template>
  <div class="overflow-y-auto divide-y divide-default">
    <div
      v-for="(message, index) in messages"
      :key="index"
      :ref="el => { messagesRefs[message.id] = el as Element }"
    >
      <div
        class="p-4 sm:px-6 text-sm cursor-pointer border-l-2 transition-colors"
        :class="[
          message.unread ? 'text-highlighted' : 'text-toned',
          selectedMessage && selectedMessage.id === message.id ? 'border-primary bg-primary/10' : 'border-bg hover:border-primary hover:bg-primary/5'
        ]"
        @click="selectedMessage = message"
      >
        <div class="flex items-center justify-between" :class="[message.unread && 'font-semibold']">
          <div class="flex items-center gap-3">
            {{ message.from.name }}

            <UChip v-if="message.unread" />
          </div>

          <span>{{ isToday(new Date(message.date)) ? format(new Date(message.date), 'HH:mm') : format(new Date(message.date), 'dd MMM') }}</span>
        </div>
        <p class="truncate" :class="[message.unread && 'font-semibold']">
          {{ message.subject }}
        </p>
        <p class="text-dimmed line-clamp-1">
          {{ message.body }}
        </p>
      </div>
    </div>
  </div>
</template>
