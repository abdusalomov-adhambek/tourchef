<script setup lang="ts">
import { imageOr } from '~/composables/useFormat'

const props = defineProps<{
  region: any
  x: number
  y: number
  selected?: boolean
}>()
const emit = defineEmits(['select', 'hover', 'leave'])
</script>

<template>
  <g
    class="rf-marker"
    :class="{ 'is-selected': selected }"
    :transform="`translate(${x}, ${y})`"
    role="button"
    tabindex="0"
    :aria-label="`${region.name}: ${region.dish_name}`"
    @click.stop="emit('select', region.slug)"
    @keydown.enter.prevent="emit('select', region.slug)"
    @keydown.space.prevent="emit('select', region.slug)"
    @mouseenter="emit('hover', region)"
    @mouseleave="emit('leave')"
  >
    <circle class="rf-hit" r="34" />
    <circle class="rf-marker-halo" r="17" />
    <circle class="rf-marker-dot" r="11" />
    <image :href="imageOr(region.dish_image_url)" x="-9" y="-9" width="18" height="18" preserveAspectRatio="xMidYMid slice" />
    <text class="rf-marker-label" y="30">{{ region.dish_name }}</text>
  </g>
</template>

<style scoped>
.rf-marker { cursor: pointer; outline: none; }
.rf-hit { fill: transparent; stroke: none; }
.rf-marker-halo { fill: rgba(198,93,58,.16); transition: r .2s, fill .2s; }
.rf-marker-dot { fill: #fff; stroke: #C65D3A; stroke-width: 2; transition: stroke-width .2s; }
.rf-marker image { clip-path: circle(9px); }
.rf-marker-label { font: 600 11px "DM Sans", sans-serif; fill: #252923; text-anchor: middle; paint-order: stroke; stroke: #F8F5EF; stroke-width: 3px; stroke-linejoin: round; opacity: 0; transition: opacity .2s; pointer-events: none; }
.rf-marker:hover .rf-marker-halo, .rf-marker:focus .rf-marker-halo, .rf-marker.is-selected .rf-marker-halo { fill: rgba(198,93,58,.28); r: 21px; }
.rf-marker:hover .rf-marker-label, .rf-marker:focus .rf-marker-label, .rf-marker.is-selected .rf-marker-label { opacity: 1; }
.rf-marker.is-selected .rf-marker-dot { stroke-width: 3.5; }
</style>
