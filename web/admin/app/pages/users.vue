<script setup lang="ts">
import type { Invitation, OrganizationUser, OrganizationUserPage, Role } from "~/types/admin"

const api = useAdminApi()
const auth = useAuthStore()

const emptyPage: OrganizationUserPage = { items: [], page: 1, pageSize: 10, storageStatus: "unavailable", total: 0 }
const pageData = ref<OrganizationUserPage>({ ...emptyPage })
const roles = ref<Role[]>([])
const invitations = ref<Invitation[]>([])
const roleDrafts = ref<Record<string, string>>({})
const filters = reactive({
  keyword: "",
  username: "",
  displayName: "",
  email: "",
  roleCode: "",
  status: ""
})
const page = ref(1)
const pageSize = ref("10")
const email = ref("")
const roleCode = ref("member")
const inviteToken = ref("")
const inviteUrl = ref("")
const loading = ref(false)
const saving = ref(false)
const error = ref("")
const success = ref("")

const users = computed(() => pageData.value.items)
const pageSizeNumber = computed(() => Math.min(100, Math.max(1, Number(pageSize.value) || 10)))
const totalPages = computed(() => Math.max(1, Math.ceil(pageData.value.total / pageSizeNumber.value)))
const roleOptions = computed(() => roles.value.map((role) => ({
  label: `${role.name} (${role.code})`,
  value: role.code
})))
const roleFilterOptions = computed(() => [
  { label: "全部角色", value: "" },
  ...roleOptions.value
])
const statusOptions = [
  { label: "全部状态", value: "" },
  { label: "启用", value: "active" },
  { label: "禁用", value: "disabled" }
]
const pageSizeOptions = [
  { label: "10 条/页", value: "10" },
  { label: "30 条/页", value: "30" },
  { label: "50 条/页", value: "50" },
  { label: "100 条/页", value: "100" }
]
const canInvite = computed(() => Boolean(email.value.trim() && roleCode.value && !saving.value))
const persisted = computed(() => pageData.value.storageStatus === "persisted")

