<script setup lang="ts">
import { SERVER_STATUS_DASHBOARD_CONFIG } from "~/config/server-status-dashboard"
import type { SystemServerInfo } from "~/types/admin"
import { createServerStatusDashboardModel } from "~/utils/serverStatusDashboard"

const api = useAdminApi()
const dashboardConfig = SERVER_STATUS_DASHBOARD_CONFIG
const info = ref<SystemServerInfo | null>(null)
const loading = ref(false)
const error = ref("")

const autoRefresh = useAdminAutoRefresh({
  blocked: loading,
  defaultEnabled: dashboardConfig.refresh.autoEnabled,
  intervalMs: dashboardConfig.refresh.intervalMs,
  load,
  manualCooldownMs: dashboardConfig.refresh.manualCooldownMs
})
const dashboard = computed(() => createServerStatusDashboardModel(info.value, dashboardConfig, autoRefresh.enabled.value))
const hasData = computed(() => Boolean(info.value))
const emptyTitle = computed(() => loading.value ? dashboardConfig.emptyStates.loading : dashboardConfig.emptyStates.noData)
const emptyIcon = computed(() => loading.value ? "loader-circle" : "activity")

async function load(options: { silent?: boolean } = {}) {
  if (!options.silent) {
    loading.value = true
  }
  error.value = ""
  try {
    info.value = await api.getSystemServerInfo()
  } catch (err) {
    error.value = errorMessage(err)
  } finally {
    if (!options.silent) {
      loading.value = false
    }
  }
}

onMounted(autoRefresh.refreshNow)

useHead({
  title: `${dashboardConfig.labels.pageTitle} - Aoi Admin`
})
</script>

