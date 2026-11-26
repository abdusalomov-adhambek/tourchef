<script setup lang="ts">
import { useApi } from '~/composables/useApi'
import { useAuth } from '~/composables/useAuth'
import { initials } from '~/composables/useFormat'

const auth = useAuth()
const api = useApi()
const router = useRouter()
const route = useRoute()
useHead({ bodyAttrs: { 'data-area': 'admin' } })

interface NavItem {
  to: string
  label: string
  icon: string
  exact?: boolean
  badge?: string
}

const sections: { label: string; items: NavItem[] }[] = [
  {
    label: 'Overview',
    items: [{ to: '/admin', label: 'Dashboard', icon: 'dashboard', exact: true }]
  },
  {
    label: 'Management',
    items: [
      { to: '/admin/users', label: 'Users', icon: 'users' },
      { to: '/admin/families', label: 'Families', icon: 'family', badge: 'pending_chefs' },
      { to: '/admin/masterclasses', label: 'Masterclasses', icon: 'book', badge: 'pending_mc' },
      { to: '/admin/sessions', label: 'Sessions', icon: 'calendar' },
      { to: '/admin/bookings', label: 'Bookings', icon: 'ticket' },
      { to: '/admin/payments', label: 'Payments', icon: 'card' }
    ]
  },
  {
    label: 'Content & insights',
    items: [
      { to: '/admin/reels', label: 'Reels moderation', icon: 'film' },
      { to: '/admin/dishes', label: 'Dishes & cuisines', icon: 'dish' },
      { to: '/admin/locations', label: 'Locations', icon: 'pin' },
      { to: '/admin/regions', label: 'Taste of Uzbekistan', icon: 'globe' },
      { to: '/admin/reports', label: 'Reports', icon: 'chart' }
    ]
  },
  {
    label: 'System',
    items: [{ to: '/admin/settings', label: 'Settings', icon: 'settings' }]
  }
]

const stats = useState<Record<string, any>>('tc-admin-stats', () => ({}))

onMounted(async () => {
  if (!auth.token.value) return
  try {
    stats.value = await api.get<Record<string, any>>('/admin/stats')
  } catch {
    /* dashboard page shows its own error state */
  }
})

function active(to: string, exact = false) {
  return exact ? route.path === to : route.path === to || route.path.startsWith(`${to}/`)
}
</script>

<template>
  <div class="admin-page">
    <div class="admin-app">
      <aside class="admin-side">
        <div class="admin-brand">
          <NuxtLink to="/admin" aria-label="TourChef admin">
            <img src="/logo.svg" alt="TourChef" class="logo-img" />
          </NuxtLink>
          <span class="badge" style="margin-left:auto;font-size:10px">Admin</span>
        </div>
        <div class="admin-side-mid">
          <div v-for="section in sections" :key="section.label">
            <div class="admin-sect">{{ section.label }}</div>
            <nav class="admin-nav">
              <NuxtLink
                v-for="item in section.items"
                :key="item.to"
                :to="item.to"
                :class="{ on: active(item.to, item.exact) }"
              >
                <AppIcon :name="item.icon === 'ticket' ? 'book' : item.icon" :size="17" />
                <span class="lbl">{{ item.label }}</span>
                <span
                  v-if="item.badge && Number(stats[item.badge] || 0) > 0"
                  class="cnt"
                >{{ stats[item.badge] }}</span>
              </NuxtLink>
            </nav>
          </div>
        </div>
        <div class="admin-side-foot">
          <NuxtLink to="/" class="abtn" style="justify-content: center">View live site</NuxtLink>
          <button class="abtn" style="justify-content: center" type="button" @click="auth.signOutAndGo(router)">
            Sign out
          </button>
        </div>
      </aside>
      <main class="admin-main">
        <div class="admin-top">
          <h1>
            <span class="crumb">Admin /</span>
            {{ route.path === '/admin' ? 'Dashboard' : String(route.path.split('/').pop() || '').replace(/-/g, ' ') }}
          </h1>
          <div class="me">
            <span class="avatar-sm">{{ initials(auth.user.value?.full_name) }}</span>
            <span>{{ auth.user.value?.full_name || 'Administrator' }}</span>
          </div>
        </div>
        <slot />
      </main>
    </div>
  </div>
</template>
