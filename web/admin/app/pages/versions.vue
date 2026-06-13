<script setup lang="ts">
import type {
  ID,
  SystemAPIEntry,
  SystemAPIGroup,
  SystemDictionary,
  SystemMenuGroup,
  SystemVersionDetail,
  SystemVersionPackage,
  SystemVersionPage,
  SystemVersionSourceCatalog
} from "~/types/admin"

const api = useAdminApi()

const pageData = ref<SystemVersionPage>({ items: [], page: 1, pageSize: 10, storageStatus: "unavailable", total: 0 })
const sources = ref<SystemVersionSourceCatalog>({
  apiCount: 0,
  apis: [],
  dictionaries: [],
  dictionaryCount: 0,
  menuCount: 0,
  menus: [],
  storageStatus: "unavailable"
})
const versionName = ref("")
const versionCode = ref("")
const startCreatedAt = ref("")
const endCreatedAt = ref("")
const page = ref(1)
const pageSize = ref("10")
const selectedIds = ref<ID[]>([])
const exportOpen = ref(false)
const importOpen = ref(false)
const detailOpen = ref(false)
const detail = ref<SystemVersionDetail | null>(null)
const formVersionName = ref("")
const formVersionCode = ref("")
const formDescription = ref("")
const selectedMenuCodes = ref<string[]>([])
const selectedAPICodes = ref<string[]>([])
const selectedDictionaryCodes = ref<string[]>([])
const menuKeyword = ref("")
const apiKeyword = ref("")
const dictionaryKeyword = ref("")
const importText = ref("")
const loading = ref(false)
const sourceLoading = ref(false)
const saving = ref(false)
const importing = ref(false)
const deleting = ref(false)
const detailLoading = ref(false)
const error = ref("")
const success = ref("")

const persisted = computed(() => pageData.value.storageStatus === "persisted")
const sourcesPersisted = computed(() => sources.value.storageStatus === "persisted")
const pageSizeNumber = computed(() => Math.min(100, Math.max(1, Number(pageSize.value) || 10)))
const totalPages = computed(() => Math.max(1, Math.ceil(pageData.value.total / pageSizeNumber.value)))
const hasSelection = computed(() => selectedIds.value.length > 0)
const detailPackage = computed(() => detail.value?.package || null)
const importPreview = computed(() => parseImportPackage(importText.value))
const canExport = computed(() => Boolean(
  formVersionName.value.trim()
  && formVersionCode.value.trim()
  && (selectedMenuCodes.value.length || selectedAPICodes.value.length || selectedDictionaryCodes.value.length)
  && !saving.value
))
const canImport = computed(() => Boolean(importPreview.value && !importing.value))

const allMenuCodes = computed(() => sources.value.menus.flatMap((group) => group.items.map((item) => menuSelector(group, item.code))))
const allAPICodes = computed(() => sources.value.apis.flatMap((group) => group.items.map(apiSelector)))
const allDictionaryCodes = computed(() => sources.value.dictionaries.map((dictionary) => dictionary.code))

const filteredMenus = computed(() => {
  const keyword = menuKeyword.value.trim().toLowerCase()
  if (!keyword) {
    return sources.value.menus
  }
  return sources.value.menus
    .map((group) => ({
      ...group,
      items: group.items.filter((item) => [group.label, group.code, item.label, item.code, item.path, item.permission]
        .some((value) => String(value || "").toLowerCase().includes(keyword)))
    }))
    .filter((group) => group.items.length)
})

const filteredAPIs = computed(() => {
  const keyword = apiKeyword.value.trim().toLowerCase()
  if (!keyword) {
    return sources.value.apis
  }
  return sources.value.apis
    .map((group) => ({
      ...group,
      items: group.items.filter((item) => [group.label, group.code, item.method, item.path, item.description, item.permission]
        .some((value) => String(value || "").toLowerCase().includes(keyword)))
    }))
    .filter((group) => group.items.length)
})

