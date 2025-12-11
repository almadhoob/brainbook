import { createSharedComposable } from '@vueuse/core'
import type { UiNotification, WebsocketNotificationPayload } from '~/types'
import { useRealtime } from '~/composables/useRealtime'

const normalizeNotification = (payload: WebsocketNotificationPayload): UiNotification => ({
  id: payload.id,
  type: payload.type,
  isRead: payload.is_read,
  createdAt: payload.created_at,
  payload: payload.payload ?? null
})

const _useNotifications = () => {
  const runtimeConfig = useRuntimeConfig()
  const apiBase = runtimeConfig.public?.apiBase || 'http://localhost:8080'
  const notifications = ref<UiNotification[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)
  const bootstrapped = ref(false)

  const { notificationBus, connect } = useRealtime()

  const upsertNotification = (incoming: UiNotification) => {
    const index = notifications.value.findIndex(notification => notification.id === incoming.id)
    if (index === -1) {
      notifications.value = [incoming, ...notifications.value]
    } else {
      const cloned = [...notifications.value]
      cloned[index] = incoming
      notifications.value = cloned
    }
  }

  const fetchNotifications = async (force = false) => {
    if (!import.meta.client) return
    if (loading.value) return
    if (bootstrapped.value && !force) return

    loading.value = true
    try {
      const response = await $fetch<{ notifications: WebsocketNotificationPayload[] }>(`${apiBase}/protected/v1/notifications?all=1`, {
        credentials: 'include'
      })
      const normalized = (response.notifications ?? []).map(normalizeNotification)
      notifications.value = normalized.sort((a, b) => b.createdAt.localeCompare(a.createdAt))
      bootstrapped.value = true
      error.value = null
    } catch (err) {
      error.value = 'Unable to load notifications'
      console.error(err)
    } finally {
      loading.value = false
    }
  }

  const markAsRead = async (notificationId: number) => {
    if (!import.meta.client) return
    try {
      await $fetch(`${apiBase}/protected/v1/notifications/${notificationId}/read`, {
        method: 'POST',
        credentials: 'include'
      })
      const cloned = notifications.value.map(notification => notification.id === notificationId
        ? { ...notification, isRead: true }
        : notification)
      notifications.value = cloned
    } catch (err) {
      console.error('Failed to mark notification as read', err)
    }
  }

  if (import.meta.client) {
    connect()
    fetchNotifications()
    notificationBus.on((payload) => {
      upsertNotification(normalizeNotification(payload))
    })
  }

  const unreadCount = computed(() => notifications.value.filter(notification => !notification.isRead).length)

  return {
    notifications: readonly(notifications),
    unreadCount,
    loading: readonly(loading),
    error: readonly(error),
    fetchNotifications,
    markAsRead
  }
}

export const useNotifications = createSharedComposable(_useNotifications)
