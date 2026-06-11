<script setup lang="ts">
import type { RouteLocationRaw } from "vue-router"
import type { AoiActionAppearance, AoiIntent } from "~/types/ui"

type ButtonSize = "sm" | "md" | "lg"
type LinkTarget = "_blank" | "_parent" | "_self" | "_top" | (string & {})

const props = withDefaults(defineProps<{
  appearance?: AoiActionAppearance
  intent?: AoiIntent
  size?: ButtonSize
  icon?: string
  trailingIcon?: string
  loading?: boolean
  disabled?: boolean
  ariaLabel?: string
  ariaPressed?: boolean
  external?: boolean
  href?: RouteLocationRaw
  noRel?: boolean
  rel?: string | null
  target?: LinkTarget | null
  to?: RouteLocationRaw
  type?: "button" | "submit" | "reset"
}>(), {
  appearance: "solid",
  intent: "primary",
  size: "md",
  icon: undefined,
  trailingIcon: undefined,
  loading: false,
  disabled: false,
  ariaLabel: undefined,
  ariaPressed: undefined,
  external: undefined,
  href: undefined,
  noRel: false,
  rel: undefined,
  target: undefined,
  to: undefined,
  type: "button"
})

const emit = defineEmits<{
  click: [event: MouseEvent]
}>()

const tagName = computed(() => {
  const map: Record<AoiActionAppearance, string> = {
    elevated: "md-elevated-button",
    outline: "md-outlined-button",
    plain: "md-text-button",
    soft: "md-filled-tonal-button",
    solid: "md-filled-button"
  }

  return map[props.appearance]
})

const hasLink = computed(() => Boolean(props.to || props.href))
const resolvedIcon = computed(() => props.loading ? "loader-circle" : props.icon)
const hasTrailingIcon = computed(() => Boolean(props.trailingIcon && !props.loading))

function onClick(event: MouseEvent) {
  emit("click", event)
}
</script>

<template>
  <AoiLink
    v-if="hasLink && !disabled && !loading"
    class="aoi-button-link"
    :aria-label="ariaLabel"
    :aria-pressed="ariaPressed"
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
      class="aoi-button"
      :class="`aoi-button--${size}`"
      :data-aoi-appearance="appearance"
      :data-aoi-intent="intent"
      aria-hidden="true"
      tabindex="-1"
      :type="type"
      :trailing-icon="hasTrailingIcon || undefined"
    >
      <AoiIcon
        v-if="resolvedIcon"
        slot="icon"
        :class="{ 'aoi-spin': loading }"
        :name="resolvedIcon"
        decorative
      />
      <AoiIcon
        v-if="hasTrailingIcon && trailingIcon"
        slot="icon"
        :name="trailingIcon"
        decorative
      />
      <slot />
    </component>
  </AoiLink>
  <component
    v-else
    :is="tagName"
    class="aoi-button"
    :class="`aoi-button--${size}`"
    :data-aoi-appearance="appearance"
    :data-aoi-intent="intent"
    :aria-label="ariaLabel"
    :aria-pressed="ariaPressed"
    :disabled="disabled || loading || undefined"
    :type="type"
    :trailing-icon="hasTrailingIcon || undefined"
    @click="onClick"
  >
    <AoiIcon
      v-if="resolvedIcon"
      slot="icon"
      :class="{ 'aoi-spin': loading }"
      :name="resolvedIcon"
      decorative
    />
    <AoiIcon
      v-if="hasTrailingIcon && trailingIcon"
      slot="icon"
      :name="trailingIcon"
      decorative
    />
    <slot />
  </component>
</template>

<style scoped>
.aoi-spin {
  animation: aoi-spin 900ms linear infinite;
}

.aoi-button-link {
  display: inline-flex;
  color: inherit;
  text-decoration: none;
}

.aoi-button-link > .aoi-button {
  pointer-events: none;
}

