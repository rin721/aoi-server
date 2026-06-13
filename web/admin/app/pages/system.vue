<script setup lang="ts">
import type { SystemConfigItem, SystemConfigSection, SystemConfigSnapshot } from "~/types/admin"
import type { AoiIntent, AoiStatItem } from "~/types/ui"

const api = useAdminApi()
const snapshot = ref<SystemConfigSnapshot>({ sections: [] })
const selectedCode = ref("")
const editingKey = ref("")
const draftValue = ref("")
const draftBoolean = ref(false)
const loading = ref(false)
const saving = ref(false)
const error = ref("")
const success = ref("")

const sections = computed<SystemConfigSection[]>(() => [...snapshot.value.sections].sort((left, right) => left.order - right.order || left.code.localeCompare(right.code)))
const selectedSection = computed(() => sections.value.find((section) => section.code === selectedCode.value) || sections.value[0] || null)
const editingItem = computed(() => selectedSection.value?.items.find((item) => item.key === editingKey.value) || null)
const itemCount = computed(() => sections.value.reduce((count, section) => count + section.items.length, 0))
const editableCount = computed(() => sections.value.reduce((count, section) => count + section.items.filter((item) => item.editable).length, 0))
const secretCount = computed(() => sections.value.reduce((count, section) => count + section.items.filter((item) => item.secret).length, 0))
const autoRefreshBlocked = computed(() => loading.value || saving.value || Boolean(editingKey.value))
const canSave = computed(() => {
  const item = editingItem.value
  if (!item || saving.value) {
    return false
  }
  if (item.valueType === "boolean") {
    return true
  }
  if (item.valueType === "number") {
    return Number.isInteger(Number(draftValue.value))
  }
  if (item.valueType === "array") {
    return true
  }
  return true
})
const editorSupportingText = computed(() => {
  const item = editingItem.value
  if (!item) {
    return ""
  }
  if (item.secret) {
    return "敏感字段不会回显原值；留空表示不修改。"
  }
  if (item.valueType === "number") {
    return "请输入整数；保存后会先经过后端配置校验。"
  }
  if (item.valueType === "array") {
    return "每行一个值；保存后会写入当前配置文件。"
  }
  return "保存会写入当前配置文件；由环境变量管理的配置项不可在此保存。"
})
const overviewItems = computed<AoiStatItem[]>(() => [
  { icon: "blocks", label: "分区", value: sections.value.length },
  { icon: "list-tree", label: "字段", value: itemCount.value },
  { icon: "edit-3", intent: editableCount.value ? "success" : "neutral", label: "可编辑", value: editableCount.value },
  { icon: "shield-check", intent: secretCount.value ? "warning" : "neutral", label: "敏感", value: secretCount.value }
])

async function load(options: { silent?: boolean } = {}) {
  if (!options.silent) {
    loading.value = true
  }
  error.value = ""
  try {
    snapshot.value = await api.getSystemConfig()
    if (!sections.value.some((section) => section.code === selectedCode.value)) {
      selectedCode.value = sections.value[0]?.code || ""
    }
    if (editingKey.value && !selectedSection.value?.items.some((item) => item.key === editingKey.value)) {
      cancelEdit()
    }
  } catch (err) {
    error.value = errorMessage(err)
  } finally {
    if (!options.silent) {
      loading.value = false
    }
  }
}

const autoRefresh = useAdminAutoRefresh({ blocked: autoRefreshBlocked, load })

function selectSection(code: string) {
  selectedCode.value = code
  cancelEdit()
}

function startEdit(item: SystemConfigItem) {
  if (!item.editable) {
    return
  }
  error.value = ""
  success.value = ""
  editingKey.value = item.key
  draftBoolean.value = Boolean(item.value)
  draftValue.value = item.secret ? "" : formatDraftValue(item.value)
}

function cancelEdit() {
  editingKey.value = ""
  draftValue.value = ""
  draftBoolean.value = false
}

async function saveConfigItem() {
  const item = editingItem.value
  if (!item || !canSave.value) {
    return
  }
  const value = configDraftPayload(item)
  saving.value = true
  error.value = ""
  success.value = ""
  try {
    snapshot.value = await api.updateSystemConfig([{ key: item.key, value }], { persist: true })
    success.value = `配置 ${item.label} 已更新。`
    cancelEdit()
  } catch (err) {
    error.value = errorMessage(err)
  } finally {
    saving.value = false
  }
}

function configDraftPayload(item: SystemConfigItem) {
  if (item.valueType === "boolean") {
    return draftBoolean.value
  }
  if (item.valueType === "number") {
    return Number(draftValue.value)
  }
  if (item.valueType === "array") {
    return draftValue.value
      .split(/\r?\n/)
      .map((value) => value.trim())
      .filter(Boolean)
  }
  return draftValue.value
}

function itemBadge(item: SystemConfigItem) {
  if (item.secret) {
    return "敏感"
  }
  if (!item.editable) {
    return "只读"
  }
  return "可编辑"
}

