<script setup lang="ts">
import { imageOr } from '~/composables/useFormat'

const props = defineProps<{ dish: any }>()
const href = computed(() => `/dishes/${props.dish?.slug || props.dish?.id}`)
const image = computed(() => imageOr(props.dish?.image_url))
</script>

<template>
  <article class="card dish-card" v-if="dish">
    <NuxtLink :to="href">
      <div class="card-media"><img :src="image" :alt="dish.name" loading="lazy" /></div>
      <div class="card-body">
        <div v-if="dish.region" class="card-kicker"><span>{{ dish.region }}</span></div>
        <h3 class="card-title">{{ dish.name }}</h3>
        <p class="card-desc line-clamp-2">{{ dish.description || 'Discover this traditional Uzbek dish.' }}</p>
        <span class="text-link mt-12">Explore <AppIcon name="arrow" :size="16" /></span>
      </div>
    </NuxtLink>
  </article>
</template>
