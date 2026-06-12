<script setup lang="ts">
import type { ID, SystemMediaAsset, SystemMediaAssetPage, SystemMediaCategory, SystemMediaCategoryCatalog } from "~/types/admin"

const api = useAdminApi()

const categories = ref<SystemMediaCategoryCatalog>({ items: [], storageStatus: "unavailable", total: 0 })
const pageData = ref<SystemMediaAssetPage>({
  items: [],
  objectStorage: "unavailable",
  page: 1,
  pageSize: 10,
  storageStatus: "unavailable",
  total: 0,
  uploadMaxBytes: 20 * 1024 * 1024,
  uploadMaxMb: 20,
  uploadUnavailable: true
})
const selectedCategoryId = ref<ID>("0")
const keyword = ref("")
const page = ref(1)
const pageSize = ref("10")
const loading = ref(false)
const categoryLoading = ref(false)
const uploading = ref(false)
const importing = ref(false)
const savingCategory = ref(false)
const deleting = ref(false)
const error = ref("")
const success = ref("")
const fileInput = ref<HTMLInputElement | null>(null)

const importOpen = ref(false)
const importText = ref("")
const renameOpen = ref(false)
const renameTarget = ref<SystemMediaAsset | null>(null)
const renameValue = ref("")
const categoryOpen = ref(false)
const editingCategory = ref<SystemMediaCategory | null>(null)
const categoryName = ref("")
const categoryParentId = ref<ID>("0")
const categorySort = ref("10")

type FlatCategory = SystemMediaCategory & { depth: number }

function normalizeCategoryCatalog(catalog: SystemMediaCategoryCatalog): SystemMediaCategoryCatalog {
  const items = Array.isArray(catalog.items) ? catalog.items : []

  return {
    ...catalog,
    items,
    total: Number(catalog.total) || items.length
  }
}

function normalizeAssetPage(page: SystemMediaAssetPage): SystemMediaAssetPage {
  return {
    ...page,
    items: Array.isArray(page.items) ? page.items : [],
    total: Number(page.total) || 0
  }
}

const persisted = computed(() => pageData.value.storageStatus === "persisted")
const categoryPersisted = computed(() => categories.value.storageStatus === "persisted")
const pageSizeNumber = computed(() => Math.min(100, Math.max(1, Number(pageSize.value) || 10)))
const totalPages = computed(() => Math.max(1, Math.ceil(pageData.value.total / pageSizeNumber.value)))
const flatCategories = computed(() => flattenCategories(categories.value.items))
const selectedCategoryName = computed(() => {
  if (selectedCategoryId.value === "0") {
    return "全部分类"
  }
  return flatCategories.value.find((item) => item.id === selectedCategoryId.value)?.name || "未命名分类"
})
const uploadDisabled = computed(() => uploading.value || !persisted.value || pageData.value.uploadUnavailable)
const categoryOptions = computed(() => [
  { label: "根分类", value: "0" },
  ...flatCategories.value
    .filter((item) => !editingCategory.value || item.id !== editingCategory.value.id)
    .map((item) => ({ label: `${"　".repeat(item.depth)}${item.name}`, value: item.id }))
])

async function loadCategories() {
  categoryLoading.value = true
  try {
    categories.value = normalizeCategoryCatalog(await api.listSystemMediaCategories())
  } catch (err) {
    error.value = errorMessage(err)
  } finally {
    categoryLoading.value = false
  }
}

async function loadAssets() {
  loading.value = true
  error.value = ""
  try {
    pageData.value = normalizeAssetPage(await api.listSystemMediaAssets({
      categoryId: selectedCategoryId.value === "0" ? undefined : selectedCategoryId.value,
      keyword: keyword.value.trim() || undefined,
      page: page.value,
      pageSize: pageSizeNumber.value
    }))
  } catch (err) {
    error.value = errorMessage(err)
  } finally {
    loading.value = false
  }
}

async function refreshAll() {
  await Promise.all([loadCategories(), loadAssets()])
}

function chooseCategory(id: ID) {
  selectedCategoryId.value = id
  page.value = 1
  void loadAssets()
}

function openFilePicker() {
  fileInput.value?.click()
}

