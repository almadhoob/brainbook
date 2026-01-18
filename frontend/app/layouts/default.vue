<script setup lang="ts">
import type { NavigationMenuItem } from '@nuxt/ui'

const { unreadCount } = useNotifications()

const open = ref(false)

const links = computed<NavigationMenuItem[][]>(() => [[{
  label: 'Home',
  icon: 'i-lucide-house',
  to: '/',
  onSelect: () => {
    open.value = false
  }
}, {
  label: 'Posts',
  icon: 'i-lucide-newspaper',
  to: '/posts',
  onSelect: () => {
    open.value = false
  }
}, {
  label: 'Groups',
  icon: 'i-lucide-users-round',
  to: '/groups',
  onSelect: () => {
    open.value = false
  }
}, {
  label: 'Users',
  icon: 'i-lucide-users',
  to: '/users',
  onSelect: () => {
    open.value = false
  }
}, {
  label: 'Messages',
  icon: 'i-lucide-inbox',
  to: '/messages',
  badge: unreadCount.value ? String(unreadCount.value) : undefined,
  onSelect: () => {
    open.value = false
  }
}, {
  label: 'Settings',
  to: '/settings',
  icon: 'i-lucide-settings',
  defaultOpen: true,
  type: 'trigger',
  children: [{
    label: 'General',
    to: '/settings',
    exact: true,
    onSelect: () => {
      open.value = false
    }
  }, {
    label: 'Security',
    to: '/settings/security',
    onSelect: () => {
      open.value = false
    }
  }]
}], []])

// onMounted(async () => {
//   const cookie = useCookie('cookie-consent')
//   if (cookie.value === 'accepted') {
//     return
//   }

//   toast.add({
//     title: 'We use first-party cookies to enhance your experience on our website.',
//     duration: 0,
//     close: false,
//     actions: [{
//       label: 'Accept',
//       color: 'neutral',
//       variant: 'outline',
//       onClick: () => {
//         cookie.value = 'accepted'
//       }
//     }, {
//       label: 'Opt out',
//       color: 'neutral',
//       variant: 'ghost'
//     }]
//   })
// })
</script>

<template>
  <UDashboardGroup unit="rem">
    <UDashboardSidebar
      id="default"
      v-model:open="open"
      collapsible
      resizable
      class="bg-elevated/25"
      :ui="{ footer: 'lg:border-t lg:border-default' }"
    >
      <template #header>
        <UButton
          to="/"
          variant="ghost"
          class="w-full justify-start px-4 py-3 data-[active=true]:bg-elevated"
        >
          <span class="truncate">BrainBook</span>
        </UButton>
      </template>

      <template #default="{ collapsed }">
        <UNavigationMenu
          :collapsed="collapsed"
          :items="links[0]"
          orientation="vertical"
          tooltip
          popover
        />

        <UNavigationMenu
          :collapsed="collapsed"
          :items="links[1]"
          orientation="vertical"
          tooltip
          class="mt-auto"
        />
      </template>

      <template #footer="{ collapsed }">
        <UserMenu :collapsed="collapsed" />
      </template>
    </UDashboardSidebar>

    <slot />

    <NotificationsSlideover />
  </UDashboardGroup>
</template>
