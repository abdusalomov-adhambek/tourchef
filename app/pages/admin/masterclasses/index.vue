<script setup lang="ts">
import { useApi } from '~/composables/useApi'
import { useToast } from '~/composables/useToast'
import { money, apPillClass, humanStatus, imageOr, duration, dateLabel } from '~/composables/useFormat'

definePageMeta({ layout: 'admin' })

const api = useApi()
const toast = useToast()
const items = ref<any[]>([])
const dishes = ref<any[]>([])
const allChefs = ref<any[]>([])
const regionChefs = ref<any[]>([])
const restaurants = ref<any[]>([])
const loading = ref(true)
const error = ref('')
const filter = ref('')
const q = ref('')
const busyId = ref<any>(null)
const showForm = ref(false)
const busy = ref(false)
const uploading = ref(false)
const fileInput = ref<HTMLInputElement | null>(null)

// "View" opens an in-page detail modal instead of leaving the admin area.
const viewing = ref<any | null>(null)
const allSessions = ref<any[]>([])
const sessionsLoaded = ref(false)
const sessionsLoading = ref(false)

const emptyForm = () => ({
  title: '',
  description: '',
  dish_id: '',
  region: '',
  chef_id: '',
  duration_min: 180,
  price: 180000,
  difficulty: 'beginner',
  language: 'uz',
  image_url: '',
  restaurant_id: '',
  status: 'approved'
})
const form = reactive(emptyForm())

// Hududlar tasdiqlangan oilalar va taomlarning regionlaridan olinadi.
const regions = computed(() =>
  Array.from(new Set([
    ...allChefs.value.map((c) => c.city).filter(Boolean),
    ...dishes.value.map((d) => d.region).filter(Boolean)
  ])).sort()
)

const filtered = computed(() =>
  items.value.filter((m) => {
    const matchFilter = !filter.value || m.status === filter.value
    const matchQ = !q.value || `${m.title} ${m.city || ''}`.toLowerCase().includes(q.value.toLowerCase())
    return matchFilter && matchQ
  })
)

async function loadRegionChefs(city: string) {
  if (!city) {
    regionChefs.value = []
    return
  }
  try {
    const data = await api.get<any[]>('/admin/families', { city })
    regionChefs.value = Array.isArray(data) ? data : []
  } catch (e: any) {
    regionChefs.value = []
    toast.error(e?.message || 'Oilalar yuklanmadi.')
  }
}

watch(() => form.region, (city) => {
  form.chef_id = ''
  void loadRegionChefs(city)
})

async function load() {
  loading.value = true
  const [mc, dishData, chefs, restData] = await Promise.allSettled([
    api.get<any[]>('/admin/masterclasses'),
    api.get<any>('/dishes', { limit: 100 }),
    api.get<any[]>('/admin/families'),
    api.get<any[]>('/restaurants')
  ])
  if (mc.status === 'fulfilled') items.value = Array.isArray(mc.value) ? mc.value : []
  if (dishData.status === 'fulfilled') dishes.value = Array.isArray(dishData.value?.items) ? dishData.value.items : []
  if (chefs.status === 'fulfilled') allChefs.value = Array.isArray(chefs.value) ? chefs.value : []
  if (restData.status === 'fulfilled') restaurants.value = Array.isArray(restData.value) ? restData.value : []
  if (form.region) await loadRegionChefs(form.region)
  loading.value = false
}

function openCreate() {
  Object.assign(form, emptyForm())
  error.value = ''
  showForm.value = true
}

async function uploadImage(event: Event) {
  const file = (event.target as HTMLInputElement).files?.[0]
  if (!file) return
  if (!file.type.startsWith('image/') || file.size > 10 * 1024 * 1024) {
    toast.error('JPG/PNG/WebP/GIF rasm, max 10 MB.')
    return
  }
  uploading.value = true
  try {
    const res = await api.upload<any>('/upload', file)
    form.image_url = res.url
    toast.success('Rasm yuklandi.')
  } catch (e: any) {
    toast.error(e?.message || 'Rasm yuklanmadi.')
  } finally {
    uploading.value = false
    if (fileInput.value) fileInput.value.value = ''
  }
}