function itemBadgeIntent(item: SystemConfigItem): AoiIntent {
  if (item.secret) {
    return "warning"
  }
  if (!item.editable) {
    return "neutral"
  }
  return "success"
}

function formatDraftValue(value: unknown) {
  if (value === undefined || value === null) {
    return ""
  }
  if (Array.isArray(value)) {
    return value.join("\n")
  }
  if (typeof value === "object") {
    return JSON.stringify(value)
  }
  return String(value)
}

function formatConfigValue(value: unknown) {
  if (value === undefined || value === null || value === "") {
    return "-"
  }
  if (typeof value === "boolean") {
    return value ? "是" : "否"
  }
  if (Array.isArray(value)) {
    return value.length ? value.join(", ") : "-"
  }
  if (typeof value === "object") {
    return JSON.stringify(value)
  }
  return String(value)
}

onMounted(autoRefresh.refreshNow)

useHead({
  title: "系统配置 - Aoi Admin"
})
</script>

<template>
  <div class="page-grid">
    <PageHeader title="系统配置" icon="settings" description="来自后端配置管理器的当前快照；保存会写入当前配置文件。">
      <template #actions>
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
    <AoiStatusMessage tone="warning" message="保存会写入当前进程使用的配置文件；环境变量或 ${ENV:default} 占位管理的配置项不可在此保存。" />

    <AoiStatGrid :items="overviewItems" :columns="4" />

    <section v-if="sections.length" class="config-layout">
      <nav class="config-tabs" aria-label="配置分区">
        <button
          v-for="section in sections"
          :key="section.code"
          class="config-tab"
          :class="{ 'config-tab--active': selectedSection?.code === section.code }"
          type="button"
          @click="selectSection(section.code)"
        >
          <AoiIcon :name="section.icon || 'settings'" decorative />
          <span>{{ section.label }}</span>
          <small>{{ section.items.length }}</small>
        </button>
      </nav>

      <AoiAdminCard
        v-if="selectedSection"
        class="config-panel"
        :badge="selectedSection.code"
        :description="selectedSection.description"
        :icon="selectedSection.icon || 'settings'"
        :title="selectedSection.label"
      >
        <div class="config-items">
          <div
            v-for="item in selectedSection.items"
            :key="item.key"
            class="config-item"
            :class="{ 'config-item--editing': editingKey === item.key, 'config-item--readonly': !item.editable }"
          >
            <button class="config-item__main" type="button" :disabled="!item.editable" @click="startEdit(item)">
              <span class="config-item__copy">
                <strong>{{ item.label }}</strong>
                <small>{{ item.key }}</small>
                <em v-if="item.description">{{ item.description }}</em>
              </span>
              <span class="config-item__value" :class="{ 'config-item__value--mono': item.valueType !== 'boolean' }">
                {{ formatConfigValue(item.value) }}
              </span>
            </button>
            <AoiMetaPill appearance="outline" :intent="itemBadgeIntent(item)">
              {{ itemBadge(item) }}
            </AoiMetaPill>
            <AoiButton
              v-if="item.editable"
              appearance="soft"
              size="sm"
              icon="edit-3"
              :aria-pressed="editingKey === item.key"
              @click="startEdit(item)"
            >
              编辑
            </AoiButton>
          </div>
        </div>

        <form v-if="editingItem" class="config-editor" @submit.prevent="saveConfigItem">
          <div class="config-editor__header">
            <div>
              <h3>编辑 {{ editingItem.label }}</h3>
              <p>{{ editingItem.key }}</p>
            </div>
            <AoiMetaPill :intent="editingItem.secret ? 'warning' : 'info'" appearance="outline">
              {{ editingItem.valueType }}
            </AoiMetaPill>
          </div>

          <AoiSwitch
            v-if="editingItem.valueType === 'boolean'"
            v-model="draftBoolean"
            label="配置值"
          />
          <AoiTextField
            v-else
            v-model="draftValue"
            :label="editingItem.secret ? '新配置值' : '配置值'"
            icon="settings-2"
            :type="editingItem.secret ? 'password' : editingItem.valueType === 'number' ? 'number' : 'text'"
            :step="editingItem.valueType === 'number' ? 1 : undefined"
            :multiline="editingItem.valueType === 'array'"
            :rows="editingItem.valueType === 'array' ? 5 : undefined"
            :supporting-text="editorSupportingText"
          />

          <div class="config-editor__actions">
            <AoiButton type="submit" icon="save" :loading="saving" :disabled="!canSave">保存</AoiButton>
            <AoiButton appearance="plain" icon="x" type="button" :disabled="saving" @click="cancelEdit">取消</AoiButton>
          </div>
        </form>
      </AoiAdminCard>
    </section>

    <AoiAdminCard v-else class="config-empty" padding="lg">
      <AoiIcon name="settings" :size="24" decorative />
      <p>{{ loading ? "配置加载中。" : "暂无配置快照。" }}</p>
    </AoiAdminCard>
  </div>