const filteredDictionaries = computed(() => {
  const keyword = dictionaryKeyword.value.trim().toLowerCase()
  if (!keyword) {
    return sources.value.dictionaries
  }
  return sources.value.dictionaries.filter((dictionary) => [
    dictionary.code,
    dictionary.name,
    dictionary.description,
    ...dictionary.items.flatMap((item) => [item.label, item.value, item.extra])
  ].some((value) => String(value || "").toLowerCase().includes(keyword)))
})

async function load(options: { silent?: boolean } = {}) {
  if (!options.silent) {
    loading.value = true
  }
  error.value = ""
  try {
    pageData.value = await api.listSystemVersions({
      endCreatedAt: toQueryDate(endCreatedAt.value),
      page: page.value,
      pageSize: pageSizeNumber.value,
      startCreatedAt: toQueryDate(startCreatedAt.value),
      versionCode: versionCode.value.trim() || undefined,
      versionName: versionName.value.trim() || undefined
    })
    selectedIds.value = selectedIds.value.filter((id) => pageData.value.items.some((item) => item.id === id))
  } catch (err) {
    error.value = errorMessage(err)
  } finally {
    if (!options.silent) {
      loading.value = false
    }
  }
}

const autoRefresh = useAdminAutoRefresh({
  blocked: computed(() => loading.value || saving.value || importing.value || deleting.value),
  load
})

async function loadSources(force = false) {
  if (sourceLoading.value || (!force && (sources.value.menuCount || sources.value.apiCount || sources.value.dictionaryCount))) {
    return
  }
  sourceLoading.value = true
  error.value = ""
  try {
    sources.value = await api.listSystemVersionSources()
  } catch (err) {
    error.value = errorMessage(err)
  } finally {
    sourceLoading.value = false
  }
}

async function openExportDialog() {
  exportOpen.value = true
  formVersionName.value = defaultVersionName()
  formVersionCode.value = defaultVersionCode()
  formDescription.value = ""
  menuKeyword.value = ""
  apiKeyword.value = ""
  dictionaryKeyword.value = ""
  await loadSources()
  selectAllSources()
}

async function openImportDialog() {
  importOpen.value = true
  importText.value = ""
  await loadSources()
}

async function exportVersion() {
  if (!canExport.value) {
    return
  }
  saving.value = true
  error.value = ""
  success.value = ""
  try {
    const created = await api.exportSystemVersion({
      apiCodes: selectedAPICodes.value,
      description: formDescription.value.trim(),
      dictionaryCodes: selectedDictionaryCodes.value,
      menuCodes: selectedMenuCodes.value,
      versionCode: formVersionCode.value.trim(),
      versionName: formVersionName.value.trim()
    })
    success.value = `版本包 ${created.item.versionName} 已创建。`
    exportOpen.value = false
    await load()
  } catch (err) {
    error.value = errorMessage(err)
  } finally {
    saving.value = false
  }
}

async function importVersion() {
  if (!canImport.value) {
    return
  }
  importing.value = true
  error.value = ""
  success.value = ""
  try {
    const result = await api.importSystemVersion(importText.value)
    success.value = `版本包 ${result.item.versionName} 已导入：字典 ${result.dictionariesCreated} 个，字典项 ${result.dictionaryItemsCreated} 个。`
    importOpen.value = false
    await load()
    await loadSources(true)
  } catch (err) {
    error.value = errorMessage(err)
  } finally {
    importing.value = false
  }
}

async function showDetail(id: ID) {
  detailOpen.value = true
  detailLoading.value = true
  error.value = ""
  try {
    detail.value = await api.getSystemVersion(id)
  } catch (err) {
    error.value = errorMessage(err)
  } finally {
    detailLoading.value = false
  }
}

async function downloadVersion(id: ID) {
  error.value = ""
  success.value = ""
  try {
    const pkg = await api.downloadSystemVersion(id)
    downloadJSON(pkg, `system-version-${pkg.version.code || id}.json`)
    success.value = "版本包已下载。"
  } catch (err) {
    error.value = errorMessage(err)
  }
}

