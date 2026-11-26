<script setup lang="ts">
import { useAuth, homeForRole, type UserRole } from '~/composables/useAuth'
import { useToast } from '~/composables/useToast'

definePageMeta({ layout: 'default' })

const auth = useAuth()
const toast = useToast()
const router = useRouter()
const route = useRoute()

const role = ref<'client' | 'chef'>(
  String(route.query.role) === 'chef' ? 'chef' : 'client'
)
const form = reactive({ full_name: '', email: '', phone: '', password: '', confirm: '', terms: false })
const busy = ref(false)
const error = ref('')

const roles: { value: 'client' | 'chef'; label: string; description: string; icon: string }[] = [
  { value: 'client', label: 'Traveler / Client', description: 'Book masterclasses and save favorites.', icon: 'user' },
  { value: 'chef', label: 'Family host', description: 'Share family recipes, host sessions and welcome guests. Appears in admin for approval.', icon: 'family' }
]

async function submit() {
  error.value = ''
  if (form.password !== form.confirm) {
    error.value = 'Passwords do not match.'
    toast.error(error.value)
    return
  }
  if (!form.terms) {
    error.value = 'Please accept the terms and privacy policy.'
    toast.error(error.value)
    return
  }
  busy.value = true
  try {
    const user = await auth.register({
      full_name: form.full_name.trim(),
      email: form.email.trim(),
      phone: form.phone.trim() || undefined,
      password: form.password,
      role: role.value
    })
    toast.success(
      role.value === 'chef'
        ? 'Your chef account was created and sent to admin for approval.'
        : `Welcome to TourChef, ${user.full_name}!`
    )
    const redirect = typeof route.query.redirect === 'string' ? route.query.redirect : ''
    const safe = redirect.startsWith('/') && !redirect.startsWith('//') ? redirect : homeForRole(user.role as UserRole)
    await router.push(safe)
  } catch (e: any) {
    error.value = e?.message || 'Could not create your account.'
    toast.error(error.value)
  } finally {
    busy.value = false
  }
}
</script>

<template>
  <div class="auth-wrap">
    <div class="auth-art">
      <img src="https://thumb.wikimedia.org/wikipedia/commons/thumb/0/04/Uzbek_national_cuisine.jpg/960px-Uzbek_national_cuisine.jpg" alt="" />
      <div class="aa-copy">
        <span class="eyebrow" style="color: #fff">TourChef</span>
        <h2>Flavors worth traveling for</h2>
        <p>Authentic kitchens, real chefs and experiences you will never forget.</p>
      </div>
    </div>
    <div class="auth-panel">
      <div class="auth-card">
        <span class="eyebrow">Welcome</span>
        <h1 style="margin-top: 12px">Create account</h1>
        <p>Book experiences, save favorites and share your reviews.</p>

        <AppAlert v-if="error" kind="error" class="mb-16">{{ error }}</AppAlert>

        <div class="chips mb-24">
          <button
            v-for="r in roles"
            :key="r.value"
            class="chip"
            :class="{ on: role === r.value }"
            type="button"
            @click="role = r.value"
          >
            <AppIcon :name="r.icon" :size="15" /> {{ r.label }}
          </button>
        </div>
        <p class="muted text-sm mb-16">
          {{ roles.find((r) => r.value === role)?.description }}
        </p>

        <form @submit.prevent="submit">
          <div class="field">
            <label for="full_name">Full name</label>
            <input id="full_name" v-model="form.full_name" class="inp" type="text" autocomplete="name" placeholder="Aziza Karimova" required minlength="2" />
          </div>
          <div class="field mt-16">
            <label for="email">Email</label>
            <input id="email" v-model="form.email" class="inp" type="email" autocomplete="email" placeholder="you@example.com" required />
          </div>
          <div class="field mt-16">
            <label for="phone">Phone number <span class="muted">(optional)</span></label>
            <input id="phone" v-model="form.phone" class="inp" type="tel" autocomplete="tel" placeholder="+998 90 123 45 67" />
          </div>
          <div class="field mt-16">
            <label for="password">Password</label>
            <input id="password" v-model="form.password" class="inp" type="password" autocomplete="new-password" placeholder="At least 8 characters" minlength="8" required />
          </div>
          <div class="field mt-16">
            <label for="confirm">Confirm password</label>
            <input id="confirm" v-model="form.confirm" class="inp" type="password" autocomplete="new-password" placeholder="Repeat password" minlength="8" required />
          </div>
          <label class="check-line mt-16">
            <input v-model="form.terms" type="checkbox" required />
            I agree to the <NuxtLink to="/terms" class="text-link">Terms</NuxtLink> and
            <NuxtLink to="/privacy" class="text-link">Privacy Policy</NuxtLink>
          </label>
          <button class="btn btn-primary btn-block btn-lg mt-24" type="submit" :disabled="busy" :class="{ 'is-loading': busy }">
            {{ busy ? 'Creating account…' : 'Create account' }}
          </button>
        </form>

        <div class="auth-alt">Already have an account? <NuxtLink to="/login">Sign in</NuxtLink></div>
      </div>
    </div>
  </div>
</template>
