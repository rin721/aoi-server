<script setup lang="ts">
const route = useRoute()

const primaryItems = [
  { icon: "layout-dashboard", label: "仪表盘", to: "/" },
  { icon: "building-2", label: "组织", to: "/organizations" },
  { icon: "users", label: "用户", to: "/users" },
  { icon: "shield-check", label: "角色权限", to: "/roles" },
  { icon: "monitor-check", label: "会话", to: "/sessions" },
  { icon: "scroll-text", label: "审计日志", to: "/audit-logs" },
  { icon: "lock-keyhole", label: "安全", to: "/security" },
  { icon: "list-checks", label: "Demo Todo", to: "/todos" }
]

function isActive(to: string) {
  return to === "/" ? route.path === "/" : route.path.startsWith(to)
}
</script>

<template>
  <aside class="app-sidebar">
    <NuxtLink class="app-sidebar__brand" to="/">
      <span class="brand-mark">A</span>
      <span>
        <strong>Aoi Admin</strong>
        <small>go-scaffold</small>
      </span>
    </NuxtLink>

    <nav class="app-sidebar__nav" aria-label="主导航">
      <NuxtLink
        v-for="item in primaryItems"
        :key="item.to"
        class="app-sidebar__item"
        :class="{ 'app-sidebar__item--active': isActive(item.to) }"
        :to="item.to"
      >
        <AoiIcon :name="item.icon" decorative />
        <span>{{ item.label }}</span>
      </NuxtLink>
    </nav>
  </aside>
</template>

<style scoped>
.app-sidebar {
  position: sticky;
  top: 0;
  display: flex;
  height: 100vh;
  flex-direction: column;
  border-right: 1px solid var(--aoi-border);
  background: rgba(255, 255, 255, .78);
  box-shadow: 10px 0 28px rgba(24, 72, 78, .06);
  padding: 14px;
  backdrop-filter: blur(18px);
}

.app-sidebar__brand {
  display: flex;
  align-items: center;
  gap: 10px;
  border-radius: var(--aoi-radius-control);
  padding: 8px;
}

.app-sidebar__brand strong,
.app-sidebar__brand small {
  display: block;
}

.app-sidebar__brand small {
  color: var(--aoi-text-muted);
  font-size: 12px;
}

.app-sidebar__nav {
  display: grid;
  gap: 4px;
  margin-top: 18px;
}

.app-sidebar__item {
  display: flex;
  min-height: 40px;
  align-items: center;
  gap: 10px;
  border-radius: var(--aoi-radius-control);
  color: var(--aoi-text-muted);
  font-weight: 760;
  padding: 0 10px;
  transition:
    background var(--aoi-motion-fast) var(--aoi-ease-out),
    color var(--aoi-motion-fast) var(--aoi-ease-out);
}

.app-sidebar__item:hover,
.app-sidebar__item--active {
  background: var(--aoi-accent-10);
  color: var(--aoi-accent-60);
}

@media (max-width: 980px) {
  .app-sidebar {
    position: static;
    height: auto;
    border-right: 0;
    border-bottom: 1px solid var(--aoi-border);
  }

  .app-sidebar__nav {
    grid-template-columns: repeat(4, minmax(0, 1fr));
  }
}

@media (max-width: 720px) {
  .app-sidebar__nav {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}
</style>
