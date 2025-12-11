import { createSharedComposable } from '@vueuse/core'

interface SessionState {
  user_id: number | null
  full_name: string
  email: string
  avatar: string | null
}

const _useSession = () => {
  const runtimeConfig = useRuntimeConfig()
  const apiBase = runtimeConfig.public?.apiBase || 'http://localhost:8080'

  const session = useState<SessionState>('session-profile', () => ({
    user_id: null,
    full_name: '',
    email: '',
    avatar: null
  }))
  const loading = ref(false)
  const error = ref<string | null>(null)

  const hydrate = async (force = false) => {
    if (!import.meta.client) return
    if (loading.value) return
    if (session.value.user_id && !force) return

    loading.value = true
    try {
      const data = await $fetch<SessionState>(`${apiBase}/protected/v1/session`, {
        credentials: 'include'
      })
      session.value = {
        user_id: data.user_id ?? null,
        full_name: data.full_name ?? '',
        email: data.email ?? '',
        avatar: data.avatar ?? null
      }
      error.value = null
    } catch (err) {
      error.value = 'Unable to load session information'
      console.error(err)
    } finally {
      loading.value = false
    }
  }

  if (import.meta.client) {
    hydrate()
  }

  return {
    session: readonly(session),
    loading: readonly(loading),
    error: readonly(error),
    hydrate
  }
}

export const useSession = createSharedComposable(_useSession)
