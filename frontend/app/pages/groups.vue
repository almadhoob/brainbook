<script setup lang="ts">
const toast = useToast()
const runtimeConfig = useRuntimeConfig()
const apiBase = typeof runtimeConfig.public?.apiBase === 'string' && runtimeConfig.public.apiBase.length > 0
  ? runtimeConfig.public.apiBase
  : 'http://localhost:8080'

type TabKey = 'all' | 'mine'

interface ApiGroup {
  id?: number
  owner_id?: number
  title?: string | null
  description?: string | null
  created_at?: string | null
}

interface ApiGroupMember {
  user_id?: number
  f_name?: string | null
  l_name?: string | null
  avatar?: string | null
  role?: string | null
  joined_at?: string | null
}

interface ApiGroupPost {
  id?: number
  group_id?: number
  user_id?: number
  f_name?: string | null
  l_name?: string | null
  avatar?: string | null
  content?: string | null
  file?: string | null
  created_at?: string | null
  comment_count?: number | null
}

interface ApiGroupComment {
  id?: number
  user_id?: number
  f_name?: string | null
  l_name?: string | null
  avatar?: string | null
  content?: string | null
  created_at?: string | null
}

interface ApiGroupEvent {
  id?: number
  title?: string | null
  description?: string | null
  time?: string | null
}

interface GroupSummary {
  id: number
  ownerId: number | null
  title: string
  description: string
  createdAtRaw: string | null
  createdAtFormatted: string
}

interface GroupMember {
  id: number
  fullName: string
  initials: string
  role: string
  joinedAt: string
  avatarSrc?: string
}

interface GroupEventItem {
  id: number
  title: string
  description: string
  timeRaw: string
  formattedTime: string
}

interface GroupPostItem {
  id: number
  content: string
  createdAtRaw: string | null
  formattedCreatedAt: string
  authorName: string
  authorInitials: string
  avatarSrc?: string
  mediaSrc?: string
  commentCount: number
}

interface GroupComment {
  id: number
  content: string
  formattedCreatedAt: string
  authorName: string
  authorInitials: string
  avatarSrc?: string
}

interface GroupDetailState {
  summary: GroupSummary
  members: GroupMember[]
  events: GroupEventItem[]
  posts: GroupPostItem[]
}

interface JoinState {
  status: string
  loading: boolean
}

const tabOptions: { label: string, value: TabKey }[] = [
  { label: 'All groups', value: 'all' },
  { label: 'My groups', value: 'mine' }
]

const activeTab = ref<TabKey>('all')
const searchQuery = ref('')
const selectedGroupId = ref<number | null>(null)
const groupDetail = ref<GroupDetailState | null>(null)
const detailLoading = ref(false)
const detailError = ref<string | null>(null)
const refreshingGroups = ref(false)

const createGroupModalOpen = ref(false)
const createGroupForm = reactive({
  title: '',
  description: ''
})
const createGroupErrors = reactive({
  title: '',
  description: ''
})
const createGroupLoading = ref(false)

const newPostForm = reactive({
  content: '',
  file: null as string | null
})
const createPostLoading = ref(false)
const postFileInput = ref<HTMLInputElement | null>(null)

const newEventForm = reactive({
  title: '',
  description: '',
  time: ''
})
const createEventLoading = ref(false)

const commentsCache = reactive<Record<number, GroupComment[]>>({})
const commentsLoading = reactive<Record<number, boolean>>({})
const newCommentDrafts = reactive<Record<number, string>>({})
const commentSubmitting = reactive<Record<number, boolean>>({})
const expandedPosts = ref(new Set<number>())

const joinRequestState = reactive<Record<number, JoinState>>({})

const currentUserId = ref<number | null>(null)

function hydrateCurrentUserId() {
  if (typeof window === 'undefined') return
  const stored = window.localStorage.getItem('user_id')
  if (!stored) return
  const parsed = Number.parseInt(stored, 10)
  if (!Number.isNaN(parsed)) {
    currentUserId.value = parsed
  }
}

if (import.meta.client) {
  hydrateCurrentUserId()
}

onMounted(() => {
  hydrateCurrentUserId()
})

const {
  data: allGroupsData,
  status: allGroupsStatus,
  error: allGroupsError,
  refresh: refreshAllGroups
} = await useFetch<{ groups: ApiGroup[] }>(`${apiBase}/protected/v1/groups`, {
  credentials: 'include',
  lazy: true
})

