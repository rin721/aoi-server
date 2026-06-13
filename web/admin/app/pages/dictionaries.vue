<script setup lang="ts">
import type { SystemDictionary, SystemDictionaryCatalog, SystemDictionaryItem } from "~/types/admin"

const api = useAdminApi()
const catalog = ref<SystemDictionaryCatalog>({ items: [], storageStatus: "unavailable", total: 0 })
const query = ref("")
const selectedDictionaryId = ref("")
const loading = ref(false)
const savingDictionary = ref(false)
const updatingDictionary = ref(false)
const savingItem = ref(false)
const updatingItem = ref(false)
const deleting = ref(false)
const error = ref("")
const success = ref("")

const createCode = ref("")
const createName = ref("")
const createDescription = ref("")
const createStatus = ref("active")

const editDictionaryId = ref("")
const editName = ref("")
const editDescription = ref("")
const editStatus = ref("active")

const itemLabel = ref("")
const itemValue = ref("")
const itemExtra = ref("")
const itemStatus = ref("active")
const itemSort = ref("10")

const editItemId = ref("")
const editItemLabel = ref("")
const editItemValue = ref("")
const editItemExtra = ref("")
const editItemStatus = ref("active")
const editItemSort = ref("10")

const statusOptions = [
  { label: "启用", value: "active" },
  { label: "停用", value: "disabled" }
]

const dictionaries = computed(() => catalog.value.items)
const itemCount = computed(() => dictionaries.value.reduce((count, dictionary) => count + dictionary.items.length, 0))
const activeDictionaryCount = computed(() => dictionaries.value.filter((dictionary) => dictionary.status === "active").length)
const activeItemCount = computed(() => dictionaries.value.reduce((count, dictionary) => count + dictionary.items.filter((item) => item.status === "active").length, 0))
const persisted = computed(() => catalog.value.storageStatus === "persisted")

const filteredDictionaries = computed(() => {
  const keyword = query.value.trim().toLowerCase()
  if (!keyword) {
    return dictionaries.value
  }
  return dictionaries.value.filter((dictionary) => matchesDictionary(dictionary, keyword))
})

const dictionaryOptions = computed(() => dictionaries.value.map((dictionary) => ({
  label: `${dictionary.name} (${dictionary.code})`,
  value: dictionary.id
})))

const selectedDictionary = computed(() => dictionaries.value.find((dictionary) => dictionary.id === selectedDictionaryId.value) || dictionaries.value[0] || null)
const selectedItems = computed(() => selectedDictionary.value?.items || [])

async function load(options: { silent?: boolean } = {}) {
  if (!options.silent) {
    loading.value = true
  }
  error.value = ""
  try {
    catalog.value = await api.listSystemDictionaries()
    if (!selectedDictionaryId.value || !dictionaries.value.some((dictionary) => dictionary.id === selectedDictionaryId.value)) {
      selectedDictionaryId.value = dictionaries.value[0]?.id || ""
    }
  } catch (err) {
    error.value = errorMessage(err)
  } finally {
    if (!options.silent) {
      loading.value = false
    }
  }
}

const autoRefresh = useAdminAutoRefresh({
  blocked: computed(() => loading.value || savingDictionary.value || updatingDictionary.value || savingItem.value || updatingItem.value || deleting.value),
  load
})

async function createDictionary() {
  if (!createCode.value.trim() || !createName.value.trim()) {
    return
  }
  savingDictionary.value = true
  error.value = ""
  success.value = ""
  try {
    const dictionary = await api.createSystemDictionary({
      code: createCode.value.trim(),
      description: createDescription.value.trim(),
      name: createName.value.trim(),
      status: createStatus.value
    })
    success.value = `字典 ${dictionary.name} 已创建。`
    createCode.value = ""
    createName.value = ""
    createDescription.value = ""
    createStatus.value = "active"
    selectedDictionaryId.value = dictionary.id
    await load()
  } catch (err) {
    error.value = errorMessage(err)
  } finally {
    savingDictionary.value = false
  }
}