<template>
  <div class="page-grid server-status-page">
    <PageHeader
      :title="dashboardConfig.labels.pageTitle"
      icon="activity"
      :description="dashboardConfig.labels.pageDescription"
    >
      <template #actions>
        <AoiMetaPill :intent="dashboard.overall.intent" appearance="soft" :icon="dashboard.overall.icon">
          {{ dashboard.overall.label }}
        </AoiMetaPill>
        <AdminAutoRefreshControls
          v-model="autoRefresh.enabled.value"
          :last-refreshed-label="autoRefresh.lastRefreshedLabel.value"
          :next-refresh-label="autoRefresh.nextRefreshLabel.value"
          :status-label="autoRefresh.statusLabel.value"
        />
        <AoiButton appearance="soft" icon="refresh-cw" :loading="loading" :disabled="autoRefresh.refreshDisabled.value" @click="autoRefresh.refreshNow">
          {{ dashboardConfig.labels.refreshAction }}
        </AoiButton>
      </template>
    </PageHeader>

    <AoiStatusMessage tone="danger" icon="circle-alert" :message="error" />
    <AoiStatusMessage
      v-if="hasData && dashboard.anomalyCount > 0"
      :intent="dashboard.overall.intent === 'danger' ? 'danger' : 'warning'"
      :icon="dashboard.overall.icon"
      :message="dashboard.anomalySummary"
    />

    <AoiStatGrid :items="dashboard.kpis" :columns="4" />

    <AoiMasonryGrid v-if="hasData" class="server-layout" gap="normal">
      <AoiAdminCard
        :title="dashboard.panels.environment.title"
        :description="dashboard.panels.environment.description"
        :icon="dashboard.panels.environment.icon"
        :badge="dashboard.panels.environment.badge"
        :badge-intent="dashboard.panels.environment.badgeIntent"
      >
        <AoiKeyValueList :items="dashboard.environmentItems" layout="cards" />
      </AoiAdminCard>

      <AoiAdminCard
        :title="dashboard.panels.cpu.title"
        :description="dashboard.panels.cpu.description"
        :icon="dashboard.panels.cpu.icon"
        :badge="dashboard.panels.cpu.badge"
        :badge-intent="dashboard.panels.cpu.badgeIntent"
      >
        <AoiProgressBar
          :value="dashboard.cpu.averageValue"
          :intent="dashboard.cpu.averageState.intent"
          size="md"
          :label="dashboard.panels.cpu.title"
        />
        <div v-if="dashboard.cpu.cores.length" class="server-scroll-area server-scroll-area--cpu">
          <div class="server-cpu-grid">
            <div
              v-for="core in dashboard.cpu.cores"
              :key="core.index"
              class="server-cpu-row"
              :data-state="core.state.level"
            >
              <span>{{ core.label }}</span>
              <AoiProgressBar :value="core.value" :intent="core.state.intent" size="sm" :label="core.label" />
              <strong>{{ core.displayValue }}</strong>
            </div>
          </div>
        </div>
        <AoiDataState
          v-else
          :title="dashboardConfig.emptyStates.cpu"
          icon="activity"
          intent="info"
        />
      </AoiAdminCard>

      <AoiAdminCard
        :title="dashboard.panels.ram.title"
        :description="dashboard.panels.ram.description"
        :icon="dashboard.panels.ram.icon"
        :badge="dashboard.panels.ram.badge"
        :badge-intent="dashboard.panels.ram.badgeIntent"
      >
        <AoiProgressBar :value="dashboard.ram.value" :intent="dashboard.ram.state.intent" size="md" :label="dashboard.panels.ram.title" />
        <AoiKeyValueList :items="dashboard.ram.items" layout="cards" density="compact" />
      </AoiAdminCard>

      <AoiAdminCard
        :title="dashboard.panels.heap.title"
        :description="dashboard.panels.heap.description"
        :icon="dashboard.panels.heap.icon"
        :badge="dashboard.panels.heap.badge"
        :badge-intent="dashboard.panels.heap.badgeIntent"
      >
        <AoiProgressBar :value="dashboard.heap.value" :intent="dashboard.heap.state.intent" size="md" :label="dashboard.panels.heap.title" />
        <AoiKeyValueList :items="dashboard.heap.items" layout="cards" density="compact" />
      </AoiAdminCard>

      <AoiAdminCard
        :title="dashboard.panels.gc.title"
        :description="dashboard.panels.gc.description"
        :icon="dashboard.panels.gc.icon"
        :badge="dashboard.panels.gc.badge"
        :badge-intent="dashboard.panels.gc.badgeIntent"
      >
        <AoiKeyValueList :items="dashboard.gcItems" layout="rows" />
      </AoiAdminCard>

      <AoiAdminCard
        :title="dashboard.panels.disk.title"
        :description="dashboard.panels.disk.description"
        :icon="dashboard.panels.disk.icon"
        :badge="dashboard.panels.disk.badge"
        :badge-intent="dashboard.panels.disk.badgeIntent"
      >
        <div v-if="dashboard.diskRows.length" class="server-scroll-area">
          <div class="server-disk-list">
            <div v-for="item in dashboard.diskRows" :key="item.mountPoint" class="server-disk-row" :data-state="item.state.level">
              <div class="server-disk-row__head">
                <strong>{{ item.mountPoint }}</strong>
                <AoiMetaPill appearance="soft" :intent="item.state.intent" :icon="item.state.icon">
                  {{ item.state.label }}
                </AoiMetaPill>
              </div>
              <AoiProgressBar :value="item.value" :intent="item.state.intent" size="sm" :label="item.mountPoint" />
              <div class="server-disk-row__meta">
                <span>{{ item.displaySize }}</span>
                <span>{{ item.fsType }}</span>
                <strong>{{ item.displayPercent }}</strong>
              </div>
            </div>
          </div>
        </div>
        <AoiDataState v-else :title="dashboardConfig.emptyStates.disk" icon="hard-drive" intent="info" />
      </AoiAdminCard>

      <AoiAdminCard
        :title="dashboard.panels.build.title"
        :description="dashboard.panels.build.description"
        :icon="dashboard.panels.build.icon"
        :badge="dashboard.panels.build.badge"
        :badge-intent="dashboard.panels.build.badgeIntent"
      >
        <div class="server-scroll-area server-scroll-area--build">
          <AoiKeyValueList :items="dashboard.buildItems" layout="rows" />
        </div>
      </AoiAdminCard>
    </AoiMasonryGrid>

    <AoiAdminCard v-else padding="lg">
      <AoiDataState
        :title="emptyTitle"
        :icon="emptyIcon"
        :loading="loading"
        :description="error || undefined"
        :intent="error ? 'danger' : 'info'"
      >
        <template #actions>
          <AoiButton appearance="soft" icon="refresh-cw" :loading="loading" :disabled="autoRefresh.refreshDisabled.value" @click="autoRefresh.refreshNow">
            {{ dashboardConfig.labels.refreshAction }}
          </AoiButton>
        </template>
      </AoiDataState>
    </AoiAdminCard>
  </div>
