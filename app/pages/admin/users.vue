<script setup lang="ts">
import { useApi } from '~/composables/useApi'
import { useToast } from '~/composables/useToast'
import { dateLabel, apPillClass, humanStatus, initials } from '~/composables/useFormat'

definePageMeta({ layout: 'admin' })

const api = useApi()
const toast = useToast()
const users = ref<any[]>([])
const loading = ref(true)
const error = ref('')
const q = ref('')
const role = ref('')
const status = ref('')
const busyId = ref<any>(null)

const filtered = computed(() =>
  users.value.filter((u) => {
    const matchQ = !q.value || `${u.full_name} ${u.email} ${u.phone || ''}`.toLowerCase().includes(q.value.toLowerCase())
    return matchQ && (!role.value || u.role === role.value) && (!status.value || u.status === status.value)
  })
)

async function load() {
  loading.value = true
  error.value = ''
  try {
    const data = await api.get<any[]>('/admin/users')
    users.value = Array.isArray(data) ? data : []
  } catch (e: any) {
    error.value = e?.message || 'Could not load users.'
  } finally {
    loading.value = false
  }
}

async function setStatus(user: any, next: string) {
  if (!window.confirm(`Change ${user.full_name}'s status to ${next}?`)) return
  busyId.value = user.id
  try {
    const updated = await api.patch<any>(`/admin/users/${user.id}`, { status: next })
    users.value = users.value.map((u) => (u.id === user.id ? { ...u, ...updated } : u))
    toast.success(`User ${next}.`)
  } catch (e: any) {
    toast.error(e?.message || 'Could not update user.')
  } finally {
    busyId.value = null
  }
}

onMounted(load)
</script>

<template>
  <div class="acard">
    <div class="acard-h">
      <h3>Users <span class="muted">({{ filtered.length }})</span></h3>
      <div class="toolbar">
        <div class="afinput"><AppIcon name="search" :size="15" /><input v-model="q" type="search" placeholder="Search name, email…" /></div>
        <select v-model="role" class="aseg" style="border: 1px solid var(--a-line); background: var(--a-card); border-radius: 9px; height: 32px; padding: 0 8px"><option value="">All roles</option><option value="client">Client</option><option value="chef">Family host</option><option value="admin">Admin</option></select>
        <select v-model="status" class="aseg" style="border: 1px solid var(--a-line); background: var(--a-card); border-radius: 9px; height: 32px; padding: 0 8px"><option value="">All statuses</option><option value="active">Active</option><option value="pending">Pending</option><option value="suspended">Suspended</option></select>
      </div>
    </div>
    <AppAlert v-if="error" kind="error" class="mx-16">{{ error }}</AppAlert>
    <div class="atbl-wrap">
      <table class="atbl" v-if="filtered.length">
        <thead><tr><th>User</th><th>Role</th><th>Status</th><th>Phone</th><th>Joined</th><th>Actions</th></tr></thead>
        <tbody>
          <tr v-for="u in filtered" :key="u.id">
            <td><div class="who"><span class="amono">{{ initials(u.full_name) }}</span><div><b>{{ u.full_name }}</b><small>{{ u.email }}</small></div></div></td>
            <td>{{ u.role }}</td>
            <td><span :class="apPillClass(u.status)"><i class="dot" /> {{ humanStatus(u.status) }}</span></td>
            <td>{{ u.phone || '—' }}</td>
            <td>{{ u.created_at ? dateLabel(u.created_at, false) : '—' }}</td>
            <td>
              <button v-if="u.status !== 'suspended'" class="abtn danger" type="button" :disabled="busyId === u.id" @click="setStatus(u, 'suspended')">Suspend</button>
              <button v-else class="abtn" type="button" :disabled="busyId === u.id" @click="setStatus(u, 'active')">Reactivate</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
    <LoadingState v-if="loading" :cards="1" label="Loading users…" />
    <EmptyState v-else-if="!filtered.length" title="No users found" message="Try different filters or check back later." icon="users" />
  </div>
</template>
