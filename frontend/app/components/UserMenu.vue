<script setup lang="ts">
import type { DropdownMenuItem } from '@nuxt/ui'
import { normalizeAvatar } from '~/utils'

defineProps<{
  collapsed?: boolean
}>()

const runtimeConfig = useRuntimeConfig()
const apiBase = runtimeConfig.public?.apiBase || 'http://localhost:8080'
const toast = useToast()
const router = useRouter()

const { session, hydrate } = useSession()
if (import.meta.client) {
  hydrate()
}

const avatarSrc = computed(() => normalizeAvatar(session.value.avatar))
const displayName = computed(() => session.value.full_name || 'Account')
const initials = computed(() => {
  const parts = displayName.value.split(/\s+/).filter(Boolean)
  if (!parts.length) return 'U'
  return parts.map(part => part[0]?.toUpperCase()).join('').slice(0, 2)
})
const buttonConfig = computed(() => ({
  avatar: {
    src: avatarSrc.value ?? undefined,
    text: initials.value,
    alt: displayName.value
  }
}))

const handleLogout = async () => {
  try {
    await $fetch('/protected/v1/logout', {
      method: 'POST',
      baseURL: apiBase,
      credentials: 'include'
    })
  } catch (err: unknown) {
    // Treat unauthorized/expired sessions as logged out
    const status = typeof err === 'object' && err !== null && 'status' in err
      ? (err as { status?: number }).status
      : undefined
    if (status !== 401 && status !== 403) {
      toast.add({
        title: 'Logout failed',
        description: 'Something went wrong while signing out',
        color: 'error'
      })
      console.error(err)
      return
    }
  }
  if (import.meta.client) {
    window.localStorage.removeItem('user_id')
  }
  await router.push('/signin')
}

const items = computed<DropdownMenuItem[][]>(() => [[{
  type: 'label',
  label: displayName.value,
  avatar: {
    src: avatarSrc.value ?? undefined,
    text: initials.value
  }
}], [{
  label: 'Profile',
  icon: 'i-lucide-user',
  to: '/profile'
}], [{
  label: 'Repository',
  icon: 'i-simple-icons-github',
  to: 'https://learn.reboot01.com/git/aalmadhoo/social-network',
  target: '_blank'
}, {
  label: 'Log out',
  icon: 'i-lucide-log-out',
  onSelect: handleLogout
}]])
</script>

<template>
  <UDropdownMenu
    :items="items"
    :content="{ align: 'center', collisionPadding: 12 }"
    :ui="{ content: collapsed ? 'w-48' : 'w-(--reka-dropdown-menu-trigger-width)' }"
  >
    <UButton
      v-bind="buttonConfig"
      :label="collapsed ? undefined : displayName"
      color="neutral"
      variant="ghost"
      block
      :square="collapsed"
      class="data-[state=open]:bg-elevated"
      :trailing-icon="collapsed ? undefined : 'i-lucide-chevrons-up-down'"
      :ui="{ trailingIcon: 'text-dimmed' }"
    />

    <template #chip-leading="{ item }">
      <span
        :style="{
          '--chip-light': `var(--color-${(item as any).chip}-500)`,
          '--chip-dark': `var(--color-${(item as any).chip}-400)`
        }"
        class="ms-0.5 size-2 rounded-full bg-(--chip-light) dark:bg-(--chip-dark)"
      />
    </template>
  </UDropdownMenu>
</template>
