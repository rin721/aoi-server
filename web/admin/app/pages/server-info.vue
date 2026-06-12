<script setup lang="ts">
import type { SystemServerDiskInfo, SystemServerInfo } from "~/types/admin"
import type { AoiKeyValueItem, AoiStatItem } from "~/types/ui"

const api = useAdminApi()
const info = ref<SystemServerInfo | null>(null)
const loading = ref(false)
const error = ref("")

const summaryCards = computed<AoiStatItem[]>(() => {
  const current = info.value
  return [
    { icon: "clock-3", label: "运行时长", value: current?.runtime.uptime || "-" },
    { icon: "activity", label: "CPU 平均", value: formatPercent(averagePercent(current?.cpu.percent)) },
    { icon: "database", label: "主机内存", value: formatPercent(current?.ram.usedPercent) },
    { icon: "hard-drive", label: "磁盘分区", value: formatNumber(current?.disk.length) }
  ]
})
const cpuAveragePercent = computed(() => averagePercent(info.value?.cpu.percent))
const cpuSamples = computed(() => info.value?.cpu.percent || [])
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
const environmentItems = computed<AoiKeyValueItem[]>(() => {
  const current = info.value
  if (!current) {
    return []
  }
  return [
    { label: "启动时间", value: formatDateTime(current.runtime.startTime) },
    { label: "刷新时间", value: formatDateTime(current.refreshedAt) },
    { label: "物理核心", value: current.cpu.cores || current.os.numCpu },
    { label: "逻辑 CPU", value: current.os.numCpu },
    { label: "编译器", value: current.os.compiler, monospace: true },
    { label: "协程", value: formatNumber(current.os.numGoroutine), monospace: true }
  ]
})
const ramItems = computed<AoiKeyValueItem[]>(() => {
  const current = info.value
  if (!current) {
    return []
  }
  return [
    { label: "总内存", value: formatMB(current.ram.totalMb), monospace: true },
    { label: "已使用", value: formatMB(current.ram.usedMb), monospace: true },
    { label: "可用估算", value: formatRAMFree(), monospace: true },
    { label: "使用率", value: formatPercent(current.ram.usedPercent), monospace: true }
  ]
})
const memoryItems = computed<AoiKeyValueItem[]>(() => {
  const current = info.value
  if (!current) {
    return []
  }
  return [
    { label: "当前分配", value: formatMB(current.memory.allocMb), monospace: true },
    { label: "系统占用", value: formatMB(current.memory.sysMb), monospace: true },
    { label: "累计分配", value: formatMB(current.memory.totalAllocMb), monospace: true },
    { label: "对象数量", value: formatNumber(current.memory.heapObjects), monospace: true }
  ]
})
const gcItems = computed<AoiKeyValueItem[]>(() => {
  const current = info.value
  if (!current) {
    return []
  }
  return [
    { label: "最近 GC", value: current.gc.lastGcAt ? formatDateTime(current.gc.lastGcAt) : "-" },
    { label: "暂停总时长", value: gcPauseMs.value, monospace: true },
    { label: "空闲堆", value: formatMB(current.memory.heapIdleMb), monospace: true },
    { label: "已释放堆", value: formatMB(current.memory.heapReleasedMb), monospace: true }
  ]
})
const buildItems = computed<AoiKeyValueItem[]>(() => {
  const current = info.value
  if (!current) {
    return []
  }
  return [
    { label: "主包", value: current.build.path || "-", monospace: true },
    { label: "Go 版本", value: current.build.goVersion, monospace: true },
    ...current.build.settings.map((setting) => ({
      label: setting.key,
      value: setting.value || "-",
      monospace: true
    }))
  ]
})

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

onMounted(load)

useHead({
  title: "服务器状态 - Aoi Admin"
})
</script>

