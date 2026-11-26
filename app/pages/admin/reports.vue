<script setup lang="ts">
import { useApi } from '~/composables/useApi'
import { useToast } from '~/composables/useToast'
import { bookingTotal, money, dateLabel } from '~/composables/useFormat'

definePageMeta({ layout: 'admin' })

const api = useApi()
const toast = useToast()
const stats = useState<Record<string, any>>('tc-admin-stats', () => ({}))
const dishes = ref<any[]>([])
const bookings = ref<any[]>([])
const loading = ref(true)
const exporting = ref(false)

function escapeCsv(value: unknown) {
  return `"${String(value ?? '').replace(/"/g, '""')}"`
}

function downloadCsv() {
  if (!bookings.value.length) {
    toast.info('There are no bookings to export yet.')
    return
  }
  exporting.value = true
  const header = ['Booking ID', 'User ID', 'Masterclass', 'Session date', 'Guests', 'Total', 'Status', 'Payment']
  const rows = bookings.value.map((b) => [
    b.id,
    b.user_id,
    b.masterclass?.title || '',
    dateLabel(b.starts_at || b.session?.starts_at),
    b.guests || 1,
    bookingTotal(b),
    b.status,
    b.payment_status
  ])
  const csv = [header, ...rows].map((row) => row.map(escapeCsv).join(',')).join('\n')
  const url = URL.createObjectURL(new Blob([csv], { type: 'text/csv;charset=utf-8' }))
  const link = document.createElement('a')
  link.href = url
  link.download = `tourchef-bookings-${new Date().toISOString().slice(0, 10)}.csv`
  document.body.appendChild(link)
  link.click()
  link.remove()
  URL.revokeObjectURL(url)
  exporting.value = false
  toast.success('Bookings report exported as CSV.')
}

onMounted(async () => {
  const [s, d, b] = await Promise.allSettled([
    api.get<Record<string, any>>('/admin/stats'),
    api.get<any>('/dishes', { limit: 200 }),
    api.get<any[]>('/admin/bookings')
  ])
  if (s.status === 'fulfilled') stats.value = s.value
  if (d.status === 'fulfilled') dishes.value = Array.isArray(d.value?.items) ? d.value.items : []
  if (b.status === 'fulfilled') bookings.value = Array.isArray(b.value) ? b.value : []
  loading.value = false
})
</script>

<template>
  <div class="agrid">
    <div class="acard">
      <div class="acard-h">
        <h3>Reports & analytics</h3>
        <button class="abtn primary" type="button" :disabled="exporting" @click="downloadCsv"><AppIcon name="download" :size="15" /> Export bookings CSV</button>
      </div>
      <div class="acard-b">
        <LoadingState v-if="loading" :cards="1" label="Building reports…" />
        <div v-else class="agrid k4">
          <div class="abox"><h4>GMV / revenue</h4><b style="font-size: 20px">{{ money(stats.revenue) }}</b></div>
          <div class="abox"><h4>Bookings</h4><b style="font-size: 20px">{{ stats.bookings ?? 0 }}</b></div>
          <div class="abox"><h4>Masterclasses</h4><b style="font-size: 20px">{{ stats.masterclasses ?? 0 }}</b></div>
          <div class="abox"><h4>Users</h4><b style="font-size: 20px">{{ stats.users ?? 0 }}</b></div>
        </div>
      </div>
    </div>

    <div class="agrid r2">
      <div class="acard">
        <div class="acard-h"><h3>Popular dish regions</h3></div>
        <div class="acard-b">
          <div v-for="region in Array.from(new Set(dishes.map((d) => d.region).filter(Boolean)))" :key="region" class="mb-12">
            <div class="asum"><span>{{ region }}</span><span>{{ dishes.filter((d) => d.region === region).length }} dishes</span></div>
            <div class="abar mt-8"><i :style="{ width: `${(dishes.filter((d) => d.region === region).length / Math.max(1, dishes.length)) * 100}%` }" /></div>
          </div>
          <p v-if="!dishes.length" class="muted text-sm">No dish data yet.</p>
        </div>
      </div>
      <div class="acard">
        <div class="acard-h"><h3>Booking status breakdown</h3></div>
        <div class="acard-b">
          <div v-for="status in ['confirmed', 'pending', 'completed', 'cancelled']" :key="status" class="mb-12">
            <div class="asum"><span>{{ status }}</span><span>{{ bookings.filter((b) => b.status === status).length }}</span></div>
            <div class="abar mt-8"><i :style="{ width: `${(bookings.filter((b) => b.status === status).length / Math.max(1, bookings.length)) * 100}%` }" /></div>
          </div>
          <p v-if="!bookings.length" class="muted text-sm">No bookings yet.</p>
        </div>
      </div>
    </div>
  </div>
</template>