async function deleteOne(id: ID, name: string) {
  if (!window.confirm(`删除版本包 ${name}？`)) {
    return
  }
  deleting.value = true
  error.value = ""
  success.value = ""
  try {
    await api.deleteSystemVersion(id)
    success.value = `版本包 ${name} 已删除。`
    await load()
  } catch (err) {
    error.value = errorMessage(err)
  } finally {
    deleting.value = false
  }
}

async function deleteSelected() {
  if (!hasSelection.value || !window.confirm(`删除 ${selectedIds.value.length} 个版本包？`)) {
    return
  }
  deleting.value = true
  error.value = ""
  success.value = ""
  try {
    await api.deleteSystemVersions(selectedIds.value)
    success.value = `已删除 ${selectedIds.value.length} 个版本包。`
    selectedIds.value = []
    await load()
  } catch (err) {
    error.value = errorMessage(err)
  } finally {
    deleting.value = false
  }
}

async function resetFilters() {
  versionName.value = ""
  versionCode.value = ""
  startCreatedAt.value = ""
  endCreatedAt.value = ""
  page.value = 1
  await autoRefresh.refreshNow()
}

async function previousPage() {
  if (page.value <= 1) {
    return
  }
  page.value -= 1
  await autoRefresh.refreshNow()
}

async function nextPage() {
  if (page.value >= totalPages.value) {
    return
  }
  page.value += 1
  await autoRefresh.refreshNow()
}

async function search() {
  page.value = 1
  await autoRefresh.refreshNow()
}

function toggleAllRows(event: Event) {
  const checked = (event.target as HTMLInputElement).checked
  selectedIds.value = checked ? pageData.value.items.map((item) => item.id) : []
}

function selectAllSources() {
  selectedMenuCodes.value = allMenuCodes.value
  selectedAPICodes.value = allAPICodes.value
  selectedDictionaryCodes.value = sourcesPersisted.value ? allDictionaryCodes.value : []
}

function clearSourceSelection() {
  selectedMenuCodes.value = []
  selectedAPICodes.value = []
  selectedDictionaryCodes.value = []
}

function toggleMenuGroup(group: SystemMenuGroup, event: Event) {
  const checked = (event.target as HTMLInputElement).checked
  const codes = group.items.map((item) => menuSelector(group, item.code))
  selectedMenuCodes.value = toggleMany(selectedMenuCodes.value, codes, checked)
}

function toggleAPIGroup(group: SystemAPIGroup, event: Event) {
  const checked = (event.target as HTMLInputElement).checked
  const codes = group.items.map(apiSelector)
  selectedAPICodes.value = toggleMany(selectedAPICodes.value, codes, checked)
}

function toggleMenuCode(code: string, event: Event) {
	const checked = (event.target as HTMLInputElement).checked
	selectedMenuCodes.value = toggleMany(selectedMenuCodes.value, [code], checked)
}

function toggleAPICode(code: string, event: Event) {
	const checked = (event.target as HTMLInputElement).checked
	selectedAPICodes.value = toggleMany(selectedAPICodes.value, [code], checked)
}

function toggleDictionaryCode(code: string, event: Event) {
	const checked = (event.target as HTMLInputElement).checked
	selectedDictionaryCodes.value = toggleMany(selectedDictionaryCodes.value, [code], checked)
}

function toggleMany(current: string[], codes: string[], checked: boolean) {
  const set = new Set(current)
  for (const code of codes) {
    if (checked) {
      set.add(code)
    } else {
      set.delete(code)
    }
  }
  return Array.from(set)
}

async function readImportFile(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) {
    return
  }
  importText.value = await file.text()
  input.value = ""
}

function parseImportPackage(value: string): SystemVersionPackage | null {
  const raw = value.trim()
  if (!raw) {
    return null
  }
  try {
    const parsed = JSON.parse(raw) as SystemVersionPackage
    if (!parsed?.version?.name || !parsed.version.code) {
      return null
    }
    return parsed
  } catch {
    return null
  }
}

function menuSelector(group: SystemMenuGroup, itemCode: string) {
  return `${group.code}:${itemCode}`
}

function apiSelector(item: SystemAPIEntry) {
  return item.code || `${item.method} ${item.path}`.toLowerCase()
}

