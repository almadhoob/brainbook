<script setup lang="ts">
import type { ReceiveGroupMessageEventPayload } from '~/types'
import type { ApiGroup, GroupSummary } from '~/composables/useGroupDetail'
import { extractErrorMessage } from '~/composables/useGroupHelpers'

const toast = useToast()
const { hydrate: hydrateSession } = useSession()
const runtimeConfig = useRuntimeConfig()
const apiBase = typeof runtimeConfig.public?.apiBase === 'string' && runtimeConfig.public.apiBase.length > 0
  ? runtimeConfig.public.apiBase
  : 'http://localhost:8080'

type TabKey = 'all' | 'mine'
type ContentTabKey = 'posts' | 'chat' | 'events' | 'members'

const tabOptions: { label: string, value: TabKey }[] = [
  { label: 'All groups', value: 'all' },
  { label: 'My groups', value: 'mine' }
]

const contentTabOptions: { label: string, value: ContentTabKey, icon: string }[] = [
  { label: 'Posts', value: 'posts', icon: 'i-lucide-file-text' },
  { label: 'Chat', value: 'chat', icon: 'i-lucide-message-square' },
  { label: 'Events', value: 'events', icon: 'i-lucide-calendar' },
  { label: 'Members', value: 'members', icon: 'i-lucide-users' }
]

const { connect: connectRealtime, groupMessageBus } = useRealtime()

const activeTab = ref<TabKey>('all')
const activeContentTab = ref<ContentTabKey>('posts')
const searchQuery = ref('')
const refreshingGroups = ref(false)

// Initialize composables
const groupDetail = useGroupDetail(apiBase)
const groupPosts = useGroupPosts(apiBase, groupDetail.selectedGroupId)
const groupChat = useGroupChat(apiBase, groupDetail.selectedGroupId, groupDetail.currentUserId)
const groupEvents = useGroupEvents(apiBase, groupDetail.selectedGroupId, groupDetail.isOwner)
const groupMembers = useGroupMembers(apiBase, groupDetail.selectedGroupId, groupDetail.isOwner)

if (import.meta.client) {
  hydrateSession()
  connectRealtime()

  const stopGroupMessages = groupMessageBus.on((event: ReceiveGroupMessageEventPayload) => {
    groupChat.handleIncomingMessage(event, groupMembers.members.value)
  })
  onScopeDispose(stopGroupMessages)
}

const {
  data: allGroupsData,
  status: allGroupsStatus,
  error: allGroupsError,
  refresh: refreshAllGroups
} = await useFetch<{ groups: ApiGroup[] }>(`${apiBase}/protected/v1/groups`, {
  credentials: 'include',
  lazy: true,
  server: false
})

const {
  data: myGroupsData,
  status: myGroupsStatus,
  error: myGroupsError,
  refresh: refreshMyGroups
} = await useFetch<{ groups: ApiGroup[] }>(`${apiBase}/protected/v1/user/groups`, {
  credentials: 'include',
  lazy: true,
  server: false
})

const combinedGroupsError = computed(() => allGroupsError.value ?? myGroupsError.value ?? null)

function normalizeGroupList(groups?: ApiGroup[]): GroupSummary[] {
  if (!Array.isArray(groups)) return []

  return groups
    .map(group => groupDetail.normalizeGroup(group))
    .filter((group): group is GroupSummary => Boolean(group))
    .sort((a, b) => a.title.localeCompare(b.title))
}

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
  if (groupDetail.selectedGroupId.value == null) return null
  if (groupDetail.groupDetail.value && groupDetail.groupDetail.value.id === groupDetail.selectedGroupId.value) {
    return groupDetail.groupDetail.value
  }
  return fallbackGroupMap.value.get(groupDetail.selectedGroupId.value) ?? null
})

const listLoading = computed(() =>
  (activeTab.value === 'all' ? allGroupsStatus.value === 'pending' : myGroupsStatus.value === 'pending')
  || refreshingGroups.value
)

const displayedGroups = computed(() => {
  const base = activeTab.value === 'all' ? normalizedAllGroups.value : normalizedMyGroups.value
  if (!searchQuery.value.trim()) return base

  const query = searchQuery.value.trim().toLowerCase()
  return base.filter((group) => {
    const haystack = `${group.title} ${group.description}`.toLowerCase()
    return haystack.includes(query)
  })
})

const isMember = computed(() =>
  groupDetail.selectedGroupId.value != null
  && membershipIds.value.has(groupDetail.selectedGroupId.value)
)

const showMemberContent = computed(() => isMember.value || groupDetail.isOwner.value)
const anyGroupsLoaded = computed(() =>
  normalizedAllGroups.value.length > 0 || normalizedMyGroups.value.length > 0
)

const joinStatusDisplay = computed(() => {
  if (groupDetail.isOwner.value) return { type: 'owner', label: 'Owner', color: 'warning' }
  if (isMember.value) return { type: 'member', label: 'Member', color: 'primary' }
  if (groupDetail.joinStatus.value === 'pending') return { type: 'pending', label: 'Request pending', color: 'neutral' }
  return { type: 'none', label: '', color: 'neutral' }
})

