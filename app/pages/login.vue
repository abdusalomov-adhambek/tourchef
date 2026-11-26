<script setup lang="ts">
import { useAuth, homeForRole } from '~/composables/useAuth'
import { useToast } from '~/composables/useToast'

definePageMeta({ layout: 'default' })

const auth = useAuth()
const toast = useToast()
const router = useRouter()
const route = useRoute()

const form = reactive({ email: '', password: '' })
const busy = ref(false)
const error = ref('')

onMounted(() => {
  auth.fetchUser()
  if (auth.user.value) router.replace(homeForRole(auth.user.value.role))
})

async function submit() {
  error.value = ''
  busy.value = true
  try {
    const user = await auth.login(form.email.trim(), form.password)
    toast.success(`Welcome back, ${user.full_name}!`)
    const redirect = typeof route.query.redirect === 'string' ? route.query.redirect : ''
    const safe = redirect.startsWith('/') && !redirect.startsWith('//') ? redirect : homeForRole(user.role)
    await router.push(safe)
  } catch (e: any) {
    error.value = e?.message || 'Invalid email or password.'
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
        <p>Authentic kitchens, welcoming families and experiences you will never forget.</p>
      </div>
    </div>
    <div class="auth-panel">
      <div class="auth-card">
        <span class="eyebrow">Welcome back</span>
        <h1 style="margin-top: 12px">Sign in</h1>
        <p>Book experiences, save favorites and share your reviews.</p>

        <AppAlert v-if="error" kind="error" class="mb-16">{{ error }}</AppAlert>

        <form @submit.prevent="submit">
          <div class="field">
            <label for="email">Email</label>
            <input id="email" v-model="form.email" class="inp" type="email" autocomplete="email" placeholder="you@example.com" required />
          </div>
          <div class="field">
            <label for="password">Password</label>
            <input id="password" v-model="form.password" class="inp" type="password" autocomplete="current-password" placeholder="Your password" minlength="6" required />
          </div>
          <NuxtLink to="/forgot-password" class="text-link mt-12" style="font-size: 13px">Forgot password?</NuxtLink>
          <button class="btn btn-primary btn-block btn-lg mt-24" type="submit" :disabled="busy" :class="{ 'is-loading': busy }">
            {{ busy ? 'Signing in…' : 'Sign in' }}
          </button>
        </form>

        <div class="auth-alt">
          New to TourChef? <NuxtLink to="/register">Create an account</NuxtLink>
        </div>
        <div class="auth-sep">or</div>
        <div class="auth-alt">
          Joining as a family host?
          <NuxtLink :to="`/register?role=chef${route.query.redirect ? `&redirect=${encodeURIComponent(String(route.query.redirect))}` : ''}`">
            Register as a partner
          </NuxtLink>
        </div>
      </div>
    </div>
  </div>
</template>
