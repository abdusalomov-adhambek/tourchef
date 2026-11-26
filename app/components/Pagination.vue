<script setup lang="ts">
const props = withDefaults(defineProps<{ page: number; pages: number; total?: number }>(), {
  page: 1,
  pages: 1
})
const emit = defineEmits<{ (e: 'update:page', value: number): void }>()

const visible = computed(() => {
  const result: (number | '…')[] = []
  const p = Math.max(1, props.page)
  const last = Math.max(1, props.pages)
  for (let i = 1; i <= last; i++) {
    if (i === 1 || i === last || Math.abs(i - p) <= 1) result.push(i)
    else if (result[result.length - 1] !== '…') result.push('…')
  }
  return result
})

function go(value: number | '…') {
  if (typeof value !== 'number') return
  if (value >= 1 && value <= props.pages) emit('update:page', value)
}
</script>

<template>
  <nav v-if="pages > 1" class="pagination" aria-label="Pagination">
    <button type="button" :disabled="page <= 1" aria-label="Previous page" @click="go(page - 1)">
      <AppIcon name="back" :size="16" />
    </button>
    <template v-for="(v, i) in visible" :key="i">
      <span v-if="v === '…'" class="pages">…</span>
      <button
        v-else
        type="button"
        :class="{ on: v === page }"
        :aria-current="v === page ? 'page' : undefined"
        @click="go(v)"
      >
        {{ v }}
      </button>
    </template>
    <button type="button" :disabled="page >= pages" aria-label="Next page" @click="go(page + 1)">
      <AppIcon name="chevron" :size="16" />
    </button>
    <span v-if="total !== undefined" class="pages">{{ total }} results</span>
  </nav>
</template>
