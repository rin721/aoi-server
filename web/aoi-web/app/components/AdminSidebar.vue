<script setup lang="ts">
const route = useRoute()
const { groups } = useAdminNavigation()
const { state } = useAdminUiPreferences()

function isActive(to: string) {
  return to === "/" ? route.path === "/" : route.path.startsWith(to)
}
</script>

<template>
  <aside class="admin-sidebar" :class="{ 'admin-sidebar--collapsed': state.sidebarCollapsed }">
    <NuxtLink class="admin-sidebar__brand" to="/">
      <span class="admin-brand-mark">A</span>
      <span class="admin-sidebar__brand-copy">
        <strong>Aoi Admin</strong>
        <small>go-scaffold</small>
      </span>
    </NuxtLink>

    <nav class="admin-sidebar__nav" aria-label="后台主导航">
      <section v-for="group in groups" :key="group.label" class="admin-sidebar__group">
        <p class="admin-sidebar__group-label">{{ group.label }}</p>
        <NuxtLink
          v-for="item in group.items"
          :key="item.to"
          class="admin-sidebar__item"
          :class="{ 'admin-sidebar__item--active': isActive(item.to) }"
          :title="item.label"
          :to="item.to"
        >
          <AoiIcon :name="item.icon" decorative />
          <span>{{ item.label }}</span>
        </NuxtLink>
      </section>
    </nav>
  </aside>
</template>


