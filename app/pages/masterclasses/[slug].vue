<script setup lang="ts">
import { useApi } from '~/composables/useApi'
import { useAuth } from '~/composables/useAuth'
import { useToast } from '~/composables/useToast'
import { useFavorites } from '~/composables/useFavorites'
import { money, duration, dateLabel, imageOr, initials } from '~/composables/useFormat'

definePageMeta({ layout: 'default' })

const route = useRoute()
const router = useRouter()
const api = useApi()
const auth = useAuth()
const toast = useToast()
const favorites = useFavorites()

const mc = ref<any>(null)
const sessions = ref<any[]>([])
const reviews = ref<any[]>([])
const loading = ref(true)
const error = ref('')
const selectedSession = ref<any>(null)
const guests = ref(1)
const booking = ref(false)
const booked = ref<any>(null)

const image = computed(() => imageOr(mc.value?.image_url || mc.value?.dish?.image_url))
const spotsLeft = (s: any) => Math.max(0, Number(s?.capacity || 0) - Number(s?.booked || 0))
const total = computed(() => Number((selectedSession.value?.price ?? mc.value?.price) || 0) * Math.max(1, guests.value))

async function load() {
  loading.value = true
  error.value = ''
  try {
    const data = await api.get<any>(`/masterclasses/${route.params.slug}`)
    mc.value = data
    sessions.value = Array.isArray(data?.sessions) ? data.sessions : []
    selectedSession.value = sessions.value.find((s) => spotsLeft(s) > 0) || sessions.value[0] || null
    try {
      const rev = await api.get<any[]>('/reviews', { masterclass_id: data?.id })
      reviews.value = Array.isArray(rev) ? rev : []
    } catch {
      reviews.value = []
    }
  } catch (e: any) {
    error.value = e?.message || 'Masterclass not found.'
    mc.value = null
  } finally {
    loading.value = false
  }
}

async function reserve() {
  if (!selectedSession.value) {
    toast.error('Please choose a session first.')
    return
  }
  if (!auth.token.value) {
    await auth.fetchUser(true)
  }
  if (!auth.token.value) {
    toast.info('Sign in to reserve your place — we will bring you right back.')
    await router.push(`/login?redirect=${encodeURIComponent(route.fullPath)}`)
    return
  }
  booking.value = true
  try {
    const result = await api.post<any>('/bookings', {
      session_id: selectedSession.value.id,
      guests: Math.max(1, Number(guests.value))
    })
    booked.value = result
    toast.success('Your place is reserved!')
    await router.push(`/client/bookings/${result?.id}`)
  } catch (e: any) {
    if (e?.status === 401) {
      auth.logout()
      toast.info('Your session expired. Sign in again to finish booking.')
      await router.push(`/login?redirect=${encodeURIComponent(route.fullPath)}`)
    } else {
      toast.error(e?.message || 'Could not complete the reservation.')
    }
  } finally {
    booking.value = false
  }
}

onMounted(load)
</script>

