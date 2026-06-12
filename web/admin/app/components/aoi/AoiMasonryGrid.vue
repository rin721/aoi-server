<script setup lang="ts">
import type { AoiAdminInfoLayout, AoiContentGridGap } from "~/types/ui"

const props = withDefaults(defineProps<{
  as?: string
  gap?: Extract<AoiContentGridGap, "compact" | "normal">
  layout?: AoiAdminInfoLayout
  minWidth?: string
}>(), {
  as: "section",
  gap: "normal",
  minWidth: "var(--aoi-admin-info-card-min-width)"
})

const { state } = useAdminUiPreferences()
const resolvedLayout = computed(() => props.layout || state.infoLayout || "adaptive")
const gridStyle = computed(() => ({
  "--aoi-masonry-grid-min-width": props.minWidth
}))
</script>

<template>
  <component
    :is="props.as"
    class="aoi-masonry-grid"
    :class="[
      `aoi-masonry-grid--${resolvedLayout}`,
      `aoi-masonry-grid--gap-${props.gap}`
    ]"
    :style="gridStyle"
  >
    <slot />
  </component>
</template>

<style scoped>
.aoi-masonry-grid {
  min-width: 0;
}

.aoi-masonry-grid--gap-normal {
  --aoi-masonry-grid-gap: var(--aoi-admin-panel-gap);
}

.aoi-masonry-grid--gap-compact {
  --aoi-masonry-grid-gap: var(--aoi-admin-panel-gap-compact);
}

.aoi-masonry-grid--adaptive {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(min(100%, var(--aoi-masonry-grid-min-width)), 1fr));
  gap: var(--aoi-masonry-grid-gap);
  align-items: start;
}

.aoi-masonry-grid--masonry {
  column-gap: var(--aoi-masonry-grid-gap);
  column-width: var(--aoi-masonry-grid-min-width);
}

.aoi-masonry-grid--masonry > :deep(*) {
  display: inline-grid;
  width: 100%;
  margin: 0 0 var(--aoi-masonry-grid-gap);
  break-inside: avoid;
}

.aoi-masonry-grid--list {
  display: grid;
  gap: var(--aoi-masonry-grid-gap);
}

@media (max-width: 760px) {
  .aoi-masonry-grid--adaptive,
  .aoi-masonry-grid--list {
    grid-template-columns: 1fr;
  }

  .aoi-masonry-grid--masonry {
    columns: auto;
  }
}
</style>
