<script setup lang="ts">
import { useApi } from '~/composables/useApi'
import { dateLabel, money, apPillClass, humanStatus } from '~/composables/useFormat'

definePageMeta({ layout: 'admin' })

const api = useApi()
const sessions = ref<any[]>([])
const masterclasses = ref<any[]>([])
const loading = ref(true)
const error = ref('')
const status = ref('')
const q = ref('')

const filtered = computed(() =>
  sessions.value.filter((s) => {
    const mc = masterclasses.value.find((m) => Number(m.id) === Number(s.masterclass_id))
    const matchQ = !q.value || `${mc?.title || ''} ${s.id}`.toLowerCase().includes(q.value.toLowerCase())
    return matchQ && (!status.value || s.status === status.value)
  })
)

onMounted(async () => {
  const [ss, mc] = await Promise.allSettled([
    api.get<any[]>('/admin/sessions'),
    api.get<any[]>('/admin/masterclasses')
  ])
  if (ss.status === 'fulfilled') sessions.value = Array.isArray(ss.value) ? ss.value : []
  else error.value = ss.reason?.message || 'Could not load sessions.'
  if (mc.status === 'fulfilled') masterclasses.value = Array.isArray(mc.value) ? mc.value : []
  loading.value = false
})
</script>

<template>
  <div class="acard">
    <div class="acard-h">
      <h3>Sessions <span class="muted">({{ filtered.length }})</span></h3>
      <div class="toolbar">
        <div class="afinput"><AppIcon name="search" :size="15" /><input v-model="q" type="search" placeholder="Search masterclass or ID…" /></div>
        <select v-model="status" class="aseg" style="border: 1px solid var(--a-line); background: var(--a-card); border-radius: 9px; height: 32px; padding: 0 8px"><option value="">All statuses</option><option value="scheduled">Scheduled</option><option value="active">Active</option><option value="completed">Completed</option><option value="cancelled">Cancelled</option></select>
      </div>
    </div>
    <AppAlert v-if="error" kind="error" class="mx-16">{{ error }}</AppAlert>
    <div class="atbl-wrap">
      <table class="atbl" v-if="filtered.length">
        <thead><tr><th>ID</th><th>Masterclass</th><th>Starts</th><th>Capacity</th><th>Price</th><th>Status</th></tr></thead>
        <tbody>
          <tr v-for="s in filtered" :key="s.id">
            <td>#{{ s.id }}</td>
            <td>{{ masterclasses.find((m) => Number(m.id) === Number(s.masterclass_id))?.title || s.masterclass?.title || `MC #${s.masterclass_id}` }}</td>
            <td>{{ dateLabel(s.starts_at) }}</td>
            <td><div class="abar" style="max-width: 120px"><i :style="{ width: `${Math.min(100, (Number(s.booked || 0) / Math.max(1, Number(s.capacity || 1))) * 100)}%` }" /></div><small>{{ s.booked || 0 }} / {{ s.capacity }}</small></td>
            <td>{{ money(s.price) }}</td>
            <td><span :class="apPillClass(s.status)"><i class="dot" /> {{ humanStatus(s.status) }}</span></td>
          </tr>
        </tbody>
      </table>
    </div>
    <LoadingState v-if="loading" :cards="1" label="Loading sessions…" />
    <EmptyState v-else-if="!filtered.length" title="No sessions found" message="Scheduled sessions from family hosts will show up here." icon="calendar" />
  </div>
</template>
