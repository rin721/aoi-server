<script setup lang="ts">
interface SettingsShellItem {
  depth: "basic" | "all"
  icon: string
  label: string
  to: string
}

interface SettingsShellGroup {
  items: SettingsShellItem[]
  label: string
}

const props = defineProps<{
  activePath: string
  description: string
  depthItems: Array<{
    description?: string
    icon?: string
    label: string
    value: string
  }>
  depthLabel: string
  depthModelValue: string
  emptyText: string
  groups: SettingsShellGroup[]
  modelValue: string
  searchLabel: string
  searchPlaceholder: string
  title: string
}>()

const emit = defineEmits<{
  "update:depthModelValue": [value: string]
  "update:modelValue": [value: string]
}>()

const { t } = useI18n()
const query = computed({
  get: () => props.modelValue,
  set: (value: string) => emit("update:modelValue", value)
})
const depth = computed({
  get: () => props.depthModelValue,
  set: (value: string) => emit("update:depthModelValue", value)
})

const desktopGroups = ref<HTMLElement | null>(null)
const mobileItems = ref<HTMLElement | null>(null)
const desktopIndicatorStyle = ref<Record<string, string>>({})
const mobileIndicatorStyle = ref<Record<string, string>>({})
let indicatorFrame = 0
let indicatorResizeObserver: ResizeObserver | undefined

function resolveIndicatorStyle(container: HTMLElement | null, placement: "bottom" | "left") {
  const activeItem = container?.querySelector<HTMLElement>(".settings-shell-nav__item--active")

  if (!container || !activeItem) {
    return {
      "--settings-shell-nav-indicator-opacity": "0",
      "--settings-shell-nav-indicator-height": "0px",
      "--settings-shell-nav-indicator-width": "0px",
      "--settings-shell-nav-indicator-x": "0px",
      "--settings-shell-nav-indicator-y": "0px"
    }
  }

  const containerRect = container.getBoundingClientRect()
  const activeRect = activeItem.getBoundingClientRect()
  const indicatorWidth = placement === "bottom" ? activeRect.width / 4 : 3
  const indicatorHeight = placement === "bottom" ? 3 : 16
  const x = activeRect.left - containerRect.left + container.scrollLeft
    + (placement === "bottom" ? (activeRect.width - indicatorWidth) / 2 : 0)
  const y = activeRect.top - containerRect.top + container.scrollTop
    + (placement === "bottom" ? activeRect.height - indicatorHeight : activeRect.height / 2 - indicatorHeight / 2)

  return {
    "--settings-shell-nav-indicator-opacity": "1",
    "--settings-shell-nav-indicator-height": `${indicatorHeight}px`,
    "--settings-shell-nav-indicator-width": `${indicatorWidth}px`,
    "--settings-shell-nav-indicator-x": `${x}px`,
    "--settings-shell-nav-indicator-y": `${y}px`
  }
}

function updateIndicators() {
  desktopIndicatorStyle.value = resolveIndicatorStyle(desktopGroups.value, "left")
  mobileIndicatorStyle.value = resolveIndicatorStyle(mobileItems.value, "bottom")
}

function scheduleIndicatorUpdate() {
  if (!import.meta.client) {
    return
  }

  window.cancelAnimationFrame(indicatorFrame)
  indicatorFrame = window.requestAnimationFrame(updateIndicators)
}

watch(
  () => [props.activePath, props.groups, props.modelValue, props.depthModelValue],
  async () => {
    await nextTick()
    scheduleIndicatorUpdate()
  },
  { deep: true, flush: "post" }
)

onMounted(async () => {
  await nextTick()
  scheduleIndicatorUpdate()

  indicatorResizeObserver = new ResizeObserver(scheduleIndicatorUpdate)

  if (desktopGroups.value) {
    indicatorResizeObserver.observe(desktopGroups.value)
  }

  if (mobileItems.value) {
    indicatorResizeObserver.observe(mobileItems.value)
  }

  window.addEventListener("resize", scheduleIndicatorUpdate)
})

