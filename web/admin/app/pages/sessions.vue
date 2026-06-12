<script setup lang="ts">
import type { Session, SessionPage } from "~/types/admin"

const api = useAdminApi()
const auth = useAuthStore()

const emptyPage: SessionPage = { items: [], page: 1, pageSize: 10, storageStatus: "unavailable", total: 0 }
const pageData = ref<SessionPage>({ ...emptyPage })
const filters = reactive({
  scope: "org",
  keyword: "",
  userId: "",
  ipAddress: "",
  status: ""
})
const page = ref(1)
const pageSize = ref("10")
const loading = ref(false)
const error = ref("")
const success = ref("")

const sessions = computed(() => pageData.value.items)
const persisted = computed(() => pageData.value.storageStatus === "persisted")
const pageSizeNumber = computed(() => Math.min(100, Math.max(1, Number(pageSize.value) || 10)))
const totalPages = computed(() => Math.max(1, Math.ceil(pageData.value.total / pageSizeNumber.value)))
const scopeOptions = [
  { label: "当前组织", value: "org" },
  { label: "仅我自己", value: "self" }
]
const statusOptions = [
  { label: "全部状态", value: "" },
  { label: "有效", value: "active" },
  { label: "已撤销", value: "revoked" },
  { label: "已过期", value: "expired" }
]
const pageSizeOptions = [
  { label: "10 条/页", value: "10" },
  { label: "30 条/页", value: "30" },
  { label: "50 条/页", value: "50" },
  { label: "100 条/页", value: "100" }
]

async function load() {
  if (!auth.currentOrgId) {
    pageData.value = { ...emptyPage }
    return
  }

  loading.value = true
  error.value = ""
  try {
    pageData.value = await api.listSessions(auth.currentOrgId, {
      scope: filters.scope === "org" ? "org" : undefined,
      keyword: filters.keyword.trim() || undefined,
      userId: filters.userId.trim() || undefined,
      ipAddress: filters.ipAddress.trim() || undefined,
      status: filters.status || undefined,
      orderKey: "last_used_at",
      desc: true,
      page: page.value,
      pageSize: pageSizeNumber.value
    })
  } catch (err) {
    error.value = errorMessage(err)
  } finally {
    loading.value = false
  }
}

async function search() {
  page.value = 1
  await load()
}

async function resetFilters() {
  filters.scope = "org"
  filters.keyword = ""
  filters.userId = ""
  filters.ipAddress = ""
  filters.status = ""
  page.value = 1
  await load()
}

async function previousPage() {
  if (page.value <= 1) {
    return
  }
  page.value -= 1
  await load()
}

async function nextPage() {
  if (page.value >= totalPages.value) {
    return
  }
  page.value += 1
  await load()
}

async function revoke(session: Session) {
  if (!auth.currentOrgId || !isRevokable(session) || !confirm(`撤销会话 #${session.id}？`)) {
    return
  }

  error.value = ""
  success.value = ""
  try {
    await api.revokeSession(auth.currentOrgId, session.id)
    success.value = `会话 #${session.id} 已撤销。`
    await load()
  } catch (err) {
    error.value = errorMessage(err)
  }
}

function sessionStatus(session: Session) {
  if (session.revokedAt) {
    return "revoked"
  }
  if (new Date(session.expiresAt).getTime() <= Date.now()) {
    return "expired"
  }
  return "active"
}

function sessionStatusLabel(session: Session) {
  const status = sessionStatus(session)
  if (status === "revoked") {
    return "已撤销"
  }
  if (status === "expired") {
    return "已过期"
  }
  return "有效"
}

function sessionStatusClass(session: Session) {
  const status = sessionStatus(session)
  if (status === "active") {
    return "badge--success"
  }
  if (status === "revoked") {
    return "badge--danger"
  }
  return "badge--warning"
}

function isCurrentSession(session: Session) {
  return session.id === auth.sessionId
}

function isRevokable(session: Session) {
  return sessionStatus(session) === "active"
}

function lastUsedAt(session: Session) {
  return session.lastUsedAt || session.createdAt
}

onMounted(load)
watch(() => auth.currentOrgId, () => {
  page.value = 1
  load()
})

useHead({
  title: "会话管理 - Aoi Admin"
})
</script>