async function onFileSelected(event: Event) {
  const input = event.target as HTMLInputElement
  const files = Array.from(input.files || [])
  input.value = ""
  if (!files.length || uploadDisabled.value) {
    return
  }

  uploading.value = true
  error.value = ""
  success.value = ""
  try {
    for (const file of files) {
      await api.uploadSystemMediaAsset(file, selectedCategoryId.value === "0" ? undefined : selectedCategoryId.value)
    }
    success.value = `已上传 ${files.length} 个文件。`
    await loadAssets()
  } catch (err) {
    error.value = errorMessage(err)
  } finally {
    uploading.value = false
  }
}

function openImportDialog() {
  importText.value = ""
  importOpen.value = true
}

async function importURLs() {
  if (!importText.value.trim()) {
    return
  }
  importing.value = true
  error.value = ""
  success.value = ""
  try {
    const result = await api.importSystemMediaURLs({
      categoryId: selectedCategoryId.value === "0" ? undefined : selectedCategoryId.value,
      text: importText.value
    })
    success.value = `已导入 ${result.imported} 条外链。`
    importOpen.value = false
    await loadAssets()
  } catch (err) {
    error.value = errorMessage(err)
  } finally {
    importing.value = false
  }
}

function openRenameDialog(item: SystemMediaAsset) {
  renameTarget.value = item
  renameValue.value = item.displayName
  renameOpen.value = true
}

async function renameAsset() {
  if (!renameTarget.value || !renameValue.value.trim()) {
    return
  }
  error.value = ""
  success.value = ""
  try {
    await api.updateSystemMediaAsset(renameTarget.value.id, { displayName: renameValue.value.trim() })
    success.value = "文件名已更新。"
    renameOpen.value = false
    await loadAssets()
  } catch (err) {
    error.value = errorMessage(err)
  }
}

async function downloadAsset(item: SystemMediaAsset) {
  error.value = ""
  success.value = ""
  try {
    if (item.external) {
      window.open(item.url, "_blank", "noopener,noreferrer")
      return
    }
    const download = await api.downloadSystemMediaAsset(item.id)
    saveBlob(download.blob, download.filename || `${item.displayName || item.id}.${item.extension || "bin"}`)
    success.value = "文件已下载。"
  } catch (err) {
    error.value = errorMessage(err)
  }
}

async function deleteAsset(item: SystemMediaAsset) {
  if (!window.confirm(`删除 ${item.displayName || item.originalName}？`)) {
    return
  }
  deleting.value = true
  error.value = ""
  success.value = ""
  try {
    await api.deleteSystemMediaAsset(item.id)
    success.value = "文件已删除。"
    await loadAssets()
  } catch (err) {
    error.value = errorMessage(err)
  } finally {
    deleting.value = false
  }
}

function openCreateCategory() {
  editingCategory.value = null
  categoryName.value = ""
  categoryParentId.value = selectedCategoryId.value === "0" ? "0" : selectedCategoryId.value
  categorySort.value = "10"
  categoryOpen.value = true
}

function openEditCategory(item: SystemMediaCategory) {
  editingCategory.value = item
  categoryName.value = item.name
  categoryParentId.value = item.parentId || "0"
  categorySort.value = String(item.sort)
  categoryOpen.value = true
}

async function saveCategory() {
  if (!categoryName.value.trim()) {
    return
  }
  savingCategory.value = true
  error.value = ""
  success.value = ""
  try {
    await api.saveSystemMediaCategory({
      id: editingCategory.value?.id,
      name: categoryName.value.trim(),
      parentId: categoryParentId.value === "0" ? undefined : categoryParentId.value,
      sort: Number(categorySort.value) || 0
    })
    success.value = editingCategory.value ? "分类已更新。" : "分类已创建。"
    categoryOpen.value = false
    await loadCategories()
  } catch (err) {
    error.value = errorMessage(err)
  } finally {
    savingCategory.value = false
  }
}

