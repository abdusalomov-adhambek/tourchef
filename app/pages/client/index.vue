<script setup lang="ts">
import { useApi } from '~/composables/useApi'
import { useAuth } from '~/composables/useAuth'
import { useFavorites } from '~/composables/useFavorites'
import { dateLabel, bookingStart, bookingMc, money } from '~/composables/useFormat'

definePageMeta({ layout: 'cabinet' })

const api = useApi()
const auth = useAuth()
const favorites = useFavorites()
const loading = ref(true)
const bookings = ref<any[]>([])
const notifications = ref<any[]>([])
const recommended = ref<any[]>([])

const upcoming = computed(() =>
  bookings.value.filter((b) => ['confirmed', 'pending', 'pending_payment'].includes(String(b.status || '').toLowerCase()))
)
const completed = computed(() =>
  bookings.value.filter((b) => ['completed', 'attended'].includes(String(b.status || '').toLowerCase()))
)

onMounted(async () => {
  const [b, n, mc] = await Promise.allSettled([
    api.get<any[]>('/client/bookings'),
    api.get<any[]>('/client/notifications'),
    api.get<any>('/masterclasses', { limit: 4, sort: 'rec' })
  ])
  if (b.status === 'fulfilled') bookings.value = Array.isArray(b.value) ? b.value : []
  if (n.status === 'fulfilled') notifications.value = Array.isArray(n.value) ? n.value : []
  if (mc.status === 'fulfilled') recommended.value = Array.isArray(mc.value?.items) ? mc.value.items : []
  await favorites.load()
  loading.value = false
})
</script>

<template>
  <div>
    <div class="cab-top">
      <div>
        <h1>Welcome back, {{ auth.user.value?.full_name?.split(' ')[0] || 'traveler' }}</h1>
        <p>Here is what is happening with your culinary journeys.</p>
      </div>
      <NuxtLink to="/masterclasses" class="btn btn-primary btn-sm">Find an experience <AppIcon name="arrow" :size="16" /></NuxtLink>
    </div>

    <div v-if="loading" class="stack">
      <div class="kpis"><div v-for="i in 4" :key="i" class="skeleton" style="height: 92px" /></div>
      <div class="panel"><div class="panel-b"><div class="skeleton" style="height: 180px" /></div></div>
    </div>
    <template v-else>
      <div class="kpis">
        <div class="kpi"><small>Upcoming bookings</small><b>{{ upcoming.length }}</b><span class="delta">{{ bookings.length }} total</span></div>
        <div class="kpi"><small>Completed experiences</small><b>{{ completed.length }}</b><span class="delta">memories made</span></div>
        <div class="kpi"><small>Favorite masterclasses</small><b>{{ favorites.ids.value.length }}</b><span class="delta">saved for later</span></div>
        <div class="kpi"><small>Notifications</small><b>{{ notifications.length }}</b><span class="delta">stay in the loop</span></div>
      </div>

      <div class="panel mb-24">
        <div class="panel-h">
          <h3>Upcoming experiences</h3>
          <NuxtLink to="/client/bookings" class="text-link">All bookings <AppIcon name="arrow" :size="16" /></NuxtLink>
        </div>
        <div v-if="upcoming.length" class="panel-b">
          <div v-for="b in upcoming.slice(0, 4)" :key="b.id" class="mini-card mb-12">
            <img :src="bookingMc(b)?.image_url || '/logo.svg'" alt="" />
            <div class="mc-body">
              <b>{{ bookingMc(b)?.title || `Booking #${b.id}` }}</b>
              <small>{{ dateLabel(bookingStart(b)) }} · {{ b.guests || 1 }} guest(s) · {{ money(b.total_price || b.total || b.amount || b.price || bookingMc(b)?.price) }}</small>
            </div>
            <NuxtLink :to="`/client/bookings/${b.id}`" class="btn btn-outline btn-sm">Details</NuxtLink>
          </div>
        </div>
        <EmptyState v-else title="No upcoming experiences" message="Your next culinary adventure is one click away." icon="calendar" action-to="/masterclasses" action-label="Explore masterclasses" />
      </div>

      <div class="two-col">
        <div class="panel">
          <div class="panel-h"><h3>Recommended for you</h3><NuxtLink to="/masterclasses" class="text-link">See all</NuxtLink></div>
          <div v-if="recommended.length" class="panel-b">
            <div v-for="mc in recommended.slice(0, 3)" :key="mc.id" class="mini-card mb-12">
              <img :src="mc.image_url || mc.dish?.image_url || '/logo.svg'" alt="" />
              <div class="mc-body">
                <b>{{ mc.title }}</b>
                <small>{{ mc.city }} · {{ money(mc.price) }} / person</small>
              </div>
              <NuxtLink :to="`/masterclasses/${mc.slug || mc.id}`" class="btn btn-outline btn-sm">View</NuxtLink>
            </div>
          </div>
          <EmptyState v-else title="No recommendations yet" message="Browse the catalog to discover experiences." icon="spark" />
        </div>
        <div class="panel">
          <div class="panel-h"><h3>Recent notifications</h3><NuxtLink to="/client/notifications" class="text-link">All</NuxtLink></div>
          <div v-if="notifications.length" class="panel-b">
            <div v-for="n in notifications.slice(0, 4)" :key="n.id" class="info-row">
              <AppIcon name="bell" :size="18" />
              <span class="truncate" style="max-width: 220px">{{ n.message || n.title || 'Notification' }}</span>
            </div>
          </div>
          <EmptyState v-else title="You are all caught up" message="Booking updates will appear here." icon="bell" />
        </div>
      </div>
    </template>
  </div>
</template>