<template>
  <div class="page-grid session-page">
    <PageHeader title="会话管理" icon="monitor-check" description="查看当前组织会话，按用户、IP 和状态筛选，并撤销不再可信的会话。">
      <template #actions>
        <AoiButton appearance="soft" icon="refresh-cw" :loading="loading" @click="load">刷新</AoiButton>
      </template>
    </PageHeader>

    <AoiStatusMessage tone="danger" :message="error" />
    <AoiStatusMessage tone="success" :message="success" />

    <section class="session-notice">
      <AoiIcon name="info" decorative />
      <span>注：组织范围查询只返回当前访问上下文所属组织的会话；撤销后 refresh token 将立即失效。</span>
    </section>

    <article class="admin-card">
      <div class="admin-card__header session-card-header">
        <div>
          <h2>会话列表</h2>
          <span>{{ persisted ? "已连接会话数据" : "等待会话数据可用" }}</span>
        </div>
        <span class="badge">{{ pageData.total }} 个</span>
      </div>

      <div class="admin-filter-toolbar session-filter-toolbar">
        <AoiSelect v-model="filters.scope" label="范围" :options="scopeOptions" />
        <AoiTextField v-model="filters.keyword" label="关键字" icon="search" @enter="search" />
        <AoiTextField v-model="filters.userId" label="User ID" icon="user" @enter="search" />
        <AoiTextField v-model="filters.ipAddress" label="IP 地址" icon="network" @enter="search" />
        <AoiSelect v-model="filters.status" label="状态" :options="statusOptions" />
        <AoiSelect v-model="pageSize" label="每页" :options="pageSizeOptions" />
        <AoiButton icon="search" :loading="loading" @click="search">查询</AoiButton>
        <AoiButton appearance="soft" icon="x" @click="resetFilters">重置</AoiButton>
      </div>

      <div class="data-table-wrap session-table-wrap">
        <table class="data-table session-table">
          <thead>
            <tr>
              <th>ID</th>
              <th>User</th>
              <th>IP</th>
              <th>设备</th>
              <th>状态</th>
              <th>最后使用</th>
              <th>过期时间</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="session in sessions" :key="session.id">
              <td class="mono">{{ session.id }}</td>
              <td>
                <div class="session-user-cell">
                  <strong>{{ session.userId }}</strong>
                  <span v-if="isCurrentSession(session)">当前会话</span>
                </div>
              </td>
              <td>{{ session.ipAddress || "-" }}</td>
              <td class="session-agent">{{ session.userAgent || "-" }}</td>
              <td>
                <span class="badge" :class="sessionStatusClass(session)">
                  {{ sessionStatusLabel(session) }}
                </span>
              </td>
              <td>{{ formatDateTime(lastUsedAt(session)) }}</td>
              <td>{{ formatDateTime(session.expiresAt) }}</td>
              <td>
                <AoiButton appearance="soft" intent="danger" icon="ban" :disabled="!isRevokable(session)" @click="revoke(session)">
                  撤销
                </AoiButton>
              </td>
            </tr>
            <tr v-if="!sessions.length">
              <td colspan="8" class="muted">暂无匹配会话。</td>
            </tr>
          </tbody>
        </table>
      </div>

      <div class="session-pagination">
        <span>共 {{ pageData.total }} 个会话</span>
        <div class="action-row">
          <AoiButton appearance="soft" icon="chevron-left" :disabled="page <= 1 || loading" @click="previousPage">上一页</AoiButton>
          <span>{{ page }} / {{ totalPages }}</span>
          <AoiButton appearance="soft" trailing-icon="chevron-right" :disabled="page >= totalPages || loading" @click="nextPage">下一页</AoiButton>
        </div>
      </div>
    </article>
  </div>
</template>

<style scoped>
.session-notice {
  display: flex;
  gap: 10px;
  align-items: flex-start;
  padding: 12px 14px;
  border: 1px solid color-mix(in srgb, var(--aoi-info) 28%, var(--aoi-border));
  border-radius: var(--aoi-radius-md);
  background: color-mix(in srgb, var(--aoi-info) 9%, var(--aoi-surface));
  color: var(--aoi-text);
  font-size: 14px;
  line-height: 1.5;
}

.session-notice :deep(.aoi-icon) {
  flex: 0 0 auto;
  color: var(--aoi-info);
}

.session-card-header > div,
.session-user-cell {
  display: grid;
  gap: 8px;
  min-width: 0;
}

.session-card-header span,
.session-user-cell span {
  color: var(--aoi-text-muted);
  font-size: 13px;
}

.session-user-cell strong {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.session-filter-toolbar {
  grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
}

.session-table-wrap {
  overflow-x: auto;
}

.session-table {
  min-width: 980px;
  table-layout: fixed;
}

.session-table th:nth-child(1),
.session-table td:nth-child(1) {
  width: 150px;
}

.session-table th:nth-child(2),
.session-table td:nth-child(2),
.session-table th:nth-child(3),
.session-table td:nth-child(3) {
  width: 120px;
}

.session-table th:nth-child(5),
.session-table td:nth-child(5),
.session-table th:nth-child(8),
.session-table td:nth-child(8) {
  width: 96px;
}

.session-table th:nth-child(6),
.session-table td:nth-child(6),
.session-table th:nth-child(7),
.session-table td:nth-child(7) {
  width: 148px;
}

.session-table th:nth-child(1),
.session-table td:nth-child(1),
.session-table th:nth-child(3),
.session-table td:nth-child(3) {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.session-agent {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.session-pagination {
  display: flex;
  gap: 12px;
  align-items: center;
  justify-content: space-between;
  padding: 14px 16px;
  border-top: 1px solid var(--aoi-border);
  color: var(--aoi-text-muted);
  font-size: 13px;
}

@media (max-width: 760px) {
  .session-filter-toolbar {
    grid-template-columns: 1fr;
  }

  .session-pagination {
    align-items: flex-start;
    flex-direction: column;
  }
}
</style>
