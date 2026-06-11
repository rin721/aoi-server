<script setup lang="ts">
import type { RouteLocationRaw } from "vue-router"
import type { AoiActionAppearance, AoiIntent } from "~/types/ui"

type IconButtonSize = "sm" | "md" | "lg"
type LinkTarget = "_blank" | "_parent" | "_self" | "_top" | (string & {})

const props = withDefaults(defineProps<{
  icon: string
  label: string
  appearance?: AoiActionAppearance
  intent?: AoiIntent
  size?: IconButtonSize
  active?: boolean
  disabled?: boolean
  external?: boolean
  href?: RouteLocationRaw
  noRel?: boolean
  rel?: string | null
  target?: LinkTarget | null
  to?: RouteLocationRaw
}>(), {
  appearance: "plain",
  intent: "secondary",
  size: "md",
  active: false,
  disabled: false,
  external: undefined,
  href: undefined,
  noRel: false,
  rel: undefined,
  target: undefined,
  to: undefined
})

const emit = defineEmits<{
  click: [event: MouseEvent]
}>()

const tagName = computed(() => {
  const map: Record<AoiActionAppearance, string> = {
    elevated: "md-filled-icon-button",
    outline: "md-outlined-icon-button",
    plain: "md-icon-button",
    soft: "md-filled-tonal-icon-button",
    solid: "md-filled-icon-button"
  }

  return map[props.appearance]
})

const iconSize = computed(() => {
  const map: Record<IconButtonSize, number> = {
    sm: 18,
    md: 21,
    lg: 24
  }

  return map[props.size]
})

const hasLink = computed(() => Boolean(props.to || props.href))
const ariaCurrent = computed(() => props.active && hasLink.value ? "page" : undefined)

function onClick(event: MouseEvent) {
  emit("click", event)
}
</script>

<template>
  <AoiLink
    v-if="hasLink && !disabled"
    class="aoi-icon-button-link"
    :aria-current="ariaCurrent"
    :aria-label="label"
    :external="external"
    :href="href"
    :no-rel="noRel"
    :rel="rel"
    :target="target"
    :to="to"
    @click="onClick"
  >
    <component
      :is="tagName"
      class="aoi-icon-button"
      :class="{ 'aoi-icon-button--active': active }"
      :data-aoi-appearance="appearance"
      :data-aoi-intent="intent"
      aria-hidden="true"
      :selected="active || undefined"
      tabindex="-1"
      :toggle="active || undefined"
    >
      <AoiIcon :name="icon" :size="iconSize" decorative />
    </component>
  </AoiLink>
  <component
    v-else
    :is="tagName"
    class="aoi-icon-button"
    :class="{ 'aoi-icon-button--active': active }"
    :data-aoi-appearance="appearance"
    :data-aoi-intent="intent"
    :aria-label="label"
    :disabled="disabled || undefined"
    :selected="active || undefined"
    :toggle="active || undefined"
    @click="onClick"
  >
    <AoiIcon :name="icon" :size="iconSize" decorative />
  </component>
</template>

<style scoped>
.aoi-icon-button-link {
  display: inline-flex;
  color: inherit;
  text-decoration: none;
}

.aoi-icon-button-link > .aoi-icon-button {
  pointer-events: none;
}

.aoi-icon-button {
  --aoi-action-color: var(--aoi-intent-secondary-color);
  --aoi-action-on-solid: var(--aoi-intent-secondary-on-solid);
  --aoi-action-solid-bg: var(--aoi-intent-secondary-solid-bg);
  --aoi-action-soft-bg: var(--aoi-intent-secondary-soft-bg);
  --aoi-action-soft-bg-hover: var(--aoi-intent-secondary-soft-bg-hover);
  --aoi-action-soft-bg-pressed: var(--aoi-intent-secondary-soft-bg-pressed);
  --aoi-action-border: var(--aoi-intent-secondary-border);
  --md-icon-button-icon-color: var(--aoi-action-color);
  --md-icon-button-hover-icon-color: var(--aoi-action-color);
  --md-icon-button-focus-icon-color: var(--aoi-action-color);
  --md-icon-button-pressed-icon-color: var(--aoi-action-color);
  --md-icon-button-hover-state-layer-color: var(--aoi-action-color);
  --md-icon-button-focus-state-layer-color: var(--aoi-action-color);
  --md-icon-button-pressed-state-layer-color: var(--aoi-action-color);
  --md-filled-icon-button-container-color: var(--aoi-action-solid-bg);
  --md-filled-icon-button-icon-color: var(--aoi-action-on-solid);
  --md-filled-icon-button-hover-icon-color: var(--aoi-action-on-solid);
  --md-filled-icon-button-focus-icon-color: var(--aoi-action-on-solid);
  --md-filled-icon-button-pressed-icon-color: var(--aoi-action-on-solid);
  --md-filled-tonal-icon-button-container-color: var(--aoi-action-soft-bg);
  --md-filled-tonal-icon-button-hover-container-color: var(--aoi-action-soft-bg-hover);
  --md-filled-tonal-icon-button-focus-container-color: var(--aoi-action-soft-bg-hover);
  --md-filled-tonal-icon-button-pressed-container-color: var(--aoi-action-soft-bg-pressed);
  --md-filled-tonal-icon-button-icon-color: var(--aoi-action-color);
  --md-filled-tonal-icon-button-hover-icon-color: var(--aoi-action-color);
  --md-filled-tonal-icon-button-focus-icon-color: var(--aoi-action-color);
  --md-filled-tonal-icon-button-pressed-icon-color: var(--aoi-action-color);
  --md-outlined-icon-button-outline-color: var(--aoi-action-border);
  --md-outlined-icon-button-hover-outline-color: var(--aoi-action-border);
  --md-outlined-icon-button-focus-outline-color: var(--aoi-action-border);
  --md-outlined-icon-button-pressed-outline-color: var(--aoi-action-border);
  --md-outlined-icon-button-icon-color: var(--aoi-action-color);
  --md-outlined-icon-button-hover-icon-color: var(--aoi-action-color);
  --md-outlined-icon-button-focus-icon-color: var(--aoi-action-color);
  --md-outlined-icon-button-pressed-icon-color: var(--aoi-action-color);
}

