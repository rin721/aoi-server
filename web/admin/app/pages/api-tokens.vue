<script setup lang="ts">
import type { APIToken, APITokenPage, OrganizationUser, Role } from "~/types/admin"

const api = useAdminApi()
const auth = useAuthStore()
const config = useRuntimeConfig()

const pageData = ref<APITokenPage>({ items: [], page: 1, pageSize: 10, storageStatus: "unavailable", total: 0 })
const users = ref<OrganizationUser[]>([])
const roles = ref<Role[]>([])
const userId = ref("")
const status = ref("")
const page = ref(1)
const pageSize = ref("10")
const loading = ref(false)
const metadataLoading = ref(false)
const issuing = ref(false)
const revoking = ref(false)
const issueOpen = ref(false)
const issuedOpen = ref(false)
const issuedToken = ref("")
const issuedItem = ref<APIToken | null>(null)
const formUserId = ref("")
const formRoleCode = ref("")
const formDays = ref("30")
const formRemark = ref("")
const error = ref("")
const success = ref("")

const statusOptions = [
  { label: "全部", value: "" },
  { label: "有效", value: "active" },
  { label: "已过期", value: "expired" },
  { label: "已撤销", value: "revoked" }
]

const validityOptions = [
  { label: "7 天", value: "7" },
  { label: "30 天", value: "30" },
  { label: "90 天", value: "90" },
  { label: "180 天", value: "180" },
  { label: "365 天", value: "365" },
  { label: "长期有效", value: "-1" }
]

const persisted = computed(() => pageData.value.storageStatus === "persisted")
const pageSizeNumber = computed(() => Math.min(100, Math.max(1, Number(pageSize.value) || 10)))
const totalPages = computed(() => Math.max(1, Math.ceil(pageData.value.total / pageSizeNumber.value)))
const userOptions = computed(() => users.value.map((item) => ({
  label: `${item.user.displayName || item.user.username} (#${item.user.id})`,
  value: item.user.id
})))
const selectedUser = computed(() => users.value.find((item) => item.user.id === formUserId.value) || null)
const selectedRoleCodes = computed(() => new Set((selectedUser.value?.roles || []).map(cleanRoleCode)))
const roleOptions = computed(() => roles.value
  .filter((role) => selectedRoleCodes.value.has(role.code))
  .map((role) => ({
    label: `${role.name} (${role.code})`,
    value: role.code
  })))
const canIssue = computed(() => Boolean(auth.currentOrgId && formUserId.value && formRoleCode.value && !issuing.value))
const apiOrigin = computed(() => config.public.apiBaseURL || (import.meta.client ? window.location.origin : ""))
const tokenExample = computed(() => {
  if (!issuedToken.value) {
    return ""
  }
  return `curl -H "Authorization: Bearer ${issuedToken.value}" "${apiOrigin.value}/api/v1/me"`
})

async function load(options: { silent?: boolean } = {}) {
  if (!auth.currentOrgId) {
    pageData.value = { items: [], page: 1, pageSize: pageSizeNumber.value, storageStatus: "unavailable", total: 0 }
    return
  }

  if (!options.silent) {
    loading.value = true
  }
  error.value = ""
  try {
    pageData.value = await api.listAPITokens(auth.currentOrgId, {
      page: page.value,
      pageSize: pageSizeNumber.value,
      status: status.value || undefined,
      userId: userId.value.trim() || undefined
    })
  } catch (err) {
    error.value = errorMessage(err)
  } finally {
    if (!options.silent) {
      loading.value = false
    }
  }
}

const autoRefresh = useAdminAutoRefresh({
  blocked: computed(() => loading.value || issuing.value || revoking.value),
  load
})

async function loadMetadata() {
  if (!auth.currentOrgId) {
    users.value = []
    roles.value = []
    return
  }

  metadataLoading.value = true
  try {
    const [userResult, roleResult] = await Promise.all([
      api.listUsers(auth.currentOrgId, { pageSize: 100 }),
      api.listRoles(auth.currentOrgId)
    ])
    users.value = userResult.items
    roles.value = roleResult
    alignIssueForm()
  } catch (err) {
    error.value = errorMessage(err)
  } finally {
    metadataLoading.value = false
  }
}