const {
  data: myGroupsData,
  status: myGroupsStatus,
  error: myGroupsError,
  refresh: refreshMyGroups
} = await useFetch<{ groups: ApiGroup[] }>(`${apiBase}/protected/v1/user/groups`, {
  credentials: 'include',
  lazy: true
})

const combinedGroupsError = computed(() => allGroupsError.value ?? myGroupsError.value ?? null)

const normalizedAllGroups = computed(() => normalizeGroupList(allGroupsData.value?.groups))
const normalizedMyGroups = computed(() => normalizeGroupList(myGroupsData.value?.groups))

const membershipIds = computed(() => new Set(normalizedMyGroups.value.map(group => group.id)))

const fallbackGroupMap = computed(() => {
  const map = new Map<number, GroupSummary>()
  for (const group of normalizedAllGroups.value) {
    map.set(group.id, group)
  }
  for (const group of normalizedMyGroups.value) {
    map.set(group.id, group)
  }
  return map
})

const selectedSummary = computed(() => {
  if (selectedGroupId.value == null) {
    return null
  }
  if (groupDetail.value && groupDetail.value.summary.id === selectedGroupId.value) {
    return groupDetail.value.summary
  }
  return fallbackGroupMap.value.get(selectedGroupId.value) ?? null
})

const listLoading = computed(() => (activeTab.value === 'all' ? allGroupsStatus.value === 'pending' : myGroupsStatus.value === 'pending') || refreshingGroups.value)

const displayedGroups = computed(() => {
  const base = activeTab.value === 'all' ? normalizedAllGroups.value : normalizedMyGroups.value
  if (!searchQuery.value.trim()) {
    return base
  }
  const query = searchQuery.value.trim().toLowerCase()
  return base.filter((group) => {
    const haystack = `${group.title} ${group.description}`.toLowerCase()
    return haystack.includes(query)
  })
})

const isMember = computed(() => selectedGroupId.value != null && membershipIds.value.has(selectedGroupId.value))
const isOwner = computed(() => selectedGroupId.value != null && selectedSummary.value?.ownerId != null && currentUserId.value === selectedSummary.value.ownerId)
const showMemberContent = computed(() => isMember.value || isOwner.value)

const joinStatus = computed(() => {
  if (selectedGroupId.value == null) {
    return 'none'
  }
  if (isOwner.value) {
    return 'owner'
  }
  if (isMember.value) {
    return 'member'
  }
  return joinRequestState[selectedGroupId.value]?.status ?? 'not-requested'
})

const joinLoading = computed(() => selectedGroupId.value != null && joinRequestState[selectedGroupId.value]?.loading === true)

watch(
  () => [normalizedMyGroups.value, normalizedAllGroups.value],
  ([mine = [], all = []]) => {
    if (selectedGroupId.value != null) {
      const stillExists = mine.some(group => group.id === selectedGroupId.value) || all.some(group => group.id === selectedGroupId.value)
      if (stillExists) {
        return
      }
    }

    const fallback = mine[0]?.id ?? all[0]?.id ?? null
    if (fallback != null) {
      selectedGroupId.value = fallback
    } else {
      selectedGroupId.value = null
      groupDetail.value = null
    }
  },
  { immediate: true }
)

let detailRequestToken = 0

watch(selectedGroupId, (groupId) => {
  expandedPosts.value = new Set<number>()
  for (const key of Object.keys(commentsCache)) {
    commentsCache[Number(key)] = undefined as unknown as GroupComment[]
  }
  for (const key of Object.keys(commentsLoading)) {
    commentsLoading[Number(key)] = false
  }
  for (const key of Object.keys(newCommentDrafts)) {
    newCommentDrafts[Number(key)] = ''
  }
  for (const key of Object.keys(commentSubmitting)) {
    commentSubmitting[Number(key)] = false
  }
  detailError.value = null
  newPostForm.content = ''
  newPostForm.file = null
  newEventForm.title = ''
  newEventForm.description = ''
  newEventForm.time = ''

  if (groupId == null) {
    groupDetail.value = null
    return
  }

  void loadGroupDetail(groupId)
})

