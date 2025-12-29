<script setup lang="ts">
import type { GroupEventItem } from '~/composables/useGroupEvents'

interface Props {
  events: GroupEventItem[]
  eventsLoading: boolean
  createEventLoading: boolean
  rsvpLoading: Record<number, boolean>
  newEventForm: { title: string, description: string, time: string }
  isOwner: boolean
}

interface Emits {
  (e: 'create-event'): void
  (e: 'rsvp', eventId: number, response: 'going' | 'not_going'): void
}

defineProps<Props>()
const emit = defineEmits<Emits>()
</script>

<template>
  <div class="space-y-6">
    <section>
      <div class="flex items-center justify-between">
        <h3 class="text-lg font-semibold">
          Events
        </h3>
        <UTooltip v-if="!isOwner" text="Only group owners can schedule events.">
          <span class="text-xs text-muted">Owner only</span>
        </UTooltip>
      </div>

      <form v-if="isOwner" class="mt-4 space-y-3" @submit.prevent="emit('create-event')">
        <UFieldGroup label="Day & time" class="w-full">
          <UInput v-model="newEventForm.time" type="datetime-local" class="w-xsm" />
        </UFieldGroup>
        <UFieldGroup label="Title" class="w-full">
          <UInput v-model="newEventForm.title" placeholder="Weekly sync" class="w-md" />
        </UFieldGroup>
        <UFieldGroup label="Description" class="w-md">
          <UTextarea v-model="newEventForm.description" placeholder="Add context or an agenda" class="w-full" />
        </UFieldGroup>
        <div class="flex justify-end">
          <UButton type="submit" :loading="createEventLoading">
            Schedule event
          </UButton>
        </div>
      </form>

      <div v-if="eventsLoading" class="mt-4 text-center text-sm text-muted">
        Loading events...
      </div>
      <div v-else-if="!events.length" class="mt-4 rounded-xl border border-default/60 p-4 text-sm text-muted">
        No events planned yet.
      </div>
      <div v-else class="mt-4 space-y-3">
        <UCard v-for="event in events" :key="event.id">
          <h4 class="font-medium">
            {{ event.title }}
          </h4>
          <p class="text-sm text-muted">
            {{ event.formattedTime }}
          </p>
          <p v-if="event.description" class="mt-2 text-sm">
            {{ event.description }}
          </p>

          <div class="mt-3 flex items-center gap-3 text-sm text-muted">
            <span class="flex items-center gap-1">
              <UIcon name="i-lucide-check" class="size-4" /> {{ event.goingCount }} going
            </span>
            <span class="flex items-center gap-1">
              <UIcon name="i-lucide-x" class="size-4" /> {{ event.notGoingCount }} not going
            </span>
          </div>

          <template #footer>
            <div class="flex flex-wrap gap-2">
              <UButton
                size="sm"
                color="primary"
                :loading="rsvpLoading[event.id]"
                @click="emit('rsvp', event.id, 'going')"
              >
                I'm going
              </UButton>
              <UButton
                size="sm"
                color="neutral"
                variant="soft"
                :loading="rsvpLoading[event.id]"
                @click="emit('rsvp', event.id, 'not_going')"
              >
                Not going
              </UButton>
            </div>
          </template>
        </UCard>
      </div>
    </section>
  </div>
</template>
