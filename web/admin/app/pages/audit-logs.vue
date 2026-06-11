<script setup lang="ts">
import type { AuditLog } from "~/types/api"

const api = useAdminApi()
const auth = useAuthStore()
const logs = ref<AuditLog[]>([])
const limit = ref("100")
const action = ref("")
const userId = ref("")
const from = ref("")
const to = ref("")
const cursor = ref("")
const loading = ref(false)
const error = ref("")

async function load() {
  if (!auth.currentOrgId) {
    return
  }

  loading.value = true
  error.value = ""
  try {
    logs.value = await api.listAuditLogs(auth.currentOrgId, {
      action: action.value.trim() || undefined,
      cursor: cursor.value.trim() || undefined,
      from: from.value ? new Date(from.value).toISOString() : undefined,
      limit: Number(limit.value) || 100,
      to: to.value ? new Date(to.value).toISOString() : undefined,
      userId: userId.value.trim() || undefined
    })
  } catch (err) {
    error.value = errorMessage(err)
  } finally {
    loading.value = false
  }
}

function prettyMetadata(value: string) {
  if (!value) {
    return "-"
  }

  try {
    return JSON.stringify(JSON.parse(value), null, 2)
  } catch {
    return value
  }
}

onMounted(load)
watch(() => auth.currentOrgId, load)

useHead({
  title: "审计日志 - Aoi Admin"
})
</script>

<template>
  <div class="page-grid">
    <PageHeader title="审计日志" icon="scroll-text" description="按当前组织读取 IAM 审计记录，用于定位管理操作。">
      <template #actions>
        <AoiButton appearance="soft" icon="refresh-cw" :loading="loading" @click="load">刷新</AoiButton>
      </template>
    </PageHeader>

    <AoiStatusMessage tone="danger" :message="error" />

    <article class="admin-card">
      <div class="admin-card__header">
        <h2>日志</h2>
        <div class="toolbar-row">
          <AoiTextField v-model="action" label="Action" icon="activity" placeholder="auth.login" @enter="load" />
          <AoiTextField v-model="userId" label="User ID" icon="user" @enter="load" />
          <AoiTextField v-model="from" label="From" type="datetime-local" icon="calendar" @enter="load" />
          <AoiTextField v-model="to" label="To" type="datetime-local" icon="calendar" @enter="load" />
          <AoiTextField v-model="cursor" label="Cursor" icon="chevrons-down" @enter="load" />
          <AoiTextField v-model="limit" label="Limit" type="number" icon="list-filter" @enter="load" />
          <AoiButton appearance="soft" icon="search" @click="load">查询</AoiButton>
        </div>
      </div>
      <div class="data-table-wrap">
        <table class="data-table">
          <thead>
            <tr>
              <th>ID</th>
              <th>动作</th>
              <th>资源</th>
              <th>User</th>
              <th>IP</th>
              <th>时间</th>
              <th>Metadata</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="log in logs" :key="log.id">
              <td>{{ log.id }}</td>
              <td class="mono">{{ log.action }}</td>
              <td>{{ log.resource }} / {{ log.resourceId || "-" }}</td>
              <td>{{ log.userId || "-" }}</td>
              <td>{{ log.ipAddress || "-" }}</td>
              <td>{{ formatDateTime(log.createdAt) }}</td>
              <td><pre class="code-block">{{ prettyMetadata(log.metadata) }}</pre></td>
            </tr>
            <tr v-if="!logs.length">
              <td colspan="7" class="muted">暂无审计日志。</td>
            </tr>
          </tbody>
        </table>
      </div>
    </article>
  </div>
</template>
