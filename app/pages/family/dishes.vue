<script setup lang="ts">
import { useApi } from '~/composables/useApi'
import { imageOr } from '~/composables/useFormat'

definePageMeta({ layout: 'family' })

const api = useApi()
const dishes = ref<any[]>([])
const loading = ref(true)
const error = ref('')
const q = ref('')
const category = ref('')

const categories = computed(() => Array.from(new Set(dishes.value.map((d) => d.category).filter(Boolean))).sort())
const filtered = computed(() =>
  dishes.value.filter((d) => {
    const matchQ = !q.value || `${d.name} ${d.description || ''} ${d.region || ''}`.toLowerCase().includes(q.value.toLowerCase())
    const matchCategory = !category.value || d.category === category.value
    return matchQ && matchCategory
  })
)

onMounted(async () => {
  try {
    const data = await api.get<any>('/dishes', { limit: 100 })
    dishes.value = Array.isArray(data?.items) ? data.items.filter((d) => !d.status || d.status === 'active') : []
  } catch (e: any) {
    error.value = e?.message || 'Could not load the dish catalog.'
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div>
    <div class="cab-top">
      <div><h1>Dish catalog</h1><p>Select a dish when creating a masterclass. These are managed by the administrator.</p></div>
      <NuxtLink to="/family/masterclasses" class="btn btn-outline btn-sm">Create masterclass</NuxtLink>
    </div>

    <div class="panel mb-24">
      <div class="panel-b row">
        <input v-model="q" class="inp" type="search" placeholder="Search dishes…" />
        <select v-model="category" class="select" style="width: auto; min-width: 180px"><option value="">All categories</option><option v-for="c in categories" :key="c" :value="c">{{ c }}</option></select>
        <span class="muted text-sm">{{ filtered.length }} of {{ dishes.length }} dishes</span>
      </div>
    </div>

    <AppAlert v-if="error" kind="error" title="Could not load dishes">{{ error }}</AppAlert>
    <LoadingState v-else-if="loading" :cards="4" label="Loading dish catalog…" />
    <div v-else-if="filtered.length" class="grid-cards four">
      <article v-for="dish in filtered" :key="dish.id" class="card dish-card">
        <div class="card-media"><img :src="imageOr(dish.image_url)" :alt="dish.name" loading="lazy" /></div>
        <div class="card-body">
          <span class="badge sage">{{ dish.region || 'Uzbekistan' }}</span>
          <h3 class="card-title">{{ dish.name }}</h3>
          <p class="card-desc line-clamp-2">{{ dish.description }}</p>
          <NuxtLink :to="`/dishes/${dish.slug || dish.id}`" class="text-link mt-12">Details <AppIcon name="arrow" :size="16" /></NuxtLink>
        </div>
      </article>
    </div>
    <EmptyState v-else title="No dishes available" message="The administrator has not published any active dishes yet." icon="dish" />
  </div>
</template>