async function create() {
  error.value = ''
  if (!form.title.trim() || !form.description.trim() || !form.dish_id || !form.region || !form.chef_id) {
    error.value = 'Sarlavha, tavsif, taom, hudud va oila majburiy.'
    toast.error(error.value)
    return
  }
  if (Number(form.price) <= 0 || Number(form.duration_min) <= 0) {
    error.value = 'Narx va davomiylik 0 dan katta bo\'lishi kerak.'
    toast.error(error.value)
    return
  }
  busy.value = true
  try {
    const created = await api.post<any>('/masterclasses', {
      title: form.title.trim(),
      description: form.description.trim(),
      dish_id: form.dish_id,
      city: form.region,
      chef_id: form.chef_id,
      duration_min: Number(form.duration_min),
      price: Number(form.price),
      difficulty: form.difficulty,
      language: form.language,
      image_url: form.image_url || undefined,
      restaurant_id: form.restaurant_id || undefined,
      status: 'approved'
    })
    items.value = [created, ...items.value]
    toast.success('Masterclass yaratildi va saytda ko\'rinadi.')
    showForm.value = false
    Object.assign(form, emptyForm())
  } catch (e: any) {
    error.value = e?.message || 'Masterclass yaratilmadi.'
    toast.error(error.value)
  } finally {
    busy.value = false
  }
}

async function setStatus(mc: any, status: string) {
  if (!window.confirm(`Set "${mc.title}" to ${status}?`)) return
  busyId.value = mc.id
  try {
    const updated = await api.patch<any>(`/admin/masterclasses/${mc.id}`, { status })
    items.value = items.value.map((m) => (m.id === mc.id ? { ...m, ...updated } : m))
    toast.success(`Masterclass ${status}.`)
  } catch (e: any) {
    toast.error(e?.message || 'Could not update masterclass.')
  } finally {
    busyId.value = null
  }
}

async function openView(mc: any) {
  viewing.value = mc
  if (sessionsLoaded.value) return
  sessionsLoaded.value = true
  sessionsLoading.value = true
  try {
    const data = await api.get<any[]>('/admin/sessions')
    allSessions.value = Array.isArray(data) ? data : []
  } catch (e: any) {
    toast.error(e?.message || 'Seanslar yuklanmadi.')
  } finally {
    sessionsLoading.value = false
  }
}

function closeView() {
  viewing.value = null
}

const viewSessions = computed(() => {
  if (!viewing.value) return []
  const id = String(viewing.value.id)
  return allSessions.value.filter((s) => String(s.masterclass_id) === id)
})

function onKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape' && viewing.value) closeView()
}

onMounted(() => {
  void load()
  window.addEventListener('keydown', onKeydown)
})
onUnmounted(() => window.removeEventListener('keydown', onKeydown))
</script>

