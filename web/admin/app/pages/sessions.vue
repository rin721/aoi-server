<script setup lang="ts">
import type { Session } from "~/types/api"

const api = useAdminApi()
const auth = useAuthStore()
const sessions = ref<Session[]>([])
const userId = ref("")
const loading = ref(false)
const error = ref("")
const success = ref("")

async function load() {
  if (!auth.currentOrgId) {
    return
  }

  loading.value = true
  error.value = ""
  try {
    const parsedUserId = userId.value.trim() ? Number(userId.value.trim()) : null
    sessions.value = await api.listSessions(auth.currentOrgId, parsedUserId)
  } catch (err) {
    error.value = errorMessage(err)
  } finally {
    loading.value = false
  }
}

async function revoke(sessionId: number) {
  if (!auth.currentOrgId) {
    return
  }

  error.value = ""
  success.value = ""
  try {
    await api.revokeSession(auth.currentOrgId, sessionId)
    success.value = `会话 #${sessionId} 已撤销。`
    await load()
  } catch (err) {
    error.value = errorMessage(err)
  }
}

onMounted(load)
watch(() => auth.currentOrgId, load)

useHead({
  title: "会话 - Aoi Admin"
})
</script>

<template>
  <div class="page-grid">
    <PageHeader title="会话" icon="monitor-check" description="查看当前组织会话，可按 userId 过滤并撤销指定会话。">
      <template #actions>
        <AoiButton appearance="soft" icon="refresh-cw" :loading="loading" @click="load">刷新</AoiButton>
      </template>
    </PageHeader>

    <AoiStatusMessage tone="danger" :message="error" />
    <AoiStatusMessage tone="success" :message="success" />

    <article class="admin-card">
      <div class="admin-card__header">
        <h2>会话列表</h2>
        <div class="toolbar-row">
          <AoiTextField v-model="userId" label="User ID" icon="user" placeholder="可选" @enter="load" />
          <AoiButton appearance="soft" icon="search" @click="load">查询</AoiButton>
        </div>
      </div>
      <div class="data-table-wrap">
        <table class="data-table">
          <thead>
            <tr>
              <th>ID</th>
              <th>User</th>
              <th>IP</th>
              <th>过期时间</th>
              <th>最后使用</th>
              <th>状态</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="session in sessions" :key="session.id">
              <td>{{ session.id }}</td>
              <td>{{ session.userId }}</td>
              <td>{{ session.ipAddress }}</td>
              <td>{{ formatDateTime(session.expiresAt) }}</td>
              <td>{{ formatDateTime(session.lastUsedAt || session.createdAt) }}</td>
              <td>
                <span class="badge" :class="session.revokedAt ? 'badge--danger' : 'badge--success'">
                  {{ session.revokedAt ? "已撤销" : "有效" }}
                </span>
              </td>
              <td>
                <AoiButton appearance="soft" intent="danger" icon="ban" :disabled="Boolean(session.revokedAt)" @click="revoke(session.id)">
                  撤销
                </AoiButton>
              </td>
            </tr>
            <tr v-if="!sessions.length">
              <td colspan="7" class="muted">暂无会话。</td>
            </tr>
          </tbody>
        </table>
      </div>
    </article>
  </div>
</template>
