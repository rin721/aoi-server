<script setup lang="ts">
import type { SystemAPIEntry, SystemAPIGroup } from "~/types/admin"

const api = useAdminApi()
const groups = ref<SystemAPIGroup[]>([])
const query = ref("")
const method = ref("")
const groupCode = ref("")
const accessMode = ref("")
const loading = ref(false)
const syncing = ref(false)
const syncingPermissions = ref(false)
const error = ref("")
const success = ref("")

const totalCount = computed(() => groups.value.reduce((count, group) => count + group.items.length, 0))
const syncedCount = computed(() => groups.value.reduce((count, group) => count + group.items.filter((entry) => entry.synced).length, 0))
const unsyncedCount = computed(() => Math.max(totalCount.value - syncedCount.value, 0))
const protectedCount = computed(() => groups.value.reduce((count, group) => count + group.items.filter((entry) => Boolean(entry.permission)).length, 0))
const registeredPermissionCount = computed(() => groups.value.reduce((count, group) => count + group.items.filter((entry) => entry.permission && entry.permissionRegistered).length, 0))
const unregisteredPermissionCount = computed(() => Math.max(protectedCount.value - registeredPermissionCount.value, 0))
const publicCount = computed(() => groups.value.reduce((count, group) => count + group.items.filter((entry) => entry.access === "public").length, 0))
const authenticatedCount = computed(() => groups.value.reduce((count, group) => count + group.items.filter((entry) => entry.access === "authenticated").length, 0))
const permissionCount = computed(() => groups.value.reduce((count, group) => count + group.items.filter((entry) => entry.access === "permission").length, 0))

const groupOptions = computed(() => [
  { label: "全部分组", value: "" },
  ...groups.value.map((group) => ({ label: `${group.label} (${group.count})`, value: group.code }))
])

const methodOptions = [
  { label: "全部方法", value: "" },
  { label: "GET", value: "GET" },
  { label: "POST", value: "POST" },
  { label: "PATCH", value: "PATCH" },
  { label: "PUT", value: "PUT" },
  { label: "DELETE", value: "DELETE" }
]

const accessOptions = [
  { label: "全部访问属性", value: "" },
  { label: "公开接口", value: "public" },
  { label: "登录后接口", value: "authenticated" },
  { label: "权限接口", value: "permission" }
]

const filteredGroups = computed(() => {
  const keyword = query.value.trim().toLowerCase()
  return groups.value
    .filter((group) => !groupCode.value || group.code === groupCode.value)
    .map((group) => ({
      ...group,
      items: group.items.filter((entry) => matchesAPI(entry, keyword))
    }))
    .filter((group) => group.items.length > 0)
})

async function load(options: { silent?: boolean } = {}) {
  if (!options.silent) {
    loading.value = true
  }
  error.value = ""
  success.value = ""
  try {
    groups.value = await api.listSystemAPIs()
  } catch (err) {
    error.value = errorMessage(err)
  } finally {
    if (!options.silent) {
      loading.value = false
    }
  }
}

const autoRefresh = useAdminAutoRefresh({
  blocked: computed(() => loading.value || syncing.value || syncingPermissions.value),
  load
})

async function syncAPIs() {
  syncing.value = true
  error.value = ""
  success.value = ""
  try {
    const result = await api.syncSystemAPIs()
    groups.value = result.groups
    if (result.persisted) {
      success.value = `已同步 ${result.total} 个接口，新增 ${result.created} 个，更新 ${result.updated} 个，标记过期 ${result.stale} 个。`
    } else if (result.storageStatus === "unavailable") {
      success.value = "已刷新当前接口目录；system_apis 表尚未创建，暂未写入同步状态。"
    } else {
      success.value = "已刷新当前接口目录；当前环境未启用接口同步存储。"
    }
  } catch (err) {
    error.value = errorMessage(err)
  } finally {
    syncing.value = false
  }
}

async function syncPermissions() {
  syncingPermissions.value = true
  error.value = ""
  success.value = ""
  try {
    const result = await api.syncSystemAPIPermissions()
    let message = ""
    if (result.persisted) {
      message = `已同步 ${result.total} 个权限码，新增 ${result.created} 个，跳过已有 ${result.skipped} 个。`
    } else {
      message = "当前环境未启用权限字典存储，暂未写入权限码。"
    }
    await load()
    success.value = message
  } catch (err) {
    error.value = errorMessage(err)
  } finally {
    syncingPermissions.value = false
  }
}

