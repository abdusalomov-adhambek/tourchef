<script setup lang="ts">
import { useApi } from '~/composables/useApi'
import { useToast } from '~/composables/useToast'
import { dateLabel, bookingStart, bookingMc, bookingTotal, money, apPillClass, humanStatus } from '~/composables/useFormat'

definePageMeta({ layout: 'admin' })

const api = useApi()
const toast = useToast()
const bookings = ref<any[]>([])
const loading = ref(true)
const error = ref('')
const status = ref('')
const q = ref('')
const busyId = ref<any>(null)

const filtered = computed(() =>
  bookings.value.filter((b) => {
    const matchQ = !q.value || `${b.id} ${bookingMc(b)?.title || ''} ${b.user?.email || ''}`.toLowerCase().includes(q.value.toLowerCase())
    return matchQ && (!status.value || b.status === status.value)
  })
)

async function load() {
  loading.value = true
  try {
    const data = await api.get<any[]>('/admin/bookings')
    bookings.value = Array.isArray(data) ? data : []
  } catch (e: any) {
    error.value = e?.message || 'Could not load bookings.'
  } finally {
    loading.value = false
  }
}

async function change(b: any, next: string) {
  if (!window.confirm(`Change booking #${b.id} status to ${next}?`)) return
  busyId.value = b.id
  try {
    const updated = await api.patch<any>(`/admin/bookings/${b.id}`, { status: next })
    bookings.value = bookings.value.map((x) => (x.id === b.id ? { ...x, ...updated } : x))
    toast.success(`Booking #${b.id} updated.`)
  } catch (e: any) {
    toast.error(e?.message || 'Could not update booking.')
  } finally {
    busyId.value = null
  }
}

onMounted(load)
</script>

<template>
  <div class="acard">
    <div class="acard-h">
      <h3>Bookings <span class="muted">({{ filtered.length }})</span></h3>
      <div class="toolbar">
        <div class="afinput"><AppIcon name="search" :size="15" /><input v-model="q" type="search" placeholder="Search ID, title, email…" /></div>
        <select v-model="status" class="aseg" style="border: 1px solid var(--a-line); background: var(--a-card); border-radius: 9px; height: 32px; padding: 0 8px"><option value="">All statuses</option><option value="pending">Pending</option><option value="confirmed">Confirmed</option><option value="completed">Completed</option><option value="cancelled">Cancelled</option><option value="refunded">Refunded</option></select>
      </div>
    </div>
    <AppAlert v-if="error" kind="error" class="mx-16">{{ error }}</AppAlert>
    <div class="atbl-wrap">
      <table class="atbl" v-if="filtered.length">
        <thead><tr><th>Booking</th><th>Guest</th><th>Experience</th><th>Date</th><th>Total</th><th>Payment</th><th>Status</th><th>Update</th></tr></thead>
        <tbody>
          <tr v-for="b in filtered" :key="b.id">
            <td>#{{ b.id }}</td>
            <td>{{ b.user?.full_name || b.user?.email || `User #${b.user_id}` }}</td>
            <td class="truncate" style="max-width: 200px">{{ bookingMc(b)?.title || 'Masterclass' }}</td>
            <td>{{ dateLabel(bookingStart(b)) }}</td>
            <td>{{ money(bookingTotal(b)) }}</td>
            <td><span :class="apPillClass(b.payment_status)"><i class="dot" /> {{ humanStatus(b.payment_status) }}</span></td>
            <td><span :class="apPillClass(b.status)"><i class="dot" /> {{ humanStatus(b.status) }}</span></td>
            <td>
              <select class="aseg" style="border: 1px solid var(--a-line); background: var(--a-card); border-radius: 8px; height: 30px; padding: 0 6px" :disabled="busyId === b.id" :value="b.status" @change="change(b, ($event.target as HTMLSelectElement).value)">
                <option value="pending">Pending</option>
                <option value="confirmed">Confirmed</option>
                <option value="completed">Completed</option>
                <option value="cancelled">Cancelled</option>
                <option value="refunded">Refunded</option>
              </select>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
    <LoadingState v-if="loading" :cards="1" label="Loading bookings…" />
    <EmptyState v-else-if="!filtered.length" title="No bookings found" message="Reservations from travelers will appear here." icon="calendar" />
  </div>
</template>
