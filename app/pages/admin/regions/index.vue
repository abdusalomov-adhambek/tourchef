<script setup lang="ts">
import { useApi } from '~/composables/useApi'
import { useToast } from '~/composables/useToast'
import { imageOr } from '~/composables/useFormat'

definePageMeta({ layout: 'admin' })
const api = useApi()
const toast = useToast()
const rows = ref<any[]>([])
const loading = ref(true)
const busy = ref(false)
const uploading = ref(false)
const error = ref('')
const showForm = ref(false)
const editing = ref<any>(null)
const fileInput = ref<HTMLInputElement | null>(null)

const emptyForm = () => ({
  name: '', name_uz: '', geo_id: '', city: '', dish_name: '', dish_image_url: '',
  description: '', ingredients: '', cultural_info: '', latitude: 0, longitude: 0, status: 'published'
})
const form = reactive(emptyForm())

async function load() {
  loading.value = true
  try {
    const data = await api.get<any[]>('/admin/regions')
    rows.value = Array.isArray(data) ? data : []
  } catch (e: any) {
    error.value = e?.message || 'Mintaqalar yuklanmadi.'
  } finally {
    loading.value = false
  }
}

function openCreate() {
  editing.value = null
  Object.assign(form, emptyForm())
  error.value = ''
  showForm.value = true
}

function openEdit(r: any) {
  editing.value = r
  Object.assign(form, {
    name: r.name || '', name_uz: r.name_uz || '', geo_id: r.geo_id || '', city: r.city || '',
    dish_name: r.dish_name || '', dish_image_url: r.dish_image_url || '', description: r.description || '',
    ingredients: Array.isArray(r.ingredients) ? r.ingredients.join(', ') : '',
    cultural_info: r.cultural_info || '', latitude: r.latitude || 0, longitude: r.longitude || 0,
    status: r.status || 'published'
  })
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
    form.dish_image_url = res.url
    toast.success('Rasm yuklandi.')
  } catch (e: any) {
    toast.error(e?.message || 'Rasm yuklanmadi.')
  } finally {
    uploading.value = false
    if (fileInput.value) fileInput.value.value = ''
  }
}

async function save() {
  error.value = ''
  if (!form.name.trim() || !form.dish_name.trim() || !form.geo_id.trim()) {
    error.value = 'Nomi, taom nomi va geo_id majburiy.'
    toast.error(error.value)
    return
  }
  busy.value = true
  const payload = {
    name: form.name.trim(), name_uz: form.name_uz.trim(), geo_id: form.geo_id.trim(),
    city: form.city.trim(), dish_name: form.dish_name.trim(), dish_image_url: form.dish_image_url,
    description: form.description.trim(),
    ingredients: form.ingredients.split(/[,;]\s*/).map((s) => s.trim()).filter(Boolean),
    cultural_info: form.cultural_info.trim(), latitude: Number(form.latitude), longitude: Number(form.longitude),
    status: form.status
  }
  try {
    if (editing.value) {
      const updated = await api.patch<any>(`/admin/regions/${editing.value.id}`, payload)
      rows.value = rows.value.map((r) => (r.id === editing.value.id ? { ...r, ...updated } : r))
      toast.success('Mintaqa yangilandi.')
    } else {
      const created = await api.post<any>('/admin/regions', payload)
      rows.value = [created, ...rows.value]
      toast.success('Mintaqa qo\'shildi.')
    }
    showForm.value = false
  } catch (e: any) {
    error.value = e?.message || 'Saqlanmadi.'
    toast.error(error.value)
  } finally {
    busy.value = false
  }
}

async function toggle(r: any) {
  try {
    const updated = await api.patch<any>(`/admin/regions/${r.id}`, { status: r.status === 'published' ? 'hidden' : 'published' })
    rows.value = rows.value.map((x) => (x.id === r.id ? { ...x, ...updated } : x))
    toast.success('Holat o\'zgartirildi.')
  } catch (e: any) {
    toast.error(e?.message || 'Xatolik.')
  }
}

async function remove(r: any) {
  if (!window.confirm(`"${r.name}" o'chirilsinmi?`)) return
  try {
    await api.del(`/admin/regions/${r.id}`)
    rows.value = rows.value.filter((x) => x.id !== r.id)
    toast.success('O\'chirildi.')
  } catch (e: any) {
    toast.error(e?.message || 'Xatolik.')
  }
}

onMounted(load)
</script>

