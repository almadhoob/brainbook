<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { breakpointsTailwind } from '@vueuse/core'
import type { Message } from '~/types'

const tabItems = [{
  label: 'All',
  value: 'all'
}, {
  label: 'Unread',
  value: 'unread'
}]
const selectedTab = ref('all')

const { data: messages } = await useFetch<Message[]>('/api/messages', { default: () => [] })

// Filter messages based on the selected tab
const filteredMessages = computed(() => {
  if (selectedTab.value === 'unread') {
    return messages.value.filter(message => !!message.unread)
  }

  return messages.value
})

const selectedMessage = ref<Message | null>()

const isMessagePanelOpen = computed({
  get() {
    return !!selectedMessage.value
  },
  set(value: boolean) {
    if (!value) {
      selectedMessage.value = null
    }
  }
})

// Reset selected message if it's not in the filtered messages
watch(filteredMessages, () => {
  if (!filteredMessages.value.find(message => message.id === selectedMessage.value?.id)) {
    selectedMessage.value = null
  }
})

const breakpoints = useBreakpoints(breakpointsTailwind)
const isMobile = breakpoints.smaller('lg')
</script>

<template>
  <UDashboardPanel
    id="inbox-1"
    :default-size="25"
    :min-size="20"
    :max-size="30"
    resizable
  >
    <UDashboardNavbar title="Messages">
      <template #leading>
        <UDashboardSidebarCollapse />
      </template>
      <template #trailing>
        <UBadge :label="filteredMessages.length" variant="subtle" />
      </template>

      <template #right>
        <UTabs
          v-model="selectedTab"
          :items="tabItems"
          :content="false"
          size="xs"
        />
      </template>
    </UDashboardNavbar>
    <InboxList v-model="selectedMessage" :messages="filteredMessages" />
  </UDashboardPanel>

  <InboxMail v-if="selectedMessage" :message="selectedMessage" @close="selectedMessage = null" />
  <div v-else class="hidden lg:flex flex-1 items-center justify-center">
    <UIcon name="i-lucide-inbox" class="size-32 text-dimmed" />
  </div>

  <ClientOnly>
    <USlideover v-if="isMobile" v-model:open="isMessagePanelOpen">
      <template #content>
        <InboxMail v-if="selectedMessage" :message="selectedMessage" @close="selectedMessage = null" />
      </template>
    </USlideover>
  </ClientOnly>
</template>
