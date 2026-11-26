<script setup lang="ts">
import { useApi } from '~/composables/useApi'

definePageMeta({ layout: 'default' })

const api = useApi()
const items = ref<any[]>([])
const loading = ref(true)
const error = ref('')
const q = ref('')
const city = ref('')
const cities = computed(() => Array.from(new Set(items.value.map((r) => r.city).filter(Boolean))).sort())
const filtered = computed(() =>
  items.value.filter((r) => {
    const matchQ = !q.value || `${r.name} ${r.cuisine} ${r.description || ''}`.toLowerCase().includes(q.value.toLowerCase())
    const matchCity = !city.value || r.city === city.value
    return matchQ && matchCity
  })
)

onMounted(async () => {
  try {
    const data = await api.get<any[]>('/restaurants')
    items.value = Array.isArray(data) ? data : []
  } catch (e: any) {
    error.value = e?.message || 'Could not load restaurants.'
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div>
    <div class="container">
      <div class="page-head">
        <span class="eyebrow">Our partner kitchens</span>
        <h1>Restaurants</h1>
        <p>The places where Uzbek culinary traditions come alive.</p>
      </div>
      <div class="panel mb-24">
        <div class="panel-b row">
          <input v-model="q" class="inp" type="search" placeholder="Search restaurants…" />
          <select v-model="city" class="select" style="width: auto; min-width: 180px">
            <option value="">All cities</option>
            <option v-for="c in cities" :key="c" :value="c">{{ c }}</option>
          </select>
          <span class="muted text-sm">{{ filtered.length }} of {{ items.length }}</span>
        </div>
      </div>
      <AppAlert v-if="error" kind="error" title="Could not load restaurants">{{ error }}</AppAlert>
      <LoadingState v-else-if="loading" :cards="3" label="Loading restaurants…" />
      <div v-else-if="filtered.length" class="grid-cards">
        <RestaurantCard v-for="r in filtered" :key="r.id" :restaurant="r" />
      </div>
      <EmptyState v-else title="No restaurants found" message="Try a different search or browse all experiences." icon="restaurant" action-to="/masterclasses" action-label="Explore masterclasses" />
    </div>
  </div>
</template>
