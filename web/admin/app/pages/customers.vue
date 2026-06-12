<script setup lang="ts">
import type { DemoCustomer, DemoCustomerPage } from "~/types/admin"

const api = useAdminApi()

const pageData = ref<DemoCustomerPage>({ items: [], page: 1, pageSize: 10, storageStatus: "unavailable", total: 0 })
const keyword = ref("")
const page = ref(1)
const pageSize = ref("10")
const loading = ref(false)
const saving = ref(false)
const deleting = ref(false)
const drawerOpen = ref(false)
const editing = ref<DemoCustomer | null>(null)
const form = reactive({
  customerName: "",
  customerPhoneData: ""
})
const error = ref("")
const success = ref("")

const pageSizeNumber = computed(() => Math.min(100, Math.max(1, Number(pageSize.value) || 10)))
const totalPages = computed(() => Math.max(1, Math.ceil(pageData.value.total / pageSizeNumber.value)))
const drawerTitle = computed(() => editing.value ? "编辑客户" : "新增客户")
const canSave = computed(() => Boolean(form.customerName.trim() && form.customerPhoneData.trim() && !saving.value))
const persisted = computed(() => pageData.value.storageStatus === "persisted")

async function load() {
  loading.value = true
  error.value = ""
  try {
    pageData.value = await api.listDemoCustomers({
      keyword: keyword.value.trim() || undefined,
      page: page.value,
      pageSize: pageSizeNumber.value
    })
  } catch (err) {
    error.value = errorMessage(err)
  } finally {
    loading.value = false
  }
}

function openCreateDrawer() {
  editing.value = null
  form.customerName = ""
  form.customerPhoneData = ""
  error.value = ""
  success.value = ""
  drawerOpen.value = true
}

function openEditDrawer(item: DemoCustomer) {
  editing.value = item
  form.customerName = item.customerName
  form.customerPhoneData = item.customerPhoneData
  error.value = ""
  success.value = ""
  drawerOpen.value = true
}

function closeDrawer() {
  if (!saving.value) {
    drawerOpen.value = false
  }
}

async function submitForm() {
  if (!canSave.value) {
    return
  }
  saving.value = true
  error.value = ""
  success.value = ""
  try {
    if (editing.value) {
      await api.updateDemoCustomer(editing.value.id, {
        customerName: form.customerName.trim(),
        customerPhoneData: form.customerPhoneData.trim()
      })
      success.value = "客户信息已更新。"
    } else {
      await api.createDemoCustomer({
        customerName: form.customerName.trim(),
        customerPhoneData: form.customerPhoneData.trim()
      })
      success.value = "客户已创建。"
    }
    drawerOpen.value = false
    await load()
  } catch (err) {
    error.value = errorMessage(err)
  } finally {
    saving.value = false
  }
}

async function deleteCustomer(item: DemoCustomer) {
  if (!confirm(`删除客户「${item.customerName}」？删除后列表中将不可见。`)) {
    return
  }
  deleting.value = true
  error.value = ""
  success.value = ""
  try {
    await api.deleteDemoCustomer(item.id)
    success.value = `客户「${item.customerName}」已删除。`
    await load()
  } catch (err) {
    error.value = errorMessage(err)
  } finally {
    deleting.value = false
  }
}

async function search() {
  page.value = 1
  await load()
}

