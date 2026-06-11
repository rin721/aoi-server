<script setup lang="ts">
import type { ID, SystemParameter, SystemParameterPage } from "~/types/admin"

const api = useAdminApi()
const pageData = ref<SystemParameterPage>({ items: [], page: 1, pageSize: 10, storageStatus: "unavailable", total: 0 })
const name = ref("")
const key = ref("")
const startCreatedAt = ref("")
const endCreatedAt = ref("")
const page = ref(1)
const pageSize = ref("10")
const selectedIds = ref<ID[]>([])
const editParameterId = ref("")
const formName = ref("")
const formKey = ref("")
const formValue = ref("")
const formDescription = ref("")
const loading = ref(false)
const saving = ref(false)
const deleting = ref(false)
const error = ref("")
const success = ref("")

const persisted = computed(() => pageData.value.storageStatus === "persisted")
const pageSizeNumber = computed(() => Math.min(100, Math.max(1, Number(pageSize.value) || 10)))
const totalPages = computed(() => Math.max(1, Math.ceil(pageData.value.total / pageSizeNumber.value)))
const hasSelection = computed(() => selectedIds.value.length > 0)
const editing = computed(() => Boolean(editParameterId.value))
const selectedParameter = computed(() => pageData.value.items.find((item) => item.id === editParameterId.value) || null)

async function load() {
  loading.value = true
  error.value = ""
  try {
    pageData.value = await api.listSystemParameters({
      endCreatedAt: toQueryDate(endCreatedAt.value),
      key: key.value.trim() || undefined,
      name: name.value.trim() || undefined,
      page: page.value,
      pageSize: pageSizeNumber.value,
      startCreatedAt: toQueryDate(startCreatedAt.value)
    })
    selectedIds.value = selectedIds.value.filter((id) => pageData.value.items.some((item) => item.id === id))
  } catch (err) {
    error.value = errorMessage(err)
  } finally {
    loading.value = false
  }
}

async function resetFilters() {
  name.value = ""
  key.value = ""
  startCreatedAt.value = ""
  endCreatedAt.value = ""
  page.value = 1
  await load()
}

async function submitParameter() {
  if (!formName.value.trim() || !formKey.value.trim() || !formValue.value.trim()) {
    return
  }
  saving.value = true
  error.value = ""
  success.value = ""
  try {
    if (editing.value) {
      const parameter = await api.updateSystemParameter(editParameterId.value, formBody())
      success.value = `参数 ${parameter.name} 已更新。`
    } else {
      const parameter = await api.createSystemParameter(formBody())
      success.value = `参数 ${parameter.name} 已创建。`
      editParameterId.value = parameter.id
    }
    await load()
  } catch (err) {
    error.value = errorMessage(err)
  } finally {
    saving.value = false
  }
}

function startCreate() {
  editParameterId.value = ""
  formName.value = ""
  formKey.value = ""
  formValue.value = ""
  formDescription.value = ""
}

function startEdit(parameter: SystemParameter) {
  editParameterId.value = parameter.id
  formName.value = parameter.name
  formKey.value = parameter.key
  formValue.value = parameter.value
  formDescription.value = parameter.description
}

async function deleteOne(parameter: SystemParameter) {
  if (!confirm(`删除参数 ${parameter.name}？`)) {
    return
  }
  deleting.value = true
  error.value = ""
  success.value = ""
  try {
    await api.deleteSystemParameter(parameter.id)
    success.value = `参数 ${parameter.name} 已删除。`
    if (editParameterId.value === parameter.id) {
      startCreate()
    }
    await load()
  } catch (err) {
    error.value = errorMessage(err)
  } finally {
    deleting.value = false
  }
}

