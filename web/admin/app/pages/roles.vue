<script setup lang="ts">
import type { Permission, Role } from "~/types/admin"

type PermissionTarget = "create" | "edit"

type PermissionGroup = {
  code: string
  count: number
  items: Permission[]
  label: string
}

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
const permissionSearch = ref("")
const permissionObjectFilter = ref("")

const permissionObjectLabels: Record<string, string> = {
  audit: "审计",
  config: "配置",
  dictionary: "字典",
  operation: "操作历史",
  org: "组织",
  parameter: "参数",
  permission: "权限",
  plugin: "插件",
  role: "角色",
  session: "会话",
  server: "服务器",
  user: "用户"
}

const permissionObjectOrder = ["org", "user", "role", "config", "server", "dictionary", "operation", "parameter", "permission", "session", "audit", "plugin"]

const permissionGroups = computed<PermissionGroup[]>(() => {
  const groups = new Map<string, PermissionGroup>()

  for (const permission of permissions.value) {
    const groupCode = permissionObjectCode(permission.code)
    const group = groups.get(groupCode) || {
      code: groupCode,
      count: 0,
      items: [],
      label: permissionObjectLabel(groupCode)
    }
    group.items.push(permission)
    group.count = group.items.length
    groups.set(groupCode, group)
  }

  return Array.from(groups.values())
    .map((group) => ({
      ...group,
      items: [...group.items].sort((left, right) => left.code.localeCompare(right.code))
    }))
    .sort((left, right) => permissionObjectRank(left.code) - permissionObjectRank(right.code) || left.code.localeCompare(right.code))
})

const filteredPermissionGroups = computed<PermissionGroup[]>(() => {
  const keyword = permissionSearch.value.trim().toLowerCase()
  return permissionGroups.value
    .filter((group) => !permissionObjectFilter.value || group.code === permissionObjectFilter.value)
    .map((group) => ({
      ...group,
      items: group.items.filter((permission) => matchesPermission(permission, group, keyword))
    }))
    .filter((group) => group.items.length > 0)
})

const permissionGroupOptions = computed(() => [
  { label: "全部权限域", value: "" },
  ...permissionGroups.value.map((group) => ({ label: `${group.label} (${group.count})`, value: group.code }))
])

const filteredPermissionCount = computed(() => filteredPermissionGroups.value.reduce((count, group) => count + group.items.length, 0))

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

function clearPermissions(target: PermissionTarget) {
  if (target === "create") {
    selectedPermissions.value = []
    return
  }
  editPermissions.value = []
}

function setPermissionGroup(target: PermissionTarget, items: Permission[], checked: boolean) {
  const codes = items.map((permission) => permission.code)
  const current = target === "create" ? selectedPermissions.value : editPermissions.value
  const next = checked
    ? Array.from(new Set([...current, ...codes]))
    : current.filter((item) => !codes.includes(item))

  if (target === "create") {
    selectedPermissions.value = next
    return
  }
  editPermissions.value = next
}

function permissionGroupSelectedCount(selected: string[], items: Permission[]) {
  const selectedSet = new Set(selected)
  return items.filter((permission) => selectedSet.has(permission.code)).length
}

function permissionGroupChecked(selected: string[], items: Permission[]) {
  return items.length > 0 && permissionGroupSelectedCount(selected, items) === items.length
}

function permissionObjectCode(code: string) {
  return code.split(":")[0] || "other"
}

function permissionObjectLabel(code: string) {
  return permissionObjectLabels[code] || code.toUpperCase()
}

function permissionObjectRank(code: string) {
  const index = permissionObjectOrder.indexOf(code)
  return index === -1 ? permissionObjectOrder.length + 1 : index
}

