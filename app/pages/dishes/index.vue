<script setup lang="ts">
import { useApi } from '~/composables/useApi'

definePageMeta({ layout: 'default' })

const api = useApi()
const query = reactive({ q: '', category: '', region: '', sort: 'name' })
const items = ref<any[]>([])
const total = ref(0)
const limit = 12
const offset = ref(0)
const loading = ref(true)
const loadingMore = ref(false)
const error = ref('')

const categories = computed(() =>
  Array.from(new Set(items.value.map((d) => d.category).filter(Boolean))).sort()
)
const regions = computed(() =>
  Array.from(new Set(items.value.map((d) => d.region).filter(Boolean))).sort()
)

async function load(reset = true) {
  if (reset) {
    offset.value = 0
    items.value = []
  }
  loading.value = reset
  loadingMore.value = !reset
  error.value = ''
  try {
    const params: Record<string, any> = { ...query, limit, offset: offset.value }
    Object.keys(params).forEach((k) => {
      if (!params[k]) delete params[k]
    })
    const result = await api.get<any>('/dishes', params)
    const batch = Array.isArray(result?.items) ? result.items : []
    items.value = reset ? batch : [...items.value, ...batch]
    total.value = Number(result?.total || items.value.length)
    offset.value = items.value.length
  } catch (e: any) {
    error.value = e?.message || 'Could not load dishes.'
  } finally {
    loading.value = false
    loadingMore.value = false
  }
}

function apply() {
  load(true)
}

onMounted(() => load(true))
</script>

<template>
  <div>
    <div class="container">
      <div class="page-head">
        <span class="eyebrow">Uzbek food culture</span>
        <h1>Traditional Dishes</h1>
        <p>From plov to somsa — discover the dishes behind every culinary experience.</p>
      </div>

      <div class="panel mb-24">
        <div class="panel-b">
          <div class="toolbar">
            <input v-model="query.q" class="inp" type="search" placeholder="Search dishes…" @keyup.enter="apply" />
            <select v-model="query.region" class="select" style="width: auto; min-width: 160px" @change="apply">
              <option value="">All regions</option>
              <option v-for="r in regions" :key="r" :value="r">{{ r }}</option>
            </select>
            <select v-model="query.sort" class="select" style="width: auto; min-width: 160px" @change="apply">
              <option value="name">Name A–Z</option>
              <option value="new">Newest</option>
              <option value="rating">Top rated</option>
            </select>
            <button class="btn btn-primary" type="button" @click="apply"><AppIcon name="search" :size="18" /> Search</button>
          </div>
          <div class="chips mt-16">
            <button class="chip" :class="{ on: !query.category }" type="button" @click="query.category = ''; apply()">All categories</button>
            <button v-for="c in categories" :key="c" class="chip" :class="{ on: query.category === c }" type="button" @click="query.category = c; apply()">{{ c }}</button>
          </div>
        </div>
      </div>

      <AppAlert v-if="error" kind="error" title="Could not load dishes">{{ error }}</AppAlert>
      <LoadingState v-else-if="loading" :cards="4" label="Loading dishes…" />
      <template v-else-if="items.length">
        <div class="grid-cards four">
          <DishCard v-for="dish in items" :key="dish.id" :dish="dish" />
        </div>
        <div v-if="items.length < total" class="center-text mt-32">
          <button class="btn btn-outline" type="button" :disabled="loadingMore" :class="{ 'is-loading': loadingMore }" @click="load(false)">
            {{ loadingMore ? 'Loading…' : `Show more (${total - items.length} left)` }}
          </button>
        </div>
      </template>
      <EmptyState v-else title="No dishes found" message="Try another search or category — our catalog keeps growing." icon="dish" action-to="/dishes" action-label="Reset search" />
    </div>
  </div>
</template>
