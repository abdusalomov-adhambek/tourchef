<script setup lang="ts">
import { useApi } from '~/composables/useApi'
import { useToast } from '~/composables/useToast'
import { imageOr } from '~/composables/useFormat'

definePageMeta({ layout: 'default' })

const api = useApi()
const router = useRouter()
const toast = useToast()

const search = reactive({ destination: '', dish: '', date: '', guests: 2 })
const loading = ref(true)
const featured = ref<any[]>([])
const dishes = ref<any[]>([])
const reels = ref<any[]>([])

// Reels: Instagram-uslubidagi grid va to'liq ekran video ko'rish
const activeReel = ref<any>(null)
const modalVideo = ref<HTMLVideoElement | null>(null)

function reelBookPath(reel: any): string | null {
  const slug = reel?.masterclass?.slug || reel?.masterclass_slug
  if (slug) return `/masterclasses/${slug}#booking`
  const id = reel?.masterclass_id || reel?.masterclass?.id
  if (id) return `/masterclasses?id=${id}`
  if (reel?.restaurant?.slug) return `/restaurants/${reel.restaurant.slug}`
  return null
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

function bookReel() {
  const path = reelBookPath(activeReel.value)
  closeReel()
  if (path) router.push(path)
  else toast.info('This story is not linked to a bookable experience yet.')
}

async function shareReel() {
  const reel = activeReel.value
  if (!reel) return
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

// Landing sahifasida har ikki bo'limda faqat 3 ta karta ko'rsatiladi.
const featuredCards = computed(() => featured.value.slice(0, 3))
const dishCards = computed(() => dishes.value.slice(0, 3))

// Taste of Uzbekistan — interaktiv xarita
const regions = ref<any[]>([])
const selectedRegion = ref('')
const regionDetail = ref<any>(null)
const regionMasterclasses = ref<any[]>([])
const regionRestaurants = ref<any[]>([])
const regionLoading = ref(false)
const popupOpen = ref(false)

async function loadRegion(slug: string) {
  selectedRegion.value = slug
  regionLoading.value = true
  regionDetail.value = null
  regionMasterclasses.value = []
  regionRestaurants.value = []
  const [detail, mc, rest] = await Promise.allSettled([
    api.get<any>(`/v1/regions/${slug}`),
    api.get<any>(`/v1/regions/${slug}/masterclasses`, { limit: 6 }),
    api.get<any[]>(`/v1/regions/${slug}/restaurants`)
  ])
  if (detail.status === 'fulfilled') regionDetail.value = detail.value
  if (mc.status === 'fulfilled') regionMasterclasses.value = Array.isArray(mc.value?.items) ? mc.value.items : []
  if (rest.status === 'fulfilled') regionRestaurants.value = Array.isArray(rest.value) ? rest.value : []
  regionLoading.value = false
}

async function openRegion(slug: string) {
  popupOpen.value = true
  await loadRegion(slug)
}

function onKeydown(e: KeyboardEvent) {
  if (e.key !== 'Escape') return
  popupOpen.value = false
  if (activeReel.value) closeReel()
}
onMounted(() => window.addEventListener('keydown', onKeydown))
onUnmounted(() => {
  window.removeEventListener('keydown', onKeydown)
  if (import.meta.client) document.body.style.overflow = ''
})

const testimonials = [
  {
    quote: 'Cooking plov with a real master in Tashkent was the highlight of our trip. We still talk about it!',
    name: 'Emily L.',
    meta: 'United Kingdom · Plov masterclass',
    initials: 'EL',
    rating: 5
  },
  {
    quote: 'The Reels made it so easy to find a place we’d never have discovered on our own. Booking took a minute.',
    name: 'Marina K.',
    meta: 'Germany · Somsa in Bukhara',
    initials: 'MK',
    rating: 5
  },
  {
    quote: 'Warm, welcoming and so authentic. Our host family made us feel at home. Highly recommend TourChef.',
    name: 'Daniel A.',
    meta: 'USA · Manti in Samarkand',
    initials: 'DA',
    rating: 5
  }
]

const steps = [
  { title: 'Discover', text: 'Explore restaurants, dishes and culinary experiences across Uzbekistan.' },
  { title: 'Book', text: 'Choose a masterclass, select a session and reserve your place in seconds.' },
  { title: 'Experience', text: 'Visit the family kitchen, learn from a host, and enjoy authentic Uzbek cuisine.' }
]

async function load() {
  loading.value = true
  const [mc, dishData, reelData, regionData] = await Promise.allSettled([
    api.get<any>('/masterclasses', { limit: 3, sort: 'rec' }),
    api.get<any>('/dishes', { limit: 3 }),
    api.get<any[]>('/reels'),
    api.get<any[]>('/v1/regions')
  ])
  if (mc.status === 'fulfilled') featured.value = Array.isArray(mc.value?.items) ? mc.value.items.slice(0, 3) : []
  if (dishData.status === 'fulfilled') dishes.value = Array.isArray(dishData.value?.items) ? dishData.value.items.slice(0, 3) : []
  if (reelData.status === 'fulfilled') reels.value = Array.isArray(reelData.value) ? reelData.value.slice(0, 4) : []
  if (regionData.status === 'fulfilled') regions.value = Array.isArray(regionData.value) ? regionData.value : []
  if (regions.value.length) {
    const preferred = regions.value.find((r) => r.slug === 'tashkent-city') || regions.value[0]
    await loadRegion(preferred.slug)
  }
  loading.value = false
}

function submitSearch() {
  const query: Record<string, string> = {}
  if (search.destination.trim()) query.city = search.destination.trim()
  if (search.dish.trim()) query.q = search.dish.trim()
  if (search.date) query.date = search.date
  if (search.guests) query.guests = String(search.guests)
  router.push({ path: '/masterclasses', query })
}

onMounted(load)
</script>

<template>
  <div>
    <!-- HERO -->
    <section class="hero">
      <div class="container">
        <div class="hero-grid">
          <div class="hero-copy">
            <span class="eyebrow reveal" style="--reveal-delay:60ms">Culinary tourism · Uzbekistan</span>
            <h1 class="hero-title reveal" style="--reveal-delay:120ms">Don't Just Taste Uzbekistan, <em>Be Its Guest.</em></h1>
            <p class="hero-lead reveal" style="--reveal-delay:180ms">
              Experience authentic Uzbek cuisine, learn traditional recipes from local families,
              and explore the culture behind every dish.
            </p>
            <div class="hero-actions reveal" style="--reveal-delay:240ms">
              <NuxtLink to="/masterclasses" class="btn btn-primary btn-lg">
                Explore Masterclasses <AppIcon name="arrow" :size="20" />
              </NuxtLink>
              <NuxtLink to="/reels" class="btn btn-outline btn-lg">
                <AppIcon name="play" :size="20" /> Watch Reels
              </NuxtLink>
            </div>
            <div class="hero-trust reveal" style="--reveal-delay:300ms">
              <div class="t"><AppIcon name="shield" :size="18" /><b>Verified</b> local kitchens</div>
              <div class="t"><AppIcon name="star" :size="18" /><b>4.9/5</b> guest rating</div>
              <div class="t"><AppIcon name="globe" :size="18" /><b>8</b> regions to explore</div>
            </div>
          </div>
          <div class="hero-visual reveal" style="--reveal-delay:150ms">
            <img
              class="hero-photo"
              src="https://thumb.wikimedia.org/wikipedia/commons/thumb/0/04/Uzbek_national_cuisine.jpg/960px-Uzbek_national_cuisine.jpg"
              alt="Traditional Uzbek dishes on a table"
            />
            <img
              class="hero-photo small"
              src="https://thumb.wikimedia.org/wikipedia/commons/thumb/9/90/Shashlik-national_food_of_Uzbekistan.jpg/960px-Shashlik-national_food_of_Uzbekistan.jpg"
              alt="Family kitchen"
            />
            <div class="float-card">
              <span class="ic"><AppIcon name="family" :size="22" /></span>
              <div><b>Cook with a family</b><small>Samarkand · 3 hours</small></div>
            </div>
          </div>
        </div>

        <form class="search-panel reveal" style="margin-top: 52px; --reveal-delay:360ms" @submit.prevent="submitSearch">
          <div class="search-field">
            <AppIcon name="pin" :size="20" />
            <div><label for="dest">Destination</label><input id="dest" v-model="search.destination" type="text" placeholder="City or region" /></div>
          </div>
          <div class="search-field">
            <AppIcon name="dish" :size="20" />
            <div><label for="dish">Dish</label><input id="dish" v-model="search.dish" type="text" placeholder="Plov, manti, somsa…" /></div>
          </div>
          <div class="search-field">
            <AppIcon name="calendar" :size="20" />
            <div><label for="date">Date</label><input id="date" v-model="search.date" type="date" /></div>
          </div>
          <div class="search-field">
            <AppIcon name="users" :size="20" />
            <div><label for="guests">Guests</label><input id="guests" v-model.number="search.guests" type="number" min="1" max="20" /></div>
          </div>
          <button class="btn btn-primary" type="submit">Find Experiences</button>
        </form>
      </div>
    </section>

    <!-- FEATURED MASTERCLASSES -->
    <section class="section">
      <div class="container">
        <div class="section-head reveal">
          <div>
            <span class="eyebrow">Handpicked for you</span>
            <h2 class="section-title">Featured <em>Masterclasses</em></h2>
            <p class="section-sub">Learn from local families in their kitchens — then taste what you have made.</p>
          </div>
          <NuxtLink to="/masterclasses" class="text-link">View all experiences <AppIcon name="arrow" :size="16" /></NuxtLink>
        </div>
        <LoadingState v-if="loading" :cards="3" />
        <div v-else-if="featuredCards.length" class="grid-cards">
          <MasterclassCard
            v-for="(mc, i) in featuredCards"
            class="reveal"
            :style="{ '--reveal-delay': `${i * 70}ms` }"
            :key="mc.id"
            :mc="mc"
            :badge="i === 0 ? 'Popular' : undefined"
          />
        </div>
        <EmptyState v-else title="No masterclasses yet" message="New culinary experiences are being added. Explore our dishes in the meantime." action-to="/masterclasses" action-label="Refresh catalog" />
      </div>
    </section>

    <!-- POPULAR DISHES -->
    <section class="section tight" style="background: var(--surface); border-block: 1px solid var(--border)">
      <div class="container">
        <div class="section-head reveal">
          <div>
            <span class="eyebrow">Taste the culture</span>
            <h2 class="section-title">Popular Uzbek <em>Dishes</em></h2>
          </div>
          <NuxtLink to="/dishes" class="text-link">Explore the food culture <AppIcon name="arrow" :size="16" /></NuxtLink>
        </div>
        <div v-if="dishCards.length" class="grid-cards">
          <DishCard v-for="(dish, i) in dishCards" :key="dish.id" :dish="dish" class="reveal" :style="{ '--reveal-delay': `${i * 60}ms` }" />
        </div>
        <EmptyState v-else-if="!loading" title="The dish catalog is empty" message="Our culinary team is adding traditional Uzbek dishes." action-to="/dishes" action-label="Go to dishes" />
      </div>
    </section>

    <!-- TASTE OF UZBEKISTAN — INTERACTIVE FOOD MAP -->
    <section class="section rf-section">
      <div class="container">
        <div class="section-head reveal">
          <div>
            <span class="eyebrow">Explore by region</span>
            <h2 class="section-title">Taste of <em>Uzbekistan</em></h2>
            <p class="section-sub">Explore Uzbekistan region by region and discover the traditional flavors that make every destination unique.</p>
          </div>
          <NuxtLink to="/regions" class="text-link">All regions <AppIcon name="arrow" :size="16" /></NuxtLink>
        </div>
        <div class="rf-map-full reveal">
          <ClientOnly>
            <UzbekistanFoodMap :regions="regions" :selected="selectedRegion" @select="openRegion" />
            <template #fallback><div class="rf-map-fallback">Xarita yuklanayapti…</div></template>
          </ClientOnly>
        </div>
      </div>
    </section>

    <!-- Mintaqa / taom haqidagi ma'lumotlar popupda -->
    <div v-if="popupOpen" class="rf-modal-back" @click.self="popupOpen = false">
      <div class="rf-modal" role="dialog" aria-modal="true" aria-label="Mintaqa va taom ma'lumotlari">
        <button class="rf-modal-close" type="button" aria-label="Close" @click="popupOpen = false"><AppIcon name="close" :size="18" /></button>
        <RegionalFoodDetail
          :region="regionDetail"
          :masterclasses="regionMasterclasses"
          :restaurants="regionRestaurants"
          :loading="regionLoading"
        />
      </div>
    </div>

    <!-- REELS PREVIEW -->
    <section class="section tight" style="background: var(--reels-bg); color: #fff">
      <div class="container">
        <div class="section-head" style="color: #fff">
          <div>
            <span class="eyebrow" style="color: var(--accent)">TourChef Reels</span>
            <h2 class="section-title" style="color: #fff">Watch. Discover. <em>Experience.</em></h2>
            <p class="section-sub" style="color: rgba(255,255,255,.65)">
              Explore real culinary experiences through videos created by restaurants, families, and travelers.
            </p>
          </div>
          <NuxtLink to="/reels" class="btn btn-primary">Explore All Reels <AppIcon name="arrow" :size="18" /></NuxtLink>
        </div>
        <div v-if="reels.length" class="ig-grid">
          <button
            v-for="(reel, i) in reels"
            :key="reel.id"
            class="ig-tile reveal"
            type="button"
            :style="{ '--reveal-delay': `${i * 70}ms` }"
            :aria-label="`Play ${reel.title || 'reel'}`"
            @click="openReel(reel)"
          >
            <video
              v-if="reel.video_url"
              class="ig-video"
              :src="reel.video_url"
              :poster="imageOr(reel.image_url)"
              muted
              loop
              playsinline
              preload="metadata"
            />
            <img v-else class="ig-video" :src="imageOr(reel.image_url)" :alt="reel.title || 'TourChef reel'" loading="lazy" />
            <span class="ig-play"><AppIcon name="play" :size="26" /></span>
            <span class="ig-label">{{ reel.title || reel.caption || 'Culinary story' }}</span>
          </button>
        </div>
        <div v-else-if="!loading" style="color: rgba(255,255,255,.7)">
          Reels are being prepared — follow our kitchens to be first to watch.
        </div>
      </div>
    </section>

    <!-- REEL FULLSCREEN PLAYER -->
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
            <button class="btn btn-outline btn-light" type="button" @click="shareReel">
              <AppIcon name="share" :size="18" /> Share
            </button>
            <button class="btn btn-primary" type="button" @click="bookReel">
              Book now <AppIcon name="arrow" :size="18" />
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- HOW IT WORKS -->
    <section class="section">
      <div class="container">
        <div class="section-head reveal"><div><span class="eyebrow">Simple by design</span><h2 class="section-title">How TourChef <em>Works</em></h2></div></div>
        <div class="steps">
          <div v-for="(step, i) in steps" :key="step.title" class="step reveal" :style="{ '--reveal-delay': `${i * 80}ms` }">
            <h3>{{ step.title }}</h3>
            <p>{{ step.text }}</p>
            <NuxtLink v-if="i === 0" to="/masterclasses" class="text-link mt-16">Start exploring <AppIcon name="arrow" :size="16" /></NuxtLink>
            <NuxtLink v-else-if="i === 1" to="/client/bookings" class="text-link mt-16">My bookings <AppIcon name="arrow" :size="16" /></NuxtLink>
          </div>
        </div>
      </div>
    </section>

    <!-- ABOUT -->
    <section id="about" class="section tight" style="background: var(--surface); border-block: 1px solid var(--border)">
      <div class="container split">
        <div class="about-visual reveal">
          <img
            src="https://thumb.wikimedia.org/wikipedia/commons/thumb/a/ac/Uzbek_national_dastarkhan_and_wheat_prepared_for_sumalak.jpg/960px-Uzbek_national_dastarkhan_and_wheat_prepared_for_sumalak.jpg"
            alt="Uzbek culinary culture"
          />
        </div>
        <div>
          <span class="eyebrow">Our mission</span>
          <h2 class="section-title reveal">Culture you can <em>taste</em></h2>
          <p class="section-sub">
            TourChef connects travelers with authentic Uzbek culinary culture through local
            restaurants, families and cooking experiences. Every dish tells a story — we help you live it.
          </p>
          <ul class="about-list">
            <li><AppIcon name="check" :size="18" /><span><b>For tourists:</b> discover real kitchens, book with confidence and take memories home.</span></li>
            <li><AppIcon name="check" :size="18" /><span><b>For restaurants:</b> reach guests worldwide and grow your culinary brand.</span></li>
            <li><AppIcon name="check" :size="18" /><span><b>For culture:</b> keep traditions alive, one shared meal at a time.</span></li>
          </ul>
          <NuxtLink to="/about" class="btn btn-outline mt-24">More about TourChef <AppIcon name="arrow" :size="16" /></NuxtLink>
        </div>
      </div>
    </section>

    <!-- PARTNER CTA -->
    <section class="section tight">
      <div class="container">
        <div class="partner-cta reveal">
          <div>
            <span class="eyebrow gold" style="color: var(--gold)">For restaurants</span>
            <h2>Share Your Culinary Experience with the World</h2>
            <p>Join TourChef and introduce your restaurant, traditional recipes and cooking experiences to travelers.</p>
          </div>
           <NuxtLink to="/register?role=chef" class="btn btn-gold btn-lg">
            Become a Family Host <AppIcon name="arrow" :size="18" />
          </NuxtLink>
        </div>
      </div>
    </section>

    <!-- TESTIMONIALS -->
    <section class="section tight" style="background: var(--surface); border-top: 1px solid var(--border)">
      <div class="container">
        <div class="section-head reveal"><div><span class="eyebrow">Guest stories</span><h2 class="section-title">Loved by <em>travelers</em></h2></div></div>
        <div class="testimonial-grid">
          <div v-for="(t, i) in testimonials" :key="t.name" class="testimonial reveal" :style="{ '--reveal-delay': `${i * 80}ms` }">
            <div class="rating">
              <AppIcon v-for="n in t.rating" :key="n" name="star" :size="15" />
            </div>
            <p class="quote">“{{ t.quote }}”</p>
            <div class="who">
              <span class="avatar">{{ t.initials }}</span>
              <div><b>{{ t.name }}</b><small class="muted">{{ t.meta }}</small></div>
            </div>
          </div>
        </div>
      </div>
    </section>
  </div>
</template>

<style scoped>
.ig-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 8px;
  max-width: 900px;
  margin: 0 auto;
}

.ig-tile {
  position: relative;
  aspect-ratio: 1 / 1;
  padding: 0;
  border: 0;
  border-radius: 10px;
  overflow: hidden;
  background: #000;
  cursor: pointer;
}

.ig-video {
  display: block;
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.ig-play {
  position: absolute;
  inset: 0;
  display: grid;
  place-items: center;
  color: #fff;
  background: rgba(0, 0, 0, .28);
  opacity: 0;
  transition: opacity .2s ease;
}

.ig-label {
  position: absolute;
  inset: auto 0 0;
  padding: 14px 12px 10px;
  color: #fff;
  font-size: 13px;
  font-weight: 600;
  text-align: left;
  background: linear-gradient(transparent, rgba(0, 0, 0, .8));
  opacity: 0;
  transition: opacity .2s ease;
}

.ig-tile:hover .ig-play,
.ig-tile:hover .ig-label,
.ig-tile:focus-visible .ig-play,
.ig-tile:focus-visible .ig-label { opacity: 1; }

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
  width: min(100%, 720px);
}

.ig-player-video {
  display: block;
  width: 100%;
  max-height: 74vh;
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

@media (max-width: 640px) {
  .ig-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); }
  .ig-modal { padding: 16px; }
  .ig-player-bottom { flex-direction: column; align-items: stretch; }
  .ig-player-actions { justify-content: stretch; }
  .ig-player-actions .btn { flex: 1; justify-content: center; }
}
</style>
