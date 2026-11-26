<script setup lang="ts">
import { useApi } from '~/composables/useApi'
import { dateLabel } from '~/composables/useFormat'

definePageMeta({ layout: 'cabinet' })

const api = useApi()
const items = ref<any[]>([])
const loading = ref(true)
const error = ref('')

onMounted(async () => {
  try {
    const data = await api.get<any[]>('/client/notifications')
    items.value = Array.isArray(data) ? data : []
  } catch (e: any) {
    error.value = e?.message || 'Could not load notifications.'
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div>
    <div class="cab-top">
      <div><h1>Notifications</h1><p>Booking confirmations, reminders and updates.</p></div>
    </div>
    <AppAlert v-if="error" kind="error" title="Could not load notifications">{{ error }}</AppAlert>
    <LoadingState v-if="loading" :cards="2" label="Loading notifications…" />
    <div v-else-if="items.length" class="panel">
      <div v-for="n in items" :key="n.id" class="panel-b" style="border-bottom: 1px solid var(--border)">
        <div class="row between">
          <div class="row" style="align-items: flex-start">
            <span class="ic" style="width: 38px; height: 38px; border-radius: 50%; display: grid; place-items: center; background: var(--accent-soft); color: var(--accent)">
              <AppIcon :name="n.type === 'booking' ? 'calendar' : 'bell'" :size="18" />
            </span>
            <div>
              <b>{{ n.title || 'TourChef update' }}</b>
              <p class="muted text-sm mt-8">{{ n.message || n.body || 'You have a new update.' }}</p>
            </div>
          </div>
          <small class="muted">{{ dateLabel(n.created_at || n.createdAt) }}</small>
        </div>
      </div>
    </div>
    <EmptyState v-else title="You are all caught up" message="There are no notifications right now. We will let you know when something changes." icon="bell" action-to="/client/bookings" action-label="View bookings" />
  </div>
</template>
