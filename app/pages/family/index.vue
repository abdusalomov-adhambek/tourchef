<script setup lang="ts">
import { useApi } from '~/composables/useApi'
import { useAuth } from '~/composables/useAuth'
import { dateLabel, money, pillClass, humanStatus } from '~/composables/useFormat'

definePageMeta({ layout: 'family' })

const api = useApi()
const auth = useAuth()
const loading = ref(true)
const masterclasses = ref<any[]>([])
const sessions = ref<any[]>([])
const dishes = ref<any[]>([])

const upcoming = computed(() =>
  sessions.value.filter((s) => !s.starts_at || new Date(s.starts_at).getTime() > Date.now())
)

onMounted(async () => {
  const [mc, ss, d] = await Promise.allSettled([
    api.get<any[]>('/family/masterclasses'),
    api.get<any[]>('/family/sessions'),
    api.get<any>('/dishes', { limit: 100 })
  ])
  if (mc.status === 'fulfilled') masterclasses.value = Array.isArray(mc.value) ? mc.value : []
  if (ss.status === 'fulfilled') sessions.value = Array.isArray(ss.value) ? ss.value : []
  if (d.status === 'fulfilled') dishes.value = Array.isArray(d.value?.items) ? d.value.items : []
  loading.value = false
})
</script>

<template>
  <div>
    <div class="cab-top">
      <div>
        <h1>Family dashboard</h1>
        <p>Share your family recipes: choose a dish, set how many guests you can welcome, and pick the date and time.</p>
      </div>
      <NuxtLink to="/family/masterclasses" class="btn btn-primary btn-sm">New masterclass <AppIcon name="plus" :size="16" /></NuxtLink>
    </div>

    <div v-if="loading" class="stack">
      <div class="kpis"><div v-for="i in 4" :key="i" class="skeleton" style="height: 92px" /></div>
    </div>
    <template v-else>
      <div class="kpis">
        <div class="kpi"><small>My masterclasses</small><b>{{ masterclasses.length }}</b><span class="delta">published & drafts</span></div>
        <div class="kpi"><small>Guest sessions</small><b>{{ sessions.length }}</b><span class="delta">scheduled</span></div>
        <div class="kpi"><small>When guests arrive</small><b>{{ upcoming.length }}</b><span class="delta">upcoming dates</span></div>
        <div class="kpi"><small>Dishes we can cook</small><b>{{ dishes.length }}</b><span class="delta">family recipes</span></div>
      </div>

      <div class="two-col">
        <div class="panel">
          <div class="panel-h"><h3>My masterclasses</h3><NuxtLink to="/family/masterclasses" class="text-link">Manage</NuxtLink></div>
          <div v-if="masterclasses.length" class="panel-b">
            <div v-for="mc in masterclasses.slice(0, 5)" :key="mc.id" class="mini-card mb-12">
              <img :src="mc.image_url || mc.dish?.image_url || '/logo.svg'" alt="" />
              <div class="mc-body"><b>{{ mc.title }}</b><small>{{ mc.city }} · {{ money(mc.price) }}</small></div>
              <span :class="pillClass(mc.status)">{{ humanStatus(mc.status) }}</span>
            </div>
          </div>
          <EmptyState v-else title="No masterclasses yet" message="Create your first experience and start hosting travelers." icon="book" action-to="/family/masterclasses" action-label="Create masterclass" />
        </div>
        <div class="panel">
          <div class="panel-h"><h3>Next sessions</h3><NuxtLink to="/family/sessions" class="text-link">Manage</NuxtLink></div>
          <div v-if="sessions.length" class="panel-b">
            <div v-for="s in sessions.slice(0, 5)" :key="s.id" class="info-row">
              <AppIcon name="calendar" :size="18" />
              <span class="truncate" style="max-width: 190px">{{ dateLabel(s.starts_at) }}</span>
              <span>{{ s.booked || 0 }}/{{ s.capacity || 0 }}</span>
            </div>
          </div>
          <EmptyState v-else title="No sessions scheduled" message="Add a date to one of your masterclasses." icon="calendar" action-to="/family/sessions" action-label="Add session" />
        </div>
      </div>
    </template>
  </div>
</template>
