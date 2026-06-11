<script setup lang="ts">
const route = useRoute()
const { mobileItems } = useAdminNavigation()

function isActive(to: string) {
  return to === "/" ? route.path === "/" : route.path.startsWith(to)
}
</script>

<template>
  <nav class="admin-mobile-nav" aria-label="移动端主导航">
    <NuxtLink
      v-for="item in mobileItems"
      :key="item.to"
      class="admin-mobile-nav__item"
      :class="{ 'admin-mobile-nav__item--active': isActive(item.to) }"
      :to="item.to"
    >
      <AoiIcon :name="item.icon" decorative />
      <span>{{ item.label }}</span>
    </NuxtLink>
  </nav>
</template>


