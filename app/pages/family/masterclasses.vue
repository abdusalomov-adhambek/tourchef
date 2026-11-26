<script setup lang="ts">
import { useApi } from '~/composables/useApi'
import { useToast } from '~/composables/useToast'
import { money, duration, pillClass, humanStatus, imageOr } from '~/composables/useFormat'

definePageMeta({ layout: 'family' })

const api = useApi()
const toast = useToast()
const items = ref<any[]>([])
const dishes = ref<any[]>([])
const restaurants = ref<any[]>([])
const showForm = ref(false)
const loading = ref(true)
const busy = ref(false)
const error = ref('')

const form = reactive({
  title: '',
  description: '',
  dish_id: '',
  city: '',
  duration_min: 120,
  price: 150000,
  difficulty: 'Beginner',
  language: 'English',
  image_url: '',
  restaurant_id: ''
})

async function load() {
  loading.value = true
  const [mc, d, r] = await Promise.allSettled([
    api.get<any[]>('/family/masterclasses'),
    api.get<any>('/dishes', { limit: 100 }),
    api.get<any[]>('/restaurants')
  ])
  if (mc.status === 'fulfilled') items.value = Array.isArray(mc.value) ? mc.value : []
  if (d.status === 'fulfilled') dishes.value = Array.isArray(d.value?.items) ? d.value.items : []
  if (r.status === 'fulfilled') restaurants.value = Array.isArray(r.value) ? r.value : []
  loading.value = false
}

async function create() {
  error.value = ''
  if (!form.title.trim() || !form.description.trim() || !form.city.trim() || !form.dish_id) {
    error.value = 'Title, description, city and dish are required.'
    toast.error(error.value)
    return
  }
  if (Number(form.price) <= 0 || Number(form.duration_min) <= 0) {
    error.value = 'Price and duration must be greater than zero.'
    toast.error(error.value)
    return
  }
  busy.value = true
  try {
    const payload: Record<string, any> = {
      title: form.title.trim(),
      description: form.description.trim(),
      dish_id: form.dish_id,
      city: form.city.trim(),
      duration_min: Number(form.duration_min),
      price: Number(form.price),
      difficulty: form.difficulty,
      language: form.language
    }
    if (form.image_url.trim()) payload.image_url = form.image_url.trim()
    if (form.restaurant_id) payload.restaurant_id = form.restaurant_id
    const created = await api.post<any>('/masterclasses', payload)
    items.value = [created, ...items.value]
    toast.success('Masterclass created. You can now schedule sessions.')
    showForm.value = false
    Object.assign(form, { title: '', description: '', dish_id: '', city: '', duration_min: 120, price: 150000, difficulty: 'Beginner', language: 'English', image_url: '', restaurant_id: '' })
  } catch (e: any) {
    error.value = e?.message || 'Could not create masterclass.'
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
      <div><h1>My masterclasses</h1><p>Create and manage the experiences you host.</p></div>
      <button class="btn btn-primary btn-sm" type="button" @click="showForm = !showForm">
        <AppIcon :name="showForm ? 'close' : 'plus'" :size="16" /> {{ showForm ? 'Close form' : 'New masterclass' }}
      </button>
    </div>

    <div v-if="showForm" class="panel mb-24">
      <div class="panel-h"><h3>Create a masterclass</h3></div>
      <div class="panel-b">
        <AppAlert v-if="error" kind="error" class="mb-16">{{ error }}</AppAlert>
        <form @submit.prevent="create">
          <div class="form-grid">
            <div class="field full"><label for="title">Title</label><input id="title" v-model="form.title" class="inp" placeholder="The Art of Uzbek Plov" required /></div>
            <div class="field full"><label for="desc">Description</label><textarea id="desc" v-model="form.description" class="textarea" rows="4" required /></div>
            <div class="field"><label for="dish">Dish</label><select id="dish" v-model="form.dish_id" class="select" required><option value="" disabled>Select a dish</option><option v-for="d in dishes" :key="d.id" :value="String(d.id)">{{ d.name }}</option></select><span class="hint">Dishes come from the admin catalog.</span></div>
            <div class="field"><label for="city">City</label><input id="city" v-model="form.city" class="inp" placeholder="Tashkent" required /></div>
            <div class="field"><label for="duration">Duration (minutes)</label><input id="duration" v-model.number="form.duration_min" class="inp" type="number" min="30" step="15" required /></div>
            <div class="field"><label for="price">Price per person (so'm)</label><input id="price" v-model.number="form.price" class="inp" type="number" min="1" required /><span class="hint">Shown on the site in USD (1 USD = 12 000 so'm).</span></div>
            <div class="field"><label for="difficulty">Difficulty</label><select id="difficulty" v-model="form.difficulty" class="select"><option>Beginner</option><option>Intermediate</option><option>Advanced</option></select></div>
            <div class="field"><label for="language">Language</label><select id="language" v-model="form.language" class="select"><option>English</option><option>Uzbek</option><option>Russian</option></select></div>
            <div class="field"><label for="image">Image URL <span class="muted">(optional)</span></label><input id="image" v-model="form.image_url" class="inp" type="url" placeholder="https://…" /></div>
            <div class="field"><label for="restaurant">Restaurant <span class="muted">(optional)</span></label><select id="restaurant" v-model="form.restaurant_id" class="select"><option value="">None</option><option v-for="r in restaurants" :key="r.id" :value="String(r.id)">{{ r.name }}</option></select></div>
          </div>
          <div class="form-actions">
            <button class="btn btn-outline" type="button" @click="showForm = false">Cancel</button>
            <button class="btn btn-primary" type="submit" :disabled="busy" :class="{ 'is-loading': busy }">{{ busy ? 'Creating…' : 'Create masterclass' }}</button>
          </div>
        </form>
      </div>
    </div>

    <LoadingState v-if="loading" :cards="2" label="Loading your masterclasses…" />
    <div v-else-if="items.length" class="grid-cards">
      <div v-for="mc in items" :key="mc.id" class="card">
        <div class="card-media"><img :src="imageOr(mc.image_url || mc.dish?.image_url)" :alt="mc.title" /></div>
        <div class="card-body">
          <div class="row between"><span class="badge neutral">{{ mc.city }}</span><span :class="pillClass(mc.status)">{{ humanStatus(mc.status) }}</span></div>
          <h3 class="card-title">{{ mc.title }}</h3>
          <p class="card-desc line-clamp-2">{{ mc.description }}</p>
          <div class="card-meta"><span><AppIcon name="clock" :size="15" /> {{ duration(mc.duration_min) }}</span><span><AppIcon name="dish" :size="15" /> {{ mc.dish?.name || '—' }}</span></div>
          <div class="card-foot"><span class="price">{{ money(mc.price) }}</span><NuxtLink :to="`/masterclasses/${mc.slug || mc.id}`" class="text-link">View <AppIcon name="arrow" :size="16" /></NuxtLink></div>
        </div>
      </div>
    </div>
    <EmptyState v-else title="No masterclasses yet" message="Use the button above to create your first culinary experience." icon="book" />
  </div>
</template>
