<script setup lang="ts">
import { homeForRole } from '~/composables/useAuth'

const auth = useAuth()
const route = useRoute()
const router = useRouter()
const mobileOpen = ref(false)

const links = [
  { to: '/', label: 'Home' },
  { to: '/masterclasses', label: 'Masterclasses' },
  { to: '/reels', label: 'Reels' },
  { to: '/restaurants', label: 'Restaurants' },
  { to: '/dishes', label: 'Dishes' },
  { to: '/about', label: 'About' }
]

watch(
  () => route.fullPath,
  () => {
    mobileOpen.value = false
  }
)

onMounted(() => {
  auth.fetchUser()
})
</script>

<template>
  <header class="site-header">
    <div class="container">
      <NuxtLink to="/" class="brand" aria-label="TourChef home">
        <img src="/logo.svg" alt="TourChef" class="logo-img" width="138" height="46" />
      </NuxtLink>

      <nav class="main-nav" aria-label="Main navigation">
        <NuxtLink
          v-for="link in links"
          :key="link.to"
          :to="link.to"
          :class="{ active: route.path === link.to }"
        >
          {{ link.label }}
        </NuxtLink>
      </nav>

      <div class="header-actions">
        <template v-if="auth.user.value">
          <NuxtLink :to="homeForRole(auth.user.value.role)" class="btn btn-ghost btn-sm">
            My cabinet
          </NuxtLink>
          <NuxtLink :to="`/register?role=chef`" class="btn btn-outline btn-sm">
            Become a Family Host
          </NuxtLink>
          <button class="btn btn-outline btn-sm" type="button" @click="auth.signOutAndGo(router)">
            Sign out
          </button>
        </template>
        <template v-else>
          <NuxtLink to="/login" class="btn btn-ghost btn-sm">Login</NuxtLink>
          <NuxtLink to="/register" class="btn btn-primary btn-sm">Register</NuxtLink>
          <NuxtLink to="/register?role=chef" class="btn btn-outline btn-sm partner-btn">
            Become a Family Host
          </NuxtLink>
        </template>
      </div>

      <button
        class="menu-toggle"
        :class="{ open: mobileOpen }"
        type="button"
        aria-label="Toggle menu"
        :aria-expanded="mobileOpen"
        @click="mobileOpen = !mobileOpen"
      >
        <span />
      </button>
    </div>

    <nav class="mobile-menu" :class="{ open: mobileOpen }" aria-label="Mobile navigation">
      <NuxtLink v-for="link in links" :key="link.to" :to="link.to">{{ link.label }}</NuxtLink>
      <template v-if="auth.user.value">
        <NuxtLink :to="homeForRole(auth.user.value.role)">My cabinet</NuxtLink>
        <button class="btn btn-outline btn-sm" type="button" @click="auth.signOutAndGo(router)">
          Sign out
        </button>
      </template>
      <template v-else>
        <NuxtLink to="/login">Login</NuxtLink>
        <NuxtLink to="/register">Register</NuxtLink>
      </template>
      <NuxtLink to="/register?role=chef">Become a Family Host</NuxtLink>
    </nav>
  </header>
</template>

<style scoped>
.partner-btn { display: inline-flex; }
@media (max-width: 1080px) { .partner-btn { display: none; } }
.mobile-menu .btn { margin-top: 8px; }
</style>
