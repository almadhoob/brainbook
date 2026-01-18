<script setup lang="ts">
import { useFollowers } from '~/composables/useFollowers'
import { normalizeAvatar } from '~/utils'

const props = defineProps<{ apiBase: string }>()
const emit = defineEmits<{ (event: 'created'): void }>()

const open = ref(false)
const loading = ref(false)

const form = reactive({
  content: '',
  visibility: 'public',
  allowedUserIds: [] as number[]
})

const errors = reactive({
  content: '',
  file: '',
  allowedUsers: ''
})

const fileName = ref('')
const filePayload = ref<string | undefined>(undefined)
const toast = useToast()
const MAX_FILE_SIZE = 10 * 1024 * 1024
const ALLOWED_IMAGE_TYPES = ['image/jpeg', 'image/png', 'image/gif']

// Content max length and counter (similar to group modal)
const MAX_CONTENT = 350
const contentCount = computed(() => form.content.length)

const followerSearch = ref('')

const { session, hydrate } = useSession()
const userId = computed(() => session.value.user_id)
const {
  followers,
  loading: followersLoading,
  error: followersError,
  loadFollowers
} = useFollowers(props.apiBase, userId)

const filteredFollowers = computed(() => {
  const query = followerSearch.value.trim().toLowerCase()
  if (!query) return followers.value
  return followers.value.filter((user) => {
    const fullName = `${user.f_name ?? ''} ${user.l_name ?? ''}`.trim().toLowerCase()
    return fullName.includes(query)
  })
})

watch(() => form.visibility, async (value) => {
  errors.allowedUsers = ''
  if (value === 'private') {
    await ensureFollowersLoaded()
  } else {
    form.allowedUserIds = []
  }
})

watch(open, (value) => {
  if (!value) {
    resetForm()
  }
})

function resetForm() {
  form.content = ''
  form.visibility = 'public'
  form.allowedUserIds = []
  errors.content = ''
  errors.file = ''
  errors.allowedUsers = ''
  fileName.value = ''
  filePayload.value = undefined
  followerSearch.value = ''
}

async function ensureFollowersLoaded() {
  if (!session.value.user_id) {
    await hydrate(true)
  }
  await loadFollowers()
}

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

async function handleFileChange(event: Event) {
  errors.file = ''
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]

  if (!file) {
    fileName.value = ''
    filePayload.value = undefined
    return
  }

  if (file.size > MAX_FILE_SIZE) {
    errors.file = 'File exceeds 10 MB limit.'
    return
  }

  if (!ALLOWED_IMAGE_TYPES.includes(file.type)) {
    errors.file = 'Only JPEG, PNG, or GIF images are allowed.'
    return
  }

  try {
    filePayload.value = await fileToBase64(file)
    fileName.value = file.name
  } catch (error) {
    errors.file = 'Unable to read file.'
    console.error(error)
  }
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

