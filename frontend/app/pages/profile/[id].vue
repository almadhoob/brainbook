<script setup lang="ts">
import { useFollowers } from '~/composables/useFollowers'
import { normalizeAvatar } from '~/utils'

interface ApiUserSummary {
  user_id?: number
  f_name?: string | null
  l_name?: string | null
  avatar?: string | null
}

interface ApiPost {
  id?: number
  content?: string | null
  file?: string | null
  created_at?: string | null
  comment_count?: number | null
  f_name?: string | null
  l_name?: string | null
  avatar?: string | null
}

interface ProfileResponse {
  user_id?: number
  full_name?: string
  email?: string
  dob?: string
  is_public?: boolean
  followers?: ApiUserSummary[]
  following?: ApiUserSummary[]
  posts?: ApiPost[]
  pending_follow_requests_count?: number
  is_self?: boolean
  avatar?: string | null
  nickname?: string
  bio?: string
  follow_request_status?: string | null
}

interface ProfilePostItem {
  id: number | string
  authorName: string
  authorInitials: string
  avatarSrc?: string
  content: string
  postedAt: string
  mediaSrc?: string
  commentCount: number
}

const toast = useToast()
const route = useRoute()
const { session, hydrate } = useSession()
const requestPending = ref(false)

if (import.meta.client) {
  hydrate()
}

const runtimeConfig = useRuntimeConfig()
const apiBase = typeof runtimeConfig.public?.apiBase === 'string' && runtimeConfig.public.apiBase.length > 0
  ? runtimeConfig.public.apiBase
  : 'http://localhost:8080'

const profileId = computed(() => Number(route.params.id))
const normalizedProfileId = computed(() => Number.isFinite(profileId.value) ? profileId.value : null)

const {
  followers,
  loading: followersLoading,
  error: followersError,
  loadFollowers
} = useFollowers(apiBase, normalizedProfileId)

const { data, status, error, refresh } = await useFetch<ProfileResponse>(
  () => `${apiBase}/guest/v1/profile/user/${profileId.value}`,
  {
    credentials: 'include',
    lazy: true,
    server: false,
    watch: [profileId]
  }
)

const profile = computed(() => data.value)
const following = computed(() => profile.value?.following ?? [])

const avatarSrc = computed(() => normalizeAvatar(profile.value?.avatar))
const initials = computed(() => {
  const name = profile.value?.full_name?.trim() || ''
  const parts = name.split(/\s+/).filter(Boolean)
  if (!parts.length) return 'U'
  return parts.map(part => part[0]?.toUpperCase()).join('').slice(0, 2)
})

const isLoading = computed(() => !error.value && !profile.value)
const isSelf = computed(() => Boolean(profile.value?.is_self))
const isFollowing = computed(() => {
  const currentId = session.value.user_id
  if (!currentId) return false
  return followers.value.some(follower => follower.user_id === currentId)
})
const isLimitedProfile = computed(() => {
  if (!profile.value) return false
  if (isSelf.value || profile.value.is_public) return false
  const hasEmail = typeof profile.value.email === 'string' && profile.value.email.length > 0
  const hasDob = typeof profile.value.dob === 'string' && profile.value.dob.length > 0
  return !(hasEmail || hasDob)
})

const isRequestPending = computed(() => {
  if (requestPending.value) return true
  return profile.value?.follow_request_status === 'pending'
})

watch([normalizedProfileId, profile], async ([id, profileValue]) => {
  if (!id || !profileValue) return
  await loadFollowers(true)
})

const posts = computed<ProfilePostItem[]>(() => {
  const rawPosts = profile.value?.posts
  if (!Array.isArray(rawPosts)) return []
  return rawPosts.map((post, index) => {
    const firstName = post.f_name?.trim() || ''
    const lastName = post.l_name?.trim() || ''
    const authorName = `${firstName} ${lastName}`.trim() || profile.value?.full_name || 'Unknown user'
    const authorInitials = [firstName, lastName]
      .filter(Boolean)
      .map(chunk => chunk[0]?.toUpperCase())
      .join('') || initials.value

    return {
      id: typeof post.id === 'number' ? post.id : `post-${index}`,
      authorName,
      authorInitials,
      avatarSrc: binaryToDataUrl(post.avatar, 'image/png') ?? avatarSrc.value,
      content: (post.content ?? '').trim(),
      postedAt: formatTimestamp(post.created_at),
      mediaSrc: binaryToDataUrl(post.file, 'image/jpeg'),
      commentCount: typeof post.comment_count === 'number' ? post.comment_count : 0
    }
  })
})

