<script setup lang="ts">
withDefaults(
  defineProps<{ cities?: string[]; dishes?: any[]; resultCount?: number }>(),
  { cities: () => [], dishes: () => [] }
)
const emit = defineEmits<{ (e: 'reset'): void; (e: 'apply'): void }>()
const model = defineModel<Record<string, any>>({ required: true })

function reset() {
  const keys = Object.keys(model.value || {})
  for (const key of keys) model.value[key] = ''
  emit('reset')
}
</script>

<template>
  <aside class="filter-box" aria-label="Masterclass filters">
    <h3>
      Filters
      <AppIcon name="filter" :size="18" />
    </h3>
    <p v-if="resultCount !== undefined" class="muted text-xs mt-8">{{ resultCount }} experiences found</p>

    <div class="filter-group">
      <h4>Search</h4>
      <input v-model="model.q" class="inp" type="search" placeholder="Title, family, dish…" @keyup.enter="emit('apply')" />
    </div>

    <div class="filter-group">
      <h4>City</h4>
      <select v-model="model.city" class="select">
        <option value="">All cities</option>
        <option v-for="city in cities" :key="city" :value="city">{{ city }}</option>
      </select>
    </div>

    <div class="filter-group">
      <h4>Dish</h4>
      <select v-model="model.dish" class="select">
        <option value="">All dishes</option>
        <option v-for="dish in dishes" :key="dish.id || dish.name" :value="dish.name || dish.slug">
          {{ dish.name }}
        </option>
      </select>
    </div>

    <div class="filter-group">
      <h4>Price per person</h4>
      <div class="row" style="gap: 8px">
        <input v-model="model.min_price" class="inp" type="number" min="0" placeholder="Min" style="width: 48%" />
        <input v-model="model.max_price" class="inp" type="number" min="0" placeholder="Max" style="width: 48%" />
      </div>
    </div>

    <div class="filter-group">
      <h4>Duration</h4>
      <select v-model="model.duration" class="select">
        <option value="">Any duration</option>
        <option value="1-2">1–2 hours</option>
        <option value="2-4">2–4 hours</option>
        <option value="4+">4+ hours</option>
      </select>
    </div>

    <div class="filter-group">
      <h4>Minimum rating</h4>
      <select v-model="model.rating" class="select">
        <option value="">Any rating</option>
        <option value="4">4.0 and up</option>
        <option value="4.5">4.5 and up</option>
        <option value="4.8">4.8 and up</option>
      </select>
    </div>

    <div class="filter-group">
      <h4>Sort by</h4>
      <select v-model="model.sort" class="select">
        <option value="rec">Recommended</option>
        <option value="price_asc">Price: low to high</option>
        <option value="price_desc">Price: high to low</option>
        <option value="rating">Highest rated</option>
        <option value="new">Newest</option>
      </select>
    </div>

    <div class="filter-actions">
      <button class="btn btn-primary btn-block" type="button" @click="emit('apply')">Apply filters</button>
      <button class="btn btn-outline btn-block" type="button" @click="reset">Reset</button>
    </div>
  </aside>
</template>
