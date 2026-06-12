<script setup lang="ts">
import type { AuditLog } from "~/types/admin"

const api = useAdminApi()
const auth = useAuthStore()

const logs = ref<AuditLog[]>([])
const userId = ref("")
const ipAddress = ref("")
const from = ref("")
const to = ref("")
const limit = ref("50")
const loading = ref(false)
const error = ref("")

const limitNumber = computed(() => Math.min(200, Math.max(1, Number(limit.value) || 50)))
const loginLogs = computed(() => logs.value.filter((item) => item.action === "auth.login"))

async function load() {
  if (!auth.currentOrgId) {
    logs.value = []
    return
  }

  loading.value = true
  error.value = ""
  try {
    const items = await api.listAuditLogs(auth.currentOrgId, {
      action: "auth.login",
      from: from.value ? new Date(from.value).toISOString() : undefined,
      limit: limitNumber.value,
      to: to.value ? new Date(to.value).toISOString() : undefined,
      userId: userId.value.trim() || undefined
    })
    const ip = ipAddress.value.trim()
    logs.value = ip ? items.filter((item) => item.ipAddress.includes(ip)) : items
  } catch (err) {
    error.value = errorMessage(err)
  } finally {
    loading.value = false
  }
}

async function resetFilters() {
  userId.value = ""
  ipAddress.value = ""
  from.value = ""
  to.value = ""
  limit.value = "50"
  await load()
}

function deviceName(value: string) {
  if (!value) {
    return "-"
  }
  if (value.includes("Firefox")) {
    return "Firefox"
  }
  if (value.includes("Edg/")) {
    return "Edge"
  }
  if (value.includes("Chrome")) {
    return "Chrome"
  }
  if (value.includes("Safari")) {
    return "Safari"
  }
  return value.slice(0, 48)
}

onMounted(load)
watch(() => auth.currentOrgId, load)

useHead({
  title: "登录日志 - Aoi Admin"
})
</script>

<template>
  <div class="page-grid">
    <PageHeader title="登录日志" icon="log-in" description="展示当前组织的登录审计记录，数据来自 IAM audit log 的 auth.login 事件。">
      <template #actions>
        <AoiButton appearance="soft" icon="refresh-cw" :loading="loading" @click="load">刷新</AoiButton>
      </template>
    </PageHeader>

    <AoiStatusMessage tone="danger" :message="error" />
    <AoiStatusMessage v-if="!auth.currentOrgId" tone="warning" message="请先选择组织后查看登录日志。" />

    <article class="admin-card">
      <div class="admin-card__header">
        <div>
          <h2>登录记录</h2>
          <p class="muted">共 {{ loginLogs.length }} 条，默认只显示 auth.login。</p>
        </div>
      </div>

      <div class="admin-filter-toolbar login-log-filter-toolbar">
        <AoiTextField v-model="userId" label="用户 ID" icon="user" placeholder="10001" @enter="load" />
        <AoiTextField v-model="ipAddress" label="登录 IP" icon="map-pin" placeholder="127.0.0.1" @enter="load" />
        <AoiTextField v-model="from" label="开始时间" type="datetime-local" icon="calendar" @enter="load" />
        <AoiTextField v-model="to" label="结束时间" type="datetime-local" icon="calendar" @enter="load" />
        <AoiTextField v-model="limit" label="数量" type="number" min="1" max="200" step="1" icon="list-filter" @enter="load" />
        <AoiButton appearance="soft" icon="search" :loading="loading" @click="load">查询</AoiButton>
        <AoiButton appearance="plain" icon="rotate-ccw" @click="resetFilters">重置</AoiButton>
      </div>

      <div class="data-table-wrap login-log-table-wrap">
        <table class="data-table login-log-table">
          <thead>
            <tr>
              <th>编号</th>
              <th>用户 ID</th>
              <th>登录 IP</th>
              <th>设备</th>
              <th>资源</th>
              <th>登录时间</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="log in loginLogs" :key="log.id">
              <td class="mono" data-label="编号">{{ log.id }}</td>
              <td data-label="用户 ID">{{ log.userId || "-" }}</td>
              <td class="mono" data-label="登录 IP">{{ log.ipAddress || "-" }}</td>
              <td data-label="设备">{{ deviceName(log.userAgent) }}</td>
              <td data-label="资源">{{ log.resource || "-" }}</td>
              <td data-label="登录时间">{{ formatDateTime(log.createdAt) }}</td>
            </tr>
            <tr v-if="!loginLogs.length">
              <td colspan="6" class="muted">暂无登录日志。</td>
            </tr>
          </tbody>
        </table>
      </div>
    </article>
  </div>
</template>

<style scoped>
.login-log-table {
  min-width: 860px;
}

.login-log-filter-toolbar {
  grid-template-columns:
    minmax(128px, 0.75fr)
    minmax(140px, 0.9fr)
    minmax(150px, 1fr)
    minmax(150px, 1fr)
    minmax(96px, 0.55fr)
    104px
    96px;
}

.login-log-table th:first-child,
.login-log-table td:first-child {
  width: 180px;
}

@media (max-width: 760px) {
  .login-log-filter-toolbar {
    grid-template-columns: 1fr;
  }

  .login-log-table {
    min-width: 0;
    width: 100%;
  }

  .login-log-table thead {
    display: none;
  }

  .login-log-table tr {
    display: grid;
    gap: 8px;
    padding: 14px 0;
  }

  .login-log-table td {
    display: grid;
    grid-template-columns: 88px minmax(0, 1fr);
    gap: 10px;
    border-bottom: 0;
    padding: 0;
  }

  .login-log-table td::before {
    color: var(--aoi-text-muted);
    content: attr(data-label);
    font-size: 12px;
    font-weight: 700;
  }
}
</style>
