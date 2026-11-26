<script setup lang="ts">
import { compactNumber, imageOr, initials } from '~/composables/useFormat'

const props = defineProps<{ reel: any }>()
const creator = computed(
  () => props.reel?.chef?.display_name || props.reel?.restaurant?.name || 'Family host'
)
const href = computed(() => `/reels?reel=${props.reel?.id || ''}`)
</script>

<template>
  <NuxtLink :to="href" class="reel-preview" v-if="reel">
    <img :src="imageOr(reel.image_url)" :alt="reel.title || 'TourChef reel'" loading="lazy" />
    <span class="play"><AppIcon name="play" :size="54" /></span>
    <span v-if="reel.views" class="views">{{ compactNumber(reel.views) }} views</span>
    <span class="shade" />
    <span class="rp-meta">
      <b>{{ creator }}</b>
      <small>{{ reel.title || reel.caption || 'Culinary story' }}</small>
    </span>
    <span class="sr-only">{{ initials(creator) }}</span>
  </NuxtLink>
</template>

<style scoped>
.sr-only { position: absolute; width: 1px; height: 1px; overflow: hidden; clip: rect(0,0,0,0); }
</style>