async function loadGroupDetail(groupId: number) {
  const token = ++detailRequestToken
  detailLoading.value = true
  detailError.value = null
  try {
    const groupResponse = await $fetch<{ group: ApiGroup }>(`${apiBase}/protected/v1/groups/${groupId}`, {
      credentials: 'include'
    })

    const summary = normalizeGroup(groupResponse.group)
    if (!summary) {
      throw new Error('Group not found')
    }

    let members: GroupMember[] = []
    let events: GroupEventItem[] = []
    let posts: GroupPostItem[] = []

    if (membershipIds.value.has(groupId) || (summary.ownerId != null && summary.ownerId === currentUserId.value)) {
      const [membersResult, eventsResult, postsResult] = await Promise.allSettled([
        $fetch<{ members: ApiGroupMember[] }>(`${apiBase}/protected/v1/groups/${groupId}/members`, {
          credentials: 'include'
        }),
        $fetch<{ events: ApiGroupEvent[] }>(`${apiBase}/protected/v1/groups/${groupId}/events`, {
          credentials: 'include'
        }),
        $fetch<{ posts: ApiGroupPost[] }>(`${apiBase}/protected/v1/groups/${groupId}/posts`, {
          credentials: 'include'
        })
      ])

      if (membersResult.status === 'fulfilled') {
        members = normalizeMembers(membersResult.value.members)
      }
      if (eventsResult.status === 'fulfilled') {
        events = normalizeEvents(eventsResult.value.events)
      }
      if (postsResult.status === 'fulfilled') {
        posts = normalizePosts(postsResult.value.posts)
      }
    }

    if (token === detailRequestToken) {
      groupDetail.value = {
        summary,
        members,
        events,
        posts
      }
    }
  } catch (error) {
    if (token === detailRequestToken) {
      detailError.value = extractErrorMessage(error) || 'Unable to load group details.'
      groupDetail.value = null
    }
  } finally {
    if (token === detailRequestToken) {
      detailLoading.value = false
    }
  }
}

async function refreshGroups() {
  refreshingGroups.value = true
  try {
    await Promise.all([refreshAllGroups(), refreshMyGroups()])
    if (selectedGroupId.value != null) {
      await loadGroupDetail(selectedGroupId.value)
    }
  } catch (error) {
    toast.add({
      title: 'Refresh failed',
      description: extractErrorMessage(error) || 'Unable to refresh groups right now.',
      color: 'error'
    })
  } finally {
    refreshingGroups.value = false
  }
}

function selectGroup(id: number) {
  if (selectedGroupId.value === id) {
    return
  }
  selectedGroupId.value = id
}

async function submitCreateGroup() {
  createGroupErrors.title = ''
  createGroupErrors.description = ''

  const title = createGroupForm.title.trim()
  const description = createGroupForm.description.trim()

  if (!title) {
    createGroupErrors.title = 'Title is required.'
  }
  if (!description) {
    createGroupErrors.description = 'Description is required.'
  }
  if (createGroupErrors.title || createGroupErrors.description) {
    return
  }

  try {
    createGroupLoading.value = true
    await $fetch(`${apiBase}/protected/v1/groups`, {
      method: 'POST',
      credentials: 'include',
      body: {
        title,
        description
      }
    })
    toast.add({ title: 'Group created', description: 'You can now invite people to your group.' })
    createGroupModalOpen.value = false
    resetCreateGroupForm()
    await refreshGroups()
  } catch (error) {
    toast.add({
      title: 'Unable to create group',
      description: extractErrorMessage(error) || 'Please try again later.',
      color: 'error'
    })
  } finally {
    createGroupLoading.value = false
  }
}

function resetCreateGroupForm() {
  createGroupForm.title = ''
  createGroupForm.description = ''
  createGroupErrors.title = ''
  createGroupErrors.description = ''
}

watch(() => createGroupModalOpen.value, (open) => {
  if (!open) {
    resetCreateGroupForm()
  }
})

async function submitJoinRequest() {
  if (selectedGroupId.value == null) {
    return
  }
  const groupId = selectedGroupId.value
  if (!joinRequestState[groupId]) {
    joinRequestState[groupId] = { status: 'not-requested', loading: false }
  }

  joinRequestState[groupId].loading = true
  try {
    const response = await $fetch<{ status?: string }>(`${apiBase}/protected/v1/groups/${groupId}/join`, {
      method: 'POST',
      credentials: 'include'
    })
    const status = response.status ?? 'pending'
    joinRequestState[groupId].status = status
    if (status === 'member' || status === 'owner') {
      await refreshGroups()
      toast.add({ title: 'Joined group', description: 'Welcome aboard!' })
    } else if (status === 'pending') {
      toast.add({ title: 'Request sent', description: 'Waiting for the group owner to respond.' })
    } else {
      toast.add({ title: 'Request updated', description: `Status: ${status}` })
    }
  } catch (error) {
    toast.add({
      title: 'Unable to send request',
      description: extractErrorMessage(error) || 'Please try again later.',
      color: 'error'
    })
  } finally {
    joinRequestState[groupId].loading = false
  }
}

