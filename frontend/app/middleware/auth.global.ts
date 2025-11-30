import { navigateTo } from '#app'

export default defineNuxtRouteMiddleware(async (to, _from) => {
  // Allow public pages
  const publicPages = ['/signin', '/signup']
  if (publicPages.includes(to.path)) return

  // Check authentication by calling backend
  try {
    // Use a reliable endpoint that returns 401 for unauthenticated users
    const publicConfig = useRuntimeConfig().public as { apiBase?: string }
    const apiBase = publicConfig.apiBase && typeof publicConfig.apiBase === 'string' && publicConfig.apiBase.length > 0
      ? publicConfig.apiBase
      : 'http://localhost:8080'
    // Use /protected/v1/user-list for authentication check
    await $fetch('/protected/v1/user-list', {
      method: 'GET',
      credentials: 'include',
      baseURL: apiBase
    })
    // If request succeeds, user is authenticated
  } catch (err: unknown) {
    // If error is 401 or 403, redirect to signin
    if (typeof err === 'object' && err !== null && 'status' in err) {
      const e = err as { status?: number }
      if (e.status === 401 || e.status === 403) {
        return navigateTo('/signin')
      }
    }
    // For other errors, allow navigation (or handle as needed)
  }
})
