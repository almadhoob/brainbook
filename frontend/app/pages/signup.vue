<script setup lang="ts">
import * as z from 'zod'
import type { FormSubmitEvent } from '@nuxt/ui'

definePageMeta({
  layout: 'auth'
})

useSeoMeta({
  title: 'Sign up',
  description: 'Create an account to get started'
})

const toast = useToast()

const fields = [
  {
    name: 'f_name',
    type: 'text' as const,
    label: 'First Name',
    placeholder: 'Enter your first name'
  },
  {
    name: 'l_name',
    type: 'text' as const,
    label: 'Last Name',
    placeholder: 'Enter your last name'
  },
  {
    name: 'email',
    type: 'text' as const,
    label: 'Email',
    placeholder: 'Enter your email'
  },
  {
    name: 'password',
    label: 'Password',
    type: 'password' as const,
    placeholder: 'Enter your password'
  },
  {
    name: 'dob',
    type: 'date' as const,
    label: 'Date of Birth',
    placeholder: 'Select your birth date'
  },
  {
    name: 'nickname',
    type: 'text' as const,
    label: 'Nickname (optional)',
    placeholder: 'Enter a nickname'
  },
  {
    name: 'bio',
    type: 'textarea' as const,
    label: 'About me (optional)',
    placeholder: 'Add a short bio'
  },
  {
    name: 'avatar',
    type: 'file' as const,
    label: 'Avatar (optional)',
    placeholder: 'Choose an avatar image',
    accept: 'image/png,image/jpeg,image/gif'
  }
]

const AVATAR_MAX_BYTES = 5 * 1024 * 1024

const schema = z.object({
  f_name: z.string().min(1, 'First name is required'),
  l_name: z.string().min(1, 'Last name is required'),
  email: z.string().email('Invalid email'),
  password: z.string().min(8, 'Must be at least 8 characters'),
  dob: z.string().regex(/^\d{4}-\d{2}-\d{2}$/, 'Select a valid date'),
  nickname: z.string().max(30, 'Nickname too long').optional().or(z.literal('')),
  bio: z.string().max(500, 'Bio limit exceeded (500 characters)').optional().or(z.literal('')),
  avatar: z.any().optional()
})

type Schema = z.output<typeof schema>

async function fileToBase64(file: File): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => {
      const result = typeof reader.result === 'string' ? reader.result : ''
      const payload = result.includes(',') ? result.split(',')[1] : result
      if (!payload) {
        reject(new Error('empty-file'))
        return
      }
      resolve(payload)
    }
    reader.onerror = () => reject(reader.error ?? new Error('read-error'))
    reader.readAsDataURL(file)
  })
}

async function onSubmit(payload: FormSubmitEvent<Schema>) {
  const dobDate = new Date(payload.data.dob)
  const dobUtc = new Date(Date.UTC(dobDate.getFullYear(), dobDate.getMonth(), dobDate.getDate()))
  const dobRFC3339 = dobUtc.toISOString()

  const nicknameValue = (payload.data.nickname || '').trim()
  const bioValue = (payload.data.bio || '').trim()
  let avatarBase64: string | undefined
  const avatarFile = payload.data.avatar
  const fileCandidate = Array.isArray(avatarFile) ? avatarFile[0] : avatarFile
  if (fileCandidate instanceof File) {
    if (!['image/jpeg', 'image/png', 'image/gif'].includes(fileCandidate.type)) {
      toast.add({ title: 'Invalid avatar', description: 'Only JPEG, PNG, or GIF images are allowed.', color: 'error' })
      return
    }
    if (fileCandidate.size > AVATAR_MAX_BYTES) {
      toast.add({ title: 'Invalid avatar', description: 'Avatar must be 5MB or smaller.', color: 'error' })
      return
    }
    avatarBase64 = await fileToBase64(fileCandidate)
  }

  const body = {
    email: payload.data.email,
    password: payload.data.password,
    f_name: payload.data.f_name,
    l_name: payload.data.l_name,
    dob: dobRFC3339,
    nickname: nicknameValue || undefined,
    bio: bioValue || undefined,
    avatar: avatarBase64
  }
  try {
    await $fetch('/v1/register', {
      method: 'POST',
      baseURL: typeof (useRuntimeConfig() as { public: { apiBase?: string } }).public.apiBase === 'string'
        ? (useRuntimeConfig() as { public: { apiBase?: string } }).public.apiBase
        : 'http://localhost:8080',
      body,
      credentials: 'include'
    })
    toast.add({ title: 'Account created', description: 'Welcome!' })
    resetOptionalFields()
    await navigateTo('/')
  } catch (err: unknown) {
    const errorMsg = (err as { data?: { Error?: string } })?.data?.Error || 'Registration error'
    toast.add({ title: 'Signup failed', description: errorMsg, color: 'error' })
  }
}
</script>
<template>
  <UAuthForm
    :fields="fields"
    :schema="schema"
    title="Create an account"
    :submit="{ label: 'Continue' }"
    @submit="onSubmit"
  >
    <template #description>
      Already have an account? <ULink
        to="/signin"
        class="text-primary font-medium"
      >Sign in</ULink>.
    </template>

    <template #footer>
      <div class="text-sm text-muted">
        By signing up, you agree to our <ULink
          to="/"
          class="text-primary font-medium"
        >Terms of Service</ULink>.
      </div>
    </template>
  </UAuthForm>
</template>
