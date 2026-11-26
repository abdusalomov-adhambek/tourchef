<script setup lang="ts">
import { useApi } from '~/composables/useApi'
import { useAuth } from '~/composables/useAuth'
import { useToast } from '~/composables/useToast'
import { compactNumber, imageOr } from '~/composables/useFormat'

definePageMeta({ layout: 'reels' })

const api = useApi()
const auth = useAuth()
const toast = useToast()
const router = useRouter()
const route = useRoute()

const reels = ref<any[]>([])
const mcMap = ref<Record<string, string>>({})
const loading = ref(true)
const error = ref('')
const tab = ref<'all' | 'masterclasses' | 'restaurants'>('all')
const videos = new Map<string, HTMLVideoElement>()
const ratios = reactive<Record<string, string>>({})
const activeReel = ref<any>(null)
const modalVideo = ref<HTMLVideoElement | null>(null)

const filtered = computed(() => {
  if (tab.value === 'masterclasses') return reels.value.filter((r) => reelMcId(r))
  if (tab.value === 'restaurants') return reels.value.filter((r) => r.restaurant || r.restaurant_id)
  return reels.value
})

function reelMcId(reel: any): string | number | null {
  return reel?.masterclass_id || reel?.masterclass?.id || reel?.mc_id || null
}

function bookPath(reel: any): string | null {
  const id = reelMcId(reel)
  if (id && mcMap.value[String(id)]) return `/masterclasses/${mcMap.value[String(id)]}#booking`
  if (reel?.masterclass?.slug) return `/masterclasses/${reel.masterclass.slug}#booking`
  if (id) return `/masterclasses?id=${id}`
  return null
}

function setVideo(el: any, id: string) {
  if (el) videos.set(id, el as HTMLVideoElement)
  else videos.delete(id)
}

function setRatioFromVideo(el: any, id: string) {
  const video = el as HTMLVideoElement
  if (video?.videoWidth && video?.videoHeight) ratios[id] = `${video.videoWidth} / ${video.videoHeight}`
}

function setRatioFromImage(el: any, id: string) {
  const img = el as HTMLImageElement
  if (img?.naturalWidth && img?.naturalHeight) ratios[id] = `${img.naturalWidth} / ${img.naturalHeight}`
}

function reelRatio(reel: any): string {
  return ratios[String(reel.id)] || '16 / 9'
}

function hoverPlay(id: string, on: boolean) {
  const video = videos.get(String(id))
  if (!video) return
  if (on) video.play().catch(() => {})
  else video.pause()
}

async function openReel(reel: any) {
  activeReel.value = reel
  if (import.meta.client) document.body.style.overflow = 'hidden'
  await nextTick()
  modalVideo.value?.play().catch(() => {})
}

function closeReel() {
  modalVideo.value?.pause()
  activeReel.value = null
  if (import.meta.client) document.body.style.overflow = ''
}

function onKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape') closeReel()
}

function openBooking(reel: any) {
  const path = bookPath(reel)
  closeReel()
  if (!path) {
    if (reel?.restaurant?.slug) router.push(`/restaurants/${reel.restaurant.slug}`)
    else toast.info('This story is not linked to a bookable experience yet.')
    return
  }
  if (!auth.token.value) {
    toast.info('Sign in to reserve your place — your experience is saved.')
    router.push(`/login?redirect=${encodeURIComponent(path)}`)
  } else {
    router.push(path)
  }
}

async function share(reel: any) {
  const url = `${window.location.origin}/reels?reel=${reel.id}`
  try {
    if (navigator.share) {
      await navigator.share({ title: reel.title || 'TourChef reel', url })
    } else {
      await navigator.clipboard.writeText(url)
      toast.success('Reel link copied.')
    }
  } catch {
    toast.info('You can copy this page URL to share the reel.')
  }
}

