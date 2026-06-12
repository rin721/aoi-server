<script setup lang="ts">
import type { AoiIntent, AoiSurfacePadding } from "~/types/ui"

const props = withDefaults(defineProps<{
  as?: string
  badge?: number | string
  badgeIntent?: AoiIntent
  description?: string | null
  flush?: boolean
  icon?: string
  padding?: AoiSurfacePadding
  title?: string
}>(), {
  as: "article",
  badge: undefined,
  badgeIntent: "neutral",
  description: undefined,
  flush: false,
  icon: undefined,
  padding: "md",
  title: undefined
})

const hasHeader = computed(() => Boolean(props.title || props.description || props.icon || props.badge !== undefined))
</script>

<template>
  <component
    :is="props.as"
    class="aoi-admin-card"
    :class="[
      `aoi-admin-card--padding-${props.padding}`,
      { 'aoi-admin-card--flush': props.flush }
    ]"
  >
    <slot name="header">
      <header v-if="hasHeader || $slots.actions" class="aoi-admin-card__header">
        <div class="aoi-admin-card__title-group">
          <span v-if="props.icon" class="aoi-admin-card__icon" aria-hidden="true">
            <AoiIcon :name="props.icon" :size="18" decorative />
          </span>
          <div class="aoi-admin-card__copy">
            <h2 v-if="props.title">{{ props.title }}</h2>
            <p v-if="props.description">{{ props.description }}</p>
          </div>
        </div>
        <div v-if="props.badge !== undefined || $slots.actions" class="aoi-admin-card__actions">
          <AoiMetaPill v-if="props.badge !== undefined" :intent="props.badgeIntent" appearance="outline">
            {{ props.badge }}
          </AoiMetaPill>
          <slot name="actions" />
        </div>
      </header>
    </slot>

    <div class="aoi-admin-card__body">
      <slot />
    </div>
  </component>
</template>

<style scoped>
.aoi-admin-card {
  display: grid;
  min-width: 0;
  overflow: hidden;
  border: 1px solid var(--aoi-admin-border);
  border-radius: var(--aoi-radius-card);
  background: var(--aoi-admin-surface);
  box-shadow: var(--aoi-admin-shadow);
  color: var(--aoi-admin-text);
}

.aoi-admin-card__header {
  display: flex;
  min-width: 0;
  min-height: var(--aoi-admin-card-header-min-height);
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--aoi-admin-card-gap);
  border-bottom: 1px solid var(--aoi-admin-border);
  background: var(--aoi-admin-surface);
  padding: var(--aoi-admin-card-header-padding);
}

.aoi-admin-card__title-group,
.aoi-admin-card__actions {
  display: flex;
  min-width: 0;
  align-items: flex-start;
  gap: var(--aoi-admin-card-gap);
}

.aoi-admin-card__title-group {
  flex: 1 1 auto;
}

.aoi-admin-card__actions {
  flex: 0 0 auto;
  flex-wrap: wrap;
  justify-content: flex-end;
}

.aoi-admin-card__icon {
  display: inline-grid;
  width: var(--aoi-admin-card-icon-size);
  height: var(--aoi-admin-card-icon-size);
  place-items: center;
  border-radius: var(--aoi-radius-control);
  background: var(--aoi-intent-primary-soft-bg);
  color: var(--aoi-active-color);
  flex: 0 0 auto;
}

.aoi-admin-card__copy {
  display: grid;
  min-width: 0;
  gap: var(--aoi-admin-card-copy-gap);
}

.aoi-admin-card__copy h2,
.aoi-admin-card__copy p {
  margin: 0;
}

.aoi-admin-card__copy h2 {
  color: var(--aoi-admin-text);
  font-size: var(--aoi-admin-card-title-size);
  line-height: 1.35;
}

.aoi-admin-card__copy p {
  color: var(--aoi-admin-text-muted);
  line-height: 1.6;
  overflow-wrap: anywhere;
}

.aoi-admin-card__body {
  min-width: 0;
  padding: var(--aoi-admin-card-body-padding);
}

.aoi-admin-card--padding-none .aoi-admin-card__body,
.aoi-admin-card--flush .aoi-admin-card__body {
  padding: 0;
}

.aoi-admin-card--padding-sm .aoi-admin-card__body {
  padding: var(--aoi-admin-card-body-padding-sm);
}

.aoi-admin-card--padding-lg .aoi-admin-card__body {
  padding: var(--aoi-admin-card-body-padding-lg);
}

@media (max-width: 680px) {
  .aoi-admin-card__header,
  .aoi-admin-card__title-group,
  .aoi-admin-card__actions {
    align-items: flex-start;
  }

  .aoi-admin-card__header {
    flex-direction: column;
  }

  .aoi-admin-card__actions {
    justify-content: flex-start;
  }
}
</style>
