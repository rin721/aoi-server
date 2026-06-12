<script setup lang="ts">
type VisitedTab = {
  icon: string
  label: string
  path: string
}

const storageKey = "aoi.admin.visited-tabs.v1"
const route = useRoute()
const router = useRouter()
const { findItem } = useAdminNavigation()
const tabs = ref<VisitedTab[]>([])

function readTabs() {
  if (!import.meta.client) {
    return []
  }
  try {
    const parsed = JSON.parse(sessionStorage.getItem(storageKey) || "[]") as VisitedTab[]
    return Array.isArray(parsed) ? parsed : []
  } catch {
    return []
  }
}

function saveTabs() {
  if (import.meta.client) {
    sessionStorage.setItem(storageKey, JSON.stringify(tabs.value))
  }
}

function addCurrent(path = route.path) {
  const item = findItem(path)
  if (!item) {
    return
  }

  const next = { icon: item.icon, label: item.label, path: item.to }
  tabs.value = [
    ...tabs.value.filter((tab) => tab.path !== next.path),
    next
  ].slice(-8)

  if (!tabs.value.some((tab) => tab.path === "/")) {
    tabs.value.unshift({ icon: "layout-dashboard", label: "仪表盘", path: "/" })
  }
  saveTabs()
}

function closeTab(path: string) {
  tabs.value = tabs.value.filter((tab) => tab.path !== path)
  if (route.path === path) {
    void router.push(tabs.value.at(-1)?.path || "/")
  }
  saveTabs()
}

onMounted(() => {
  tabs.value = readTabs()
  addCurrent()
})

watch(() => route.path, (path) => addCurrent(path))
</script>

<template>
  <div class="admin-tabs" aria-label="访问标签页">
    <AoiLink
      v-for="tab in tabs"
      :key="tab.path"
      class="admin-tabs__item"
      :class="{ 'admin-tabs__item--active': route.path === tab.path }"
      :to="tab.path"
    >
      <AoiIcon :name="tab.icon" decorative />
      <span>{{ tab.label }}</span>
      <button
        v-if="tab.path !== '/'"
        aria-label="关闭标签"
        class="admin-tabs__close"
        type="button"
        @click.prevent="closeTab(tab.path)"
      >
        <AoiIcon name="x" decorative />
      </button>
    </AoiLink>
  </div>
</template>