async function load() {
  if (!auth.currentOrgId) {
    pageData.value = { ...emptyPage }
    roles.value = []
    invitations.value = []
    return
  }

  loading.value = true
  error.value = ""
  try {
    const [userResult, roleResult, invitationResult] = await Promise.all([
      api.listUsers(auth.currentOrgId, {
        keyword: filters.keyword.trim() || undefined,
        username: filters.username.trim() || undefined,
        displayName: filters.displayName.trim() || undefined,
        email: filters.email.trim() || undefined,
        roleCode: filters.roleCode || undefined,
        status: filters.status || undefined,
        orderKey: "id",
        desc: true,
        page: page.value,
        pageSize: pageSizeNumber.value
      }),
      api.listRoles(auth.currentOrgId),
      api.listInvitations(auth.currentOrgId)
    ])
    pageData.value = userResult
    roles.value = roleResult
    invitations.value = invitationResult
    if (!roles.value.some((role) => role.code === roleCode.value)) {
      roleCode.value = roles.value[0]?.code || "member"
    }
    roleDrafts.value = Object.fromEntries(users.value.map((item) => [
      item.user.id,
      firstRoleCode(item.roles) || roleCode.value
    ]))
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
  filters.keyword = ""
  filters.username = ""
  filters.displayName = ""
  filters.email = ""
  filters.roleCode = ""
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

async function invite() {
  if (!auth.currentOrgId || !canInvite.value) {
    return
  }

  saving.value = true
  error.value = ""
  success.value = ""
  inviteToken.value = ""
  inviteUrl.value = ""
  try {
    const result = await api.inviteUser(auth.currentOrgId, {
      email: email.value.trim(),
      roleCode: roleCode.value
    })
    inviteToken.value = result.token || ""
    inviteUrl.value = result.url || ""
    success.value = inviteToken.value || inviteUrl.value ? "邀请已创建，调试信息如下。" : "邀请已创建。"
    email.value = ""
    await load()
  } catch (err) {
    error.value = errorMessage(err)
  } finally {
    saving.value = false
  }
}

async function updateMemberStatus(item: OrganizationUser, status: "active" | "disabled") {
  if (!auth.currentOrgId) {
    return
  }
  error.value = ""
  success.value = ""
  try {
    await api.updateUser(auth.currentOrgId, item.user.id, { status })
    success.value = `${item.user.username} 已${status === "active" ? "启用" : "禁用"}。`
    await load()
  } catch (err) {
    error.value = errorMessage(err)
  }
}

async function updateMemberRole(item: OrganizationUser) {
  const role = roleDrafts.value[item.user.id]
  if (!auth.currentOrgId || !role) {
    return
  }
  error.value = ""
  success.value = ""
  try {
    await api.updateUser(auth.currentOrgId, item.user.id, { roles: [role] })
    success.value = `${item.user.username} 的角色已更新。`
    await load()
  } catch (err) {
    error.value = errorMessage(err)
  }
}

async function revokeInvitation(invitation: Invitation) {
  if (!auth.currentOrgId || !confirm(`撤销 ${invitation.email} 的邀请？`)) {
    return
  }
  error.value = ""
  success.value = ""
  try {
    await api.revokeInvitation(auth.currentOrgId, invitation.id)
    success.value = "邀请已撤销。"
    await load()
  } catch (err) {
    error.value = errorMessage(err)
  }
}

function firstRoleCode(roles: string[]) {
  return roles[0]?.replace("role:", "") || ""
}

function roleLabel(role: string) {
  return role.replace("role:", "")
}

function displayUserName(item: OrganizationUser) {
  return item.user.displayName || item.user.username
}

onMounted(load)
watch(() => auth.currentOrgId, () => {
  page.value = 1
  load()
})

useHead({
  title: "用户管理 - Aoi Admin"
})
</script>

<template>
  <div class="page-grid user-page">
    <PageHeader title="用户管理" icon="users" description="按当前组织管理成员、角色和启停状态。">
      <template #actions>
        <AoiButton appearance="soft" icon="refresh-cw" :loading="loading" @click="load">刷新</AoiButton>
      </template>
    </PageHeader>

    <AoiStatusMessage tone="danger" :message="error" />
    <AoiStatusMessage tone="success" :message="success" />
    <AoiStatusMessage v-if="inviteToken" tone="success" :message="`调试 token：${inviteToken}`" />
    <AoiStatusMessage v-if="inviteUrl" tone="success" :message="`调试链接：${inviteUrl}`" />

    <section class="user-notice">
      <AoiIcon name="info" decorative />
      <span>注：本系统按当前组织和当前登录角色进行权限判断，成员列表仅展示当前组织范围。</span>
    </section>

    <section class="user-workspace">
      <article class="admin-card user-main-card">
        <div class="admin-card__header user-card-header">
          <div>
            <h2>组织用户</h2>
            <span>{{ persisted ? "已连接组织成员表" : "等待用户数据可用" }}</span>
          </div>
          <span class="badge">{{ pageData.total }} 人</span>
        </div>

        <div class="admin-filter-toolbar user-filter-toolbar">
          <AoiTextField v-model="filters.keyword" label="关键字" icon="search" @enter="search" />
          <AoiTextField v-model="filters.username" label="用户名" icon="user" @enter="search" />
          <AoiTextField v-model="filters.displayName" label="昵称 / 显示名" icon="badge" @enter="search" />
          <AoiTextField v-model="filters.email" label="邮箱" icon="mail" @enter="search" />
          <AoiSelect v-model="filters.roleCode" label="角色" :options="roleFilterOptions" />
          <AoiSelect v-model="filters.status" label="状态" :options="statusOptions" />
          <AoiSelect v-model="pageSize" label="每页" :options="pageSizeOptions" />
          <AoiButton icon="search" :loading="loading" @click="search">查询</AoiButton>
          <AoiButton appearance="soft" icon="x" @click="resetFilters">重置</AoiButton>
        </div>

        <div class="data-table-wrap user-table-wrap">
          <table class="data-table user-table">
            <thead>
              <tr>
                <th>ID</th>
                <th>用户</th>
                <th>邮箱</th>
                <th>角色</th>
                <th>成员状态</th>
                <th>MFA</th>
                <th>最近登录</th>
                <th>操作</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="item in users" :key="item.user.id">
                <td class="mono">{{ item.user.id }}</td>
                <td>
                  <div class="user-identity">
                    <strong>{{ displayUserName(item) }}</strong>
                    <span>{{ item.user.username }}</span>
                  </div>
                </td>
                <td>{{ item.user.email }}</td>
                <td>
                  <div class="user-role-list">
                    <span v-for="role in item.roles" :key="role" class="badge">{{ roleLabel(role) }}</span>
                  </div>
                </td>
                <td><span class="badge" :class="item.membershipStatus === 'active' ? 'badge--success' : 'badge--warning'">{{ formatStatus(item.membershipStatus) }}</span></td>
                <td><span class="badge" :class="item.user.mfaEnabled ? 'badge--success' : ''">{{ item.user.mfaEnabled ? "已启用" : "未启用" }}</span></td>
                <td>{{ formatDateTime(item.user.lastLoginAt) }}</td>
                <td>
                  <div class="user-actions">
                    <AoiSelect v-model="roleDrafts[item.user.id]" label="角色" :options="roleOptions" />
                    <AoiButton appearance="soft" icon="save" @click="updateMemberRole(item)">保存</AoiButton>
                    <AoiButton
                      appearance="soft"
                      :intent="item.membershipStatus === 'active' ? 'danger' : 'neutral'"
                      :icon="item.membershipStatus === 'active' ? 'ban' : 'check'"
                      @click="updateMemberStatus(item, item.membershipStatus === 'active' ? 'disabled' : 'active')"
                    >
                      {{ item.membershipStatus === "active" ? "禁用" : "启用" }}
                    </AoiButton>
                  </div>
                </td>
              </tr>
              <tr v-if="!users.length">
                <td colspan="8" class="muted">暂无匹配用户。</td>
              </tr>
            </tbody>
          </table>
        </div>

        <div class="user-pagination">
          <span>共 {{ pageData.total }} 人</span>
          <div class="action-row">
            <AoiButton appearance="soft" icon="chevron-left" :disabled="page <= 1 || loading" @click="previousPage">上一页</AoiButton>
            <span>{{ page }} / {{ totalPages }}</span>
            <AoiButton appearance="soft" trailing-icon="chevron-right" :disabled="page >= totalPages || loading" @click="nextPage">下一页</AoiButton>
          </div>
        </div>
      </article>

      <aside class="user-side-panel">
        <article class="admin-card">
          <div class="admin-card__header">
            <h2>邀请用户</h2>
          </div>
          <form class="admin-card__body user-invite-form" @submit.prevent="invite">
            <AoiTextField v-model="email" label="邮箱" icon="mail" type="email" placeholder="member@example.com" />
            <AoiSelect v-model="roleCode" label="角色" :options="roleOptions" />
            <AoiButton type="submit" icon="send" :loading="saving" :disabled="!canInvite">发送邀请</AoiButton>
          </form>
        </article>

        <article class="admin-card">
          <div class="admin-card__header">
            <h2>邀请列表</h2>
            <span class="badge">{{ invitations.length }} 个</span>
          </div>
          <div class="data-table-wrap">
            <table class="data-table user-invitation-table">
              <thead>
                <tr>
                  <th>邮箱</th>
                  <th>角色</th>
                  <th>状态</th>
                  <th>操作</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="invitation in invitations" :key="invitation.id">
                  <td>{{ invitation.email }}</td>
                  <td class="mono">{{ invitation.roleCode }}</td>
                  <td><span class="badge" :class="invitation.status === 'pending' ? 'badge--warning' : ''">{{ formatStatus(invitation.status) }}</span></td>
                  <td>
                    <AoiButton appearance="soft" intent="danger" icon="ban" :disabled="invitation.status !== 'pending'" @click="revokeInvitation(invitation)">
                      撤销
                    </AoiButton>
                  </td>
                </tr>
                <tr v-if="!invitations.length">
                  <td colspan="4" class="muted">暂无邀请。</td>
                </tr>
              </tbody>
            </table>
          </div>
        </article>
      </aside>
    </section>
  </div>
</template>

<style scoped>
.user-notice {
  display: flex;
  gap: 10px;
  align-items: flex-start;
  padding: 12px 14px;
  border: 1px solid color-mix(in srgb, var(--aoi-warning) 28%, var(--aoi-border));
  border-radius: var(--aoi-radius-md);
  background: color-mix(in srgb, var(--aoi-warning) 10%, var(--aoi-surface));
  color: var(--aoi-text);
  font-size: 14px;
  line-height: 1.5;
}

.user-notice :deep(.aoi-icon) {
  flex: 0 0 auto;
  color: var(--aoi-warning);
}

.user-workspace {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(300px, 360px);
  gap: 16px;
  align-items: start;
}

.user-side-panel,
.user-invite-form,
.user-card-header > div,
.user-identity,
.user-role-list {
  display: grid;
  gap: 8px;
}

.user-card-header span,
.user-identity span {
  color: var(--aoi-text-muted);
  font-size: 13px;
}

.user-filter-toolbar {
  grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
}

.user-table-wrap {
  overflow-x: auto;
}

.user-table {
  min-width: 1060px;
}

.user-identity strong {
  color: var(--aoi-text);
}

.user-role-list {
  grid-auto-flow: column;
  justify-content: start;
}

.user-actions {
  display: grid;
  grid-template-columns: minmax(150px, 1fr) auto auto;
  gap: 8px;
  align-items: center;
  min-width: 330px;
}

.user-pagination {
  display: flex;
  gap: 12px;
  align-items: center;
  justify-content: space-between;
  padding: 14px 16px;
  border-top: 1px solid var(--aoi-border);
  color: var(--aoi-text-muted);
  font-size: 13px;
}

.user-invitation-table {
  min-width: 100%;
}

@media (max-width: 1180px) {
  .user-workspace {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 760px) {
  .user-filter-toolbar {
    grid-template-columns: 1fr;
  }

  .user-pagination {
    align-items: flex-start;
    flex-direction: column;
  }
}
</style>
