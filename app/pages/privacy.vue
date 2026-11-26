<script setup lang="ts">
definePageMeta({ layout: 'default' })
</script>

<template>
  <div>
    <div class="container" style="max-width: 940px">
      <div class="page-head">
        <span class="eyebrow">Legal</span>
        <h1>Privacy Policy</h1>
        <p>Last updated: 1 January 2026</p>
      </div>
      <div class="panel"><div class="panel-b prose">
        <h2 class="section-title small">What we store</h2>
        <p>TourChef stores the account details you provide (name, email, phone, role), your bookings, favorites, reviews and notifications. Your session is kept in a browser cookie named <code>tc_token</code> containing your access token.</p>
        <h2 class="section-title small mt-24">How we use data</h2>
        <p>Your data is used to provide the service: authenticate you, show your reservations, notify you about sessions and let hosts prepare for your visit. We do not sell personal data.</p>
        <h2 class="section-title small mt-24">Your choices</h2>
        <p>You can update your profile from your cabinet or sign out at any time. For account or data requests, contact <a href="mailto:support@tourchef.uz" class="text-link">support@tourchef.uz</a>.</p>
        <p class="mt-24">See also our <NuxtLink to="/terms" class="text-link">Terms and Conditions</NuxtLink>.</p>
      </div></div>
    </div>
  </div>
</template>