.aoi-button {
  --aoi-action-color: var(--aoi-intent-primary-color);
  --aoi-action-on-solid: var(--aoi-intent-primary-on-solid);
  --aoi-action-solid-bg: var(--aoi-intent-primary-solid-bg);
  --aoi-action-solid-bg-hover: var(--aoi-intent-primary-solid-bg-hover);
  --aoi-action-solid-bg-pressed: var(--aoi-intent-primary-solid-bg-pressed);
  --aoi-action-soft-bg: var(--aoi-intent-primary-soft-bg);
  --aoi-action-soft-bg-hover: var(--aoi-intent-primary-soft-bg-hover);
  --aoi-action-soft-bg-pressed: var(--aoi-intent-primary-soft-bg-pressed);
  --aoi-action-border: var(--aoi-intent-primary-border);
  --aoi-action-plain-hover: var(--aoi-intent-primary-plain-hover);
  --aoi-action-plain-pressed: var(--aoi-intent-primary-plain-pressed);
  --md-filled-button-container-color: var(--aoi-action-solid-bg);
  --md-filled-button-focus-container-color: var(--aoi-action-solid-bg-hover);
  --md-filled-button-hover-container-color: var(--aoi-action-solid-bg-hover);
  --md-filled-button-pressed-container-color: var(--aoi-action-solid-bg-pressed);
  --md-filled-button-label-text-color: var(--aoi-action-on-solid);
  --md-filled-button-focus-label-text-color: var(--aoi-action-on-solid);
  --md-filled-button-hover-label-text-color: var(--aoi-action-on-solid);
  --md-filled-button-pressed-label-text-color: var(--aoi-action-on-solid);
  --md-filled-button-icon-color: var(--aoi-action-on-solid);
  --md-filled-button-focus-icon-color: var(--aoi-action-on-solid);
  --md-filled-button-hover-icon-color: var(--aoi-action-on-solid);
  --md-filled-button-pressed-icon-color: var(--aoi-action-on-solid);
  --md-filled-tonal-button-container-color: var(--aoi-action-soft-bg);
  --md-filled-tonal-button-focus-container-color: var(--aoi-action-soft-bg-hover);
  --md-filled-tonal-button-hover-container-color: var(--aoi-action-soft-bg-hover);
  --md-filled-tonal-button-pressed-container-color: var(--aoi-action-soft-bg-pressed);
  --md-filled-tonal-button-label-text-color: var(--aoi-action-color);
  --md-filled-tonal-button-focus-label-text-color: var(--aoi-action-color);
  --md-filled-tonal-button-hover-label-text-color: var(--aoi-action-color);
  --md-filled-tonal-button-pressed-label-text-color: var(--aoi-action-color);
  --md-filled-tonal-button-icon-color: var(--aoi-action-color);
  --md-filled-tonal-button-focus-icon-color: var(--aoi-action-color);
  --md-filled-tonal-button-hover-icon-color: var(--aoi-action-color);
  --md-filled-tonal-button-pressed-icon-color: var(--aoi-action-color);
  --md-outlined-button-outline-color: var(--aoi-action-border);
  --md-outlined-button-hover-outline-color: var(--aoi-action-border);
  --md-outlined-button-focus-outline-color: var(--aoi-action-border);
  --md-outlined-button-pressed-outline-color: var(--aoi-action-border);
  --md-outlined-button-label-text-color: var(--aoi-action-color);
  --md-outlined-button-focus-label-text-color: var(--aoi-action-color);
  --md-outlined-button-hover-label-text-color: var(--aoi-action-color);
  --md-outlined-button-pressed-label-text-color: var(--aoi-action-color);
  --md-outlined-button-icon-color: var(--aoi-action-color);
  --md-outlined-button-focus-icon-color: var(--aoi-action-color);
  --md-outlined-button-hover-icon-color: var(--aoi-action-color);
  --md-outlined-button-pressed-icon-color: var(--aoi-action-color);
  --md-outlined-button-hover-state-layer-color: var(--aoi-action-color);
  --md-outlined-button-focus-state-layer-color: var(--aoi-action-color);
  --md-outlined-button-pressed-state-layer-color: var(--aoi-action-color);
  --md-text-button-label-text-color: var(--aoi-action-color);
  --md-text-button-focus-label-text-color: var(--aoi-action-color);
  --md-text-button-hover-label-text-color: var(--aoi-action-color);
  --md-text-button-pressed-label-text-color: var(--aoi-action-color);
  --md-text-button-icon-color: var(--aoi-action-color);
  --md-text-button-focus-icon-color: var(--aoi-action-color);
  --md-text-button-hover-icon-color: var(--aoi-action-color);
  --md-text-button-pressed-icon-color: var(--aoi-action-color);
  --md-text-button-hover-state-layer-color: var(--aoi-action-color);
  --md-text-button-focus-state-layer-color: var(--aoi-action-color);
  --md-text-button-pressed-state-layer-color: var(--aoi-action-color);
  --md-elevated-button-container-color: var(--aoi-action-soft-bg);
  --md-elevated-button-focus-container-color: var(--aoi-action-soft-bg-hover);
  --md-elevated-button-hover-container-color: var(--aoi-action-soft-bg-hover);
  --md-elevated-button-pressed-container-color: var(--aoi-action-soft-bg-pressed);
  --md-elevated-button-label-text-color: var(--aoi-action-color);
  --md-elevated-button-focus-label-text-color: var(--aoi-action-color);
  --md-elevated-button-hover-label-text-color: var(--aoi-action-color);
  --md-elevated-button-pressed-label-text-color: var(--aoi-action-color);
  --md-elevated-button-icon-color: var(--aoi-action-color);
  --md-elevated-button-focus-icon-color: var(--aoi-action-color);
  --md-elevated-button-hover-icon-color: var(--aoi-action-color);
  --md-elevated-button-pressed-icon-color: var(--aoi-action-color);
}

