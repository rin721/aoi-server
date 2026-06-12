<script setup lang="ts">
import type { AoiKeyValueItem, AoiKeyValueListDensity, AoiKeyValueListLayout } from "~/types/ui"

const props = withDefaults(defineProps<{
  columns?: number
  density?: AoiKeyValueListDensity
  items: AoiKeyValueItem[]
  layout?: AoiKeyValueListLayout
}>(), {
  columns: 2,
  density: "default",
  layout: "rows"
})

const listStyle = computed(() => ({
  "--aoi-kv-columns": `repeat(${Math.max(1, props.columns)}, minmax(0, 1fr))`
}))

function displayValue(item: AoiKeyValueItem) {
  if (item.value === undefined || item.value === null || item.value === "") {
    return "-"
  }
  if (typeof item.value === "boolean") {
    return item.value ? "true" : "false"
  }
  return String(item.value)
}
</script>

<template>
  <dl
    class="aoi-key-value-list"
    :class="[
      `aoi-key-value-list--${props.layout}`,
      `aoi-key-value-list--density-${props.density}`
    ]"
    :style="listStyle"
  >
    <div v-for="item in props.items" :key="`${item.label}-${item.meta || item.value || ''}`" class="aoi-key-value-list__item">
      <dt>
        <AoiIcon v-if="item.icon" :name="item.icon" :size="15" decorative />
        <span>{{ item.label }}</span>
        <small v-if="item.meta">{{ item.meta }}</small>
      </dt>
      <dd>
        <span
          class="aoi-key-value-list__value"
          :class="{
            'aoi-key-value-list__value--empty': displayValue(item) === '-',
            'aoi-key-value-list__value--mono': item.monospace,
            'aoi-key-value-list__value--secret': item.secret
          }"
        >
          {{ displayValue(item) }}
        </span>
        <AoiMetaPill v-if="item.badge !== undefined" appearance="outline" :intent="item.intent || 'neutral'">
          {{ item.badge }}
        </AoiMetaPill>
        <AoiMetaPill v-if="item.secret" appearance="soft" intent="warning">
          敏感
        </AoiMetaPill>
        <small v-if="item.description" class="aoi-key-value-list__description">{{ item.description }}</small>
      </dd>
    </div>
  </dl>
</template>

<style scoped>
.aoi-key-value-list {
  display: grid;
  min-width: 0;
  margin: 0;
  gap: var(--aoi-admin-kv-gap);
}

.aoi-key-value-list--cards {
  grid-template-columns: var(--aoi-kv-columns);
}

.aoi-key-value-list__item {
  min-width: 0;
}

.aoi-key-value-list--rows .aoi-key-value-list__item {
  display: grid;
  min-height: var(--aoi-admin-kv-row-min-height);
  align-items: start;
  gap: var(--aoi-admin-kv-row-gap);
  grid-template-columns: minmax(var(--aoi-admin-kv-label-min-width), var(--aoi-admin-kv-label-max-width)) minmax(0, 1fr);
  border-top: 1px solid var(--aoi-admin-border-soft);
  padding: var(--aoi-admin-kv-row-padding);
}

.aoi-key-value-list--rows .aoi-key-value-list__item:first-child {
  border-top: 0;
}

.aoi-key-value-list--cards .aoi-key-value-list__item {
  display: grid;
  gap: var(--aoi-admin-kv-card-gap);
  border: 1px solid var(--aoi-admin-border);
  border-radius: var(--aoi-radius-card);
  background: var(--aoi-admin-surface-muted);
  padding: var(--aoi-admin-kv-card-padding);
}

.aoi-key-value-list__item dt,
.aoi-key-value-list__item dd {
  min-width: 0;
}

.aoi-key-value-list__item dt {
  display: grid;
  align-content: start;
  gap: var(--aoi-admin-kv-label-gap);
  color: var(--aoi-admin-text-muted);
  font-size: var(--aoi-admin-kv-meta-size);
  font-weight: 780;
}

.aoi-key-value-list__item dt span {
  color: var(--aoi-admin-text);
  font-size: var(--aoi-admin-kv-label-size);
  overflow-wrap: anywhere;
}

.aoi-key-value-list__item dt small {
  color: var(--aoi-admin-text-muted);
  font-family: var(--aoi-font-mono);
  font-weight: 640;
  overflow-wrap: anywhere;
}

.aoi-key-value-list__item dd {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: flex-end;
  gap: var(--aoi-admin-kv-value-gap);
  margin: 0;
}

.aoi-key-value-list--cards .aoi-key-value-list__item dd {
  justify-content: flex-start;
}

.aoi-key-value-list__value {
  display: inline-flex;
  max-width: 100%;
  min-height: var(--aoi-admin-kv-value-min-height);
  align-items: center;
  border: 1px solid var(--aoi-admin-border);
  border-radius: var(--aoi-radius-field);
  background: var(--aoi-admin-surface);
  color: var(--aoi-admin-text);
  font-weight: 720;
  line-height: 1.45;
  overflow-wrap: anywhere;
  padding: var(--aoi-admin-kv-value-padding);
  white-space: pre-wrap;
}

.aoi-key-value-list__value--mono {
  font-family: var(--aoi-font-mono);
  font-size: var(--aoi-admin-kv-mono-size);
}

.aoi-key-value-list__value--empty {
  color: var(--aoi-admin-text-muted);
}

.aoi-key-value-list__value--secret {
  border-color: var(--aoi-intent-warning-border);
  background: var(--aoi-intent-warning-soft-bg);
  color: var(--aoi-intent-warning-color);
}

.aoi-key-value-list__description {
  flex-basis: 100%;
  color: var(--aoi-admin-text-muted);
  line-height: 1.5;
}

.aoi-key-value-list--density-compact {
  gap: var(--aoi-admin-kv-gap-compact);
}

@media (max-width: 760px) {
  .aoi-key-value-list--cards,
  .aoi-key-value-list--rows .aoi-key-value-list__item {
    grid-template-columns: 1fr;
  }

  .aoi-key-value-list__item dd {
    justify-content: flex-start;
  }
}
</style>
