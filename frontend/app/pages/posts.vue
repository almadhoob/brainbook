<script setup lang="ts">
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

interface PostFeedItem {
  id: number | string
  authorName: string
  authorInitials: string
  avatarSrc?: string
  postedAt: string
  content: string
  mediaSrc?: string
  commentCount: number
}

const runtimeConfig = useRuntimeConfig()
const apiBase = typeof runtimeConfig.public?.apiBase === 'string' && runtimeConfig.public.apiBase.length > 0
  ? runtimeConfig.public.apiBase
  : 'http://localhost:8080'

const { data, status, error, refresh } = await useFetch<{ posts: ApiPost[] }>(`${apiBase}/protected/v1/posts`, {
  credentials: 'include',
  lazy: true
})

const posts = computed<PostFeedItem[]>(() => {
  const rawPosts = data.value?.posts
  if (!Array.isArray(rawPosts)) return []

  return rawPosts.map((post, index) => {
    const firstName = post.f_name?.trim() || ''
    const lastName = post.l_name?.trim() || ''
    const authorName = `${firstName} ${lastName}`.trim() || 'Unknown user'
    const initials = [firstName, lastName]
      .filter(Boolean)
      .map(chunk => chunk[0]?.toUpperCase())
      .join('') || 'U'

    return {
      id: typeof post.id === 'number' ? post.id : `post-${index}`,
      authorName,
      authorInitials: initials,
      avatarSrc: binaryToDataUrl(post.avatar, 'image/png'),
      postedAt: formatTimestamp(post.created_at),
      content: (post.content ?? '').trim(),
      mediaSrc: binaryToDataUrl(post.file, 'image/jpeg'),
      commentCount: typeof post.comment_count === 'number' ? post.comment_count : 0
    }
  })
})

const isInitialLoading = computed(() => status.value === 'pending' && !data.value)
const isEmpty = computed(() => status.value === 'success' && !error.value && posts.value.length === 0)
const isRefreshing = computed(() => status.value === 'pending')

const errorMessage = computed(() => {
  if (!error.value) return ''
  if (typeof error.value === 'string') return error.value
  if (error.value instanceof Error) return error.value.message
  if (typeof error.value === 'object' && 'statusMessage' in (error.value as Record<string, unknown>)) {
    const statusMessage = (error.value as Record<string, unknown>).statusMessage
    if (typeof statusMessage === 'string') return statusMessage
  }
  return 'Something went wrong while loading the posts feed.'
})

function binaryToDataUrl(raw?: string | null, mime = 'image/png') {
  if (typeof raw !== 'string') return undefined
  const trimmed = raw.trim()
  if (!trimmed) return undefined
  if (trimmed.startsWith('data:')) return trimmed
  return `data:${mime};base64,${trimmed}`
}

function formatTimestamp(timestamp?: string | null) {
  if (!timestamp) return 'Unknown date'
  const normalized = timestamp.includes('T') ? timestamp : timestamp.replace(' ', 'T')
  const parsed = new Date(normalized)
  if (Number.isNaN(parsed.getTime())) return timestamp
  return parsed.toLocaleString()
}

function handlePostCreated() {
  refresh()
}

function handleRefreshClick(_: MouseEvent) {
  return refresh()
}
</script>

<template>
  <UDashboardPanel id="posts">
    <template #header>
      <UDashboardNavbar title="Posts" :ui="{ right: 'gap-3' }">
        <template #leading>
          <UDashboardSidebarCollapse />
        </template>

        <template #right>
          <UButton
            color="neutral"
            variant="ghost"
            icon="i-lucide-refresh-cw"
            :loading="isRefreshing"
            @click="handleRefreshClick"
          >
            Refresh
          </UButton>
          <PostsCreateModal :api-base="apiBase" @created="handlePostCreated" />
        </template>
      </UDashboardNavbar>
    </template>

    <template #body>
      <div class="p-6 space-y-4">
        <UAlert
          v-if="error"
          variant="subtle"
          color="error"
          title="Unable to load posts"
          :description="errorMessage"
        />

        <div v-if="isInitialLoading" class="py-8 text-center text-muted">
          Loading posts...
        </div>

        <div v-else-if="isEmpty" class="py-8 text-center text-muted">
          No posts to display yet. Create one to get the conversation started!
        </div>

        <div v-else class="space-y-4">
          <UCard v-for="post in posts" :key="post.id" class="bg-elevated/40">
            <template #header>
              <div class="flex items-center gap-3">
                <UAvatar size="lg" :src="post.avatarSrc" :text="post.authorInitials" />
                <div class="flex flex-col">
                  <span class="font-medium">{{ post.authorName }}</span>
                  <span class="text-xs text-muted">{{ post.postedAt }}</span>
                </div>
              </div>
            </template>

            <div class="space-y-4">
              <p class="text-sm text-default whitespace-pre-line">
                {{ post.content || 'No content provided.' }}
              </p>
              <div v-if="post.mediaSrc" class="overflow-hidden rounded-xl border border-default/50">
                <img
                  :src="post.mediaSrc"
                  alt="Post attachment"
                  class="w-full max-h-96 object-cover"
                  loading="lazy"
                  decoding="async"
                >
              </div>
            </div>

            <template #footer>
              <div class="flex items-center gap-2 text-sm text-muted">
                <UIcon name="i-lucide-message-square" class="size-4" />
                <span>{{ post.commentCount }} comments</span>
              </div>
            </template>
          </UCard>
        </div>
      </div>
    </template>
  </UDashboardPanel>
</template>