async function openIssueDialog() {
  error.value = ""
  success.value = ""
  issueOpen.value = true
  if (!users.value.length || !roles.value.length) {
    await loadMetadata()
  } else {
    alignIssueForm()
  }
}

async function issueToken() {
  if (!auth.currentOrgId || !canIssue.value) {
    return
  }

  issuing.value = true
  error.value = ""
  success.value = ""
  issuedToken.value = ""
  issuedItem.value = null
  try {
    const created = await api.createAPIToken(auth.currentOrgId, {
      days: Number(formDays.value),
      remark: formRemark.value.trim(),
      roleCode: formRoleCode.value,
      userId: formUserId.value
    })
    issuedToken.value = created.token
    issuedItem.value = created.item
    success.value = "API Token 已签发，请立即复制完整 token。"
    issueOpen.value = false
    issuedOpen.value = true
    formRemark.value = ""
    await load()
  } catch (err) {
    error.value = errorMessage(err)
  } finally {
    issuing.value = false
  }
}

async function revokeToken(item: APIToken) {
  if (!auth.currentOrgId || !confirm(`撤销 API Token #${item.id}？撤销后无法再次使用。`)) {
    return
  }

  revoking.value = true
  error.value = ""
  success.value = ""
  try {
    await api.revokeAPIToken(auth.currentOrgId, item.id)
    success.value = `API Token #${item.id} 已撤销。`
    await load()
  } catch (err) {
    error.value = errorMessage(err)
  } finally {
    revoking.value = false
  }
}

