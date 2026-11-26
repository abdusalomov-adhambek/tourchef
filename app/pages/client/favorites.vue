<script setup lang="ts">
import { useApi } from '~/composables/useApi'
import { useFavorites } from '~/composables/useFavorites'
import { useToast } from '~/composables/useToast'

definePageMeta({ layout: 'cabinet' })

const api = useApi()
const favorites = useFavorites()
const toast = useToast()
const route = useRoute()
const loading = ref(true)
const error = ref('')
const catalog = ref<any[]>([])

const items = computed(() =>
  favorites.ids.value.map((id) =>
    catalog.value.find((m) => Number(m.id) === Number(id)) ||
    favorites.records.value.find((r) => Number(r?.id) === Number(id) || Number(r?.masterclass?.id) === Number(id))?.masterclass ||
    { id, title: `Masterclass #${id}`, price: 0 }
  )
)

async function remove(id: number) {
  if (!window.confirm('Remove this masterclass from favorites?')) return
  const ok = await favorites.toggle(id, route.fullPath)
  if (ok) toast.success('Removed from favorites.')
}

onMounted(async () => {
  await favorites.load(true)
  try {
    const data = await api.get<any>('/masterclasses', { limit: 100 })
    catalog.value = Array.isArray(data?.items) ? data.items : []
  } catch (e: any) {
    error.value = e?.message || 'Could not load masterclass details.'
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div>
    <div class="cab-top">
      <div><h1>Favorites</h1><p>Experiences you saved for later.</p></div>
      <NuxtLink to="/masterclasses" class="btn btn-outline btn-sm">Explore experiences</NuxtLink>
    </div>

    <AppAlert v-if="error" kind="error" title="Some details could not be loaded">{{ error }}</AppAlert>
    <LoadingState v-if="loading" :cards="3" label="Loading favorites…" />
    <div v-else-if="items.length" class="grid-cards">
      <div v-for="mc in items" :key="mc.id" class="relative">
        <MasterclassCard :mc="mc" />
        <button class="btn btn-outline btn-sm mt-12" type="button" @click="remove(mc.id)">
          <AppIcon name="trash" :size="15" /> Remove
        </button>
      </div>
    </div>
    <EmptyState v-else title="No favorites yet" message="Tap the heart on any masterclass to save it here." icon="heart" action-to="/masterclasses" action-label="Find experiences" />
  </div>
</template>
