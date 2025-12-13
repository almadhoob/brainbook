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
  <UModal v-model:open="open">
    <UButton icon="i-lucide-plus">
      New group
    </UButton>

    <template #header>
      <h3 class="text-lg font-semibold">
        Create a group
      </h3>
    </template>

    <template #body>
      <form class="space-y-4" @submit.prevent="handleSubmit">
        <UFieldGroup label="Title" :error="errors.title">
          <UInput v-model="form.title" placeholder="AI Researchers" />
        </UFieldGroup>
        <UFieldGroup label="Description" :error="errors.description">
          <UTextarea v-model="form.description" placeholder="Describe the purpose of your group" />
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
