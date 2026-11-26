<script setup lang="ts">
import { useApi } from '~/composables/useApi'
import { useToast } from '~/composables/useToast'
import { dateLabel, imageOr, apPillClass, humanStatus } from '~/composables/useFormat'

definePageMeta({ layout: 'admin' })

const api = useApi()
const toast = useToast()
const dishes = ref<any[]>([])
const loading = ref(true)
const busy = ref(false)
const uploading = ref(false)
const error = ref('')
const q = ref('')
const editing = ref<any>(null)
const showForm = ref(false)

const emptyForm = () => ({
  name: '',
  category: 'Main',
  region: 'Tashkent',
  description: '',
  ingredients: '',
  image_url: '',
  multiimage: [] as string[],
  status: 'active'
})
const form = reactive(emptyForm())

const filtered = computed(() =>
  dishes.value.filter((d) => !q.value || `${d.name} ${d.region || ''} ${d.category || ''}`.toLowerCase().includes(q.value.toLowerCase()))
)

async function load() {
  loading.value = true
  try {
    const data = await api.get<any>('/dishes', { limit: 200 })
    dishes.value = Array.isArray(data?.items) ? data.items : []
  } catch (e: any) {
    error.value = e?.message || 'Could not load dishes.'
  } finally {
    loading.value = false
  }
}

function openCreate() {
  editing.value = null
  Object.assign(form, emptyForm())
  showForm.value = true
}

function openEdit(dish: any) {
  editing.value = dish
  Object.assign(form, {
    name: dish.name || '',
    category: dish.category || 'Main',
    region: dish.region || 'Tashkent',
    description: dish.description || '',
    ingredients: Array.isArray(dish.ingredients) ? dish.ingredients.join(', ') : String(dish.ingredients || ''),
    image_url: dish.image_url || '',
    multiimage: Array.isArray(dish.multiimage) ? [...dish.multiimage] : (dish.image_url ? [dish.image_url] : []),
    status: dish.status || 'active'
  })
  showForm.value = true
}

const fileInput = ref<HTMLInputElement | null>(null)
const dragging = ref(false)

function pickFile() {
  fileInput.value?.click()
}

async function uploadFiles(files?: File[] | FileList | null) {
  const list = Array.from(files || [])
  if (!list.length) return
  if (list.some((f) => !f.type.startsWith('image/'))) {
    toast.error('Faqat rasm fayllari tanlang (JPG, PNG, WebP, GIF).')
    return
  }
  if (list.some((f) => f.size > 10 * 1024 * 1024)) {
    toast.error('Har bir rasm 10 MB dan kichik bo\'lishi kerak.')
    return
  }
  uploading.value = true
  try {
    const result = await api.uploadMany<any>('/upload', list)
    form.multiimage = [...form.multiimage, ...(Array.isArray(result?.urls) ? result.urls : [result.url])]
    form.image_url = form.multiimage[0] || ''
    toast.success(`${list.length} ta rasm yuklandi. Saqlagandan keyin saytda ko'rinadi.`)
  } catch (e: any) {
    toast.error(e?.message || 'Rasmlar yuklanmadi.')
  } finally {
    uploading.value = false
    if (fileInput.value) fileInput.value.value = ''
  }
}

function removeImage(index: number) {
  form.multiimage.splice(index, 1)
  form.image_url = form.multiimage[0] || ''
}

function uploadImage(event: Event) {
  void uploadFiles((event.target as HTMLInputElement).files)
}

function onDrop(event: DragEvent) {
  dragging.value = false
  void uploadFiles(event.dataTransfer?.files)
}

