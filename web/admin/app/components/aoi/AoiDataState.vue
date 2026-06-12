<script setup lang="ts">
import type { AoiIntent } from "~/types/ui"

const props = withDefaults(defineProps<{
  as?: string
  description?: string
  icon?: string
  intent?: Extract<AoiIntent, "danger" | "info" | "neutral" | "success" | "warning">
  loading?: boolean
  title: string
}>(), {
  as: "div",
  description: undefined,
  icon: "info",
  intent: "neutral",
  loading: false
})

const resolvedIcon = computed(() => props.loading ? "loader-circle" : props.icon)
</script>

<template>
  <component
    :is="props.as"
    class="aoi-data-state"
    :class="[
      `aoi-data-state--${props.intent}`,
      { 'aoi-data-state--loading': props.loading }
    ]"
  >
    <span class="aoi-data-state__icon" aria-hidden="true">
      <AoiIcon :name="resolvedIcon" :size="20" decorative />
    </span>
    <div class="aoi-data-state__copy">
      <strong>{{ props.title }}</strong>
      <p v-if="props.description">{{ props.description }}</p>
    </div>
    <div v-if="$slots.actions" class="aoi-data-state__actions">
      <slot name="actions" />
    </div>
  </component>
</template>

<style scoped>
.aoi-data-state {
  display: grid;
  min-height: var(--aoi-admin-data-state-min-height);
  min-width: 0;
  place-items: center;
  gap: var(--aoi-admin-card-gap);
  border: 1px dashed var(--aoi-data-state-border, var(--aoi-admin-border));
  border-radius: var(--aoi-radius-card);
  background: var(--aoi-data-state-bg, var(--aoi-admin-surface-muted));
  color: var(--aoi-data-state-color, var(--aoi-admin-text-muted));
  padding: var(--aoi-admin-data-state-padding);
  text-align: center;
}

.aoi-data-state__icon {
  display: inline-grid;
  width: var(--aoi-admin-data-state-icon-size);
  height: var(--aoi-admin-data-state-icon-size);
  place-items: center;
  border-radius: var(--aoi-radius-control);
  background: color-mix(in srgb, currentColor 12%, transparent);
  color: inherit;
}

.aoi-data-state--loading .aoi-data-state__icon {
  animation: aoi-data-state-spin 1.2s linear infinite;
}

.aoi-data-state__copy {
  display: grid;
  max-width: var(--aoi-admin-data-state-copy-max-width);
  gap: var(--aoi-admin-card-copy-gap);
}

.aoi-data-state__copy strong,
.aoi-data-state__copy p {
  margin: 0;
}

.aoi-data-state__copy strong {
  color: var(--aoi-admin-text);
  font-size: var(--aoi-admin-data-state-title-size);
}

.aoi-data-state__copy p {
  line-height: 1.6;
  overflow-wrap: anywhere;
}

.aoi-data-state__actions {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  gap: var(--aoi-admin-card-gap);
}

.aoi-data-state--success {
  --aoi-data-state-bg: var(--aoi-intent-success-soft-bg);
  --aoi-data-state-border: var(--aoi-intent-success-border);
  --aoi-data-state-color: var(--aoi-intent-success-color);
}

.aoi-data-state--warning {
  --aoi-data-state-bg: var(--aoi-intent-warning-soft-bg);
  --aoi-data-state-border: var(--aoi-intent-warning-border);
  --aoi-data-state-color: var(--aoi-intent-warning-color);
}

.aoi-data-state--danger {
  --aoi-data-state-bg: var(--aoi-intent-danger-soft-bg);
  --aoi-data-state-border: var(--aoi-intent-danger-border);
  --aoi-data-state-color: var(--aoi-intent-danger-color);
}

.aoi-data-state--info {
  --aoi-data-state-bg: var(--aoi-intent-info-soft-bg);
  --aoi-data-state-border: var(--aoi-intent-info-border);
  --aoi-data-state-color: var(--aoi-intent-info-color);
}

@keyframes aoi-data-state-spin {
  to {
    transform: rotate(360deg);
  }
}
</style>
