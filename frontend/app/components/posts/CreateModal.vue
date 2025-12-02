<script setup lang="ts">
const props = defineProps<{ apiBase: string }>()
const emit = defineEmits<{ (event: 'created'): void }>()

const open = ref(false)
const loading = ref(false)

const form = reactive({
  content: ''
})

const errors = reactive({
  content: '',
  file: ''
})

const fileName = ref('')
const filePayload = ref<string | undefined>(undefined)
const toast = useToast()
const MAX_FILE_SIZE = 5 * 1024 * 1024 // 5MB limit to prevent oversized uploads

watch(open, (value) => {
  if (!value) {
    resetForm()
  }
})

function resetForm() {
  form.content = ''
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
    errors.file = 'File exceeds 5 MB limit.'
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

  try {
    loading.value = true
    await $fetch(`${props.apiBase}/protected/v1/posts`, {
      method: 'POST',
      credentials: 'include',
      body: {
        content,
        file: filePayload.value
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
  <UModal v-model:open="open">
    <UButton icon="i-lucide-plus">
      New post
    </UButton>

    <template #header>
      <h3 class="text-lg font-semibold">
        Share an update
      </h3>
    </template>

    <template #body>
      <form class="space-y-4" @submit.prevent="handleSubmit">
        <UFormGroup label="Content" :error="errors.content">
          <UTextarea
            v-model="form.content"
            placeholder="What's on your mind?"
            :rows="5"
            autoresize
          />
        </UFormGroup>

        <UFormGroup label="Attachment" :description="fileName || 'Optional image or file'" :error="errors.file">
          <input
            type="file"
            accept="image/*,video/*"
            class="block w-full text-sm"
            @change="handleFileChange"
          >
        </UFormGroup>

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
