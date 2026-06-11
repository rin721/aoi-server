<script setup lang="ts">
import type { SystemServerDiskInfo, SystemServerInfo } from "~/types/admin"

const api = useAdminApi()
const info = ref<SystemServerInfo | null>(null)
const loading = ref(false)
const error = ref("")

const summaryCards = computed(() => {
  const current = info.value
  return [
    {
      icon: "clock-3",
      label: "运行时长",
      value: current?.runtime.uptime || "-"
    },
    {
      icon: "activity",
      label: "CPU 平均",
      value: formatPercent(averagePercent(current?.cpu.percent))
    },
    {
      icon: "database",
      label: "主机内存",
      value: formatPercent(current?.ram.usedPercent)
    },
    {
      icon: "hard-drive",
      label: "磁盘分区",
      value: formatNumber(current?.disk.length)
    }
  ]
})

const cpuAveragePercent = computed(() => averagePercent(info.value?.cpu.percent))
const cpuPreview = computed(() => info.value?.cpu.percent.slice(0, 16) || [])
const diskItems = computed(() => info.value?.disk || [])
const heapUsagePercent = computed(() => percent(info.value?.memory.heapInuseMb, info.value?.memory.heapSysMb))
const ramUsagePercent = computed(() => boundedPercent(info.value?.ram.usedPercent))
const gcPauseMs = computed(() => {
  const ns = info.value?.gc.pauseTotalNs || 0
  if (!ns) {
    return "0 ms"
  }
  return `${(ns / 1_000_000).toFixed(ns < 10_000_000 ? 2 : 0)} ms`
})

const buildSettings = computed(() => info.value?.build.settings.slice(0, 10) || [])

async function load() {
  loading.value = true
  error.value = ""
  try {
    info.value = await api.getSystemServerInfo()
  } catch (err) {
    error.value = errorMessage(err)
  } finally {
    loading.value = false
  }
}

function formatMB(value?: number | null) {
  if (value === undefined || value === null || !Number.isFinite(value)) {
    return "-"
  }
  return `${value} MB`
}

function formatGB(value?: number | null) {
  if (value === undefined || value === null || !Number.isFinite(value)) {
    return "-"
  }
  return `${value} GB`
}

function formatNumber(value?: number | null) {
  if (value === undefined || value === null || !Number.isFinite(value)) {
    return "-"
  }
  return new Intl.NumberFormat("zh-CN").format(value)
}

function formatPercent(value?: number | null) {
  if (value === undefined || value === null || !Number.isFinite(value)) {
    return "-"
  }
  const rounded = Math.round(value * 10) / 10
  return `${Number.isInteger(rounded) ? rounded.toFixed(0) : rounded.toFixed(1)}%`
}

function percent(value?: number | null, total?: number | null) {
  if (!value || !total || total <= 0) {
    return 0
  }
  return Math.min(100, Math.round(value / total * 100))
}

function boundedPercent(value?: number | null) {
  if (value === undefined || value === null || !Number.isFinite(value)) {
    return 0
  }
  return Math.min(100, Math.max(0, Math.round(value)))
}

function averagePercent(values?: number[]) {
  if (!values?.length) {
    return null
  }
  const total = values.reduce((sum, value) => sum + (Number.isFinite(value) ? value : 0), 0)
  return Math.round(total / values.length * 10) / 10
}

function formatDiskSize(item: SystemServerDiskInfo) {
  if (item.totalGb > 0) {
    return `${formatGB(item.usedGb)} / ${formatGB(item.totalGb)}`
  }
  return `${formatMB(item.usedMb)} / ${formatMB(item.totalMb)}`
}

function formatRAMFree() {
  const current = info.value
  if (!current?.ram.totalMb) {
    return "-"
  }
  return formatMB(Math.max(0, current.ram.totalMb - current.ram.usedMb))
}

function formatBuildValue(value?: string) {
  if (!value) {
    return "-"
  }
  if (value.length <= 72) {
    return value
  }
  return `${value.slice(0, 68)}...`
}

onMounted(load)

useHead({
  title: "服务器状态 - Aoi Admin"
})
</script>

