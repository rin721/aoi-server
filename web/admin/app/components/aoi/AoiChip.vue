<script setup lang="ts">
import type { AoiActionAppearance, AoiIntent } from "~/types/ui"

const props = withDefaults(defineProps<{
  label: string
  appearance?: Extract<AoiActionAppearance, "outline" | "plain" | "soft">
  icon?: string
  intent?: AoiIntent
  selected?: boolean
  removable?: boolean
  removeLabel?: string
  disabled?: boolean
}>(), {
  appearance: "outline",
  disabled: false,
  icon: undefined,
  intent: "secondary",
  removable: false,
  removeLabel: undefined,
  selected: false
})

const emit = defineEmits<{
  remove: []
}>()

function remove() {
  if (!props.disabled) {
    emit("remove")
  }
}
</script>

<template>
  <span
    class="aoi-chip"
    :class="[
      `aoi-chip--${appearance}`,
      `aoi-chip--intent-${intent}`,
      { 'aoi-chip--selected': selected, 'aoi-chip--disabled': disabled }
    ]"
  >
    <AoiIcon v-if="icon" :name="icon" :size="14" decorative />
    <span class="aoi-chip__label">{{ label }}</span>
    <button
      v-if="removable"
      class="aoi-chip__remove"
      type="button"
      :aria-label="removeLabel || label"
      :disabled="disabled || undefined"
      @click="remove"
    >
      <AoiIcon name="x" :size="14" decorative />
    </button>
  </span>
</template>

<style scoped>
.aoi-chip {
  display: inline-flex;
  min-height: 30px;
  max-width: 100%;
  align-items: center;
  gap: 6px;
  border: 1px solid var(--aoi-border);
  border-radius: var(--aoi-radius-sm);
  background: var(--aoi-surface-solid);
  color: var(--aoi-chip-color, var(--aoi-intent-secondary-color));
  font-size: 12px;
  font-weight: 800;
  padding: 4px 7px 4px 9px;
}

.aoi-chip--intent-primary {
  --aoi-chip-color: var(--aoi-intent-primary-color);
  --aoi-chip-bg: var(--aoi-intent-primary-soft-bg);
  --aoi-chip-border: var(--aoi-intent-primary-border);
}

.aoi-chip--intent-secondary {
  --aoi-chip-color: var(--aoi-intent-secondary-color);
  --aoi-chip-bg: var(--aoi-intent-secondary-soft-bg);
  --aoi-chip-border: var(--aoi-intent-secondary-border);
}

.aoi-chip--intent-neutral {
  --aoi-chip-color: var(--aoi-intent-neutral-color);
  --aoi-chip-bg: var(--aoi-intent-neutral-soft-bg);
  --aoi-chip-border: var(--aoi-intent-neutral-border);
}

.aoi-chip--intent-success {
  --aoi-chip-color: var(--aoi-intent-success-color);
  --aoi-chip-bg: var(--aoi-intent-success-soft-bg);
  --aoi-chip-border: var(--aoi-intent-success-border);
}

.aoi-chip--intent-warning {
  --aoi-chip-color: var(--aoi-intent-warning-color);
  --aoi-chip-bg: var(--aoi-intent-warning-soft-bg);
  --aoi-chip-border: var(--aoi-intent-warning-border);
}

.aoi-chip--intent-danger {
  --aoi-chip-color: var(--aoi-intent-danger-color);
  --aoi-chip-bg: var(--aoi-intent-danger-soft-bg);
  --aoi-chip-border: var(--aoi-intent-danger-border);
}

.aoi-chip--intent-info {
  --aoi-chip-color: var(--aoi-intent-info-color);
  --aoi-chip-bg: var(--aoi-intent-info-soft-bg);
  --aoi-chip-border: var(--aoi-intent-info-border);
}

.aoi-chip--soft {
  border-color: transparent;
  background: var(--aoi-chip-bg);
}

.aoi-chip--outline {
  border-color: var(--aoi-chip-border);
}

.aoi-chip--plain {
  border-color: transparent;
  background: transparent;
}

.aoi-chip--selected {
  border-color: var(--aoi-chip-border);
  background: var(--aoi-chip-bg);
}

.aoi-chip--disabled {
  opacity: .58;
}

.aoi-chip__label {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.aoi-chip__remove {
  display: inline-grid;
  width: 22px;
  height: 22px;
  place-items: center;
  border: 0;
  border-radius: var(--aoi-radius-xs);
  background: transparent;
  color: currentColor;
  cursor: pointer;
  padding: 0;
}

.aoi-chip__remove:hover {
  background: var(--aoi-chip-bg);
}

.aoi-chip__remove:disabled {
  cursor: not-allowed;
}
</style>


