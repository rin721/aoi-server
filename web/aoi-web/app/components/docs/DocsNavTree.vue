<script setup lang="ts">
defineOptions({
  name: "DocsNavTree"
})

export interface DocsNavigationItem {
  children?: DocsNavigationItem[]
  description?: string
  icon?: string
  path?: string
  title?: string
}

const props = withDefaults(defineProps<{
  activePath?: string
  items?: DocsNavigationItem[]
  level?: number
}>(), {
  activePath: "",
  items: () => [],
  level: 0
})

function isActive(item: DocsNavigationItem): boolean {
  return Boolean(item.path && props.activePath === item.path)
}

function isOpen(item: DocsNavigationItem): boolean {
  return Boolean(item.children?.some((child) => child.path === props.activePath || isOpen(child)))
}
</script>

<template>
  <ul class="docs-nav-tree" :class="`docs-nav-tree--level-${props.level}`">
    <li v-for="item in props.items" :key="item.path || item.title" class="docs-nav-tree__item">
      <AoiLink
        v-if="item.path"
        class="docs-nav-tree__link"
        :class="{ 'docs-nav-tree__link--active': isActive(item) }"
        :to="item.path"
        :aria-current="isActive(item) ? 'page' : undefined"
      >
        <AoiIcon v-if="item.icon" :name="item.icon" :size="15" decorative />
        <span>{{ item.title || item.path }}</span>
      </AoiLink>
      <span v-else class="docs-nav-tree__label">
        <AoiIcon v-if="item.icon" :name="item.icon" :size="15" decorative />
        <span>{{ item.title }}</span>
      </span>

      <DocsNavTree
        v-if="item.children?.length && (props.level === 0 || isOpen(item) || isActive(item))"
        :items="item.children"
        :active-path="props.activePath"
        :level="props.level + 1"
      />
    </li>
  </ul>
</template>

<style scoped>
.docs-nav-tree {
  display: grid;
  gap: 4px;
  list-style: none;
  margin: 0;
  padding: 0;
}

.docs-nav-tree--level-1 {
  margin: 4px 0 6px 13px;
  padding-left: 10px;
  border-left: 1px solid var(--aoi-border);
}

.docs-nav-tree--level-2 {
  margin-left: 8px;
}

.docs-nav-tree__item {
  min-width: 0;
}

.docs-nav-tree__link,
.docs-nav-tree__label {
  display: flex;
  width: 100%;
  min-width: 0;
  min-height: 34px;
  align-items: center;
  gap: 8px;
  border-radius: var(--aoi-radius-control);
  color: var(--aoi-text-muted);
  font-weight: 720;
  line-height: 1.3;
  padding: 7px 9px;
}

.docs-nav-tree__link:hover,
.docs-nav-tree__link:focus-visible {
  background: var(--aoi-state-hover);
  color: var(--aoi-text);
}

.docs-nav-tree__link--active {
  background: var(--aoi-nav-active-bg);
  color: var(--aoi-nav-active-color);
}

.docs-nav-tree__link span,
.docs-nav-tree__label span {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>


