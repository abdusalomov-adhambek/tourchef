<script setup lang="ts">
import { useApi } from '~/composables/useApi'
import { useAuth } from '~/composables/useAuth'
import { useToast } from '~/composables/useToast'
import { dateLabel, bookingStart, bookingMc, bookingSession, bookingTotal, money, pillClass, humanStatus } from '~/composables/useFormat'

definePageMeta({ layout: 'cabinet' })

const route = useRoute()
const router = useRouter()
const api = useApi()
const auth = useAuth()
const toast = useToast()
const booking = ref<any>(null)
const qrUrl = ref('')
const loading = ref(true)
const error = ref('')

async function ensureQR() {
  if (!booking.value || booking.value.qr_code_url) return
  try {
    const res = await api.get<any>(`/client/bookings/${booking.value.id}/qr`)
    qrUrl.value = res.url
    booking.value.qr_code_url = res.url
  } catch (e: any) {
    toast.error(e?.message || 'QR kod yaratilmadi.')
  }
}

const price = computed(() => bookingTotal(booking.value))
const platformFee = computed(() => Number(booking.value?.platform_fee || booking.value?.commission || 0))
const grandTotal = computed(() => price.value + platformFee.value)

async function load() {
  loading.value = true
  try {
    const data = await api.get<any[]>('/client/bookings')
    const list = Array.isArray(data) ? data : []
    booking.value = list.find((b) => String(b.id) === String(route.params.id)) || null
    if (!booking.value) error.value = 'Booking not found in your account.'
    else {
      qrUrl.value = booking.value.qr_code_url || ''
      await ensureQR()
    }
  } catch (e: any) {
    error.value = e?.message || 'Could not load this booking.'
  } finally {
    loading.value = false
  }
}

async function cancel() {
  if (!window.confirm('Cancel this booking? This cannot be undone.')) return
  try {
    const updated = await api.patch<any>(`/client/bookings/${booking.value.id}/cancel`, {})
    booking.value = { ...booking.value, ...updated }
    toast.success('Booking cancelled.')
  } catch (e: any) {
    toast.error(e?.message || 'Could not cancel booking.')
  }
}

onMounted(load)
</script>

<template>
  <div>
    <NuxtLink to="/client/bookings" class="text-link"><AppIcon name="back" :size="16" /> All bookings</NuxtLink>

    <div v-if="loading" class="skeleton mt-16" style="height: 420px; border-radius: 16px" />
    <AppAlert v-else-if="error" kind="error" title="Booking unavailable">{{ error }}</AppAlert>

    <template v-else-if="booking">
      <div class="cab-top mt-16">
        <div>
          <h1>{{ booking.booking_number || booking.id }}</h1>
          <p>
            <span :class="pillClass(booking.status)">{{ humanStatus(booking.status) }}</span>
            <span :class="pillClass(booking.payment_status)" style="margin-left: 8px">{{ humanStatus(booking.payment_status) }}</span>
          </p>
        </div>
        <button
          v-if="['confirmed', 'pending', 'pending_payment'].includes(String(booking.status || '').toLowerCase())"
          class="btn btn-outline btn-sm" type="button" @click="cancel"
        >Cancel booking</button>
      </div>

      <div class="panel mt-16 mb-16 qr-ticket">
        <div class="panel-h">
          <h3>QR chipta</h3>
          <span class="badge sage">{{ booking.booking_number }}</span>
        </div>
        <div class="panel-b qr-ticket-body">
          <img v-if="qrUrl" :src="qrUrl" alt="Booking QR code" class="qr-image" />
          <div v-else class="skeleton" style="width:150px;height:150px;border-radius:12px" />
          <div class="qr-info">
            <b>Bu QR ni restoranda ko'rsating</b>
            <p class="muted">Chipta bron qilinganda avtomatik yaratiladi. Unda shu bron havolasi yozilgan.</p>
            <a v-if="qrUrl" :href="qrUrl" :download="`tourchef-${booking.booking_number}.png`" class="btn btn-outline btn-sm">Yuklab olish</a>
          </div>
        </div>
      </div>

      <div class="two-col">
        <div class="panel">
          <div class="panel-h"><h3>Experience</h3></div>
          <div class="panel-b">
            <div v-if="bookingMc(booking)?.slug" class="mini-card mb-16">
              <img :src="bookingMc(booking)?.image_url || '/logo.svg'" alt="" />
              <div class="mc-body">
                <b>{{ bookingMc(booking)?.title }}</b>
                <small>{{ bookingMc(booking)?.city }}</small>
              </div>
              <NuxtLink :to="`/masterclasses/${bookingMc(booking).slug}`" class="btn btn-outline btn-sm">View</NuxtLink>
            </div>
            <dl class="akv" style="grid-template-columns: 170px 1fr">
              <dt>Booking number</dt><dd>{{ booking.booking_number || booking.id }}</dd>
              <dt>Masterclass</dt><dd>{{ bookingMc(booking)?.title || '—' }}</dd>
              <dt>Restaurant</dt><dd>{{ bookingMc(booking)?.restaurant?.name || booking.restaurant?.name || '—' }}</dd>
              <dt>Session date</dt><dd>{{ dateLabel(bookingStart(booking)) }}</dd>
              <dt>Duration</dt><dd>{{ bookingMc(booking)?.duration_min ? `${bookingMc(booking).duration_min} min` : '—' }}</dd>
              <dt>Guests</dt><dd>{{ booking.guests || 1 }}</dd>
              <dt>Guest info</dt><dd>{{ booking.guest_name || booking.guest_email || auth.user.value?.full_name || '—' }}</dd>
            </dl>
          </div>
        </div>
        <div class="panel">
          <div class="panel-h"><h3>Payment summary</h3></div>
          <div class="panel-b">
            <div class="w-summary">
              <div><span>Price × guests</span><span>{{ money(price) }}</span></div>
              <div><span>Platform fee</span><span>{{ money(platformFee) }}</span></div>
              <div><span>Payment status</span><span>{{ humanStatus(booking.payment_status) }}</span></div>
            </div>
            <div class="w-total"><span>Total</span><span>{{ money(grandTotal || price) }}</span></div>
            <AppAlert kind="info" class="mt-16">Payments are recorded in development mode (<code>paid_dev</code>). No real charge was made.</AppAlert>
          </div>
        </div>
      </div>

      <div class="panel mt-24">
        <div class="panel-h"><h3>Location & policy</h3></div>
        <div class="panel-b">
          <p class="prose">
            {{ bookingMc(booking)?.restaurant?.name || 'Your host kitchen' }} ·
            {{ bookingMc(booking)?.city || bookingSession(booking)?.city || 'Uzbekistan' }}
          </p>
          <p class="muted text-sm mt-12">
            Free cancellation up to 24 hours before the session. See the
            <NuxtLink to="/faq#cancellation" class="text-link">cancellation policy</NuxtLink>.
          </p>
        </div>
      </div>
    </template>
  </div>
</template>
