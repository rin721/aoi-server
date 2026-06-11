<script setup lang="ts">
import type { AoiIntent, AoiRevealProp, AoiSurfaceKind, AoiSurfacePadding } from "~/types/ui"

const props = withDefaults(defineProps<{
  as?: string
  interactive?: boolean
  padding?: AoiSurfacePadding
  reveal?: AoiRevealProp
  selected?: boolean
  surface?: AoiSurfaceKind
  intent?: AoiIntent
}>(), {
  as: "div",
  interactive: false,
  padding: "md",
  reveal: false,
  selected: false,
  surface: "card",
  intent: "neutral"
})
</script>

<template>
  <component
    :is="props.as"
    v-aoi-reveal="props.reveal"
    class="aoi-surface"
    :class="[
      `aoi-surface--${props.surface}`,
      `aoi-surface--intent-${props.intent}`,
      `aoi-surface--padding-${props.padding}`,
      {
        'aoi-surface--interactive': props.interactive,
        'aoi-surface--selected': props.selected
      }
    ]"
  >
    <slot />
  </component>
</template>

<style scoped>
.aoi-surface {
  min-width: 0;
  --aoi-surface-intent-bg: var(--aoi-intent-neutral-soft-bg);
  --aoi-surface-intent-bg-hover: var(--aoi-intent-neutral-soft-bg-hover);
  --aoi-surface-intent-border: var(--aoi-intent-neutral-border);
  color: var(--aoi-text);
}

.aoi-surface--plain {
  background: transparent;
}

.aoi-surface--panel,
.aoi-surface--card,
.aoi-surface--state,
.aoi-surface--code,
.aoi-surface--toolbar {
  border: 1px solid var(--aoi-border);
  box-shadow: var(--aoi-shadow-sm);
}

.aoi-surface--panel {
  border-radius: var(--aoi-radius-container);
  background: var(--aoi-panel-bg);
}

.aoi-surface--card,
.aoi-surface--state,
.aoi-surface--toolbar {
  border-radius: var(--aoi-radius-card);
  background: var(--aoi-card-bg);
}

.aoi-surface--state {
  background: var(--aoi-surface);
}

.aoi-surface--code {
  overflow: hidden;
  border-radius: var(--aoi-radius-card);
  background: var(--aoi-bg);
}

.aoi-surface--toolbar {
  background: var(--aoi-surface);
}

.aoi-surface--padding-none {
  padding: 0;
}

.aoi-surface--padding-sm {
  padding: 10px;
}

.aoi-surface--padding-md {
  padding: var(--aoi-card-padding);
}

.aoi-surface--padding-lg {
  padding: var(--aoi-panel-padding);
}

.aoi-surface--intent-primary {
  --aoi-surface-intent-bg: color-mix(in srgb, var(--aoi-intent-primary-soft-bg) 70%, var(--aoi-card-bg));
  --aoi-surface-intent-bg-hover: var(--aoi-intent-primary-soft-bg-hover);
  --aoi-surface-intent-border: var(--aoi-intent-primary-border);
  border-color: var(--aoi-surface-intent-border);
  background: var(--aoi-surface-intent-bg);
}

.aoi-surface--intent-danger {
  --aoi-surface-intent-bg: color-mix(in srgb, var(--aoi-intent-danger-soft-bg) 70%, var(--aoi-card-bg));
  --aoi-surface-intent-bg-hover: var(--aoi-intent-danger-soft-bg-hover);
  --aoi-surface-intent-border: var(--aoi-intent-danger-border);
  border-color: var(--aoi-surface-intent-border);
}

.aoi-surface--intent-secondary {
  background: var(--aoi-surface-muted);
}

.aoi-surface--intent-success,
.aoi-surface--intent-warning,
.aoi-surface--intent-info {
  border-color: var(--aoi-surface-intent-border);
  background: var(--aoi-surface-intent-bg);
}

.aoi-surface--intent-success {
  --aoi-surface-intent-bg: color-mix(in srgb, var(--aoi-intent-success-soft-bg) 70%, var(--aoi-card-bg));
  --aoi-surface-intent-bg-hover: var(--aoi-intent-success-soft-bg-hover);
  --aoi-surface-intent-border: var(--aoi-intent-success-border);
}

.aoi-surface--intent-warning {
  --aoi-surface-intent-bg: color-mix(in srgb, var(--aoi-intent-warning-soft-bg) 70%, var(--aoi-card-bg));
  --aoi-surface-intent-bg-hover: var(--aoi-intent-warning-soft-bg-hover);
  --aoi-surface-intent-border: var(--aoi-intent-warning-border);
}

.aoi-surface--intent-info {
  --aoi-surface-intent-bg: color-mix(in srgb, var(--aoi-intent-info-soft-bg) 70%, var(--aoi-card-bg));
  --aoi-surface-intent-bg-hover: var(--aoi-intent-info-soft-bg-hover);
  --aoi-surface-intent-border: var(--aoi-intent-info-border);
}

.aoi-surface--interactive {
  transition:
    border-color var(--aoi-motion-fast) var(--aoi-ease-out),
    background-color var(--aoi-motion-fast) var(--aoi-ease-out),
    box-shadow var(--aoi-motion-fast) var(--aoi-ease-out),
    transform var(--aoi-motion-fast) var(--aoi-ease-out);
}

.aoi-surface--interactive:hover,
.aoi-surface--selected {
  border-color: var(--aoi-surface-intent-border);
  background: var(--aoi-surface-intent-bg-hover);
}

.aoi-surface--interactive:hover {
  transform: translate3d(0, -1px, 0);
}

@media (prefers-reduced-motion: reduce) {
  .aoi-surface--interactive {
    transition: none;
  }

  .aoi-surface--interactive:hover {
    transform: none;
  }
}
</style>


