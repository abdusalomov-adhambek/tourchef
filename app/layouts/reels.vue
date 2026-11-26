<script setup lang="ts">
import { homeForRole } from '~/composables/useAuth'

const auth = useAuth()
useHead({ bodyAttrs: { 'data-area': 'reels' } })
</script>

<template>
  <div class="reels-page">
    <div class="reels-top">
      <NuxtLink to="/" class="brand" aria-label="TourChef home">
        <img src="/logo.svg" alt="TourChef" class="logo-img" width="138" height="46" />
      </NuxtLink>
      <nav class="reels-nav" aria-label="Reels navigation">
        <NuxtLink to="/masterclasses">Masterclasses</NuxtLink>
        <NuxtLink to="/restaurants">Restaurants</NuxtLink>
        <NuxtLink v-if="auth.user.value" :to="homeForRole(auth.user.value.role)">My cabinet</NuxtLink>
        <NuxtLink v-else to="/login">Sign in</NuxtLink>
      </nav>
    </div>
    <slot />
  </div>
</template>
