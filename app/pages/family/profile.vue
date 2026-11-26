<script setup lang="ts">
import { useAuth } from '~/composables/useAuth'
import { useToast } from '~/composables/useToast'
import { dateLabel } from '~/composables/useFormat'

definePageMeta({ layout: 'family' })

const auth = useAuth()
const toast = useToast()
const router = useRouter()
const form = reactive({ full_name: '', email: '', phone: '' })
const busy = ref(false)
const error = ref('')

onMounted(() => {
  form.full_name = auth.user.value?.full_name || ''
  form.email = auth.user.value?.email || ''
  form.phone = auth.user.value?.phone || ''
})

async function save() {
  error.value = ''
  if (!form.full_name.trim() || !/^\S+@\S+\.\S+$/.test(form.email)) {
    error.value = 'Enter a valid name and email.'
    toast.error(error.value)
    return
  }
  busy.value = true
  try {
    await auth.api.patch('/auth/me', {
      full_name: form.full_name.trim(),
      email: form.email.trim(),
      phone: form.phone.trim()
    })
    if (auth.user.value) auth.user.value = { ...auth.user.value, full_name: form.full_name.trim(), email: form.email.trim(), phone: form.phone.trim() }
    toast.success('Profile details saved.')
  } catch (e: any) {
    error.value = e?.message || 'Profile updates are not available in this API version.'
    toast.error(error.value)
  } finally {
    busy.value = false
  }
}
</script>

<template>
  <div>
    <div class="cab-top">
      <div><h1>Profile settings</h1><p>Your family account details.</p></div>
    </div>
    <div class="two-col">
      <div class="panel">
        <div class="panel-h"><h3>Personal details</h3></div>
        <div class="panel-b">
          <AppAlert v-if="error" kind="error" class="mb-16">{{ error }}</AppAlert>
          <form @submit.prevent="save">
            <div class="field"><label for="name">Full name</label><input id="name" v-model="form.full_name" class="inp" required /></div>
            <div class="field mt-16"><label for="email">Email</label><input id="email" v-model="form.email" class="inp" type="email" required /></div>
            <div class="field mt-16"><label for="phone">Phone</label><input id="phone" v-model="form.phone" class="inp" type="tel" /></div>
            <button class="btn btn-primary mt-24" type="submit" :disabled="busy" :class="{ 'is-loading': busy }">{{ busy ? 'Saving…' : 'Save changes' }}</button>
          </form>
        </div>
      </div>
      <div class="panel">
        <div class="panel-h"><h3>Account</h3></div>
        <div class="panel-b">
          <dl class="akv" style="grid-template-columns: 110px 1fr">
            <dt>Role</dt><dd>{{ auth.user.value?.role }}</dd>
            <dt>Status</dt><dd>{{ auth.user.value?.status }}</dd>
            <dt>Member since</dt><dd>{{ auth.user.value?.created_at ? dateLabel(auth.user.value.created_at, false) : '—' }}</dd>
          </dl>
          <button class="btn btn-outline btn-block mt-24" type="button" @click="auth.signOutAndGo(router)"><AppIcon name="logout" :size="16" /> Sign out</button>
        </div>
      </div>
    </div>
  </div>
</template>
