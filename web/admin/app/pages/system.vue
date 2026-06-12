<script setup lang="ts">
import type { SystemConfigSection, SystemConfigSnapshot } from "~/types/admin"
import type { AoiKeyValueItem, AoiStatItem } from "~/types/ui"

const api = useAdminApi()
const snapshot = ref<SystemConfigSnapshot>({ sections: [] })
const selectedCode = ref("")
const loading = ref(false)
const error = ref("")

const sections = computed<SystemConfigSection[]>(() => [...snapshot.value.sections].sort((left, right) => left.order - right.order || left.code.localeCompare(right.code)))
const selectedSection = computed(() => sections.value.find((section) => section.code === selectedCode.value) || sections.value[0] || null)
const itemCount = computed(() => sections.value.reduce((count, section) => count + section.items.length, 0))
const secretCount = computed(() => sections.value.reduce((count, section) => count + section.items.filter((item) => item.secret).length, 0))
const overviewItems = computed<AoiStatItem[]>(() => [
  { icon: "blocks", label: "分区", value: sections.value.length },
  { icon: "list-tree", label: "字段", value: itemCount.value },
  { icon: "shield-check", intent: secretCount.value ? "warning" : "neutral", label: "脱敏", value: secretCount.value },
  { icon: "server-cog", label: "来源", value: "runtime" }
])
const selectedItems = computed<AoiKeyValueItem[]>(() =>
  selectedSection.value?.items.map((item) => ({
    badge: item.secret ? "已脱敏" : undefined,
    intent: item.secret ? "warning" : "neutral",
    label: item.label,
    meta: item.key,
    monospace: typeof item.value !== "boolean",
    secret: item.secret,
    value: formatConfigValue(item.value)
  })) || []
)

async function load() {
  loading.value = true
  error.value = ""
  try {
    snapshot.value = await api.getSystemConfig()
    if (!sections.value.some((section) => section.code === selectedCode.value)) {
      selectedCode.value = sections.value[0]?.code || ""
    }
  } catch (err) {
    error.value = errorMessage(err)
  } finally {
    loading.value = false
  }
}

function selectSection(code: string) {
  selectedCode.value = code
}

function formatConfigValue(value: unknown) {
  if (value === undefined || value === null || value === "") {
    return "-"
  }
  if (typeof value === "boolean") {
    return value ? "是" : "否"
  }
  if (Array.isArray(value)) {
    return value.length ? value.join(", ") : "-"
  }
  if (typeof value === "object") {
    return JSON.stringify(value)
  }
  return String(value)
}

onMounted(load)

useHead({
  title: "系统配置 - Aoi Admin"
})
</script>

<template>
  <div class="page-grid">
    <PageHeader title="系统配置" icon="settings" description="当前进程的运行配置快照，敏感字段只显示配置状态。">
      <template #actions>
        <AoiButton appearance="soft" icon="refresh-cw" :loading="loading" @click="load">刷新</AoiButton>
      </template>
    </PageHeader>

    <AoiStatusMessage tone="danger" :message="error" />

    <AoiStatGrid :items="overviewItems" :columns="4" />

    <section v-if="sections.length" class="config-layout">
      <nav class="config-tabs" aria-label="配置分区">
        <button
          v-for="section in sections"
          :key="section.code"
          class="config-tab"
          :class="{ 'config-tab--active': selectedSection?.code === section.code }"
          type="button"
          @click="selectSection(section.code)"
        >
          <AoiIcon :name="section.icon || 'settings'" decorative />
          <span>{{ section.label }}</span>
          <small>{{ section.items.length }}</small>
        </button>
      </nav>

      <AoiAdminCard
        v-if="selectedSection"
        class="config-panel"
        :badge="selectedSection.code"
        :description="selectedSection.description"
        :icon="selectedSection.icon || 'settings'"
        :title="selectedSection.label"
      >
        <AoiKeyValueList :items="selectedItems" layout="rows" />
      </AoiAdminCard>
    </section>

    <AoiAdminCard v-else class="config-empty" padding="lg">
      <AoiIcon name="settings" :size="24" decorative />
      <p>{{ loading ? "配置加载中。" : "暂无配置快照。" }}</p>
    </AoiAdminCard>
  </div>
</template>

<style scoped>
.config-layout {
  align-items: start;
  display: grid;
  gap: var(--aoi-admin-panel-gap);
  grid-template-columns: minmax(var(--aoi-admin-config-nav-min-width), var(--aoi-admin-config-nav-width)) minmax(0, 1fr);
}

.config-tabs {
  background: var(--aoi-surface-solid);
  border: 1px solid var(--aoi-border);
  border-radius: var(--aoi-radius-card);
  display: grid;
  overflow: hidden;
  scrollbar-width: none;
}

.config-tabs::-webkit-scrollbar {
  display: none;
}

.config-tab {
  align-items: center;
  background: transparent;
  border: 0;
  border-bottom: 1px solid var(--aoi-border);
  color: var(--aoi-text);
  cursor: pointer;
  display: grid;
  font: inherit;
  gap: var(--aoi-admin-card-gap);
  grid-template-columns: var(--aoi-nav-icon-size) minmax(0, 1fr) auto;
  min-height: var(--aoi-admin-nav-row-height);
  padding: var(--aoi-admin-nav-row-padding);
  text-align: left;
}

.config-tab:last-child {
  border-bottom: 0;
}

.config-tab:hover,
.config-tab:focus-visible,
.config-tab--active {
  background: var(--aoi-state-active);
  color: var(--aoi-active-color);
}

.config-tab span {
  font-weight: 760;
  min-width: 0;
  overflow-wrap: anywhere;
}

.config-tab small {
  background: var(--aoi-surface-muted);
  border: 1px solid var(--aoi-border);
  border-radius: var(--aoi-radius-round);
  color: var(--aoi-text-muted);
  font-size: 11px;
  font-weight: 800;
  min-width: 24px;
  padding: 2px 6px;
  text-align: center;
}

.config-panel {
  min-width: 0;
}

.config-empty {
  align-items: center;
  color: var(--aoi-text-muted);
  display: flex;
  gap: var(--aoi-admin-card-gap);
  justify-content: center;
  min-height: 160px;
}

.config-empty p {
  margin: 0;
}

@media (max-width: 920px) {
  .config-layout {
    grid-template-columns: 1fr;
  }

  .config-tabs {
    display: flex;
    overflow-x: auto;
  }

  .config-tab {
    border-bottom: 0;
    border-right: 1px solid var(--aoi-border);
    flex: 0 0 auto;
    min-width: var(--aoi-admin-config-tab-min-width);
  }
}
</style>