</template>

<style scoped>
.server-status-page {
  min-width: 0;
}

.server-layout :deep(.aoi-admin-card__body) {
  display: grid;
  gap: var(--aoi-admin-panel-gap-compact);
}

.server-scroll-area {
  max-height: var(--aoi-admin-server-scroll-max-height);
  min-width: 0;
  overflow: auto;
  padding-inline-end: var(--aoi-admin-card-copy-gap);
}

.server-scroll-area--cpu {
  max-height: var(--aoi-admin-server-cpu-scroll-max-height);
}

.server-scroll-area--build {
  max-height: var(--aoi-admin-server-build-scroll-max-height);
}

.server-cpu-grid,
.server-disk-list {
  display: grid;
  min-width: 0;
  gap: var(--aoi-admin-panel-gap-compact);
}

.server-cpu-grid {
  grid-template-columns: repeat(auto-fit, minmax(min(100%, var(--aoi-admin-cpu-row-min-width)), 1fr));
}

.server-cpu-row {
  display: grid;
  min-height: var(--aoi-admin-cpu-row-height);
  min-width: 0;
  align-items: center;
  gap: var(--aoi-admin-kv-value-gap);
  grid-template-columns: var(--aoi-admin-cpu-label-width) minmax(0, 1fr) var(--aoi-admin-cpu-value-width);
}

.server-cpu-row span,
.server-cpu-row strong,
.server-disk-row__meta span,
.server-disk-row__meta strong {
  font-family: var(--aoi-font-mono);
  font-size: var(--aoi-admin-kv-mono-size);
}

.server-cpu-row span {
  color: var(--aoi-admin-text-muted);
}

.server-cpu-row strong {
  color: var(--aoi-admin-text);
  text-align: right;
}

.server-disk-row {
  display: grid;
  min-width: 0;
  gap: var(--aoi-admin-kv-value-gap);
  border-top: 1px solid var(--aoi-admin-border-soft);
  padding-top: var(--aoi-admin-kv-card-padding);
}

.server-disk-row:first-child {
  border-top: 0;
  padding-top: 0;
}

.server-disk-row__head,
.server-disk-row__meta {
  display: flex;
  min-width: 0;
  align-items: center;
  justify-content: space-between;
  gap: var(--aoi-admin-card-gap);
}

.server-disk-row__head strong {
  min-width: 0;
  overflow-wrap: anywhere;
  color: var(--aoi-admin-text);
}

.server-disk-row__meta {
  color: var(--aoi-admin-text-muted);
}

.server-disk-row__meta span {
  min-width: 0;
  overflow-wrap: anywhere;
}

.server-disk-row__meta strong {
  color: var(--aoi-admin-text);
}

.server-note {
  color: var(--aoi-admin-text-muted);
  font-size: var(--aoi-admin-server-note-size);
  line-height: 1.6;
}

@media (max-width: 680px) {
  .server-disk-row__head,
  .server-disk-row__meta {
    align-items: flex-start;
    flex-direction: column;
    gap: var(--aoi-admin-card-copy-gap);
  }
}
</style>
