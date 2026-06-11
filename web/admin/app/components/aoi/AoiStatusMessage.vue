<script setup lang="ts">
import type { AoiIntent } from "~/types/ui"

const props = withDefaults(defineProps<{
  as?: string
  icon?: string
  message?: string
  intent?: Extract<AoiIntent, "danger" | "info" | "success" | "warning">
  tone?: Extract<AoiIntent, "danger" | "info" | "success" | "warning">
}>(), {
  as: "p",
  icon: undefined,
  message: undefined,
  intent: "info",
  tone: undefined
})

const resolvedIntent = computed(() => props.tone || props.intent)
</script>

<template>
  <component
    :is="props.as"
    v-if="props.message || $slots.default"
    class="aoi-status-message"
    :class="`aoi-status-message--${resolvedIntent}`"
  >
    <AoiIcon v-if="props.icon" :name="props.icon" :size="15" decorative />
    <slot>{{ props.message }}</slot>
  </component>
</template>

<style scoped>
.aoi-status-message {
  display: flex;
  align-items: flex-start;
  gap: 6px;
  margin: 0;
  border: 1px solid var(--aoi-border);
  border-radius: var(--aoi-radius-card);
  line-height: 1.7;
  padding: 10px 12px;
}

.aoi-status-message--success {
  border-color: var(--aoi-intent-success-border);
  background: var(--aoi-intent-success-soft-bg);
  color: var(--aoi-intent-success-color);
}

.aoi-status-message--danger {
  border-color: var(--aoi-intent-danger-border);
  background: var(--aoi-intent-danger-soft-bg);
  color: var(--aoi-intent-danger-color);
}

.aoi-status-message--warning {
  border-color: var(--aoi-intent-warning-border);
  background: var(--aoi-intent-warning-soft-bg);
  color: var(--aoi-intent-warning-color);
}

.aoi-status-message--info {
  border-color: var(--aoi-intent-info-border);
  background: var(--aoi-intent-info-soft-bg);
  color: var(--aoi-intent-info-color);
}
</style>