<template>
  <div class="page-grid">
    <PageHeader title="服务器状态" icon="activity" description="对齐 GVA 服务器信息能力，展示主机 CPU、内存、磁盘以及当前 Go 进程运行快照。">
      <template #actions>
        <AoiButton appearance="soft" icon="refresh-cw" :loading="loading" @click="load">刷新</AoiButton>
      </template>
    </PageHeader>

    <AoiStatusMessage tone="danger" :message="error" />

    <section class="server-overview" aria-label="服务器状态概览">
      <article v-for="card in summaryCards" :key="card.label" class="admin-card server-stat">
        <AoiIcon :name="card.icon" :size="22" decorative />
        <div>
          <span>{{ card.label }}</span>
          <strong>{{ card.value }}</strong>
        </div>
      </article>
    </section>

    <section v-if="info" class="server-layout">
      <article class="admin-card server-panel">
        <div class="admin-card__header">
          <div>
            <h2>运行环境</h2>
            <p>{{ info.os.goos }} / {{ info.os.goarch }}</p>
          </div>
          <span class="badge">{{ info.os.goVersion }}</span>
        </div>
        <dl class="server-kv">
          <div>
            <dt>启动时间</dt>
            <dd>{{ formatDateTime(info.runtime.startTime) }}</dd>
          </div>
          <div>
            <dt>刷新时间</dt>
            <dd>{{ formatDateTime(info.refreshedAt) }}</dd>
          </div>
          <div>
            <dt>物理核心</dt>
            <dd>{{ info.cpu.cores || info.os.numCpu }}</dd>
          </div>
          <div>
            <dt>逻辑 CPU</dt>
            <dd>{{ info.os.numCpu }}</dd>
          </div>
          <div>
            <dt>编译器</dt>
            <dd>{{ info.os.compiler }}</dd>
          </div>
          <div>
            <dt>协程</dt>
            <dd>{{ formatNumber(info.os.numGoroutine) }}</dd>
          </div>
        </dl>
      </article>

      <article class="admin-card server-panel">
        <div class="admin-card__header">
          <div>
            <h2>CPU 负载</h2>
            <p>{{ cpuPreview.length ? `${cpuPreview.length} 个逻辑核心采样` : "暂无核心采样" }}</p>
          </div>
          <span class="badge">{{ formatPercent(cpuAveragePercent) }}</span>
        </div>
        <AoiProgressBar :value="boundedPercent(cpuAveragePercent)" intent="info" size="md" label="CPU 平均负载" />
        <div v-if="cpuPreview.length" class="server-cpu-list">
          <div
            v-for="(value, index) in cpuPreview"
            :key="index"
            class="server-cpu-row"
            :style="{ '--server-percent': `${boundedPercent(value)}%` }"
          >
            <span>Core {{ index + 1 }}</span>
            <i aria-hidden="true" />
            <strong>{{ formatPercent(value) }}</strong>
          </div>
        </div>
        <p v-else class="server-note">CPU 采样暂不可用，刷新后会再次尝试读取。</p>
      </article>

      <article class="admin-card server-panel">
        <div class="admin-card__header">
          <div>
            <h2>主机内存</h2>
            <p>{{ formatMB(info.ram.usedMb) }} / {{ formatMB(info.ram.totalMb) }}</p>
          </div>
          <span class="badge">{{ formatPercent(info.ram.usedPercent) }}</span>
        </div>
        <AoiProgressBar :value="ramUsagePercent" intent="info" size="md" label="主机内存使用率" />
        <dl class="server-kv server-kv--compact">
          <div>
            <dt>总内存</dt>
            <dd>{{ formatMB(info.ram.totalMb) }}</dd>
          </div>
          <div>
            <dt>已使用</dt>
            <dd>{{ formatMB(info.ram.usedMb) }}</dd>
          </div>
          <div>
            <dt>可用估算</dt>
            <dd>{{ formatRAMFree() }}</dd>
          </div>
          <div>
            <dt>使用率</dt>
            <dd>{{ formatPercent(info.ram.usedPercent) }}</dd>
          </div>
        </dl>
      </article>

      <article class="admin-card server-panel">
        <div class="admin-card__header">
          <div>
            <h2>Go 堆内存</h2>
            <p>{{ formatMB(info.memory.heapInuseMb) }} / {{ formatMB(info.memory.heapSysMb) }}</p>
          </div>
          <span class="badge">{{ heapUsagePercent }}%</span>
        </div>
        <AoiProgressBar :value="heapUsagePercent" intent="info" size="md" label="堆内存使用率" />
        <dl class="server-kv server-kv--compact">
          <div>
            <dt>当前分配</dt>
            <dd>{{ formatMB(info.memory.allocMb) }}</dd>
          </div>
          <div>
            <dt>系统占用</dt>
            <dd>{{ formatMB(info.memory.sysMb) }}</dd>
          </div>
          <div>
            <dt>累计分配</dt>
            <dd>{{ formatMB(info.memory.totalAllocMb) }}</dd>
          </div>
          <div>
            <dt>对象数量</dt>
            <dd>{{ formatNumber(info.memory.heapObjects) }}</dd>
          </div>
        </dl>
      </article>

      <article class="admin-card server-panel">
        <div class="admin-card__header">
          <div>
            <h2>GC 状态</h2>
            <p>下次目标 {{ formatMB(info.gc.nextGcMb) }}</p>
          </div>
          <span class="badge">{{ info.gc.numGc }} 次</span>
        </div>
        <dl class="server-kv">
          <div>
            <dt>最近 GC</dt>
            <dd>{{ info.gc.lastGcAt ? formatDateTime(info.gc.lastGcAt) : "-" }}</dd>
          </div>
          <div>
            <dt>暂停总时长</dt>
            <dd>{{ gcPauseMs }}</dd>
          </div>
          <div>
            <dt>空闲堆</dt>
            <dd>{{ formatMB(info.memory.heapIdleMb) }}</dd>
          </div>
          <div>
            <dt>已释放堆</dt>
            <dd>{{ formatMB(info.memory.heapReleasedMb) }}</dd>
          </div>
        </dl>
      </article>

      <article class="admin-card server-panel server-panel--wide">
        <div class="admin-card__header">
          <div>
            <h2>磁盘空间</h2>
            <p>{{ diskItems.length ? `${diskItems.length} 个挂载点` : "暂无磁盘采样" }}</p>
          </div>
          <span class="badge">{{ formatNumber(diskItems.length) }}</span>
        </div>
        <div v-if="diskItems.length" class="server-disk-list">
          <div v-for="item in diskItems" :key="item.mountPoint" class="server-disk-row">
            <div class="server-disk-row__head">
              <strong>{{ item.mountPoint }}</strong>
              <span>{{ item.fsType || "未知文件系统" }}</span>
            </div>
            <AoiProgressBar :value="boundedPercent(item.usedPercent)" intent="info" size="sm" :label="`${item.mountPoint} 磁盘使用率`" />
            <div class="server-disk-row__meta">
              <span>{{ formatDiskSize(item) }}</span>
              <strong>{{ formatPercent(item.usedPercent) }}</strong>
            </div>
          </div>
        </div>
        <p v-else class="server-note">当前运行环境没有返回可读磁盘分区。</p>
      </article>

      <article class="admin-card server-panel server-panel--wide">
        <div class="admin-card__header">
          <div>
            <h2>构建信息</h2>
            <p>{{ info.build.module || info.build.path || "-" }}</p>
          </div>
          <span class="badge">{{ info.build.version || "devel" }}</span>
        </div>
        <dl class="server-kv server-kv--build">
          <div>
            <dt>主包</dt>
            <dd>{{ info.build.path || "-" }}</dd>
          </div>
          <div>
            <dt>Go 版本</dt>
            <dd>{{ info.build.goVersion }}</dd>
          </div>
          <div v-for="setting in buildSettings" :key="setting.key">
            <dt>{{ setting.key }}</dt>
            <dd>{{ formatBuildValue(setting.value) }}</dd>
          </div>
        </dl>
      </article>
    </section>

    <article v-else class="admin-card server-empty">
      <AoiIcon name="activity" :size="24" decorative />
      <p>{{ loading ? "服务器状态加载中。" : "暂无服务器状态。" }}</p>
    </article>
  </div>
