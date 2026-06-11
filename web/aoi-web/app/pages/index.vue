<script setup lang="ts">
import type { AuditLog, HealthStatus, ReadyStatus, Session } from "~/types/admin"

const api = useAdminApi()
const auth = useAuthStore()
const loading = ref(false)
const error = ref("")
const health = ref<HealthStatus | null>(null)
const ready = ref<ReadyStatus | null>(null)
const sessions = ref<Session[]>([])
const auditLogs = ref<AuditLog[]>([])

async function refresh() {
  loading.value = true
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
        api.listSessions(auth.currentOrgId),
        api.listAuditLogs(auth.currentOrgId, 6)
      ])
      sessions.value = sessionResult
      auditLogs.value = auditResult
    }
  } catch (err) {
    error.value = errorMessage(err)
  } finally {
    loading.value = false
  }
}

onMounted(refresh)
watch(() => auth.currentOrgId, refresh)

useHead({
  title: "仪表盘 - Aoi Admin"
})
</script>

<template>
  <div class="page-grid">
    <PageHeader title="仪表盘" icon="layout-dashboard" description="查看服务状态、当前组织和最近 IAM 活动。">
      <template #actions>
        <AoiButton appearance="soft" icon="refresh-cw" :loading="loading" @click="refresh">刷新</AoiButton>
      </template>
    </PageHeader>

    <AoiStatusMessage tone="danger" :message="error" />

    <section class="summary-grid">
      <article class="admin-card stat-card">
        <span class="stat-card__label">登录账号</span>
        <strong class="stat-card__value">{{ auth.user?.username || "-" }}</strong>
        <span class="stat-card__meta">{{ auth.user?.email || "-" }}</span>
      </article>
      <article class="admin-card stat-card">
        <span class="stat-card__label">当前组织</span>
        <strong class="stat-card__value">{{ auth.currentOrg?.code || "-" }}</strong>
        <span class="stat-card__meta">{{ auth.currentOrg?.name || "-" }}</span>
      </article>
      <article class="admin-card stat-card">
        <span class="stat-card__label">Health</span>
        <strong class="stat-card__value">{{ health?.status || "-" }}</strong>
        <span class="stat-card__meta">/health</span>
      </article>
      <article class="admin-card stat-card">
        <span class="stat-card__label">Ready</span>
        <strong class="stat-card__value">{{ ready?.status || "-" }}</strong>
        <span class="stat-card__meta">{{ ready?.checks ? Object.entries(ready.checks).map(([k, v]) => `${k}:${v}`).join(" · ") : "/ready" }}</span>
      </article>
    </section>

    <section class="two-column-grid">
      <article class="admin-card">
        <div class="admin-card__header">
          <h2>最近审计</h2>
          <AoiButton appearance="plain" icon="arrow-right" to="/audit-logs">全部</AoiButton>
        </div>
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
      </article>

      <article class="admin-card">
        <div class="admin-card__header">
          <h2>活跃会话</h2>
          <AoiButton appearance="plain" icon="arrow-right" to="/sessions">管理</AoiButton>
        </div>
        <div class="admin-card__body page-grid">
          <div v-for="session in sessions.slice(0, 5)" :key="session.id">
            <strong>#{{ session.id }}</strong>
            <div class="muted">{{ session.ipAddress }} · {{ formatDateTime(session.lastUsedAt || session.createdAt) }}</div>
          </div>
          <p v-if="!sessions.length" class="muted">暂无会话。</p>
        </div>
      </article>
    </section>
  </div>
</template>




