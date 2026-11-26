<script setup lang="ts">
import { imageOr } from '~/composables/useFormat'

const props = defineProps<{ restaurant: any; count?: number }>()
const href = computed(() => `/restaurants/${props.restaurant?.slug || props.restaurant?.id}`)
const image = computed(() => imageOr(props.restaurant?.image_url))
</script>

<template>
  <article class="card rest-card" v-if="restaurant">
    <NuxtLink :to="href">
      <div class="card-media"><img :src="image" :alt="restaurant.name" loading="lazy" /></div>
      <div class="card-body" style="padding-top: 0">
        <div class="rest-top" style="padding-left: 0; padding-right: 0">
          <img class="rest-logo" :src="image" alt="" />
          <div>
            <h3 class="card-title" style="margin: 0">{{ restaurant.name }}</h3>
            <span class="muted" style="font-size: 13px">
              {{ restaurant.city }} · {{ restaurant.cuisine || 'Uzbek cuisine' }}
            </span>
          </div>
        </div>
        <div class="card-meta">
          <span v-if="restaurant.rating" class="rating"><AppIcon name="star" :size="15" /> {{ Number(restaurant.rating).toFixed(1) }}</span>
          <span v-if="count !== undefined">{{ count }} masterclasses</span>
          <span v-else-if="restaurant.masterclasses_count !== undefined">{{ restaurant.masterclasses_count }} masterclasses</span>
          <span><AppIcon name="pin" :size="15" /> {{ restaurant.city || 'Uzbekistan' }}</span>
        </div>
      </div>
    </NuxtLink>
  </article>
</template>