onBeforeUnmount(() => {
  window.cancelAnimationFrame(indicatorFrame)
  indicatorResizeObserver?.disconnect()
  window.removeEventListener("resize", scheduleIndicatorUpdate)
})
</script>

<template>
  <aside
    v-aoi-scroll-native
    class="settings-shell-nav"
    :aria-label="t('settings.shell.categoryAria')"
  >
    <div class="settings-shell-nav__intro">
      <span class="settings-shell-nav__mark" aria-hidden="true">
        <AoiIcon name="settings" :size="22" decorative />
      </span>
      <div>
        <h1>{{ props.title }}</h1>
        <p>{{ props.description }}</p>
      </div>
    </div>

    <AoiTextField
      v-model="query"
      class="settings-shell-nav__search-field"
      icon="search"
      :label="props.searchLabel"
      :placeholder="props.searchPlaceholder"
      appearance="outlined"
      type="search"
    />

    <AoiSegmentedControl
      v-model="depth"
      class="settings-shell-nav__depth"
      :items="props.depthItems"
      :aria-label="props.depthLabel"
      :columns="2"
    />

    <nav
      ref="desktopGroups"
      class="settings-shell-nav__groups"
      :aria-label="t('settings.shell.pagesAria')"
    >
      <span
        class="settings-shell-nav__indicator"
        :style="desktopIndicatorStyle"
        aria-hidden="true"
      />
      <section
        v-for="group in props.groups"
        :key="group.label"
        class="settings-shell-nav__group"
      >
        <h2>{{ group.label }}</h2>
        <AoiLink
          v-for="item in group.items"
          :key="item.to"
          class="settings-shell-nav__item"
          :class="{ 'settings-shell-nav__item--active': props.activePath === item.to }"
          :to="item.to"
        >
          <AoiIcon :name="item.icon" :size="17" decorative />
          <span>{{ item.label }}</span>
        </AoiLink>
      </section>

      <p v-if="props.groups.length === 0" class="settings-shell-nav__empty">
        {{ props.emptyText }}
      </p>
    </nav>
  </aside>

  <div class="settings-shell-nav-mobile">
    <AoiSegmentedControl
      v-model="depth"
      class="settings-shell-nav__depth"
      :items="props.depthItems"
      :aria-label="props.depthLabel"
      :columns="2"
    />

    <nav
      ref="mobileItems"
      v-aoi-scroll-native
      class="settings-shell-nav-mobile__items"
      :aria-label="t('settings.shell.pagesAria')"
    >
      <span
        class="settings-shell-nav__indicator"
        :style="mobileIndicatorStyle"
        aria-hidden="true"
      />
      <template
        v-for="group in props.groups"
        :key="group.label"
      >
        <AoiLink
          v-for="item in group.items"
          :key="item.to"
          class="settings-shell-nav__item"
          :class="{ 'settings-shell-nav__item--active': props.activePath === item.to }"
          :to="item.to"
        >
          <AoiIcon :name="item.icon" :size="17" decorative />
          <span>{{ item.label }}</span>
        </AoiLink>
      </template>

      <p v-if="props.groups.length === 0" class="settings-shell-nav__empty">
        {{ props.emptyText }}
      </p>
    </nav>
  </div>
</template>

<style scoped>
.settings-shell-nav {
  position: sticky;
  top: var(--aoi-settings-sticky-top);
  z-index: var(--aoi-z-sticky);
  display: grid;
  max-height: calc(100dvh - var(--aoi-settings-sticky-top) * 2);
  align-self: start;
  gap: var(--aoi-grid-gap-compact);
  overflow: auto;
  border: 1px solid var(--aoi-border);
  border-radius: var(--aoi-radius-container);
  background: var(--aoi-panel-bg);
  box-shadow: var(--aoi-shadow-sm);
  padding: var(--aoi-card-padding);
}