const errorMessage = computed(() => {
  if (!error.value) return ''
  if (typeof error.value === 'string') return error.value
  if (error.value instanceof Error) return error.value.message
  if (typeof error.value === 'object' && error.value !== null) {
    const statusCode = 'status' in error.value ? (error.value as { status?: number }).status : undefined
    if (statusCode === 401 || statusCode === 403) {
      return 'This profile is private. Send a follow request to see more.'
    }
    if ('statusMessage' in error.value && typeof (error.value as { statusMessage?: string }).statusMessage === 'string') {
      return (error.value as { statusMessage?: string }).statusMessage as string
    }
  }
  return 'Unable to load profile information.'
})

watchEffect(() => {
  if (isFollowing.value) {
    requestPending.value = false
  }
})

function formatTimestamp(timestamp?: string | null) {
  if (!timestamp) return 'Unknown date'
  const normalized = timestamp.includes('T') ? timestamp : timestamp.replace(' ', 'T')
  const parsed = new Date(normalized)
  if (Number.isNaN(parsed.getTime())) return timestamp
  return parsed.toLocaleString()
}

function formatDate(dateString?: string | null) {
  if (!dateString) return '—'
  const parsed = new Date(dateString)
  if (Number.isNaN(parsed.getTime())) return dateString
  return parsed.toLocaleDateString()
}

function binaryToDataUrl(raw?: string | null, mime = 'image/png') {
  if (typeof raw !== 'string') return undefined
  const trimmed = raw.trim()
  if (!trimmed) return undefined
  if (trimmed.startsWith('data:')) return trimmed
  return `data:${mime};base64,${trimmed}`
}

async function followUser() {
  if (!profile.value?.user_id) return
  try {
    await $fetch(`${apiBase}/protected/v1/users/${profile.value.user_id}/follow`, {
      method: 'POST',
      credentials: 'include'
    })
    requestPending.value = true
    toast.add({
      title: 'Follow request sent',
      description: profile.value.is_public
        ? 'You are now following this user.'
        : 'Pending approval if the profile is private.'
    })
    await refresh()
  } catch (err) {
    toast.add({ title: 'Error', description: 'Could not follow this user.', color: 'error' })
    console.error(err)
  }
}

async function unfollowUser() {
  if (!profile.value?.user_id) return
  try {
    await $fetch(`${apiBase}/protected/v1/users/${profile.value.user_id}/unfollow`, {
      method: 'POST',
      credentials: 'include'
    })
    toast.add({ title: 'Unfollowed', description: 'You are no longer following this user.' })
    await refresh()
  } catch (err) {
    toast.add({ title: 'Error', description: 'Could not unfollow this user.', color: 'error' })
    console.error(err)
  }
}

function handleRefresh() {
  return refresh()
}
</script>