function matchesPermission(permission: Permission, group: PermissionGroup, keyword: string) {
  if (!keyword) {
    return true
  }
  return [
    permission.code,
    permission.name,
    permission.description,
    group.code,
    group.label,
    permission.code.includes(":") ? permission.code.split(":")[1] : ""
  ].some((value) => String(value || "").toLowerCase().includes(keyword))
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
        <AoiButton appearance="soft" aria-label="API 管理" icon="code-2" to="/apis">API 管理</AoiButton>
        <AoiButton appearance="soft" icon="refresh-cw" :loading="loading" @click="load">刷新</AoiButton>
      </template>
    </PageHeader>

    <AoiStatusMessage tone="danger" :message="error" />
    <AoiStatusMessage tone="success" :message="success" />

    <section class="admin-management-grid">
      <article class="admin-card admin-management-grid__primary">
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
                <th>权限</th>
                <th>系统角色</th>
                <th>说明</th>
                <th>创建时间</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="role in roles" :key="role.id">
                <td class="mono">{{ role.code }}</td>
                <td>{{ role.name }}</td>
                <td><span class="badge">{{ role.permissions?.length || 0 }} 个</span></td>
                <td><span class="badge" :class="role.system ? 'badge--success' : ''">{{ role.system ? "是" : "否" }}</span></td>
                <td>{{ role.description || "-" }}</td>
                <td>{{ formatDateTime(role.createdAt) }}</td>
              </tr>
              <tr v-if="!roles.length">
                <td colspan="6" class="muted">暂无角色。</td>
              </tr>
            </tbody>
          </table>
        </div>
      </article>

      <article class="admin-card">
        <div class="admin-card__header">
          <h2>编辑自定义角色</h2>
          <span class="badge badge--success">{{ editPermissions.length }} 已选</span>
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

          <div class="permission-panel">
            <div class="permission-panel__toolbar">
              <AoiTextField v-model="permissionSearch" label="权限关键词" icon="search" placeholder="role:update" />
              <AoiSelect
                :model-value="permissionObjectFilter"
                label="权限域"
                :options="permissionGroupOptions"
                @update:model-value="permissionObjectFilter = $event"
              />
            </div>
            <div class="permission-panel__meta">
              <span class="badge">{{ filteredPermissionCount }} 可见</span>
              <AoiButton appearance="plain" intent="neutral" size="sm" icon="x" @click="clearPermissions('edit')">清空</AoiButton>
            </div>
            <div class="permission-groups">
              <section v-for="group in filteredPermissionGroups" :key="group.code" class="permission-group">
                <div class="permission-group__header">
                  <div>
                    <h3>{{ group.label }}</h3>
                    <p>{{ group.code }}</p>
                  </div>
                  <div class="permission-group__actions">
                    <span class="badge">{{ permissionGroupSelectedCount(editPermissions, group.items) }}/{{ group.items.length }}</span>
                    <AoiButton
                      appearance="plain"
                      intent="neutral"
                      size="sm"
                      @click="setPermissionGroup('edit', group.items, !permissionGroupChecked(editPermissions, group.items))"
                    >
                      {{ permissionGroupChecked(editPermissions, group.items) ? "清空" : "全选" }}
                    </AoiButton>
                  </div>
                </div>
                <label
                  v-for="permission in group.items"
                  :key="permission.code"
                  class="permission-option"
                  :class="{ 'permission-option--selected': editPermissions.includes(permission.code) }"
                >
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
              <p v-if="filteredPermissionGroups.length === 0" class="permission-empty muted">暂无匹配权限。</p>
            </div>
          </div>

          <AoiButton type="submit" icon="save" :loading="updating" :disabled="!editRoleId || !editName">保存角色</AoiButton>
        </form>
      </article>

      <article class="admin-card">
        <div class="admin-card__header">
          <h2>创建角色</h2>
          <span class="badge badge--success">{{ selectedPermissions.length }} 已选</span>
        </div>
        <form class="admin-card__body form-grid" @submit.prevent="createRole">
          <AoiTextField v-model="code" label="角色 Code" icon="badge" placeholder="operator" />
          <AoiTextField v-model="name" label="角色名称" icon="id-card" placeholder="Operator" />
          <AoiTextField v-model="description" label="说明" type="textarea" icon="file-text" />

          <div class="permission-panel">
            <div class="permission-panel__toolbar">
              <AoiTextField v-model="permissionSearch" label="权限关键词" icon="search" placeholder="role:update" />
              <AoiSelect
                :model-value="permissionObjectFilter"
                label="权限域"
                :options="permissionGroupOptions"
                @update:model-value="permissionObjectFilter = $event"
              />
            </div>
            <div class="permission-panel__meta">
              <span class="badge">{{ filteredPermissionCount }} 可见</span>
              <AoiButton appearance="plain" intent="neutral" size="sm" icon="x" @click="clearPermissions('create')">清空</AoiButton>
            </div>
            <div class="permission-groups">
              <section v-for="group in filteredPermissionGroups" :key="group.code" class="permission-group">
                <div class="permission-group__header">
                  <div>
                    <h3>{{ group.label }}</h3>
                    <p>{{ group.code }}</p>
                  </div>
                  <div class="permission-group__actions">
                    <span class="badge">{{ permissionGroupSelectedCount(selectedPermissions, group.items) }}/{{ group.items.length }}</span>
                    <AoiButton
                      appearance="plain"
                      intent="neutral"
                      size="sm"
                      @click="setPermissionGroup('create', group.items, !permissionGroupChecked(selectedPermissions, group.items))"
                    >
                      {{ permissionGroupChecked(selectedPermissions, group.items) ? "清空" : "全选" }}
                    </AoiButton>
                  </div>
                </div>
                <label
                  v-for="permission in group.items"
                  :key="permission.code"
                  class="permission-option"
                  :class="{ 'permission-option--selected': selectedPermissions.includes(permission.code) }"
                >
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
              <p v-if="filteredPermissionGroups.length === 0" class="permission-empty muted">暂无匹配权限。</p>
            </div>
          </div>

          <AoiButton type="submit" icon="plus" :loading="saving" :disabled="!code || !name">创建角色</AoiButton>
        </form>
      </article>
    </section>
  </div>
</template>

<style scoped>
.permission-panel {
  border: 1px solid var(--aoi-border);
  border-radius: var(--aoi-radius-card);
  background: var(--aoi-surface-solid);
  overflow: hidden;
}

.permission-panel__toolbar {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
  gap: 10px;
  padding: 10px;
}

.permission-panel__toolbar :deep(.aoi-text-field) {
  min-width: 0;
}

.permission-panel__meta {
  align-items: center;
  border-top: 1px solid var(--aoi-border);
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  justify-content: space-between;
  padding: 8px 10px;
}

.permission-groups {
  display: grid;
  border-top: 1px solid var(--aoi-border);
  max-height: 360px;
  overflow: auto;
}

.permission-group {
  display: grid;
  border-bottom: 1px solid var(--aoi-border);
  gap: 2px;
  padding: 10px;
}

.permission-group:last-child {
  border-bottom: 0;
}

.permission-group__header {
  align-items: center;
  display: flex;
  gap: 10px;
  justify-content: space-between;
  margin-bottom: 4px;
}

.permission-group__header h3,
.permission-group__header p {
  margin: 0;
}

.permission-group__header h3 {
  color: var(--aoi-text);
  font-size: 13px;
}

.permission-group__header p {
  color: var(--aoi-text-muted);
  font-size: 12px;
  margin-top: 2px;
}

.permission-group__actions {
  align-items: center;
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  justify-content: flex-end;
}

.permission-option {
  align-items: flex-start;
  background: transparent;
  border-radius: var(--aoi-radius-control);
  display: grid;
  gap: 8px;
  grid-template-columns: 18px minmax(0, 1fr);
  padding: 8px 6px;
}

.permission-option:hover,
.permission-option--selected {
  background: var(--aoi-surface-muted);
}

.permission-option input {
  margin-top: 2px;
}

.permission-option small,
.permission-option strong {
  display: block;
  min-width: 0;
}

.permission-option small {
  color: var(--aoi-text-muted);
  margin-top: 2px;
  overflow-wrap: anywhere;
}

.permission-option strong {
  overflow-wrap: anywhere;
}

.permission-empty {
  margin: 0;
  padding: 14px 10px;
}

@media (max-width: 640px) {
  .permission-panel__toolbar {
    grid-template-columns: 1fr;
  }

  .permission-group__header {
    align-items: flex-start;
    flex-direction: column;
  }

  .permission-group__actions {
    justify-content: flex-start;
  }
}
</style>