<template>
  <div class="agrid">
    <div v-if="showForm" class="acard">
      <div class="acard-h"><h3>{{ editing ? 'Mintaqani tahrirlash' : 'Yangi mintaqa' }}</h3><button class="abtn" type="button" @click="showForm = false"><AppIcon name="close" :size="15" /> Close</button></div>
      <div class="acard-b">
        <AppAlert v-if="error" kind="error">{{ error }}</AppAlert>
        <form @submit.prevent="save">
          <div class="form-grid">
            <div class="field"><label>Mintaqa nomi</label><input v-model="form.name" class="inp" placeholder="Jizzakh region" required /></div>
            <div class="field"><label>Nomi (o'zbekcha)</label><input v-model="form.name_uz" class="inp" placeholder="Jizzax viloyati" /></div>
            <div class="field"><label>Geo ID <span class="muted">(GeoJSON dagi id)</span></label><input v-model="form.geo_id" class="inp" placeholder="3" required /></div>
            <div class="field"><label>Shahar / hudud <span class="muted">(masterclass filteri)</span></label><input v-model="form.city" class="inp" placeholder="Jizzakh" /></div>
            <div class="field"><label>Taom nomi</label><input v-model="form.dish_name" class="inp" placeholder="Jizzakh Somsa" required /></div>
            <div class="field"><label>Status</label><select v-model="form.status" class="select"><option value="published">Published</option><option value="hidden">Hidden</option></select></div>
            <div class="field full"><label>Taom tavsifi</label><textarea v-model="form.description" class="textarea" rows="3" /></div>
            <div class="field full"><label>Ingredientlar <span class="muted">(vergul bilan)</span></label><input v-model="form.ingredients" class="inp" placeholder="flour, lamb, onion" /></div>
            <div class="field full"><label>Madaniy ma'lumot</label><textarea v-model="form.cultural_info" class="textarea" rows="2" /></div>
            <div class="field"><label>Kenglik (lat)</label><input v-model.number="form.latitude" class="inp" type="number" step="0.0001" /></div>
            <div class="field"><label>Uzunlik (lon)</label><input v-model.number="form.longitude" class="inp" type="number" step="0.0001" /></div>
            <div class="field full">
              <label>Taom rasmi</label>
              <div class="upload-zone" :class="{ 'has-image': !!form.dish_image_url }" @click.self="fileInput?.click()">
                <div class="upload-preview"><img v-if="form.dish_image_url" :src="imageOr(form.dish_image_url)" alt="Preview" /><span v-else class="upload-icon"><AppIcon name="dish" :size="24" /></span></div>
                <div class="upload-info">
                  <b>{{ form.dish_image_url ? 'Rasm tayyor' : 'Rasm tanlang' }}</b>
                  <small>JPG, PNG, WebP · max 10 MB</small>
                  <div class="upload-actions"><button class="abtn primary" type="button" :disabled="uploading" @click="fileInput?.click()">{{ uploading ? 'Yuklanayapti…' : 'Rasm tanlash' }}</button><button v-if="form.dish_image_url" class="abtn danger" type="button" @click="form.dish_image_url = ''">O'chirish</button></div>
                </div>
                <input ref="fileInput" class="upload-input" type="file" accept="image/jpeg,image/png,image/webp,image/gif" :disabled="uploading" @change="uploadImage" />
              </div>
            </div>
          </div>
          <div class="form-actions"><button class="abtn" type="button" @click="showForm = false">Cancel</button><button class="abtn primary" type="submit" :disabled="busy">{{ busy ? 'Saving…' : 'Save' }}</button></div>
        </form>
      </div>
    </div>

    <div class="acard">
      <div class="acard-h"><h3>Regional food map <span class="muted">({{ rows.length }})</span></h3><button class="abtn primary lg" type="button" @click="openCreate"><AppIcon name="plus" :size="15" /> Add region</button></div>
      <div class="atbl-wrap">
        <table class="atbl" v-if="rows.length">
          <thead><tr><th>Region</th><th>Featured dish</th><th>City</th><th>Geo ID</th><th>Status</th><th>Actions</th></tr></thead>
          <tbody>
            <tr v-for="r in rows" :key="r.id">
              <td><div class="who"><img :src="imageOr(r.dish_image_url)" alt="" style="width:34px;height:34px;border-radius:8px;object-fit:cover" /><div><b>{{ r.name }}</b><small>{{ r.name_uz }}</small></div></div></td>
              <td>{{ r.dish_name }}</td><td>{{ r.city || '—' }}</td><td>{{ r.geo_id }}</td>
              <td><span :class="r.status === 'published' ? 'apill g' : 'apill n'"><i class="dot" /> {{ r.status }}</span></td>
              <td><button class="abtn" type="button" @click="openEdit(r)">Edit</button> <button class="abtn" type="button" @click="toggle(r)">{{ r.status === 'published' ? 'Hide' : 'Publish' }}</button> <button class="abtn danger" type="button" @click="remove(r)">Delete</button></td>
            </tr>
          </tbody>
        </table>
      </div>
      <LoadingState v-if="loading" :cards="1" label="Yuklanayapti…" />
      <EmptyState v-else-if="!rows.length" title="Mintaqa yo'q" message="Add region tugmasi bilan qo'shing." icon="pin" />
    </div>
  </div>
</template>
