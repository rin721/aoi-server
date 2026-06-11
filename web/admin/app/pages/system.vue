<script setup lang="ts">
import type { SystemConfigItem, SystemConfigSection, SystemConfigSnapshot } from "~/types/admin"

const api = useAdminApi()
const snapshot = ref<SystemConfigSnapshot>({ sections: [] })
const selectedCode = ref("")
const loading = ref(false)
const error = ref("")

const sections = computed<SystemConfigSection[]>(() => [...snapshot.value.sections].sort((left, right) => left.order - right.order || left.code.localeCompare(right.code)))
const selectedSection = computed(() => sections.value.find((section) => section.code === selectedCode.value) || sections.value[0] || null)
const itemCount = computed(() => sections.value.reduce((count, section) => count + section.items.length, 0))
const secretCount = computed(() => sections.value.reduce((count, section) => count + section.items.filter((item) => item.secret).length, 0))

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

function itemValueClass(item: SystemConfigItem) {
  return {
    "config-item__value--empty": formatConfigValue(item.value) === "-",
    "config-item__value--secret": item.secret
  }
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

    <section class="config-overview" aria-label="配置概览">
      <div class="config-stat">
        <span>分区</span>
        <strong>{{ sections.length }}</strong>
      </div>
      <div class="config-stat">
        <span>字段</span>
        <strong>{{ itemCount }}</strong>
      </div>
      <div class="config-stat">
        <span>脱敏</span>
        <strong>{{ secretCount }}</strong>
      </div>
      <div class="config-stat">
        <span>来源</span>
        <strong>runtime</strong>
      </div>
    </section>

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

      <article v-if="selectedSection" class="admin-card config-panel">
        <div class="admin-card__header config-panel__header">
          <div class="config-panel__title">
            <AoiIcon :name="selectedSection.icon || 'settings'" decorative />
            <div>
              <h2>{{ selectedSection.label }}</h2>
              <p>{{ selectedSection.description }}</p>
            </div>
          </div>
          <span class="badge">{{ selectedSection.code }}</span>
        </div>

        <div class="config-items">
          <div v-for="item in selectedSection.items" :key="item.key" class="config-item">
            <div class="config-item__meta">
              <strong>{{ item.label }}</strong>
              <small>{{ item.key }}</small>
            </div>
            <div class="config-item__readout">
              <span class="config-item__value" :class="itemValueClass(item)">{{ formatConfigValue(item.value) }}</span>
              <span v-if="item.secret" class="badge badge--warning">脱敏</span>
            </div>
          </div>
        </div>
      </article>
    </section>

    <article v-else class="admin-card config-empty">
      <AoiIcon name="settings" :size="24" decorative />
      <p>{{ loading ? "配置加载中。" : "暂无配置快照。" }}</p>
    </article>
  </div>
</template>

<style scoped>
.config-overview {
  display: grid;
  gap: 12px;
  grid-template-columns: repeat(4, minmax(0, 1fr));
}

.config-stat {
  background: var(--aoi-surface-solid);
  border: 1px solid var(--aoi-border);
  border-radius: var(--aoi-radius-card);
  display: grid;
  gap: 8px;
  min-width: 0;
  padding: 14px;
}

.config-stat span {
  color: var(--aoi-text-muted);
  font-size: 12px;
  font-weight: 700;
}

.config-stat strong {
  color: var(--aoi-text);
  font-size: 22px;
  line-height: 1.15;
  min-width: 0;
  overflow-wrap: anywhere;
}

.config-layout {
  align-items: start;
  display: grid;
  gap: 14px;
  grid-template-columns: minmax(180px, 240px) minmax(0, 1fr);
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
  gap: 10px;
  grid-template-columns: 22px minmax(0, 1fr) auto;
  min-height: 46px;
  padding: 10px 12px;
  text-align: left;
}

.config-tab:last-child {
  border-bottom: 0;
}

.config-tab:hover,
.config-tab--active {
  background: var(--aoi-active-bg);
  color: var(--aoi-active-color);
}

.config-tab span {
  font-weight: 700;
  min-width: 0;
  overflow-wrap: anywhere;
}

.config-tab small {
  background: var(--aoi-surface-muted);
  border: 1px solid var(--aoi-border);
  border-radius: 999px;
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

.config-panel__header {
  align-items: flex-start;
}

.config-panel__title {
  align-items: flex-start;
  display: flex;
  gap: 10px;
  min-width: 0;
}

.config-panel__title > svg {
  color: var(--aoi-accent-60);
  flex: 0 0 auto;
  margin-top: 2px;
}

.config-panel__title h2,
.config-panel__title p {
  margin: 0;
}

.config-panel__title p {
  color: var(--aoi-text-muted);
  line-height: 1.6;
  margin-top: 4px;
  overflow-wrap: anywhere;
}

.config-items {
  display: grid;
}

.config-item {
  align-items: start;
  border-top: 1px solid var(--aoi-border);
  display: grid;
  gap: 14px;
  grid-template-columns: minmax(180px, 260px) minmax(0, 1fr);
  padding: 13px 0;
}

.config-item:first-child {
  border-top: 0;
  padding-top: 0;
}

.config-item:last-child {
  padding-bottom: 0;
}

.config-item__meta,
.config-item__readout {
  min-width: 0;
}

.config-item__meta {
  display: grid;
  gap: 4px;
}

.config-item__meta strong {
  color: var(--aoi-text);
  overflow-wrap: anywhere;
}

.config-item__meta small {
  color: var(--aoi-text-muted);
  font-family: var(--aoi-font-mono);
  overflow-wrap: anywhere;
}

.config-item__readout {
  align-items: center;
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  justify-content: flex-end;
}

.config-item__value {
  background: var(--aoi-surface-soft);
  border: 1px solid var(--aoi-border);
  border-radius: 8px;
  color: var(--aoi-text);
  display: inline-block;
  max-width: 100%;
  overflow-wrap: anywhere;
  padding: 6px 9px;
  text-align: right;
  white-space: pre-wrap;
}

.config-item__value--empty {
  color: var(--aoi-text-muted);
}

.config-item__value--secret {
  font-weight: 800;
}

.config-empty {
  align-items: center;
  color: var(--aoi-text-muted);
  display: flex;
  gap: 10px;
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
    min-width: 150px;
  }
}

@media (max-width: 680px) {
  .config-overview {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .config-panel__header {
    align-items: flex-start;
    flex-direction: column;
  }

  .config-item {
    gap: 8px;
    grid-template-columns: 1fr;
  }

  .config-item__readout {
    justify-content: flex-start;
  }

  .config-item__value {
    text-align: left;
  }
}
</style>
