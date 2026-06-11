<script setup lang="ts">
import type { Permission, Role } from "~/types/admin"

const api = useAdminApi()
const auth = useAuthStore()
const roles = ref<Role[]>([])
const permissions = ref<Permission[]>([])
const code = ref("")
const name = ref("")
const description = ref("")
const selectedPermissions = ref<string[]>([])
const editRoleId = ref("")
const editName = ref("")
const editDescription = ref("")
const editPermissions = ref<string[]>([])
const loading = ref(false)
const saving = ref(false)
const updating = ref(false)
const error = ref("")
const success = ref("")

const groupedPermissions = computed(() => {
  const groups = new Map<string, Permission[]>()

  for (const permission of permissions.value) {
    const group = permission.code.split(":")[0] || "other"
    groups.set(group, [...(groups.get(group) || []), permission])
  }

  return Array.from(groups.entries())
})

const editableRoleOptions = computed(() => roles.value
  .filter((role) => !role.system)
  .map((role) => ({ label: `${role.name} (${role.code})`, value: role.id })))

const selectedEditRole = computed(() => roles.value.find((role) => role.id === editRoleId.value) || null)

async function load() {
  if (!auth.currentOrgId) {
    return
  }

  loading.value = true
  error.value = ""
  try {
    const [roleResult, permissionResult] = await Promise.all([
      api.listRoles(auth.currentOrgId),
      api.listPermissions(auth.currentOrgId)
    ])
    roles.value = roleResult
    permissions.value = permissionResult
  } catch (err) {
    error.value = errorMessage(err)
  } finally {
    loading.value = false
  }
}

function togglePermission(code: string, checked: boolean) {
  selectedPermissions.value = checked
    ? Array.from(new Set([...selectedPermissions.value, code]))
    : selectedPermissions.value.filter((item) => item !== code)
}

function toggleEditPermission(code: string, checked: boolean) {
  editPermissions.value = checked
    ? Array.from(new Set([...editPermissions.value, code]))
    : editPermissions.value.filter((item) => item !== code)
}

async function createRole() {
  if (!auth.currentOrgId || !code.value.trim() || !name.value.trim()) {
    return
  }

  saving.value = true
  error.value = ""
  success.value = ""
  try {
    const role = await api.createRole(auth.currentOrgId, {
      code: code.value.trim(),
      description: description.value.trim(),
      name: name.value.trim(),
      permissions: selectedPermissions.value
    })
    success.value = `角色 ${role.name} 已创建。`
    code.value = ""
    name.value = ""
    description.value = ""
    selectedPermissions.value = []
    await load()
  } catch (err) {
    error.value = errorMessage(err)
  } finally {
    saving.value = false
  }
}

async function updateRole() {
  if (!auth.currentOrgId || !editRoleId.value || !editName.value.trim()) {
    return
  }

  updating.value = true
  error.value = ""
  success.value = ""
  try {
    const role = await api.updateRole(auth.currentOrgId, editRoleId.value, {
      description: editDescription.value.trim(),
      name: editName.value.trim(),
      permissions: editPermissions.value
    })
    success.value = `角色 ${role.name} 已更新。`
    await load()
  } catch (err) {
    error.value = errorMessage(err)
  } finally {
    updating.value = false
  }
}

onMounted(load)
watch(() => auth.currentOrgId, load)
watch(selectedEditRole, (role) => {
  editName.value = role?.name || ""
  editDescription.value = role?.description || ""
  editPermissions.value = [...(role?.permissions || [])]
}, { immediate: true })

useHead({
  title: "角色权限 - Aoi Admin"
})
</script>