</template>

<style scoped>
.server-overview {
  display: grid;
  gap: 12px;
  grid-template-columns: repeat(4, minmax(0, 1fr));
}

.server-stat {
  align-items: center;
  display: flex;
  gap: 12px;
  min-width: 0;
  padding: 14px;
}

.server-stat > svg {
  color: var(--aoi-accent-60);
  flex: 0 0 auto;
}

.server-stat div {
  display: grid;
  gap: 5px;
  min-width: 0;
}

.server-stat span {
  color: var(--aoi-text-muted);
  font-size: 12px;
  font-weight: 800;
}

.server-stat strong {
  color: var(--aoi-text);
  font-size: 22px;
  line-height: 1.15;
  min-width: 0;
  overflow-wrap: anywhere;
}

.server-layout {
  align-items: start;
  display: grid;
  gap: 14px;
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.server-panel {
  display: grid;
  gap: 16px;
  min-width: 0;
}

.server-panel--wide {
  grid-column: 1 / -1;
}

.server-panel .admin-card__header {
  align-items: flex-start;
}

.server-panel h2,
.server-panel p {
  margin: 0;
}

.server-panel p {
  color: var(--aoi-text-muted);
  line-height: 1.5;
  margin-top: 4px;
  overflow-wrap: anywhere;
}

.server-kv {
  display: grid;
  gap: 10px;
  margin: 0;
}

.server-kv--compact,
.server-kv--build {
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.server-kv div {
  align-items: start;
  border-top: 1px solid var(--aoi-border);
  display: grid;
  gap: 8px;
  grid-template-columns: minmax(110px, 150px) minmax(0, 1fr);
  min-width: 0;
  padding-top: 10px;
}

.server-kv--compact div,
.server-kv--build div {
  grid-template-columns: 1fr;
}

.server-kv dt {
  color: var(--aoi-text-muted);
  font-size: 12px;
  font-weight: 800;
}

.server-kv dd {
  color: var(--aoi-text);
  font-family: var(--aoi-font-mono);
  margin: 0;
  min-width: 0;
  overflow-wrap: anywhere;
}

.server-cpu-list,
.server-disk-list {
  display: grid;
  gap: 10px;
  min-width: 0;
}

.server-cpu-row {
  --server-percent: 0%;
  align-items: center;
  display: grid;
  gap: 10px;
  grid-template-columns: 66px minmax(0, 1fr) 54px;
  min-height: 28px;
  min-width: 0;
}

.server-cpu-row span,
.server-cpu-row strong,
.server-disk-row__meta span,
.server-disk-row__meta strong {
  font-family: var(--aoi-font-mono);
  font-size: 12px;
}

.server-cpu-row span {
  color: var(--aoi-text-muted);
}

.server-cpu-row strong {
  color: var(--aoi-text);
  text-align: right;
}

.server-cpu-row i {
  display: block;
  height: 8px;
  overflow: hidden;
  border-radius: var(--aoi-radius-round);
  background: var(--aoi-surface-muted);
}

.server-cpu-row i::before {
  display: block;
  width: var(--server-percent);
  height: 100%;
  border-radius: inherit;
  background: var(--aoi-accent-50);
  content: "";
}

.server-disk-row {
  display: grid;
  gap: 8px;
  min-width: 0;
  border-top: 1px solid var(--aoi-border);
  padding-top: 12px;
}

.server-disk-row:first-child {
  border-top: 0;
  padding-top: 0;
}

.server-disk-row__head,
.server-disk-row__meta {
  align-items: center;
  display: flex;
  gap: 10px;
  justify-content: space-between;
  min-width: 0;
}

.server-disk-row__head strong {
  color: var(--aoi-text);
  min-width: 0;
  overflow-wrap: anywhere;
}

.server-disk-row__head span,
.server-disk-row__meta span {
  color: var(--aoi-text-muted);
}

.server-disk-row__meta strong {
  color: var(--aoi-text);
}

.server-note {
  color: var(--aoi-text-muted);
  font-size: 13px;
  line-height: 1.6;
}

.server-empty {
  align-items: center;
  color: var(--aoi-text-muted);
  display: flex;
  gap: 10px;
  justify-content: center;
  min-height: 160px;
}

.server-empty p {
  margin: 0;
}

@media (max-width: 960px) {
  .server-overview,
  .server-layout {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 680px) {
  .server-overview,
  .server-layout,
  .server-kv--compact,
  .server-kv--build {
    grid-template-columns: 1fr;
  }

  .server-panel .admin-card__header {
    align-items: flex-start;
    flex-direction: column;
  }

  .server-kv div {
    grid-template-columns: 1fr;
  }

  .server-cpu-row {
    grid-template-columns: 58px minmax(0, 1fr) 48px;
  }

  .server-disk-row__head,
  .server-disk-row__meta {
    align-items: flex-start;
    flex-direction: column;
    gap: 4px;
  }
}
</style>