</template>

<style scoped>
.config-layout {
  align-items: start;
  display: grid;
  gap: var(--aoi-admin-panel-gap);
  grid-template-columns: minmax(var(--aoi-admin-config-nav-min-width), var(--aoi-admin-config-nav-width)) minmax(0, 1fr);
}

.config-tabs {
  background: var(--aoi-surface-solid);
  border: 1px solid var(--aoi-border);
  border-radius: var(--aoi-radius-card);
  display: grid;
  overflow: hidden;
  scrollbar-width: none;
}

.config-tabs::-webkit-scrollbar {
  display: none;
}

.config-tab {
  align-items: center;
  background: transparent;
  border: 0;
  border-bottom: 1px solid var(--aoi-border);
  color: var(--aoi-text);
  cursor: pointer;
  display: grid;
  font: inherit;
  gap: var(--aoi-admin-card-gap);
  grid-template-columns: var(--aoi-nav-icon-size) minmax(0, 1fr) auto;
  min-height: var(--aoi-admin-nav-row-height);
  padding: var(--aoi-admin-nav-row-padding);
  text-align: left;
}

.config-tab:last-child {
  border-bottom: 0;
}

.config-tab:hover,
.config-tab:focus-visible,
.config-tab--active {
  background: var(--aoi-state-active);
  color: var(--aoi-active-color);
}

.config-tab span {
  font-weight: 760;
  min-width: 0;
  overflow-wrap: anywhere;
}

.config-tab small {
  background: var(--aoi-surface-muted);
  border: 1px solid var(--aoi-border);
  border-radius: var(--aoi-radius-round);
  color: var(--aoi-text-muted);
  font-size: 11px;
  font-weight: 800;
  min-width: 24px;
  padding: 2px 6px;
  text-align: center;
}

.config-panel,
.config-items {
  min-width: 0;
}

.config-items {
  display: grid;
  gap: var(--aoi-admin-card-gap);
}

.config-item {
  align-items: center;
  border: 1px solid var(--aoi-border);
  border-radius: var(--aoi-radius-control);
  display: grid;
  gap: var(--aoi-admin-card-gap);
  grid-template-columns: minmax(0, 1fr) auto auto;
  min-width: 0;
  padding: 10px;
}

.config-item--editing {
  border-color: var(--aoi-active-color);
  background: var(--aoi-state-active);
}

.config-item__main {
  align-items: center;
  background: transparent;
  border: 0;
  color: inherit;
  cursor: pointer;
  display: grid;
  gap: var(--aoi-admin-card-gap);
  grid-template-columns: minmax(0, 1fr) minmax(160px, 38%);
  min-width: 0;
  padding: 0;
  text-align: left;
}

.config-item__main:disabled {
  cursor: default;
}

.config-item__copy {
  display: grid;
  gap: 4px;
  min-width: 0;
}

.config-item__copy strong,
.config-editor h3 {
  color: var(--aoi-text);
  line-height: 1.35;
}

.config-item__copy small,
.config-item__copy em,
.config-editor p {
  color: var(--aoi-text-muted);
  font-style: normal;
  line-height: 1.5;
  overflow-wrap: anywhere;
}

.config-item__value {
  color: var(--aoi-text);
  min-width: 0;
  overflow-wrap: anywhere;
}

.config-item__value--mono {
  font-family: var(--aoi-font-mono);
  font-size: 13px;
}

.config-editor {
  border-top: 1px solid var(--aoi-border);
  display: grid;
  gap: var(--aoi-admin-card-gap);
  margin-top: var(--aoi-admin-panel-gap);
  padding-top: var(--aoi-admin-panel-gap);
}

.config-editor__header,
.config-editor__actions {
  align-items: flex-start;
  display: flex;
  flex-wrap: wrap;
  gap: var(--aoi-admin-card-gap);
  justify-content: space-between;
  min-width: 0;
}

.config-editor h3,
.config-editor p {
  margin: 0;
}

.config-empty {
  align-items: center;
  color: var(--aoi-text-muted);
  display: flex;
  gap: var(--aoi-admin-card-gap);
  justify-content: center;
  min-height: 160px;
}

.config-empty p {
  margin: 0;
}

@media (max-width: 920px) {
  .config-layout {
    grid-template-columns: 1fr;
  }

  .config-tabs {
    display: flex;
    overflow-x: auto;
  }

  .config-tab {
    border-bottom: 0;
    border-right: 1px solid var(--aoi-border);
    flex: 0 0 auto;
    min-width: var(--aoi-admin-config-tab-min-width);
  }
}

@media (max-width: 680px) {
  .config-item,
  .config-item__main {
    grid-template-columns: 1fr;
  }

  .config-editor__actions {
    justify-content: flex-start;
  }
}
</style>
