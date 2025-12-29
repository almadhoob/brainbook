<script setup lang="ts">
const props = defineProps<{ apiBase: string }>()
const emit = defineEmits<{ (event: 'created'): void }>()

const open = ref(false)
const form = reactive({
  title: '',
  description: ''
})
const errors = reactive({
  title: '',
  description: ''
})
const loading = ref(false)
const toast = useToast()

watch(open, (value) => {
  if (!value) {
    resetForm()
  }
})

function resetForm() {
  form.title = ''
  form.description = ''
  errors.title = ''
  errors.description = ''
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

const MAX_TITLE = 20
const MAX_DESCRIPTION = 50
const titleCount = computed(() => form.title.length)
const descriptionCount = computed(() => form.description.length)

async function handleSubmit() {
  errors.title = ''
  errors.description = ''

  const title = form.title.trim()
  const description = form.description.trim()

  if (!title) {
    errors.title = 'Title is required.'
  }
  if (!description) {
    errors.description = 'Description is required.'
  }
  // Enforce max length as an extra safety measure
  if (title.length > MAX_TITLE) {
    errors.title = `Title must be at most ${MAX_TITLE} characters.`
  }
  if (description.length > MAX_DESCRIPTION) {
    errors.description = `Description must be at most ${MAX_DESCRIPTION} characters.`
  }
  if (errors.title || errors.description) {
    return
  }

  try {
    loading.value = true
    await $fetch(`${props.apiBase}/protected/v1/groups`, {
      method: 'POST',
      credentials: 'include',
      body: {
        title,
        description
      }
    })
    toast.add({ title: 'Group created', description: 'You can now invite people to your group.' })
    emit('created')
    open.value = false
  } catch (error) {
    toast.add({
      title: 'Unable to create group',
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
    title="Create a group"
    description="Provide a title and description for your new group."
  >
    <UButton icon="i-lucide-plus">
      New group
    </UButton>

    <template #body>
      <form class="space-y-4" @submit.prevent="handleSubmit">
        <UFieldGroup
          label="Title"
          :error="errors.title"
          class="w-full max-w-none"
          :ui="{ container: 'w-full flex flex-col', label: 'w-full', wrapper: 'w-full max-w-none' }"
        >
          <div class="relative w-full">
            <UInput
              v-model="form.title"
              placeholder="e.g. AI Researchers, BioNauts, etc."
              :maxlength="MAX_TITLE"
              class="w-full"
            />
            <span class="pointer-events-none absolute bottom-2 right-2 text-xs text-neutral-500 z-10">
              {{ titleCount }} / {{ MAX_TITLE }}
            </span>
          </div>
        </UFieldGroup>

        <UFieldGroup
          label="Description"
          :error="errors.description"
          class="w-full max-w-none"
          :ui="{ container: 'w-full flex flex-col', label: 'w-full', wrapper: 'w-full max-w-none' }"
        >
          <div class="relative w-full">
            <UTextarea
              v-model="form.description"
              placeholder="e.g. A friendly space to discuss all things AI!"
              :maxlength="MAX_DESCRIPTION"
              :rows="2"
              autoresize
              class="w-full"
            />
            <span class="pointer-events-none absolute bottom-2 right-2 text-xs text-neutral-500">
              {{ descriptionCount }} / {{ MAX_DESCRIPTION }}
            </span>
          </div>
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
            Create group
          </UButton>
        </div>
      </form>
    </template>
  </UModal>
</template>
