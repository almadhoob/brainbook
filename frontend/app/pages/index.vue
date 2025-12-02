<script setup lang="ts">
import { computed } from 'vue'
import type { DropdownMenuItem } from '@nuxt/ui'

const { isNotificationsSlideoverOpen } = useDashboard()

const items = [[{
  label: 'New post',
  icon: 'i-lucide-file-text',
  to: '/posts'
}, {
  label: 'New group',
  icon: 'i-lucide-users',
  to: '/groups'
}]] satisfies DropdownMenuItem[][]

const stats = [{
  label: 'Daily active users',
  value: '3,482',
  delta: '+12%',
  icon: 'i-lucide-users',
  trend: 'up'
}, {
  label: 'New posts',
  value: '426',
  delta: '+6%',
  icon: 'i-lucide-file-text',
  trend: 'up'
}, {
  label: 'Support tickets',
  value: '37',
  delta: '-3%',
  icon: 'i-lucide-life-buoy',
  trend: 'down'
}, {
  label: 'Pending invites',
  value: '58',
  delta: '+18%',
  icon: 'i-lucide-users-round',
  trend: 'up'
}] as const

const heroBackground = '/background.jpg'

const heroBackgroundStyle = computed(() => ({
  backgroundImage: `linear-gradient(140deg, rgba(15, 23, 42, 0.92), rgba(6, 182, 212, 0.75)), url(${heroBackground})`
}))
</script>

<template>
  <UDashboardPanel id="home">
    <template #header>
      <UDashboardNavbar title="Home" :ui="{ right: 'gap-3' }">
        <template #leading>
          <UDashboardSidebarCollapse />
        </template>

        <template #right>
          <UTooltip text="Notifications" :shortcuts="['N']">
            <UButton
              color="neutral"
              variant="ghost"
              square
              @click="isNotificationsSlideoverOpen = true"
            >
              <UChip color="error" inset>
                <UIcon name="i-lucide-bell" class="size-5 shrink-0" />
              </UChip>
            </UButton>
          </UTooltip>

          <UDropdownMenu :items="items">
            <UButton icon="i-lucide-plus" size="md" class="rounded-full" />
          </UDropdownMenu>
        </template>
      </UDashboardNavbar>
    </template>

    <template #body>
      <div class="p-6 space-y-8">
        <section
          class="home-hero relative overflow-hidden rounded-3xl text-white shadow-lg"
          :style="heroBackgroundStyle"
        >
          <div class="relative z-10 flex flex-col gap-6 p-8 lg:flex-row lg:items-center lg:justify-between">
            <div class="space-y-4 max-w-2xl">
              <p class="text-sm font-medium uppercase tracking-[0.3em] text-white/70">
                Dashboard overview
              </p>
              <div>
                <h1 class="text-3xl font-semibold lg:text-4xl">
                  Welcome back to BrainBook
                </h1>
                <p class="mt-3 text-base text-white/80">
                  Here's a quick pulse on what your members are doing right now.
                </p>
              </div>
              <div class="flex flex-wrap gap-3 text-sm text-white/80">
                <span class="inline-flex items-center gap-2 rounded-full border border-white/20 px-4 py-2">
                  <UIcon name="i-lucide-activity" class="size-4" /> Live engagement up 18%
                </span>
                <span class="inline-flex items-center gap-2 rounded-full border border-white/20 px-4 py-2">
                  <UIcon name="i-lucide-shield-check" class="size-4" /> Moderation queue clear
                </span>
              </div>
            </div>

            <div class="grid w-full max-w-sm grid-cols-2 gap-4 text-center">
              <div class="rounded-2xl bg-white/15 p-4 backdrop-blur">
                <p class="text-sm text-white/80">
                  Active conversations
                </p>
                <p class="text-3xl font-semibold">
                  128
                </p>
                <p class="text-xs text-emerald-200">
                  +9% vs last week
                </p>
              </div>
              <div class="rounded-2xl bg-white/15 p-4 backdrop-blur">
                <p class="text-sm text-white/80">
                  Avg. response time
                </p>
                <p class="text-3xl font-semibold">
                  2m 16s
                </p>
                <p class="text-xs text-emerald-200">
                  -35% support wait
                </p>
              </div>
            </div>
          </div>
        </section>

        <section>
          <div class="mb-4 flex flex-wrap items-center justify-between gap-3">
            <div>
              <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
                Today's statistics
              </h2>
              <p class="text-sm text-gray-500 dark:text-gray-400">
                A snapshot of the most important community health metrics.
              </p>
            </div>
          </div>

          <div class="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
            <UCard
              v-for="stat in stats"
              :key="stat.label"
              :ui="{ body: ['space-y-3'] }"
              class="h-full"
            >
              <div class="flex items-center justify-between gap-4">
                <div class="space-y-1">
                  <p class="text-sm font-medium text-gray-500 dark:text-gray-400">
                    {{ stat.label }}
                  </p>
                  <p class="text-2xl font-semibold text-gray-900 dark:text-white">
                    {{ stat.value }}
                  </p>
                </div>
                <span class="inline-flex rounded-full bg-primary-50 p-3 text-primary-500 dark:bg-primary-500/10">
                  <UIcon :name="stat.icon" class="size-5" />
                </span>
              </div>
              <p
                :class="[
                  'text-sm font-medium',
                  stat.trend === 'down'
                    ? 'text-error-500 dark:text-error-400'
                    : 'text-success-500 dark:text-success-400'
                ]"
              >
                {{ stat.delta }} vs last 24h
              </p>
            </UCard>
          </div>
        </section>
      </div>
    </template>
  </UDashboardPanel>
</template>

<style scoped>
.home-hero {
  background-size: cover;
  background-position: center;
}
</style>