function openPostFilePicker() {
  postFileInput.value?.click()
}

function handlePostFileChange(event: Event) {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]
  if (!file) {
    newPostForm.file = null
    return
  }
  fileToBase64(file)
    .then((base64) => {
      newPostForm.file = base64
    })
    .catch(() => {
      toast.add({ title: 'Unable to process file', color: 'error' })
      newPostForm.file = null
    })
}

async function submitGroupPost() {
  if (!showMemberContent.value || selectedGroupId.value == null) {
    return
  }
  const content = newPostForm.content.trim()
  if (!content) {
    toast.add({ title: 'Post content required', description: 'Share something with your group first.', color: 'error' })
    return
  }

  try {
    createPostLoading.value = true
    await $fetch(`${apiBase}/protected/v1/groups/${selectedGroupId.value}/create`, {
      method: 'POST',
      credentials: 'include',
      body: {
        content,
        file: newPostForm.file
      }
    })
    toast.add({ title: 'Post published', description: 'Your group can see it now.' })
    newPostForm.content = ''
    newPostForm.file = null
    await loadGroupDetail(selectedGroupId.value)
  } catch (error) {
    toast.add({
      title: 'Unable to publish',
      description: extractErrorMessage(error) || 'Please try again later.',
      color: 'error'
    })
  } finally {
    createPostLoading.value = false
  }
}

function togglePostComments(postId: number) {
  const next = new Set(expandedPosts.value)
  if (next.has(postId)) {
    next.delete(postId)
  } else {
    next.add(postId)
    if (!commentsCache[postId]) {
      void loadPostComments(postId)
    }
  }
  expandedPosts.value = next
}

function isPostExpanded(postId: number) {
  return expandedPosts.value.has(postId)
}

async function loadPostComments(postId: number) {
  if (!showMemberContent.value || selectedGroupId.value == null) {
    return
  }
  commentsLoading[postId] = true
  try {
    const response = await $fetch<{ comments: ApiGroupComment[] }>(`${apiBase}/protected/v1/groups/${selectedGroupId.value}/posts/${postId}/comments`, {
      credentials: 'include'
    })
    commentsCache[postId] = normalizeComments(response.comments)
  } catch (error) {
    toast.add({ title: 'Unable to load comments', description: extractErrorMessage(error) || 'Please try again later.', color: 'error' })
  } finally {
    commentsLoading[postId] = false
  }
}

async function submitComment(postId: number) {
  if (!showMemberContent.value || selectedGroupId.value == null) {
    return
  }
  const draft = newCommentDrafts[postId]?.trim()
  if (!draft) {
    toast.add({ title: 'Comment required', color: 'error' })
    return
  }
  commentSubmitting[postId] = true
  try {
    await $fetch(`${apiBase}/protected/v1/groups/${selectedGroupId.value}/posts/${postId}/comments`, {
      method: 'POST',
      credentials: 'include',
      body: { content: draft }
    })
    newCommentDrafts[postId] = ''
    const targetPost = groupDetail.value?.posts.find(post => post.id === postId)
    if (targetPost) {
      targetPost.commentCount += 1
    }
    await loadPostComments(postId)
    toast.add({ title: 'Comment added' })
  } catch (error) {
    toast.add({ title: 'Unable to comment', description: extractErrorMessage(error) || 'Please try again later.', color: 'error' })
  } finally {
    commentSubmitting[postId] = false
  }
}

