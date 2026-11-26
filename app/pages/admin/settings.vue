<script setup lang="ts">
import { useToast } from '~/composables/useToast'

definePageMeta({ layout: 'admin' })

const toast = useToast()
const form = reactive({
  currency: 'USD',
  language: 'English',
  commission: 10,
  cancellationHours: 24,
  emailNotifications: true,
  reelModeration: true,
  maintenance: false
})
const saving = ref(false)
const error = ref('')

onMounted(() => {
  if (!import.meta.client) return
  try {
    const saved = JSON.parse(localStorage.getItem('tc-admin-settings') || 'null')
    if (saved) Object.assign(form, saved)
  } catch { /* ignore bad local data */ }
})

async function save() {
  error.value = ''
  if (Number(form.commission) < 0 || Number(form.commission) > 100) {
    error.value = 'Commission must be between 0 and 100.'
    toast.error(error.value)
    return
  }
  if (Number(form.cancellationHours) < 0) {
    error.value = 'Cancellation window cannot be negative.'
    toast.error(error.value)
    return
  }
  saving.value = true
  await nextTick()
  try {
    if (import.meta.client) localStorage.setItem('tc-admin-settings', JSON.stringify({ ...form }))
    toast.success('Platform preferences saved for this admin session.')
  } catch {
    error.value = 'Could not save preferences in this browser.'
    toast.error(error.value)
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <div class="agrid r2">
    <div class="acard">
      <div class="acard-h"><h3>Platform settings</h3></div>
      <div class="acard-b">
        <AppAlert v-if="error" kind="error">{{ error }}</AppAlert>
        <form @submit.prevent="save">
          <div class="form-grid">
            <div class="field"><label for="currency">Currency</label><select id="currency" v-model="form.currency" class="select"><option>USD</option><option>so'm</option></select><span class="hint">Narxlar interfeysda dollarda ko'rsatiladi.</span></div>
            <div class="field"><label for="language">Default language</label><select id="language" v-model="form.language" class="select"><option>English</option><option>Uzbek</option><option>Russian</option></select></div>
            <div class="field"><label for="commission">Platform commission (%)</label><input id="commission" v-model.number="form.commission" class="inp" type="number" min="0" max="100" /></div>
            <div class="field"><label for="hours">Free cancellation window (hours)</label><input id="hours" v-model.number="form.cancellationHours" class="inp" type="number" min="0" /></div>
          </div>
          <div class="stack sm mt-16">
            <label class="inline-check"><input v-model="form.emailNotifications" type="checkbox" /> Email notifications enabled</label>
            <label class="inline-check"><input v-model="form.reelModeration" type="checkbox" /> Require reel moderation before publishing</label>
            <label class="inline-check"><input v-model="form.maintenance" type="checkbox" /> Maintenance mode</label>
          </div>
          <div class="form-actions">
            <button class="abtn primary lg" type="submit" :disabled="saving">{{ saving ? 'Saving…' : 'Save settings' }}</button>
          </div>
        </form>
      </div>
    </div>
    <div class="acard">
      <div class="acard-h"><h3>About settings</h3></div>
      <div class="acard-b">
        <AppAlert kind="info">This API version does not expose a settings endpoint. Preferences are saved in your browser and applied to admin tooling; production settings are configured via environment variables on the server.</AppAlert>
        <div class="abox mt-16"><h4>Quick links</h4><ul class="atl"><li><NuxtLink to="/admin/dishes" class="text-link">Manage content catalog</NuxtLink></li><li><NuxtLink to="/admin/reels" class="text-link">Review reel moderation queue</NuxtLink></li><li><NuxtLink to="/admin/reports" class="text-link">Export platform reports</NuxtLink></li></ul></div>
      </div>
    </div>
  </div>
</template>