.aoi-icon-button[data-aoi-intent="primary"] {
  --aoi-action-color: var(--aoi-intent-primary-color);
  --aoi-action-on-solid: var(--aoi-intent-primary-on-solid);
  --aoi-action-solid-bg: var(--aoi-intent-primary-solid-bg);
  --aoi-action-soft-bg: var(--aoi-intent-primary-soft-bg);
  --aoi-action-soft-bg-hover: var(--aoi-intent-primary-soft-bg-hover);
  --aoi-action-soft-bg-pressed: var(--aoi-intent-primary-soft-bg-pressed);
  --aoi-action-border: var(--aoi-intent-primary-border);
}

.aoi-icon-button[data-aoi-intent="neutral"] {
  --aoi-action-color: var(--aoi-intent-neutral-color);
  --aoi-action-on-solid: var(--aoi-intent-neutral-on-solid);
  --aoi-action-solid-bg: var(--aoi-intent-neutral-solid-bg);
  --aoi-action-soft-bg: var(--aoi-intent-neutral-soft-bg);
  --aoi-action-soft-bg-hover: var(--aoi-intent-neutral-soft-bg-hover);
  --aoi-action-soft-bg-pressed: var(--aoi-intent-neutral-soft-bg-pressed);
  --aoi-action-border: var(--aoi-intent-neutral-border);
}

.aoi-icon-button[data-aoi-intent="success"] {
  --aoi-action-color: var(--aoi-intent-success-color);
  --aoi-action-on-solid: var(--aoi-intent-success-on-solid);
  --aoi-action-solid-bg: var(--aoi-intent-success-solid-bg);
  --aoi-action-soft-bg: var(--aoi-intent-success-soft-bg);
  --aoi-action-soft-bg-hover: var(--aoi-intent-success-soft-bg-hover);
  --aoi-action-soft-bg-pressed: var(--aoi-intent-success-soft-bg-pressed);
  --aoi-action-border: var(--aoi-intent-success-border);
}

.aoi-icon-button[data-aoi-intent="warning"] {
  --aoi-action-color: var(--aoi-intent-warning-color);
  --aoi-action-on-solid: var(--aoi-intent-warning-on-solid);
  --aoi-action-solid-bg: var(--aoi-intent-warning-solid-bg);
  --aoi-action-soft-bg: var(--aoi-intent-warning-soft-bg);
  --aoi-action-soft-bg-hover: var(--aoi-intent-warning-soft-bg-hover);
  --aoi-action-soft-bg-pressed: var(--aoi-intent-warning-soft-bg-pressed);
  --aoi-action-border: var(--aoi-intent-warning-border);
}

.aoi-icon-button[data-aoi-intent="danger"] {
  --aoi-action-color: var(--aoi-intent-danger-color);
  --aoi-action-on-solid: var(--aoi-intent-danger-on-solid);
  --aoi-action-solid-bg: var(--aoi-intent-danger-solid-bg);
  --aoi-action-soft-bg: var(--aoi-intent-danger-soft-bg);
  --aoi-action-soft-bg-hover: var(--aoi-intent-danger-soft-bg-hover);
  --aoi-action-soft-bg-pressed: var(--aoi-intent-danger-soft-bg-pressed);
  --aoi-action-border: var(--aoi-intent-danger-border);
}

.aoi-icon-button[data-aoi-intent="info"] {
  --aoi-action-color: var(--aoi-intent-info-color);
  --aoi-action-on-solid: var(--aoi-intent-info-on-solid);
  --aoi-action-solid-bg: var(--aoi-intent-info-solid-bg);
  --aoi-action-soft-bg: var(--aoi-intent-info-soft-bg);
  --aoi-action-soft-bg-hover: var(--aoi-intent-info-soft-bg-hover);
  --aoi-action-soft-bg-pressed: var(--aoi-intent-info-soft-bg-pressed);
  --aoi-action-border: var(--aoi-intent-info-border);
}

.aoi-icon-button--active {
  color: var(--aoi-action-color);
  box-shadow: inset 0 0 0 1px var(--aoi-action-border);
}

.aoi-icon-button[data-aoi-appearance="elevated"] {
  box-shadow: var(--aoi-shadow-sm);
}
</style>


