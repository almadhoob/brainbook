import { extractErrorMessage, formatDate, toDataUrl, initialsFromName, buildFullName } from './useGroupHelpers'

export interface ApiGroupMember {
  user_id?: number
  f_name?: string | null
  l_name?: string | null
  avatar?: string | null
  role?: string | null
  joined_at?: string | null
}

export interface GroupMember {
  id: number
  fullName: string
  initials: string
  role: string
  joinedAt: string
  avatarSrc?: string
}

export function useGroupMembers(apiBase: string, groupId: Ref<number | null>, isOwner: Ref<boolean>) {
  const toast = useToast()

  const membersCache = reactive<Record<number, GroupMember[]>>({})
  const membersLoading = reactive<Record<number, boolean>>({})

  const members = computed(() =>
    groupId.value ? membersCache[groupId.value] ?? [] : []
  )

  const isLoading = computed(() =>
    groupId.value ? membersLoading[groupId.value] ?? false : false
  )

  const ownerRequestResponse = reactive({
    requestId: '',
    action: 'accept' as 'accept' | 'decline',
    loading: false
  })

  async function loadMembers() {
    if (!groupId.value) return

    const gid = groupId.value
    membersLoading[gid] = true
    try {
      const response = await $fetch<{ members: ApiGroupMember[] }>(
        `${apiBase}/protected/v1/groups/${gid}/members`,
        { credentials: 'include' }
      )
      membersCache[gid] = normalizeMembers(response.members)
    } catch (error) {
      toast.add({
        title: 'Unable to load members',
        description: extractErrorMessage(error) || 'Please try again later.',
        color: 'error'
      })
    } finally {
      membersLoading[gid] = false
    }
  }

  async function respondGroupRequestById() {
    if (!isOwner.value || !groupId.value) return

    const requestId = Number.parseInt(ownerRequestResponse.requestId, 10)
    if (!Number.isInteger(requestId) || requestId <= 0) {
      toast.add({ title: 'Invalid request ID', color: 'error' })
      return
    }

    ownerRequestResponse.loading = true
    try {
      await $fetch(
        `${apiBase}/protected/v1/groups/${groupId.value}/requests/${requestId}`,
        {
          method: 'POST',
          credentials: 'include',
          body: { action: ownerRequestResponse.action }
        }
      )

      toast.add({
        title: 'Request updated',
        description: `Marked ${ownerRequestResponse.action}.`
      })
      ownerRequestResponse.requestId = ''
      await loadMembers()
    } catch (error) {
      toast.add({
        title: 'Unable to update request',
        description: extractErrorMessage(error) || 'Try again later.',
        color: 'error'
      })
    } finally {
      ownerRequestResponse.loading = false
    }
  }

  function normalizeMembers(members?: ApiGroupMember[]): GroupMember[] {
    if (!Array.isArray(members)) return []

    return members.map((member) => {
      const fullName = buildFullName(member.f_name, member.l_name)
      return {
        id: typeof member.user_id === 'number' ? member.user_id : -1,
        fullName,
        initials: initialsFromName(fullName),
        role: (member.role ?? 'member').toLowerCase(),
        joinedAt: formatDate(member.joined_at),
        avatarSrc: toDataUrl(member.avatar)
      }
    })
  }

  return {
    members,
    membersLoading: isLoading,
    ownerRequestResponse,
    loadMembers,
    respondGroupRequestById
  }
}
