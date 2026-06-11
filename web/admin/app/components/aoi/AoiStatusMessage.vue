<script setup lang="ts">
withDefaults(defineProps<{
  message?: string | null
  tone?: "danger" | "info" | "success" | "warning"
}>(), {
  message: undefined,
  tone: "info"
})
</script>

<template>
  <div v-if="message" class="aoi-status" :class="`aoi-status--${tone}`">
    <AoiIcon :name="tone === 'danger' ? 'circle-alert' : tone === 'success' ? 'circle-check' : 'info'" decorative />
    <span>{{ message }}</span>
  </div>
</template>

<style scoped>
.aoi-status {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  border: 1px solid var(--aoi-border);
  border-radius: var(--aoi-radius-control);
  line-height: 1.6;
  padding: 10px 12px;
}

.aoi-status--info {
  background: color-mix(in srgb, var(--aoi-info) 8%, white);
  color: var(--aoi-info);
}

.aoi-status--success {
  background: color-mix(in srgb, var(--aoi-success) 10%, white);
  color: var(--aoi-success);
}

.aoi-status--warning {
  background: color-mix(in srgb, var(--aoi-sun-50) 16%, white);
  color: var(--aoi-warning);
}

.aoi-status--danger {
  background: color-mix(in srgb, var(--aoi-danger) 8%, white);
  color: var(--aoi-danger);
}
</style>
