<script setup lang="ts">
import { fileToBase64 } from '~/composables/useGroupHelpers'
import type { GroupPostItem, GroupComment } from '~/composables/useGroupPosts'
import { MAX_GROUP_POST_LENGTH, MAX_GROUP_COMMENT_LENGTH } from '~/composables/useGroupPosts'

interface Props {
  posts: GroupPostItem[]
  postsLoading: boolean
  createPostLoading: boolean
  postCount: number
  commentsCache: Record<number, GroupComment[]>
  commentsLoading: Record<number, boolean>
  commentSubmitting: Record<number, boolean>
  expandedPosts: Set<number>
}

interface Emits {
  (e: 'submit-post'): void
  (e: 'toggle-comments' | 'submit-comment', postId: number): void
}

defineProps<Props>()
const newPostForm = defineModel<{ content: string, file: string | null }>('newPostForm', { required: true })
const newCommentDrafts = defineModel<Record<number, string>>('newCommentDrafts', { required: true })
const emit = defineEmits<Emits>()

const toast = useToast()
const postFileInput = ref<HTMLInputElement | null>(null)

function openPostFilePicker() {
  postFileInput.value?.click()
}

function handlePostFileChange(event: Event) {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]
  if (!file) {
    newPostForm.value.file = null
    return
  }

  fileToBase64(file)
    .then((base64) => {
      newPostForm.value.file = base64
    })
    .catch(() => {
      toast.add({ title: 'Unable to process file', color: 'error' })
      newPostForm.value.file = null
    })
}

function getCommentCount(postId: number) {
  const val = newCommentDrafts.value[postId]
  return typeof val === 'string' ? val.length : 0
}
</script>

<template>
  <div class="space-y-6">
    <section>
      <h3 class="text-lg font-semibold">
        Posts
      </h3>
      <form class="mt-4 space-y-3" @submit.prevent="emit('submit-post')">
        <UFieldGroup label="Share an update" class="w-full">
          <div class="relative w-full">
            <UTextarea
              v-model="newPostForm.content"
              placeholder="What is new?"
              :maxlength="MAX_GROUP_POST_LENGTH"
              autoresize
              :rows="3"
              class="w-full"
            />
            <span class="pointer-events-none absolute bottom-2 right-2 text-xs text-neutral-500 z-10">
              {{ postCount }} / {{ MAX_GROUP_POST_LENGTH }}
            </span>
          </div>
        </UFieldGroup>
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

      <div v-if="postsLoading" class="mt-4 text-center text-sm text-muted">
        Loading posts...
      </div>
      <div v-else-if="!posts.length" class="mt-4 rounded-xl border border-default/60 p-4 text-sm text-muted">
        No posts yet. Start the conversation!
      </div>
      <div v-else class="mt-4 space-y-4">
        <UCard v-for="post in posts" :key="post.id" class="bg-elevated/30">
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
            <p class="text-sm whitespace-pre-line break-words">
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
              <UButton size="xs" variant="ghost" @click="emit('toggle-comments', post.id)">
                {{ expandedPosts.has(post.id) ? 'Hide comments' : 'View comments' }}
              </UButton>
            </div>
            <div v-if="expandedPosts.has(post.id)" class="mt-4 space-y-3 rounded-2xl border border-default/60 p-4">
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
                  <p class="mt-2 text-sm whitespace-pre-line break-words">
                    {{ comment.content }}
                  </p>
                </div>
              </div>
              <div class="flex gap-2">
                <div class="relative w-full">
                  <UTextarea
                    v-model="newCommentDrafts[post.id]"
                    placeholder="Write a comment"
                    :maxlength="MAX_GROUP_COMMENT_LENGTH"
                    autoresize
                    :rows="3"
                    class="w-full"
                  />
                  <span class="pointer-events-none absolute bottom-2 right-2 text-xs text-neutral-500 z-10">
                    {{ getCommentCount(post.id) }} / {{ MAX_GROUP_COMMENT_LENGTH }}
                  </span>
                </div>
                <UButton
                  color="primary"
                  :loading="commentSubmitting[post.id]"
                  @click="emit('submit-comment', post.id)"
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
