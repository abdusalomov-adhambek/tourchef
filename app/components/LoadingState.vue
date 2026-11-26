<script setup lang="ts">
withDefaults(defineProps<{ label?: string; cards?: number }>(), { label: 'Loading experiences…', cards: 3 })
</script>

<template>
  <div class="loading-state" role="status" :aria-label="label">
    <p class="muted text-sm mb-16">{{ label }}</p>
    <div class="grid-cards">
      <div v-for="n in cards" :key="n" class="card">
        <div class="card-media skeleton" />
        <div class="card-body">
          <div class="skeleton" style="height: 12px; width: 55%; margin-bottom: 12px" />
          <div class="skeleton" style="height: 22px; width: 80%; margin-bottom: 10px" />
          <div class="skeleton" style="height: 12px; width: 100%; margin-bottom: 8px" />
          <div class="skeleton" style="height: 12px; width: 70%" />
        </div>
      </div>
    </div>
  </div>
</template>
