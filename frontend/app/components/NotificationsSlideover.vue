<script setup lang="ts">
import { formatTimeAgo } from '@vueuse/core'
import type { RouteLocationRaw } from 'vue-router'
import type { UiNotification } from '~/types'

const router = useRouter()
const { isNotificationsSlideoverOpen } = useDashboard()
const { notifications, loading, markAsRead } = useNotifications()

interface NotificationDescriptor {
  title: string
  description: string
  icon: string
  accent: string
  navigateTo?: RouteLocationRaw
}

const readString = (payload: Record<string, unknown>, key: string) => {
  const value = payload[key]
  return typeof value === 'string' ? value : undefined
}

const readNumber = (payload: Record<string, unknown>, key: string) => {
  const value = payload[key]
  if (typeof value === 'number') {
    return value
  }
  if (typeof value === 'string') {
    const parsed = Number(value)
    return Number.isNaN(parsed) ? undefined : parsed
  }
  return undefined
}

const describeNotification = (notification: UiNotification): NotificationDescriptor => {
  const payload = notification.payload ?? {}

  switch (notification.type) {
    case 'direct_message': {
      const senderName = readString(payload, 'sender_name') ?? 'Direct message'
      const directMessage = readString(payload, 'message') ?? 'Sent you a private message'
      const senderId = readNumber(payload, 'sender_id')
      const directNavigate = typeof senderId === 'number'
        ? { path: '/messages', query: { user: String(senderId) } }
        : undefined
      return {
        title: senderName,
        description: directMessage,
        icon: 'i-lucide-message-circle',
        accent: notification.isRead ? 'text-muted' : 'text-primary',
        navigateTo: directNavigate
      }
    }
    case 'group_message': {
      const groupTitle = readString(payload, 'group_title')
      const groupMessage = readString(payload, 'message') ?? 'New activity in one of your groups'
      const groupId = readNumber(payload, 'group_id')
      const groupNavigate = typeof groupId === 'number'
        ? { path: '/groups', query: { id: String(groupId) } }
        : undefined
      return {
        title: groupTitle ? `Group · ${groupTitle}` : 'Group message',
        description: groupMessage,
        icon: 'i-lucide-users-round',
        accent: notification.isRead ? 'text-muted' : 'text-emerald-500',
        navigateTo: groupNavigate
      }
    }
    default: {
      const genericMessage = readString(payload, 'message') ?? 'You have a new notification'
      return {
        title: 'Notification',
        description: genericMessage,
        icon: 'i-lucide-bell',
        accent: notification.isRead ? 'text-muted' : 'text-amber-500'
      }
    }
  }
}

const decoratedNotifications = computed(() => notifications.value.map(notification => ({
  notification,
  meta: describeNotification(notification)
})))

const handleNotificationClick = async (notification: UiNotification) => {
  const descriptor = describeNotification(notification)
  if (descriptor.navigateTo) {
    await router.push(descriptor.navigateTo)
    isNotificationsSlideoverOpen.value = false
  }

  if (!notification.isRead) {
    await markAsRead(notification.id)
  }
}

const markSingleAsRead = async (notification: UiNotification, event: MouseEvent) => {
  event.stopPropagation()
  if (!notification.isRead) {
    await markAsRead(notification.id)
  }
}
</script>

<template>
  <USlideover
    v-model:open="isNotificationsSlideoverOpen"
    title="Notifications"
    description="Latest alerts, direct messages, and group updates"
  >
    <template #body>
      <div v-if="loading" class="py-12 flex items-center justify-center text-muted">
        <UIcon name="i-lucide-loader" class="size-6 animate-spin" />
      </div>

      <div v-else-if="!decoratedNotifications.length" class="py-10 text-center text-sm text-muted">
        <UIcon name="i-lucide-bell" class="size-8 mx-auto mb-3" />
        You are all caught up.
      </div>

      <div v-else class="space-y-2">
        <button
          v-for="entry in decoratedNotifications"
          :key="entry.notification.id"
          type="button"
          class="w-full rounded-lg border border-default px-3 py-2 text-left hover:border-primary/40 hover:bg-primary/5 transition"
          :class="!entry.notification.isRead && 'bg-primary/5 border-primary/50'"
          @click="handleNotificationClick(entry.notification)"
        >
          <div class="flex items-start gap-3">
            <span class="rounded-full bg-default/60 p-2" :class="entry.meta.accent">
              <UIcon :name="entry.meta.icon" class="size-5" />
            </span>

            <div class="flex-1 min-w-0">
              <div class="flex items-start justify-between gap-2">
                <p class="font-medium text-highlighted truncate">
                  {{ entry.meta.title }}
                </p>
                <time
                  :datetime="entry.notification.createdAt"
                  class="text-xs text-muted whitespace-nowrap"
                >
                  {{ formatTimeAgo(new Date(entry.notification.createdAt)) }}
                </time>
              </div>
              <p class="text-sm text-muted line-clamp-2">
                {{ entry.meta.description }}
              </p>
            </div>

            <UButton
              v-if="!entry.notification.isRead"
              icon="i-lucide-check"
              color="neutral"
              variant="ghost"
              size="xs"
              @click="markSingleAsRead(entry.notification, $event)"
            />
          </div>
        </button>
      </div>
    </template>
  </USlideover>
</template>