function selectedCountInGroup(codes: string[], groupCodes: string[]) {
  const selected = new Set(codes)
  return groupCodes.filter((code) => selected.has(code)).length
}

function sourceLabel(source: string) {
  return source === "import" ? "导入" : "导出"
}

function sourceBadgeClass(source: string) {
  return source === "import" ? "badge--warning" : "badge--success"
}

function defaultVersionName() {
  const now = new Date()
  return `Release ${now.getFullYear()}-${pad(now.getMonth() + 1)}-${pad(now.getDate())}`
}

function defaultVersionCode() {
  const now = new Date()
  return `v${now.getFullYear()}.${pad(now.getMonth() + 1)}.${pad(now.getDate())}.${pad(now.getHours())}${pad(now.getMinutes())}`
}

function pad(value: number) {
  return String(value).padStart(2, "0")
}

function toQueryDate(value: string) {
  return value || undefined
}

function downloadJSON(value: unknown, filename: string) {
  if (!import.meta.client) {
    return
  }
  const blob = new Blob([JSON.stringify(value, null, 2)], { type: "application/json;charset=utf-8" })
  const url = URL.createObjectURL(blob)
  const link = document.createElement("a")
  link.href = url
  link.download = filename
  link.click()
  URL.revokeObjectURL(url)
}

onMounted(async () => {
  await Promise.all([
    autoRefresh.refreshNow(),
    loadSources()
  ])
})

watch(pageSize, async () => {
  page.value = 1
  await autoRefresh.refreshNow()
})

useHead({
  title: "版本管理 - Aoi Admin"
})
</script>

