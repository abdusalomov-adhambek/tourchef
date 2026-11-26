<script setup lang="ts">
import { useApi } from '~/composables/useApi'
import { money } from '~/composables/useFormat'

definePageMeta({ layout: 'admin' })

const api = useApi()
const stats = useState<Record<string, any>>('tc-admin-stats', () => ({}))
const loading = ref(true)
const error = ref('')

const cards = computed(() => [
  { label: 'Users', value: stats.value.users, icon: 'users', tint: 'tint-a', to: '/admin/users' },
  { label: 'Families', value: stats.value.chefs, icon: 'family', tint: 'tint-g', to: '/admin/families' },
  { label: 'Restaurants', value: stats.value.restaurants, icon: 'restaurant', tint: 'tint-n', to: '/admin/locations' },
  { label: 'Dishes', value: stats.value.dishes, icon: 'dish', tint: 'tint-w', to: '/admin/dishes' },
  { label: 'Masterclasses', value: stats.value.masterclasses, icon: 'book', tint: 'tint-a', to: '/admin/masterclasses' },
  { label: 'Bookings', value: stats.value.bookings, icon: 'calendar', tint: 'tint-g', to: '/admin/bookings' },
  { label: 'Platform revenue', value: money(stats.value.revenue), icon: 'card', tint: 'tint-w', to: '/admin/payments' }
])

onMounted(async () => {
  try {
    stats.value = await api.get<Record<string, any>>('/admin/stats')
  } catch (e: any) {
    error.value = e?.message || 'Could not load platform statistics.'
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div class="agrid">
    <AppAlert v-if="error" kind="error" title="Stats unavailable">{{ error }}</AppAlert>

    <div v-if="loading" class="agrid k4">
      <div v-for="i in 7" :key="i" class="acard skeleton" style="height: 70px" />
    </div>
    <div v-else class="agrid k4">
      <NuxtLink v-for="c in cards" :key="c.label" :to="c.to" class="acard akpi">
        <span class="ic" :class="c.tint"><AppIcon :name="c.icon" :size="17" /></span>
        <span><small>{{ c.label }}</small><b>{{ c.value ?? '—' }}</b></span>
      </NuxtLink>
    </div>

    <div class="agrid r2">
      <div class="acard">
        <div class="acard-h"><h3>Platform overview</h3></div>
        <div class="acard-b">
          <div class="asum"><span>Total users</span><b>{{ stats.users ?? '—' }}</b></div>
          <div class="asum"><span>Verified families</span><b>{{ stats.chefs ?? '—' }}</b></div>
          <div class="asum"><span>Active dishes</span><b>{{ stats.dishes ?? '—' }}</b></div>
          <div class="asum"><span>Masterclasses</span><b>{{ stats.masterclasses ?? '—' }}</b></div>
          <div class="asum"><span>Total bookings</span><b>{{ stats.bookings ?? '—' }}</b></div>
          <div class="asum"><span>Revenue</span><b>{{ money(stats.revenue) }}</b></div>
        </div>
      </div>
      <div class="acard">
        <div class="acard-h"><h3>Needs attention</h3></div>
        <div class="acard-b">
          <NuxtLink to="/admin/families" class="abtn" style="justify-content: space-between">
            Pending family applications <span class="cnt" style="background: var(--a-bad-soft); color: var(--a-bad); border-radius: 999px; padding: 1px 8px">{{ stats.pending_chefs ?? 0 }}</span>
          </NuxtLink>
          <NuxtLink to="/admin/masterclasses" class="abtn" style="justify-content: space-between">
            Pending masterclasses <span class="cnt" style="background: var(--a-warn-soft); color: var(--a-warn); border-radius: 999px; padding: 1px 8px">{{ stats.pending_mc ?? 0 }}</span>
          </NuxtLink>
          <NuxtLink to="/admin/reels" class="abtn" style="justify-content: space-between">Reels moderation <AppIcon name="chevron" :size="15" /></NuxtLink>
          <NuxtLink to="/admin/reports" class="abtn" style="justify-content: space-between">Reports & analytics <AppIcon name="chevron" :size="15" /></NuxtLink>
        </div>
      </div>
    </div>
  </div>
</template>
