<script setup lang="ts">
import type { AoiActionAppearance, AoiIntent } from "~/types/ui"

const props = withDefaults(defineProps<{
  appearance?: Extract<AoiActionAppearance, "outline" | "plain" | "soft">
  icon?: string
  intent?: AoiIntent
  size?: "sm" | "md"
}>(), {
  appearance: "outline",
  icon: undefined,
  intent: "secondary",
  size: "sm",
})
</script>

<template>
  <span
    class="aoi-meta-pill"
    :class="[
      `aoi-meta-pill--${props.size}`,
      `aoi-meta-pill--${props.appearance}`,
      `aoi-meta-pill--intent-${props.intent}`
    ]"
  >
    <AoiIcon v-if="props.icon" :name="props.icon" :size="props.size === 'sm' ? 14 : 16" decorative />
    <slot />
  </span>
</template>

<style scoped>
.aoi-meta-pill {
  display: inline-flex;
  width: fit-content;
  max-width: 100%;
  align-items: center;
  gap: 5px;
  border: 1px solid var(--aoi-border);
  border-radius: var(--aoi-radius-control);
  background: var(--aoi-surface);
  color: var(--aoi-pill-color, var(--aoi-intent-secondary-color));
  font-weight: 760;
  line-height: 1.25;
}

.aoi-meta-pill--sm {
  min-height: 26px;
  font-size: 12px;
  padding: 4px 8px;
}

.aoi-meta-pill--md {
  min-height: 32px;
  font-size: 13px;
  padding: 5px 10px;
}

.aoi-meta-pill--intent-primary {
  --aoi-pill-color: var(--aoi-intent-primary-color);
  --aoi-pill-bg: var(--aoi-intent-primary-soft-bg);
  --aoi-pill-border: var(--aoi-intent-primary-border);
}

.aoi-meta-pill--intent-secondary {
  --aoi-pill-color: var(--aoi-intent-secondary-color);
  --aoi-pill-bg: var(--aoi-intent-secondary-soft-bg);
  --aoi-pill-border: var(--aoi-intent-secondary-border);
}

.aoi-meta-pill--intent-neutral {
  --aoi-pill-color: var(--aoi-intent-neutral-color);
  --aoi-pill-bg: var(--aoi-intent-neutral-soft-bg);
  --aoi-pill-border: var(--aoi-intent-neutral-border);
}

.aoi-meta-pill--intent-success {
  --aoi-pill-color: var(--aoi-intent-success-color);
  --aoi-pill-bg: var(--aoi-intent-success-soft-bg);
  --aoi-pill-border: var(--aoi-intent-success-border);
}

.aoi-meta-pill--intent-warning {
  --aoi-pill-color: var(--aoi-intent-warning-color);
  --aoi-pill-bg: var(--aoi-intent-warning-soft-bg);
  --aoi-pill-border: var(--aoi-intent-warning-border);
}

.aoi-meta-pill--intent-danger {
  --aoi-pill-color: var(--aoi-intent-danger-color);
  --aoi-pill-bg: var(--aoi-intent-danger-soft-bg);
  --aoi-pill-border: var(--aoi-intent-danger-border);
}

.aoi-meta-pill--intent-info {
  --aoi-pill-color: var(--aoi-intent-info-color);
  --aoi-pill-bg: var(--aoi-intent-info-soft-bg);
  --aoi-pill-border: var(--aoi-intent-info-border);
}

.aoi-meta-pill--soft {
  border-color: transparent;
  background: var(--aoi-pill-bg);
}

.aoi-meta-pill--outline {
  border-color: var(--aoi-pill-border);
}

.aoi-meta-pill--plain {
  border-color: transparent;
  background: transparent;
}
</style>


