<script setup lang="ts">
import { useApi } from '~/composables/useApi'
import { useToast } from '~/composables/useToast'

definePageMeta({ layout: 'default' })

const api = useApi()
const toast = useToast()
const email = ref('')
const busy = ref(false)
const sent = ref(false)
const error = ref('')

async function submit() {
  error.value = ''
  busy.value = true
  try {
    await api.post('/auth/forgot-password', { email: email.value.trim() })
    sent.value = true
    toast.success('If this email is registered, recovery instructions are on their way.')
  } catch (e: any) {
    error.value = e?.message || 'Could not send recovery instructions.'
    toast.error(error.value)
  } finally {
    busy.value = false
  }
}
</script>

<template>
  <div class="auth-wrap">
    <div class="auth-art">
      <img src="https://thumb.wikimedia.org/wikipedia/commons/thumb/9/90/Shashlik-national_food_of_Uzbekistan.jpg/960px-Shashlik-national_food_of_Uzbekistan.jpg" alt="" />
      <div class="aa-copy">
        <span class="eyebrow" style="color: #fff">TourChef</span>
        <h2>Your next experience awaits</h2>
        <p>Sign back in to continue planning your culinary journey.</p>
      </div>
    </div>
    <div class="auth-panel">
      <div class="auth-card">
        <span class="eyebrow">Account recovery</span>
        <h1 style="margin-top: 12px">Forgot password?</h1>
        <p>Enter your email and we will send you instructions to reset it.</p>

        <AppAlert v-if="error" kind="error" class="mb-16">{{ error }}</AppAlert>
        <AppAlert v-if="sent" kind="success" class="mb-16">
          Recovery request accepted. Check your inbox — or contact
          <a href="mailto:support@tourchef.uz" class="text-link">support@tourchef.uz</a>.
        </AppAlert>

        <form @submit.prevent="submit">
          <div class="field">
            <label for="email">Email</label>
            <input id="email" v-model="email" class="inp" type="email" autocomplete="email" placeholder="you@example.com" required />
          </div>
          <button class="btn btn-primary btn-block btn-lg mt-24" type="submit" :disabled="busy" :class="{ 'is-loading': busy }">
            {{ busy ? 'Sending…' : 'Send recovery instructions' }}
          </button>
        </form>
        <div class="auth-alt"><NuxtLink to="/login"><AppIcon name="back" :size="14" /> Back to sign in</NuxtLink></div>
      </div>
    </div>
  </div>
</template>
