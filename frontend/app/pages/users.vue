<script setup lang="ts">
import type { TableColumn } from '@nuxt/ui'
import { getPaginationRowModel } from '@tanstack/table-core'
import upperFirst from 'lodash/upperFirst'

const UAvatar = resolveComponent('UAvatar')
const UButton = resolveComponent('UButton')
const toast = useToast()
const table = ref()

const columnFilters = ref([{ id: 'fullName', value: '' }])
const columnVisibility = ref()
const rowSelection = ref({})

const statusFilter = ref('all')

const config = useRuntimeConfig()
const apiBase = config.public?.apiBase || 'http://localhost:8080'

interface ApiUser {
  user_id?: number
  user_full_name?: string
  user_avatar?: string | null
  last_message_time?: string | null
  follows?: boolean
  followed_by?: boolean
  follow_request_status?: string | null
}

interface ApiFollowRequest {
  id?: number
  requester_id?: number
  f_name?: string | null
  l_name?: string | null
  avatar?: string | null
  created_at?: string | null
}

interface TableUser {
  id: number | null
  fullName: string
  initials: string
  avatarSrc?: string
  lastMessageTime?: string | null
  follows: boolean
  followedBy: boolean
  followRequestStatus: string | null
}

const { data, status, error, refresh } = await useFetch<{ users: ApiUser[] }>(`${apiBase}/protected/v1/user-list`, {
  credentials: 'include',
  lazy: true,
  server: false
})

const { data: followRequestsData, refresh: refreshFollowRequests } = await useFetch<{ requests: ApiFollowRequest[] }>(
  `${apiBase}/protected/v1/follow-requests`,
  { credentials: 'include', lazy: true, server: false }
)

const normalizedUsers = computed<TableUser[]>(() => {
  const users = data.value?.users
  if (!Array.isArray(users)) return []

  return users.map((user) => {
    const fullName = user.user_full_name?.trim() || 'Unknown user'
    const initials = fullName
      .split(/\s+/)
      .filter(Boolean)
      .slice(0, 2)
      .map(part => part[0]?.toUpperCase())
      .join('') || 'U'

    const avatarSrc = user.user_avatar ? normalizeAvatar(user.user_avatar) : undefined

    return {
      id: typeof user.user_id === 'number' ? user.user_id : null,
      fullName,
      initials,
      avatarSrc,
      lastMessageTime: user.last_message_time ?? null,
      follows: Boolean(user.follows),
      followedBy: Boolean(user.followed_by),
      followRequestStatus: typeof user.follow_request_status === 'string' ? user.follow_request_status : null
    }
  })
})

const filteredUsers = computed<TableUser[]>(() => {
  if (statusFilter.value === 'following') {
    return normalizedUsers.value.filter(user => user.follows)
  }
  if (statusFilter.value === 'followers') {
    return normalizedUsers.value.filter(user => user.followedBy)
  }
  return normalizedUsers.value
})

function normalizeAvatar(raw?: string | null) {
  if (!raw) return undefined
  if (raw.startsWith('data:')) return raw
  return `data:image/png;base64,${raw}`
}

function formatLastActive(timestamp?: string | null) {
  if (!timestamp) return 'No recent activity'
  const normalized = timestamp.includes('T') ? timestamp : timestamp.replace(' ', 'T')
  const date = new Date(normalized)
  if (Number.isNaN(date.getTime())) {
    return timestamp
  }
  return date.toLocaleString()
}

async function followUser(userId: number) {
  try {
    await $fetch(`${apiBase}/protected/v1/users/${userId}/follow`, {
      method: 'POST',
      credentials: 'include'
    })
    toast.add({ title: 'Follow request sent', description: 'Pending if the profile is private.' })
    refresh()
  } catch (e) {
    toast.add({ title: 'Error', description: 'Could not follow user.' })
    console.error(e)
  }
}

async function unfollowUser(userId: number) {
  try {
    await $fetch(`${apiBase}/protected/v1/users/${userId}/unfollow`, {
      method: 'POST',
      credentials: 'include'
    })
    toast.add({ title: 'Unfollowed', description: 'You are no longer following this user.' })
    refresh()
  } catch (e) {
    toast.add({ title: 'Error', description: 'Could not unfollow user.' })
    console.error(e)
  }
}