function matchesAPI(entry: SystemAPIEntry, keyword: string) {
  if (method.value && entry.method !== method.value) {
    return false
  }
  if (accessMode.value && entry.access !== accessMode.value) {
    return false
  }
  if (!keyword) {
    return true
  }
  return [
    entry.code,
    entry.description,
    entry.group,
    entry.method,
    entry.path,
    accessLabel(entry.access),
    entry.permission || "",
    entry.synced ? "synced 已同步" : "unsynced 未同步",
    entry.permissionRegistered ? "registered 已登记" : "unregistered 未登记"
  ].some((value) => value.toLowerCase().includes(keyword))
}

function methodClass(value: string) {
  return `api-method api-method--${value.toLowerCase()}`
}

function accessLabel(value: SystemAPIEntry["access"]) {
  switch (value) {
    case "public":
      return "公开"
    case "permission":
      return "权限"
    default:
      return "登录"
  }
}

function accessClass(value: SystemAPIEntry["access"]) {
  return {
    "badge--success": value === "public",
    "badge--warning": value === "authenticated",
    "badge--neutral": value === "permission"
  }
}

onMounted(autoRefresh.refreshNow)

useHead({
  title: "API 管理 - Aoi Admin"
})
</script>

<template>
  <div class="page-grid">
    <PageHeader title="API 管理" icon="code-2" description="查看当前后端注册的 HTTP API 目录，作为权限、菜单和审计治理的基础索引。">
      <template #actions>
        <AdminAutoRefreshControls
          v-model="autoRefresh.enabled.value"
          :last-refreshed-label="autoRefresh.lastRefreshedLabel.value"
          :next-refresh-label="autoRefresh.nextRefreshLabel.value"
          :status-label="autoRefresh.statusLabel.value"
        />
        <AoiButton appearance="soft" icon="refresh-cw" :loading="loading" :disabled="autoRefresh.refreshDisabled.value" @click="autoRefresh.refreshNow">刷新</AoiButton>
        <AoiButton icon="repeat" :loading="syncing" @click="syncAPIs">同步路由</AoiButton>
        <AoiButton icon="shield-check" :loading="syncingPermissions" @click="syncPermissions">同步权限</AoiButton>
      </template>
    </PageHeader>

    <AoiStatusMessage tone="danger" :message="error" />
    <AoiStatusMessage tone="success" :message="success" />

    <article class="admin-card">
      <div class="admin-card__header">
        <h2>接口目录</h2>
        <div class="api-summary">
          <span class="badge">{{ totalCount }} 个</span>
          <span class="badge badge--success">{{ publicCount }} 公开</span>
          <span class="badge badge--warning">{{ authenticatedCount }} 登录</span>
          <span class="badge">{{ permissionCount }} 权限</span>
          <span class="badge badge--success">{{ syncedCount }} 已同步</span>
          <span v-if="unsyncedCount" class="badge badge--warning">{{ unsyncedCount }} 未同步</span>
          <span class="badge badge--success">{{ registeredPermissionCount }} 权限已登记</span>
          <span v-if="unregisteredPermissionCount" class="badge badge--warning">{{ unregisteredPermissionCount }} 权限未登记</span>
        </div>
      </div>
      <div class="admin-filter-toolbar">
        <AoiTextField v-model="query" label="关键字" icon="search" placeholder="/api/v1/orgs" @enter="autoRefresh.refreshNow" />
        <AoiSelect
          :model-value="groupCode"
          label="分组"
          :options="groupOptions"
          @update:model-value="groupCode = $event"
        />
        <AoiSelect
          :model-value="method"
          label="Method"
          :options="methodOptions"
          @update:model-value="method = $event"
        />
        <AoiSelect
          :model-value="accessMode"
          label="访问属性"
          :options="accessOptions"
          @update:model-value="accessMode = $event"
        />
      </div>

      <div class="api-groups">
        <section v-for="group in filteredGroups" :key="group.code" class="api-group">
          <div class="api-group__header">
            <div>
              <h3>{{ group.label }}</h3>
              <p>{{ group.code }}</p>
            </div>
            <span class="badge">{{ group.items.length }} 个</span>
          </div>
          <div class="data-table-wrap">
            <table class="data-table api-table">
              <thead>
                <tr>
                  <th>Method</th>
                  <th>Path</th>
                  <th>访问</th>
                  <th>权限</th>
                  <th>登记</th>
                  <th>同步</th>
                  <th>说明</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="entry in group.items" :key="entry.code">
                  <td data-label="Method"><span :class="methodClass(entry.method)">{{ entry.method }}</span></td>
                  <td class="mono api-table__path" data-label="Path">{{ entry.path }}</td>
                  <td data-label="访问">
                    <span :class="['badge', accessClass(entry.access)]">{{ accessLabel(entry.access) }}</span>
                  </td>
                  <td data-label="权限">
                    <span v-if="entry.permission" class="badge">{{ entry.permission }}</span>
                    <span v-else class="muted">未绑定权限</span>
                  </td>
                  <td data-label="登记">
                    <span v-if="entry.permission" :class="['badge', entry.permissionRegistered ? 'badge--success' : 'badge--warning']">
                      {{ entry.permissionRegistered ? "已登记" : "未登记" }}
                    </span>
                    <span v-else class="muted">无权限码</span>
                  </td>
                  <td data-label="同步">
                    <span :class="['badge', entry.synced ? 'badge--success' : 'badge--warning']">
                      {{ entry.synced ? "已同步" : "未同步" }}
                    </span>
                    <span v-if="entry.syncedAt" class="api-table__synced-at">{{ formatDateTime(entry.syncedAt) }}</span>
                  </td>
                  <td data-label="说明">{{ entry.description }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </section>

        <p v-if="!loading && filteredGroups.length === 0" class="api-empty muted">暂无匹配接口。</p>
      </div>
    </article>
  </div>
</template>

<style scoped>
.api-groups {
  display: grid;
  gap: 16px;
  padding: 0 14px 14px;
}

.api-group {
  display: grid;
  gap: 10px;
}

.api-summary {
  align-items: center;
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  justify-content: flex-end;
}

.api-group__header {
  align-items: center;
  display: flex;
  justify-content: space-between;
  gap: 12px;
}

.api-group__header h3,
.api-group__header p {
  margin: 0;
}

.api-group__header h3 {
  font-size: 15px;
}

.api-group__header p {
  color: var(--aoi-text-muted);
  font-size: 12px;
  margin-top: 3px;
}

.api-table th:first-child,
.api-table td:first-child {
  width: 96px;
}

.api-table__path {
  min-width: 260px;
}

.api-table__synced-at {
  color: var(--aoi-text-muted);
  display: block;
  font-size: 12px;
  margin-top: 6px;
  white-space: nowrap;
}

.api-method {
  border: 1px solid var(--aoi-border);
  border-radius: 6px;
  display: inline-flex;
  font-size: 12px;
  font-weight: 800;
  justify-content: center;
  min-width: 58px;
  padding: 4px 8px;
}

.api-method--get {
  background: var(--aoi-intent-success-soft-bg);
  border-color: var(--aoi-intent-success-border);
  color: var(--aoi-intent-success-color);
}

.api-method--post {
  background: var(--aoi-intent-info-soft-bg);
  border-color: var(--aoi-intent-info-border);
  color: var(--aoi-intent-info-color);
}

.api-method--patch,
.api-method--put {
  background: var(--aoi-intent-warning-soft-bg);
  border-color: var(--aoi-intent-warning-border);
  color: var(--aoi-intent-warning-color);
}

.api-method--delete {
  background: var(--aoi-intent-danger-soft-bg);
  border-color: var(--aoi-intent-danger-border);
  color: var(--aoi-intent-danger-color);
}

.api-empty {
  margin: 0;
  padding: 18px 0 4px;
}

@media (max-width: 640px) {
  .admin-card__header {
    align-items: flex-start;
    flex-direction: column;
    gap: 10px;
  }

  .admin-card__header h2 {
    width: 100%;
  }

  .api-groups {
    padding-inline: 14px;
  }

  .api-summary {
    justify-content: flex-start;
    width: 100%;
  }

  .api-table,
  .api-table tbody,
  .api-table tr,
  .api-table td {
    display: block;
  }

  .api-table thead {
    display: none;
  }

  .api-table tr {
    border-bottom: 1px solid var(--aoi-border);
    padding: 10px 0;
  }

  .api-table tr:last-child {
    border-bottom: 0;
  }

  .api-table td {
    align-items: flex-start;
    border-bottom: 0;
    display: grid;
    gap: 8px;
    grid-template-columns: 72px minmax(0, 1fr);
    padding: 5px 0;
  }

  .api-table td::before {
    color: var(--aoi-text-muted);
    content: attr(data-label);
    font-size: 12px;
    font-weight: 700;
  }

  .api-table td .badge {
    justify-self: start;
  }

  .api-table__path {
    min-width: 0;
    overflow-wrap: anywhere;
  }

  .api-table__synced-at {
    white-space: normal;
  }
}
</style>
