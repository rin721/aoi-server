<script setup lang="ts">
import type { SystemOperationRecord, SystemOperationRecordPage } from "~/types/admin"

const api = useAdminApi()

const pageData = ref<SystemOperationRecordPage>({ items: [], page: 1, pageSize: 10, storageStatus: "unavailable", total: 0 })
const method = ref("")
const path = ref("")
const status = ref("")
const statusClass = ref("5xx")
const page = ref(1)
const pageSize = ref("10")
const loading = ref(false)
const error = ref("")

const statusClassOptions = [
  { label: "5xx 服务错误", value: "5xx" },
  { label: "4xx 客户端错误", value: "4xx" },
  { label: "全部错误", value: "error" }
]

const persisted = computed(() => pageData.value.storageStatus === "persisted")
const pageSizeNumber = computed(() => Math.min(100, Math.max(1, Number(pageSize.value) || 10)))
const totalPages = computed(() => Math.max(1, Math.ceil(pageData.value.total / pageSizeNumber.value)))

async function load(options: { silent?: boolean } = {}) {
  if (!options.silent) {
    loading.value = true
  }
  error.value = ""
  const exactStatus = status.value.trim()
  try {
    pageData.value = await api.listSystemOperationRecords({
      method: method.value.trim() || undefined,
      page: page.value,
      pageSize: pageSizeNumber.value,
      path: path.value.trim() || undefined,
      status: exactStatus || undefined,
      statusClass: exactStatus ? undefined : statusClass.value
    })
  } catch (err) {
    error.value = errorMessage(err)
  } finally {
    if (!options.silent) {
      loading.value = false
    }
  }
}

const autoRefresh = useAdminAutoRefresh({ blocked: loading, load })

async function resetFilters() {
  method.value = ""
  path.value = ""
  status.value = ""
  statusClass.value = "5xx"
  page.value = 1
  await autoRefresh.refreshNow()
}

async function previousPage() {
  if (page.value <= 1) {
    return
  }
  page.value -= 1
  await autoRefresh.refreshNow()
}

async function nextPage() {
  if (page.value >= totalPages.value) {
    return
  }
  page.value += 1
  await autoRefresh.refreshNow()
}

async function search() {
  page.value = 1
  await autoRefresh.refreshNow()
}

function formatStatusTone(statusCode: number) {
  if (statusCode >= 500) {
    return "badge--danger"
  }
  if (statusCode >= 400) {
    return "badge--warning"
  }
  return "badge--neutral"
}

function displayUser(record: SystemOperationRecord) {
  return record.username || record.userId || "-"
}

function payloadPreview(value: string) {
  if (!value) {
    return "-"
  }
  try {
    return JSON.stringify(JSON.parse(value), null, 2)
  } catch {
    return value
  }
}

function errorPreview(record: SystemOperationRecord) {
  return payloadPreview(record.errorMessage || record.response || record.body)
}

onMounted(autoRefresh.refreshNow)

watch([pageSize, statusClass], async () => {
  page.value = 1
  await autoRefresh.refreshNow()
})

useHead({
  title: "错误日志 - Aoi Admin"
})
</script>

