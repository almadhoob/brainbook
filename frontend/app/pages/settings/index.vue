<script setup lang="ts">
import * as z from 'zod'
import type { FormSubmitEvent } from '@nuxt/ui'
import { normalizeAvatar } from '~/utils'

interface ProfileResponse {
  user_id?: number
  full_name?: string
  email?: string
  nickname?: string
  bio?: string
  is_public?: boolean
  avatar?: string | null
}

const fileRef = ref<HTMLInputElement>()
const { session, hydrate } = useSession()
const toast = useToast()

const runtimeConfig = useRuntimeConfig()
const apiBase = typeof runtimeConfig.public?.apiBase === 'string' && runtimeConfig.public.apiBase.length > 0
  ? runtimeConfig.public.apiBase
  : 'http://localhost:8080'

const profileSchema = z.object({
  nickname: z.string().max(50, 'Nickname must be 50 characters or less').optional(),
  bio: z.string().max(500, 'Bio limit exceeded (500 characters)').optional(),
  is_public: z.boolean().optional(),
  avatar: z.string().optional()
})

type ProfileSchema = z.output<typeof profileSchema>

const profile = reactive<Partial<ProfileSchema>>({
  nickname: '',
  bio: '',
  is_public: true
})

const avatarPreview = ref<string | undefined>(undefined)
const avatarPayload = ref<string | null>(null)

const isLoadingProfile = ref(false)
const isSaving = ref(false)
const loadError = ref<string | null>(null)

const errors = reactive({ avatar: '' })
const MAX_AVATAR_SIZE = 5 * 1024 * 1024
const ALLOWED_AVATAR_TYPES = ['image/jpeg', 'image/png', 'image/gif']

async function loadProfile() {
  const userId = session.value.user_id
  if (!userId || isLoadingProfile.value) return

  isLoadingProfile.value = true
  loadError.value = null
  try {
    const data = await $fetch<ProfileResponse>(`${apiBase}/protected/v1/profile/user/${userId}`, {
      credentials: 'include'
    })

    profile.nickname = data.nickname ?? ''
    profile.bio = data.bio ?? ''
    profile.is_public = typeof data.is_public === 'boolean' ? data.is_public : true
    avatarPreview.value = normalizeAvatar(data.avatar)
    avatarPayload.value = null
  } catch (err) {
    loadError.value = 'Unable to load profile settings.'
    console.error(err)
  } finally {
    isLoadingProfile.value = false
  }
}

async function onSubmit(_event: FormSubmitEvent<ProfileSchema>) {
  if (isSaving.value) return
  if (!session.value.user_id) {
    toast.add({ title: 'Sign in required', description: 'Please sign in to update your profile.', color: 'error' })
    return
  }

  const payload: Record<string, unknown> = {}
  if (typeof profile.nickname === 'string' && profile.nickname.trim().length > 0) {
    payload.nickname = profile.nickname.trim()
  }
  if (typeof profile.bio === 'string' && profile.bio.trim().length > 0) {
    payload.bio = profile.bio.trim()
  }
  if (typeof profile.is_public === 'boolean') {
    payload.is_public = profile.is_public
  }
  if (avatarPayload.value) {
    payload.avatar = avatarPayload.value
  }

  isSaving.value = true
  try {
    await $fetch(`${apiBase}/protected/v1/profile/update`, {
      method: 'POST',
      credentials: 'include',
      body: payload
    })
    toast.add({
      title: 'Success',
      description: 'Your settings have been updated.',
      icon: 'i-lucide-check',
      color: 'success'
    })
    avatarPayload.value = null
    await hydrate(true)
  } catch (err) {
    toast.add({
      title: 'Update failed',
      description: 'We could not update your profile settings.',
      color: 'error'
    })
    console.error(err)
  } finally {
    isSaving.value = false
  }
}

function toBase64Payload(dataUrl: string) {
  const parts = dataUrl.split(',')
  return parts.length > 1 ? parts[1] : dataUrl
}

async function onFileChange(e: Event) {
  errors.avatar = ''
  const input = e.target as HTMLInputElement

  const file = input.files?.[0]
  if (!file) {
    return
  }

  if (!ALLOWED_AVATAR_TYPES.includes(file.type)) {
    errors.avatar = 'Only JPEG, PNG, or GIF images are allowed.'
    input.value = ''
    return
  }

  if (file.size > MAX_AVATAR_SIZE) {
    errors.avatar = 'File exceeds 5 MB limit.'
    input.value = ''
    return
  }

  const dataUrl = await new Promise<string>((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => resolve(String(reader.result))
    reader.onerror = () => reject(new Error('Unable to read file'))
    reader.readAsDataURL(file)
  })

  avatarPreview.value = dataUrl
  avatarPayload.value = toBase64Payload(dataUrl)
}

function onFileClick() {
  fileRef.value?.click()
}

if (import.meta.client) {
  hydrate(true).then(loadProfile)
}
</script>

<template>
  <UForm
    id="settings"
    :schema="profileSchema"
    :state="profile"
    @submit="onSubmit"
  >
    <UPageCard
      title="Profile"
      description="These informations will be displayed publicly."
      variant="naked"
      orientation="horizontal"
      class="mb-4"
      aria-describedby="undefined"
    >
      <UButton
        form="settings"
        label="Save changes"
        color="neutral"
        type="submit"
        class="w-fit lg:ms-auto"
        :loading="isSaving"
      />
    </UPageCard>

    <UPageCard variant="subtle">
      <UFormField
        name="nickname"
        label="Username"
        description="Your unique username for your profile URL."
        class="flex max-sm:flex-col justify-between items-start gap-4"
      >
        <UInput
          v-model="profile.nickname"
          autocomplete="off"
        />
      </UFormField>
      <USeparator />
      <UFormField
        name="is_public"
        label="Profile visibility"
        description="Control who can view your profile details."
        class="flex max-sm:flex-col justify-between items-start gap-4"
      >
        <USwitch v-model="profile.is_public" />
      </UFormField>
      <USeparator />
      <UFormField
        name="avatar"
        label="Avatar"
        description="JPG, GIF or PNG. 5MB Max."
        :error="errors.avatar"
        class="flex max-sm:flex-col justify-between sm:items-center gap-4"
      >
        <div class="flex flex-wrap items-center gap-3">
          <UAvatar
            :src="avatarPreview"
            :alt="profile.nickname || session.full_name"
            size="lg"
          />
          <UButton
            label="Choose"
            color="neutral"
            @click="onFileClick"
          />
          <input
            ref="fileRef"
            type="file"
            class="hidden"
            accept=".jpg, .jpeg, .png, .gif"
            @change="onFileChange"
          >
        </div>
      </UFormField>
      <USeparator />
      <UFormField
        name="bio"
        label="Bio"
        description="Brief description for your profile. URLs are hyperlinked."
        class="flex max-sm:flex-col justify-between items-start gap-4"
        :ui="{ container: 'w-full' }"
      >
        <UTextarea
          v-model="profile.bio"
          :rows="5"
          autoresize
          class="w-full"
        />
      </UFormField>
    </UPageCard>
  </UForm>
</template>
