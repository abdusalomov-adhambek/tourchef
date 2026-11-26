<script setup lang="ts">
import { useApi } from '~/composables/useApi'
import { useToast } from '~/composables/useToast'
import { dateLabel, money, apPillClass, humanStatus, imageOr } from '~/composables/useFormat'

definePageMeta({ layout: 'admin' })

const api = useApi()
const toast = useToast()
const route = useRoute()
const slug = String(route.params.slug)

const data = ref<any>(null)
const loading = ref(true)
const error = ref('')

const mc = computed(() => data.value?.masterclass || null)
const stats = computed(() => data.value?.stats || {})
const sessions = computed(() => data.value?.sessions || [])
const bookings = computed(() => data.value?.bookings || [])

async function load() {
  loading.value = true
  try {
    data.value = await api.get<any>(`/admin/masterclasses/${slug}`)
  } catch (e: any) {
    error.value = e?.message || 'Masterclass topilmadi.'
    toast.error(error.value)
  } finally {
    loading.value = false
  }
}

function seatsLeft(s: any) {
  return Math.max(0, Number(s.capacity || 0) - Number(s.booked || 0))
}

onMounted(load)
</script>

<template>
  <div>
    <div class="admin-top" style="margin-bottom: 18px">
      <h1><AppIcon name="book" :size="18" /> Masterclass tafsilotlari</h1>
      <span class="crumb">Admin / Masterclasses / {{ slug }}</span>
      <NuxtLink to="/admin/masterclasses" class="abtn"><AppIcon name="back" :size="15" /> Orqaga</NuxtLink>
    </div>

    <AppAlert v-if="error" kind="error">{{ error }}</AppAlert>
    <LoadingState v-if="loading" :cards="1" label="Yuklanayapti…" />

    <template v-else-if="mc">
      <div class="agrid k4 mb-16">
        <div class="acard akpi"><span class="ic tint-a"><AppIcon name="users" :size="17" /></span><div><small>Jami o'rin</small><b>{{ stats.capacity ?? 0 }}</b></div></div>
        <div class="acard akpi"><span class="ic tint-g"><AppIcon name="check" :size="17" /></span><div><small>Bron qilgan kishilar</small><b>{{ stats.booked ?? 0 }}</b></div></div>
        <div class="acard akpi"><span class="ic tint-w"><AppIcon name="calendar" :size="17" /></span><div><small>Qolgan o'rinlar</small><b>{{ stats.seats_left ?? 0 }}</b></div></div>
        <div class="acard akpi"><span class="ic tint-b"><AppIcon name="book" :size="17" /></span><div><small>Bronlar soni</small><b>{{ stats.bookings_count ?? 0 }}</b></div></div>
      </div>

      <div class="acard mb-16">
        <div class="panel-b" style="display:flex;gap:18px;align-items:center;flex-wrap:wrap">
          <img :src="imageOr(mc.image_url || mc.dish?.image_url)" :alt="mc.title" style="width:110px;height:86px;object-fit:cover;border-radius:11px" />
          <div style="flex:1;min-width:220px">
            <span class="eyebrow">{{ mc.dish?.name || 'Uzbek cuisine' }} · {{ mc.city }}</span>
            <h2 style="font-family:var(--font-display);font-size:24px;margin:4px 0">{{ mc.title }}</h2>
            <p class="muted" style="font-size:13px;max-width:70ch">{{ mc.description }}</p>
          </div>
          <div style="display:grid;gap:6px;justify-items:end">
            <b style="font-size:18px">{{ money(mc.price) }}</b>
            <span :class="apPillClass(mc.status)"><i class="dot" /> {{ humanStatus(mc.status) }}</span>
            <small class="muted">{{ mc.family?.display_name || mc.chef?.display_name || 'Family host' }}</small>
          </div>
        </div>
      </div>

      <div class="acard mb-16">
        <div class="acard-h"><h3>Sessionlar <span class="muted">({{ sessions.length }})</span></h3></div>
        <div class="atbl-wrap">
          <table class="atbl" v-if="sessions.length">
            <thead><tr><th>Sana va vaqt</th><th>Sig'im</th><th>Bron qilingan</th><th>Qoldi</th><th>Narx</th><th>Holat</th></tr></thead>
            <tbody>
              <tr v-for="s in sessions" :key="s.id">
                <td><b>{{ dateLabel(s.starts_at) }}</b></td>
                <td>{{ s.capacity }}</td>
                <td>{{ s.booked }}</td>
                <td><b style="color:var(--a-accent)">{{ seatsLeft(s) }}</b></td>
                <td>{{ money(s.price) }}</td>
                <td><span :class="apPillClass(s.status)"><i class="dot" /> {{ humanStatus(s.status) }}</span></td>
              </tr>
            </tbody>
          </table>
          <EmptyState v-else title="Session yo'q" message="Bu masterclass uchun hali sana qo'shilmagan." icon="calendar" />
        </div>
      </div>

      <div class="acard">
        <div class="acard-h"><h3>Bron qilganlar <span class="muted">({{ bookings.length }})</span></h3></div>
        <div class="atbl-wrap">
          <table class="atbl" v-if="bookings.length">
            <thead><tr><th>Bron</th><th>Mehmon</th><th>Email</th><th>Session</th><th>Kishi</th><th>Holat</th><th>To'lov</th><th>Sana</th></tr></thead>
            <tbody>
              <tr v-for="b in bookings" :key="b.id">
                <td><b>{{ b.booking_number }}</b></td>
                <td>{{ b.guest_name || '—' }}</td>
                <td>{{ b.guest_email || '—' }}</td>
                <td>{{ dateLabel(b.starts_at) }}</td>
                <td>{{ b.guests }}</td>
                <td><span :class="apPillClass(b.status)"><i class="dot" /> {{ humanStatus(b.status) }}</span></td>
                <td><span :class="apPillClass(b.payment_status)"><i class="dot" /> {{ humanStatus(b.payment_status) }}</span></td>
                <td>{{ dateLabel(b.created_at, false) }}</td>
              </tr>
            </tbody>
          </table>
          <EmptyState v-else title="Hali bron yo'q" message="Bu masterclassga hech kim bron qilmadi." icon="book" />
        </div>
      </div>
    </template>
  </div>
</template>
