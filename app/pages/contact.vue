<script setup lang="ts">
definePageMeta({ layout: 'default' })
const sent = ref(false)
const form = reactive({ name: '', email: '', message: '' })

function submit() {
  const subject = encodeURIComponent(`TourChef enquiry from ${form.name}`)
  const body = encodeURIComponent(`${form.message}\n\nFrom: ${form.name} <${form.email}>`)
  window.location.href = `mailto:support@tourchef.uz?subject=${subject}&body=${body}`
  sent.value = true
}
</script>

<template>
  <div>
    <div class="container">
      <div class="page-head">
        <span class="eyebrow">We are here to help</span>
        <h1>Contact TourChef</h1>
        <p>Questions about a booking, a kitchen or becoming a partner? Reach out to our team.</p>
      </div>
      <div class="grid-2">
        <div class="panel">
          <div class="panel-b">
            <h2 class="section-title small">Send us a message</h2>
            <AppAlert v-if="sent" kind="success" class="mt-16">Your email app should now be open with your message ready to send.</AppAlert>
            <form class="mt-16" @submit.prevent="submit">
              <div class="field"><label for="name">Your name</label><input id="name" v-model="form.name" class="inp" required /></div>
              <div class="field mt-16"><label for="email">Email</label><input id="email" v-model="form.email" class="inp" type="email" required /></div>
              <div class="field mt-16"><label for="message">Message</label><textarea id="message" v-model="form.message" class="textarea" rows="6" required /></div>
              <button class="btn btn-primary mt-16" type="submit">Open email <AppIcon name="mail" :size="16" /></button>
            </form>
          </div>
        </div>
        <div class="panel">
          <div class="panel-b">
            <h2 class="section-title small">Other ways to reach us</h2>
            <div class="info-row mt-16"><AppIcon name="mail" :size="18" /><span>Email</span> <a href="mailto:support@tourchef.uz" class="text-link">support@tourchef.uz</a></div>
            <div class="info-row"><AppIcon name="phone" :size="18" /><span>Phone</span> +998 71 200 00 00</div>
            <div class="info-row"><AppIcon name="pin" :size="18" /><span>Office</span> Tashkent, Uzbekistan</div>
            <div class="info-row"><AppIcon name="globe" :size="18" /><span>Social</span> <a href="https://www.instagram.com/" target="_blank" rel="noopener" class="text-link">Instagram</a></div>
            <p class="muted text-sm mt-24">Looking for an answer right away? See our <NuxtLink to="/faq" class="text-link">FAQ</NuxtLink>.</p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