function startEditDictionary(dictionary: SystemDictionary) {
  editDictionaryId.value = dictionary.id
  editName.value = dictionary.name
  editDescription.value = dictionary.description
  editStatus.value = dictionary.status
  selectedDictionaryId.value = dictionary.id
}

async function updateDictionary() {
  if (!editDictionaryId.value || !editName.value.trim()) {
    return
  }
  updatingDictionary.value = true
  error.value = ""
  success.value = ""
  try {
    const dictionary = await api.updateSystemDictionary(editDictionaryId.value, {
      description: editDescription.value.trim(),
      name: editName.value.trim(),
      status: editStatus.value
    })
    success.value = `字典 ${dictionary.name} 已更新。`
    await load()
  } catch (err) {
    error.value = errorMessage(err)
  } finally {
    updatingDictionary.value = false
  }
}

async function deleteDictionary(dictionary: SystemDictionary) {
  if (!window.confirm(`删除字典 ${dictionary.name}？关联字典项会一并删除。`)) {
    return
  }
  deleting.value = true
  error.value = ""
  success.value = ""
  try {
    await api.deleteSystemDictionary(dictionary.id)
    success.value = `字典 ${dictionary.name} 已删除。`
    if (selectedDictionaryId.value === dictionary.id) {
      selectedDictionaryId.value = ""
    }
    if (editDictionaryId.value === dictionary.id) {
      editDictionaryId.value = ""
    }
    await load()
  } catch (err) {
    error.value = errorMessage(err)
  } finally {
    deleting.value = false
  }
}

async function createItem() {
  if (!selectedDictionary.value || !itemLabel.value.trim() || !itemValue.value.trim()) {
    return
  }
  savingItem.value = true
  error.value = ""
  success.value = ""
  try {
    await api.createSystemDictionaryItem(selectedDictionary.value.id, {
      extra: itemExtra.value.trim(),
      label: itemLabel.value.trim(),
      sort: toSort(itemSort.value),
      status: itemStatus.value,
      value: itemValue.value.trim()
    })
    success.value = `字典项 ${itemLabel.value.trim()} 已创建。`
    itemLabel.value = ""
    itemValue.value = ""
    itemExtra.value = ""
    itemSort.value = "10"
    itemStatus.value = "active"
    await load()
  } catch (err) {
    error.value = errorMessage(err)
  } finally {
    savingItem.value = false
  }
}

function startEditItem(item: SystemDictionaryItem) {
  editItemId.value = item.id
  editItemLabel.value = item.label
  editItemValue.value = item.value
  editItemExtra.value = item.extra
  editItemSort.value = String(item.sort)
  editItemStatus.value = item.status
}

async function updateItem() {
  if (!editItemId.value || !editItemLabel.value.trim() || !editItemValue.value.trim()) {
    return
  }
  updatingItem.value = true
  error.value = ""
  success.value = ""
  try {
    await api.updateSystemDictionaryItem(editItemId.value, {
      extra: editItemExtra.value.trim(),
      label: editItemLabel.value.trim(),
      sort: toSort(editItemSort.value),
      status: editItemStatus.value,
      value: editItemValue.value.trim()
    })
    success.value = `字典项 ${editItemLabel.value.trim()} 已更新。`
    await load()
  } catch (err) {
    error.value = errorMessage(err)
  } finally {
    updatingItem.value = false
  }
}

async function deleteItem(item: SystemDictionaryItem) {
  if (!window.confirm(`删除字典项 ${item.label}？`)) {
    return
  }
  deleting.value = true
  error.value = ""
  success.value = ""
  try {
    await api.deleteSystemDictionaryItem(item.id)
    success.value = `字典项 ${item.label} 已删除。`
    if (editItemId.value === item.id) {
      editItemId.value = ""
    }
    await load()
  } catch (err) {
    error.value = errorMessage(err)
  } finally {
    deleting.value = false
  }
}

