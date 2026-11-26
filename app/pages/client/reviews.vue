<script setup lang="ts">
import { useApi } from '~/composables/useApi'
import { useAuth } from '~/composables/useAuth'
import { useToast } from '~/composables/useToast'
import { dateLabel, bookingMc, initials } from '~/composables/useFormat'

definePageMeta({ layout: 'cabinet' })

const api = useApi()
const auth = useAuth()
const toast = useToast()

const form = reactive({ masterclass_id: '', rating: 5, comment: '' })
const bookings = ref<any[]>([])
const reviews = ref<any[]>([])
const loading = ref(true)
const busy = ref(false)
const error = ref('')

const options = computed(() => {
  const seen = new Set<string>()
  return bookings.value
    .map((b) => bookingMc(b))
    .filter((mc) => {
      if (!mc || !mc.id || seen.has(String(mc.id))) return false
      seen.add(String(mc.id))
      return true
    })
})

async function load() {
  loading.value = true
  const [b, r] = await Promise.allSettled([
    api.get<any[]>('/client/bookings'),
    api.get<any[]>('/reviews')
  ])
  if (b.status === 'fulfilled') {
    bookings.value = Array.isArray(b.value) ? b.value : []
    if (!form.masterclass_id && options.value.length) form.masterclass_id = String(options.value[0].id)
  }
  if (r.status === 'fulfilled') {
    const all = Array.isArray(r.value) ? r.value : []
    const mine = all.filter(
      (x) =>
        String(x.user_id ?? x.user?.id ?? '') === String(auth.user.value?.id ?? '') ||
        String(x.email ?? x.user?.email ?? '') === String(auth.user.value?.email ?? '')
    )
    reviews.value = mine.length ? mine : all
  }
  loading.value = false
}

async function submit() {
  error.value = ''
  if (!form.masterclass_id) {
    error.value = 'Choose a masterclass to review.'
    toast.error(error.value)
    return
  }
  if (form.comment.trim().length < 5) {
    error.value = 'Please write at least 5 characters.'
    toast.error(error.value)
    return
  }
  busy.value = true
  try {
    await api.post('/client/reviews', {
      masterclass_id: Number(form.masterclass_id),
      rating: Number(form.rating),
      comment: form.comment.trim()
    })
    form.comment = ''
    toast.success('Thank you! Your review has been published.')
    await load()
  } catch (e: any) {
    error.value = e?.message || 'Could not publish your review.'
    toast.error(error.value)
  } finally {
    busy.value = false
  }
}

onMounted(load)
</script>

<template>
  <div>
    <div class="cab-top">
      <div><h1>My reviews</h1><p>Share what you learned and help other travelers choose.</p></div>
    </div>

    <div class="two-col">
      <div class="panel">
        <div class="panel-h"><h3>Write a review</h3></div>
        <div class="panel-b">
          <AppAlert v-if="error" kind="error" class="mb-16">{{ error }}</AppAlert>
          <form @submit.prevent="submit">
            <div class="field">
              <label for="mc">Masterclass</label>
              <select id="mc" v-model="form.masterclass_id" class="select" required>
                <option value="" disabled>Select an experience</option>
                <option v-for="mc in options" :key="mc.id" :value="String(mc.id)">{{ mc.title }}</option>
              </select>
              <span class="hint">Choose from your booked experiences.</span>
            </div>
            <div class="field mt-16">
              <label>Rating</label>
              <div class="row" style="gap: 4px">
                <button
                  v-for="n in 5" :key="n" type="button"
                  class="icon-btn" :class="{ on: form.rating >= n }"
                  :aria-label="`${n} stars`" @click="form.rating = n"
                ><AppIcon name="star" :size="18" /></button>
                <span class="muted text-sm" style="margin-left: 8px">{{ form.rating }} / 5</span>
              </div>
            </div>
            <div class="field mt-16">
              <label for="comment">Your review</label>
              <textarea id="comment" v-model="form.comment" class="textarea" rows="5" minlength="5" placeholder="What did you enjoy most?" required />
            </div>
            <button class="btn btn-primary btn-block mt-16" type="submit" :disabled="busy || !options.length" :class="{ 'is-loading': busy }">
              {{ busy ? 'Publishing…' : 'Publish review' }}
            </button>
            <p v-if="!options.length" class="muted text-xs center-text mt-8">Book an experience first to leave a review.</p>
          </form>
        </div>
      </div>

      <div class="panel">
        <div class="panel-h"><h3>Recent reviews</h3></div>
        <div v-if="reviews.length" class="panel-b">
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
        <EmptyState v-else title="No reviews yet" message="Your published reviews will appear here." icon="star" />
      </div>
    </div>
  </div>
</template>