<template>
  <div class="page-grid">
    <PageHeader title="错误日志" icon="bug" description="展示系统操作历史中的异常状态记录，辅助排查后台错误。">
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
    <AoiStatusMessage v-if="!persisted" tone="warning" message="操作记录表尚不可用，请先执行数据库迁移。" />

    <article class="admin-card">
      <div class="admin-card__header">
        <div>
          <h2>异常请求</h2>
          <p class="muted">共 {{ pageData.total }} 条</p>
        </div>
        <div class="error-log-pager">
          <AoiButton appearance="soft" size="sm" icon="chevron-left" :disabled="page <= 1" @click="previousPage">上一页</AoiButton>
          <span class="badge">{{ page }} / {{ totalPages }}</span>
          <AoiButton appearance="soft" size="sm" icon="chevron-right" :disabled="page >= totalPages" @click="nextPage">下一页</AoiButton>
        </div>
      </div>

      <div class="admin-filter-toolbar error-log-filter-toolbar">
        <AoiSelect v-model="statusClass" label="错误范围" :options="statusClassOptions" />
        <AoiTextField v-model="status" label="状态码" icon="hash" placeholder="500" @enter="search" />
        <AoiTextField v-model="method" label="请求方法" icon="activity" placeholder="POST" @enter="search" />
        <AoiTextField v-model="path" label="请求路径" icon="route" placeholder="/api/v1" @enter="search" />
        <AoiTextField v-model="pageSize" label="每页" icon="list-filter" type="number" min="1" max="100" step="1" @enter="search" />
        <AoiButton appearance="soft" icon="search" :loading="loading" @click="search">查询</AoiButton>
        <AoiButton appearance="plain" icon="rotate-ccw" @click="resetFilters">重置</AoiButton>
      </div>

      <div class="data-table-wrap error-log-table-wrap">
        <table class="data-table error-log-table">
          <thead>
            <tr>
              <th>状态码</th>
              <th>请求方法</th>
              <th>请求路径</th>
              <th>操作人</th>
              <th>请求 IP</th>
              <th>错误摘要</th>
              <th>Trace ID</th>
              <th>时间</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="record in pageData.items" :key="record.id">
              <td data-label="状态码">
                <span class="badge" :class="formatStatusTone(record.status)">{{ record.status }}</span>
              </td>
              <td data-label="请求方法">
                <span class="badge">{{ record.method }}</span>
              </td>
              <td class="mono error-log-path" data-label="请求路径">{{ record.path }}</td>
              <td data-label="操作人">{{ displayUser(record) }}</td>
              <td class="mono" data-label="请求 IP">{{ record.ipAddress || "-" }}</td>
              <td data-label="错误摘要">
                <pre v-if="errorPreview(record) !== '-'" class="code-block error-log-payload">{{ errorPreview(record) }}</pre>
                <span v-else class="muted">-</span>
              </td>
              <td class="mono error-log-trace" data-label="Trace ID">{{ record.traceId || "-" }}</td>
              <td data-label="时间">{{ formatDateTime(record.createdAt) }}</td>
            </tr>
            <tr v-if="!pageData.items.length">
              <td colspan="8" class="muted">暂无错误日志。</td>
            </tr>
          </tbody>
        </table>
      </div>
    </article>
  </div>
</template>

<style scoped>
.error-log-pager {
  align-items: center;
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.error-log-filter-toolbar {
  grid-template-columns:
    minmax(146px, 0.9fr)
    minmax(96px, 0.55fr)
    minmax(112px, 0.65fr)
    minmax(160px, 1.2fr)
    minmax(88px, 0.5fr)
    104px
    96px;
}

.error-log-table {
  min-width: 1040px;
}

.error-log-path,
.error-log-payload,
.error-log-trace {
  max-width: 300px;
  overflow-wrap: anywhere;
  white-space: pre-wrap;
}

.error-log-payload {
  margin: 0;
  max-height: 112px;
  overflow: auto;
}

@media (max-width: 760px) {
  .admin-card__header {
    align-items: flex-start;
    flex-direction: column;
  }

  .error-log-filter-toolbar {
    grid-template-columns: 1fr;
  }

  .error-log-table {
    min-width: 0;
    width: 100%;
  }

  .error-log-table thead {
    display: none;
  }

  .error-log-table,
  .error-log-table tbody,
  .error-log-table tr,
  .error-log-table td {
    display: block;
  }

  .error-log-table tr {
    border-bottom: 1px solid var(--aoi-border);
    padding: 12px 0;
    width: 100%;
  }

  .error-log-table tr:last-child {
    border-bottom: 0;
  }

  .error-log-table td,
  .error-log-table td:last-child {
    align-items: flex-start;
    border-bottom: 0;
    display: grid;
    gap: 8px;
    grid-template-columns: 82px minmax(0, 1fr);
    padding: 5px 0;
    white-space: normal;
    width: 100%;
  }

  .error-log-table td::before {
    color: var(--aoi-text-muted);
    content: attr(data-label);
    font-size: 12px;
    font-weight: 700;
  }

  .error-log-table td > * {
    justify-self: start;
    max-width: 100%;
    min-width: 0;
  }

  .error-log-payload {
    justify-self: stretch;
    max-width: 100%;
  }
}
</style>
