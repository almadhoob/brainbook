<script setup lang="ts">
import * as z from 'zod'
import type { FormSubmitEvent } from '@nuxt/ui'

definePageMeta({
  layout: 'auth'
})

useSeoMeta({
  title: 'Login',
  description: 'Login to your account to continue'
})

const toast = useToast()

const fields = [{
  name: 'email',
  type: 'text' as const,
  label: 'Email',
  placeholder: 'Enter your email',
  required: true
}, {
  name: 'password',
  label: 'Password',
  type: 'password' as const,
  placeholder: 'Enter your password'
}, {
  name: 'remember',
  label: 'Remember me',
  type: 'checkbox' as const
}]

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
  email: z.string().email('Invalid email'),
  password: z.string().min(8, 'Must be at least 8 characters')
})

type Schema = z.output<typeof schema>

async function onSubmit(payload: FormSubmitEvent<Schema>) {
  // Prepare login payload
  const body = {
    identifier: payload.data.email,
    password: payload.data.password
  }
  try {
    const publicConfig = useRuntimeConfig().public as { apiBase?: string }
    const apiBase = publicConfig.apiBase && typeof publicConfig.apiBase === 'string' && publicConfig.apiBase.length > 0
      ? publicConfig.apiBase
      : 'http://localhost:8080'
    await $fetch('/v1/login', {
      method: 'POST',
      baseURL: apiBase,
      body,
      credentials: 'include'
    })
    // Fetch current user profile to get user_id
    let userId: string | undefined
    try {
      // Try to get user list and find the matching email
      const userList = await $fetch<{ users: Array<{ user_id: string, email: string }> }>(
        '/protected/v1/user-list', {
          method: 'GET',
          baseURL: apiBase,
          credentials: 'include'
        }
      )
      const user = userList.users.find(u => u.email === payload.data.email)
      if (user) {
        userId = user.user_id
        localStorage.setItem('user_id', userId)
      }
    } catch {
      // fallback: do not set user_id
    }
    toast.add({ title: 'Login successful', description: 'Welcome back!' })
    await navigateTo('/')
  } catch (err: unknown) {
    let errorMsg = 'Authentication error'
    type ErrorResponse = { data?: { Error?: string } }
    if (typeof err === 'object' && err !== null && 'data' in err) {
      const e = err as ErrorResponse
      if (e.data && typeof e.data.Error === 'string') {
        errorMsg = e.data.Error
      }
    }
    toast.add({ title: 'Login failed', description: errorMsg, color: 'error' })
  }
}
</script>

<template>
  <UAuthForm
    :fields="fields"
    :schema="schema"
    :providers="providers"
    title="Welcome back"
    icon="i-lucide-lock"
    @submit="onSubmit"
  >
    <template #description>
      Don't have an account? <ULink
        to="/signup"
        class="text-primary font-medium"
      >Sign up</ULink>.
    </template>

    <template #password-hint>
      <ULink
        to="/"
        class="text-primary font-medium"
        tabindex="-1"
      >Forgot password?</ULink>
    </template>

    <template #footer>
      By signing in, you agree to our <ULink
        to="/"
        class="text-primary font-medium"
      >Terms of Service</ULink>.
    </template>
  </UAuthForm>
</template>
