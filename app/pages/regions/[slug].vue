<script setup lang="ts">
import { useApi } from '~/composables/useApi'
import { imageOr } from '~/composables/useFormat'

definePageMeta({ layout: 'default' })
const api = useApi()
const route = useRoute()
const slug = String(route.params.slug)
const region = ref<any>(null)
const masterclasses = ref<any[]>([])
const restaurants = ref<any[]>([])
const loading = ref(true)
const error = ref('')

onMounted(async () => {
  const [d, m, r] = await Promise.allSettled([
    api.get<any>(`/v1/regions/${slug}`),
    api.get<any>(`/v1/regions/${slug}/masterclasses`, { limit: 12 }),
    api.get<any[]>(`/v1/regions/${slug}/restaurants`)
  ])
  if (d.status === 'fulfilled') region.value = d.value
  else error.value = 'Mintaqa topilmadi.'
  if (m.status === 'fulfilled') masterclasses.value = Array.isArray(m.value?.items) ? m.value.items : []
  if (r.status === 'fulfilled') restaurants.value = Array.isArray(r.value) ? r.value : []
  loading.value = false
})
</script>

<template>
  <div class="section container rf-section">
    <div class="rf-back"><NuxtLink to="/regions" class="text-link">← Barcha mintaqalar</NuxtLink></div>
    <AppAlert v-if="error" kind="error" class="mt-16">{{ error }}</AppAlert>
    <LoadingState v-if="loading" :cards="1" label="Yuklanayapti…" />
    <template v-else-if="region">
      <span class="eyebrow mt-24">{{ region.name }} · {{ region.name_uz }}</span>
      <h1 class="page-title" style="margin:10px 0">{{ region.dish_name }}</h1>
      <p class="muted" style="max-width:70ch">{{ region.description }}</p>
      <div class="rf-map-full mt-24">
        <ClientOnly>
          <UzbekistanFoodMap :regions="[region]" :selected="region.slug" />
        </ClientOnly>
      </div>
      <div class="rf-detail-below mt-24">
        <RegionalFoodDetail :region="region" :masterclasses="masterclasses" :restaurants="restaurants" />
      </div>
    </template>
  </div>
</template>