.aoi-button[data-aoi-intent="secondary"] {
  --aoi-action-color: var(--aoi-intent-secondary-color);
  --aoi-action-on-solid: var(--aoi-intent-secondary-on-solid);
  --aoi-action-solid-bg: var(--aoi-intent-secondary-solid-bg);
  --aoi-action-solid-bg-hover: var(--aoi-intent-secondary-solid-bg-hover);
  --aoi-action-solid-bg-pressed: var(--aoi-intent-secondary-solid-bg-pressed);
  --aoi-action-soft-bg: var(--aoi-intent-secondary-soft-bg);
  --aoi-action-soft-bg-hover: var(--aoi-intent-secondary-soft-bg-hover);
  --aoi-action-soft-bg-pressed: var(--aoi-intent-secondary-soft-bg-pressed);
  --aoi-action-border: var(--aoi-intent-secondary-border);
  --aoi-action-plain-hover: var(--aoi-intent-secondary-plain-hover);
  --aoi-action-plain-pressed: var(--aoi-intent-secondary-plain-pressed);
}

.aoi-button[data-aoi-intent="neutral"] {
  --aoi-action-color: var(--aoi-intent-neutral-color);
  --aoi-action-on-solid: var(--aoi-intent-neutral-on-solid);
  --aoi-action-solid-bg: var(--aoi-intent-neutral-solid-bg);
  --aoi-action-solid-bg-hover: var(--aoi-intent-neutral-solid-bg-hover);
  --aoi-action-solid-bg-pressed: var(--aoi-intent-neutral-solid-bg-pressed);
  --aoi-action-soft-bg: var(--aoi-intent-neutral-soft-bg);
  --aoi-action-soft-bg-hover: var(--aoi-intent-neutral-soft-bg-hover);
  --aoi-action-soft-bg-pressed: var(--aoi-intent-neutral-soft-bg-pressed);
  --aoi-action-border: var(--aoi-intent-neutral-border);
  --aoi-action-plain-hover: var(--aoi-intent-neutral-plain-hover);
  --aoi-action-plain-pressed: var(--aoi-intent-neutral-plain-pressed);
}

.aoi-button[data-aoi-intent="success"] {
  --aoi-action-color: var(--aoi-intent-success-color);
  --aoi-action-on-solid: var(--aoi-intent-success-on-solid);
  --aoi-action-solid-bg: var(--aoi-intent-success-solid-bg);
  --aoi-action-solid-bg-hover: var(--aoi-intent-success-solid-bg-hover);
  --aoi-action-solid-bg-pressed: var(--aoi-intent-success-solid-bg-pressed);
  --aoi-action-soft-bg: var(--aoi-intent-success-soft-bg);
  --aoi-action-soft-bg-hover: var(--aoi-intent-success-soft-bg-hover);
  --aoi-action-soft-bg-pressed: var(--aoi-intent-success-soft-bg-pressed);
  --aoi-action-border: var(--aoi-intent-success-border);
  --aoi-action-plain-hover: var(--aoi-intent-success-plain-hover);
  --aoi-action-plain-pressed: var(--aoi-intent-success-plain-pressed);
}

