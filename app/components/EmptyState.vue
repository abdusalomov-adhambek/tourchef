<script setup lang="ts">
withDefaults(
  defineProps<{ title?: string; message?: string; icon?: string; actionTo?: string; actionLabel?: string }>(),
  {
    title: 'Nothing here yet',
    message: 'Check back soon — there is more to discover.',
    icon: 'dish'
  }
)
</script>

<template>
  <div class="empty">
    <div class="ic"><AppIcon :name="icon" :size="30" /></div>
    <h3>{{ title }}</h3>
    <p class="muted" style="font-size: 14px; max-width: 42ch; margin: 0 auto">{{ message }}</p>
    <NuxtLink v-if="actionTo" :to="actionTo" class="btn btn-primary btn-sm mt-16">
      {{ actionLabel || 'Explore' }} <AppIcon name="arrow" :size="16" />
    </NuxtLink>
    <slot />
  </div>
</template>