<template>
  <div class="page-grid">
    <PageHeader title="版本管理" icon="package-check">
      <template #actions>
        <AoiButton appearance="soft" icon="package-plus" :disabled="!persisted" @click="openExportDialog">创建发版</AoiButton>
        <AoiButton appearance="soft" icon="upload" :disabled="!persisted" @click="openImportDialog">导入版本</AoiButton>
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
    <AoiStatusMessage v-if="!persisted" tone="warning" message="system_versions 表尚不可用，请先执行数据库迁移。" />

    <article class="admin-card">
      <div class="admin-card__header">
        <div>
          <h2>版本列表</h2>
          <p class="muted">共 {{ pageData.total }} 个</p>
        </div>
        <div class="version-pager">
          <AoiButton appearance="soft" size="sm" icon="chevron-left" :disabled="page <= 1" @click="previousPage">上一页</AoiButton>
          <span class="badge">{{ page }} / {{ totalPages }}</span>
          <AoiButton appearance="soft" size="sm" icon="chevron-right" :disabled="page >= totalPages" @click="nextPage">下一页</AoiButton>
        </div>
      </div>

      <div class="admin-filter-toolbar version-filter-toolbar">
        <AoiTextField v-model="versionName" label="版本名称" icon="search" placeholder="Release" @enter="search" />
        <AoiTextField v-model="versionCode" label="版本号" icon="hash" placeholder="v2026.06" @enter="search" />
        <AoiTextField v-model="startCreatedAt" label="开始日期" icon="calendar" type="date" />
        <AoiTextField v-model="endCreatedAt" label="结束日期" icon="calendar" type="date" />
        <AoiTextField v-model="pageSize" label="每页" icon="list-filter" type="number" min="1" max="100" step="1" @enter="search" />
        <AoiButton appearance="soft" icon="search" :loading="loading" @click="search">查询</AoiButton>
        <AoiButton appearance="plain" icon="rotate-ccw" @click="resetFilters">重置</AoiButton>
        <AoiButton appearance="soft" intent="danger" icon="trash-2" :disabled="!hasSelection || deleting" :loading="deleting" @click="deleteSelected">删除</AoiButton>
      </div>

      <div class="data-table-wrap">
        <table class="data-table version-table">
          <thead>
            <tr>
              <th>
                <input
                  aria-label="选择全部版本"
                  class="version-check"
                  type="checkbox"
                  :checked="Boolean(pageData.items.length) && selectedIds.length === pageData.items.length"
                  :disabled="!pageData.items.length"
                  @change="toggleAllRows"
                >
              </th>
              <th>创建时间</th>
              <th>版本</th>
              <th>来源</th>
              <th>菜单</th>
              <th>API</th>
              <th>字典</th>
              <th>创建人</th>
              <th>说明</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="item in pageData.items" :key="item.id">
              <td data-label="选择">
                <input v-model="selectedIds" class="version-check" type="checkbox" :value="item.id" :aria-label="`选择版本 ${item.id}`">
              </td>
              <td data-label="创建时间">{{ formatDateTime(item.createdAt) }}</td>
              <td data-label="版本">
                <button class="version-name" type="button" @click="showDetail(item.id)">
                  <strong>{{ item.versionName }}</strong>
                  <small>{{ item.versionCode }}</small>
                </button>
              </td>
              <td data-label="来源"><span class="badge" :class="sourceBadgeClass(item.source)">{{ sourceLabel(item.source) }}</span></td>
              <td class="mono" data-label="菜单">{{ item.menuCount }}</td>
              <td class="mono" data-label="API">{{ item.apiCount }}</td>
              <td class="mono" data-label="字典">{{ item.dictionaryCount }}</td>
              <td data-label="创建人">{{ item.createdByUsername || item.createdBy }}</td>
              <td data-label="说明">{{ item.description || "-" }}</td>
              <td data-label="操作">
                <div class="table-actions">
                  <AoiButton appearance="soft" size="sm" icon="eye" @click="showDetail(item.id)">查看</AoiButton>
                  <AoiButton appearance="soft" size="sm" icon="download" @click="downloadVersion(item.id)">下载</AoiButton>
                  <AoiButton appearance="soft" intent="danger" size="sm" icon="trash-2" :disabled="deleting" @click="deleteOne(item.id, item.versionName)">删除</AoiButton>
                </div>
              </td>
            </tr>
            <tr v-if="!pageData.items.length">
              <td colspan="10" class="muted">暂无版本包。</td>
            </tr>
          </tbody>
        </table>
      </div>
    </article>

    <AoiDialog v-model:open="exportOpen" class="version-export-shell">
      <template #headline>创建发版</template>
      <div class="version-dialog version-export-dialog">
        <div class="version-form-grid">
          <AoiTextField v-model="formVersionName" label="版本名称" icon="package-check" />
          <AoiTextField v-model="formVersionCode" label="版本号" icon="hash" />
          <AoiTextField v-model="formDescription" label="说明" icon="file-text" multiline :rows="3" />
        </div>

        <div class="version-source-actions">
          <span class="badge">{{ selectedMenuCodes.length }} 菜单</span>
          <span class="badge">{{ selectedAPICodes.length }} API</span>
          <span class="badge">{{ selectedDictionaryCodes.length }} 字典</span>
          <AoiButton appearance="soft" size="sm" icon="check-check" :disabled="sourceLoading" @click="selectAllSources">全选</AoiButton>
          <AoiButton appearance="plain" size="sm" icon="x" @click="clearSourceSelection">清空</AoiButton>
        </div>

        <div class="version-source-grid">
          <section class="version-source-panel">
            <div class="version-source-panel__header">
              <h3>菜单</h3>
              <AoiTextField v-model="menuKeyword" label="筛选" icon="search" />
            </div>
            <div class="version-source-list">
              <div v-for="group in filteredMenus" :key="group.code" class="version-source-group">
                <label class="version-source-group__title">
                  <input
                    class="version-check"
                    type="checkbox"
                    :checked="selectedCountInGroup(selectedMenuCodes, group.items.map((item) => menuSelector(group, item.code))) === group.items.length"
                    @change="toggleMenuGroup(group, $event)"
                  >
                  <span>{{ group.label }}</span>
                  <small>{{ selectedCountInGroup(selectedMenuCodes, group.items.map((item) => menuSelector(group, item.code))) }}/{{ group.items.length }}</small>
                </label>
                <label v-for="item in group.items" :key="item.code" class="version-source-option">
                  <input
                    class="version-check"
                    type="checkbox"
                    :checked="selectedMenuCodes.includes(menuSelector(group, item.code))"
                    @change="toggleMenuCode(menuSelector(group, item.code), $event)"
                  >
                  <span>{{ item.label }}</span>
                  <code>{{ item.path }}</code>
                </label>
              </div>
            </div>
          </section>

          <section class="version-source-panel">
            <div class="version-source-panel__header">
              <h3>API</h3>
              <AoiTextField v-model="apiKeyword" label="筛选" icon="search" />
            </div>
            <div class="version-source-list">
              <div v-for="group in filteredAPIs" :key="group.code" class="version-source-group">
                <label class="version-source-group__title">
                  <input
                    class="version-check"
                    type="checkbox"
                    :checked="selectedCountInGroup(selectedAPICodes, group.items.map(apiSelector)) === group.items.length"
                    @change="toggleAPIGroup(group, $event)"
                  >
                  <span>{{ group.label }}</span>
                  <small>{{ selectedCountInGroup(selectedAPICodes, group.items.map(apiSelector)) }}/{{ group.items.length }}</small>
                </label>
                <label v-for="item in group.items" :key="apiSelector(item)" class="version-source-option">
                  <input
                    class="version-check"
                    type="checkbox"
                    :checked="selectedAPICodes.includes(apiSelector(item))"
                    @change="toggleAPICode(apiSelector(item), $event)"
                  >
                  <span class="method-badge">{{ item.method }}</span>
                  <code>{{ item.path }}</code>
                </label>
              </div>
            </div>
          </section>

          <section class="version-source-panel">
            <div class="version-source-panel__header">
              <h3>字典</h3>
              <AoiTextField v-model="dictionaryKeyword" label="筛选" icon="search" />
            </div>
            <AoiStatusMessage v-if="!sourcesPersisted" tone="warning" message="字典表不可用，当前只能打包菜单和 API。" />
            <div class="version-source-list">
              <label v-for="dictionary in filteredDictionaries" :key="dictionary.code" class="version-source-option version-source-option--dictionary">
                <input
                  class="version-check"
                  type="checkbox"
                  :checked="selectedDictionaryCodes.includes(dictionary.code)"
                  :disabled="!sourcesPersisted"
                  @change="toggleDictionaryCode(dictionary.code, $event)"
                >
                <span>{{ dictionary.name }}</span>
                <code>{{ dictionary.code }}</code>
                <small>{{ dictionary.items.length }} 项</small>
              </label>
            </div>
          </section>
        </div>
      </div>
      <template #actions>
        <AoiButton appearance="plain" :disabled="saving" @click="exportOpen = false">取消</AoiButton>
        <AoiButton icon="package-plus" :loading="saving" :disabled="!canExport" @click="exportVersion">创建发版</AoiButton>
      </template>
    </AoiDialog>

    <AoiDialog v-model:open="importOpen">
      <template #headline>导入版本</template>
      <div class="version-dialog version-import-dialog">
        <label class="version-file">
          <AoiIcon name="upload" :size="20" decorative />
          <span>选择 JSON 文件</span>
          <input type="file" accept="application/json,.json" @change="readImportFile">
        </label>
        <AoiTextField v-model="importText" label="版本 JSON" icon="braces" multiline :rows="12" />
        <section v-if="importPreview" class="version-preview">
          <div>
            <h3>{{ importPreview.version.name }}</h3>
            <p class="mono">{{ importPreview.version.code }}</p>
          </div>
          <span class="badge">{{ importPreview.menus.reduce((count, group) => count + group.items.length, 0) }} 菜单</span>
          <span class="badge">{{ importPreview.apis.length }} API</span>
          <span class="badge">{{ importPreview.dictionaries.length }} 字典</span>
        </section>
        <AoiStatusMessage v-else-if="importText.trim()" tone="danger" message="JSON 无法识别为版本包。" />
      </div>
      <template #actions>
        <AoiButton appearance="plain" :disabled="importing" @click="importOpen = false">取消</AoiButton>
        <AoiButton icon="upload" :loading="importing" :disabled="!canImport" @click="importVersion">导入</AoiButton>
      </template>
    </AoiDialog>

    <AoiDialog v-model:open="detailOpen">
      <template #headline>版本详情</template>
      <div class="version-dialog version-detail-dialog">
        <AoiStatusMessage v-if="detailLoading" tone="info" message="正在加载版本包。" />
        <template v-else-if="detail && detailPackage">
          <section class="version-detail-head">
            <div>
              <h3>{{ detail.item.versionName }}</h3>
              <p class="mono">{{ detail.item.versionCode }}</p>
            </div>
            <span class="badge" :class="sourceBadgeClass(detail.item.source)">{{ sourceLabel(detail.item.source) }}</span>
          </section>
          <div class="version-detail-stats">
            <span class="badge">{{ detail.item.menuCount }} 菜单</span>
            <span class="badge">{{ detail.item.apiCount }} API</span>
            <span class="badge">{{ detail.item.dictionaryCount }} 字典</span>
            <span class="badge">{{ formatDateTime(detail.item.createdAt) }}</span>
          </div>
          <section class="version-json-preview">
            <pre>{{ JSON.stringify(detailPackage, null, 2) }}</pre>
          </section>
        </template>
      </div>
      <template #actions>
        <AoiButton appearance="plain" @click="detailOpen = false">关闭</AoiButton>
        <AoiButton v-if="detail" icon="download" @click="downloadVersion(detail.item.id)">下载</AoiButton>
      </template>
    </AoiDialog>
  </div>
