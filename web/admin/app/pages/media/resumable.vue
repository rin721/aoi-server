<script setup lang="ts">
import type { SystemMediaAsset, SystemMediaAssetPage, SystemMediaResumableCheckResult } from "~/types/admin"

const api = useAdminApi()

const fileInput = ref<HTMLInputElement | null>(null)
const selectedFile = ref<File | null>(null)
const fileHash = ref("")
const checkResult = ref<SystemMediaResumableCheckResult | null>(null)
const lastAsset = ref<SystemMediaAsset | null>(null)
const loadingStatus = ref(false)
const hashing = ref(false)
const uploading = ref(false)
const aborting = ref(false)
const status = ref<"aborted" | "completed" | "error" | "hashing" | "idle" | "ready" | "uploading">("idle")
const error = ref("")
const success = ref("")
const uploadMaxBytes = ref(20 * 1024 * 1024)
const uploadMaxMb = ref(20)
const uploadUnavailable = ref(false)

const chunkSize = 1024 * 1024

const progress = computed(() => {
  if (status.value === "completed") {
    return 100
  }
  return Math.max(0, Math.min(100, checkResult.value?.progress || 0))
})
const selectedFileName = computed(() => selectedFile.value?.name || "未选择文件")
const selectedFileSize = computed(() => selectedFile.value ? formatBytes(selectedFile.value.size) : "-")
const chunkTotal = computed(() => selectedFile.value ? Math.ceil(selectedFile.value.size / chunkSize) : 0)
const missingCount = computed(() => checkResult.value?.missingChunks.length ?? chunkTotal.value)
const uploadedCount = computed(() => checkResult.value?.uploadedChunks.length ?? 0)
const canPickFile = computed(() => !hashing.value && !uploading.value)
const canUpload = computed(() => Boolean(selectedFile.value && checkResult.value && !uploading.value && !hashing.value && !uploadUnavailable.value && status.value !== "completed"))
const canAbort = computed(() => Boolean(checkResult.value?.session.id && !uploading.value && status.value !== "completed" && status.value !== "aborted"))
const statusLabel = computed(() => {
  switch (status.value) {
    case "aborted":
      return "已中止"
    case "completed":
      return "已完成"
    case "error":
      return "异常"
    case "hashing":
      return "校验中"
    case "ready":
      return "待上传"
    case "uploading":
      return "上传中"
    default:
      return "待选择"
  }
})

async function loadMediaStatus(options: { silent?: boolean } = {}) {
  if (!options.silent) {
    loadingStatus.value = true
  }
  try {
    const page = await api.listSystemMediaAssets({ page: 1, pageSize: 1 }) as SystemMediaAssetPage
    uploadMaxBytes.value = page.uploadMaxBytes
    uploadMaxMb.value = page.uploadMaxMb
    uploadUnavailable.value = page.uploadUnavailable
  } catch (err) {
    error.value = errorMessage(err)
  } finally {
    if (!options.silent) {
      loadingStatus.value = false
    }
  }
}

const autoRefresh = useAdminAutoRefresh({
  blocked: computed(() => loadingStatus.value || hashing.value || uploading.value || aborting.value),
  load: loadMediaStatus
})

function openFilePicker() {
  if (canPickFile.value) {
    fileInput.value?.click()
  }
}

async function onFileSelected(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  input.value = ""
  if (!file) {
    return
  }
  error.value = ""
  success.value = ""
  lastAsset.value = null
  checkResult.value = null
  selectedFile.value = file

  if (file.size <= 0 || file.size > uploadMaxBytes.value) {
    status.value = "error"
    error.value = `文件大小需在 1 B 到 ${uploadMaxMb.value} MB 之间。`
    return
  }

  hashing.value = true
  status.value = "hashing"
  try {
    fileHash.value = await hashBlob(file)
    const check = await api.checkSystemMediaResumableUpload({
      chunkSize,
      chunkTotal: chunkTotal.value,
      fileHash: fileHash.value,
      fileName: file.name,
      sizeBytes: file.size
    })
    applyCheckResult(check)
    if (check.session.status === "completed" && check.asset) {
      lastAsset.value = check.asset
      success.value = "相同文件已存在，可直接在媒体库查看。"
      status.value = "completed"
      return
    }
    status.value = "ready"
    success.value = check.uploadedChunks.length ? `已识别 ${check.uploadedChunks.length} 个历史分片。` : ""
  } catch (err) {
    status.value = "error"
    error.value = errorMessage(err)
  } finally {
    hashing.value = false
  }
}