async function deleteCategory(item: SystemMediaCategory) {
  if (!window.confirm(`删除分类 ${item.name}？分类下存在子分类或文件时不能删除。`)) {
    return
  }
  error.value = ""
  success.value = ""
  try {
    await api.deleteSystemMediaCategory(item.id)
    success.value = "分类已删除。"
    if (selectedCategoryId.value === item.id) {
      selectedCategoryId.value = "0"
      page.value = 1
      await loadAssets()
    }
    await loadCategories()
  } catch (err) {
    error.value = errorMessage(err)
  }
}

async function resetFilters() {
  keyword.value = ""
  page.value = 1
  await loadAssets()
}

async function previousPage() {
  if (page.value <= 1) {
    return
  }
  page.value -= 1
  await loadAssets()
}

async function nextPage() {
  if (page.value >= totalPages.value) {
    return
  }
  page.value += 1
  await loadAssets()
}

function flattenCategories(items: SystemMediaCategory[] | null | undefined, depth = 0): FlatCategory[] {
  return (items || []).flatMap((item) => [
    { ...item, depth },
    ...flattenCategories(item.children || [], depth + 1)
  ])
}

function isImageAsset(item: SystemMediaAsset) {
  return item.mimeType.startsWith("image/") || ["jpg", "jpeg", "png", "gif", "webp", "svg"].includes(item.extension.toLowerCase())
}

function sourceLabel(item: SystemMediaAsset) {
  return item.external ? "外链" : "本地"
}

function assetTypeLabel(item: SystemMediaAsset) {
  return item.extension ? item.extension.toUpperCase() : item.mimeType || "-"
}

function formatBytes(value: number) {
  if (!value) {
    return "-"
  }
  const units = ["B", "KB", "MB", "GB"]
  let size = value
  let unit = 0
  while (size >= 1024 && unit < units.length - 1) {
    size /= 1024
    unit += 1
  }
  return `${size.toFixed(unit === 0 ? 0 : 1)} ${units[unit]}`
}

function saveBlob(blob: Blob, filename: string) {
  const url = URL.createObjectURL(blob)
  const link = document.createElement("a")
  link.href = url
  link.download = filename
  link.click()
  URL.revokeObjectURL(url)
}

onMounted(async () => {
  await refreshAll()
})

watch(pageSize, async () => {
  page.value = 1
  await loadAssets()
})

useHead({
  title: "媒体库 - Aoi Admin"
})
</script>