async function respondFollowRequest(requestId: number, action: 'accept' | 'decline') {
  try {
    await $fetch(`${apiBase}/protected/v1/follow-requests/${requestId}`, {
      method: 'POST',
      credentials: 'include',
      body: { action }
    })
    toast.add({
      title: action === 'accept' ? 'Request accepted' : 'Request declined',
      description: 'Follow request updated.'
    })
    await refreshFollowRequests()
  } catch (e) {
    toast.add({ title: 'Error', description: 'Could not update follow request.' })
    console.error(e)
  }
}

const columns: TableColumn<TableUser>[] = [
  {
    accessorKey: 'fullName',
    header: 'User',
    cell: ({ row }) =>
      h('div', { class: 'flex items-center gap-3' }, [
        h(UAvatar, {
          size: 'lg',
          src: row.original.avatarSrc,
          text: row.original.initials
        }),
        h('div', { class: 'flex flex-col' }, [
          h('span', { class: 'font-medium' }, row.original.fullName),
          h('span', { class: 'text-xs text-muted' }, formatLastActive(row.original.lastMessageTime))
        ])
      ])
  },
  {
    accessorKey: 'lastMessageTime',
    header: 'Last Activity',
    cell: ({ row }) =>
      h('span', { class: 'text-sm text-muted' }, formatLastActive(row.original.lastMessageTime))
  },
  {
    id: 'relationship',
    header: 'Status',
    cell: ({ row }) => {
      if (row.original.follows) {
        return h('span', { class: 'text-xs text-success' }, 'Following')
      }
      if (row.original.followRequestStatus === 'pending') {
        return h('span', { class: 'text-xs text-warning' }, 'Request pending')
      }
      if (row.original.followedBy) {
        return h('span', { class: 'text-xs text-muted' }, 'Follows you')
      }
      return h('span', { class: 'text-xs text-muted' }, 'Not connected')
    }
  },
  {
    id: 'actions',
    header: '',
    cell: ({ row }) =>
      row.original.id == null
        ? h('span', { class: 'text-xs text-muted' }, 'Follow unavailable')
        : h('div', { class: 'flex gap-2' }, [
            h(UButton, {
              label: row.original.followRequestStatus === 'pending'
                ? 'Requested'
                : (row.original.follows ? 'Unfollow' : 'Follow'),
              color: row.original.follows ? 'neutral' : 'primary',
              variant: row.original.follows ? 'soft' : 'solid',
              size: 'sm',
              disabled: row.original.followRequestStatus === 'pending',
              onClick: () => (row.original.follows
                ? unfollowUser(row.original.id as number)
                : followUser(row.original.id as number))
            }),
            h(UButton, {
              label: 'View profile',
              color: 'neutral',
              variant: 'outline',
              size: 'sm',
              to: `/profile/${row.original.id}`
            })
          ])
  }
]

const pagination = ref({ pageIndex: 0, pageSize: 10 })

// Comment out watcher to prevent recursion
// watch(() => statusFilter.value, (newVal) => {
//   if (!table?.value?.tableApi) return
//   const statusColumn = table.value.tableApi.getColumn('status')
//   if (!statusColumn) return
//   if (newVal === 'all') {
//     statusColumn.setFilterValue(undefined)
//   } else {
//     statusColumn.setFilterValue(newVal)
//   }
// })
</script>

