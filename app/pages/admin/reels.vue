<script setup lang="ts">
import { useApi } from '~/composables/useApi'
import { useToast } from '~/composables/useToast'
import { compactNumber, imageOr, apPillClass, humanStatus } from '~/composables/useFormat'

definePageMeta({ layout: 'admin' })

const api = useApi()
const toast = useToast()
const reels = ref<any[]>([])
const loading = ref(true)
const error = ref('')
const busyId = ref<any>(null)

async function load() {
  loading.value = true
  try {
    const data = await api.get<any[]>('/admin/reels')
    reels.value = Array.isArray(data) ? data : []
  } catch (e: any) {
    error.value = e?.message || 'Could not load reels.'
  } finally {
    loading.value = false
  }
}

async function moderate(reel: any, status: string) {
  if (!window.confirm(`Set reel "${reel.title || reel.id}" to ${status}?`)) return
  busyId.value = reel.id
  try {
    const updated = await api.patch<any>(`/admin/reels/${reel.id}`, { status })
    reels.value = reels.value.map((r) => (r.id === reel.id ? { ...r, ...updated } : r))
    toast.success(`Reel ${status}.`)
  } catch (e: any) {
    toast.error(e?.message || 'Reel moderation is not available in this API version.')
  } finally {
    busyId.value = null
  }
}

onMounted(load)
</script>

<template>
  <div class="acard">
    <div class="acard-h"><h3>Reels moderation <span class="muted">({{ reels.length }})</span></h3></div>
    <AppAlert v-if="error" kind="error" class="mx-16">{{ error }}</AppAlert>
    <div class="atbl-wrap">
      <table class="atbl" v-if="reels.length">
        <thead><tr><th>Reel</th><th>Creator</th><th>Views</th><th>Likes</th><th>Status</th><th>Actions</th></tr></thead>
        <tbody>
          <tr v-for="reel in reels" :key="reel.id">
            <td><div class="who"><img :src="imageOr(reel.image_url)" alt="" style="width: 34px; height: 44px; border-radius: 8px; object-fit: cover" /><div><b class="truncate" style="max-width: 220px">{{ reel.title || reel.caption || `Reel #${reel.id}` }}</b><small>{{ reel.restaurant?.name || reel.city || 'Uzbekistan' }}</small></div></div></td>
            <td>{{ reel.chef?.display_name || reel.restaurant?.name || '—' }}</td>
            <td>{{ compactNumber(reel.views) }}</td>
            <td>{{ compactNumber(reel.likes) }}</td>
            <td><span :class="apPillClass(reel.status)"><i class="dot" /> {{ humanStatus(reel.status) }}</span></td>
            <td>
              <button class="abtn primary" type="button" :disabled="busyId === reel.id" @click="moderate(reel, 'published')">Approve</button>
              <button class="abtn danger" type="button" :disabled="busyId === reel.id" @click="moderate(reel, 'rejected')">Reject</button>
              <button class="abtn" type="button" :disabled="busyId === reel.id" @click="moderate(reel, 'hidden')">Hide</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
    <LoadingState v-if="loading" :cards="1" label="Loading reels…" />
    <EmptyState v-else-if="!reels.length" title="No reels to moderate" message="Published culinary stories will appear here." icon="film" />
  </div>
</template>