<template>
  <div class="page-grid">
    <PageHeader title="角色权限" icon="shield-check" description="查看系统角色与权限集合，并为当前组织创建自定义角色。">
      <template #actions>
        <AoiButton appearance="soft" icon="refresh-cw" :loading="loading" @click="load">刷新</AoiButton>
      </template>
    </PageHeader>

    <AoiStatusMessage tone="danger" :message="error" />
    <AoiStatusMessage tone="success" :message="success" />

    <section class="two-column-grid">
      <article class="admin-card">
        <div class="admin-card__header">
          <h2>角色列表</h2>
          <span class="badge">{{ roles.length }} 个</span>
        </div>
        <div class="data-table-wrap">
          <table class="data-table">
            <thead>
              <tr>
                <th>Code</th>
                <th>名称</th>
                <th>系统角色</th>
                <th>说明</th>
                <th>创建时间</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="role in roles" :key="role.id">
                <td class="mono">{{ role.code }}</td>
                <td>{{ role.name }}</td>
                <td><span class="badge" :class="role.system ? 'badge--success' : ''">{{ role.system ? "是" : "否" }}</span></td>
                <td>{{ role.description || "-" }}</td>
                <td>{{ formatDateTime(role.createdAt) }}</td>
              </tr>
              <tr v-if="!roles.length">
                <td colspan="5" class="muted">暂无角色。</td>
              </tr>
            </tbody>
          </table>
        </div>
      </article>

      <article class="admin-card">
        <div class="admin-card__header">
          <h2>编辑自定义角色</h2>
        </div>
        <form class="admin-card__body form-grid" @submit.prevent="updateRole">
          <AoiSelect
            :model-value="editRoleId"
            label="选择角色"
            :options="editableRoleOptions"
            @update:model-value="editRoleId = $event"
          />
          <AoiTextField v-model="editName" label="角色名称" icon="id-card" />
          <AoiTextField v-model="editDescription" label="说明" type="textarea" icon="file-text" />

          <div class="permission-groups">
            <section v-for="[group, items] in groupedPermissions" :key="group" class="permission-group">
              <h3>{{ group }}</h3>
              <label v-for="permission in items" :key="permission.code" class="permission-option">
                <input
                  type="checkbox"
                  :checked="editPermissions.includes(permission.code)"
                  @change="toggleEditPermission(permission.code, ($event.target as HTMLInputElement).checked)"
                >
                <span>
                  <strong>{{ permission.code }}</strong>
                  <small>{{ permission.description || permission.name }}</small>
                </span>
              </label>
            </section>
          </div>

          <AoiButton type="submit" icon="save" :loading="updating" :disabled="!editRoleId || !editName">保存角色</AoiButton>
        </form>
      </article>

      <article class="admin-card">
        <div class="admin-card__header">
          <h2>创建角色</h2>
        </div>
        <form class="admin-card__body form-grid" @submit.prevent="createRole">
          <AoiTextField v-model="code" label="角色 Code" icon="badge" placeholder="operator" />
          <AoiTextField v-model="name" label="角色名称" icon="id-card" placeholder="Operator" />
          <AoiTextField v-model="description" label="说明" type="textarea" icon="file-text" />

          <div class="permission-groups">
            <section v-for="[group, items] in groupedPermissions" :key="group" class="permission-group">
              <h3>{{ group }}</h3>
              <label v-for="permission in items" :key="permission.code" class="permission-option">
                <input
                  type="checkbox"
                  :checked="selectedPermissions.includes(permission.code)"
                  @change="togglePermission(permission.code, ($event.target as HTMLInputElement).checked)"
                >
                <span>
                  <strong>{{ permission.code }}</strong>
                  <small>{{ permission.description || permission.name }}</small>
                </span>
              </label>
            </section>
          </div>

          <AoiButton type="submit" icon="plus" :loading="saving" :disabled="!code || !name">创建角色</AoiButton>
        </form>
      </article>
    </section>
  </div>
</template>

<style scoped>
.permission-groups {
  display: grid;
  gap: 10px;
  max-height: 360px;
  overflow: auto;
  padding-right: 2px;
}

.permission-group {
  display: grid;
  gap: 6px;
}

.permission-group h3 {
  margin: 0;
  color: var(--aoi-text-muted);
  font-size: 12px;
  text-transform: uppercase;
}

.permission-option {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  border: 1px solid var(--aoi-border);
  border-radius: var(--aoi-radius-control);
  background: rgba(255, 255, 255, .65);
  padding: 9px;
}

.permission-option small,
.permission-option strong {
  display: block;
}

.permission-option small {
  color: var(--aoi-text-muted);
  margin-top: 2px;
}
</style>




