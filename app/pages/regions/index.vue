<script setup lang="ts">
import { useApi } from '~/composables/useApi'
import { imageOr } from '~/composables/useFormat'

definePageMeta({ layout: 'default' })
const api = useApi()
const regions = ref<any[]>([])
const loading = ref(true)
const error = ref('')

onMounted(async () => {
  try {
    const data = await api.get<any[]>('/v1/regions')
    regions.value = Array.isArray(data) ? data : []
  } catch (e: any) {
    error.value = e?.message || 'Mintaqalar yuklanmadi.'
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div class="section container">
    <span class="eyebrow">Culinary geography</span>
    <h1 class="page-title" style="margin:10px 0">Taste of Uzbekistan</h1>
    <p class="muted">Every region has its own flavor — explore the dishes and traditions of Uzbekistan.</p>
    <AppAlert v-if="error" kind="error" class="mt-16">{{ error }}</AppAlert>
    <LoadingState v-if="loading" :cards="6" label="Mintaqalar yuklanayapti…" />
    <div v-else-if="regions.length" class="grid-cards three mt-24">
      <NuxtLink v-for="r in regions" :key="r.slug" :to="`/regions/${r.slug}`" class="card">
        <div class="card-media"><img :src="imageOr(r.dish_image_url)" :alt="r.dish_name" /></div>
        <div class="card-body">
          <span class="eyebrow">{{ r.name }}</span>
          <h3 class="card-title">{{ r.dish_name }}</h3>
          <p class="card-desc muted">{{ r.description }}</p>
        </div>
      </NuxtLink>
    </div>
    <EmptyState v-else title="Mintaqalar yo'q" message="Ma'muriyat mintaqa ma'lumotlarini hali qo'shmadi." icon="pin" />
  </div>
</template>
