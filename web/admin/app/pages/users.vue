<script setup lang="ts">
import type { OrganizationUser, Role } from "~/types/api"

const api = useAdminApi()
const auth = useAuthStore()
const users = ref<OrganizationUser[]>([])
const roles = ref<Role[]>([])
const email = ref("")
const roleCode = ref("member")
const inviteToken = ref("")
const loading = ref(false)
const saving = ref(false)
const error = ref("")

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
    users.value = userResult
    roles.value = roleResult
    if (!roles.value.some((role) => role.code === roleCode.value)) {
      roleCode.value = roles.value[0]?.code || "member"
    }
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
  inviteToken.value = ""
  try {
    const result = await api.inviteUser(auth.currentOrgId, {
      email: email.value.trim(),
      roleCode: roleCode.value
    })
    inviteToken.value = result.token
    email.value = ""
  } catch (err) {
    error.value = errorMessage(err)
  } finally {
    saving.value = false
  }
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
    <AoiStatusMessage v-if="inviteToken" tone="success" :message="`邀请 token：${inviteToken}`" />

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
                <th>MFA</th>
                <th>最近登录</th>
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
                <td><span class="badge" :class="item.user.mfaEnabled ? 'badge--success' : ''">{{ item.user.mfaEnabled ? "已启用" : "未启用" }}</span></td>
                <td>{{ formatDateTime(item.user.lastLoginAt) }}</td>
              </tr>
              <tr v-if="!users.length">
                <td colspan="6" class="muted">暂无用户。</td>
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
    </section>
  </div>
</template>
