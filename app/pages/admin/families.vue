<script setup lang="ts">
import { useApi } from '~/composables/useApi'
import { useToast } from '~/composables/useToast'
import { apPillClass, humanStatus, initials } from '~/composables/useFormat'

definePageMeta({ layout: 'admin' })

const api = useApi()
const toast = useToast()
const families = ref<any[]>([])
const loading = ref(true)
const error = ref('')
const filter = ref('')
const busyId = ref<any>(null)

const filtered = computed(() => families.value.filter((c) => !filter.value || c.status === filter.value))
const pending = computed(() => families.value.filter((c) => c.status === 'pending').length)

async function load() {
  loading.value = true
  try {
    const data = await api.get<any[]>('/admin/families')
    families.value = Array.isArray(data) ? data : []
  } catch (e: any) {
    error.value = e?.message || 'Could not load families.'
  } finally {
    loading.value = false
  }
}

async function setStatus(family: any, status: string) {
  if (!window.confirm(`Set ${family.display_name || 'this family'} to ${status}?`)) return
  busyId.value = family.id
  try {
    const updated = await api.patch<any>(`/admin/families/${family.id}`, { status })
    families.value = families.value.map((c) => (c.id === family.id ? { ...c, ...updated } : c))
    toast.success(`Family ${status}.`)
  } catch (e: any) {
    toast.error(e?.message || 'Could not update family.')
  } finally {
    busyId.value = null
  }
}

onMounted(load)
</script>

<template>
  <div class="acard">
    <div class="acard-h">
      <h3>Families <span class="muted">({{ families.length }})</span></h3>
      <span v-if="pending" class="apill w"><i class="dot" /> {{ pending }} pending approval</span>
      <div class="aseg">
        <button v-for="f in [{ v: '', l: 'All' }, { v: 'pending', l: 'Pending' }, { v: 'approved', l: 'Approved' }, { v: 'suspended', l: 'Suspended' }]" :key="f.v" type="button" :class="{ on: filter === f.v }" @click="filter = f.v">{{ f.l }}</button>
      </div>
    </div>
    <AppAlert v-if="error" kind="error" class="mx-16">{{ error }}</AppAlert>
    <div class="atbl-wrap">
      <table class="atbl" v-if="filtered.length">
        <thead><tr><th>Family</th><th>City</th><th>Specialty</th><th>Rating</th><th>Status</th><th>Actions</th></tr></thead>
        <tbody>
          <tr v-for="family in filtered" :key="family.id">
            <td><div class="who"><span class="amono">{{ initials(family.display_name) }}</span><div><b>{{ family.display_name || 'Unnamed family' }}</b><small>{{ family.user?.email || family.email || `User #${family.user_id}` }}</small></div></div></td>
            <td>{{ family.city || '—' }}</td>
            <td>{{ family.specialty || '—' }}</td>
            <td>{{ family.rating ? Number(family.rating).toFixed(1) : '—' }}</td>
            <td><span :class="apPillClass(family.status)"><i class="dot" /> {{ humanStatus(family.status) }}</span></td>
            <td>
              <button v-if="family.status !== 'approved'" class="abtn primary" type="button" :disabled="busyId === family.id" @click="setStatus(family, 'approved')">Approve</button>
              <button v-if="family.status !== 'suspended'" class="abtn danger" type="button" :disabled="busyId === family.id" @click="setStatus(family, 'suspended')">Suspend</button>
              <button v-if="family.status === 'suspended'" class="abtn" type="button" :disabled="busyId === family.id" @click="setStatus(family, 'approved')">Restore</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
    <LoadingState v-if="loading" :cards="1" label="Loading families…" />
    <EmptyState v-else-if="!filtered.length" title="No families in this category" message="New family registrations appear here as pending." icon="family" />
  </div>
</template>
