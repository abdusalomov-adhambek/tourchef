<script setup lang="ts">
import { useApi } from '~/composables/useApi'
import { useToast } from '~/composables/useToast'
import { bookingTotal, dateLabel, money, apPillClass, humanStatus, initials } from '~/composables/useFormat'

definePageMeta({ layout: 'admin' })

const api = useApi()
const toast = useToast()
const bookings = ref<any[]>([])
const stats = useState<Record<string, any>>('tc-admin-stats', () => ({}))
const loading = ref(true)
const error = ref('')
const filter = ref('')

const payments = computed(() =>
  bookings.value.filter((b) => b.payment_status && (!filter.value || b.payment_status === filter.value))
)
const totalPaid = computed(() =>
  payments.value
    .filter((b) => ['paid', 'paid_dev'].includes(String(b.payment_status)))
    .reduce((sum, b) => sum + bookingTotal(b), 0)
)

async function load() {
  loading.value = true
  try {
    const [b, s] = await Promise.allSettled([
      api.get<any[]>('/admin/bookings'),
      api.get<Record<string, any>>('/admin/stats')
    ])
    if (b.status === 'fulfilled') bookings.value = Array.isArray(b.value) ? b.value : []
    if (s.status === 'fulfilled') stats.value = s.value
  } catch (e: any) {
    error.value = e?.message || 'Could not load payment records.'
  } finally {
    loading.value = false
  }
}

async function refund(b: any) {
  if (!window.confirm(`Record a refund for booking #${b.id}?`)) return
  try {
    await api.patch<any>(`/admin/bookings/${b.id}`, { payment_status: 'refunded', status: 'refunded' })
    bookings.value = bookings.value.map((x) => (x.id === b.id ? { ...x, payment_status: 'refunded', status: 'refunded' } : x))
    toast.success('Refund recorded.')
  } catch (e: any) {
    toast.error(e?.message || 'Could not record refund.')
  }
}

onMounted(load)
</script>

<template>
  <div class="agrid">
    <div class="agrid k4">
      <div class="acard akpi"><span class="ic tint-g"><AppIcon name="card" :size="17" /></span><span><small>Platform revenue</small><b>{{ money(stats.revenue) }}</b></span></div>
      <div class="acard akpi"><span class="ic tint-a"><AppIcon name="check" :size="17" /></span><span><small>Paid records</small><b>{{ money(totalPaid) }}</b></span></div>
      <div class="acard akpi"><span class="ic tint-w"><AppIcon name="refresh" :size="17" /></span><span><small>Payment records</small><b>{{ payments.length }}</b></span></div>
      <div class="acard akpi"><span class="ic tint-n"><AppIcon name="info" :size="17" /></span><span><small>Mode</small><b>Dev</b></span></div>
    </div>

    <div class="acard">
      <div class="acard-h">
        <h3>Payment records <span class="muted">({{ payments.length }})</span></h3>
        <div class="aseg">
          <button v-for="f in [{ v: '', l: 'All' }, { v: 'paid_dev', l: 'Paid (dev)' }, { v: 'pending', l: 'Pending' }, { v: 'refunded', l: 'Refunded' }]" :key="f.v" type="button" :class="{ on: filter === f.v }" @click="filter = f.v">{{ f.l }}</button>
        </div>
      </div>
      <AppAlert kind="info" class="mx-16 mb-12">Payments are recorded from bookings in clearly labelled development mode. No real charges are made.</AppAlert>
      <AppAlert v-if="error" kind="error" class="mx-16">{{ error }}</AppAlert>
      <div class="atbl-wrap">
        <table class="atbl" v-if="payments.length">
          <thead><tr><th>Booking</th><th>Customer</th><th>Amount</th><th>Method</th><th>Date</th><th>Status</th><th>Action</th></tr></thead>
          <tbody>
            <tr v-for="b in payments" :key="b.id">
              <td>#{{ b.id }}</td>
              <td><div class="who"><span class="amono">{{ initials(b.user?.full_name || 'CL') }}</span><div><b>{{ b.user?.full_name || `User #${b.user_id}` }}</b><small>{{ b.user?.email || '' }}</small></div></div></td>
              <td>{{ money(bookingTotal(b)) }}</td>
              <td>{{ b.payment_provider || 'dev' }}</td>
              <td>{{ dateLabel(b.created_at || b.booking_date) }}</td>
              <td><span :class="apPillClass(b.payment_status)"><i class="dot" /> {{ humanStatus(b.payment_status) }}</span></td>
              <td><button v-if="!['refunded', 'failed'].includes(String(b.payment_status))" class="abtn danger" type="button" @click="refund(b)">Refund</button><span v-else class="muted">—</span></td>
            </tr>
          </tbody>
        </table>
      </div>
      <LoadingState v-if="loading" :cards="1" label="Loading payments…" />
      <EmptyState v-else-if="!payments.length" title="No payment records" message="Payments appear here when travelers reserve experiences." icon="card" />
    </div>
  </div>
</template>
