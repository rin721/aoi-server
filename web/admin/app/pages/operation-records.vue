<script setup lang="ts">
import type { ID, SystemOperationRecord, SystemOperationRecordPage } from "~/types/admin"

const api = useAdminApi()
const pageData = ref<SystemOperationRecordPage>({ items: [], page: 1, pageSize: 10, storageStatus: "unavailable", total: 0 })
const method = ref("")
const path = ref("")
const status = ref("")
const page = ref(1)
const pageSize = ref("10")
const selectedIds = ref<ID[]>([])
const loading = ref(false)
const deleting = ref(false)
const error = ref("")

const persisted = computed(() => pageData.value.storageStatus === "persisted")
const pageSizeNumber = computed(() => Math.min(100, Math.max(1, Number(pageSize.value) || 10)))
const totalPages = computed(() => Math.max(1, Math.ceil(pageData.value.total / pageSizeNumber.value)))
const hasSelection = computed(() => selectedIds.value.length > 0)

async function load(options: { silent?: boolean } = {}) {
  if (!options.silent) {
    loading.value = true
  }
  error.value = ""
  try {
    pageData.value = await api.listSystemOperationRecords({
      method: method.value.trim() || undefined,
      page: page.value,
      pageSize: pageSizeNumber.value,
      path: path.value.trim() || undefined,
      status: status.value.trim() || undefined
    })
    selectedIds.value = selectedIds.value.filter((id) => pageData.value.items.some((item) => item.id === id))
  } catch (err) {
    error.value = errorMessage(err)
  } finally {
    if (!options.silent) {
      loading.value = false
    }
  }
}

const autoRefresh = useAdminAutoRefresh({
  blocked: computed(() => loading.value || deleting.value),
  load
})

async function resetFilters() {
  method.value = ""
  path.value = ""
  status.value = ""
  page.value = 1
  await autoRefresh.refreshNow()
}