// Watch for group changes and auto-select
watch(
  () => [normalizedMyGroups.value, normalizedAllGroups.value],
  ([mine = [], all = []]) => {
    if (groupDetail.selectedGroupId.value != null) {
      const stillExists = mine.some(group => group.id === groupDetail.selectedGroupId.value)
        || all.some(group => group.id === groupDetail.selectedGroupId.value)
      if (stillExists) return
    }

    const fallback = mine[0]?.id ?? all[0]?.id ?? null
    if (fallback != null) {
      groupDetail.selectGroup(fallback)
    } else {
      groupDetail.selectedGroupId.value = null
      groupDetail.groupDetail.value = null
    }
  },
  { immediate: true }
)

// Watch for selected group changes to load content
watch(() => groupDetail.selectedGroupId.value, async (groupId) => {
  groupPosts.clearPostsState()
  groupEvents.clearEventForm()

  if (groupId == null) {
    groupDetail.groupDetail.value = null
    return
  }

  try {
    const summary = await groupDetail.loadGroupSummary(groupId)

    // Check membership based on the freshly loaded data
    const isGroupMember = membershipIds.value.has(groupId)
    const isGroupOwner = summary.ownerId != null && groupDetail.currentUserId.value === summary.ownerId

    if (isGroupMember || isGroupOwner) {
      await Promise.all([
        groupPosts.loadPosts(),
        groupChat.loadMessages(),
        groupEvents.loadEvents(),
        groupMembers.loadMembers()
      ])
    }
  } catch {
    // Error already handled in loadGroupSummary
  }
})

async function refreshGroups() {
  refreshingGroups.value = true
  try {
    await Promise.all([refreshAllGroups(), refreshMyGroups()])
    if (groupDetail.selectedGroupId.value != null) {
      const summary = await groupDetail.loadGroupSummary(groupDetail.selectedGroupId.value)

      // Check membership based on the freshly loaded data
      const isGroupMember = membershipIds.value.has(groupDetail.selectedGroupId.value)
      const isGroupOwner = summary.ownerId != null && groupDetail.currentUserId.value === summary.ownerId

      if (isGroupMember || isGroupOwner) {
        await Promise.all([
          groupPosts.loadPosts(),
          groupChat.loadMessages(),
          groupEvents.loadEvents(),
          groupMembers.loadMembers()
        ])
      }
    }
  } catch (err) {
    toast.add({
      title: 'Refresh failed',
      description: extractErrorMessage(err) || 'Unable to refresh groups right now.',
      color: 'error'
    })
  } finally {
    refreshingGroups.value = false
  }
}

async function handleGroupCreated() {
  await refreshGroups()
}

async function handleJoinRequest() {
  try {
    const status = await groupDetail.submitJoinRequest()
    if (status === 'member' || status === 'owner') {
      await refreshGroups()
    }
  } catch {
    // Error already handled in composable
  }
}

