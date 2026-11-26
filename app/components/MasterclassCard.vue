<script setup lang="ts">
import { useFavorites } from '~/composables/useFavorites'
import { money, duration, imageOr } from '~/composables/useFormat'

const props = withDefaults(
  defineProps<{ mc: any; badge?: string; showFavorite?: boolean }>(),
  { showFavorite: true }
)

const favorites = useFavorites()
const route = useRoute()
const href = computed(() => `/masterclasses/${props.mc?.slug || props.mc?.id}`)
const image = computed(() =>
  imageOr(props.mc?.image_url || props.mc?.dish?.image_url || props.mc?.restaurant?.image_url)
)
const familyName = computed(
  () => props.mc?.chef?.display_name || props.mc?.chef?.full_name || props.mc?.chef_name || ''
)
const ratingCount = computed(() => Number(props.mc?.reviews_count || props.mc?.review_count || 0))
</script>

<template>
  <article class="card" v-if="mc">
    <div class="card-media">
      <NuxtLink :to="href" :aria-label="mc.title">
        <img :src="image" :alt="mc.title" loading="lazy" />
      </NuxtLink>
      <span v-if="badge" class="badge">{{ badge }}</span>
      <button
        v-if="showFavorite"
        class="fav-btn"
        type="button"
        :class="{ on: favorites.has(mc.id) }"
        :aria-label="favorites.has(mc.id) ? 'Remove from favorites' : 'Save to favorites'"
        @click="favorites.toggle(mc.id, route.fullPath)"
      >
        <AppIcon name="heart" :size="18" />
      </button>
    </div>
    <div class="card-body">
      <div class="card-kicker">
        <span class="truncate">{{ mc.restaurant?.name || mc.chef?.display_name || 'TourChef experience' }}</span>
        <span v-if="mc.rating" class="rating">
          <AppIcon name="star" :size="15" /> {{ Number(mc.rating).toFixed(1) }}
          <span v-if="ratingCount">({{ ratingCount }})</span>
        </span>
      </div>
      <h3 class="card-title">
        <NuxtLink :to="href">{{ mc.title }}</NuxtLink>
      </h3>
      <p class="card-desc line-clamp-2">{{ mc.description || 'A hands-on culinary experience in Uzbekistan.' }}</p>
      <div class="card-meta">
        <span v-if="mc.dish?.name || mc.dish_name"><AppIcon name="dish" :size="15" /> {{ mc.dish?.name || mc.dish_name }}</span>
        <span><AppIcon name="clock" :size="15" /> {{ duration(mc.duration_min) }}</span>
        <span v-if="mc.city"><AppIcon name="pin" :size="15" /> {{ mc.city }}</span>
        <span v-if="familyName"><AppIcon name="family" :size="15" /> {{ familyName }}</span>
      </div>
      <div class="card-foot">
        <span class="price">{{ money(mc.price) }} <small>/ person</small></span>
        <NuxtLink :to="`${href}#booking`" class="text-link">Book now <AppIcon name="arrow" :size="16" /></NuxtLink>
      </div>
    </div>
  </article>
</template>

<style scoped>
.card-media > a { display: block; width: 100%; height: 100%; }
.card-title a:hover { color: var(--accent); }
.fav-btn {
  position: absolute; top: 12px; right: 12px; z-index: 2;
  width: 34px; height: 34px; display: grid; place-items: center;
  border: 0; border-radius: 50%; background: rgba(255,255,255,.94);
  color: var(--text-2); transition: .18s;
}
.fav-btn:hover { color: var(--accent); transform: scale(1.08); }
.fav-btn.on { color: var(--accent); }
.fav-btn.on svg { fill: var(--accent); }
</style>