async function deleteSelected() {
  if (!hasSelection.value || !confirm(`删除 ${selectedIds.value.length} 条操作记录？`)) {
    return
  }
  deleting.value = true
  error.value = ""
  try {
    await api.deleteSystemOperationRecords(selectedIds.value)
    selectedIds.value = []
    await load()
  } catch (err) {
    error.value = errorMessage(err)
  } finally {
    deleting.value = false
  }
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

function toggleAll(event: Event) {
  const checked = (event.target as HTMLInputElement).checked
  selectedIds.value = checked ? pageData.value.items.map((item) => item.id) : []
}

function formatStatusTone(statusCode: number) {
  if (statusCode >= 500) {
    return "badge--danger"
  }
  if (statusCode >= 400) {
    return "badge--warning"
  }
  if (statusCode >= 300) {
    return "badge--neutral"
  }
  return "badge--success"
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

onMounted(autoRefresh.refreshNow)

watch([pageSize], async () => {
  page.value = 1
  await autoRefresh.refreshNow()
})

useHead({
  title: "操作历史 - Aoi Admin"
})
</script>

<template>
  <div class="page-grid">
    <PageHeader title="操作历史" icon="history" description="记录后台受保护 API 的请求方法、路径、状态码、耗时和操作者。">
      <template #actions>
        <AdminAutoRefreshControls
          v-model="autoRefresh.enabled.value"
          :last-refreshed-label="autoRefresh.lastRefreshedLabel.value"
          :next-refresh-label="autoRefresh.nextRefreshLabel.value"
          :status-label="autoRefresh.statusLabel.value"
        />
        <AoiButton appearance="soft" icon="refresh-cw" :loading="loading" :disabled="autoRefresh.refreshDisabled.value" @click="autoRefresh.refreshNow">刷新</AoiButton>
        <AoiButton appearance="soft" intent="danger" icon="trash-2" :disabled="!hasSelection || !persisted" :loading="deleting" @click="deleteSelected">删除</AoiButton>
      </template>
    </PageHeader>

    <AoiStatusMessage tone="danger" :message="error" />
    <AoiStatusMessage v-if="!persisted" tone="warning" message="操作记录表尚不可用，请先执行数据库迁移。" />

    <article class="admin-card">
      <div class="admin-card__header">
        <div>
          <h2>操作记录</h2>
          <p class="muted">共 {{ pageData.total }} 条</p>
        </div>
        <div class="operation-pager">
          <AoiButton appearance="soft" size="sm" icon="chevron-left" :disabled="page <= 1" @click="previousPage">上一页</AoiButton>
          <span class="badge">{{ page }} / {{ totalPages }}</span>
          <AoiButton appearance="soft" size="sm" icon="chevron-right" :disabled="page >= totalPages" @click="nextPage">下一页</AoiButton>
        </div>
      </div>

      <div class="admin-filter-toolbar">
        <AoiTextField v-model="method" label="请求方法" icon="activity" placeholder="GET" @enter="search" />
        <AoiTextField v-model="path" label="请求路径" icon="route" placeholder="/api/v1/system" @enter="search" />
        <AoiTextField v-model="status" label="结果状态码" icon="hash" placeholder="200" @enter="search" />
        <AoiTextField v-model="pageSize" label="每页" icon="list-filter" type="number" min="1" max="100" step="1" @enter="search" />
        <AoiButton appearance="soft" icon="search" :loading="loading" @click="search">查询</AoiButton>
        <AoiButton appearance="plain" icon="rotate-ccw" @click="resetFilters">重置</AoiButton>
      </div>

      <div class="data-table-wrap operation-table-wrap">
        <table class="data-table operation-table">
          <thead>
            <tr>
              <th>
                <input
                  aria-label="选择所有记录"
                  class="operation-check"
                  type="checkbox"
                  :checked="Boolean(pageData.items.length) && selectedIds.length === pageData.items.length"
                  :disabled="!pageData.items.length"
                  @change="toggleAll"
                >
              </th>
              <th>操作人</th>
              <th>日期</th>
              <th>状态码</th>
              <th>请求 IP</th>
              <th>请求方法</th>
              <th>请求路径</th>
              <th>请求</th>
              <th>响应</th>
              <th>耗时</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="record in pageData.items" :key="record.id">
              <td data-label="选择">
                <input v-model="selectedIds" class="operation-check" type="checkbox" :value="record.id" :aria-label="`选择记录 ${record.id}`">
              </td>
              <td data-label="操作人">{{ displayUser(record) }}</td>
              <td data-label="日期">{{ formatDateTime(record.createdAt) }}</td>
              <td data-label="状态码">
                <span class="badge" :class="formatStatusTone(record.status)">{{ record.status }}</span>
              </td>
              <td class="mono" data-label="请求 IP">{{ record.ipAddress || "-" }}</td>
              <td data-label="请求方法">
                <span class="badge">{{ record.method }}</span>
              </td>
              <td class="mono operation-path" data-label="请求路径">{{ record.path }}</td>
              <td data-label="请求">
                <pre v-if="record.body" class="code-block operation-payload">{{ payloadPreview(record.body) }}</pre>
                <span v-else class="muted">-</span>
              </td>
              <td data-label="响应">
                <pre v-if="record.response || record.errorMessage" class="code-block operation-payload">{{ payloadPreview(record.response || record.errorMessage) }}</pre>
                <span v-else class="muted">-</span>
              </td>
              <td class="mono" data-label="耗时">{{ record.latencyMs }} ms</td>
            </tr>
            <tr v-if="!pageData.items.length">
              <td colspan="10" class="muted">暂无操作记录。</td>
            </tr>
          </tbody>
        </table>
      </div>
    </article>
  </div>
</template>

<style scoped>
.operation-pager {
  align-items: center;
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.operation-check {
  accent-color: var(--aoi-accent-60);
  block-size: 16px;
  inline-size: 16px;
}

.operation-path,
.operation-payload {
  max-width: 280px;
  overflow-wrap: anywhere;
  white-space: pre-wrap;
}

.operation-payload {
  margin: 0;
  max-height: 96px;
  overflow: auto;
}

.operation-table th:first-child,
.operation-table td:first-child {
  width: 42px;
}

@media (max-width: 760px) {
  .admin-card__header {
    align-items: flex-start;
    flex-direction: column;
  }

  .operation-table {
    min-width: 0;
    width: 100%;
  }

  .operation-table thead {
    display: none;
  }

  .operation-table,
  .operation-table tbody,
  .operation-table tr,
  .operation-table td {
    display: block;
  }

  .operation-table tr {
    border-bottom: 1px solid var(--aoi-border);
    padding: 10px 0;
    width: 100%;
  }

  .operation-table tr:last-child {
    border-bottom: 0;
  }

  .operation-table td,
  .operation-table td:last-child {
    align-items: flex-start;
    border-bottom: 0;
    display: grid;
    gap: 8px;
    grid-template-columns: 78px minmax(0, 1fr);
    padding: 5px 0;
    white-space: normal;
    width: 100%;
  }

  .operation-table td::before {
    color: var(--aoi-text-muted);
    content: attr(data-label);
    font-size: 12px;
    font-weight: 700;
  }

  .operation-table td > * {
    justify-self: start;
    max-width: 100%;
    min-width: 0;
  }

  .operation-payload {
    justify-self: stretch;
    max-width: 100%;
  }
}
</style>