async function handleSubmit() {
  errors.content = ''
  errors.file = ''
  errors.allowedUsers = ''

  const content = form.content.trim()
  if (!content) {
    errors.content = 'Content is required.'
    return
  }
  // Enforce max length as an extra safety measure (same style as group modal)
  if (content.length > MAX_CONTENT) {
    errors.content = `Content must be at most ${MAX_CONTENT} characters.`
    return
  }

  const visibility = form.visibility
  const allowedIds: number[] = []
  if (visibility === 'private') {
    const cleaned = form.allowedUserIds.filter(id => Number.isInteger(id) && id > 0)
    if (cleaned.length === 0) {
      errors.allowedUsers = 'Select at least one follower for limited posts.'
      return
    }
    allowedIds.push(...cleaned)
  }

  try {
    loading.value = true
    await $fetch(`${props.apiBase}/protected/v1/posts`, {
      method: 'POST',
      credentials: 'include',
      body: {
        content,
        file: filePayload.value,
        visibility,
        allowed_user_ids: allowedIds
      }
    })
    toast.add({ title: 'Post created', description: 'Your update is live.' })
    emit('created')
    open.value = false
  } catch (error) {
    toast.add({
      title: 'Unable to create post',
      description: extractErrorMessage(error) || 'Please try again later.',
      color: 'error'
    })
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <UModal
    v-model:open="open"
    title="Share an update"
    description="Write a post and optionally attach media."
  >
    <UButton icon="i-lucide-plus">
      New post
    </UButton>

    <template #body>
      <form class="space-y-4 w-full" @submit.prevent="handleSubmit">
        <UFieldGroup
          label="Content"
          :error="errors.content"
          class="w-full max-w-none"
          :ui="{ container: 'w-full flex flex-col', label: 'w-full', wrapper: 'w-full max-w-none' }"
        >
          <div class="relative w-full">
            <UTextarea
              v-model="form.content"
              placeholder="What's on your mind?"
              :rows="6"
              :maxlength="MAX_CONTENT"
              autoresize
              class="w-full"
            />
            <span class="pointer-events-none absolute bottom-2 right-2 text-xs text-neutral-500 z-10">
              {{ contentCount }} / {{ MAX_CONTENT }}
            </span>
          </div>
        </UFieldGroup>

        <div class="mt-2 flex flex-col gap-2 items-start">
          <USelect
            v-model="form.visibility"
            class="max-w-xs"
            :items="[
              { label: 'Public (everyone)', value: 'public' },
              { label: 'Private (followers only)', value: 'almost_private' },
              { label: 'Limited (select followers)', value: 'private' }
            ]"
          />
        </div>

        <div v-if="form.visibility === 'private'" class="rounded-xl border border-default/60 p-4 space-y-3">
          <div class="flex items-center justify-between">
            <p class="text-sm font-semibold">
              Choose followers who can view this post
            </p>
            <UButton
              size="xs"
              color="neutral"
              variant="ghost"
              :loading="followersLoading"
              @click="ensureFollowersLoaded"
            >
              Refresh
            </UButton>
          </div>

          <UInput
            v-model="followerSearch"
            placeholder="Search followers"
            class="max-w-sm"
          />

          <p v-if="errors.allowedUsers" class="text-xs text-red-500">
            {{ errors.allowedUsers }}
          </p>

          <div v-if="followersLoading" class="text-sm text-muted">
            Loading followers...
          </div>
          <div v-else-if="followersError" class="text-sm text-red-500">
            {{ followersError }}
          </div>
          <div v-else-if="filteredFollowers.length === 0" class="text-sm text-muted">
            No followers found.
          </div>
          <div v-else class="grid gap-2 sm:grid-cols-2">
            <label
              v-for="user in filteredFollowers"
              :key="user.user_id"
              class="flex items-center gap-3 rounded-lg border border-default/60 p-2 cursor-pointer hover:border-primary/60"
            >
              <input
                v-model="form.allowedUserIds"
                type="checkbox"
                :value="user.user_id"
                class="h-4 w-4 rounded border-default/60"
              >
              <UAvatar :src="normalizeAvatar(user.avatar)" :text="user.f_name?.[0] || 'U'" size="xs" />
              <span class="text-sm">
                {{ `${user.f_name ?? ''} ${user.l_name ?? ''}`.trim() || 'Unknown user' }}
              </span>
            </label>
          </div>
        </div>

        <UFieldGroup label="Attachment" :description="fileName || 'Optional image or file'" :error="errors.file">
          <input
            type="file"
            accept="image/jpeg,image/png,image/gif"
            class="block w-full text-sm"
            @change="handleFileChange"
          >
        </UFieldGroup>

        <div class="flex justify-end gap-2">
          <UButton
            type="button"
            color="neutral"
            variant="ghost"
            @click="open = false"
          >
            Cancel
          </UButton>
          <UButton type="submit" :loading="loading">
            Create post
          </UButton>
        </div>
      </form>
    </template>
  </UModal>
</template>