const currentMessageDraft = computed({
  get: () => groupChat.groupMessageDrafts[groupDetail.selectedGroupId.value || -1] || '',
  set: (value: string) => {
    if (groupDetail.selectedGroupId.value) {
      groupChat.groupMessageDrafts[groupDetail.selectedGroupId.value] = value
    }
  }
})
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
            :loading="refreshingGroups || groupDetail.detailLoading.value"
            @click="refreshGroups"
          >
            Refresh
          </UButton>
          <GroupsCreateModal :api-base="apiBase" @created="handleGroupCreated" />
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

        <div class="grid gap-6 lg:grid-cols-[minmax(0,320px)_minmax(0,1fr)] lg:h-[calc(100vh-12rem)]">
          <!-- Group List Sidebar -->
          <UCard class="bg-elevated/40 flex flex-col h-full">
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
              <div class="overflow-y-auto flex-1">
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
                    :class="groupDetail.selectedGroupId.value === group.id ? 'border-primary bg-primary/10 shadow-sm' : 'border-default/60 hover:bg-elevated/60'"
                    @click="groupDetail.selectGroup(group.id)"
                  >
                    <div class="flex items-start justify-between gap-2">
                      <div class="min-w-0">
                        <p class="font-medium truncate">
                          {{ group.title }}
                        </p>
                        <p class="text-xs text-muted">
                          Created {{ group.createdAtFormatted }}
                        </p>
                      </div>
                      <div class="flex items-center gap-1">
                        <UBadge
                          v-if="group.ownerId != null && group.ownerId === groupDetail.currentUserId.value"
                          size="xs"
                          color="warning"
                        >
                          Owner
                        </UBadge>
                        <UBadge
                          v-else-if="membershipIds.has(group.id)"
                          size="xs"
                          color="primary"
                        >
                          Member
                        </UBadge>
                      </div>
                    </div>
                    <p class="mt-3 text-sm truncate">
                      {{ group.description || 'No description provided.' }}
                    </p>
                  </button>
                </div>
              </div>
            </template>
          </UCard>

          <!-- Group Detail Content -->
          <div class="space-y-4 h-full overflow-y-auto">
            <UCard v-if="selectedSummary" class="bg-elevated/40">
              <template #header>
                <div class="flex flex-wrap items-center justify-end gap-2">
                  <UBadge v-if="joinStatusDisplay.type === 'owner'" color="warning">
                    {{ joinStatusDisplay.label }}
                  </UBadge>
                  <UBadge v-else-if="joinStatusDisplay.type === 'member'" color="primary">
                    {{ joinStatusDisplay.label }}
                  </UBadge>
                  <UBadge v-else-if="joinStatusDisplay.type === 'pending'" color="neutral" variant="subtle">
                    {{ joinStatusDisplay.label }}
                  </UBadge>
                  <UButton
                    v-if="joinStatusDisplay.type === 'none'"
                    color="primary"
                    :loading="groupDetail.joinLoading.value"
                    @click="handleJoinRequest"
                  >
                    Request to join
                  </UButton>
                </div>
              </template>

              <template #default>
                <div class="space-y-6">
                  <UAlert
                    v-if="groupDetail.detailError.value"
                    color="error"
                    variant="subtle"
                    title="Unable to load group data"
                    :description="groupDetail.detailError.value"
                  />

                  <div v-if="groupDetail.detailLoading.value" class="text-center text-sm text-muted">
                    Loading group details...
                  </div>

                  <template v-else>
                    <div v-if="!showMemberContent" class="rounded-2xl border border-dashed border-default/70 p-6 text-center text-sm text-muted">
                      Join this group to unlock posts, members, and events.
                    </div>

                    <div v-else class="space-y-6">
                      <!-- Tab Navigation -->
                      <div class="flex flex-wrap gap-2 border-b border-default/60 pb-2">
                        <UButton
                          v-for="tab in contentTabOptions"
                          :key="tab.value"
                          :label="tab.label"
                          :icon="tab.icon"
                          size="sm"
                          :color="activeContentTab === tab.value ? 'primary' : 'neutral'"
                          :variant="activeContentTab === tab.value ? 'solid' : 'ghost'"
                          @click="activeContentTab = tab.value"
                        />
                      </div>

                      <!-- Tab Content -->
                      <GroupsPostsTab
                        v-if="activeContentTab === 'posts'"
                        v-model:new-post-form="groupPosts.newPostForm"
                        v-model:new-comment-drafts="groupPosts.newCommentDrafts"
                        v-model:new-comment-files="groupPosts.newCommentFiles"
                        :posts="groupPosts.posts.value"
                        :posts-loading="groupPosts.postsLoading.value"
                        :create-post-loading="groupPosts.createPostLoading.value"
                        :post-count="groupPosts.postCount.value"
                        :comments-cache="groupPosts.commentsCache"
                        :comments-loading="groupPosts.commentsLoading"
                        :comment-submitting="groupPosts.commentSubmitting"
                        :expanded-posts="groupPosts.expandedPosts.value"
                        @submit-post="groupPosts.submitPost"
                        @toggle-comments="groupPosts.togglePostComments"
                        @submit-comment="groupPosts.submitComment"
                      />

                      <GroupsChatTab
                        v-if="activeContentTab === 'chat'"
                        v-model:message-draft="currentMessageDraft"
                        :messages="groupChat.activeGroupMessages.value"
                        :messages-loading="groupChat.groupMessagesLoading[groupDetail.selectedGroupId.value || -1] || false"
                        :message-sending="groupChat.groupMessageSending[groupDetail.selectedGroupId.value || -1] || false"
                        :current-user-id="groupDetail.currentUserId.value"
                        :is-online="groupChat.isUserOnline(groupDetail.currentUserId.value)"
                        @send-message="groupChat.sendMessage"
                      />

                      <GroupsEventsTab
                        v-if="activeContentTab === 'events'"
                        v-model:new-event-form="groupEvents.newEventForm"
                        :events="groupEvents.events.value"
                        :events-loading="groupEvents.eventsLoading.value"
                        :create-event-loading="groupEvents.createEventLoading.value"
                        :rsvp-loading="groupEvents.rsvpLoading"
                        :is-owner="groupDetail.isOwner.value"
                        @create-event="groupEvents.createEvent"
                        @rsvp="groupEvents.respondRsvp"
                      />

                      <GroupsMembersTab
                        v-if="activeContentTab === 'members'"
                        v-model:owner-request-response="groupMembers.ownerRequestResponse"
                        :members="groupMembers.members.value"
                        :members-loading="groupMembers.membersLoading.value"
                        :is-owner="groupDetail.isOwner.value"
                        @respond-request="groupMembers.respondGroupRequestById"
                        @refresh-members="groupMembers.loadMembers"
                      />
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
</template>
