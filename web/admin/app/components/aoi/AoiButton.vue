<script setup lang="ts">
import type { RouteLocationRaw } from "vue-router"

const props = withDefaults(defineProps<{
  appearance?: "outline" | "plain" | "soft" | "solid"
  disabled?: boolean
  icon?: string
  intent?: "danger" | "neutral" | "primary" | "success" | "warning"
  loading?: boolean
  to?: RouteLocationRaw
  type?: "button" | "reset" | "submit"
}>(), {
  appearance: "solid",
  disabled: false,
  icon: undefined,
  intent: "primary",
  loading: false,
  to: undefined,
  type: "button"
})

const emit = defineEmits<{
  click: [event: MouseEvent]
}>()

const classes = computed(() => [
  "aoi-button",
  `aoi-button--${props.appearance}`,
  `aoi-button--${props.intent}`
])

function onClick(event: MouseEvent) {
  if (!props.disabled && !props.loading) {
    emit("click", event)
  }
}
</script>

<template>
  <NuxtLink v-if="to && !disabled" class="aoi-button-link" :to="to">
    <span :class="classes">
      <AoiIcon v-if="loading" name="loader-circle" class="aoi-button__spin" decorative />
      <AoiIcon v-else-if="icon" :name="icon" decorative />
      <slot />
    </span>
  </NuxtLink>
  <button
    v-else
    :class="classes"
    :disabled="disabled || loading"
    :type="type"
    @click="onClick"
  >
    <AoiIcon v-if="loading" name="loader-circle" class="aoi-button__spin" decorative />
    <AoiIcon v-else-if="icon" :name="icon" decorative />
    <slot />
  </button>
</template>

<style scoped>
.aoi-button,
.aoi-button-link {
  display: inline-flex;
  min-height: 36px;
  align-items: center;
  justify-content: center;
  gap: 8px;
  border: 1px solid transparent;
  border-radius: var(--aoi-radius-control);
  font-weight: 780;
  line-height: 1;
  padding: 0 13px;
  text-decoration: none;
  transition:
    background var(--aoi-motion-fast) var(--aoi-ease-out),
    border-color var(--aoi-motion-fast) var(--aoi-ease-out),
    color var(--aoi-motion-fast) var(--aoi-ease-out),
    transform var(--aoi-motion-fast) var(--aoi-ease-out);
}

.aoi-button {
  cursor: pointer;
}

.aoi-button:active,
.aoi-button-link:active {
  transform: translateY(1px);
}

.aoi-button:disabled {
  cursor: not-allowed;
  opacity: .58;
}

.aoi-button--primary {
  --button-color: var(--aoi-accent-60);
  --button-soft: var(--aoi-accent-10);
}

.aoi-button--neutral {
  --button-color: var(--aoi-text);
  --button-soft: var(--aoi-surface-muted);
}

.aoi-button--success {
  --button-color: var(--aoi-success);
  --button-soft: color-mix(in srgb, var(--aoi-success) 11%, white);
}

.aoi-button--warning {
  --button-color: var(--aoi-warning);
  --button-soft: color-mix(in srgb, var(--aoi-sun-50) 17%, white);
}

.aoi-button--danger {
  --button-color: var(--aoi-danger);
  --button-soft: color-mix(in srgb, var(--aoi-danger) 9%, white);
}

.aoi-button--solid {
  background: var(--button-color);
  color: white;
}

.aoi-button--solid:hover {
  background: color-mix(in srgb, var(--button-color) 88%, black);
}

.aoi-button--soft {
  border-color: color-mix(in srgb, var(--button-color) 26%, var(--aoi-border));
  background: var(--button-soft);
  color: var(--button-color);
}

.aoi-button--soft:hover {
  background: color-mix(in srgb, var(--button-color) 14%, white);
}

.aoi-button--outline {
  border-color: color-mix(in srgb, var(--button-color) 30%, var(--aoi-border));
  background: transparent;
  color: var(--button-color);
}

.aoi-button--outline:hover,
.aoi-button--plain:hover {
  background: var(--button-soft);
}

.aoi-button--plain {
  background: transparent;
  color: var(--button-color);
}

.aoi-button__spin {
  animation: aoi-spin 900ms linear infinite;
}

@keyframes aoi-spin {
  to {
    transform: rotate(360deg);
  }
}
</style>