function matchesDictionary(dictionary: SystemDictionary, keyword: string) {
  return [
    dictionary.code,
    dictionary.name,
    dictionary.description,
    dictionary.status,
    ...dictionary.items.flatMap((item) => [item.label, item.value, item.extra, item.status])
  ].some((value) => String(value || "").toLowerCase().includes(keyword))
}

function toSort(value: string) {
  const parsed = Number.parseInt(value, 10)
  return Number.isFinite(parsed) ? parsed : 0
}

onMounted(autoRefresh.refreshNow)

watch(selectedDictionary, (dictionary) => {
  if (dictionary && !editDictionaryId.value) {
    startEditDictionary(dictionary)
  }
}, { immediate: true })

watch(editDictionaryId, (id) => {
  const dictionary = dictionaries.value.find((item) => item.id === id)
  if (!dictionary) {
    return
  }
  selectedDictionaryId.value = dictionary.id
  editName.value = dictionary.name
  editDescription.value = dictionary.description
  editStatus.value = dictionary.status
})

useHead({
  title: "字典管理 - Aoi Admin"
})
</script>

<template>
  <div class="page-grid">
    <PageHeader title="字典管理" icon="book-open" description="维护系统字典与字典项，为状态、枚举和表单选项提供统一配置。">
      <template #actions>
        <AoiButton appearance="soft" icon="code-2" to="/apis">API 管理</AoiButton>
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
    <AoiStatusMessage
      v-if="!persisted"
      tone="warning"
      message="system_dictionaries 表尚未可用，请先执行数据库迁移。"
    />

    <section class="admin-management-grid dictionary-grid">
      <article class="admin-card admin-management-grid__primary">
        <div class="admin-card__header">
          <h2>字典列表</h2>
          <div class="dictionary-summary">
            <span class="badge">{{ catalog.total }} 个字典</span>
            <span class="badge badge--success">{{ activeDictionaryCount }} 启用</span>
            <span class="badge">{{ itemCount }} 个字典项</span>
            <span class="badge badge--success">{{ activeItemCount }} 启用项</span>
          </div>
        </div>

        <div class="admin-filter-toolbar">
          <AoiTextField v-model="query" label="关键词" icon="search" placeholder="status 或 gender" />
        </div>

        <div class="data-table-wrap">
          <table class="data-table dictionary-table">
            <thead>
              <tr>
                <th>字典</th>
                <th>状态</th>
                <th>字典项</th>
                <th>说明</th>
                <th>操作</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="dictionary in filteredDictionaries"
                :key="dictionary.id"
                :class="{ 'dictionary-table__row--selected': selectedDictionary?.id === dictionary.id }"
              >
                <td data-label="字典">
                  <button class="dictionary-name" type="button" @click="selectedDictionaryId = dictionary.id; startEditDictionary(dictionary)">
                    <strong>{{ dictionary.name }}</strong>
                    <small>{{ dictionary.code }}</small>
                  </button>
                </td>
                <td data-label="状态">
                  <span class="badge" :class="dictionary.status === 'active' ? 'badge--success' : 'badge--warning'">
                    {{ formatStatus(dictionary.status) }}
                  </span>
                </td>
                <td data-label="字典项"><span class="badge">{{ dictionary.items.length }} 项</span></td>
                <td data-label="说明">{{ dictionary.description || "-" }}</td>
                <td data-label="操作">
                  <div class="table-actions">
                    <AoiButton appearance="soft" size="sm" icon="edit-3" @click="startEditDictionary(dictionary)">编辑</AoiButton>
                    <AoiButton appearance="soft" intent="danger" size="sm" icon="trash-2" @click="deleteDictionary(dictionary)">删除</AoiButton>
                  </div>
                </td>
              </tr>
              <tr v-if="!filteredDictionaries.length">
                <td colspan="5" class="muted">暂无匹配字典。</td>
              </tr>
            </tbody>
          </table>
        </div>

        <section v-if="selectedDictionary" class="dictionary-items">
          <div class="dictionary-items__header">
            <div>
              <h3>{{ selectedDictionary.name }} 字典项</h3>
              <p>{{ selectedDictionary.code }}</p>
            </div>
            <span class="badge">{{ selectedItems.length }} 项</span>
          </div>
          <div class="data-table-wrap">
            <table class="data-table dictionary-item-table">
              <thead>
                <tr>
                  <th>Label</th>
                  <th>Value</th>
                  <th>排序</th>
                  <th>状态</th>
                  <th>扩展</th>
                  <th>操作</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="item in selectedItems" :key="item.id">
                  <td data-label="Label"><strong>{{ item.label }}</strong></td>
                  <td class="mono" data-label="Value">{{ item.value }}</td>
                  <td class="mono" data-label="排序">{{ item.sort }}</td>
                  <td data-label="状态">
                    <span class="badge" :class="item.status === 'active' ? 'badge--success' : 'badge--warning'">
                      {{ formatStatus(item.status) }}
                    </span>
                  </td>
                  <td data-label="扩展">{{ item.extra || "-" }}</td>
                  <td data-label="操作">
                    <div class="table-actions">
                      <AoiButton appearance="soft" size="sm" icon="edit-3" @click="startEditItem(item)">编辑</AoiButton>
                      <AoiButton appearance="soft" intent="danger" size="sm" icon="trash-2" @click="deleteItem(item)">删除</AoiButton>
                    </div>
                  </td>
                </tr>
                <tr v-if="!selectedItems.length">
                  <td colspan="6" class="muted">暂无字典项。</td>
                </tr>
              </tbody>
            </table>
          </div>
        </section>
      </article>

      <article class="admin-card">
        <div class="admin-card__header">
          <h2>创建字典</h2>
        </div>
        <form class="admin-card__body form-grid" @submit.prevent="createDictionary">
          <AoiTextField v-model="createCode" label="字典 Code" icon="badge" placeholder="status" />
          <AoiTextField v-model="createName" label="字典名称" icon="book-open" placeholder="状态" />
          <AoiTextField v-model="createDescription" label="说明" icon="file-text" multiline :rows="3" />
          <AoiSelect :model-value="createStatus" label="状态" :options="statusOptions" @update:model-value="createStatus = $event" />
          <AoiButton type="submit" icon="plus" :loading="savingDictionary" :disabled="!createCode || !createName || !persisted">创建字典</AoiButton>
        </form>
      </article>

      <article class="admin-card">
        <div class="admin-card__header">
          <h2>编辑字典</h2>
        </div>
        <form class="admin-card__body form-grid" @submit.prevent="updateDictionary">
          <AoiSelect :model-value="editDictionaryId" label="选择字典" :options="dictionaryOptions" @update:model-value="editDictionaryId = $event" />
          <AoiTextField v-model="editName" label="字典名称" icon="book-open" />
          <AoiTextField v-model="editDescription" label="说明" icon="file-text" multiline :rows="3" />
          <AoiSelect :model-value="editStatus" label="状态" :options="statusOptions" @update:model-value="editStatus = $event" />
          <AoiButton type="submit" icon="save" :loading="updatingDictionary" :disabled="!editDictionaryId || !editName || !persisted">保存字典</AoiButton>
        </form>
      </article>

      <article class="admin-card">
        <div class="admin-card__header">
          <h2>创建字典项</h2>
          <span v-if="selectedDictionary" class="badge">{{ selectedDictionary.code }}</span>
        </div>
        <form class="admin-card__body form-grid" @submit.prevent="createItem">
          <AoiTextField v-model="itemLabel" label="Label" icon="tag" placeholder="启用" />
          <AoiTextField v-model="itemValue" label="Value" icon="hash" placeholder="active" />
          <AoiTextField v-model="itemSort" label="排序" icon="arrow-down-up" type="number" min="0" step="1" />
          <AoiSelect :model-value="itemStatus" label="状态" :options="statusOptions" @update:model-value="itemStatus = $event" />
          <AoiTextField v-model="itemExtra" label="扩展信息" icon="braces" multiline :rows="3" placeholder='{"color":"green"}' />
          <AoiButton type="submit" icon="plus" :loading="savingItem" :disabled="!selectedDictionary || !itemLabel || !itemValue || !persisted">创建字典项</AoiButton>
        </form>
      </article>

      <article class="admin-card">
        <div class="admin-card__header">
          <h2>编辑字典项</h2>
        </div>
        <form class="admin-card__body form-grid" @submit.prevent="updateItem">
          <AoiTextField :model-value="editItemId || '-'" label="字典项 ID" icon="fingerprint" disabled />
          <AoiTextField v-model="editItemLabel" label="Label" icon="tag" />
          <AoiTextField v-model="editItemValue" label="Value" icon="hash" />
          <AoiTextField v-model="editItemSort" label="排序" icon="arrow-down-up" type="number" min="0" step="1" />
          <AoiSelect :model-value="editItemStatus" label="状态" :options="statusOptions" @update:model-value="editItemStatus = $event" />
          <AoiTextField v-model="editItemExtra" label="扩展信息" icon="braces" multiline :rows="3" />
          <AoiButton type="submit" icon="save" :loading="updatingItem" :disabled="!editItemId || !editItemLabel || !editItemValue || !persisted">保存字典项</AoiButton>
        </form>
      </article>
    </section>
  </div>
