<script setup lang="ts">
import { useAuth } from '~/composables/useAuth'

const auth = useAuth()
const router = useRouter()
const route = useRoute()
useHead({ bodyAttrs: { 'data-area': 'cabinet' } })

const primary = [
  { to: '/family', label: 'Dashboard', icon: 'dashboard', exact: true },
  { to: '/family/masterclasses', label: 'My masterclasses', icon: 'book' },
  { to: '/family/sessions', label: 'Sessions', icon: 'calendar' },
  { to: '/family/dishes', label: 'Dish catalog', icon: 'dish' }
]
const account = [{ to: '/family/profile', label: 'Profile settings', icon: 'user' }]

function active(to: string, exact = false) {
  return exact ? route.path === to : route.path === to || route.path.startsWith(`${to}/`)
}
</script>

<template>
  <div class="cab-shell">
    <aside class="cab-side">
      <div class="cab-brand">
        <NuxtLink to="/"><img src="/logo.svg" alt="TourChef" class="logo-img side" width="138" height="46" /></NuxtLink>
        <span class="badge gold">Family</span>
      </div>
      <div class="side-label">Kitchen</div>
      <nav class="cab-nav">
        <NuxtLink v-for="item in primary" :key="item.to" :to="item.to" :class="{ on: active(item.to, item.exact) }">
          <AppIcon :name="item.icon" :size="18" /> {{ item.label }}
        </NuxtLink>
      </nav>
      <div class="side-label">Account</div>
      <nav class="cab-nav">
        <NuxtLink v-for="item in account" :key="item.to" :to="item.to" :class="{ on: active(item.to) }">
          <AppIcon :name="item.icon" :size="18" /> {{ item.label }}
        </NuxtLink>
        <button type="button" @click="auth.signOutAndGo(router)">
          <AppIcon name="logout" :size="18" /> Sign out
        </button>
      </nav>
      <div class="side-label" style="margin-top: 18px">
        {{ auth.user.value?.full_name || 'Family account' }}
      </div>
    </aside>
    <main class="cab-main">
      <slot />
    </main>
  </div>
</template>

<style scoped>
.cab-nav button {
  display: flex; align-items: center; gap: 11px; width: 100%;
  padding: 10px 12px; border: 0; border-radius: 10px; background: transparent;
  font-size: 14px; font-weight: 600; color: var(--text-2); text-align: left;
}
.cab-nav button:hover { background: var(--bg); color: var(--text); }
</style>
