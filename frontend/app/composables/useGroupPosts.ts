import { extractErrorMessage, formatDate, toDataUrl, initialsFromName, buildFullName } from './useGroupHelpers'

export interface ApiGroupPost {
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

export interface ApiGroupComment {
  id?: number
  user_id?: number
  f_name?: string | null
  l_name?: string | null
  avatar?: string | null
  content?: string | null
  created_at?: string | null
}

export interface GroupPostItem {
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

export interface GroupComment {
  id: number
  content: string
  formattedCreatedAt: string
  authorName: string
  authorInitials: string
  avatarSrc?: string
}

export const MAX_GROUP_POST_LENGTH = 350
export const MAX_GROUP_COMMENT_LENGTH = 350

export function useGroupPosts(apiBase: string, groupId: Ref<number | null>) {
  const toast = useToast()

  const postsCache = reactive<Record<number, GroupPostItem[]>>({})
  const postsLoading = reactive<Record<number, boolean>>({})
  const createPostLoading = ref(false)

  const posts = computed(() =>
    groupId.value ? postsCache[groupId.value] ?? [] : []
  )

  const isLoading = computed(() =>
    groupId.value ? postsLoading[groupId.value] ?? false : false
  )

  const newPostForm = reactive({
    content: '',
    file: null as string | null
  })

  const commentsCache = reactive<Record<number, GroupComment[]>>({})
  const commentsLoading = reactive<Record<number, boolean>>({})
  const newCommentDrafts = reactive<Record<number, string>>({})
  const commentSubmitting = reactive<Record<number, boolean>>({})
  const expandedPosts = ref(new Set<number>())

  const postCount = computed(() => newPostForm.content.length)

  async function loadPosts() {
    if (!groupId.value) return

    const gid = groupId.value
    postsLoading[gid] = true
    try {
      const response = await $fetch<{ posts: ApiGroupPost[] }>(
        `${apiBase}/protected/v1/groups/${gid}/posts`,
        { credentials: 'include' }
      )
      postsCache[gid] = normalizePosts(response.posts)
    } catch (error) {
      toast.add({
        title: 'Unable to load posts',
        description: extractErrorMessage(error) || 'Please try again later.',
        color: 'error'
      })
    } finally {
      postsLoading[gid] = false
    }
  }

  async function submitPost() {
    if (!groupId.value) return

    const content = newPostForm.content.trim()
    if (!content) {
      toast.add({
        title: 'Post content required',
        description: 'Share something with your group first.',
        color: 'error'
      })
      return
    }

    if (content.length > MAX_GROUP_POST_LENGTH) {
      toast.add({
        title: 'Post too long',
        description: `Keep it under ${MAX_GROUP_POST_LENGTH} characters.`,
        color: 'error'
      })
      return
    }

    createPostLoading.value = true
    try {
      await $fetch(`${apiBase}/protected/v1/groups/${groupId.value}/create`, {
        method: 'POST',
        credentials: 'include',
        body: { content, file: newPostForm.file }
      })

      toast.add({ title: 'Post published', description: 'Your group can see it now.' })
      newPostForm.content = ''
      newPostForm.file = null
      await loadPosts()
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

  async function loadPostComments(postId: number) {
    if (!groupId.value) return

    commentsLoading[postId] = true
    try {
      const response = await $fetch<{ comments: ApiGroupComment[] }>(
        `${apiBase}/protected/v1/groups/${groupId.value}/posts/${postId}/comments`,
        { credentials: 'include' }
      )
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

  async function submitComment(postId: number) {
    if (!groupId.value) return

    const draft = newCommentDrafts[postId]?.trim()
    if (!draft) {
      toast.add({ title: 'Comment required', color: 'error' })
      return
    }

    if (draft.length > MAX_GROUP_COMMENT_LENGTH) {
      toast.add({
        title: 'Comment too long',
        description: `Max ${MAX_GROUP_COMMENT_LENGTH} characters.`,
        color: 'error'
      })
      return
    }

    commentSubmitting[postId] = true
    try {
      await $fetch(
        `${apiBase}/protected/v1/groups/${groupId.value}/posts/${postId}/comments`,
        {
          method: 'POST',
          credentials: 'include',
          body: { content: draft }
        }
      )

      newCommentDrafts[postId] = ''
      const targetPost = posts.value.find(post => post.id === postId)
      if (targetPost) {
        targetPost.commentCount += 1
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

  function getCommentCount(postId: number) {
    const val = newCommentDrafts[postId]
    return typeof val === 'string' ? val.length : 0
  }

  function clearPostsState() {
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
    newPostForm.content = ''
    newPostForm.file = null
  }

  function normalizePosts(posts?: ApiGroupPost[]): GroupPostItem[] {
    if (!Array.isArray(posts)) return []

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
    if (!Array.isArray(comments)) return []

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

  return {
    posts,
    postsLoading: isLoading,
    createPostLoading,
    newPostForm,
    postCount,
    commentsCache,
    commentsLoading,
    newCommentDrafts,
    commentSubmitting,
    expandedPosts,
    loadPosts,
    submitPost,
    loadPostComments,
    submitComment,
    togglePostComments,
    isPostExpanded,
    getCommentCount,
    clearPostsState
  }
}
