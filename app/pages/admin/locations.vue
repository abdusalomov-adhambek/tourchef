<script setup lang="ts">
import { useApi } from '~/composables/useApi'

definePageMeta({ layout: 'admin' })

const api = useApi()
const regions = [
  'Tashkent', 'Samarkand', 'Bukhara', 'Fergana', 'Andijan', 'Namangan', 'Kashkadarya',
  'Surkhandarya', 'Jizzakh', 'Syrdarya', 'Navoiy', 'Karakalpakstan', 'Khorezm', 'Tashkent Region'
]
const dishes = ref<any[]>([])
const restaurants = ref<any[]>([])
const loading = ref(true)

const dishRegions = computed(() =>
  Array.from(new Set(dishes.value.map((d) => d.region).filter(Boolean))).sort()
)
const cities = computed(() =>
  Array.from(new Set(restaurants.value.map((r) => r.city).filter(Boolean))).sort()
)

onMounted(async () => {
  const [d, r] = await Promise.allSettled([
    api.get<any>('/dishes', { limit: 200 }),
    api.get<any[]>('/restaurants')
  ])
  if (d.status === 'fulfilled') dishes.value = Array.isArray(d.value?.items) ? d.value.items : []
  if (r.status === 'fulfilled') restaurants.value = Array.isArray(r.value) ? r.value : []
  loading.value = false
})
</script>

<template>
  <div class="agrid r2">
    <div class="acard">
      <div class="acard-h"><h3>Uzbekistan regions</h3><span class="muted">Country → Region</span></div>
      <div class="acard-b">
        <div class="chips">
          <span v-for="region in regions" :key="region" class="achip" :class="{ on: dishRegions.includes(region) }">
            {{ region }} <small v-if="dishRegions.includes(region)">(in catalog)</small>
          </span>
        </div>
      </div>
    </div>
    <div class="acard">
      <div class="acard-h"><h3>Active locations</h3><span class="muted">From live data</span></div>
      <div class="acard-b">
        <LoadingState v-if="loading" :cards="1" label="Loading locations…" />
        <template v-else>
          <div class="abox"><h4>Regions used by dishes</h4><p>{{ dishRegions.length ? dishRegions.join(', ') : 'No regions assigned yet.' }}</p></div>
          <div class="abox"><h4>Cities with partner restaurants</h4><p>{{ cities.length ? cities.join(', ') : 'No restaurants listed yet.' }}</p></div>
          <NuxtLink to="/admin/dishes" class="abtn mt-12">Manage dish regions <AppIcon name="chevron" :size="15" /></NuxtLink>
        </template>
      </div>
    </div>
  </div>
</template>
