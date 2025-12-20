<script setup lang="ts">
const props = defineProps<{ apiBase: string }>()
const emit = defineEmits<{ (event: 'created'): void }>()

const open = ref(false)
const loading = ref(false)

const form = reactive({
  content: '',
  visibility: 'public',
  allowedUserIds: ''
})

const errors = reactive({
  content: '',
  file: ''
})

const fileName = ref('')
const filePayload = ref<string | undefined>(undefined)
const toast = useToast()
const MAX_FILE_SIZE = 10 * 1024 * 1024

const limitedInfoOpen = ref(false)

watch(() => form.visibility, (v) => {
  if (v === 'private') {
    limitedInfoOpen.value = true
    form.visibility = 'almost_private'
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
  form.allowedUserIds = ''
  errors.content = ''
  errors.file = ''
  fileName.value = ''
  filePayload.value = undefined
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

  const content = form.content.trim()
  if (!content) {
    errors.content = 'Content is required.'
    return
  }

  const visibility = form.visibility
  const allowedIds: number[] = []
  if (visibility === 'private') {
    const cleaned = form.allowedUserIds
      .split(',')
      .map(part => Number.parseInt(part.trim(), 10))
      .filter(n => Number.isInteger(n) && n > 0)
    if (cleaned.length === 0) {
      errors.content = 'Provide at least one allowed follower ID for private posts.'
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
              :rows="5"
              autoresize
              class="w-full"
            />
            <span class="absolute bottom-2 right-2 text-xs text-neutral-500">
              {{ form.content.length }} chars
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

        <UFieldGroup label="Attachment" :description="fileName || 'Optional image or file'" :error="errors.file">
          <input
            type="file"
            accept="image/*,video/*"
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

  <!-- Popup shown when 'Limited' is selected -->
  <UModal
    v-model:open="limitedInfoOpen"
    title="Limited visibility"
    description="Only your followers can view limited posts."
  >
    <template #body>
        <div class="flex justify-end">
          <UButton color="primary" @click="limitedInfoOpen = false">Got it</UButton>
        </div>
    </template>
  </UModal>
</template>