onMounted(async () => {
  window.addEventListener('keydown', onKeydown)
  const [reelData, mcData] = await Promise.allSettled([
    api.get<any[]>('/reels'),
    api.get<any>('/masterclasses', { limit: 100 })
  ])
  if (reelData.status === 'fulfilled') reels.value = Array.isArray(reelData.value) ? reelData.value : []
  else error.value = reelData.reason?.message || 'Could not load reels.'
  if (mcData.status === 'fulfilled' && Array.isArray(mcData.value?.items)) {
    const map: Record<string, string> = {}
    for (const mc of mcData.value.items) map[String(mc.id)] = mc.slug || String(mc.id)
    mcMap.value = map
  }
  loading.value = false
  const target = String(route.query.reel || '')
  if (target) {
    await nextTick()
    document.getElementById(`reel-${target}`)?.scrollIntoView({ behavior: 'smooth', block: 'center' })
  }
})

onBeforeUnmount(() => {
  window.removeEventListener('keydown', onKeydown)
  if (import.meta.client) document.body.style.overflow = ''
})
</script>

<template>
  <div class="reels-page">
    <div class="reels-filter-bar">
      <button class="btn btn-sm" :class="tab === 'all' ? 'btn-primary' : 'btn-outline'" type="button" @click="tab = 'all'">For you</button>
      <button class="btn btn-sm" :class="tab === 'masterclasses' ? 'btn-primary' : 'btn-outline'" type="button" @click="tab = 'masterclasses'">Masterclasses</button>
      <button class="btn btn-sm" :class="tab === 'restaurants' ? 'btn-primary' : 'btn-outline'" type="button" @click="tab = 'restaurants'">Restaurants</button>
    </div>

    <div v-if="loading" class="reels-empty"><div class="muted">Loading culinary stories…</div></div>
    <AppAlert v-else-if="error" kind="error" class="mt-32">{{ error }}</AppAlert>
    <div v-else-if="!filtered.length" class="reels-empty">
      <div>
        <h2 style="font-family: var(--font-display); font-size: 32px">No reels in this feed</h2>
        <p class="muted mt-12">Try another tab or explore our kitchens.</p>
        <NuxtLink to="/masterclasses" class="btn btn-primary mt-16">Explore masterclasses</NuxtLink>
      </div>
    </div>

    <div v-else class="ig-grid">
      <article
        v-for="reel in filtered"
        :key="reel.id"
        :id="`reel-${reel.id}`"
        class="ig-tile"
        @mouseenter="hoverPlay(String(reel.id), true)"
        @mouseleave="hoverPlay(String(reel.id), false)"
      >
        <video
          v-if="reel.video_url"
          :ref="(el) => setVideo(el, String(reel.id))"
          class="ig-media"
          :style="{ aspectRatio: reelRatio(reel) }"
          :src="reel.video_url"
          :poster="imageOr(reel.image_url)"
          muted
          loop
          playsinline
          preload="metadata"
          @loadedmetadata="setRatioFromVideo($event, String(reel.id))"
        />
        <img
          v-else
          class="ig-media"
          :style="{ aspectRatio: reelRatio(reel) }"
          :src="imageOr(reel.image_url)"
          :alt="reel.title || reel.caption || 'TourChef reel'"
          loading="lazy"
          @load="setRatioFromImage($event, String(reel.id))"
        />
        <button class="ig-open" type="button" :aria-label="`Play ${reel.title || 'reel'}`" @click="openReel(reel)">
          <AppIcon name="play" :size="26" />
        </button>
        <div class="ig-meta">
          <b>{{ reel.title || reel.caption || 'Culinary story' }}</b>
          <small>{{ reel.chef?.display_name || reel.restaurant?.name || 'TourChef' }} · {{ compactNumber(reel.views) }} views</small>
        </div>
      </article>
    </div>

    <!-- FULLSCREEN PLAYER -->
    <div v-if="activeReel" class="ig-modal" role="dialog" aria-modal="true" aria-label="Reel video" @click.self="closeReel">
      <button class="ig-close" type="button" aria-label="Close video" @click="closeReel">
        <AppIcon name="close" :size="24" />
      </button>
      <div class="ig-player">
        <video
          ref="modalVideo"
          class="ig-player-video"
          :src="activeReel.video_url || ''"
          :poster="imageOr(activeReel.image_url)"
          autoplay
          loop
          playsinline
          controls
        />
        <div class="ig-player-bottom">
          <div class="ig-player-meta">
            <b>{{ activeReel.title || activeReel.caption || 'TourChef reel' }}</b>
            <small>{{ activeReel.chef?.display_name || activeReel.restaurant?.name || 'TourChef' }}</small>
          </div>
          <div class="ig-player-actions">
            <button class="btn btn-outline btn-light" type="button" @click="share(activeReel)">
              <AppIcon name="share" :size="18" /> Share
            </button>
            <button class="btn btn-primary" type="button" @click="openBooking(activeReel)">
              Book now <AppIcon name="arrow" :size="18" />
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.ig-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 8px;
  max-width: 1100px;
  margin: 0 auto;
  padding: 110px 20px 40px;
}

