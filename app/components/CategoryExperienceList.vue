<script setup lang="ts">
import { useApi } from '~/composables/useApi'
import { money, duration, imageOr } from '~/composables/useFormat'

const props = defineProps<{
  title: string
  description: string
  category: string
  keywords: string[]
  icon?: string
}>()

const api = useApi()
const items = ref<any[]>([])
const loading = ref(true)
const error = ref('')
const chip = ref<'all' | 'available' | 'top'>('all')

const visible = computed(() => {
  let list = items.value
  if (chip.value === 'top') list = list.filter((m) => Number(m.rating || 0) >= 4.7)
  if (chip.value === 'available') list = list.filter((m) => Number(m.available_seats ?? m.seats ?? 1) > 0)
  return list
})

function matches(mc: any): boolean {
  const haystack = [
    mc.title,
    mc.description,
    mc.category,
    mc.difficulty,
    mc.dish?.name,
    mc.dish_name,
    mc.chef?.display_name,
    mc.restaurant?.name
  ]
    .filter(Boolean)
    .join(' ')
    .toLowerCase()
  return props.keywords.some((word) => haystack.includes(word.toLowerCase()))
}

onMounted(async () => {
  let list: any[] = []
  try {
    const result = await api.get<any>('/masterclasses', { q: props.category, limit: 100 })
    list = Array.isArray(result?.items) ? result.items : []
  } catch { /* fall back to full catalog */ }
  if (!list.length) {
    try {
      const all = await api.get<any>('/masterclasses', { limit: 100 })
      list = Array.isArray(all?.items) ? all.items.filter(matches) : []
    } catch (e: any) {
      error.value = e?.message || 'Could not load experiences.'
    }
  }
  items.value = list
  loading.value = false
})
</script>

<template>
  <div>
    <div class="cab-top">
      <div><h1>{{ title }}</h1><p>{{ description }}</p></div>
      <NuxtLink to="/masterclasses" class="btn btn-outline btn-sm">Browse all experiences</NuxtLink>
    </div>

    <div class="panel mb-24">
      <div class="panel-b row">
        <span style="width: 44px; height: 44px; border-radius: 13px; display: grid; place-items: center; flex: none; background: var(--accent-soft); color: var(--accent)">
          <AppIcon :name="icon || 'family'" :size="22" />
        </span>
        <div class="grow"><b>{{ title }}</b><p class="muted text-sm mt-8">{{ description }}</p></div>
        <span class="badge" :class="category === 'family' ? 'sage' : ''">{{ items.length }} experiences</span>
      </div>
    </div>

    <div class="chips mb-16">
      <button class="chip" :class="{ on: chip === 'all' }" type="button" @click="chip = 'all'">All experiences</button>
      <button class="chip" :class="{ on: chip === 'available' }" type="button" @click="chip = 'available'">Available now</button>
      <button class="chip" :class="{ on: chip === 'top' }" type="button" @click="chip = 'top'">Top rated</button>
    </div>

    <AppAlert v-if="error" kind="error" title="Could not load experiences">{{ error }}</AppAlert>
    <LoadingState v-else-if="loading" :cards="3" label="Loading experiences…" />
    <div v-else-if="visible.length" class="stack">
      <article v-for="mc in visible" :key="mc.id" class="list-card">
        <div class="lc-media"><NuxtLink :to="`/masterclasses/${mc.slug || mc.id}`"><img :src="imageOr(mc.image_url || mc.dish?.image_url)" :alt="mc.title" /></NuxtLink></div>
        <div>
          <span class="badge sage">{{ title }}</span>
          <h3><NuxtLink :to="`/masterclasses/${mc.slug || mc.id}`">{{ mc.title }}</NuxtLink></h3>
          <div class="muted text-sm">{{ mc.restaurant?.name || 'TourChef kitchen' }}<span v-if="mc.chef?.display_name"> · {{ mc.chef.display_name }}</span></div>
          <div class="lc-meta">
            <span><AppIcon name="dish" :size="15" /> {{ mc.dish?.name || mc.dish_name || 'Uzbek cuisine' }}</span>
            <span><AppIcon name="clock" :size="15" /> {{ duration(mc.duration_min) }}</span>
            <span v-if="mc.rating" class="rating"><AppIcon name="star" :size="15" /> {{ Number(mc.rating).toFixed(1) }}</span>
          </div>
          <NuxtLink :to="`/masterclasses/${mc.slug || mc.id}`" class="btn btn-outline btn-sm">View details</NuxtLink>
        </div>
        <div class="lc-side">
          <div><div class="price">{{ money(mc.price) }}</div><div class="muted text-xs mt-8">per person</div></div>
          <NuxtLink :to="`/masterclasses/${mc.slug || mc.id}#booking`" class="btn btn-primary btn-sm">Book now</NuxtLink>
        </div>
      </article>
    </div>
    <EmptyState v-else title="No experiences in this category yet" message="Try another filter or browse the full masterclass catalog." icon="family" action-to="/masterclasses" action-label="Browse all experiences" />
  </div>
</template>
