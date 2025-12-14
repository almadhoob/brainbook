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

interface ApiPostComment {
  id?: number
  user_id?: number
  user_full_name?: string | null
  user_avatar?: string | null
  content?: string | null
  created_at?: string | null
}

interface PostComment {
  id: number | string
  authorName: string
  authorInitials: string
  avatarSrc?: string
  content: string
  formattedCreatedAt: string
}

const toast = useToast()
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

const commentsCache = reactive<Record<number | string, PostComment[]>>({})
const commentsLoading = reactive<Record<number | string, boolean>>({})
const commentDrafts = reactive<Record<number | string, string>>({})
const commentSubmitting = reactive<Record<number | string, boolean>>({})
const expandedPosts = ref(new Set<number | string>())

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

function extractErrorMessage(reason: unknown) {
  if (!reason) return ''
  if (reason instanceof Error) return reason.message
  if (typeof reason === 'string') return reason
  if (typeof reason === 'object' && 'data' in (reason as Record<string, unknown>)) {
    const data = (reason as Record<string, any>).data
    if (data && typeof data === 'object' && 'error' in data && typeof (data as any).error === 'string') {
      return (data as any).error
    }
  }
  return ''
}

function normalizeComments(comments?: ApiPostComment[]): PostComment[] {
  if (!Array.isArray(comments)) return []

  return comments.map((comment, index) => {
    const authorName = comment.user_full_name?.trim() || 'Unknown user'
    const initials = authorName
      .split(' ')
      .filter(Boolean)
      .map(chunk => chunk[0]?.toUpperCase())
      .join('')
      .slice(0, 2) || 'U'

    return {
      id: typeof comment.id === 'number' ? comment.id : `comment-${index}`,
      authorName,
      authorInitials: initials,
      avatarSrc: binaryToDataUrl(comment.user_avatar, 'image/png'),
      content: (comment.content ?? '').trim(),
      formattedCreatedAt: formatTimestamp(comment.created_at)
    }
  })
}

function handlePostCreated() {
  refresh()
}

function handleRefreshClick(_: MouseEvent) {
  return refresh()
}

function isPostExpanded(postId: number | string) {
  return typeof postId === 'number' && expandedPosts.value.has(postId)
}

function togglePostComments(postId: number | string) {
  if (typeof postId !== 'number') return
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

async function loadPostComments(postId: number) {
  commentsLoading[postId] = true
  try {
    const response = await $fetch<{ comments: ApiPostComment[] }>(`${apiBase}/protected/v1/posts/${postId}/comments`, {
      credentials: 'include'
    })
    commentsCache[postId] = normalizeComments(response.comments)
  } catch (error) {
    toast.add({
      title: 'Unable to load comments',
      description: extractErrorMessage(error) || 'Please try again later.',
      color: 'error'
    })
  } finally {
    commentsLoading[postId] = false
  }
}

async function submitComment(postId: number | string) {
  if (typeof postId !== 'number') return
  const draft = commentDrafts[postId]?.trim()
  if (!draft) {
    toast.add({ title: 'Comment required', color: 'error' })
    return
  }
  commentSubmitting[postId] = true
  try {
    await $fetch(`${apiBase}/protected/v1/posts/${postId}/comments`, {
      method: 'POST',
      credentials: 'include',
      body: { content: draft }
    })
    commentDrafts[postId] = ''
    const rawPosts = data.value?.posts
    const targetPost = Array.isArray(rawPosts) ? rawPosts.find(post => post.id === postId) : null
    if (targetPost && typeof targetPost.comment_count === 'number') {
      targetPost.comment_count += 1
    }
    await loadPostComments(postId)
    toast.add({ title: 'Comment added' })
  } catch (error) {
    toast.add({
      title: 'Unable to comment',
      description: extractErrorMessage(error) || 'Please try again later.',
      color: 'error'
    })
  } finally {
    commentSubmitting[postId] = false
  }
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
              <div class="space-y-3">
                <div class="flex flex-wrap items-center gap-3 text-sm text-muted">
                  <div class="flex items-center gap-1">
                    <UIcon name="i-lucide-message-square" class="size-4" />
                    <span>{{ post.commentCount }} comments</span>
                  </div>
                  <UButton size="xs" variant="ghost" @click="togglePostComments(post.id)">
                    {{ isPostExpanded(post.id) ? 'Hide comments' : 'View comments' }}
                  </UButton>
                </div>

                <div v-if="isPostExpanded(post.id)" class="space-y-3 rounded-2xl border border-default/60 p-4">
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
                      v-model="commentDrafts[post.id]"
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
              </div>
            </template>
          </UCard>
        </div>
      </div>
    </template>
  </UDashboardPanel>
</template>
