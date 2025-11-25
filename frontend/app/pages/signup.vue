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
    name: 'name',
    type: 'text' as const,
    label: 'Name',
    placeholder: 'Enter your name'
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
    type: 'text' as const,
    label: 'Date of Birth',
    placeholder: 'YYYY-MM-DD'
  }
]

const providers = [{
  label: 'Google',
  icon: 'i-simple-icons-google',
  onClick: () => {
    toast.add({ title: 'Google', description: 'Login with Google' })
  }
}, {
  label: 'GitHub',
  icon: 'i-simple-icons-github',
  onClick: () => {
    toast.add({ title: 'GitHub', description: 'Login with GitHub' })
  }
}]

const schema = z.object({
  name: z.string().min(1, 'Name is required'),
  email: z.string().email('Invalid email'),
  password: z.string().min(8, 'Must be at least 8 characters'),
  dob: z.string().min(1, 'Date of birth is required')
})

type Schema = z.output<typeof schema>

async function onSubmit(payload: FormSubmitEvent<Schema>) {
  // Prepare registration payload
  // Format dob as RFC3339 (YYYY-MM-DDT00:00:00Z)
  let dobRFC3339 = payload.data.dob
  if (/^\d{4}-\d{2}-\d{2}$/.test(payload.data.dob)) {
    dobRFC3339 = payload.data.dob + 'T00:00:00Z'
  }
  const body = {
    email: payload.data.email,
    password: payload.data.password,
    f_name: payload.data.name.split(' ')[0] || '',
    l_name: payload.data.name.split(' ').slice(1).join(' ') || '',
    dob: dobRFC3339
    // avatar, nickname, about_me can be added here
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
    :providers="providers"
    title="Create an account"
    :submit="{ label: 'Create account' }"
    @submit="onSubmit"
  >
    <template #description>
      Already have an account? <ULink
        to="/login"
        class="text-primary font-medium"
      >Login</ULink>.
    </template>

    <template #footer>
      By signing up, you agree to our <ULink
        to="/"
        class="text-primary font-medium"
      >Terms of Service</ULink>.
    </template>
  </UAuthForm>
</template>
