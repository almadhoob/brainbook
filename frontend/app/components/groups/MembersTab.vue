<script setup lang="ts">
import type { GroupMember } from '~/composables/useGroupMembers'

interface Props {
  members: GroupMember[]
  membersLoading: boolean
  isOwner: boolean
  ownerRequestResponse: {
    requestId: string
    action: 'accept' | 'decline'
    loading: boolean
  }
}

interface Emits {
  (e: 'respond-request'): void
  (e: 'refresh-members'): void
}

defineProps<Props>()
const emit = defineEmits<Emits>()
</script>

<template>
  <div class="space-y-6">
    <section>
      <div class="flex items-center justify-between">
        <h3 class="text-lg font-semibold">
          Members
        </h3>
        <UButton
          v-if="isOwner"
          size="sm"
          color="primary"
          @click="emit('refresh-members')"
        >
          Refresh members
        </UButton>
      </div>

      <div v-if="membersLoading" class="mt-4 text-center text-sm text-muted">
        Loading members...
      </div>
      <div v-else-if="!members.length" class="mt-4 rounded-xl border border-default/60 p-4 text-sm text-muted">
        No members found.
      </div>
      <div v-else class="mt-4 grid gap-3 md:grid-cols-2">
        <div
          v-for="member in members"
          :key="member.id"
          class="flex items-center gap-3 rounded-xl border border-default/60 p-3"
        >
          <UAvatar :src="member.avatarSrc" :text="member.initials" />
          <div>
            <p class="font-medium">
              {{ member.fullName }}
            </p>
            <p class="text-xs text-muted capitalize">
              {{ member.role }}
            </p>
          </div>
        </div>
      </div>

      <div v-if="isOwner" class="mt-6 rounded-xl border border-default/60 p-4">
        <div class="flex items-center justify-between gap-3">
          <div>
            <p class="font-medium">
              Process join/invite request
            </p>
            <p class="text-xs text-muted">
              Enter request ID from notifications to accept or decline.
            </p>
          </div>
          <USelect
            v-model="ownerRequestResponse.action"
            :items="[
              { label: 'Accept', value: 'accept' },
              { label: 'Decline', value: 'decline' }
            ]"
            class="w-28"
          />
        </div>
        <div class="mt-3 flex gap-2">
          <UInput
            v-model="ownerRequestResponse.requestId"
            placeholder="Request ID"
            class="max-w-xs"
          />
          <UButton
            :loading="ownerRequestResponse.loading"
            @click="emit('respond-request')"
          >
            Update request
          </UButton>
        </div>
      </div>
    </section>
  </div>
</template>
