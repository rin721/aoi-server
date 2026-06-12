<script setup lang="ts">
import type { Organization, OrganizationPage } from "~/types/admin"

const api = useAdminApi()
const auth = useAuthStore()

const emptyPage: OrganizationPage = { items: [], page: 1, pageSize: 10, storageStatus: "unavailable", total: 0 }
const pageData = ref<OrganizationPage>({ ...emptyPage })
const filters = reactive({
  keyword: "",
  code: "",
  name: "",
  status: ""
})
const code = ref("")
const name = ref("")
const currentOrgName = ref("")
const page = ref(1)
const pageSize = ref("10")
const loading = ref(false)
const saving = ref(false)
const updating = ref(false)
const error = ref("")
const success = ref("")

const organizations = computed(() => pageData.value.items)
const persisted = computed(() => pageData.value.storageStatus === "persisted")
const pageSizeNumber = computed(() => Math.min(100, Math.max(1, Number(pageSize.value) || 10)))
const totalPages = computed(() => Math.max(1, Math.ceil(pageData.value.total / pageSizeNumber.value)))
const canCreate = computed(() => Boolean(code.value.trim() && name.value.trim() && !saving.value))
const canUpdate = computed(() => Boolean(auth.currentOrgId && currentOrgName.value.trim() && !updating.value))
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

