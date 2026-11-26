<script setup lang="ts">
import { money, imageOr, duration } from '~/composables/useFormat'

defineProps<{ items: any[]; loading?: boolean }>()
</script>

<template>
  <div>
    <LoadingState v-if="loading" :cards="2" label="Masterclasslar yuklanayapti…" />
    <div v-else-if="items.length" class="rf-mc-list">
      <div v-for="mc in items" :key="mc.id" class="rf-mc">
        <NuxtLink :to="`/masterclasses/${mc.slug || mc.id}`" class="rf-mc-media">
          <img :src="imageOr(mc.image_url || mc.dish?.image_url)" :alt="mc.title" />
        </NuxtLink>
        <div class="rf-mc-body">
          <NuxtLink :to="`/masterclasses/${mc.slug || mc.id}`" class="rf-mc-title">{{ mc.title }}</NuxtLink>
          <small class="muted">{{ mc.restaurant?.name || mc.rest_name || 'Family host' }} · {{ mc.city }}</small>
          <div class="rf-mc-meta">
            <span v-if="mc.rating">★ {{ mc.rating }}</span>
            <span v-if="mc.duration_min">{{ duration(mc.duration_min) }}</span>
            <span v-if="mc.sessions?.[0]">Next: {{ new Date(mc.sessions[0].starts_at).toLocaleDateString('en-GB', { day: 'numeric', month: 'short' }) }}</span>
          </div>
        </div>
        <div class="rf-mc-side">
          <b>{{ money(mc.price) }}</b>
          <NuxtLink :to="`/masterclasses/${mc.slug || mc.id}#booking`" class="btn btn-primary btn-sm">Book Now</NuxtLink>
        </div>
      </div>
    </div>
    <EmptyState v-else title="Bu hududda masterclass yo'q" message="Tez orada yangi tajribalar qo'shiladi." icon="book" />
  </div>
</template>

<style scoped>
.rf-mc-list { display: grid; gap: 10px; }
.rf-mc { display: grid; grid-template-columns: 72px minmax(0,1fr) auto; gap: 12px; align-items: center; padding: 9px; border: 1px solid var(--border); border-radius: 12px; background: #fff; }
.rf-mc-media img { width: 72px; height: 62px; object-fit: cover; border-radius: 8px; }
.rf-mc-title { display: block; font-size: 14px; font-weight: 700; }
.rf-mc-title:hover { color: var(--accent); }
.rf-mc-body small { font-size: 12px; }
.rf-mc-meta { display: flex; gap: 10px; flex-wrap: wrap; margin-top: 4px; font-size: 11px; color: var(--text-2); }
.rf-mc-side { display: grid; gap: 6px; justify-items: end; }
.rf-mc-side b { font-size: 14px; }
@media (max-width: 640px) { .rf-mc { grid-template-columns: 56px minmax(0,1fr); } .rf-mc-side { grid-column: 1 / -1; display: flex; justify-content: space-between; align-items: center; } }
</style>
