<script setup lang="ts">
import type { AuditLog, HealthStatus, ReadyStatus, Session } from "~/types/admin"
import type { AoiStatItem } from "~/types/ui"

const api = useAdminApi()
const auth = useAuthStore()
const loading = ref(false)
const error = ref("")
const health = ref<HealthStatus | null>(null)
const ready = ref<ReadyStatus | null>(null)
const sessions = ref<Session[]>([])
const auditLogs = ref<AuditLog[]>([])
const dashboardStats = computed<AoiStatItem[]>(() => [
  { icon: "user", label: "登录账号", value: auth.user?.username || "-", description: auth.user?.email || "-" },
  { icon: "building-2", label: "当前组织", value: auth.currentOrg?.code || "-", description: auth.currentOrg?.name || "-" },
  { icon: "heart-pulse", intent: health.value?.status === "ok" ? "success" : "neutral", label: "Health", value: health.value?.status || "-", description: "/health" },
  {
    icon: "badge-check",
    intent: ready.value?.status === "ready" ? "success" : "neutral",
    label: "Ready",
    value: ready.value?.status || "-",
    description: ready.value?.checks ? Object.entries(ready.value.checks).map(([key, value]) => `${key}:${value}`).join(" · ") : "/ready"
  }
])

async function refresh(options: { silent?: boolean } = {}) {
  if (!options.silent) {
    loading.value = true
  }
  error.value = ""
  try {
    const [healthResult, readyResult] = await Promise.all([
      api.getHealth(),
      api.getReady()
    ])
    health.value = healthResult
    ready.value = readyResult

    if (auth.currentOrgId) {
      const [sessionResult, auditResult] = await Promise.all([
        api.listSessions(auth.currentOrgId, { pageSize: 6 }),
        api.listAuditLogs(auth.currentOrgId, 6)
      ])
      sessions.value = sessionResult.items
      auditLogs.value = auditResult
    }
  } catch (err) {
    error.value = errorMessage(err)
  } finally {
    if (!options.silent) {
      loading.value = false
    }
  }
}

const autoRefresh = useAdminAutoRefresh({ blocked: loading, load: refresh })

onMounted(autoRefresh.refreshNow)
watch(() => auth.currentOrgId, () => {
  void autoRefresh.refreshNow()
})

useHead({
  title: "仪表盘 - Aoi Admin"
})
</script>

<template>
  <div class="page-grid">
    <PageHeader title="仪表盘" icon="layout-dashboard" description="查看服务状态、当前组织和最近 IAM 活动。">
      <template #actions>
        <AdminAutoRefreshControls
          v-model="autoRefresh.enabled.value"
          :last-refreshed-label="autoRefresh.lastRefreshedLabel.value"
          :next-refresh-label="autoRefresh.nextRefreshLabel.value"
          :status-label="autoRefresh.statusLabel.value"
        />
        <AoiButton appearance="soft" icon="refresh-cw" :loading="loading" :disabled="autoRefresh.refreshDisabled.value" @click="autoRefresh.refreshNow">刷新</AoiButton>
      </template>
    </PageHeader>

    <AoiStatusMessage tone="danger" :message="error" />

    <AoiStatGrid :items="dashboardStats" :columns="4" />

    <section class="two-column-grid">
      <AoiAdminCard title="最近审计" flush>
        <template #actions>
          <AoiButton appearance="plain" icon="arrow-right" to="/audit-logs">全部</AoiButton>
        </template>
        <div class="data-table-wrap">
          <table class="data-table">
            <thead>
              <tr>
                <th>动作</th>
                <th>资源</th>
                <th>用户</th>
                <th>时间</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="log in auditLogs" :key="log.id">
                <td class="mono">{{ log.action }}</td>
                <td>{{ log.resource }} / {{ log.resourceId || "-" }}</td>
                <td>{{ log.userId || "-" }}</td>
                <td>{{ formatDateTime(log.createdAt) }}</td>
              </tr>
              <tr v-if="!auditLogs.length">
                <td colspan="4" class="muted">暂无审计日志。</td>
              </tr>
            </tbody>
          </table>
        </div>
      </AoiAdminCard>

      <AoiAdminCard title="活跃会话">
        <template #actions>
          <AoiButton appearance="plain" icon="arrow-right" to="/sessions">管理</AoiButton>
        </template>
        <div class="dashboard-session-list">
          <div v-for="session in sessions.slice(0, 5)" :key="session.id" class="dashboard-session-list__item">
            <strong>#{{ session.id }}</strong>
            <div class="muted">{{ session.ipAddress }} · {{ formatDateTime(session.lastUsedAt || session.createdAt) }}</div>
          </div>
          <p v-if="!sessions.length" class="muted">暂无会话。</p>
        </div>
      </AoiAdminCard>
    </section>
  </div>
</template>

<style scoped>
.dashboard-session-list {
  display: grid;
  gap: var(--aoi-admin-panel-gap-compact);
}

.dashboard-session-list__item {
  display: grid;
  gap: var(--aoi-admin-card-copy-gap);
  min-width: 0;
}

.dashboard-session-list__item strong,
.dashboard-session-list__item div {
  min-width: 0;
  overflow-wrap: anywhere;
}
</style>


