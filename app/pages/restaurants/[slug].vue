<script setup lang="ts">
import { useApi } from '~/composables/useApi'
import { imageOr } from '~/composables/useFormat'

definePageMeta({ layout: 'default' })

const route = useRoute()
const api = useApi()
const restaurant = ref<any>(null)
const masterclasses = ref<any[]>([])
const loading = ref(true)
const error = ref('')

onMounted(async () => {
  try {
    const data = await api.get<any>(`/restaurants/${route.params.slug}`)
    restaurant.value = data
    if (Array.isArray(data?.masterclasses)) {
      masterclasses.value = data.masterclasses
    } else {
      try {
        const mc = await api.get<any>('/masterclasses', { city: data?.city, limit: 20 })
        masterclasses.value = Array.isArray(mc?.items)
          ? mc.items.filter((m: any) => !data?.id || m.restaurant_id === data.id || m.restaurant?.id === data.id)
          : []
        if (!masterclasses.value.length) masterclasses.value = Array.isArray(mc?.items) ? mc.items : []
      } catch { masterclasses.value = [] }
    }
  } catch (e: any) {
    error.value = e?.message || 'Restaurant not found.'
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div>
    <div class="container">
      <AppAlert v-if="error" kind="error" title="Restaurant unavailable">{{ error }}
        <div class="mt-16"><NuxtLink to="/restaurants" class="btn btn-outline btn-sm">All restaurants</NuxtLink></div>
      </AppAlert>
      <div v-else-if="loading" class="stack">
        <div class="skeleton" style="height: 380px; border-radius: 24px" />
        <div class="skeleton" style="height: 28px; width: 45%" />
      </div>
      <template v-else-if="restaurant">
        <div class="page-head">
          <NuxtLink to="/restaurants" class="text-link"><AppIcon name="back" :size="16" /> All restaurants</NuxtLink>
          <h1 class="mt-16">{{ restaurant.name }}</h1>
          <p>{{ restaurant.city }} · {{ restaurant.cuisine || 'Uzbek cuisine' }}</p>
        </div>
        <div class="detail-hero" style="height: 420px">
          <img :src="imageOr(restaurant.image_url)" :alt="restaurant.name" />
          <div class="dh-shade" />
          <div class="dh-copy">
            <span v-if="restaurant.rating" class="badge gold"><AppIcon name="star" :size="14" /> {{ Number(restaurant.rating).toFixed(1) }} rating</span>
            <h1>{{ restaurant.name }}</h1>
          </div>
        </div>

        <div class="detail-layout mt-32">
          <div>
            <div class="detail-section">
              <span class="eyebrow">Welcome</span>
              <h2 class="section-title small">About this kitchen</h2>
              <p class="prose">{{ restaurant.description || 'A TourChef partner restaurant serving authentic Uzbek cuisine.' }}</p>
            </div>
            <div class="detail-section">
              <h2 class="section-title small">Experiences here</h2>
              <div v-if="masterclasses.length" class="grid-cards">
                <MasterclassCard v-for="mc in masterclasses" :key="mc.id" :mc="mc" />
              </div>
              <EmptyState v-else title="No experiences scheduled" message="This kitchen has not published a masterclass yet." icon="family" action-to="/masterclasses" action-label="Browse experiences" />
            </div>
          </div>
          <aside class="widget">
            <h3>Visit {{ restaurant.name }}</h3>
            <div class="info-row mt-12"><AppIcon name="pin" :size="18" /><span>City</span> {{ restaurant.city || 'Uzbekistan' }}</div>
            <div class="info-row"><AppIcon name="dish" :size="18" /><span>Cuisine</span> {{ restaurant.cuisine || 'Uzbek' }}</div>
            <div class="info-row"><AppIcon name="star" :size="18" /><span>Rating</span> {{ restaurant.rating ? Number(restaurant.rating).toFixed(1) : 'New' }}</div>
            <NuxtLink to="/masterclasses" class="btn btn-primary btn-block mt-16">Explore experiences <AppIcon name="arrow" :size="16" /></NuxtLink>
          </aside>
        </div>
      </template>
    </div>
  </div>
</template>