<template>
  <div class="agrid">
    <div v-if="showForm" class="acard">
      <div class="acard-h">
        <h3>Create a masterclass</h3>
        <button class="abtn" type="button" @click="showForm = false"><AppIcon name="close" :size="15" /> Close</button>
      </div>
      <div class="acard-b">
        <AppAlert v-if="error" kind="error">{{ error }}</AppAlert>
        <form @submit.prevent="create">
          <div class="form-grid">
            <div class="field full"><label for="mc-title">Title</label><input id="mc-title" v-model="form.title" class="inp" placeholder="The Art of Uzbek Plov" required /></div>
            <div class="field full"><label for="mc-desc">Description</label><textarea id="mc-desc" v-model="form.description" class="textarea" rows="3" placeholder="What will guests learn?" required /></div>
            <div class="field">
              <label for="mc-region">Hudud (region)</label>
              <select id="mc-region" v-model="form.region" class="select" required>
                <option value="" disabled>Tanlang…</option>
                <option v-for="r in regions" :key="r" :value="r">{{ r }}</option>
              </select>
            </div>
            <div class="field">
              <label for="mc-chef">Oila <span class="muted">(faqat shu hududagi)</span></label>
              <select id="mc-chef" v-model="form.chef_id" class="select" :disabled="!form.region" required>
                <option value="" disabled>{{ form.region ? 'Tanlang…' : 'Avval hududni tanlang' }}</option>
                <option v-for="chef in regionChefs" :key="chef.id" :value="chef.id">{{ chef.display_name || chef.email || chef.id }} · {{ chef.status === 'approved' ? 'tasdiqlangan' : chef.status }}</option>
              </select>
              <span v-if="form.region && !regionChefs.length" class="hint">Bu hududda oila yo'q.</span>
            </div>
            <div class="field">
              <label for="mc-dish">Taom</label>
              <select id="mc-dish" v-model="form.dish_id" class="select" required>
                <option value="" disabled>Tanlang…</option>
                <option v-for="d in dishes" :key="d.id" :value="d.id">{{ d.name }}</option>
              </select>
            </div>
            <div class="field">
              <label for="mc-rest">Restoran <span class="muted">(ixtiyoriy)</span></label>
              <select id="mc-rest" v-model="form.restaurant_id" class="select">
                <option value="">—</option>
                <option v-for="r in restaurants" :key="r.id" :value="r.id">{{ r.name }} · {{ r.city }}</option>
              </select>
            </div>
            <div class="field"><label for="mc-duration">Davomiylik (daqiqa)</label><input id="mc-duration" v-model.number="form.duration_min" class="inp" type="number" min="30" required /></div>
            <div class="field"><label for="mc-price">Narx (so'm / kishi)</label><input id="mc-price" v-model.number="form.price" class="inp" type="number" min="1000" required /><span class="hint">Saytda dollarda ko'rsatiladi (1 USD = 12 000 so'm).</span></div>
            <div class="field"><label for="mc-diff">Qiyinchilik</label><select id="mc-diff" v-model="form.difficulty" class="select"><option value="beginner">Beginner</option><option value="intermediate">Intermediate</option><option value="advanced">Advanced</option></select></div>
            <div class="field"><label for="mc-lang">Til</label><select id="mc-lang" v-model="form.language" class="select"><option value="uz">Uzbek</option><option value="ru">Russian</option><option value="en">English</option></select></div>
            <div class="field full">
              <label>Rasm <span class="muted">(ixtiyoriy)</span></label>
              <div class="upload-zone" :class="{ 'has-image': !!form.image_url }" @click.self="fileInput?.click()">
                <div class="upload-preview">
                  <img v-if="form.image_url" :src="imageOr(form.image_url)" alt="Preview" />
                  <span v-else class="upload-icon"><AppIcon name="book" :size="24" /></span>
                </div>
                <div class="upload-info">
                  <b>{{ form.image_url ? 'Rasm tayyor' : 'Rasm tanlang' }}</b>
                  <small>JPG, PNG, WebP · max 10 MB</small>
                  <div class="upload-actions">
                    <button class="abtn primary" type="button" :disabled="uploading" @click="fileInput?.click()">
                      <AppIcon name="folder" :size="15" /> {{ uploading ? 'Yuklanayapti…' : 'Rasm tanlash' }}
                    </button>
                    <button v-if="form.image_url" class="abtn danger" type="button" @click="form.image_url = ''">O'chirish</button>
                  </div>
                </div>
                <input ref="fileInput" class="upload-input" type="file" accept="image/jpeg,image/png,image/webp,image/gif" :disabled="uploading" @change="uploadImage" />
              </div>
            </div>
          </div>
          <div class="form-actions">
            <button class="abtn" type="button" @click="showForm = false">Cancel</button>
            <button class="abtn primary" type="submit" :disabled="busy">{{ busy ? 'Creating…' : 'Create masterclass' }}</button>
          </div>
        </form>
      </div>
    </div>

    <div class="acard">
      <div class="acard-h">
        <h3>Masterclass moderation <span class="muted">({{ filtered.length }})</span></h3>
        <div class="toolbar">
          <div class="afinput"><AppIcon name="search" :size="15" /><input v-model="q" type="search" placeholder="Search title or city…" /></div>
          <div class="aseg">
            <button v-for="f in [{ v: '', l: 'All' }, { v: 'pending', l: 'Pending' }, { v: 'approved', l: 'Active' }, { v: 'suspended', l: 'Hidden' }]" :key="f.v" type="button" :class="{ on: filter === f.v }" @click="filter = f.v">{{ f.l }}</button>
          </div>
          <button class="abtn primary lg" type="button" @click="openCreate"><AppIcon name="plus" :size="15" /> Create masterclass</button>
        </div>
      </div>
      <AppAlert v-if="error && !showForm" kind="error" class="mx-16">{{ error }}</AppAlert>
      <div class="atbl-wrap">
        <table class="atbl" v-if="filtered.length">
          <thead><tr><th>Experience</th><th>City</th><th>Price</th><th>Family</th><th>Status</th><th>Actions</th></tr></thead>
          <tbody>
            <tr v-for="mc in filtered" :key="mc.id">
              <td><div class="who"><img :src="imageOr(mc.image_url || mc.dish?.image_url)" alt="" style="width: 34px; height: 34px; border-radius: 8px; object-fit: cover" /><div><b>{{ mc.title }}</b><small>{{ mc.dish?.name || 'Uzbek cuisine' }} · {{ duration(mc.duration_min) }}</small></div></div></td>
              <td>{{ mc.city || '—' }}</td>
              <td>{{ money(mc.price) }}</td>
              <td>{{ mc.chef?.display_name || '—' }}</td>
              <td><span :class="apPillClass(mc.status)"><i class="dot" /> {{ humanStatus(mc.status) }}</span></td>
              <td>
                <button v-if="mc.status !== 'approved'" class="abtn primary" type="button" :disabled="busyId === mc.id" @click="setStatus(mc, 'approved')">Approve</button>
                <button v-if="mc.status !== 'suspended'" class="abtn danger" type="button" :disabled="busyId === mc.id" @click="setStatus(mc, 'suspended')">Hide</button>
                <NuxtLink :to="`/admin/masterclasses/${mc.slug || mc.id}`" class="abtn">View</NuxtLink>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
      <LoadingState v-if="loading" :cards="1" label="Loading masterclasses…" />
      <EmptyState v-else-if="!filtered.length" title="No masterclasses found" message="Create a masterclass or wait for family submissions." icon="book" />
    </div>

    <!-- Masterclass detail (stays inside the admin area) -->
    <div v-if="viewing" class="mc-modal-back" @click.self="closeView">
      <div class="mc-modal acard" role="dialog" aria-modal="true" aria-label="Masterclass details">
        <div class="mc-modal-head">
          <div>
            <h3>{{ viewing.title }}</h3>
            <span :class="apPillClass(viewing.status)"><i class="dot" /> {{ humanStatus(viewing.status) }}</span>
          </div>
          <button class="abtn" type="button" aria-label="Close" @click="closeView"><AppIcon name="close" :size="15" /></button>
        </div>
        <div class="mc-modal-body">
          <img class="mc-modal-img" :src="imageOr(viewing.image_url || viewing.dish?.image_url)" alt="" />
          <p class="mc-modal-desc">{{ viewing.description || 'Tavsif kiritilmagan.' }}</p>

          <div class="mc-detail-grid">
            <div><span>Shahar</span><b>{{ viewing.city || '—' }}</b></div>
            <div><span>Narx</span><b>{{ money(viewing.price) }}</b></div>
            <div><span>Davomiylik</span><b>{{ duration(viewing.duration_min) }}</b></div>
            <div><span>Qiyinchilik</span><b>{{ humanStatus(viewing.difficulty) }}</b></div>
            <div><span>Til</span><b>{{ humanStatus(viewing.language) }}</b></div>
            <div><span>Taom</span><b>{{ viewing.dish?.name || viewing.dish_name || '—' }}</b></div>
            <div><span>Oila</span><b>{{ viewing.chef?.display_name || viewing.chef_name || '—' }}</b></div>
            <div><span>Restoran</span><b>{{ viewing.restaurant?.name || viewing.restaurant_name || '—' }}</b></div>
            <div><span>Reyting</span><b>{{ viewing.rating ?? '—' }} ({{ viewing.reviews_count || 0 }} sharh)</b></div>
            <div><span>Slug</span><b>{{ viewing.slug || '—' }}</b></div>
            <div><span>ID</span><b>{{ viewing.id }}</b></div>
            <div><span>Yaratilgan</span><b>{{ dateLabel(viewing.created_at) }}</b></div>
          </div>

          <div class="mc-modal-section">
            <h4>Seanslar <span class="muted">({{ viewSessions.length }})</span></h4>
            <LoadingState v-if="sessionsLoading" :cards="1" label="Loading sessions…" />
            <div v-else-if="viewSessions.length" class="mc-session">
              <div v-for="s in viewSessions" :key="s.id">
                <b>{{ dateLabel(s.starts_at) }}</b>
                <span>{{ s.booked || 0 }} / {{ s.capacity || 0 }} o'rin · {{ money(s.price) }}</span>
                <span :class="apPillClass(s.status)"><i class="dot" /> {{ humanStatus(s.status) }}</span>
              </div>
            </div>
            <p v-else class="muted">Bu masterclass uchun seans yo'q.</p>
          </div>
        </div>
        <div class="mc-modal-foot">
          <button class="abtn primary" type="button" @click="closeView">Yopish</button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.mc-modal-back { position: fixed; inset: 0; z-index: 1200; display: grid; place-items: center; padding: 20px; background: rgba(20, 24, 19, .58); backdrop-filter: blur(3px); }
.mc-modal { width: min(720px, 100%); max-height: min(88vh, 900px); display: flex; flex-direction: column; overflow: hidden; animation: fadeUp .2s ease; }
.mc-modal-head { display: flex; align-items: flex-start; gap: 12px; padding: 15px 16px; border-bottom: 1px solid var(--a-line); }
.mc-modal-head > div { flex: 1; min-width: 0; display: flex; flex-direction: column; align-items: flex-start; gap: 7px; }
.mc-modal-head h3 { margin: 0; font-size: 16px; font-weight: 600; }
.mc-modal-body { padding: 16px; overflow-y: auto; display: grid; gap: 14px; }
.mc-modal-img { width: 100%; height: 210px; object-fit: cover; border-radius: 10px; background: var(--a-soft); }
.mc-modal-desc { margin: 0; font-size: 13px; line-height: 1.6; color: var(--a-fg); white-space: pre-line; }
.mc-detail-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 10px 16px; }
.mc-detail-grid > div { display: flex; flex-direction: column; gap: 3px; min-width: 0; }
.mc-detail-grid span { color: var(--a-muted); font-size: 11.5px; }
.mc-detail-grid b { font-size: 12.5px; font-weight: 600; word-break: break-word; }
.mc-modal-section h4 { margin: 0 0 8px; font-size: 13px; font-weight: 600; }
.mc-session { display: grid; gap: 7px; }
.mc-session > div { display: flex; gap: 10px; align-items: center; flex-wrap: wrap; border: 1px solid var(--a-line); border-radius: 9px; padding: 8px 10px; font-size: 12px; }
.mc-session > div span.muted, .mc-session > div span:not([class]) { color: var(--a-muted); }
.mc-modal-foot { display: flex; justify-content: flex-end; padding: 12px 16px; border-top: 1px solid var(--a-line); }
@media (max-width: 640px) { .mc-detail-grid { grid-template-columns: 1fr; } .mc-modal-back { padding: 10px; } }
</style>