<template>
  <div class="page-grid">
    <PageHeader title="服务器状态" icon="activity" description="展示主机 CPU、内存、磁盘以及当前 Go 进程运行快照。">
      <template #actions>
        <AoiButton appearance="soft" icon="refresh-cw" :loading="loading" @click="load">刷新</AoiButton>
      </template>
    </PageHeader>

    <AoiStatusMessage tone="danger" :message="error" />

    <AoiStatGrid :items="summaryCards" :columns="4" />

    <AoiMasonryGrid v-if="info" class="server-layout">
      <AoiAdminCard title="运行环境" :description="`${info.os.goos} / ${info.os.goarch}`" icon="server" :badge="info.os.goVersion">
        <AoiKeyValueList :items="environmentItems" layout="cards" />
      </AoiAdminCard>

      <AoiAdminCard
        title="CPU 负载"
        :description="cpuSamples.length ? `${cpuSamples.length} 个逻辑核心采样` : '暂无核心采样'"
        icon="activity"
        :badge="formatPercent(cpuAveragePercent)"
      >
        <AoiProgressBar :value="boundedPercent(cpuAveragePercent)" intent="info" size="md" label="CPU 平均负载" />
        <div v-if="cpuSamples.length" class="server-cpu-grid">
          <div
            v-for="(value, index) in cpuSamples"
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
      </AoiAdminCard>

      <AoiAdminCard title="主机内存" :description="`${formatMB(info.ram.usedMb)} / ${formatMB(info.ram.totalMb)}`" icon="database" :badge="formatPercent(info.ram.usedPercent)">
        <AoiProgressBar :value="ramUsagePercent" intent="info" size="md" label="主机内存使用率" />
        <AoiKeyValueList :items="ramItems" layout="cards" density="compact" />
      </AoiAdminCard>

      <AoiAdminCard title="Go 堆内存" :description="`${formatMB(info.memory.heapInuseMb)} / ${formatMB(info.memory.heapSysMb)}`" icon="memory-stick" :badge="`${heapUsagePercent}%`">
        <AoiProgressBar :value="heapUsagePercent" intent="info" size="md" label="堆内存使用率" />
        <AoiKeyValueList :items="memoryItems" layout="cards" density="compact" />
      </AoiAdminCard>

      <AoiAdminCard title="GC 状态" :description="`下次目标 ${formatMB(info.gc.nextGcMb)}`" icon="refresh-cw" :badge="`${info.gc.numGc} 次`">
        <AoiKeyValueList :items="gcItems" layout="rows" />
      </AoiAdminCard>

      <AoiAdminCard title="磁盘空间" :description="diskItems.length ? `${diskItems.length} 个挂载点` : '暂无磁盘采样'" icon="hard-drive" :badge="formatNumber(diskItems.length)">
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
      </AoiAdminCard>

      <AoiAdminCard title="构建信息" :description="info.build.module || info.build.path || '-'" icon="package-check" :badge="info.build.version || 'devel'">
        <AoiKeyValueList :items="buildItems" layout="rows" />
      </AoiAdminCard>
    </AoiMasonryGrid>

    <AoiAdminCard v-else class="server-empty" padding="lg">
      <AoiIcon name="activity" :size="24" decorative />
      <p>{{ loading ? "服务器状态加载中。" : "暂无服务器状态。" }}</p>
    </AoiAdminCard>
  </div>
</template>

<style scoped>
.server-layout :deep(.aoi-admin-card__body) {
  display: grid;
  gap: var(--aoi-admin-panel-gap-compact);
}

.server-cpu-grid,
.server-disk-list {
  display: grid;
  gap: var(--aoi-admin-panel-gap-compact);
  min-width: 0;
}

.server-cpu-grid {
  grid-template-columns: repeat(auto-fit, minmax(var(--aoi-admin-cpu-row-min-width), 1fr));
}

.server-cpu-row {
  --server-percent: 0%;
  align-items: center;
  display: grid;
  gap: var(--aoi-admin-kv-value-gap);
  grid-template-columns: var(--aoi-admin-cpu-label-width) minmax(0, 1fr) var(--aoi-admin-cpu-value-width);
  min-height: var(--aoi-admin-cpu-row-height);
  min-width: 0;
}

.server-cpu-row span,
.server-cpu-row strong,
.server-disk-row__meta span,
.server-disk-row__meta strong {
  font-family: var(--aoi-font-mono);
  font-size: var(--aoi-admin-kv-mono-size);
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
  height: var(--aoi-admin-meter-height);
  overflow: hidden;
  border-radius: var(--aoi-radius-round);
  background: var(--aoi-surface-muted);
}

.server-cpu-row i::before {
  display: block;
  width: var(--server-percent);
  height: 100%;
  border-radius: inherit;
  background: var(--aoi-info);
  content: "";
}

.server-disk-row {
  display: grid;
  gap: var(--aoi-admin-kv-value-gap);
  min-width: 0;
  border-top: 1px solid var(--aoi-border);
  padding-top: var(--aoi-admin-kv-card-padding);
}

.server-disk-row:first-child {
  border-top: 0;
  padding-top: 0;
}

.server-disk-row__head,
.server-disk-row__meta {
  align-items: center;
  display: flex;
  gap: var(--aoi-admin-card-gap);
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
  gap: var(--aoi-admin-card-gap);
  justify-content: center;
  min-height: 160px;
}

.server-empty p {
  margin: 0;
}

@media (max-width: 680px) {
  .server-disk-row__head,
  .server-disk-row__meta {
    align-items: flex-start;
    flex-direction: column;
    gap: var(--aoi-admin-card-copy-gap);
  }
}
</style>
