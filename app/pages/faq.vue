<script setup lang="ts">
definePageMeta({ layout: 'default' })

const faqs = [
  { q: 'How do I book a masterclass?', a: 'Open any experience from the catalog, choose a session and the number of guests, then press “Book this experience”. You will need to sign in first — TourChef keeps your selected masterclass ready after login.' },
  { q: 'Where can I see my bookings?', a: 'All your reservations appear in My Bookings inside your client cabinet. Open a booking to see the session, venue, price breakdown and payment status.' },
  { q: 'How does payment work in development?', a: 'Payments run in a clearly labelled development mode. Bookings are recorded with a paid_dev payment status; no real money is charged.' },
  { q: 'Can I cancel a booking?', a: 'Yes — open the booking in your cabinet and use “Cancel booking” while the session still allows it. Cancellation is reflected in both the client and admin panels.' },
  { q: 'How do I become a chef or restaurant partner?', a: 'Choose “Chef / Restaurant partner” on the registration page. Your account is created as pending and appears in the admin panel for approval.' },
  { q: 'Can I leave a review?', a: 'Reviews can be added from the My reviews page in your client cabinet after you have attended an experience.' }
]
</script>

<template>
  <div>
    <div class="container" style="max-width: 940px">
      <div class="page-head">
        <span class="eyebrow">Good to know</span>
        <h1>Frequently asked questions</h1>
      </div>
      <div class="panel mb-16">
        <div class="panel-b stack">
          <details v-for="f in faqs" :key="f.q" class="faq-item">
            <summary class="fw-600">{{ f.q }}</summary>
            <p class="prose mt-12">{{ f.a }}</p>
          </details>
        </div>
      </div>
      <div class="panel" id="cancellation">
        <div class="panel-b">
          <h2 class="section-title small">Cancellation policy</h2>
          <p class="prose mt-12">
            You may cancel a reservation from your client cabinet up to 24 hours before the session
            starts. Once a session has begun, cancellations are handled by our support team at
            <a href="mailto:support@tourchef.uz" class="text-link">support@tourchef.uz</a>.
            Refunds in development mode are recorded by the administrator.
          </p>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.faq-item { border-bottom: 1px solid var(--border); padding-bottom: 14px; cursor: pointer; }
.faq-item:last-child { border-bottom: 0; padding-bottom: 0; }
summary::marker { color: var(--accent); }
</style>