.aoi-button[data-aoi-intent="warning"] {
  --aoi-action-color: var(--aoi-intent-warning-color);
  --aoi-action-on-solid: var(--aoi-intent-warning-on-solid);
  --aoi-action-solid-bg: var(--aoi-intent-warning-solid-bg);
  --aoi-action-solid-bg-hover: var(--aoi-intent-warning-solid-bg-hover);
  --aoi-action-solid-bg-pressed: var(--aoi-intent-warning-solid-bg-pressed);
  --aoi-action-soft-bg: var(--aoi-intent-warning-soft-bg);
  --aoi-action-soft-bg-hover: var(--aoi-intent-warning-soft-bg-hover);
  --aoi-action-soft-bg-pressed: var(--aoi-intent-warning-soft-bg-pressed);
  --aoi-action-border: var(--aoi-intent-warning-border);
  --aoi-action-plain-hover: var(--aoi-intent-warning-plain-hover);
  --aoi-action-plain-pressed: var(--aoi-intent-warning-plain-pressed);
}

.aoi-button[data-aoi-intent="danger"] {
  --aoi-action-color: var(--aoi-intent-danger-color);
  --aoi-action-on-solid: var(--aoi-intent-danger-on-solid);
  --aoi-action-solid-bg: var(--aoi-intent-danger-solid-bg);
  --aoi-action-solid-bg-hover: var(--aoi-intent-danger-solid-bg-hover);
  --aoi-action-solid-bg-pressed: var(--aoi-intent-danger-solid-bg-pressed);
  --aoi-action-soft-bg: var(--aoi-intent-danger-soft-bg);
  --aoi-action-soft-bg-hover: var(--aoi-intent-danger-soft-bg-hover);
  --aoi-action-soft-bg-pressed: var(--aoi-intent-danger-soft-bg-pressed);
  --aoi-action-border: var(--aoi-intent-danger-border);
  --aoi-action-plain-hover: var(--aoi-intent-danger-plain-hover);
  --aoi-action-plain-pressed: var(--aoi-intent-danger-plain-pressed);
}

.aoi-button[data-aoi-intent="info"] {
  --aoi-action-color: var(--aoi-intent-info-color);
  --aoi-action-on-solid: var(--aoi-intent-info-on-solid);
  --aoi-action-solid-bg: var(--aoi-intent-info-solid-bg);
  --aoi-action-solid-bg-hover: var(--aoi-intent-info-solid-bg-hover);
  --aoi-action-solid-bg-pressed: var(--aoi-intent-info-solid-bg-pressed);
  --aoi-action-soft-bg: var(--aoi-intent-info-soft-bg);
  --aoi-action-soft-bg-hover: var(--aoi-intent-info-soft-bg-hover);
  --aoi-action-soft-bg-pressed: var(--aoi-intent-info-soft-bg-pressed);
  --aoi-action-border: var(--aoi-intent-info-border);
  --aoi-action-plain-hover: var(--aoi-intent-info-plain-hover);
  --aoi-action-plain-pressed: var(--aoi-intent-info-plain-pressed);
}

.aoi-button[data-aoi-appearance="plain"][data-aoi-intent="secondary"] {
  --md-text-button-hover-state-layer-color: var(--aoi-text);
  --md-text-button-focus-state-layer-color: var(--aoi-text);
  --md-text-button-pressed-state-layer-color: var(--aoi-text);
  --md-text-button-hover-state-layer-opacity: .06;
  --md-text-button-focus-state-layer-opacity: .08;
  --md-text-button-pressed-state-layer-opacity: .1;
}

@keyframes aoi-spin {
  to {
    rotate: 360deg;
  }
}
</style>