async function resetFilters() {
  userId.value = ""
  status.value = ""
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

async function copyIssuedToken() {
  if (!issuedToken.value || !import.meta.client) {
    return
  }
  try {
    await navigator.clipboard.writeText(issuedToken.value)
    success.value = "完整 token 已复制。"
  } catch {
    success.value = "浏览器未允许复制，请手动选中 token。"
  }
}

async function copyTokenExample() {
  if (!tokenExample.value || !import.meta.client) {
    return
  }
  try {
    await navigator.clipboard.writeText(tokenExample.value)
    success.value = "调用示例已复制。"
  } catch {
    success.value = "浏览器未允许复制，请手动选中调用示例。"
  }
}

function alignIssueForm() {
  if (!formUserId.value || !users.value.some((item) => item.user.id === formUserId.value)) {
    formUserId.value = users.value[0]?.user.id || ""
  }
  if (!formRoleCode.value || !roleOptions.value.some((item) => item.value === formRoleCode.value)) {
    formRoleCode.value = roleOptions.value[0]?.value || ""
  }
}

function cleanRoleCode(value: string) {
  return value.replace(/^role:/, "")
}

function displayUser(item: APIToken) {
  return item.userDisplayName || item.username || item.userId
}

function statusBadgeClass(value: string) {
  if (value === "active") {
    return "badge--success"
  }
  if (value === "expired") {
    return "badge--warning"
  }
  if (value === "revoked") {
    return "badge--danger"
  }
  return "badge--neutral"
}

function isRevocable(item: APIToken) {
  return item.status === "active"
}

onMounted(async () => {
  await Promise.all([
    autoRefresh.refreshNow(),
    loadMetadata()
  ])
})

watch(() => auth.currentOrgId, async () => {
  page.value = 1
  formUserId.value = ""
  formRoleCode.value = ""
  await Promise.all([
    autoRefresh.refreshNow(),
    loadMetadata()
  ])
})

watch(pageSize, async () => {
  page.value = 1
  await autoRefresh.refreshNow()
})

watch([formUserId, users, roles], alignIssueForm)

useHead({
  title: "API Token - Aoi Admin"
})
</script>

<template>
  <div class="page-grid">
    <PageHeader title="API Token" icon="key-round" description="签发和撤销组织内可用于 Bearer 认证的 API Token。">
      <template #actions>
        <AoiButton appearance="soft" icon="plus" :disabled="!auth.currentOrgId" @click="openIssueDialog">签发</AoiButton>
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
    <AoiStatusMessage tone="success" :message="success" />
    <AoiStatusMessage v-if="!auth.currentOrgId" tone="warning" message="请先选择组织后再管理 API Token。" />
    <AoiStatusMessage v-else-if="!persisted" tone="warning" message="iam_api_tokens 表尚不可用，请先执行数据库迁移。" />

    <article class="admin-card">
      <div class="admin-card__header">
        <div>
          <h2>Token 列表</h2>
          <p class="muted">共 {{ pageData.total }} 条，只显示 token 前缀，完整 token 仅在签发成功时展示一次。</p>
        </div>
        <div class="api-token-pager">
          <AoiButton appearance="soft" size="sm" icon="chevron-left" :disabled="page <= 1" @click="previousPage">上一页</AoiButton>
          <span class="badge">{{ page }} / {{ totalPages }}</span>
          <AoiButton appearance="soft" size="sm" icon="chevron-right" :disabled="page >= totalPages" @click="nextPage">下一页</AoiButton>
        </div>
      </div>

      <div class="admin-filter-toolbar api-token-filter-toolbar">
        <AoiTextField v-model="userId" label="用户 ID" icon="user" placeholder="10001" @enter="search" />
        <AoiSelect v-model="status" label="状态" :options="statusOptions" />
        <AoiTextField v-model="pageSize" label="每页" icon="list-filter" type="number" min="1" max="100" step="1" @enter="search" />
        <AoiButton appearance="soft" icon="search" :loading="loading" @click="search">查询</AoiButton>
        <AoiButton appearance="plain" icon="rotate-ccw" @click="resetFilters">重置</AoiButton>
      </div>

      <div class="data-table-wrap api-token-table-wrap">
        <table class="data-table api-token-table">
          <thead>
            <tr>
              <th>ID</th>
              <th>用户</th>
              <th>角色</th>
              <th>状态</th>
              <th>Token 前缀</th>
              <th>过期时间</th>
              <th>最后使用</th>
              <th>备注</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="item in pageData.items" :key="item.id">
              <td class="mono" data-label="ID">{{ item.id }}</td>
              <td data-label="用户">
                <strong>{{ displayUser(item) }}</strong>
                <span class="muted api-token-user-id">#{{ item.userId }}</span>
              </td>
              <td class="mono" data-label="角色">{{ item.roleCode }}</td>
              <td data-label="状态">
                <span class="badge" :class="statusBadgeClass(item.status)">{{ formatStatus(item.status) }}</span>
              </td>
              <td class="mono" data-label="Token 前缀">{{ item.tokenPrefix }}</td>
              <td data-label="过期时间">{{ item.expiresAt ? formatDateTime(item.expiresAt) : "长期有效" }}</td>
              <td data-label="最后使用">{{ formatDateTime(item.lastUsedAt) }}</td>
              <td data-label="备注">{{ item.remark || "-" }}</td>
              <td data-label="操作">
                <AoiButton
                  appearance="soft"
                  intent="danger"
                  icon="ban"
                  size="sm"
                  :disabled="!isRevocable(item) || revoking"
                  @click="revokeToken(item)"
                >
                  撤销
                </AoiButton>
              </td>
            </tr>
            <tr v-if="!pageData.items.length">
              <td colspan="9" class="muted">暂无 API Token。</td>
            </tr>
          </tbody>
        </table>
      </div>
    </article>

    <AoiDialog v-model:open="issueOpen">
      <template #headline>签发 API Token</template>
      <form class="api-token-form" @submit.prevent="issueToken">
        <AoiSelect v-model="formUserId" label="用户" :options="userOptions" :disabled="metadataLoading || issuing" />
        <AoiSelect v-model="formRoleCode" label="角色" :options="roleOptions" :disabled="!formUserId || metadataLoading || issuing" />
        <AoiSelect v-model="formDays" label="有效期" :options="validityOptions" :disabled="issuing" />
        <AoiTextField v-model="formRemark" label="备注" icon="file-text" placeholder="用途、调用方或工单号" :disabled="issuing" multiline :rows="3" />
      </form>
      <AoiStatusMessage v-if="formUserId && !roleOptions.length" tone="warning" message="该用户暂无可用于签发的角色。" />
      <template #actions>
        <AoiButton appearance="plain" :disabled="issuing" @click="issueOpen = false">取消</AoiButton>
        <AoiButton icon="key-round" :loading="issuing" :disabled="!canIssue" @click="issueToken">签发</AoiButton>
      </template>
    </AoiDialog>

    <AoiDialog v-model:open="issuedOpen" :dismissible="false">
      <template #headline>Token 已签发</template>
      <div class="api-token-issued">
        <AoiStatusMessage tone="warning" message="完整 token 只会显示这一次，关闭后只能撤销并重新签发。" />
        <div class="api-token-issued__block">
          <span>完整 token</span>
          <code>{{ issuedToken }}</code>
        </div>
        <div class="api-token-issued__meta">
          <span class="badge">#{{ issuedItem?.id }}</span>
          <span class="badge">{{ issuedItem?.roleCode }}</span>
          <span class="badge">{{ issuedItem?.expiresAt ? formatDateTime(issuedItem.expiresAt) : "长期有效" }}</span>
        </div>
        <div class="api-token-issued__block">
          <span>调用示例</span>
          <code>{{ tokenExample }}</code>
        </div>
      </div>
      <template #actions>
        <AoiButton appearance="soft" icon="copy" @click="copyIssuedToken">复制 token</AoiButton>
        <AoiButton appearance="soft" icon="terminal" @click="copyTokenExample">复制示例</AoiButton>
        <AoiButton icon="check" @click="issuedOpen = false">完成</AoiButton>
      </template>
    </AoiDialog>
  </div>
</template>

<style scoped>
.api-token-pager {
  align-items: center;
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.api-token-filter-toolbar {
  grid-template-columns:
    minmax(140px, 0.8fr)
    minmax(132px, 0.7fr)
    minmax(88px, 0.45fr)
    104px
    96px;
}

.api-token-table {
  min-width: 1080px;
}

.api-token-user-id {
  display: block;
  margin-top: 2px;
}

.api-token-form {
  display: grid;
  gap: 14px;
  min-width: min(520px, 82vw);
}

.api-token-issued {
  display: grid;
  gap: 14px;
  min-width: min(620px, 82vw);
}

.api-token-issued__block {
  border: 1px solid var(--aoi-border-subtle);
  border-radius: 8px;
  display: grid;
  gap: 8px;
  padding: 12px;
}

.api-token-issued__block span {
  color: var(--aoi-text-muted);
  font-size: 0.82rem;
}

.api-token-issued__block code {
  overflow-wrap: anywhere;
  white-space: pre-wrap;
}

.api-token-issued__meta {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

@media (max-width: 760px) {
  .admin-card__header {
    align-items: flex-start;
    flex-direction: column;
  }

  .api-token-filter-toolbar {
    grid-template-columns: 1fr;
  }

  .api-token-table {
    min-width: 0;
    width: 100%;
  }

  .api-token-table thead {
    display: none;
  }

  .api-token-table tr {
    display: grid;
    gap: 8px;
    padding: 14px 0;
  }

  .api-token-table td {
    display: grid;
    grid-template-columns: 92px minmax(0, 1fr);
    gap: 10px;
    border-bottom: 0;
    padding: 0;
  }

  .api-token-table td::before {
    color: var(--aoi-text-muted);
    content: attr(data-label);
    font-size: 0.82rem;
    font-weight: 700;
  }

  .api-token-table td:last-child {
    justify-content: stretch;
  }
}
</style>