.ig-tile {
  position: relative;
  overflow: hidden;
  border-radius: 6px;
  background: #000;
  line-height: 0;
}

.ig-media {
  display: block;
  width: 100%;
  height: auto;
  object-fit: contain;
  background: #000;
}

.ig-open {
  position: absolute;
  inset: 0;
  display: grid;
  place-items: center;
  border: 0;
  color: #fff;
  background: rgba(0, 0, 0, .18);
  opacity: 0;
  cursor: pointer;
  transition: opacity .2s ease;
}

.ig-meta {
  position: absolute;
  inset: auto 0 0;
  padding: 26px 12px 10px;
  color: #fff;
  line-height: 1.35;
  text-align: left;
  background: linear-gradient(transparent, rgba(0, 0, 0, .85));
  opacity: 0;
  transition: opacity .2s ease;
  pointer-events: none;
}

.ig-meta b { display: block; font-size: 14px; }
.ig-meta small { color: rgba(255, 255, 255, .75); font-size: 11px; }

.ig-tile:hover .ig-open,
.ig-tile:hover .ig-meta,
.ig-tile:focus-within .ig-open,
.ig-tile:focus-within .ig-meta { opacity: 1; }

.ig-modal {
  position: fixed;
  inset: 0;
  z-index: 1000;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24px;
  background: rgba(0, 0, 0, .94);
}

.ig-close {
  position: absolute;
  top: 20px;
  right: 24px;
  display: grid;
  place-items: center;
  width: 44px;
  height: 44px;
  border: 1px solid rgba(255, 255, 255, .35);
  border-radius: 50%;
  color: #fff;
  background: rgba(255, 255, 255, .08);
  cursor: pointer;
}

.ig-close:hover { background: rgba(255, 255, 255, .2); }

.ig-player {
  display: flex;
  flex-direction: column;
  gap: 16px;
  width: min(100%, 900px);
}

.ig-player-video {
  display: block;
  width: 100%;
  max-height: 76vh;
  border-radius: 12px;
  background: #000;
  object-fit: contain;
}

.ig-player-bottom {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  color: #fff;
}

.ig-player-meta b { display: block; font-family: var(--font-display); font-size: 20px; }
.ig-player-meta small { color: rgba(255, 255, 255, .65); }
.ig-player-actions { display: flex; align-items: center; gap: 10px; }
.ig-player-actions .btn-outline { color: #fff; border-color: rgba(255, 255, 255, .5); }

@media (max-width: 900px) {
  .ig-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); }
}

@media (max-width: 640px) {
  .ig-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 4px; padding: 110px 8px 24px; }
  .ig-modal { padding: 16px; }
  .ig-player-bottom { flex-direction: column; align-items: stretch; }
  .ig-player-actions { justify-content: stretch; }
  .ig-player-actions .btn { flex: 1; justify-content: center; }
}
</style>