async function uploadFile() {
  if (!selectedFile.value || !checkResult.value || uploadUnavailable.value) {
    return
  }
  error.value = ""
  success.value = ""
  uploading.value = true
  status.value = "uploading"
  try {
    const file = selectedFile.value
    const sessionId = checkResult.value.session.id
    const missing = [...checkResult.value.missingChunks]
    for (const index of missing) {
      const start = index * chunkSize
      const end = Math.min(file.size, start + chunkSize)
      const chunk = file.slice(start, end)
      const chunkHash = await hashBlob(chunk)
      const result = await api.uploadSystemMediaChunk(chunk, {
        chunkHash,
        chunkIndex: index,
        chunkTotal: chunkTotal.value,
        fileHash: fileHash.value,
        fileName: file.name,
        sessionId
      })
      checkResult.value = {
        ...checkResult.value,
        missingChunks: result.missingChunks,
        progress: result.progress,
        uploadedChunks: result.uploadedChunks
      }
    }
    const complete = await api.completeSystemMediaResumableUpload({
      fileHash: fileHash.value,
      sessionId
    })
    lastAsset.value = complete.asset
    checkResult.value = {
      ...checkResult.value,
      missingChunks: [],
      progress: 100,
      session: {
        ...checkResult.value.session,
        finalAssetId: complete.asset.id,
        status: "completed"
      },
      uploadedChunks: Array.from({ length: chunkTotal.value }, (_, index) => index)
    }
    status.value = "completed"
    success.value = "上传完成，文件已写入媒体库。"
  } catch (err) {
    status.value = "error"
    error.value = errorMessage(err)
  } finally {
    uploading.value = false
  }
}

async function abortUpload() {
  if (!checkResult.value) {
    return
  }
  aborting.value = true
  error.value = ""
  success.value = ""
  try {
    await api.abortSystemMediaResumableUpload({
      fileHash: fileHash.value,
      sessionId: checkResult.value.session.id
    })
    status.value = "aborted"
    checkResult.value = {
      ...checkResult.value,
      missingChunks: Array.from({ length: chunkTotal.value }, (_, index) => index),
      progress: 0,
      uploadedChunks: []
    }
    success.value = "断点上传会话已中止。"
  } catch (err) {
    status.value = "error"
    error.value = errorMessage(err)
  } finally {
    aborting.value = false
  }
}

function resetState() {
  selectedFile.value = null
  fileHash.value = ""
  checkResult.value = null
  lastAsset.value = null
  error.value = ""
  success.value = ""
  status.value = "idle"
}

function applyCheckResult(check: SystemMediaResumableCheckResult) {
  checkResult.value = {
    ...check,
    missingChunks: Array.isArray(check.missingChunks) ? check.missingChunks : [],
    uploadedChunks: Array.isArray(check.uploadedChunks) ? check.uploadedChunks : []
  }
  uploadMaxBytes.value = check.uploadMaxBytes
  uploadMaxMb.value = check.uploadMaxMb
  uploadUnavailable.value = check.uploadUnavailable
}

async function hashBlob(blob: Blob) {
  const buffer = await blob.arrayBuffer()
  const digest = await crypto.subtle.digest("SHA-256", buffer)
  return Array.from(new Uint8Array(digest)).map((value) => value.toString(16).padStart(2, "0")).join("")
}