async function deleteSelected() {
  if (!hasSelection.value || !confirm(`删除 ${selectedIds.value.length} 个参数？`)) {
    return
  }
  deleting.value = true
  error.value = ""
  success.value = ""
  try {
    await api.deleteSystemParameters(selectedIds.value)
    success.value = `已删除 ${selectedIds.value.length} 个参数。`
    if (selectedIds.value.includes(editParameterId.value)) {
      startCreate()
    }
    selectedIds.value = []
    await load()
  } catch (err) {
    error.value = errorMessage(err)
  } finally {
    deleting.value = false
  }
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

function toggleAll(event: Event) {
  const checked = (event.target as HTMLInputElement).checked
  selectedIds.value = checked ? pageData.value.items.map((item) => item.id) : []
}

function formBody() {
  return {
    description: formDescription.value.trim(),
    key: formKey.value.trim(),
    name: formName.value.trim(),
    value: formValue.value.trim()
  }
}

function toQueryDate(value: string) {
  return value || undefined
}

onMounted(load)

watch(pageSize, async () => {
  page.value = 1
  await load()
})

useHead({
  title: "参数管理 - Aoi Admin"
})
</script>

<template>
  <div class="page-grid">
    <PageHeader title="参数管理" icon="compass" description="维护系统运行期可读取的键值参数，与 GVA 的 sys_params 能力保持同类。">
      <template #actions>
        <AoiButton appearance="soft" icon="plus" @click="startCreate">新增</AoiButton>
        <AoiButton appearance="soft" icon="refresh-cw" :loading="loading" @click="load">刷新</AoiButton>
        <AoiButton appearance="soft" intent="danger" icon="trash-2" :disabled="!hasSelection || !persisted" :loading="deleting" @click="deleteSelected">删除</AoiButton>
      </template>
    </PageHeader>

    <AoiStatusMessage tone="danger" :message="error" />
    <AoiStatusMessage tone="success" :message="success" />
    <AoiStatusMessage v-if="!persisted" tone="warning" message="system_parameters 表尚未可用，请先执行数据库迁移。" />

    <section class="admin-management-grid parameter-grid">
      <article class="admin-card admin-management-grid__primary">
        <div class="admin-card__header">
          <div>
            <h2>参数列表</h2>
            <p class="muted">共 {{ pageData.total }} 个</p>
          </div>
          <div class="parameter-pager">
            <AoiButton appearance="soft" size="sm" icon="chevron-left" :disabled="page <= 1" @click="previousPage">上一页</AoiButton>
            <span class="badge">{{ page }} / {{ totalPages }}</span>
            <AoiButton appearance="soft" size="sm" icon="chevron-right" :disabled="page >= totalPages" @click="nextPage">下一页</AoiButton>
          </div>
        </div>

        <div class="admin-filter-toolbar">
          <AoiTextField v-model="name" label="参数名称" icon="search" placeholder="站点名称" @enter="page = 1; load()" />
          <AoiTextField v-model="key" label="参数键" icon="key-round" placeholder="site.name" @enter="page = 1; load()" />
          <AoiTextField v-model="startCreatedAt" label="开始日期" icon="calendar" type="date" />
          <AoiTextField v-model="endCreatedAt" label="结束日期" icon="calendar" type="date" />
          <AoiTextField v-model="pageSize" label="每页" icon="list-filter" type="number" min="1" max="100" step="1" @enter="page = 1; load()" />
          <AoiButton appearance="soft" icon="search" :loading="loading" @click="page = 1; load()">查询</AoiButton>
          <AoiButton appearance="plain" icon="rotate-ccw" @click="resetFilters">重置</AoiButton>
        </div>

        <div class="data-table-wrap">
          <table class="data-table parameter-table">
            <thead>
              <tr>
                <th>
                  <input
                    aria-label="选择所有参数"
                    class="parameter-check"
                    type="checkbox"
                    :checked="Boolean(pageData.items.length) && selectedIds.length === pageData.items.length"
                    :disabled="!pageData.items.length"
                    @change="toggleAll"
                  >
                </th>
                <th>日期</th>
                <th>参数名称</th>
                <th>参数键</th>
                <th>参数值</th>
                <th>参数说明</th>
                <th>操作</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="parameter in pageData.items"
                :key="parameter.id"
                :class="{ 'parameter-table__row--selected': selectedParameter?.id === parameter.id }"
              >
                <td data-label="选择">
                  <input v-model="selectedIds" class="parameter-check" type="checkbox" :value="parameter.id" :aria-label="`选择参数 ${parameter.id}`">
                </td>
                <td data-label="日期">{{ formatDateTime(parameter.createdAt) }}</td>
                <td data-label="参数名称">
                  <button class="parameter-name" type="button" @click="startEdit(parameter)">
                    <strong>{{ parameter.name }}</strong>
                    <small>{{ parameter.id }}</small>
                  </button>
                </td>
                <td class="mono" data-label="参数键">{{ parameter.key }}</td>
                <td data-label="参数值"><span class="parameter-value">{{ parameter.value }}</span></td>
                <td data-label="参数说明">{{ parameter.description || "-" }}</td>
                <td data-label="操作">
                  <div class="table-actions">
                    <AoiButton appearance="soft" size="sm" icon="edit-3" @click="startEdit(parameter)">编辑</AoiButton>
                    <AoiButton appearance="soft" intent="danger" size="sm" icon="trash-2" :loading="deleting" @click="deleteOne(parameter)">删除</AoiButton>
                  </div>
                </td>
              </tr>
              <tr v-if="!pageData.items.length">
                <td colspan="7" class="muted">暂无参数。</td>
              </tr>
            </tbody>
          </table>
        </div>
      </article>

      <article class="admin-card parameter-form-card">
        <div class="admin-card__header">
          <h2>{{ editing ? "编辑参数" : "新增参数" }}</h2>
          <span v-if="editing" class="badge">{{ editParameterId }}</span>
        </div>
        <form class="admin-card__body form-grid" @submit.prevent="submitParameter">
          <AoiTextField v-model="formName" label="参数名称" icon="compass" placeholder="站点名称" />
          <AoiTextField v-model="formKey" label="参数键" icon="key-round" placeholder="site.name" />
          <AoiTextField v-model="formValue" label="参数值" icon="file-text" multiline :rows="5" />
          <AoiTextField v-model="formDescription" label="参数说明" icon="file-text" multiline :rows="3" />
          <div class="parameter-form-actions">
            <AoiButton type="submit" icon="save" :loading="saving" :disabled="!formName || !formKey || !formValue || !persisted">{{ editing ? "保存参数" : "创建参数" }}</AoiButton>
            <AoiButton appearance="plain" icon="x" type="button" @click="startCreate">清空</AoiButton>
          </div>
        </form>
      </article>
    </section>
  </div>
</template>

<style scoped>
.parameter-grid {
  align-items: start;
}

.parameter-pager,
.table-actions,
.parameter-form-actions {
  align-items: center;
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.parameter-check {
  accent-color: var(--aoi-accent-60);
  block-size: 16px;
  inline-size: 16px;
}

.parameter-name {
  background: transparent;
  border: 0;
  color: inherit;
  cursor: pointer;
  display: grid;
  font: inherit;
  gap: 2px;
  padding: 0;
  text-align: left;
}

.parameter-name strong,
.parameter-name small {
  min-width: 0;
  overflow-wrap: anywhere;
}

.parameter-name small {
  color: var(--aoi-text-muted);
  font-size: 12px;
}

.parameter-value {
  background: var(--aoi-surface-soft);
  border: 1px solid var(--aoi-border);
  border-radius: 8px;
  color: var(--aoi-text);
  display: block;
  max-width: 260px;
  overflow-wrap: anywhere;
  padding: 6px 8px;
  white-space: pre-wrap;
}

.parameter-table__row--selected {
  background: var(--aoi-accent-10);
}

.parameter-form-card {
  position: sticky;
  top: 18px;
}

.parameter-table th:first-child,
.parameter-table td:first-child {
  width: 42px;
}

@media (max-width: 980px) {
  .parameter-form-card {
    position: static;
  }
}

@media (max-width: 760px) {
  .admin-card__header {
    align-items: flex-start;
    flex-direction: column;
  }

  .parameter-table {
    min-width: 0;
    width: 100%;
  }

  .parameter-table thead {
    display: none;
  }

  .parameter-table,
  .parameter-table tbody,
  .parameter-table tr,
  .parameter-table td {
    display: block;
  }

  .parameter-table tr {
    border-bottom: 1px solid var(--aoi-border);
    padding: 10px 0;
    width: 100%;
  }

  .parameter-table tr:last-child {
    border-bottom: 0;
  }

  .parameter-table td,
  .parameter-table td:last-child {
    align-items: flex-start;
    border-bottom: 0;
    display: grid;
    gap: 8px;
    grid-template-columns: 78px minmax(0, 1fr);
    padding: 5px 0;
    white-space: normal;
    width: 100%;
  }

  .parameter-table td::before {
    color: var(--aoi-text-muted);
    content: attr(data-label);
    font-size: 12px;
    font-weight: 700;
  }

  .parameter-table td > * {
    justify-self: start;
    max-width: 100%;
    min-width: 0;
  }

  .parameter-value {
    justify-self: stretch;
    max-width: 100%;
  }
}
</style>
