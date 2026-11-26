<script setup lang="ts">
import { geoIdentity, geoPath } from 'd3-geo'
import { imageOr } from '~/composables/useFormat'

const props = defineProps<{ regions: any[]; selected?: string }>()
const emit = defineEmits(['select'])

const W = 1000
// O'zbekiston keng va past: viewBox balandligini xarita nisbatiga moslashtiramiz
// (eski 700 qiymati tepa/pastda katta bo'sh qoldirar edi).
const H = 520
const geo = ref<any>(null)
const loading = ref(true)
const error = ref('')
const hovered = ref<any>(null)
const tip = reactive({ x: 0, y: 0 })

// O'zbekiston chegaralari kichik hududda: tekis proyeksiya yetarli va
// ba'zi GeoJSON halqalarining teskari yo'nalishidan kelib chiqadigan
// "butun dunyo" muammosining oldini oladi.
const MIN_X = 55.5
const MAX_X = 73.5
const MIN_Y = 36.8
const MAX_Y = 45.9

function makeProjection() {
  const scale = Math.min((W - 60) / (MAX_X - MIN_X), (H - 60) / (MAX_Y - MIN_Y))
  const projection = geoIdentity().reflectY(true).scale(scale).translate([30 - MIN_X * scale, H - 30 + MIN_Y * scale])
  return { projection, path: geoPath(projection) }
}

const shapes = computed(() => {
  if (!geo.value) return []
  const { path } = makeProjection()
  return geo.value.features.map((f: any) => ({
    id: String(f.properties.id),
    name: f.properties.name,
    d: path(f),
    label: path.centroid(f)
  }))
})

const markers = computed(() => {
  if (!geo.value) return []
  const { projection } = makeProjection()
  return props.regions
    .map((r) => {
      const p = projection([Number(r.longitude), Number(r.latitude)])
      return p ? { region: r, x: p[0], y: p[1] } : null
    })
    .filter(Boolean) as { region: any; x: number; y: number }[]
})

function regionForGeo(id: string) {
  return props.regions.find((r) => String(r.geo_id) === id)
}

function onMove(e: MouseEvent) {
  const box = (e.currentTarget as SVGElement).getBoundingClientRect()
  tip.x = e.clientX - box.left
  tip.y = e.clientY - box.top
}

onMounted(async () => {
  try {
    const res = await fetch('/geo/uzbekistan-regions.geojson')
    if (!res.ok) throw new Error(`GeoJSON: ${res.status}`)
    geo.value = await res.json()
  } catch (e: any) {
    error.value = e?.message || 'Xarita ma\'lumotlari yuklanmadi.'
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div class="rf-map-wrap">
    <LoadingState v-if="loading" :cards="1" label="Xarita yuklanayapti…" />
    <AppAlert v-else-if="error" kind="error">{{ error }}</AppAlert>
    <template v-else>
      <svg class="rf-map" :viewBox="`0 0 ${W} ${H}`" role="img" aria-label="Uzbekistan culinary regions map" @mousemove="onMove">
        <g class="rf-regions">
          <path
            v-for="s in shapes"
            :key="s.id"
            :d="s.d"
            :class="{ 'is-active': selected === regionForGeo(s.id)?.slug, 'is-hovered': hovered?.geo_id === s.id }"
            @click="regionForGeo(s.id) && emit('select', regionForGeo(s.id).slug)"
            @mouseenter="hovered = regionForGeo(s.id)"
            @mouseleave="hovered = null"
          />
        </g>
        <g class="rf-region-names" aria-hidden="true">
          <text v-for="s in shapes" :key="s.id" :x="s.label[0]" :y="s.label[1]">{{ s.name.replace(/ region| city| province|Republic of/g, '') }}</text>
        </g>
        <RegionFoodMarker
          v-for="m in markers"
          :key="m.region.slug"
          :region="m.region"
          :x="m.x"
          :y="m.y"
          :selected="selected === m.region.slug"
          @select="emit('select', $event)"
          @hover="hovered = $event"
          @leave="hovered = null"
        />
      </svg>
      <div v-if="hovered" class="rf-tip" :style="{ left: `${tip.x}px`, top: `${tip.y}px` }">
        <img :src="imageOr(hovered.dish_image_url)" alt="" />
        <div><b>{{ hovered.name }}</b><small>{{ hovered.dish_name }}</small></div>
      </div>
    </template>
  </div>
</template>

<style scoped>
.rf-map-wrap { position: relative; width: 100%; min-height: 0; }
.rf-map { width: 100%; height: auto; display: block; }
.rf-regions path { fill: #EFE6D9; stroke: #D8D0C3; stroke-width: 1.4; cursor: pointer; transition: fill .2s, stroke .2s; }
.rf-regions path:hover, .rf-regions path.is-hovered { fill: var(--accent-soft); stroke: var(--accent); }
.rf-regions path.is-active { fill: #E9C7B7; stroke: var(--accent); stroke-width: 2.4; }
.rf-region-names text { font: 600 12px "DM Sans", sans-serif; fill: #77766F; text-anchor: middle; pointer-events: none; }
.rf-tip { position: absolute; z-index: 5; transform: translate(12px, -50%); display: flex; gap: 8px; align-items: center; padding: 6px 10px 6px 6px; background: #fff; border: 1px solid var(--border); border-radius: 12px; box-shadow: 0 8px 24px rgba(37,41,35,.12); pointer-events: none; white-space: nowrap; }
.rf-tip img { width: 36px; height: 36px; object-fit: cover; border-radius: 8px; }
.rf-tip b { display: block; font-size: 12px; }
.rf-tip small { display: block; font-size: 11px; color: var(--text-2); }
@media (max-width: 640px) { .rf-region-names text { display: none; } }
</style>
