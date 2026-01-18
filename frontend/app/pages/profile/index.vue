<script setup lang="ts">
const router = useRouter()
const { session, hydrate } = useSession()
const isRedirecting = ref(false)

const redirectToProfile = async () => {
  if (isRedirecting.value) return
  isRedirecting.value = true
  await hydrate(true)
  if (session.value.user_id) {
    await router.replace(`/profile/${session.value.user_id}`)
  } else {
    await router.replace('/signin')
  }
}

if (import.meta.client) {
  redirectToProfile()
}
</script>

<template>
  <UDashboardPanel id="profile-redirect">
    <template #header>
      <UDashboardNavbar title="Profile">
        <template #leading>
          <UDashboardSidebarCollapse />
        </template>
      </UDashboardNavbar>
    </template>

    <template #body>
      <div class="py-10 text-center text-muted">
        Loading your profile...
      </div>
    </template>
  </UDashboardPanel>
</template>