<template>
  <div class="page-grid">
    <PageHeader title="媒体库" icon="image-up" description="按 GVA 媒体库入口管理本地上传文件和外链资源。">
      <template #actions>
        <AoiButton appearance="soft" icon="refresh-cw" :loading="loading || categoryLoading" @click="refreshAll">刷新</AoiButton>
      </template>
    </PageHeader>

    <AoiStatusMessage tone="danger" :message="error" />
    <AoiStatusMessage tone="success" :message="success" />
    <AoiStatusMessage v-if="!persisted" tone="warning" message="媒体库数据表尚不可用，请先执行数据库迁移。" />

    <section class="media-workspace">
      <aside class="media-categories">
        <div class="media-panel-header">
          <div>
            <h2>分类</h2>
            <p>{{ categories.total }} 个分类</p>
          </div>
          <AoiButton appearance="soft" size="sm" icon="plus" :disabled="!categoryPersisted" @click="openCreateCategory">新增</AoiButton>
        </div>

        <button
          class="category-row"
          :class="{ 'category-row--active': selectedCategoryId === '0' }"
          type="button"
          @click="chooseCategory('0')"
        >
          <AoiIcon name="folder-open" decorative />
          <span>全部分类</span>
        </button>

        <div class="category-list">
          <div v-for="item in flatCategories" :key="item.id" class="category-line">
            <button
              class="category-row"
              :class="{ 'category-row--active': selectedCategoryId === item.id }"
              :style="{ paddingLeft: `${12 + item.depth * 16}px` }"
              type="button"
              @click="chooseCategory(item.id)"
            >
              <AoiIcon name="folder" decorative />
              <span>{{ item.name }}</span>
            </button>
            <div class="category-actions">
              <AoiButton appearance="plain" size="sm" icon="edit-3" aria-label="编辑分类" @click="openEditCategory(item)" />
              <AoiButton appearance="plain" intent="danger" size="sm" icon="trash-2" aria-label="删除分类" @click="deleteCategory(item)" />
            </div>
          </div>
          <p v-if="!flatCategories.length" class="muted media-empty">暂无分类</p>
        </div>
      </aside>

      <article class="admin-card media-assets">
        <div class="admin-card__header media-assets-header">
          <div>
            <h2>{{ selectedCategoryName }}</h2>
            <p class="muted">共 {{ pageData.total }} 个文件，上传上限 {{ pageData.uploadMaxMb }} MB</p>
          </div>
          <div class="media-pager">
            <AoiButton appearance="soft" size="sm" icon="chevron-left" :disabled="page <= 1" @click="previousPage">上一页</AoiButton>
            <span class="badge">{{ page }} / {{ totalPages }}</span>
            <AoiButton appearance="soft" size="sm" icon="chevron-right" :disabled="page >= totalPages" @click="nextPage">下一页</AoiButton>
          </div>
        </div>

        <div class="media-alert">
          <AoiIcon name="circle-alert" decorative />
          <span>点击文件名可编辑；当前选中的分类会作为上传分类。外链导入只保存 URL，不抓取远程文件。</span>
        </div>

        <div class="media-toolbar">
          <input ref="fileInput" class="media-file-input" multiple type="file" @change="onFileSelected">
          <AoiButton appearance="soft" icon="upload" :disabled="uploadDisabled" :loading="uploading" @click="openFilePicker">普通上传</AoiButton>
          <AoiButton appearance="soft" icon="link" :disabled="!persisted" @click="openImportDialog">导入URL</AoiButton>
          <AoiTextField v-model="keyword" label="文件名或备注" icon="search" placeholder="请输入文件名或备注" @enter="page = 1; loadAssets()" />
          <AoiTextField v-model="pageSize" label="每页" icon="list-filter" type="number" min="1" max="100" step="1" @enter="page = 1; loadAssets()" />
          <AoiButton appearance="soft" icon="search" :loading="loading" @click="page = 1; loadAssets()">查询</AoiButton>
          <AoiButton appearance="plain" icon="rotate-ccw" @click="resetFilters">重置</AoiButton>
        </div>

        <AoiStatusMessage
          v-if="pageData.uploadUnavailable"
          tone="warning"
          message="对象存储未启用：可以浏览和导入外链，但不能上传或下载本地文件。"
        />

        <div class="data-table-wrap">
          <table class="data-table media-table">
            <thead>
              <tr>
                <th>预览</th>
                <th>日期</th>
                <th>文件名/备注</th>
                <th>链接</th>
                <th>标签</th>
                <th>大小</th>
                <th>操作</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="item in pageData.items" :key="item.id">
                <td data-label="预览">
                  <div class="asset-preview">
                    <img v-if="item.external && isImageAsset(item)" :src="item.url" :alt="item.displayName" loading="lazy">
                    <AoiIcon v-else :name="isImageAsset(item) ? 'file-image' : 'file'" decorative />
                  </div>
                </td>
                <td data-label="日期">{{ formatDateTime(item.createdAt) }}</td>
                <td data-label="文件名/备注">
                  <button class="asset-name" type="button" @click="openRenameDialog(item)">
                    <strong>{{ item.displayName }}</strong>
                    <small>{{ item.originalName }}</small>
                  </button>
                </td>
                <td data-label="链接">
                  <a v-if="item.external" class="asset-link" :href="item.url" target="_blank" rel="noopener noreferrer">
                    <AoiIcon name="external-link" decorative />
                    <span>打开外链</span>
                  </a>
                  <span v-else class="asset-link asset-link--local">
                    <AoiIcon name="download" decorative />
                    <span>鉴权下载</span>
                  </span>
                </td>
                <td data-label="标签">
                  <span class="badge">{{ sourceLabel(item) }}</span>
                  <span class="badge">{{ assetTypeLabel(item) }}</span>
                </td>
                <td class="mono" data-label="大小">{{ formatBytes(item.sizeBytes) }}</td>
                <td data-label="操作">
                  <div class="table-actions">
                    <AoiButton appearance="soft" size="sm" icon="download" @click="downloadAsset(item)">
                      {{ item.external ? "打开" : "下载" }}
                    </AoiButton>
                    <AoiButton appearance="soft" size="sm" icon="edit-3" @click="openRenameDialog(item)">编辑</AoiButton>
                    <AoiButton appearance="soft" intent="danger" size="sm" icon="trash-2" :disabled="deleting" @click="deleteAsset(item)">删除</AoiButton>
                  </div>
                </td>
              </tr>
              <tr v-if="!pageData.items.length">
                <td colspan="7" class="muted media-empty">暂无媒体资源</td>
              </tr>
            </tbody>
          </table>
        </div>
      </article>
    </section>

    <AoiDialog v-model:open="importOpen">
      <template #headline>导入</template>
      <div class="media-dialog">
        <p class="muted">格式：文件名|链接或者仅链接。</p>
        <AoiTextField
          v-model="importText"
          label="URL 列表"
          icon="link"
          multiline
          :rows="6"
          placeholder="我的图片|https://example.com/my.png&#10;https://example.com/my_1.png"
        />
      </div>
      <template #actions>
        <AoiButton appearance="plain" @click="importOpen = false">取消</AoiButton>
        <AoiButton icon="check" :disabled="!importText.trim()" :loading="importing" @click="importURLs">确定</AoiButton>
      </template>
    </AoiDialog>

    <AoiDialog v-model:open="renameOpen">
      <template #headline>编辑文件名</template>
      <div class="media-dialog">
        <AoiTextField v-model="renameValue" label="文件名/备注" icon="edit-3" @enter="renameAsset" />
      </div>
      <template #actions>
        <AoiButton appearance="plain" @click="renameOpen = false">取消</AoiButton>
        <AoiButton icon="save" :disabled="!renameValue.trim()" @click="renameAsset">保存</AoiButton>
      </template>
    </AoiDialog>

    <AoiDialog v-model:open="categoryOpen">
      <template #headline>{{ editingCategory ? "编辑分类" : "新增分类" }}</template>
      <div class="media-dialog category-dialog">
        <AoiTextField v-model="categoryName" label="分类名称" icon="folder" @enter="saveCategory" />
        <AoiSelect v-model="categoryParentId" label="上级分类" icon="folder-open" :options="categoryOptions" />
        <AoiTextField v-model="categorySort" label="排序" icon="list-filter" type="number" min="0" step="1" />
      </div>
      <template #actions>
        <AoiButton appearance="plain" @click="categoryOpen = false">取消</AoiButton>
        <AoiButton icon="save" :disabled="!categoryName.trim()" :loading="savingCategory" @click="saveCategory">保存</AoiButton>
      </template>
    </AoiDialog>
  </div>
