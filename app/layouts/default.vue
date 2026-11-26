<script setup lang="ts">
useHead({ bodyAttrs: { 'data-area': 'public' } })
</script>

<template>
  <div class="public-layout">
    <SiteHeader />
    <main>
      <slot />
    </main>
    <SiteFooter />
  </div>
</template>
