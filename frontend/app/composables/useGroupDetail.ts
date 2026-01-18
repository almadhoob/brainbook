import { extractErrorMessage, formatDate } from './useGroupHelpers'

export interface ApiGroup {
  id?: number
  owner_id?: number
  title?: string | null
  description?: string | null
  created_at?: string | null
}

export interface GroupSummary {
  id: number
  ownerId: number | null
  title: string
  description: string
  createdAtRaw: string | null
  createdAtFormatted: string
}

export interface JoinState {
  status: string
  loading: boolean
}

export function useGroupDetail(apiBase: string) {
  const toast = useToast()
  const { session } = useSession()

  const selectedGroupId = ref<number | null>(null)
  const groupDetail = ref<GroupSummary | null>(null)
  const detailLoading = ref(false)
  const detailError = ref<string | null>(null)
  const joinRequestState = reactive<Record<number, JoinState>>({})

  const currentUserId = computed(() =>
    typeof session.value.user_id === 'number' ? session.value.user_id : null
  )

  const isOwner = computed(() =>
    selectedGroupId.value != null
    && groupDetail.value?.ownerId != null
    && currentUserId.value === groupDetail.value.ownerId
  )

  const joinStatus = computed(() => {
    if (selectedGroupId.value == null) return 'none'
    if (isOwner.value) return 'owner'
    return joinRequestState[selectedGroupId.value]?.status ?? 'not-requested'
  })

  const joinLoading = computed(() =>
    selectedGroupId.value != null
    && joinRequestState[selectedGroupId.value]?.loading === true
  )

  let detailRequestToken = 0

  async function loadGroupSummary(groupId: number) {
    const token = ++detailRequestToken
    detailLoading.value = true
    detailError.value = null

    try {
      const response = await $fetch<{ group: ApiGroup }>(
        `${apiBase}/protected/v1/groups/${groupId}`,
        { credentials: 'include' }
      )

      const summary = normalizeGroup(response.group)
      if (!summary) {
        throw new Error('Group not found')
      }

      if (token === detailRequestToken) {
        groupDetail.value = summary
      }

      return summary
    } catch (error) {
      if (token === detailRequestToken) {
        detailError.value = extractErrorMessage(error) || 'Unable to load group details.'
        groupDetail.value = null
      }
      throw error
    } finally {
      if (token === detailRequestToken) {
        detailLoading.value = false
      }
    }
  }

  async function submitJoinRequest() {
    if (selectedGroupId.value == null) return

    const groupId = selectedGroupId.value
    if (!joinRequestState[groupId]) {
      joinRequestState[groupId] = { status: 'not-requested', loading: false }
    }

    joinRequestState[groupId].loading = true
    try {
      const response = await $fetch<{ status?: string }>(
        `${apiBase}/protected/v1/groups/${groupId}/join`,
        { method: 'POST', credentials: 'include' }
      )

      const status = response.status ?? 'pending'
      joinRequestState[groupId].status = status

      if (status === 'member' || status === 'owner') {
        toast.add({ title: 'Joined group', description: 'Welcome aboard!' })
      } else if (status === 'pending') {
        toast.add({ title: 'Request sent', description: 'Waiting for the group owner to respond.' })
      } else {
        toast.add({ title: 'Request updated', description: `Status: ${status}` })
      }

      return status
    } catch (error) {
      toast.add({
        title: 'Unable to send request',
        description: extractErrorMessage(error) || 'Please try again later.',
        color: 'error'
      })
      throw error
    } finally {
      joinRequestState[groupId].loading = false
    }
  }

  function selectGroup(id: number) {
    if (selectedGroupId.value === id) return
    selectedGroupId.value = id
  }

  function normalizeGroup(group?: ApiGroup | null): GroupSummary | null {
    if (!group || typeof group.id !== 'number') return null

    const title = (group.title ?? '').trim() || 'Untitled group'
    return {
      id: group.id,
      ownerId: typeof group.owner_id === 'number' ? group.owner_id : null,
      title,
      description: (group.description ?? '').trim(),
      createdAtRaw: group.created_at ?? null,
      createdAtFormatted: formatDate(group.created_at)
    }
  }

  return {
    selectedGroupId,
    groupDetail,
    detailLoading,
    detailError,
    currentUserId,
    isOwner,
    joinStatus,
    joinLoading,
    loadGroupSummary,
    submitJoinRequest,
    selectGroup,
    normalizeGroup
  }
}