</template>

<style scoped>
.media-workspace {
  display: grid;
  grid-template-columns: minmax(220px, 280px) minmax(0, 1fr);
  gap: 16px;
  align-items: start;
}

.media-categories {
  display: grid;
  gap: 8px;
  padding: 16px;
  border: 1px solid rgb(var(--md-sys-color-outline-variant-rgb, 210 216 224));
  border-radius: 8px;
  background: rgb(var(--md-sys-color-surface-rgb, 255 255 255));
}

.media-panel-header,
.media-assets-header,
.media-pager,
.media-toolbar,
.table-actions,
.category-actions,
.asset-link {
  display: flex;
  align-items: center;
}

.media-panel-header,
.media-assets-header {
  justify-content: space-between;
  gap: 12px;
}

.media-panel-header h2,
.media-assets-header h2 {
  margin: 0;
  font-size: 1rem;
  line-height: 1.3;
}

.media-panel-header p,
.media-assets-header p {
  margin: 4px 0 0;
}

.category-list {
  display: grid;
  gap: 4px;
}

.category-line {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: 4px;
  align-items: center;
}

.category-row {
  display: flex;
  min-height: 38px;
  align-items: center;
  gap: 8px;
  width: 100%;
  padding: 8px 12px;
  border: 0;
  border-radius: 6px;
  background: transparent;
  color: rgb(var(--md-sys-color-on-surface-rgb, 30 36 43));
  cursor: pointer;
  font: inherit;
  text-align: left;
}