<template>
  <UDashboardPanel id="profile">
    <template #header>
      <UDashboardNavbar title="Profile">
        <template #leading>
          <UDashboardSidebarCollapse />
        </template>
      </UDashboardNavbar>
    </template>

    <template #body>
      <div v-if="isLoading" class="py-10 text-center text-muted">
        Loading profile...
      </div>

      <div v-else-if="error" class="py-10 text-center text-error">
        {{ errorMessage }}
      </div>

      <div v-else class="space-y-6">
        <UCard class="border border-default/60">
          <div class="flex flex-wrap items-center justify-between gap-6">
            <div class="flex items-center gap-4">
              <UAvatar :src="avatarSrc" :text="initials" size="xl" />
              <div>
                <p class="text-xl font-semibold">
                  {{ profile?.full_name || 'Unknown user' }}
                </p>
                <p v-if="profile?.nickname" class="text-sm text-muted">
                  {{ profile.nickname }}
                </p>
                <p class="text-xs text-muted">
                  {{ profile?.is_public ? 'Public profile' : 'Private profile' }}
                </p>
              </div>
            </div>
            <div v-if="!isSelf" class="flex flex-wrap gap-2">
              <UButton
                v-if="isFollowing"
                color="neutral"
                variant="soft"
                @click="unfollowUser"
              >
                Unfollow
              </UButton>
              <UButton
                v-else
                color="primary"
                :disabled="isRequestPending"
                @click="followUser"
              >
                {{ isRequestPending ? 'Request pending' : 'Follow' }}
              </UButton>
            </div>
          </div>

          <div v-if="isLimitedProfile" class="mt-4 rounded-lg border border-primary/30 bg-primary/5 p-3 text-sm text-primary">
            This is a private profile. Follow to unlock full details and posts.
          </div>

          <div class="mt-6 grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
            <div class="rounded-lg border border-default/60 p-4">
              <p class="text-xs text-muted">
                Followers
              </p>
              <p class="text-lg font-semibold">
                {{ followers.length }}
              </p>
            </div>
            <div class="rounded-lg border border-default/60 p-4">
              <p class="text-xs text-muted">
                Following
              </p>
              <p class="text-lg font-semibold">
                {{ following.length }}
              </p>
            </div>
            <div class="rounded-lg border border-default/60 p-4">
              <p class="text-xs text-muted">
                Posts
              </p>
              <p class="text-lg font-semibold">
                {{ posts.length }}
              </p>
            </div>
            <div v-if="isSelf" class="rounded-lg border border-default/60 p-4">
              <p class="text-xs text-muted">
                Pending follow requests
              </p>
              <p class="text-lg font-semibold">
                {{ profile?.pending_follow_requests_count ?? 0 }}
              </p>
            </div>
          </div>
        </UCard>

        <div class="grid gap-4 lg:grid-cols-3">
          <UCard v-if="!isLimitedProfile" class="lg:col-span-1">
            <template #header>
              <p class="text-sm font-semibold">
                About
              </p>
            </template>
            <div class="space-y-3 text-sm">
              <div>
                <p class="text-xs text-muted">
                  Email
                </p>
                <p>{{ profile?.email || '—' }}</p>
              </div>
              <div>
                <p class="text-xs text-muted">
                  Date of birth
                </p>
                <p>{{ formatDate(profile?.dob) }}</p>
              </div>
              <div>
                <p class="text-xs text-muted">
                  Bio
                </p>
                <p>{{ profile?.bio || 'No bio added yet.' }}</p>
              </div>
            </div>
          </UCard>

          <UCard class="lg:col-span-2">
            <template #header>
              <p class="text-sm font-semibold">
                Followers
              </p>
            </template>
            <div v-if="followersLoading" class="text-sm text-muted">
              Loading followers...
            </div>
            <div v-else-if="followersError" class="text-sm text-red-500">
              {{ followersError }}
            </div>
            <div v-else-if="followers.length === 0" class="text-sm text-muted">
              No followers yet.
            </div>
            <div v-else class="grid gap-3 sm:grid-cols-2">
              <NuxtLink
                v-for="user in followers"
                :key="user.user_id"
                :to="`/profile/${user.user_id}`"
                class="flex items-center gap-3 rounded-lg border border-default/60 p-3 hover:border-primary/60"
              >
                <UAvatar :src="normalizeAvatar(user.avatar)" :text="user.f_name?.[0] || 'U'" />
                <div>
                  <p class="text-sm font-medium">
                    {{ `${user.f_name ?? ''} ${user.l_name ?? ''}`.trim() || 'Unknown user' }}
                  </p>
                </div>
              </NuxtLink>
            </div>
          </UCard>
        </div>

        <UCard>
          <template #header>
            <p class="text-sm font-semibold">
              Following
            </p>
          </template>
          <div v-if="following.length === 0" class="text-sm text-muted">
            Not following anyone yet.
          </div>
          <div v-else class="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
            <NuxtLink
              v-for="user in following"
              :key="user.user_id"
              :to="`/profile/${user.user_id}`"
              class="flex items-center gap-3 rounded-lg border border-default/60 p-3 hover:border-primary/60"
            >
              <UAvatar :src="normalizeAvatar(user.avatar)" :text="user.f_name?.[0] || 'U'" />
              <div>
                <p class="text-sm font-medium">
                  {{ `${user.f_name ?? ''} ${user.l_name ?? ''}`.trim() || 'Unknown user' }}
                </p>
              </div>
            </NuxtLink>
          </div>
        </UCard>

        <UCard>
          <template #header>
            <div class="flex items-center justify-between">
              <p class="text-sm font-semibold">
                Posts
              </p>
              <UButton
                color="neutral"
                variant="ghost"
                icon="i-lucide-refresh-cw"
                @click="handleRefresh"
              >
                Refresh
              </UButton>
            </div>
          </template>

          <div v-if="posts.length === 0" class="text-sm text-muted">
            No posts to show yet.
          </div>

          <div v-else class="space-y-4">
            <UCard v-for="post in posts" :key="post.id" class="border border-default/50">
              <div class="flex items-center gap-3">
                <UAvatar :src="post.avatarSrc" :text="post.authorInitials" />
                <div>
                  <p class="text-sm font-medium">
                    {{ post.authorName }}
                  </p>
                  <p class="text-xs text-muted">
                    {{ post.postedAt }}
                  </p>
                </div>
              </div>
              <div class="mt-3 whitespace-pre-line text-sm">
                {{ post.content || 'No content' }}
              </div>
              <img
                v-if="post.mediaSrc"
                :src="post.mediaSrc"
                alt="Post media"
                class="mt-3 max-h-96 w-full rounded-lg object-cover"
              >
              <div class="mt-3 text-xs text-muted">
                {{ post.commentCount }} comment(s)
              </div>
            </UCard>
          </div>
        </UCard>
      </div>
    </template>
  </UDashboardPanel>
</template>