function formatBytes(value: number) {
  if (!value) {
    return "0 B"
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

onMounted(autoRefresh.refreshNow)

useHead({
  title: "断点上传 - Aoi Admin"
})
</script>

<template>
  <div class="page-grid">
    <PageHeader title="断点上传" icon="upload-cloud" description="按分片会话上传本地文件，完成后写入媒体库。">
      <template #actions>
        <AoiButton appearance="soft" icon="image-up" to="/media">媒体库</AoiButton>
        <AdminAutoRefreshControls
          v-model="autoRefresh.enabled.value"
          :last-refreshed-label="autoRefresh.lastRefreshedLabel.value"
          :next-refresh-label="autoRefresh.nextRefreshLabel.value"
          :status-label="autoRefresh.statusLabel.value"
        />
        <AoiButton appearance="soft" icon="refresh-cw" :loading="loadingStatus" :disabled="autoRefresh.refreshDisabled.value" @click="autoRefresh.refreshNow">刷新</AoiButton>
      </template>
    </PageHeader>

    <AoiStatusMessage tone="danger" :message="error" />
    <AoiStatusMessage tone="success" :message="success" />
    <AoiStatusMessage
      v-if="uploadUnavailable"
      tone="warning"
      message="对象存储未启用：断点上传需要先启用 storage。"
    />

    <AoiAdminCard class="resumable-card">
      <template #header>
        <div class="admin-card__header">
        <div>
          <h2>大文件上传</h2>
          <p class="muted">单文件上限 {{ uploadMaxMb }} MB，默认分片 {{ formatBytes(chunkSize) }}。</p>
        </div>
        <span class="badge">{{ statusLabel }}</span>
      </div>
      </template>

      <div class="resumable-actions">
        <input ref="fileInput" class="resumable-file-input" type="file" @change="onFileSelected">
        <AoiButton icon="file-plus-2" :disabled="!canPickFile" :loading="hashing" @click="openFilePicker">选择文件</AoiButton>
        <AoiButton icon="upload" :disabled="!canUpload" :loading="uploading" @click="uploadFile">上传文件</AoiButton>
        <AoiButton appearance="plain" icon="rotate-ccw" :disabled="uploading || hashing" @click="resetState">重置</AoiButton>
        <AoiButton appearance="soft" intent="danger" icon="ban" :disabled="!canAbort" :loading="aborting" @click="abortUpload">中止</AoiButton>
      </div>

      <div class="resumable-file">
        <AoiIcon name="file" decorative />
        <div>
          <strong>{{ selectedFileName }}</strong>
          <span>{{ selectedFileSize }} · {{ chunkTotal }} 个分片 · 已上传 {{ uploadedCount }} / {{ chunkTotal }}</span>
        </div>
        <span class="resumable-percent">{{ progress }}%</span>
      </div>
      <AoiProgressBar :value="progress" label="断点上传进度" />

      <dl class="resumable-meta">
        <div>
          <dt>会话</dt>
          <dd>{{ checkResult?.session.id || "-" }}</dd>
        </div>
        <div>
          <dt>缺失分片</dt>
          <dd>{{ missingCount }}</dd>
        </div>
        <div>
          <dt>过期时间</dt>
          <dd>{{ checkResult?.session.expiresAt ? formatDateTime(checkResult.session.expiresAt) : "-" }}</dd>
        </div>
        <div>
          <dt>文件 Hash</dt>
          <dd class="mono">{{ fileHash ? `${fileHash.slice(0, 16)}...` : "-" }}</dd>
        </div>
      </dl>
    </AoiAdminCard>

    <AoiAdminCard v-if="lastAsset" class="resumable-result">
      <template #header>
        <div class="admin-card__header">
        <div>
          <h2>上传结果</h2>
          <p class="muted">{{ lastAsset.displayName }} · {{ formatBytes(lastAsset.sizeBytes) }}</p>
        </div>
        <AoiButton icon="image-up" to="/media">查看媒体库</AoiButton>
      </div>
      </template>
    </AoiAdminCard>
  </div>
</template>

<style scoped>
.resumable-card,
.resumable-result {
  max-width: var(--aoi-admin-upload-card-max-width);
}

.resumable-actions,
.resumable-file {
  display: flex;
  align-items: center;
}

.resumable-actions {
  flex-wrap: wrap;
  gap: var(--aoi-admin-panel-gap-compact);
  padding: var(--aoi-admin-card-body-padding) 0;
}

.resumable-file-input {
  display: none;
}

.resumable-file {
  gap: var(--aoi-admin-card-gap);
  min-height: var(--aoi-admin-summary-min-height);
  padding: var(--aoi-admin-card-body-padding);
  border: 1px solid var(--aoi-admin-border);
  border-radius: var(--aoi-radius-card);
  background: var(--aoi-admin-surface-muted);
}

.resumable-file > div {
  display: grid;
  min-width: 0;
  flex: 1;
  gap: var(--aoi-admin-card-copy-gap);
}

.resumable-file strong,
.resumable-file span {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.resumable-file span {
  color: var(--aoi-admin-text-muted);
}

.resumable-percent {
  flex: 0 0 auto;
  font-weight: 760;
}

.resumable-meta {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: var(--aoi-admin-panel-gap-compact);
  margin: var(--aoi-admin-panel-gap) 0 0;
}

.resumable-meta div {
  display: grid;
  gap: var(--aoi-admin-card-copy-gap);
  min-width: 0;
  padding: var(--aoi-admin-card-body-padding-sm);
  border: 1px solid var(--aoi-admin-border);
  border-radius: var(--aoi-radius-card);
}

.resumable-meta dt {
  color: var(--aoi-admin-text-muted);
  font-size: 0.78rem;
}

.resumable-meta dd {
  min-width: 0;
  margin: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.mono {
  font-family: "JetBrains Mono", "SFMono-Regular", Consolas, monospace;
}

@media (max-width: 720px) {
  .resumable-actions,
  .resumable-file {
    align-items: stretch;
    flex-direction: column;
  }

  .resumable-file {
    text-align: left;
  }

  .resumable-meta {
    grid-template-columns: 1fr;
  }
}
</style>
