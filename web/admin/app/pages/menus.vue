<script setup lang="ts">
import type { SystemMenuGroup, SystemMenuItem } from "~/types/admin"

const api = useAdminApi()
const groups = ref<SystemMenuGroup[]>([])
const query = ref("")
const groupCode = ref("")
const loading = ref(false)
const error = ref("")

const totalCount = computed(() => groups.value.reduce((count, group) => count + group.items.length, 0))
const protectedCount = computed(() => groups.value.reduce((count, group) => count + group.items.filter((item) => Boolean(item.permission)).length, 0))
const mobileCount = computed(() => groups.value.reduce((count, group) => count + group.items.filter((item) => item.mobile).length, 0))

const groupOptions = computed(() => [
  { label: "全部分组", value: "" },
  ...groups.value.map((group) => ({ label: `${group.label} (${group.items.length})`, value: group.code }))
])

const filteredGroups = computed(() => {
  const keyword = query.value.trim().toLowerCase()
  return groups.value
    .filter((group) => !groupCode.value || group.code === groupCode.value)
    .map((group) => ({
      ...group,
      items: group.items.filter((item) => matchesMenu(group, item, keyword))
    }))
    .filter((group) => group.items.length > 0)
})

async function load() {
  loading.value = true
  error.value = ""
  try {
    groups.value = await api.listSystemMenus()
  } catch (err) {
    error.value = errorMessage(err)
  } finally {
    loading.value = false
  }
}

function matchesMenu(group: SystemMenuGroup, item: SystemMenuItem, keyword: string) {
  if (!keyword) {
    return true
  }
  return [
    group.code,
    group.label,
    item.code,
    item.icon,
    item.label,
    item.path,
    item.permission || "",
    item.mobile ? "mobile 移动端" : "desktop 桌面端"
  ].some((value) => value.toLowerCase().includes(keyword))
}

onMounted(load)

useHead({
  title: "菜单管理 - Aoi Admin"
})
</script>

<template>
  <div class="page-grid">
    <PageHeader title="菜单管理" icon="panel-left" description="查看当前用户可见的后台菜单目录，核对路由、权限码、移动端入口和排序。">
      <template #actions>
        <AoiButton appearance="soft" aria-label="API 管理" icon="code-2" to="/apis">API 管理</AoiButton>
        <AoiButton appearance="soft" icon="refresh-cw" :loading="loading" @click="load">刷新</AoiButton>
      </template>
    </PageHeader>

    <AoiStatusMessage tone="danger" :message="error" />

    <article class="admin-card">
      <div class="admin-card__header">
        <h2>菜单目录</h2>
        <div class="menu-summary">
          <span class="badge">{{ groups.length }} 组</span>
          <span class="badge">{{ totalCount }} 项</span>
          <span class="badge badge--success">{{ protectedCount }} 权限控制</span>
          <span class="badge">{{ mobileCount }} 移动入口</span>
        </div>
      </div>

      <div class="admin-filter-toolbar">
        <AoiTextField v-model="query" label="关键词" icon="search" placeholder="/roles 或 role:read" />
        <AoiSelect
          :model-value="groupCode"
          label="分组"
          :options="groupOptions"
          @update:model-value="groupCode = $event"
        />
      </div>

      <div class="menu-groups">
        <section v-for="group in filteredGroups" :key="group.code" class="menu-group">
          <div class="menu-group__header">
            <div>
              <h3>{{ group.label }}</h3>
              <p>{{ group.code }} · order {{ group.order }}</p>
            </div>
            <span class="badge">{{ group.items.length }} 项</span>
          </div>

          <div class="data-table-wrap">
            <table class="data-table menu-table">
              <thead>
                <tr>
                  <th>菜单</th>
                  <th>Path</th>
                  <th>权限</th>
                  <th>入口</th>
                  <th>图标</th>
                  <th>排序</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="item in group.items" :key="item.code">
                  <td data-label="菜单">
                    <div class="menu-name">
                      <AoiIcon :name="item.icon" decorative />
                      <div>
                        <strong>{{ item.label }}</strong>
                        <small>{{ item.code }}</small>
                      </div>
                    </div>
                  </td>
                  <td class="mono menu-table__path" data-label="Path">{{ item.path }}</td>
                  <td data-label="权限">
                    <span v-if="item.permission" class="badge">{{ item.permission }}</span>
                    <span v-else class="muted">登录可见</span>
                  </td>
                  <td data-label="入口">
                    <span class="badge" :class="item.mobile ? 'badge--success' : ''">
                      {{ item.mobile ? "移动 + 桌面" : "桌面" }}
                    </span>
                  </td>
                  <td class="mono" data-label="图标">{{ item.icon }}</td>
                  <td class="mono" data-label="排序">{{ item.order }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </section>

        <p v-if="!loading && filteredGroups.length === 0" class="menu-empty muted">暂无匹配菜单。</p>
      </div>
    </article>
  </div>
</template>

<style scoped>
.menu-summary {
  align-items: center;
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  justify-content: flex-end;
}

.menu-groups {
  display: grid;
  gap: 16px;
  padding: 0 14px 14px;
}

.menu-group {
  display: grid;
  gap: 10px;
}

.menu-group__header {
  align-items: center;
  display: flex;
  gap: 12px;
  justify-content: space-between;
}

.menu-group__header h3,
.menu-group__header p {
  margin: 0;
}

.menu-group__header h3 {
  font-size: 15px;
}

.menu-group__header p {
  color: var(--aoi-text-muted);
  font-size: 12px;
  margin-top: 3px;
}

.menu-name {
  align-items: center;
  display: grid;
  gap: 10px;
  grid-template-columns: 22px minmax(0, 1fr);
}

.menu-name strong,
.menu-name small {
  display: block;
  min-width: 0;
}

.menu-name small {
  color: var(--aoi-text-muted);
  font-size: 12px;
  margin-top: 2px;
}

.menu-table__path {
  min-width: 180px;
}

.menu-empty {
  margin: 0;
  padding: 18px 0 4px;
}

@media (max-width: 640px) {
  .admin-card__header {
    align-items: flex-start;
    flex-direction: column;
    gap: 10px;
  }

  .admin-card__header h2,
  .menu-summary {
    width: 100%;
  }

  .menu-summary {
    justify-content: flex-start;
  }

  .menu-groups {
    padding-inline: 14px;
  }

  .menu-table,
  .menu-table tbody,
  .menu-table tr,
  .menu-table td {
    display: block;
  }

  .menu-table thead {
    display: none;
  }

  .menu-table tr {
    border-bottom: 1px solid var(--aoi-border);
    padding: 10px 0;
  }

  .menu-table tr:last-child {
    border-bottom: 0;
  }

  .menu-table td {
    align-items: flex-start;
    border-bottom: 0;
    display: grid;
    gap: 8px;
    grid-template-columns: 72px minmax(0, 1fr);
    padding: 5px 0;
    white-space: normal;
  }

  .menu-table td::before {
    color: var(--aoi-text-muted);
    content: attr(data-label);
    font-size: 12px;
    font-weight: 700;
  }

  .menu-table td .badge {
    justify-self: start;
  }

  .menu-table__path {
    min-width: 0;
    overflow-wrap: anywhere;
  }
}
</style>
