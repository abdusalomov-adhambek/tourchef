<script setup lang="ts">
withDefaults(defineProps<{ kind?: 'info' | 'success' | 'error' | 'warn'; title?: string }>(), {
  kind: 'info'
})
</script>

<template>
  <div class="alert" :class="kind" role="alert">
    <AppIcon
      :name="kind === 'success' ? 'check' : kind === 'error' || kind === 'warn' ? 'warning' : 'info'"
      :size="18"
      class="ic"
    />
    <div>
      <strong v-if="title">{{ title }}</strong>
      <div v-if="title && $slots.default" style="margin-top: 3px"><slot /></div>
      <slot v-else />
    </div>
  </div>
</template>