async function save() {
  error.value = ''
  if (!form.name.trim() || !form.description.trim()) {
    error.value = 'Name and description are required.'
    toast.error(error.value)
    return
  }
  busy.value = true
  try {
    const payload = {
      name: form.name.trim(),
      category: form.category.trim() || 'Main',
      region: form.region.trim() || 'Uzbekistan',
      description: form.description.trim(),
      ingredients: form.ingredients
        .split(/[,;]\s*/)
        .map((s) => s.trim())
        .filter(Boolean),
      image_url: form.multiimage[0] || '',
      multiimage: form.multiimage,
      status: form.status
    }
    if (editing.value) {
      const updated = await api.patch<any>(`/dishes/${editing.value.id}`, payload)
      dishes.value = dishes.value.map((d) => (d.id === editing.value.id ? { ...d, ...updated } : d))
      toast.success('Dish updated — changes are live across TourChef.')
    } else {
      const created = await api.post<any>('/dishes', payload)
      dishes.value = [created, ...dishes.value]
      toast.success('Dish added — it now appears on the landing page and family catalog.')
    }
    showForm.value = false
  } catch (e: any) {
    error.value = e?.message || 'Could not save dish.'
    toast.error(error.value)
  } finally {
    busy.value = false
  }
}

async function remove(dish: any) {
  if (!window.confirm(`Delete "${dish.name}"? This cannot be undone.`)) return
  try {
    await api.del(`/dishes/${dish.id}`)
    dishes.value = dishes.value.filter((d) => d.id !== dish.id)
    toast.success('Dish deleted.')
  } catch (e: any) {
    toast.error(e?.message || 'Could not delete dish.')
  }
}

async function toggleStatus(dish: any) {
  const next = dish.status === 'active' ? 'hidden' : 'active'
  try {
    const updated = await api.patch<any>(`/dishes/${dish.id}`, { status: next })
    dishes.value = dishes.value.map((d) => (d.id === dish.id ? { ...d, ...updated } : d))
    toast.success(`Dish ${next}.`)
  } catch (e: any) {
    toast.error(e?.message || 'Could not update status.')
  }
}

onMounted(load)
</script>

