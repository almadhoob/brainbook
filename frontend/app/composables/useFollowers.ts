import { extractErrorMessage } from './useGroupHelpers'

export interface ApiFollowerSummary {
  user_id: number
  f_name?: string | null
  l_name?: string | null
  avatar?: string | null
}

export function useFollowers(apiBase: string, userId: Ref<number | null>) {
  const followers = ref<ApiFollowerSummary[]>([])
  const loading = ref(false)
  const error = ref('')
  const loaded = ref(false)

  async function loadFollowers(force = false) {
    if (!userId.value) return
    if (loading.value) return
    if (loaded.value && !force) return

    loading.value = true
    error.value = ''

    try {
      const data = await $fetch<{ followers?: ApiFollowerSummary[] }>(
        `${apiBase}/protected/v1/user/${userId.value}/followers`,
        { credentials: 'include' }
      )
      followers.value = data.followers ?? []
      loaded.value = true
    } catch (err) {
      error.value = extractErrorMessage(err) || 'Unable to load followers.'
    } finally {
      loading.value = false
    }
  }

  watch(userId, (value) => {
    if (!value) {
      followers.value = []
      loaded.value = false
      error.value = ''
    }
  })

  return {
    followers: readonly(followers),
    loading: readonly(loading),
    error: readonly(error),
    loaded: readonly(loaded),
    loadFollowers
  }
}