async function resetFilters() {
  keyword.value = ""
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

function ownerLabel(item: DemoCustomer) {
  return item.ownerUsername || `#${item.ownerUserId}`
}

onMounted(load)

useHead({
  title: "客户列表 - Aoi Admin"
})
</script>

<template>
  <div class="page-grid customer-page">
    <PageHeader title="客户列表" icon="id-card" description="演示登录主体、资源归属与可见范围。">
      <template #actions>
        <AoiButton appearance="soft" icon="refresh-cw" :loading="loading" @click="load">刷新</AoiButton>
        <AoiButton icon="plus" @click="openCreateDrawer">新增</AoiButton>
      </template>
    </PageHeader>

    <AoiStatusMessage tone="danger" :message="error" />
    <AoiStatusMessage tone="success" :message="success" />

    <section class="customer-notice">
      <AoiIcon name="info" decorative />
      <span>资源权限示例：当前列表只展示本组织内自己创建的客户；当请求主体带有角色上下文时，也会展示同角色归属的客户资源。</span>
    </section>

    <article class="admin-card">
      <div class="admin-card__header customer-card-header">
        <div>
          <h2>客户资源</h2>
          <span>{{ persisted ? "已连接持久化表" : "等待数据表可用" }}</span>
        </div>
        <span class="badge">{{ pageData.total }} 条</span>
      </div>

      <div class="admin-filter-toolbar customer-filter-toolbar">
        <AoiTextField v-model="keyword" label="姓名 / 电话 / 接入人" icon="search" @enter="search" />
        <AoiSelect v-model="pageSize" label="每页" :options="[
          { label: '10 条/页', value: '10' },
          { label: '20 条/页', value: '20' },
          { label: '50 条/页', value: '50' }
        ]" />
        <AoiButton icon="search" :loading="loading" @click="search">查询</AoiButton>
        <AoiButton appearance="soft" icon="x" @click="resetFilters">重置</AoiButton>
      </div>

      <div class="data-table-wrap">
        <table class="data-table customer-table">
          <thead>
            <tr>
              <th>接入日期</th>
              <th>姓名</th>
              <th>电话</th>
              <th>接入人</th>
              <th>角色</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="item in pageData.items" :key="item.id">
              <td>{{ formatDateTime(item.createdAt) }}</td>
              <td>
                <strong>{{ item.customerName }}</strong>
              </td>
              <td>{{ item.customerPhoneData }}</td>
              <td>{{ ownerLabel(item) }}</td>
              <td>
                <span class="badge">{{ item.ownerRoleCode || "个人" }}</span>
              </td>
              <td>
                <div class="action-row">
                  <AoiButton appearance="soft" icon="edit-3" @click="openEditDrawer(item)">编辑</AoiButton>
                  <AoiButton appearance="soft" intent="danger" icon="trash-2" :disabled="deleting" @click="deleteCustomer(item)">删除</AoiButton>
                </div>
              </td>
            </tr>
            <tr v-if="!pageData.items.length">
              <td colspan="6" class="muted">暂无数据</td>
            </tr>
          </tbody>
        </table>
      </div>

      <div class="customer-pagination">
        <span>共 {{ pageData.total }} 条</span>
        <div class="action-row">
          <AoiButton appearance="soft" icon="chevron-left" :disabled="page <= 1 || loading" @click="previousPage">上一页</AoiButton>
          <span>{{ page }} / {{ totalPages }}</span>
          <AoiButton appearance="soft" trailing-icon="chevron-right" :disabled="page >= totalPages || loading" @click="nextPage">下一页</AoiButton>
        </div>
      </div>
    </article>

    <Teleport to="body">
      <div v-if="drawerOpen" class="customer-drawer-layer">
        <button class="customer-drawer-layer__scrim" type="button" aria-label="关闭客户表单" @click="closeDrawer" />
        <aside class="customer-drawer" aria-label="客户表单">
          <header class="customer-drawer__header">
            <div>
              <strong>{{ drawerTitle }}</strong>
              <span>填写客户名和客户电话，保存后按资源归属显示。</span>
            </div>
            <AoiIconButton icon="x" label="关闭客户表单" :disabled="saving" @click="closeDrawer" />
          </header>
          <form class="customer-drawer__body" @submit.prevent="submitForm">
            <AoiTextField v-model="form.customerName" label="客户名" icon="user" :disabled="saving" />
            <AoiTextField v-model="form.customerPhoneData" label="客户电话" icon="id-card" :disabled="saving" />
          </form>
          <footer class="customer-drawer__footer">
            <AoiButton appearance="plain" :disabled="saving" @click="closeDrawer">取消</AoiButton>
            <AoiButton icon="save" :loading="saving" :disabled="!canSave" @click="submitForm">确定</AoiButton>
          </footer>
        </aside>
      </div>
    </Teleport>
  </div>
</template>

<style scoped>
.customer-notice {
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

.customer-notice :deep(.aoi-icon) {
  flex: 0 0 auto;
  color: var(--aoi-warning);
}

.customer-card-header > div {
  display: grid;
  gap: 3px;
}

.customer-card-header span {
  color: var(--aoi-text-muted);
  font-size: 13px;
}

.customer-filter-toolbar {
  grid-template-columns: minmax(220px, 1fr) 150px auto auto;
}

.customer-table strong {
  color: var(--aoi-text);
  font-size: 14px;
}

.customer-pagination {
  display: flex;
  gap: 12px;
  align-items: center;
  justify-content: space-between;
  padding: 14px 16px;
  border-top: 1px solid var(--aoi-border);
  color: var(--aoi-text-muted);
  font-size: 13px;
}

.customer-drawer-layer {
  position: fixed;
  inset: 0;
  z-index: var(--aoi-z-dialog);
  display: flex;
  justify-content: flex-end;
}

.customer-drawer-layer__scrim {
  position: absolute;
  inset: 0;
  border: 0;
  background: rgba(15, 23, 42, .32);
}

.customer-drawer {
  position: relative;
  z-index: 1;
  display: grid;
  grid-template-rows: auto 1fr auto;
  width: min(420px, 100vw);
  min-height: 100%;
  border-left: 1px solid var(--aoi-border);
  background: var(--aoi-surface);
  box-shadow: var(--aoi-shadow-md);
}

.customer-drawer__header,
.customer-drawer__footer {
  display: flex;
  gap: 12px;
  align-items: center;
  justify-content: space-between;
  padding: 16px 18px;
  border-bottom: 1px solid var(--aoi-border);
}

.customer-drawer__footer {
  border-top: 1px solid var(--aoi-border);
  border-bottom: 0;
}

.customer-drawer__header div {
  display: grid;
  gap: 4px;
}

.customer-drawer__header strong {
  color: var(--aoi-text);
  font-size: 17px;
}

.customer-drawer__header span {
  color: var(--aoi-text-muted);
  font-size: 13px;
  line-height: 1.45;
}

.customer-drawer__body {
  display: grid;
  align-content: start;
  gap: 14px;
  padding: 18px;
}

@media (max-width: 760px) {
  .customer-filter-toolbar {
    grid-template-columns: 1fr;
  }

  .customer-pagination {
    align-items: flex-start;
    flex-direction: column;
  }

  .customer-drawer {
    width: 100vw;
  }
}
</style>
