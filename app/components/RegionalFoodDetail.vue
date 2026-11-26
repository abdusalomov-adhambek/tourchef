<script setup lang="ts">
import { imageOr } from '~/composables/useFormat'

const props = defineProps<{
  region: any
  masterclasses: any[]
  restaurants: any[]
  loading?: boolean
}>()
</script>

<template>
  <div class="rf-detail">
    <template v-if="region">
      <div class="rf-detail-media">
        <img :src="imageOr(region.dish_image_url)" :alt="region.dish_name" />
        <span class="badge sage">{{ region.name }}</span>
      </div>
      <div class="rf-detail-body">
        <span class="eyebrow">Featured dish</span>
        <h3>{{ region.dish_name }}</h3>
        <p class="muted">{{ region.description }}</p>
        <div class="rf-detail-section" v-if="region.ingredients?.length">
          <h4>Traditional ingredients</h4>
          <div class="chips"><span v-for="i in region.ingredients" :key="i" class="chip">{{ i }}</span></div>
        </div>
        <div class="rf-detail-section" v-if="region.cultural_info">
          <h4>Cultural background</h4>
          <p class="muted">{{ region.cultural_info }}</p>
        </div>
        <div class="rf-actions">
          <NuxtLink :to="`/dishes`" class="btn btn-primary btn-sm">Explore Dish</NuxtLink>
          <NuxtLink :to="`/masterclasses?city=${encodeURIComponent(region.city)}`" class="btn btn-outline btn-sm">Find Masterclasses</NuxtLink>
          <NuxtLink :to="`/regions/${region.slug}`" class="btn btn-ghost btn-sm">Explore Region</NuxtLink>
        </div>
        <div class="rf-detail-section" v-if="restaurants.length">
          <h4>Restaurants in {{ region.city }}</h4>
          <div class="rf-rest-list">
            <NuxtLink v-for="r in restaurants" :key="r.id" :to="`/restaurants/${r.slug}`" class="rf-rest">
              <img :src="imageOr(r.image_url)" alt="" /><span>{{ r.name }}</span><small>★ {{ r.rating }}</small>
            </NuxtLink>
          </div>
        </div>
        <div class="rf-detail-section">
          <h4>Experience this dish</h4>
          <RegionalMasterclassList :items="masterclasses" :loading="loading" />
        </div>
      </div>
    </template>
    <EmptyState v-else title="Tanlangan mintaqa yo'q" message="Xaritadan bir mintaqani tanlang." icon="pin" />
  </div>
</template>

<style scoped>
.rf-detail { border: 1px solid var(--border); border-radius: var(--radius); background: #fff; overflow: hidden; }
.rf-detail-media { position: relative; height: 190px; }
.rf-detail-media img { width: 100%; height: 100%; object-fit: cover; }
.rf-detail-media .badge { position: absolute; left: 14px; bottom: 14px; }
.rf-detail-body { padding: 20px; }
.rf-detail-body h3 { font-family: var(--font-display); font-size: 26px; margin: 6px 0; }
.rf-detail-body p { font-size: 13px; line-height: 1.6; }
.rf-detail-section { margin-top: 18px; }
.rf-detail-section h4 { font-size: 13px; text-transform: uppercase; letter-spacing: .08em; color: var(--text-2); margin-bottom: 8px; }
.rf-actions { display: flex; gap: 8px; flex-wrap: wrap; margin-top: 18px; }
.rf-rest-list { display: grid; gap: 6px; }
.rf-rest { display: flex; gap: 9px; align-items: center; padding: 7px 9px; border: 1px solid var(--border); border-radius: 9px; font-size: 13px; }
.rf-rest img { width: 34px; height: 34px; object-fit: cover; border-radius: 7px; }
.rf-rest span { flex: 1; }
.rf-rest small { color: var(--text-2); }
</style>