<template>
  <div>
    <div v-if="loading" class="container" style="padding: 40px 0">
      <div class="skeleton" style="height: 400px; border-radius: 24px" />
      <div class="skeleton mt-24" style="height: 32px; width: 50%" />
    </div>

    <AppAlert v-else-if="error" kind="error" title="We could not open this experience">
      {{ error }}
      <div class="mt-16"><NuxtLink to="/masterclasses" class="btn btn-outline btn-sm">Back to catalog</NuxtLink></div>
    </AppAlert>

    <template v-else-if="mc">
      <div class="container" style="padding-top: 26px">
        <div class="detail-hero">
          <img :src="image" :alt="mc.title" />
          <div class="dh-shade" />
          <div class="dh-copy">
            <span class="badge">{{ mc.city || 'Uzbekistan' }}</span>
            <h1>{{ mc.title }}</h1>
            <p style="opacity: .85">
              {{ mc.restaurant?.name || 'TourChef partner kitchen' }}
              <span v-if="mc.chef?.display_name"> · with {{ mc.chef.display_name }}</span>
            </p>
          </div>
        </div>

        <div class="detail-layout mt-32">
          <div>
            <div class="detail-section">
              <span class="eyebrow">About the experience</span>
              <h2 class="section-title small">What you will learn</h2>
              <p class="prose">{{ mc.description || 'A hands-on Uzbek cooking experience with a local master.' }}</p>
            </div>

            <div class="detail-section">
              <h2 class="section-title small">Experience details</h2>
              <div class="detail-grid-2">
                <div class="info-row"><AppIcon name="dish" :size="18" /><span>Dish</span> {{ mc.dish?.name || 'Traditional Uzbek cuisine' }}</div>
                <div class="info-row"><AppIcon name="clock" :size="18" /><span>Duration</span> {{ duration(mc.duration_min) }}</div>
                <div class="info-row"><AppIcon name="family" :size="18" /><span>Family</span> {{ mc.chef?.display_name || mc.chef?.full_name || 'Local family' }}</div>
                <div class="info-row"><AppIcon name="restaurant" :size="18" /><span>Venue</span> {{ mc.restaurant?.name || 'Partner kitchen' }}</div>
                <div class="info-row"><AppIcon name="globe" :size="18" /><span>Language</span> {{ mc.language || 'Uzbek / Russian / English' }}</div>
                <div class="info-row"><AppIcon name="spark" :size="18" /><span>Level</span> {{ mc.difficulty || 'All levels' }}</div>
              </div>
            </div>

            <div v-if="sessions.length" class="detail-section" id="sessions">
              <h2 class="section-title small">Available sessions</h2>
              <div v-for="s in sessions" :key="s.id" class="session-opt" :class="{ on: selectedSession?.id === s.id }" @click="selectedSession = s">
                <div style="flex: 1">
                  <div class="d">{{ dateLabel(s.starts_at) }}</div>
                  <small>{{ spotsLeft(s) }} seats left · {{ money(s.price || mc.price) }} per person</small>
                </div>
                <span class="badge" :class="spotsLeft(s) > 0 ? '' : 'neutral'">{{ spotsLeft(s) > 0 ? 'Available' : 'Full' }}</span>
              </div>
            </div>
            <EmptyState v-else title="No sessions scheduled" message="The family has not published any dates yet. Check back soon." icon="calendar" />

            <div class="detail-section">
              <h2 class="section-title small">Guest reviews <span class="muted text-sm">({{ reviews.length }})</span></h2>
              <div v-if="reviews.length">
                <div v-for="review in reviews" :key="review.id" class="review">
                  <span class="avatar">{{ initials(review.user?.full_name || review.full_name || 'Guest') }}</span>
                  <div class="rb">
                    <b>{{ review.user?.full_name || review.full_name || 'Guest' }}</b>
                    <div class="rating"><AppIcon v-for="n in Number(review.rating || 0)" :key="n" name="star" :size="14" /></div>
                    <p>{{ review.comment }}</p>
                    <small class="muted">{{ dateLabel(review.created_at, false) }}</small>
                  </div>
                </div>
              </div>
              <EmptyState v-else title="No reviews yet" message="Be the first to share your experience after your session." icon="star" />
            </div>
          </div>

          <aside class="widget" id="booking">
            <div class="row between">
              <h3>Reserve your place</h3>
              <button class="icon-btn" type="button" :class="{ on: favorites.has(mc.id) }" :aria-label="'Save to favorites'" @click="favorites.toggle(mc.id, route.fullPath)">
                <AppIcon name="heart" :size="18" />
              </button>
            </div>
            <div class="w-price">{{ money(selectedSession?.price || mc.price) }} <small>/ person</small></div>

            <div v-if="sessions.length">
              <label class="text-sm fw-600">Choose a session</label>
              <div class="mt-12">
                <div v-for="s in sessions" :key="s.id" class="session-opt" :class="{ on: selectedSession?.id === s.id }" @click="selectedSession = s">
                  <div>
                    <div class="d">{{ dateLabel(s.starts_at) }}</div>
                    <small>{{ spotsLeft(s) }} seats left</small>
                  </div>
                </div>
              </div>
            </div>

            <div class="mt-16">
              <label class="text-sm fw-600">Guests</label>
              <div class="stepper mt-8">
                <button type="button" aria-label="Fewer guests" :disabled="guests <= 1" @click="guests--"><AppIcon name="minus" :size="16" /></button>
                <span>{{ guests }}</span>
                <button type="button" aria-label="More guests" :disabled="guests >= 20" @click="guests++"><AppIcon name="plus" :size="16" /></button>
              </div>
            </div>

            <div class="w-summary">
              <div><span>Price</span><span>{{ money(selectedSession?.price || mc.price) }} × {{ guests }}</span></div>
              <div><span>Payment</span><span>Development mode</span></div>
            </div>
            <div class="w-total"><span>Total</span><span>{{ money(total) }}</span></div>
            <button class="btn btn-primary btn-block btn-lg mt-16" type="button" :disabled="booking || !selectedSession" :class="{ 'is-loading': booking }" @click="reserve">
              {{ booking ? 'Reserving…' : booked ? 'Booked!' : 'Book this experience' }}
            </button>
            <p class="muted text-xs mt-12 center-text">
              Free cancellation up to 24 hours before the session. Payments run in clearly labelled development mode.
            </p>
          </aside>
        </div>
      </div>
    </template>
  </div>
</template>
