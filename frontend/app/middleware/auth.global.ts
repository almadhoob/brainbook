import { navigateTo } from '#app'

export default defineNuxtRouteMiddleware(async (to, _from) => {
  // Allow public pages
  const publicPages = ['/signin', '/signup']
  const isPublicPage = publicPages.includes(to.path)
  if (isPublicPage) {
    if (import.meta.client) {
      const { session, hydrate } = useSession()
      await hydrate()
      if (session.value.user_id) {
        return navigateTo('/')
      }
    }
    return
  }
  // Check authentication by calling backend
  try {
    // Use a reliable endpoint that returns 401 for unauthenticated users
    const publicConfig = useRuntimeConfig().public as { apiBase?: string }
    const apiBase = publicConfig.apiBase && typeof publicConfig.apiBase === 'string' && publicConfig.apiBase.length > 0
      ? publicConfig.apiBase
      : 'http://localhost:8080'
    const headers = import.meta.server ? useRequestHeaders(['cookie']) : undefined
    // Use /protected/v1/session for authentication check
    await $fetch('/protected/v1/session', {
      method: 'GET',
      credentials: 'include',
      baseURL: apiBase,
      headers
    })
    // If request succeeds, user is authenticated
  } catch (err: unknown) {
    // If error is 401 or 403, redirect to signin
    if (typeof err === 'object' && err !== null && 'status' in err) {
      const e = err as { status?: number }
      if (e.status === 401 || e.status === 403) {
        if (!publicPages.includes(to.path)) {
          return navigateTo('/signin')
        }
      }
    }
    // For other errors, allow navigation (or handle as needed)
  }
})
