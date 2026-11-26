<script setup lang="ts">
import { useToast } from '~/composables/useToast'

const toast = useToast()
</script>

<template>
  <div class="toast-host" aria-live="polite" aria-atomic="true">
    <div
      v-for="t in toast.toasts.value"
      :key="t.id"
      class="toast"
      :class="t.kind === 'ok' ? 'ok' : t.kind === 'err' ? 'err' : ''"
      role="status"
    >
      <AppIcon :name="t.kind === 'err' ? 'warning' : t.kind === 'ok' ? 'check' : 'info'" :size="18" />
      <span>{{ t.message }}</span>
      <button class="t-close" type="button" aria-label="Dismiss" @click="toast.dismiss(t.id)">
        <AppIcon name="close" :size="14" />
      </button>
    </div>
  </div>
</template>
