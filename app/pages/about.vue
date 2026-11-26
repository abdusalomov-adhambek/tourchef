<script setup lang="ts">
definePageMeta({ layout: 'default' })
</script>

<template>
  <div>
    <section class="section tight" style="background: var(--surface); border-bottom: 1px solid var(--border)">
      <div class="container split">
        <div>
          <span class="eyebrow">Our mission</span>
          <h1 class="page-title">Culture you can <em style="color: var(--accent)">taste</em></h1>
          <p class="page-sub">
            TourChef connects travelers with authentic Uzbek culinary culture through local
            restaurants, families and cooking experiences. Every dish tells a story — we help you live it.
          </p>
          <NuxtLink to="/masterclasses" class="btn btn-primary mt-24">Explore experiences <AppIcon name="arrow" :size="16" /></NuxtLink>
        </div>
        <img
          src="https://thumb.wikimedia.org/wikipedia/commons/thumb/a/ac/Uzbek_national_dastarkhan_and_wheat_prepared_for_sumalak.jpg/960px-Uzbek_national_dastarkhan_and_wheat_prepared_for_sumalak.jpg"
          alt="Uzbek culinary culture"
          style="border-radius: 24px; width: 100%; aspect-ratio: 4/3; object-fit: cover"
        />
      </div>
    </section>
    <section class="section">
      <div class="container">
        <div class="grid-3">
          <div class="panel"><div class="panel-b"><span class="badge sage">For travelers</span><h2 class="section-title small mt-16">Real kitchens</h2><p class="prose">Discover real restaurants, book with confidence, cook with local families and take memories home.</p></div></div>
          <div class="panel"><div class="panel-b"><span class="badge gold">For restaurants</span><h2 class="section-title small mt-16">New guests</h2><p class="prose">Reach travelers worldwide, showcase your signature dishes and grow your culinary brand.</p></div></div>
          <div class="panel"><div class="panel-b"><span class="badge">For culture</span><h2 class="section-title small mt-16">Living tradition</h2><p class="prose">Keep Uzbek food traditions alive, one shared meal and one masterclass at a time.</p></div></div>
        </div>
        <div class="panel mt-24"><div class="panel-b">
          <h2 class="section-title small">How it works</h2>
          <p class="prose mt-12">Browse the <NuxtLink to="/dishes" class="text-link">dish catalog</NuxtLink>, watch <NuxtLink to="/reels" class="text-link">reels</NuxtLink>, pick a masterclass and reserve a session. After your experience, your booking appears in your cabinet and you can leave a review.</p>
        </div></div>
      </div>
    </section>
  </div>
</template>
