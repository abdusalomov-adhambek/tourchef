<script setup lang="ts">
import { useApi } from '~/composables/useApi'
import { money, duration, imageOr } from '~/composables/useFormat'

definePageMeta({ layout: 'default' })

const api = useApi()
const route = useRoute()
const router = useRouter()

const defaults = () => ({
  q: '',
  city: '',
  dish: '',
  min_price: '',
  max_price: '',
  duration: '',
  rating: '',
  sort: 'rec',
  page: 1
})
const filters = reactive<Record<string, any>>(defaults())
const items = ref<any[]>([])
const total = ref(0)
const pages = ref(1)
const loading = ref(true)
const error = ref('')
const cities = ref<string[]>([])
const dishes = ref<any[]>([])

function syncFromRoute() {
  const q = route.query
  for (const key of Object.keys(defaults())) {
    const value = q[key]
    filters[key] = value === undefined || value === null ? (key === 'sort' ? 'rec' : key === 'page' ? 1 : '') : String(value)
  }
}

function queryFromFilters(): Record<string, string> {
  const out: Record<string, string> = {}
  for (const [key, value] of Object.entries(filters)) {
    if (value !== '' && value !== undefined && value !== null) out[key] = String(value)
  }
  return out
}

async function fetchItems() {
  loading.value = true
  error.value = ''
  const query = queryFromFilters()
  delete query.page
  try {
    const result = await api.get<any>('/masterclasses', { ...query, page: filters.page, limit: 12 })
    items.value = Array.isArray(result?.items) ? result.items : []
    total.value = Number(result?.total || items.value.length)
    pages.value = Number(result?.pages || 1)
  } catch (e: any) {
    error.value = e?.message || 'Could not load masterclasses.'
    items.value = []
  } finally {
    loading.value = false
  }
}

async function apply() {
  filters.page = 1
  await router.push({ path: '/masterclasses', query: queryFromFilters() })
  await fetchItems()
}

async function reset() {
  Object.assign(filters, defaults())
  await router.push({ path: '/masterclasses' })
  await fetchItems()
}

async function goPage(page: number) {
  filters.page = page
  await router.push({ path: '/masterclasses', query: queryFromFilters() })
  await fetchItems()
  window?.scrollTo?.({ top: 0, behavior: 'smooth' })
}

watch(() => route.fullPath, () => {
  syncFromRoute()
  fetchItems()
})

onMounted(async () => {
  syncFromRoute()
  await Promise.all([
    fetchItems(),
    api.get<any[]>('/restaurants').then((r) => {
      cities.value = Array.from(new Set((Array.isArray(r) ? r : []).map((x) => x.city).filter(Boolean))).sort()
    }).catch(() => {}),
    api.get<any>('/dishes', { limit: 100 }).then((r) => {
      dishes.value = Array.isArray(r?.items) ? r.items : []
    }).catch(() => {})
  ])
})
</script>

<template>
  <div>
    <div class="container">
      <div class="page-head">
        <span class="eyebrow">Discover experiences</span>
        <h1>Explore Masterclasses</h1>
        <p>Cook traditional Uzbek dishes with local families in their home kitchens.</p>
      </div>

      <div class="catalog-layout">
        <FilterBar
          v-model="filters"
          :cities="cities"
          :dishes="dishes"
          :result-count="total"
          @apply="apply"
          @reset="reset"
        />

        <div>
          <div class="catalog-toolbar">
            <div class="chips">
              <button class="chip" :class="{ on: filters.sort === 'rec' }" type="button" @click="filters.sort = 'rec'; apply()">Recommended</button>
              <button class="chip" :class="{ on: filters.sort === 'new' }" type="button" @click="filters.sort = 'new'; apply()">Newest</button>
              <button class="chip" :class="{ on: filters.sort === 'rating' }" type="button" @click="filters.sort = 'rating'; apply()">Top rated</button>
              <button class="chip" :class="{ on: filters.sort === 'price_asc' }" type="button" @click="filters.sort = 'price_asc'; apply()">Price: low to high</button>
              <button class="chip" :class="{ on: filters.sort === 'price_desc' }" type="button" @click="filters.sort = 'price_desc'; apply()">Price: high to low</button>
            </div>
            <span class="muted text-sm">{{ total }} experience{{ total === 1 ? '' : 's' }}</span>
          </div>

          <AppAlert v-if="error" kind="error" title="Could not load results">{{ error }}</AppAlert>

          <div v-if="loading" class="stack">
            <div v-for="n in 4" :key="n" class="list-card">
              <div class="lc-media skeleton" />
              <div><div class="skeleton" style="height: 22px; width: 60%; margin-bottom: 12px" /><div class="skeleton" style="height: 12px; width: 90%" /></div>
            </div>
          </div>

          <template v-else-if="items.length">
            <article v-for="mc in items" :key="mc.id" class="list-card">
              <div class="lc-media">
                <NuxtLink :to="`/masterclasses/${mc.slug || mc.id}`">
                  <img :src="imageOr(mc.image_url || mc.dish?.image_url)" :alt="mc.title" loading="lazy" />
                </NuxtLink>
              </div>
              <div>
                <span class="badge neutral">{{ mc.city || 'Uzbekistan' }}</span>
                <h3><NuxtLink :to="`/masterclasses/${mc.slug || mc.id}`">{{ mc.title }}</NuxtLink></h3>
                <p class="muted text-sm line-clamp-2" style="margin: 0">{{ mc.description }}</p>
                <div class="lc-meta">
                  <span v-if="mc.restaurant?.name"><AppIcon name="restaurant" :size="15" /> {{ mc.restaurant.name }}</span>
                  <span v-if="mc.dish?.name"><AppIcon name="dish" :size="15" /> {{ mc.dish.name }}</span>
                  <span><AppIcon name="clock" :size="15" /> {{ duration(mc.duration_min) }}</span>
                  <span v-if="mc.rating" class="rating"><AppIcon name="star" :size="15" /> {{ Number(mc.rating).toFixed(1) }}</span>
                </div>
                <NuxtLink :to="`/masterclasses/${mc.slug || mc.id}`" class="btn btn-outline btn-sm">View details</NuxtLink>
              </div>
              <div class="lc-side">
                <div>
                  <div class="price">{{ money(mc.price) }} <small>/ person</small></div>
                  <div class="muted text-xs mt-8">{{ mc.difficulty || 'All levels' }}</div>
                </div>
                <NuxtLink :to="`/masterclasses/${mc.slug || mc.id}#booking`" class="btn btn-primary btn-sm">Book now</NuxtLink>
              </div>
            </article>

            <Pagination :page="filters.page" :pages="pages" :total="total" @update:page="goPage" />
          </template>

          <EmptyState
            v-else
            title="No experiences match your filters"
            message="Try changing the city, dish or price range — or reset all filters."
            icon="search"
            action-to="/masterclasses"
            action-label="Clear filters"
          />
        </div>
      </div>
    </div>
  </div>
</template>
