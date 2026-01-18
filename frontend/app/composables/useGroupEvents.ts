import { extractErrorMessage, formatDate, toSqlDateTime } from './useGroupHelpers'

export interface ApiGroupEvent {
  id?: number
  title?: string | null
  description?: string | null
  time?: string | null
  interested?: number | null
  not_interested?: number | null
}

export interface GroupEventItem {
  id: number
  title: string
  description: string
  timeRaw: string
  formattedTime: string
  goingCount: number
  notGoingCount: number
}

export function useGroupEvents(apiBase: string, groupId: Ref<number | null>, isOwner: Ref<boolean>) {
  const toast = useToast()

  const eventsCache = reactive<Record<number, GroupEventItem[]>>({})
  const eventsLoading = reactive<Record<number, boolean>>({})
  const createEventLoading = ref(false)
  const rsvpLoading = reactive<Record<number, boolean>>({})

  const events = computed(() =>
    groupId.value ? eventsCache[groupId.value] ?? [] : []
  )

  const isLoading = computed(() =>
    groupId.value ? eventsLoading[groupId.value] ?? false : false
  )

  const newEventForm = reactive({
    title: '',
    description: '',
    time: ''
  })

  async function loadEvents() {
    if (!groupId.value) return

    const gid = groupId.value
    eventsLoading[gid] = true
    try {
      const response = await $fetch<{ events: ApiGroupEvent[] }>(
        `${apiBase}/protected/v1/groups/${gid}/events`,
        { credentials: 'include' }
      )
      eventsCache[gid] = normalizeEvents(response.events)
    } catch (error) {
      toast.add({
        title: 'Unable to load events',
        description: extractErrorMessage(error) || 'Please try again later.',
        color: 'error'
      })
    } finally {
      eventsLoading[gid] = false
    }
  }

  async function createEvent() {
    if (!isOwner.value || !groupId.value) return

    const title = newEventForm.title.trim()
    const time = newEventForm.time.trim()

    if (!title || !time) {
      toast.add({
        title: 'Event details required',
        description: 'Title and time are mandatory.',
        color: 'error'
      })
      return
    }

    createEventLoading.value = true
    try {
      await $fetch(`${apiBase}/protected/v1/groups/${groupId.value}/events`, {
        method: 'POST',
        credentials: 'include',
        body: {
          title,
          description: newEventForm.description.trim(),
          time: toSqlDateTime(time)
        }
      })

      toast.add({ title: 'Event scheduled' })
      newEventForm.title = ''
      newEventForm.description = ''
      newEventForm.time = ''
      await loadEvents()
    } catch (error) {
      toast.add({
        title: 'Unable to schedule',
        description: extractErrorMessage(error) || 'Please try again later.',
        color: 'error'
      })
    } finally {
      createEventLoading.value = false
    }
  }

  async function respondRsvp(eventId: number, response: 'going' | 'not_going') {
    if (!groupId.value) return

    rsvpLoading[eventId] = true
    try {
      await $fetch(
        `${apiBase}/protected/v1/groups/${groupId.value}/events/${eventId}/rsvp`,
        {
          method: 'POST',
          credentials: 'include',
          body: { response }
        }
      )

      toast.add({
        title: response === 'going' ? 'RSVP saved' : 'RSVP updated',
        description: response === 'going' ? 'See you there!' : 'You marked not going.'
      })
      await loadEvents()
    } catch (error) {
      toast.add({
        title: 'Unable to RSVP',
        description: extractErrorMessage(error) || 'Try again later.',
        color: 'error'
      })
    } finally {
      rsvpLoading[eventId] = false
    }
  }

  function clearEventForm() {
    newEventForm.title = ''
    newEventForm.description = ''
    newEventForm.time = ''
  }

  function normalizeEvents(events?: ApiGroupEvent[]): GroupEventItem[] {
    if (!Array.isArray(events)) return []

    return events.map(event => ({
      id: typeof event.id === 'number' ? event.id : Math.random(),
      title: (event.title ?? '').trim() || 'Untitled event',
      description: (event.description ?? '').trim(),
      timeRaw: event.time ?? '',
      formattedTime: formatDate(event.time),
      goingCount: typeof event.interested === 'number' ? event.interested : 0,
      notGoingCount: typeof event.not_interested === 'number' ? event.not_interested : 0
    }))
  }

  return {
    events,
    eventsLoading: isLoading,
    createEventLoading,
    rsvpLoading,
    newEventForm,
    loadEvents,
    createEvent,
    respondRsvp,
    clearEventForm
  }
}