<template>
  <div class="agrid">
    <div v-if="showForm" class="acard">
      <div class="acard-h">
        <h3>{{ editing ? 'Edit dish' : 'Add a new dish' }}</h3>
        <button class="abtn" type="button" @click="showForm = false"><AppIcon name="close" :size="15" /> Close</button>
      </div>
      <div class="acard-b">
        <AppAlert v-if="error" kind="error">{{ error }}</AppAlert>
        <form @submit.prevent="save">
          <div class="form-grid">
            <div class="field"><label for="name">Name</label><input id="name" v-model="form.name" class="inp" placeholder="Plov" required /></div>
            <div class="field"><label for="category">Category</label><input id="category" v-model="form.category" class="inp" placeholder="Main course" required /></div>
            <div class="field"><label for="region">Origin region</label><input id="region" v-model="form.region" class="inp" placeholder="Tashkent" required /></div>
            <div class="field"><label for="status">Status</label><select id="status" v-model="form.status" class="select"><option value="active">Active — visible everywhere</option><option value="hidden">Hidden — not visible publicly</option></select></div>
            <div class="field full"><label for="desc">Description</label><textarea id="desc" v-model="form.description" class="textarea" rows="3" placeholder="Rice, meat and carrots slow-cooked in a cauldron…" required /></div>
            <div class="field full"><label for="ingredients">Ingredients <span class="muted">(comma separated)</span></label><input id="ingredients" v-model="form.ingredients" class="inp" placeholder="Rice, lamb, carrots, onions" /></div>
            <div class="field full">
              <label>Dish image</label>
              <div
                class="upload-zone"
                :class="{ 'is-dragging': dragging, 'has-image': form.multiimage.length > 0 }"
                @dragover.prevent="dragging = true"
                @dragleave.prevent="dragging = false"
                @drop.prevent="onDrop"
                @click.self="pickFile"
              >
                <div v-if="form.multiimage.length" class="upload-thumbs">
                  <div v-for="(src, i) in form.multiimage" :key="src + i" class="upload-thumb">
                    <img :src="imageOr(src)" :alt="`Dish image ${i + 1}`" />
                    <button type="button" :aria-label="`Remove image ${i + 1}`" @click="removeImage(i)">
                      <AppIcon name="close" :size="12" />
                    </button>
                    <span v-if="i === 0" class="upload-primary-badge">Asosiy</span>
                  </div>
                </div>
                <div v-else class="upload-preview">
                  <span class="upload-icon"><AppIcon name="dish" :size="26" /></span>
                </div>
                <div class="upload-info">
                  <b>{{ form.multiimage.length ? `${form.multiimage.length} ta rasm` : 'Raslarni shu yerga tashlang' }}</b>
                  <small>{{ form.multiimage.length ? 'Birinchi rasm asosiy sifatida ishlatiladi' : 'Bir nechta rasm tanlash mumkin' }}</small>
                  <div class="upload-actions">
                    <button class="abtn primary" type="button" :disabled="uploading" @click="pickFile">
                      <AppIcon name="folder" :size="15" /> {{ uploading ? 'Yuklanayapti…' : 'Rasm tanlash' }}
                    </button>
                    <button v-if="form.multiimage.length" class="abtn danger" type="button" @click="form.multiimage = []; form.image_url = ''">
                      <AppIcon name="trash" :size="14" /> Hammasini o'chirish
                    </button>
                  </div>
                </div>
                <input ref="fileInput" class="upload-input" type="file" multiple accept="image/jpeg,image/png,image/webp,image/gif" :disabled="uploading" @change="uploadImage" />
              </div>
              <span class="hint">JPG, PNG, WebP yoki GIF · har biri max 10 MB</span>
            </div>
          </div>
          <div class="form-actions">
            <button class="abtn" type="button" @click="showForm = false">Cancel</button>
            <button class="abtn primary" type="submit" :disabled="busy">{{ busy ? 'Saving…' : editing ? 'Save changes' : 'Add dish' }}</button>
          </div>
        </form>
      </div>
    </div>

    <div class="acard">
      <div class="acard-h">
        <h3>Dishes & cuisines <span class="muted">({{ filtered.length }})</span></h3>
        <div class="toolbar">
          <div class="afinput"><AppIcon name="search" :size="15" /><input v-model="q" type="search" placeholder="Search dishes…" /></div>
          <button class="abtn primary lg" type="button" @click="openCreate"><AppIcon name="plus" :size="15" /> Add dish</button>
        </div>
      </div>
      <AppAlert v-if="error && !showForm" kind="error" class="mx-16">{{ error }}</AppAlert>
      <div class="atbl-wrap">
        <table class="atbl" v-if="filtered.length">
          <thead><tr><th>Dish</th><th>Category</th><th>Region</th><th>Status</th><th>Added</th><th>Actions</th></tr></thead>
          <tbody>
            <tr v-for="dish in filtered" :key="dish.id">
              <td><div class="who"><img :src="imageOr(dish.image_url)" alt="" style="width: 34px; height: 34px; border-radius: 8px; object-fit: cover" /><div><b>{{ dish.name }}</b><small class="truncate" style="max-width: 200px">{{ dish.slug }}</small></div></div></td>
              <td>{{ dish.category || '—' }}</td>
              <td>{{ dish.region || '—' }}</td>
              <td><span :class="apPillClass(dish.status)"><i class="dot" /> {{ humanStatus(dish.status) }}</span></td>
              <td>{{ dish.created_at ? dateLabel(dish.created_at, false) : '—' }}</td>
              <td>
                <button class="abtn" type="button" @click="openEdit(dish)"><AppIcon name="edit" :size="14" /> Edit</button>
                <button class="abtn" type="button" @click="toggleStatus(dish)">{{ dish.status === 'active' ? 'Hide' : 'Activate' }}</button>
                <button class="abtn danger" type="button" @click="remove(dish)"><AppIcon name="trash" :size="14" /> Delete</button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
      <LoadingState v-if="loading" :cards="1" label="Loading dishes…" />
      <EmptyState v-else-if="!filtered.length" title="No dishes found" message="Add a dish with the button above — it will appear on the public site and family catalog." icon="dish" />
    </div>
  </div>
</template>