async function submitEvent() {
  if (!isOwner.value || selectedGroupId.value == null) {
    return
  }
  const title = newEventForm.title.trim()
  const time = newEventForm.time.trim()
  if (!title || !time) {
    toast.add({ title: 'Event details required', description: 'Title and time are mandatory.', color: 'error' })
    return
  }
  try {
    createEventLoading.value = true
    await $fetch(`${apiBase}/protected/v1/groups/${selectedGroupId.value}/events`, {
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
    await loadGroupDetail(selectedGroupId.value)
  } catch (error) {
    toast.add({ title: 'Unable to schedule', description: extractErrorMessage(error) || 'Please try again later.', color: 'error' })
  } finally {
    createEventLoading.value = false
  }
}

const anyGroupsLoaded = computed(() => normalizedAllGroups.value.length > 0 || normalizedMyGroups.value.length > 0)

function extractErrorMessage(error: unknown): string {
  if (!error) return ''
  if (typeof error === 'string') return error
  if (error instanceof Error) return error.message
  if (typeof error === 'object') {
    const data = (error as { data?: Record<string, unknown>, message?: string }).data
    if (data) {
      if (typeof data.Error === 'string') return data.Error
      if (typeof data.error === 'string') return data.error
      if (typeof data.message === 'string') return data.message
    }
    if (typeof (error as { message?: string }).message === 'string') {
      return (error as { message: string }).message
    }
  }
  return ''
}

function normalizeGroupList(groups?: ApiGroup[]): GroupSummary[] {
  if (!Array.isArray(groups)) {
    return []
  }
  return groups
    .map(group => normalizeGroup(group))
    .filter((group): group is GroupSummary => Boolean(group))
    .sort((a, b) => a.title.localeCompare(b.title))
}

function normalizeGroup(group?: ApiGroup | null): GroupSummary | null {
  if (!group || typeof group.id !== 'number') {
    return null
  }
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

function normalizeMembers(members?: ApiGroupMember[]): GroupMember[] {
  if (!Array.isArray(members)) {
    return []
  }
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

function normalizeEvents(events?: ApiGroupEvent[]): GroupEventItem[] {
  if (!Array.isArray(events)) {
    return []
  }
  return events.map((event) => {
    return {
      id: typeof event.id === 'number' ? event.id : Math.random(),
      title: (event.title ?? '').trim() || 'Untitled event',
      description: (event.description ?? '').trim(),
      timeRaw: event.time ?? '',
      formattedTime: formatDate(event.time)
    }
  })
}

function normalizePosts(posts?: ApiGroupPost[]): GroupPostItem[] {
  if (!Array.isArray(posts)) {
    return []
  }
  return posts.map((post) => {
    const authorName = buildFullName(post.f_name, post.l_name) || 'Unknown member'
    return {
      id: typeof post.id === 'number' ? post.id : Math.random(),
      content: (post.content ?? '').trim(),
      createdAtRaw: post.created_at ?? null,
      formattedCreatedAt: formatDate(post.created_at),
      authorName,
      authorInitials: initialsFromName(authorName),
      avatarSrc: toDataUrl(post.avatar),
      mediaSrc: toDataUrl(post.file, 'image/jpeg'),
      commentCount: typeof post.comment_count === 'number' ? post.comment_count : 0
    }
  })
}

function normalizeComments(comments?: ApiGroupComment[]): GroupComment[] {
  if (!Array.isArray(comments)) {
    return []
  }
  return comments.map((comment) => {
    const authorName = buildFullName(comment.f_name, comment.l_name) || 'Unknown member'
    return {
      id: typeof comment.id === 'number' ? comment.id : Math.random(),
      content: (comment.content ?? '').trim(),
      formattedCreatedAt: formatDate(comment.created_at),
      authorName,
      authorInitials: initialsFromName(authorName),
      avatarSrc: toDataUrl(comment.avatar)
    }
  })
}

function buildFullName(first?: string | null, last?: string | null) {
  return `${(first ?? '').trim()} ${(last ?? '').trim()}`.trim()
}

function initialsFromName(name: string) {
  if (!name) {
    return '??'
  }
  return name
    .split(/\s+/)
    .filter(Boolean)
    .slice(0, 2)
    .map(part => part[0]?.toUpperCase() ?? '')
    .join('') || '??'
}

function formatDate(value?: string | null) {
  if (!value) {
    return 'Unknown'
  }
  const normalized = value.includes('T') ? value : value.replace(' ', 'T')
  const parsed = new Date(normalized)
  if (Number.isNaN(parsed.getTime())) {
    return value
  }
  return parsed.toLocaleString()
}

function toDataUrl(raw?: string | null, mime = 'image/png') {
  if (!raw) {
    return undefined
  }
  const trimmed = raw.trim()
  if (!trimmed) {
    return undefined
  }
  if (trimmed.startsWith('data:')) {
    return trimmed
  }
  return `data:${mime};base64,${trimmed}`
}

function toSqlDateTime(localValue: string) {
  if (!localValue) {
    return ''
  }
  if (localValue.includes('T')) {
    const [date, time] = localValue.split('T')
    const safeTime = time ?? ''
    if (!safeTime.includes(':')) {
      return `${date} ${safeTime}:00`
    }
    return `${date} ${safeTime.length === 5 ? `${safeTime}:00` : safeTime}`.replace('Z', '')
  }
  return localValue
}

function fileToBase64(file: File) {
  return new Promise<string>((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => {
      const result = typeof reader.result === 'string' ? reader.result : ''
      const payload = result.includes(',') ? result.split(',')[1] : result
      if (payload) {
        resolve(payload)
      } else {
        reject(new Error('empty-file'))
      }
    }
    reader.onerror = () => reject(reader.error ?? new Error('read-error'))
    reader.readAsDataURL(file)
  })
}
</script>

<template>
  <UDashboardPanel id="groups">
    <template #header>
      <UDashboardNavbar title="Groups" :ui="{ right: 'gap-3' }">
        <template #leading>
          <UDashboardSidebarCollapse />
        </template>

        <template #right>
          <UButton
            color="neutral"
            variant="ghost"
            icon="i-lucide-refresh-cw"
            :loading="refreshingGroups || detailLoading"
            @click="refreshGroups"
          >
            Refresh
          </UButton>
          <UButton icon="i-lucide-plus" @click="createGroupModalOpen = true">
            New group
          </UButton>
        </template>
      </UDashboardNavbar>
    </template>

    <template #body>
      <div class="p-6 space-y-6">
        <UAlert
          v-if="combinedGroupsError"
          color="error"
          variant="subtle"
          title="Unable to load groups"
          :description="extractErrorMessage(combinedGroupsError) || 'Please refresh the page.'"
        />

        <div class="grid gap-6 lg:grid-cols-[minmax(0,320px)_minmax(0,1fr)]">
          <UCard class="bg-elevated/40">
            <template #header>
              <div class="space-y-4">
                <UInput
                  v-model="searchQuery"
                  icon="i-lucide-search"
                  placeholder="Search groups..."
                />
                <div class="flex flex-wrap gap-2">
                  <UButton
                    v-for="tab in tabOptions"
                    :key="tab.value"
                    :label="tab.label"
                    size="sm"
                    :color="activeTab === tab.value ? 'primary' : 'neutral'"
                    :variant="activeTab === tab.value ? 'solid' : 'ghost'"
                    class="flex-1"
                    @click="activeTab = tab.value"
                  />
                </div>
              </div>
            </template>

            <template #default>
              <div v-if="listLoading" class="py-8 text-center text-muted">
                Loading groups...
              </div>
              <div v-else-if="!displayedGroups.length">
                <p class="py-8 text-center text-muted">
                  {{ anyGroupsLoaded ? 'No groups match your filters yet.' : 'No groups available. Create one to get started!' }}
                </p>
              </div>
              <div v-else class="space-y-3">
                <button
                  v-for="group in displayedGroups"
                  :key="group.id"
                  type="button"
                  class="w-full rounded-2xl border p-4 text-left transition"
                  :class="selectedGroupId === group.id ? 'border-primary bg-primary/10 shadow-sm' : 'border-default/60 hover:bg-elevated/60'"
                  @click="selectGroup(group.id)"
                >
                  <div class="flex items-start justify-between gap-2">
                    <div>
                      <p class="font-medium">
                        {{ group.title }}
                      </p>
                      <p class="text-xs text-muted">
                        Created {{ group.createdAtFormatted }}
                      </p>
                    </div>
                    <div class="flex items-center gap-1">
                      <UBadge
                        v-if="membershipIds.has(group.id)"
                        size="xs"
                        color="primary"
                      >
                        Member
                      </UBadge>
                      <UBadge
                        v-else-if="group.ownerId != null && group.ownerId === currentUserId"
                        size="xs"
                        color="warning"
                      >
                        Owner
                      </UBadge>
                    </div>
                  </div>
                  <p class="mt-3 line-clamp-2 text-sm text-muted">
                    {{ group.description || 'No description provided.' }}
                  </p>
                </button>
              </div>
            </template>
          </UCard>

          <div class="space-y-4">
            <UCard v-if="selectedSummary" class="bg-elevated/40">
              <template #header>
                <div class="flex flex-wrap items-start justify-between gap-3">
                  <div>
                    <h2 class="text-xl font-semibold">
                      {{ selectedSummary.title }}
                    </h2>
                    <p class="text-xs text-muted">
                      Created {{ selectedSummary.createdAtFormatted }}
                    </p>
                  </div>
                  <div class="flex flex-wrap items-center gap-2">
                    <UBadge v-if="joinStatus === 'owner'" color="warning">
                      Owner
                    </UBadge>
                    <UBadge v-else-if="joinStatus === 'member'" color="primary">
                      Member
                    </UBadge>
                    <UBadge v-else-if="joinStatus === 'pending'" color="neutral" variant="subtle">
                      Request pending
                    </UBadge>
                    <UButton
                      v-if="joinStatus === 'not-requested'"
                      color="primary"
                      :loading="joinLoading"
                      @click="submitJoinRequest"
                    >
                      Request to join
                    </UButton>
                  </div>
                </div>
              </template>

              <template #default>
                <div class="space-y-6">
                  <div>
                    <h3 class="text-sm font-semibold uppercase tracking-wide text-muted">
                      About
                    </h3>
                    <p class="mt-2 text-sm leading-relaxed">
                      {{ selectedSummary.description || 'No description provided yet.' }}
                    </p>
                  </div>

                  <UAlert
                    v-if="detailError"
                    color="error"
                    variant="subtle"
                    title="Unable to load group data"
                    :description="detailError"
                  />

                  <div v-if="detailLoading" class="text-center text-sm text-muted">
                    Loading group details...
                  </div>

                  <template v-else>
                    <div v-if="!showMemberContent" class="rounded-2xl border border-dashed border-default/70 p-6 text-center text-sm text-muted">
                      Join this group to unlock posts, members, and events.
                    </div>

                    <div v-else class="space-y-6">
                      <section>
                        <div class="flex items-center justify-between">
                          <h3 class="text-lg font-semibold">
                            Events
                          </h3>
                          <UTooltip v-if="!isOwner" text="Only group owners can schedule events.">
                            <span class="text-xs text-muted">Owner only</span>
                          </UTooltip>
                        </div>
                        <form v-if="isOwner" class="mt-4 space-y-3" @submit.prevent="submitEvent">
                          <div class="grid gap-3 lg:grid-cols-2">
                            <UFormGroup label="Title">
                              <UInput v-model="newEventForm.title" placeholder="Weekly sync" />
                            </UFormGroup>
                            <UFormGroup label="Day & time">
                              <UInput v-model="newEventForm.time" type="datetime-local" />
                            </UFormGroup>
                          </div>
                          <UFormGroup label="Description">
                            <UTextarea v-model="newEventForm.description" placeholder="Add context or an agenda" />
                          </UFormGroup>
                          <div class="flex justify-end">
                            <UButton type="submit" :loading="createEventLoading">
                              Schedule event
                            </UButton>
                          </div>
                        </form>
                        <div v-if="!groupDetail?.events.length" class="mt-4 rounded-xl border border-default/60 p-4 text-sm text-muted">
                          No events planned yet.
                        </div>
                        <div v-else class="mt-4 space-y-3">
                          <UCard v-for="event in groupDetail?.events" :key="event.id">
                            <h4 class="font-medium">
                              {{ event.title }}
                            </h4>
                            <p class="text-sm text-muted">
                              {{ event.formattedTime }}
                            </p>
                            <p v-if="event.description" class="mt-2 text-sm">
                              {{ event.description }}
                            </p>
                          </UCard>
                        </div>
                      </section>

                      <section>
                        <h3 class="text-lg font-semibold">
                          Members
                        </h3>
                        <div v-if="!groupDetail?.members.length" class="mt-4 rounded-xl border border-default/60 p-4 text-sm text-muted">
                          No members found.
                        </div>
                        <div v-else class="mt-4 grid gap-3 md:grid-cols-2">
                          <div
                            v-for="member in groupDetail?.members"
                            :key="member.id"
                            class="flex items-center gap-3 rounded-xl border border-default/60 p-3"
                          >
                            <UAvatar :src="member.avatarSrc" :text="member.initials" />
                            <div>
                              <p class="font-medium">
                                {{ member.fullName }}
                              </p>
                              <p class="text-xs text-muted capitalize">
                                {{ member.role }}
                              </p>
                            </div>
                          </div>
                        </div>
                      </section>

                      <section>
                        <div class="flex items-center justify-between">
                          <h3 class="text-lg font-semibold">
                            Posts
                          </h3>
                        </div>
                        <form class="mt-4 space-y-3" @submit.prevent="submitGroupPost">
                          <UFormGroup label="Share an update">
                            <UTextarea v-model="newPostForm.content" placeholder="What is new?" />
                          </UFormGroup>
                          <div class="flex flex-wrap items-center gap-3">
                            <UButton
                              type="button"
                              color="neutral"
                              variant="subtle"
                              icon="i-lucide-paperclip"
                              @click.prevent="openPostFilePicker"
                            >
                              Attach image
                            </UButton>
                            <input
                              ref="postFileInput"
                              type="file"
                              class="hidden"
                              accept="image/*"
                              @change="handlePostFileChange"
                            >
                            <span v-if="newPostForm.file" class="text-xs text-muted">
                              Attachment ready
                            </span>
                          </div>
                          <div class="flex justify-end">
                            <UButton type="submit" :loading="createPostLoading">
                              Post to group
                            </UButton>
                          </div>
                        </form>
                        <div v-if="!groupDetail?.posts.length" class="mt-4 rounded-xl border border-default/60 p-4 text-sm text-muted">
                          No posts yet. Start the conversation!
                        </div>
                        <div v-else class="mt-4 space-y-4">
                          <UCard v-for="post in groupDetail?.posts" :key="post.id" class="bg-elevated/30">
                            <template #header>
                              <div class="flex items-center gap-3">
                                <UAvatar :src="post.avatarSrc" :text="post.authorInitials" />
                                <div>
                                  <p class="font-medium">
                                    {{ post.authorName }}
                                  </p>
                                  <p class="text-xs text-muted">
                                    {{ post.formattedCreatedAt }}
                                  </p>
                                </div>
                              </div>
                            </template>
                            <div class="space-y-4">
                              <p class="text-sm whitespace-pre-line">
                                {{ post.content || 'No content provided.' }}
                              </p>
                              <div v-if="post.mediaSrc" class="overflow-hidden rounded-xl border border-default/60">
                                <img
                                  :src="post.mediaSrc"
                                  alt="Group post attachment"
                                  class="w-full"
                                  loading="lazy"
                                >
                              </div>
                            </div>
                            <template #footer>
                              <div class="flex flex-wrap items-center gap-3">
                                <div class="flex items-center gap-1 text-sm text-muted">
                                  <UIcon name="i-lucide-message-square" class="size-4" />
                                  <span>{{ post.commentCount }} comments</span>
                                </div>
                                <UButton size="xs" variant="ghost" @click="togglePostComments(post.id)">
                                  {{ isPostExpanded(post.id) ? 'Hide comments' : 'View comments' }}
                                </UButton>
                              </div>
                              <div v-if="isPostExpanded(post.id)" class="mt-4 space-y-3 rounded-2xl border border-default/60 p-4">
                                <div v-if="commentsLoading[post.id]" class="text-sm text-muted">
                                  Loading comments...
                                </div>
                                <div v-else-if="!commentsCache[post.id]?.length" class="text-sm text-muted">
                                  No comments yet.
                                </div>
                                <div v-else class="space-y-3">
                                  <div v-for="comment in commentsCache[post.id]" :key="comment.id" class="rounded-xl border border-default/40 p-3">
                                    <div class="flex items-center gap-2 text-xs text-muted">
                                      <span class="font-medium text-default">{{ comment.authorName }}</span>
                                      <span>•</span>
                                      <span>{{ comment.formattedCreatedAt }}</span>
                                    </div>
                                    <p class="mt-2 text-sm">
                                      {{ comment.content }}
                                    </p>
                                  </div>
                                </div>
                                <div class="flex gap-2">
                                  <UTextarea
                                    v-model="newCommentDrafts[post.id]"
                                    placeholder="Write a comment"
                                    class="flex-1"
                                  />
                                  <UButton
                                    color="primary"
                                    :loading="commentSubmitting[post.id]"
                                    @click="submitComment(post.id)"
                                  >
                                    Send
                                  </UButton>
                                </div>
                              </div>
                            </template>
                          </UCard>
                        </div>
                      </section>
                    </div>
                  </template>
                </div>
              </template>
            </UCard>

            <div v-else class="rounded-2xl border border-dashed border-default/70 p-8 text-center text-muted">
              Select a group to view its details.
            </div>
          </div>
        </div>
      </div>
    </template>
  </UDashboardPanel>

  <UModal v-model="createGroupModalOpen">
    <UCard>
      <template #header>
        <h3 class="text-lg font-semibold">
          Create a group
        </h3>
      </template>
      <form class="space-y-4" @submit.prevent="submitCreateGroup">
        <UFormGroup label="Title" :error="createGroupErrors.title">
          <UInput v-model="createGroupForm.title" placeholder="AI Researchers" />
        </UFormGroup>
        <UFormGroup label="Description" :error="createGroupErrors.description">
          <UTextarea v-model="createGroupForm.description" placeholder="Describe the purpose of your group" />
        </UFormGroup>
        <div class="flex justify-end gap-2">
          <UButton
            type="button"
            color="neutral"
            variant="ghost"
            @click="createGroupModalOpen = false"
          >
            Cancel
          </UButton>
          <UButton type="submit" :loading="createGroupLoading">
            Create group
          </UButton>
        </div>
      </form>
    </UCard>
  </UModal>
</template>
