<script setup lang="ts">
import { useApi } from '~/composables/useApi'
import { useToast } from '~/composables/useToast'
import { dateLabel, bookingStart, bookingMc, money, pillClass, humanStatus } from '~/composables/useFormat'

definePageMeta({ layout: 'cabinet' })

const api = useApi()
const toast = useToast()
const bookings = ref<any[]>([])
const loading = ref(true)
const error = ref('')
const tab = ref<'all' | 'upcoming' | 'completed' | 'cancelled'>('all')
const busyId = ref<any>(null)

const tabs = [
  { value: 'all', label: 'All' },
  { value: 'upcoming', label: 'Upcoming' },
  { value: 'completed', label: 'Completed' },
  { value: 'cancelled', label: 'Cancelled' }
]

const filtered = computed(() => {
  const list = bookings.value
  const status = (b: any) => String(b.status || '').toLowerCase()
  if (tab.value === 'upcoming') return list.filter((b) => ['confirmed', 'pending', 'pending_payment'].includes(status(b)))
  if (tab.value === 'completed') return list.filter((b) => ['completed', 'attended'].includes(status(b)))
  if (tab.value === 'cancelled') return list.filter((b) => ['cancelled', 'refunded'].includes(status(b)))
  return list
})

async function load() {
  loading.value = true
  error.value = ''
  try {
    const data = await api.get<any[]>('/client/bookings')
    bookings.value = Array.isArray(data) ? data : []
  } catch (e: any) {
    error.value = e?.message || 'Could not load your bookings.'
  } finally {
    loading.value = false
  }
}

async function cancel(b: any) {
  if (!window.confirm(`Cancel booking #${b.id}? This cannot be undone.`)) return
  busyId.value = b.id
  try {
    const updated = await api.patch<any>(`/client/bookings/${b.id}/cancel`, {})
    bookings.value = bookings.value.map((x) => (x.id === b.id ? { ...x, ...updated } : x))
    toast.success('Booking cancelled.')
  } catch (e: any) {
    toast.error(e?.message || 'Could not cancel this booking.')
  } finally {
    busyId.value = null
  }
}

onMounted(load)
</script>

<template>
  <div>
    <div class="cab-top">
      <div><h1>My bookings</h1><p>Every culinary experience you have reserved.</p></div>
      <NuxtLink to="/masterclasses" class="btn btn-outline btn-sm">Browse experiences</NuxtLink>
    </div>

    <AppAlert v-if="error" kind="error" title="Could not load bookings">{{ error }}</AppAlert>

    <div class="tabs">
      <button v-for="t in tabs" :key="t.value" type="button" :class="{ on: tab === t.value }" @click="tab = t.value as typeof tab">
        {{ t.label }} <span class="muted">({{ t.value === 'all' ? bookings.length : filtered.length }})</span>
      </button>
    </div>

    <LoadingState v-if="loading" :cards="2" label="Loading bookings…" />
    <div v-else-if="filtered.length" class="stack">
      <article v-for="b in filtered" :key="b.id" class="list-card">
        <div class="lc-media"><img :src="bookingMc(b)?.image_url || '/logo.svg'" :alt="bookingMc(b)?.title || 'Booking'" /></div>
        <div>
          <span class="badge neutral">#{{ b.id }}</span>
          <span :class="pillClass(b.status)" style="margin-left: 8px">{{ humanStatus(b.status) }}</span>
          <h3>{{ bookingMc(b)?.title || `Booking #${b.id}` }}</h3>
          <p class="muted text-sm" style="margin: 0">{{ bookingMc(b)?.restaurant?.name || 'TourChef partner kitchen' }}</p>
          <div class="lc-meta">
            <span><AppIcon name="calendar" :size="15" /> {{ dateLabel(bookingStart(b)) }}</span>
            <span><AppIcon name="users" :size="15" /> {{ b.guests || 1 }} guest(s)</span>
            <span><AppIcon name="card" :size="15" /> {{ humanStatus(b.payment_status) }}</span>
          </div>
          <div class="row" style="gap: 8px">
            <NuxtLink :to="`/client/bookings/${b.id}`" class="btn btn-outline btn-sm">View details</NuxtLink>
            <button
              v-if="['confirmed', 'pending', 'pending_payment'].includes(String(b.status || '').toLowerCase())"
              class="btn btn-ghost btn-sm" type="button" :disabled="busyId === b.id"
              @click="cancel(b)"
            >Cancel booking</button>
            <NuxtLink v-if="['completed', 'attended'].includes(String(b.status || '').toLowerCase())" to="/client/reviews" class="btn btn-ghost btn-sm">Leave a review</NuxtLink>
          </div>
        </div>
        <div class="lc-side">
          <div><div class="price">{{ money(b.total_price || b.total || b.amount || b.price || bookingMc(b)?.price) }}</div><div class="muted text-xs">total paid</div></div>
          <NuxtLink v-if="bookingMc(b)?.restaurant?.slug" :to="`/restaurants/${bookingMc(b).restaurant.slug}`" class="btn btn-outline btn-sm">View restaurant</NuxtLink>
        </div>
      </article>
    </div>
    <EmptyState v-else title="No bookings in this category" message="Explore our experiences and reserve your next culinary moment." icon="calendar" action-to="/masterclasses" action-label="Find experiences" />
  </div>
</template>
