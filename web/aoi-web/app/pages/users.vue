<script setup lang="ts">
import type { Invitation, OrganizationUser, Role } from "~/types/admin"

const api = useAdminApi()
const auth = useAuthStore()
const users = ref<OrganizationUser[]>([])
const roles = ref<Role[]>([])
const invitations = ref<Invitation[]>([])
const email = ref("")
const roleCode = ref("member")
const inviteToken = ref("")
const inviteUrl = ref("")
const roleDrafts = ref<Record<string, string>>({})
const loading = ref(false)
const saving = ref(false)
const error = ref("")
const success = ref("")

const roleOptions = computed(() => roles.value.map((role) => ({
  label: `${role.name} (${role.code})`,
  value: role.code
})))

async function load() {
  if (!auth.currentOrgId) {
    return
  }

  loading.value = true
  error.value = ""
  try {
    const [userResult, roleResult] = await Promise.all([
      api.listUsers(auth.currentOrgId),
      api.listRoles(auth.currentOrgId)
    ])
    const invitationResult = await api.listInvitations(auth.currentOrgId)
    users.value = userResult
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

async function invite() {
  if (!auth.currentOrgId || !email.value.trim() || !roleCode.value) {
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

onMounted(load)
watch(() => auth.currentOrgId, load)

useHead({
  title: "用户 - Aoi Admin"
})
</script>

<template>
  <div class="page-grid">
    <PageHeader title="用户" icon="users" description="查看当前组织成员，并通过 no-op 邀请流程生成邀请 token。">
      <template #actions>
        <AoiButton appearance="soft" icon="refresh-cw" :loading="loading" @click="load">刷新</AoiButton>
      </template>
    </PageHeader>

    <AoiStatusMessage tone="danger" :message="error" />
    <AoiStatusMessage tone="success" :message="success" />
    <AoiStatusMessage v-if="inviteToken" tone="success" :message="`调试 token：${inviteToken}`" />
    <AoiStatusMessage v-if="inviteUrl" tone="success" :message="`调试链接：${inviteUrl}`" />

    <section class="two-column-grid">
      <article class="admin-card">
        <div class="admin-card__header">
          <h2>组织用户</h2>
          <span class="badge">{{ users.length }} 人</span>
        </div>
        <div class="data-table-wrap">
          <table class="data-table">
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
                <td>{{ item.user.id }}</td>
                <td>{{ item.user.displayName || item.user.username }}</td>
                <td>{{ item.user.email }}</td>
                <td>
                  <span v-for="role in item.roles" :key="role" class="badge">{{ role.replace('role:', '') }}</span>
                </td>
                <td><span class="badge" :class="item.membershipStatus === 'active' ? 'badge--success' : 'badge--warning'">{{ formatStatus(item.membershipStatus) }}</span></td>
                <td><span class="badge" :class="item.user.mfaEnabled ? 'badge--success' : ''">{{ item.user.mfaEnabled ? "已启用" : "未启用" }}</span></td>
                <td>{{ formatDateTime(item.user.lastLoginAt) }}</td>
                <td>
                  <div class="toolbar-row">
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
                <td colspan="8" class="muted">暂无用户。</td>
              </tr>
            </tbody>
          </table>
        </div>
      </article>

      <article class="admin-card">
        <div class="admin-card__header">
          <h2>邀请用户</h2>
        </div>
        <form class="admin-card__body form-grid" @submit.prevent="invite">
          <AoiTextField v-model="email" label="邮箱" icon="mail" type="email" placeholder="member@example.com" />
          <AoiSelect v-model="roleCode" label="角色" :options="roleOptions" />
          <AoiButton type="submit" icon="send" :loading="saving" :disabled="!email || !roleCode">发送邀请</AoiButton>
        </form>
      </article>

      <article class="admin-card">
        <div class="admin-card__header">
          <h2>邀请列表</h2>
          <span class="badge">{{ invitations.length }} 个</span>
        </div>
        <div class="data-table-wrap">
          <table class="data-table">
            <thead>
              <tr>
                <th>邮箱</th>
                <th>角色</th>
                <th>状态</th>
                <th>过期时间</th>
                <th>操作</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="invitation in invitations" :key="invitation.id">
                <td>{{ invitation.email }}</td>
                <td class="mono">{{ invitation.roleCode }}</td>
                <td><span class="badge" :class="invitation.status === 'pending' ? 'badge--warning' : ''">{{ formatStatus(invitation.status) }}</span></td>
                <td>{{ formatDateTime(invitation.expiresAt) }}</td>
                <td>
                  <AoiButton appearance="soft" intent="danger" icon="ban" :disabled="invitation.status !== 'pending'" @click="revokeInvitation(invitation)">
                    撤销
                  </AoiButton>
                </td>
              </tr>
              <tr v-if="!invitations.length">
                <td colspan="5" class="muted">暂无邀请。</td>
              </tr>
            </tbody>
          </table>
        </div>
      </article>
    </section>
  </div>
</template>