.settings-shell-nav__search-field {
  --md-filled-text-field-container-height: var(--aoi-control-height-md);
  --md-outlined-text-field-container-height: var(--aoi-control-height-md);
}

.settings-shell-nav__intro {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr);
  gap: var(--aoi-grid-gap-compact);
  align-items: start;
}

.settings-shell-nav__mark {
  display: inline-grid;
  width: var(--aoi-settings-shell-mark-size);
  height: var(--aoi-settings-shell-mark-size);
  place-items: center;
  border-radius: var(--aoi-radius-round);
  background: var(--aoi-accent-10);
  color: var(--aoi-accent-60);
}

.settings-shell-nav__intro h1,
.settings-shell-nav__intro p,
.settings-shell-nav__group h2 {
  margin: 0;
}

.settings-shell-nav__intro h1 {
  font-size: var(--aoi-settings-shell-title-size);
  line-height: 1.1;
}

.settings-shell-nav__intro p,
.settings-shell-nav__empty {
  color: var(--aoi-text-muted);
  line-height: 1.7;
}

.settings-shell-nav__groups {
  position: relative;
  display: grid;
  gap: var(--aoi-grid-gap);
}

.settings-shell-nav__depth :deep(.aoi-segmented__item) {
  min-height: var(--aoi-control-height-md);
}

.settings-shell-nav-mobile {
  display: none;
}

.settings-shell-nav__group {
  display: grid;
  gap: max(6px, calc(var(--aoi-grid-gap-compact) - 5px));
}

.settings-shell-nav__group h2 {
  color: var(--aoi-text-muted);
  font-size: 12px;
  font-weight: 780;
  letter-spacing: 0;
}

.settings-shell-nav__item {
  display: flex;
  min-height: var(--aoi-settings-nav-item-height);
  align-items: center;
  gap: var(--aoi-nav-group-gap);
  border-radius: var(--aoi-radius-choice);
  color: var(--aoi-text-muted);
  font-weight: 740;
  padding: 0 8px;
}

.settings-shell-nav__indicator {
  position: absolute;
  top: 0;
  left: 0;
  z-index: 1;
  display: block;
  width: var(--settings-shell-nav-indicator-width, 3px);
  height: var(--settings-shell-nav-indicator-height, 16px);
  background: var(--aoi-active-color);
  opacity: var(--settings-shell-nav-indicator-opacity, 0);
  pointer-events: none;
  transform: translate3d(
    var(--settings-shell-nav-indicator-x, 0),
    var(--settings-shell-nav-indicator-y, 0),
    0
  );
  transition:
    transform var(--aoi-motion-base) var(--aoi-ease-out),
    opacity var(--aoi-motion-fast) var(--aoi-ease-out);
}

.settings-shell-nav__item:hover,
.settings-shell-nav__item--active {
  background: var(--aoi-state-hover);
  color: var(--aoi-active-color);
}

.settings-shell-nav__item--active {
  position: relative;
  padding: 0 8px;
}

.settings-shell-nav__empty {
  margin: 0;
}

@media (max-width: 960px) {
  .settings-shell-nav {
    display: none;
  }

  .settings-shell-nav-mobile {
    position: sticky;
    top: calc(var(--aoi-mobile-header-height) + 8px);
    z-index: var(--aoi-z-sticky);
    display: grid;
    gap: var(--aoi-grid-gap-compact);
    border: 1px solid var(--aoi-border);
    border-radius: var(--aoi-radius-container);
    background: var(--aoi-panel-bg);
    box-shadow: var(--aoi-shadow-sm);
    padding: 8px;
  }

  .settings-shell-nav-mobile__items {
    position: relative;
    display: flex;
    overflow-x: auto;
    gap: var(--aoi-grid-gap-compact);
  }

  .settings-shell-nav-mobile .settings-shell-nav__item {
    flex: 0 0 auto;
  }
}

@media (prefers-reduced-motion: reduce) {
  .settings-shell-nav__indicator {
    transition: none;
  }
}
</style>