async function load() {
  loading.value = true
  error.value = ""
  try {
    pageData.value = await api.listOrganizations({
      keyword: filters.keyword.trim() || undefined,
      code: filters.code.trim() || undefined,
      name: filters.name.trim() || undefined,
      status: filters.status || undefined,
      orderKey: "id",
      desc: true,
      page: page.value,
      pageSize: pageSizeNumber.value
    })
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
  filters.code = ""
  filters.name = ""
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

async function createOrg() {
  if (!canCreate.value) {
    return
  }

  saving.value = true
  error.value = ""
  success.value = ""
  try {
    const org = await api.createOrganization({
      code: code.value.trim(),
      name: name.value.trim()
    })
    success.value = `组织 ${org.name} 已创建。`
    code.value = ""
    name.value = ""
    page.value = 1
    await Promise.all([load(), auth.fetchSession()])
  } catch (err) {
    error.value = errorMessage(err)
  } finally {
    saving.value = false
  }
}

async function updateCurrentOrg() {
  if (!canUpdate.value || !auth.currentOrgId) {
    return
  }

  updating.value = true
  error.value = ""
  success.value = ""
  try {
    const org = await api.updateOrganization(auth.currentOrgId, { name: currentOrgName.value.trim() })
    success.value = `组织 ${org.code} 已更新。`
    await Promise.all([load(), auth.fetchSession()])
  } catch (err) {
    error.value = errorMessage(err)
  } finally {
    updating.value = false
  }
}

async function switchOrganization(org: Organization) {
  if (org.id === auth.currentOrgId || loading.value) {
    return
  }
  error.value = ""
  success.value = ""
  try {
    await auth.switchOrg(org.id)
    success.value = `已切换到 ${org.name}。`
    await load()
  } catch (err) {
    error.value = errorMessage(err)
  }
}

function isCurrentOrg(org: Organization) {
  return org.id === auth.currentOrgId
}

onMounted(load)
watch(() => auth.currentOrg?.name, (value) => {
  currentOrgName.value = value || ""
}, { immediate: true })
watch(() => auth.currentOrgId, () => {
  page.value = 1
  load()
})

useHead({
  title: "组织管理 - Aoi Admin"
})
</script>

<template>
  <div class="page-grid org-page">
    <PageHeader title="组织管理" icon="building-2" description="查看组织、切换访问上下文，并维护当前组织信息。">
      <template #actions>
        <AoiButton appearance="soft" icon="refresh-cw" :loading="loading" @click="load">刷新</AoiButton>
      </template>
    </PageHeader>

    <AoiStatusMessage tone="danger" :message="error" />
    <AoiStatusMessage tone="success" :message="success" />

    <section class="org-notice">
      <AoiIcon name="info" decorative />
      <span>注：组织切换会重新签发访问上下文；组织名称只能维护当前登录令牌绑定的组织。</span>
    </section>

    <section class="org-workspace">
      <article class="admin-card org-main-card">
        <div class="admin-card__header org-card-header">
          <div>
            <h2>组织列表</h2>
            <span>{{ persisted ? "已连接组织数据" : "等待组织数据可用" }}</span>
          </div>
          <span class="badge">{{ pageData.total }} 个</span>
        </div>

        <div class="admin-filter-toolbar org-filter-toolbar">
          <AoiTextField v-model="filters.keyword" label="关键字" icon="search" @enter="search" />
          <AoiTextField v-model="filters.code" label="组织 Code" icon="badge" @enter="search" />
          <AoiTextField v-model="filters.name" label="组织名称" icon="building" @enter="search" />
          <AoiSelect v-model="filters.status" label="状态" :options="statusOptions" />
          <AoiSelect v-model="pageSize" label="每页" :options="pageSizeOptions" />
          <AoiButton icon="search" :loading="loading" @click="search">查询</AoiButton>
          <AoiButton appearance="soft" icon="x" @click="resetFilters">重置</AoiButton>
        </div>

        <div class="data-table-wrap org-table-wrap">
          <table class="data-table org-table">
            <thead>
              <tr>
                <th>Code</th>
                <th>名称</th>
                <th>状态</th>
                <th>当前</th>
                <th>创建时间</th>
                <th>操作</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="org in organizations" :key="org.id">
                <td class="mono">{{ org.code }}</td>
                <td>
                  <div class="org-identity">
                    <strong>{{ org.name }}</strong>
                    <span>{{ org.code }}</span>
                  </div>
                </td>
                <td><span class="badge" :class="org.status === 'active' ? 'badge--success' : 'badge--warning'">{{ formatStatus(org.status) }}</span></td>
                <td><span class="badge" :class="isCurrentOrg(org) ? 'badge--success' : ''">{{ isCurrentOrg(org) ? "当前组织" : "可切换" }}</span></td>
                <td>{{ formatDateTime(org.createdAt) }}</td>
                <td>
                  <AoiButton appearance="soft" icon="repeat" :disabled="isCurrentOrg(org) || loading" @click="switchOrganization(org)">
                    切换
                  </AoiButton>
                </td>
              </tr>
              <tr v-if="!organizations.length">
                <td colspan="6" class="muted">暂无匹配组织。</td>
              </tr>
            </tbody>
          </table>
        </div>

        <div class="org-pagination">
          <span>共 {{ pageData.total }} 个组织</span>
          <div class="action-row">
            <AoiButton appearance="soft" icon="chevron-left" :disabled="page <= 1 || loading" @click="previousPage">上一页</AoiButton>
            <span>{{ page }} / {{ totalPages }}</span>
            <AoiButton appearance="soft" trailing-icon="chevron-right" :disabled="page >= totalPages || loading" @click="nextPage">下一页</AoiButton>
          </div>
        </div>
      </article>

      <aside class="org-side-panel">
        <article class="admin-card">
          <div class="admin-card__header">
            <h2>当前组织</h2>
          </div>
          <form class="admin-card__body form-grid" @submit.prevent="updateCurrentOrg">
            <AoiTextField :model-value="auth.currentOrg?.code || ''" label="组织 Code" icon="badge" disabled />
            <AoiTextField v-model="currentOrgName" label="组织名称" icon="building" placeholder="Acme Corp" />
            <AoiButton type="submit" icon="save" :loading="updating" :disabled="!canUpdate">保存组织</AoiButton>
          </form>
        </article>

        <article class="admin-card">
          <div class="admin-card__header">
            <h2>创建组织</h2>
          </div>
          <form class="admin-card__body form-grid" @submit.prevent="createOrg">
            <AoiTextField v-model="code" label="组织 Code" icon="badge" placeholder="acme" />
            <AoiTextField v-model="name" label="组织名称" icon="building" placeholder="Acme Corp" />
            <AoiButton type="submit" icon="plus" :loading="saving" :disabled="!canCreate">创建组织</AoiButton>
          </form>
        </article>
      </aside>
    </section>
  </div>
</template>

<style scoped>
.org-notice {
  display: flex;
  gap: 10px;
  align-items: flex-start;
  padding: 12px 14px;
  border: 1px solid color-mix(in srgb, var(--aoi-info) 28%, var(--aoi-border));
  border-radius: var(--aoi-radius-md);
  background: color-mix(in srgb, var(--aoi-info) 9%, var(--aoi-surface));
  color: var(--aoi-text);
  font-size: 14px;
  line-height: 1.5;
}

.org-notice :deep(.aoi-icon) {
  flex: 0 0 auto;
  color: var(--aoi-info);
}

.org-workspace {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(300px, 360px);
  gap: 16px;
  align-items: start;
}

.org-side-panel,
.org-card-header > div,
.org-identity {
  display: grid;
  gap: 8px;
}

.org-card-header span,
.org-identity span {
  color: var(--aoi-text-muted);
  font-size: 13px;
}

.org-filter-toolbar {
  grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
}

.org-table-wrap {
  overflow-x: auto;
}

.org-table {
  min-width: 720px;
}

.org-identity strong {
  color: var(--aoi-text);
}

.org-pagination {
  display: flex;
  gap: 12px;
  align-items: center;
  justify-content: space-between;
  padding: 14px 16px;
  border-top: 1px solid var(--aoi-border);
  color: var(--aoi-text-muted);
  font-size: 13px;
}

@media (max-width: 1180px) {
  .org-workspace {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 760px) {
  .org-filter-toolbar {
    grid-template-columns: 1fr;
  }

  .org-pagination {
    align-items: flex-start;
    flex-direction: column;
  }
}
</style>