</template>

<style scoped>
.version-pager,
.table-actions,
.version-source-actions,
.version-detail-stats {
  align-items: center;
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.version-filter-toolbar {
  grid-template-columns:
    minmax(150px, 0.9fr)
    minmax(128px, 0.75fr)
    minmax(132px, 0.65fr)
    minmax(132px, 0.65fr)
    minmax(118px, 0.48fr)
    96px
    88px
    88px;
}

.version-check {
  accent-color: var(--aoi-accent-60);
  block-size: 16px;
  inline-size: 16px;
}

.version-table {
  min-width: 1180px;
}

.version-table th:first-child,
.version-table td:first-child {
  width: 42px;
}

.version-name {
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

.version-name strong,
.version-name small {
  min-width: 0;
  overflow-wrap: anywhere;
}

.version-name small {
  color: var(--aoi-text-muted);
  font-size: 12px;
}

.version-dialog {
  display: grid;
  gap: 16px;
  max-width: 1120px;
  width: 100%;
}

:deep(md-dialog.version-export-shell) {
  height: min(88dvh, 820px);
  max-height: calc(100dvh - 40px);
  max-width: min(96vw, 1120px);
  width: min(96vw, 1120px);
}

:deep(md-dialog.version-export-shell [slot="content"]) {
  max-height: calc(100dvh - 230px);
  overflow: auto;
  overscroll-behavior: contain;
  width: min(92vw, 1040px);
}

.version-form-grid {
  display: grid;
  gap: 12px;
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.version-form-grid > :last-child {
  grid-column: 1 / -1;
}

.version-source-grid {
  display: grid;
  gap: 12px;
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.version-source-panel {
  border: 1px solid var(--aoi-border-subtle);
  border-radius: 8px;
  display: grid;
  gap: 10px;
  min-height: 420px;
  padding: 12px;
}

.version-source-panel__header {
  display: grid;
  gap: 10px;
}

.version-source-panel h3,
.version-preview h3,
.version-detail-head h3 {
  font-size: 15px;
  margin: 0;
}

.version-source-list {
  display: grid;
  gap: 10px;
  max-height: 430px;
  overflow: auto;
  padding-right: 2px;
}

.version-source-group {
  display: grid;
  gap: 6px;
}

.version-source-group__title,
.version-source-option {
  align-items: center;
  border-radius: 8px;
  display: grid;
  gap: 8px;
  grid-template-columns: 18px minmax(0, 1fr) auto;
}

.version-source-group__title {
  background: var(--aoi-surface-soft);
  border: 1px solid var(--aoi-border-subtle);
  font-weight: 700;
  padding: 8px;
}

.version-source-group__title small,
.version-source-option small {
  color: var(--aoi-text-muted);
  font-size: 12px;
}

.version-source-option {
  border: 1px solid transparent;
  padding: 7px 8px;
}

.version-source-option:hover {
  background: var(--aoi-surface-soft);
  border-color: var(--aoi-border-subtle);
}

.version-source-option span,
.version-source-option code {
  min-width: 0;
  overflow-wrap: anywhere;
}

.version-source-option code {
  color: var(--aoi-text-muted);
  font-size: 12px;
}

.version-source-option--dictionary {
  grid-template-columns: 18px minmax(0, 1fr) minmax(0, 0.8fr) auto;
}

.method-badge {
  background: var(--aoi-accent-10);
  border: 1px solid var(--aoi-border-subtle);
  border-radius: 6px;
  color: var(--aoi-active-color);
  font-family: ui-monospace, SFMono-Regular, Consolas, monospace;
  font-size: 11px;
  font-weight: 800;
  padding: 3px 5px;
}

.version-file {
  align-items: center;
  border: 1px dashed var(--aoi-border);
  border-radius: 8px;
  color: var(--aoi-text);
  cursor: pointer;
  display: flex;
  gap: 10px;
  justify-content: center;
  min-height: 86px;
  padding: 18px;
}

.version-file input {
  display: none;
}

.version-preview,
.version-detail-head {
  align-items: center;
  border: 1px solid var(--aoi-border-subtle);
  border-radius: 8px;
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  justify-content: space-between;
  padding: 12px;
}

.version-preview p,
.version-detail-head p {
  color: var(--aoi-text-muted);
  margin: 3px 0 0;
}

.version-json-preview {
  background: var(--aoi-surface-soft);
  border: 1px solid var(--aoi-border-subtle);
  border-radius: 8px;
  max-height: 460px;
  overflow: auto;
  padding: 12px;
}

.version-json-preview pre {
  font-size: 12px;
  margin: 0;
  white-space: pre-wrap;
  word-break: break-word;
}

@media (max-width: 980px) {
  .version-source-grid,
  .version-form-grid {
    grid-template-columns: 1fr;
  }

  .version-dialog {
    width: min(720px, 84vw);
  }
}

@media (max-width: 760px) {
  .versions-page {
    padding-bottom: 88px;
  }

  :deep(md-dialog.version-export-shell) {
    height: calc(100dvh - 24px);
    max-height: calc(100dvh - 24px);
    width: calc(100vw - 20px);
  }

  :deep(md-dialog.version-export-shell [slot="content"]) {
    max-height: calc(100dvh - 126px);
    width: calc(100vw - 44px);
  }

  .admin-card__header {
    align-items: flex-start;
    flex-direction: column;
  }

  .version-filter-toolbar {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }

  .version-filter-toolbar > :nth-child(-n+5) {
    grid-column: 1 / -1;
  }

  .version-table {
    min-width: 0;
    width: 100%;
  }

  .version-table thead {
    display: none;
  }

  .version-table,
  .version-table tbody,
  .version-table tr,
  .version-table td {
    display: block;
  }

  .version-table tr {
    border-bottom: 1px solid var(--aoi-border);
    padding: 12px 0;
    width: 100%;
  }

  .version-table tr:last-child {
    border-bottom: 0;
  }

  .version-table td,
  .version-table td:last-child {
    align-items: flex-start;
    border-bottom: 0;
    display: grid;
    gap: 8px;
    grid-template-columns: 78px minmax(0, 1fr);
    padding: 5px 0;
    white-space: normal;
    width: 100%;
  }

  .version-table td::before {
    color: var(--aoi-text-muted);
    content: attr(data-label);
    font-size: 12px;
    font-weight: 700;
  }

  .version-table td > * {
    max-width: 100%;
    min-width: 0;
  }

  .version-source-panel {
    min-height: 0;
  }

  .version-source-list {
    max-height: 320px;
  }

  .version-dialog {
    width: 82vw;
  }
}
</style>
