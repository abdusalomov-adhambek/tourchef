<script setup lang="ts">
import { useApi } from '~/composables/useApi'
import { useToast } from '~/composables/useToast'
import { dateLabel, money, pillClass, humanStatus } from '~/composables/useFormat'

definePageMeta({ layout: 'family' })

const api = useApi()
const toast = useToast()
const sessions = ref<any[]>([])
const masterclasses = ref<any[]>([])
const showForm = ref(false)
const loading = ref(true)
const busy = ref(false)
const error = ref('')

const form = reactive({ masterclass_id: '', starts_at: '', capacity: 8, price: 150000 })

async function load() {
  loading.value = true
  const [ss, mc] = await Promise.allSettled([
    api.get<any[]>('/family/sessions'),
    api.get<any[]>('/family/masterclasses')
  ])
  if (ss.status === 'fulfilled') sessions.value = Array.isArray(ss.value) ? ss.value : []
  if (mc.status === 'fulfilled') {
    masterclasses.value = Array.isArray(mc.value) ? mc.value : []
    if (!form.masterclass_id && masterclasses.value.length) form.masterclass_id = String(masterclasses.value[0].id)
  }
  loading.value = false
}

async function create() {
  error.value = ''
  if (!form.masterclass_id || !form.starts_at || Number(form.capacity) < 1 || Number(form.price) < 1) {
    error.value = 'Choose a masterclass, date, capacity and price.'
    toast.error(error.value)
    return
  }
  busy.value = true
  try {
    const created = await api.post<any>('/family/sessions', {
      masterclass_id: form.masterclass_id,
      starts_at: new Date(form.starts_at).toISOString(),
      capacity: Number(form.capacity),
      price: Number(form.price)
    })
    sessions.value = [created, ...sessions.value]
    toast.success('Session scheduled. Travelers can now book it.')
    showForm.value = false
    form.starts_at = ''
  } catch (e: any) {
    error.value = e?.message || 'Could not create session.'
    toast.error(error.value)
  } finally {
    busy.value = false
  }
}

onMounted(load)
</script>

<template>
  <div>
    <div class="cab-top">
      <div><h1>Sessions</h1><p>Schedule dates and seats for your masterclasses.</p></div>
      <button class="btn btn-primary btn-sm" type="button" @click="showForm = !showForm">
        <AppIcon :name="showForm ? 'close' : 'plus'" :size="16" /> {{ showForm ? 'Close form' : 'New session' }}
      </button>
    </div>

    <div v-if="showForm" class="panel mb-24">
      <div class="panel-h"><h3>Schedule a session</h3></div>
      <div class="panel-b">
        <AppAlert v-if="error" kind="error" class="mb-16">{{ error }}</AppAlert>
        <form @submit.prevent="create">
          <div class="form-grid">
            <div class="field"><label for="mc">Masterclass</label><select id="mc" v-model="form.masterclass_id" class="select" required><option value="" disabled>Select a masterclass</option><option v-for="mc in masterclasses" :key="mc.id" :value="String(mc.id)">{{ mc.title }}</option></select><span class="hint">Create a masterclass first if the list is empty.</span></div>
            <div class="field"><label for="starts">Date and time</label><input id="starts" v-model="form.starts_at" class="inp" type="datetime-local" required /></div>
            <div class="field"><label for="capacity">Capacity</label><input id="capacity" v-model.number="form.capacity" class="inp" type="number" min="1" max="100" required /></div>
            <div class="field"><label for="price">Price per person (so'm)</label><input id="price" v-model.number="form.price" class="inp" type="number" min="1" required /><span class="hint">Shown on the site in USD (1 USD = 12 000 so'm).</span></div>
          </div>
          <div class="form-actions">
            <button class="btn btn-outline" type="button" @click="showForm = false">Cancel</button>
            <button class="btn btn-primary" type="submit" :disabled="busy || !masterclasses.length" :class="{ 'is-loading': busy }">{{ busy ? 'Scheduling…' : 'Schedule session' }}</button>
          </div>
        </form>
      </div>
    </div>

    <div class="panel">
      <div class="panel-h"><h3>All sessions</h3><span class="muted text-sm">{{ sessions.length }} total</span></div>
      <div class="tbl-wrap">
        <table class="tbl" v-if="sessions.length">
          <thead><tr><th>Session</th><th>Masterclass</th><th>Date</th><th>Booked</th><th>Price</th><th>Status</th></tr></thead>
          <tbody>
            <tr v-for="s in sessions" :key="s.id">
              <td>#{{ s.id }}</td>
              <td>{{ masterclasses.find((m) => Number(m.id) === Number(s.masterclass_id))?.title || (s.masterclass?.title || `MC #${s.masterclass_id}`) }}</td>
              <td>{{ dateLabel(s.starts_at) }}</td>
              <td>{{ s.booked || 0 }} / {{ s.capacity }}</td>
              <td>{{ money(s.price) }}</td>
              <td><span :class="pillClass(s.status)">{{ humanStatus(s.status) }}</span></td>
            </tr>
          </tbody>
        </table>
      </div>
      <LoadingState v-if="loading" :cards="1" label="Loading sessions…" />
      <EmptyState v-else-if="!sessions.length" title="No sessions scheduled" message="Add a date and capacity to start taking bookings." icon="calendar" />
    </div>
  </div>
</template>