<template>
  <UDashboardPanel id="users">
    <template #header>
      <UDashboardNavbar title="Users">
        <template #leading>
          <UDashboardSidebarCollapse />
        </template>
      </UDashboardNavbar>
    </template>

    <template #body>
      <div class="flex flex-wrap items-center justify-between gap-1.5">
        <UInput
          :model-value="(table?.tableApi?.getColumn('fullName')?.getFilterValue() as string)"
          class="max-w-sm"
          icon="i-lucide-search"
          placeholder="Filter users..."
          @update:model-value="table?.tableApi?.getColumn('fullName')?.setFilterValue($event)"
        />

        <div class="flex flex-wrap items-center gap-1.5">
          <USelect
            v-model="statusFilter"
            :items="[
              { label: 'All', value: 'all' },
              { label: 'Following', value: 'following' },
              { label: 'Followers', value: 'followers' }
            ]"
            :ui="{ trailingIcon: 'group-data-[state=open]:rotate-180 transition-transform duration-200' }"
            placeholder="Filter status"
            class="min-w-28"
          />
          <UDropdownMenu
            :items="
              table?.tableApi
                ?.getAllColumns()
                .filter((column: any) => column.getCanHide())
                .map((column: any) => ({
                  label: upperFirst(column.id),
                  type: 'checkbox' as const,
                  checked: column.getIsVisible(),
                  onUpdateChecked(checked: boolean) {
                    table?.tableApi?.getColumn(column.id)?.toggleVisibility(!!checked)
                  },
                  onSelect(e?: Event) {
                    e?.preventDefault()
                  }
                }))
            "
            :content="{ align: 'end' }"
          >
            <UButton
              label="Display"
              color="neutral"
              variant="outline"
              trailing-icon="i-lucide-settings-2"
            />
          </UDropdownMenu>
        </div>
      </div>

      <div v-if="status === 'pending'" class="py-8 text-center text-muted">
        Loading users...
      </div>
      <div v-else-if="error" class="py-8 text-center text-error">
        Error loading users.
      </div>
      <div v-if="filteredUsers.length === 0 && status === 'success' && !error" class="py-8 text-center text-muted">
        No users found.
      </div>
      <UTable
        v-if="filteredUsers.length > 0"
        ref="table"
        v-model:column-filters="columnFilters"
        v-model:column-visibility="columnVisibility"
        v-model:row-selection="rowSelection"
        v-model:pagination="pagination"
        :pagination-options="{
          getPaginationRowModel: getPaginationRowModel()
        }"
        class="shrink-0"
        :data="filteredUsers"
        :columns="columns"
        :loading="status === 'pending'"
        :ui="{
          base: 'table-fixed border-separate border-spacing-0',
          thead: '[&>tr]:bg-elevated/50 [&>tr]:after:content-none',
          tbody: '[&>tr]:last:[&>td]:border-b-0',
          th: 'py-2 first:rounded-l-lg last:rounded-r-lg border-y border-default first:border-l last:border-r',
          td: 'border-b border-default',
          separator: 'h-0'
        }"
      />

      <section class="mt-8 space-y-4">
        <div class="flex items-center justify-between">
          <h3 class="text-lg font-semibold">
            Pending follow requests
          </h3>
          <UButton
            color="neutral"
            variant="ghost"
            icon="i-lucide-refresh-cw"
            @click="refreshFollowRequests()"
          >
            Refresh
          </UButton>
        </div>

        <div v-if="!followRequestsData?.requests?.length" class="rounded-xl border border-default/60 p-4 text-sm text-muted">
          No pending requests right now.
        </div>

        <div v-else class="grid gap-3 md:grid-cols-2 lg:grid-cols-3">
          <UCard v-for="req in followRequestsData.requests" :key="req.id">
            <div class="flex items-center gap-3">
              <UAvatar :src="req.avatar ? `data:image/png;base64,${req.avatar}` : undefined" :text="(req.f_name || '?')[0]" />
              <div>
                <p class="font-medium">
                  {{ `${req.f_name ?? ''} ${req.l_name ?? ''}`.trim() || 'Unknown user' }}
                </p>
                <p class="text-xs text-muted">
                  Requested {{ formatLastActive(req.created_at) }}
                </p>
              </div>
            </div>
            <template #footer>
              <div class="flex gap-2">
                <UButton size="sm" color="primary" @click="respondFollowRequest(req.id as number, 'accept')">
                  Accept
                </UButton>
                <UButton
                  size="sm"
                  color="neutral"
                  variant="soft"
                  @click="respondFollowRequest(req.id as number, 'decline')"
                >
                  Decline
                </UButton>
              </div>
            </template>
          </UCard>
        </div>
      </section>

      <div class="flex items-center justify-between gap-3 border-t border-default pt-4 mt-auto">
        <div class="text-sm text-muted">
          {{ table?.tableApi?.getFilteredSelectedRowModel().rows.length || 0 }} of
          {{ table?.tableApi?.getFilteredRowModel().rows.length || 0 }} row(s) selected.
        </div>

        <div class="flex items-center gap-1.5">
          <UPagination
            :default-page="(table?.tableApi?.getState().pagination.pageIndex || 0) + 1"
            :items-per-page="table?.tableApi?.getState().pagination.pageSize"
            :total="table?.tableApi?.getFilteredRowModel().rows.length"
            @update:page="(p: number) => table?.tableApi?.setPageIndex(p - 1)"
          />
        </div>
      </div>
    </template>
  </UDashboardPanel>
</template>