.category-row span {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.category-row:hover,
.category-row--active {
  background: rgb(var(--md-sys-color-secondary-container-rgb, 222 234 236));
}

.category-actions {
  gap: 2px;
  opacity: 0;
}

.category-line:hover .category-actions,
.category-line:focus-within .category-actions {
  opacity: 1;
}

.media-assets {
  min-width: 0;
}

.media-pager {
  gap: 8px;
  white-space: nowrap;
}

.media-alert {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  margin: 0 0 14px;
  padding: 10px 12px;
  border: 1px solid rgb(245 196 88);
  border-radius: 6px;
  background: rgb(255 248 225);
  color: rgb(92 64 14);
}

.media-toolbar {
  flex-wrap: wrap;
  gap: 10px;
  margin-bottom: 14px;
}

.media-toolbar :deep(md-filled-text-field),
.media-toolbar :deep(md-outlined-text-field) {
  min-width: 180px;
}

.media-file-input {
  display: none;
}

.media-table {
  min-width: 860px;
}

.media-table th:first-child,
.media-table td:first-child {
  width: 88px;
}

.asset-preview {
  display: grid;
  width: 56px;
  height: 56px;
  place-items: center;
  overflow: hidden;
  border: 1px solid rgb(var(--md-sys-color-outline-variant-rgb, 210 216 224));
  border-radius: 6px;
  background: rgb(var(--md-sys-color-surface-container-rgb, 244 247 249));
  color: rgb(var(--md-sys-color-primary-rgb, 23 91 112));
}

.asset-preview img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.asset-name {
  display: grid;
  gap: 2px;
  border: 0;
  background: transparent;
  color: inherit;
  cursor: pointer;
  font: inherit;
  text-align: left;
}

.asset-name strong,
.asset-name small {
  max-width: 260px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.asset-name small {
  color: rgb(var(--md-sys-color-on-surface-variant-rgb, 91 97 105));
}

.asset-link {
  gap: 6px;
  color: rgb(var(--md-sys-color-primary-rgb, 23 91 112));
  text-decoration: none;
}

.asset-link--local {
  color: rgb(var(--md-sys-color-on-surface-variant-rgb, 91 97 105));
}

.table-actions {
  flex-wrap: wrap;
  gap: 6px;
}

.media-empty {
  padding: 18px 8px;
  text-align: center;
}

.media-dialog {
  display: grid;
  gap: 14px;
  min-width: min(520px, 84vw);
}

.category-dialog {
  min-width: min(420px, 84vw);
}

.mono {
  font-family: "JetBrains Mono", "SFMono-Regular", Consolas, monospace;
}

@media (max-width: 920px) {
  .media-workspace {
    grid-template-columns: 1fr;
  }

  .media-categories {
    position: static;
  }
}

@media (max-width: 720px) {
  .media-assets-header,
  .media-panel-header,
  .media-pager,
  .media-toolbar {
    align-items: stretch;
    flex-direction: column;
  }

  .media-toolbar :deep(md-filled-text-field),
  .media-toolbar :deep(md-outlined-text-field) {
    width: 100%;
    min-width: 0;
  }

  .media-table {
    min-width: 0;
  }

  .media-table thead {
    display: none;
  }

  .media-table,
  .media-table tbody,
  .media-table tr,
  .media-table td {
    display: block;
    width: 100%;
  }

  .media-table tr {
    padding: 12px 0;
    border-bottom: 1px solid rgb(var(--md-sys-color-outline-variant-rgb, 210 216 224));
  }

  .media-table tr:last-child {
    border-bottom: 0;
  }

  .media-table td,
  .media-table td:last-child {
    display: grid;
    grid-template-columns: 92px minmax(0, 1fr);
    gap: 10px;
    align-items: center;
    padding: 8px 0;
    border-bottom: 0;
  }

  .media-table td::before {
    color: rgb(var(--md-sys-color-on-surface-variant-rgb, 91 97 105));
    content: attr(data-label);
    font-size: 0.78rem;
  }

  .asset-name strong,
  .asset-name small {
    max-width: 100%;
  }

  .table-actions {
    align-items: stretch;
  }
}
</style>
