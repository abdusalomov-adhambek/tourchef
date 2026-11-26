<script setup lang="ts">
import { useApi } from '~/composables/useApi'
import { imageOr } from '~/composables/useFormat'

definePageMeta({ layout: 'default' })

const route = useRoute()
const api = useApi()
const dish = ref<any>(null)
const related = ref<any[]>([])
const loading = ref(true)
const error = ref('')

const ingredients = computed(() => {
  const raw = dish.value?.ingredients
  if (Array.isArray(raw)) return raw.map((x) => String(x))
  if (typeof raw === 'string') return raw.split(/[,;]\s*/).filter(Boolean)
  return []
})

onMounted(async () => {
  try {
    const data = await api.get<any>(`/dishes/${route.params.slug}`)
    dish.value = data
    if (data?.name) {
      try {
        const mc = await api.get<any>('/masterclasses', { dish: data.name, limit: 6 })
        related.value = Array.isArray(mc?.items) ? mc.items : []
      } catch { related.value = [] }
    }
  } catch (e: any) {
    error.value = e?.message || 'Dish not found.'
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div>
    <div class="container">
      <AppAlert v-if="error" kind="error" title="Dish unavailable">{{ error }}
        <div class="mt-16"><NuxtLink to="/dishes" class="btn btn-outline btn-sm">All dishes</NuxtLink></div>
      </AppAlert>
      <div v-else-if="loading" class="stack">
        <div class="skeleton" style="height: 360px; border-radius: 24px" />
        <div class="skeleton" style="height: 28px; width: 40%" />
      </div>
      <template v-else-if="dish">
        <div class="page-head">
          <NuxtLink to="/dishes" class="text-link"><AppIcon name="back" :size="16" /> All dishes</NuxtLink>
          <span class="eyebrow mt-16">{{ dish.region || 'Uzbekistan' }} · {{ dish.category || 'Traditional cuisine' }}</span>
          <h1>{{ dish.name }}</h1>
        </div>
        <div class="detail-layout">
          <div>
            <div class="detail-hero" style="height: 380px">
              <img :src="imageOr(dish.image_url)" :alt="dish.name" />
            </div>
            <div class="detail-section mt-24">
              <h2 class="section-title small">The story of this dish</h2>
              <p class="prose">{{ dish.description || 'A beloved part of the Uzbek dastorkhon.' }}</p>
            </div>
            <div v-if="ingredients.length" class="detail-section">
              <h2 class="section-title small">Ingredients</h2>
              <div class="chips">
                <span v-for="(item, i) in ingredients" :key="i" class="chip">{{ item }}</span>
              </div>
            </div>
          </div>
          <aside class="widget">
            <h3>Cook it yourself</h3>
            <p class="muted text-sm mt-8">Join a hands-on masterclass and learn this recipe with a local family.</p>
            <NuxtLink :to="`/masterclasses?q=${encodeURIComponent(dish.name || '')}`" class="btn btn-primary btn-block mt-16">
              Find masterclasses <AppIcon name="arrow" :size="16" />
            </NuxtLink>
          </aside>
        </div>

        <section class="section tight">
          <div class="section-head">
            <div><span class="eyebrow">Learn to cook it</span><h2 class="section-title small">Masterclasses with {{ dish.name }}</h2></div>
            <NuxtLink to="/masterclasses" class="text-link">All experiences <AppIcon name="arrow" :size="16" /></NuxtLink>
          </div>
          <div v-if="related.length" class="grid-cards">
            <MasterclassCard v-for="mc in related" :key="mc.id" :mc="mc" />
          </div>
          <EmptyState v-else title="No sessions for this dish yet" message="Explore all culinary experiences to find something similar." icon="family" action-to="/masterclasses" action-label="Browse masterclasses" />
        </section>
      </template>
    </div>
  </div>
</template>