</template>

<style scoped>
.dictionary-grid {
  align-items: start;
}

.dictionary-summary,
.table-actions {
  align-items: center;
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.dictionary-summary {
  justify-content: flex-end;
}

.dictionary-name {
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

.dictionary-name strong,
.dictionary-name small {
  min-width: 0;
}

.dictionary-name small,
.dictionary-items__header p {
  color: var(--aoi-text-muted);
  font-size: 12px;
  margin: 0;
}

.dictionary-table__row--selected {
  background: var(--aoi-accent-10);
}

.dictionary-items {
  border-top: 1px solid var(--aoi-border);
  display: grid;
  gap: 12px;
  padding: 16px 14px 14px;
}

.dictionary-items__header {
  align-items: center;
  display: flex;
  justify-content: space-between;
  gap: 12px;
}

.dictionary-items__header h3 {
  font-size: 15px;
  margin: 0 0 3px;
}

@media (max-width: 760px) {
  .admin-card__header {
    align-items: flex-start;
    flex-direction: column;
    gap: 10px;
  }

  .dictionary-summary {
    justify-content: flex-start;
    width: 100%;
  }

  .dictionary-table,
  .dictionary-item-table {
    min-width: 0;
    width: 100%;
  }

  .dictionary-table,
  .dictionary-item-table,
  .dictionary-table tbody,
  .dictionary-item-table tbody,
  .dictionary-table tr,
  .dictionary-item-table tr,
  .dictionary-table td,
  .dictionary-item-table td {
    display: block;
  }

  .dictionary-table thead,
  .dictionary-item-table thead {
    display: none;
  }

  .dictionary-table tr,
  .dictionary-item-table tr {
    border-bottom: 1px solid var(--aoi-border);
    padding: 10px 0;
    width: 100%;
  }

  .dictionary-table tr:last-child,
  .dictionary-item-table tr:last-child {
    border-bottom: 0;
  }

  .dictionary-table td,
  .dictionary-item-table td {
    align-items: flex-start;
    border-bottom: 0;
    display: grid;
    gap: 8px;
    grid-template-columns: 76px minmax(0, 1fr);
    padding: 5px 0;
    white-space: normal;
    width: 100%;
  }

  .dictionary-table td:last-child,
  .dictionary-item-table td:last-child {
    white-space: normal;
  }

  .dictionary-table td > *,
  .dictionary-item-table td > * {
    max-width: 100%;
    min-width: 0;
    overflow-wrap: anywhere;
  }

  .dictionary-name {
    width: 100%;
  }

  .dictionary-table td::before,
  .dictionary-item-table td::before {
    color: var(--aoi-text-muted);
    content: attr(data-label);
    font-size: 12px;
    font-weight: 700;
  }

  .table-actions {
    align-items: stretch;
  }
}
</style>
