import { SERVER_STATUS_DASHBOARD_CONFIG } from "~/config/server-status-dashboard"
import type {
  ServerStatusDashboardConfig,
  ServerStatusHealthLevel,
  ServerStatusPanelId,
  ServerStatusThresholdId
} from "~/config/server-status-dashboard"
import type { SystemServerDiskInfo, SystemServerInfo } from "~/types/admin"
import type { AoiIntent, AoiKeyValueItem, AoiStatItem } from "~/types/ui"
import { formatDateTime } from "~/utils/format"

export type ServerStatusMetricState = {
  description: string
  icon: string
  intent: AoiIntent
  label: string
  level: ServerStatusHealthLevel
  weight: number
}

export type ServerStatusKpiItem = AoiStatItem & {
  id: string
  state: ServerStatusMetricState
}

export type ServerStatusCpuCoreRow = {
  displayValue: string
  index: number
  label: string
  state: ServerStatusMetricState
  value: number
}

export type ServerStatusDiskRow = {
  displayPercent: string
  displaySize: string
  fsType: string
  mountPoint: string
  state: ServerStatusMetricState
  value: number
}

export type ServerStatusResourcePanel = {
  badge: string
  badgeIntent: AoiIntent
  description: string
  icon: string
  id: ServerStatusPanelId
  title: string
}

export type ServerStatusDashboardModel = {
  anomalyCount: number
  anomalySummary: string
  buildItems: AoiKeyValueItem[]
  cpu: {
    averageDisplay: string
    averageState: ServerStatusMetricState
    averageValue: number
    cores: ServerStatusCpuCoreRow[]
    sampleCount: number
  }
  diskRows: ServerStatusDiskRow[]
  environmentItems: AoiKeyValueItem[]
  gcItems: AoiKeyValueItem[]
  heap: {
    display: string
    items: AoiKeyValueItem[]
    state: ServerStatusMetricState
    value: number
  }
  kpis: ServerStatusKpiItem[]
  overall: ServerStatusMetricState
  panels: Record<ServerStatusPanelId, ServerStatusResourcePanel>
  ram: {
    display: string
    items: AoiKeyValueItem[]
    state: ServerStatusMetricState
    value: number
  }
  refreshLabel: string
  refreshedAt: string
}

type NumericValue = number | null

export function createServerStatusDashboardModel(
  info: SystemServerInfo | null,
  config: ServerStatusDashboardConfig = SERVER_STATUS_DASHBOARD_CONFIG
): ServerStatusDashboardModel {
  // 页面只消费派生模型；阈值、文案、排序和格式化规则必须从配置进入这里统一收敛。
  const cpuAverage = averagePercent(info?.cpu.percent)
  const heapPercent = ratioPercent(info?.memory.heapInuseMb, info?.memory.heapSysMb)
  const ramPercent = finiteNumber(info?.ram.usedPercent)
  const diskRows = createDiskRows(info?.disk || [], config)
  const cpuAverageState = stateFromPercent(cpuAverage, "cpuAverage", config)
  const ramState = stateFromPercent(ramPercent, "ramUsed", config)
  const heapState = stateFromPercent(heapPercent, "heapUsed", config)
  const diskState = diskRows[0]?.state || unknownState(config)
  const candidates = [cpuAverageState, ramState, heapState, diskState]
  const overall = worstState(candidates, config)
  const anomalyCount = candidates.filter((state) => state.weight >= config.statusWeights.warning).length
  const refreshedAt = formatMaybeDate(info?.refreshedAt, config)

  return {
    anomalyCount,
    anomalySummary: anomalyCount
      ? `${config.labels.anomalyPrefix} ${formatNumber(anomalyCount, config)}`
      : config.statusLabels.healthy.description,
    buildItems: createBuildItems(info, config),
    cpu: {
      averageDisplay: formatPercent(cpuAverage, config),
      averageState: cpuAverageState,
      averageValue: boundedPercent(cpuAverage),
      cores: createCpuCoreRows(info?.cpu.percent || [], config),
      sampleCount: info?.cpu.percent.length || 0
    },
    diskRows,
    environmentItems: createFieldItems("environment", info, config),
    gcItems: createFieldItems("gc", info, config),
    heap: {
      display: formatPercent(heapPercent, config),
      items: createFieldItems("heap", info, config),
      state: heapState,
      value: boundedPercent(heapPercent)
    },
    kpis: createKpis(info, { cpuAverage, diskState, heapPercent, heapState, ramPercent, ramState, refreshedAt }, config),
    overall,
    panels: createPanels(info, { cpuAverage, cpuAverageState, diskRows, heapPercent, heapState, ramPercent, ramState }, config),
    ram: {
      display: formatPercent(ramPercent, config),
      items: createFieldItems("ram", info, config),
      state: ramState,
      value: boundedPercent(ramPercent)
    },
    refreshLabel: config.refresh.autoEnabled ? config.labels.autoRefreshEnabled : config.labels.autoRefreshDisabled,
    refreshedAt
  }
}

function createKpis(
  info: SystemServerInfo | null,
  derived: {
    cpuAverage: NumericValue
    diskState: ServerStatusMetricState
    heapPercent: NumericValue
    heapState: ServerStatusMetricState
    ramPercent: NumericValue
    ramState: ServerStatusMetricState
    refreshedAt: string
  },
  config: ServerStatusDashboardConfig
): ServerStatusKpiItem[] {
  return config.kpis.map((item) => {
    switch (item.id) {
      case "cpuAverage":
        const cpuState = stateFromPercent(derived.cpuAverage, "cpuAverage", config)
        return {
          ...item,
          description: derived.cpuAverage === null ? config.emptyStates.cpu : cpuState.description,
          intent: cpuState.intent,
          state: cpuState,
          value: formatPercent(derived.cpuAverage, config)
        }
      case "diskWorst":
        return {
          ...item,
          description: derived.diskState.description,
          intent: derived.diskState.intent,
          state: derived.diskState,
          value: info?.disk.length ? formatWorstDisk(info.disk, config) : config.format.emptyText
        }
      case "heapUsed":
        return {
          ...item,
          description: derived.heapState.description,
          intent: derived.heapState.intent,
          state: derived.heapState,
          value: formatPercent(derived.heapPercent, config)
        }
      case "ramUsed":
        return {
          ...item,
          description: derived.ramState.description,
          intent: derived.ramState.intent,
          state: derived.ramState,
          value: formatPercent(derived.ramPercent, config)
        }
      case "refreshedAt":
        return {
          ...item,
          description: config.refresh.autoEnabled ? config.labels.autoRefreshEnabled : config.labels.autoRefreshDisabled,
          intent: "neutral",
          state: info ? healthyState(config) : unknownState(config),
          value: derived.refreshedAt
        }
      case "uptime":
      default:
        return {
          ...item,
          description: formatMaybeDate(info?.runtime.startTime, config),
          intent: info ? "success" : "neutral",
          state: info ? healthyState(config) : unknownState(config),
          value: info?.runtime.uptime || config.format.emptyText
        }
    }
  })
}

function createPanels(
  info: SystemServerInfo | null,
  derived: {
    cpuAverage: NumericValue
    cpuAverageState: ServerStatusMetricState
    diskRows: ServerStatusDiskRow[]
    heapPercent: NumericValue
    heapState: ServerStatusMetricState
    ramPercent: NumericValue
    ramState: ServerStatusMetricState
  },
  config: ServerStatusDashboardConfig
): Record<ServerStatusPanelId, ServerStatusResourcePanel> {
  const empty = config.format.emptyText
  const panelMap = {} as Record<ServerStatusPanelId, ServerStatusResourcePanel>

  for (const panel of config.panels) {
    panelMap[panel.id] = {
      ...panel,
      badge: empty,
      badgeIntent: "neutral",
      description: empty
    }
  }

  panelMap.environment = {
    ...panelMap.environment,
    badge: info?.os.goVersion || empty,
    badgeIntent: info ? "info" : "neutral",
    description: info ? `${info.os.goos} / ${info.os.goarch}` : config.emptyStates.noData
  }
  panelMap.cpu = {
    ...panelMap.cpu,
    badge: formatPercent(derived.cpuAverage, config),
    badgeIntent: derived.cpuAverageState.intent,
    description: info?.cpu.percent.length
      ? `${formatNumber(info.cpu.percent.length, config)} ${config.labels.cpuSampleCountSuffix}`
      : config.emptyStates.cpu
  }
  panelMap.ram = {
    ...panelMap.ram,
    badge: formatPercent(derived.ramPercent, config),
    badgeIntent: derived.ramState.intent,
    description: info ? `${formatCapacityMb(info.ram.usedMb, config)} / ${formatCapacityMb(info.ram.totalMb, config)}` : config.emptyStates.noData
  }
  panelMap.heap = {
    ...panelMap.heap,
    badge: formatPercent(derived.heapPercent, config),
    badgeIntent: derived.heapState.intent,
    description: info ? `${formatCapacityMb(info.memory.heapInuseMb, config)} / ${formatCapacityMb(info.memory.heapSysMb, config)}` : config.emptyStates.noData
  }
  panelMap.gc = {
    ...panelMap.gc,
    badge: info ? `${formatNumber(info.gc.numGc, config)} ${config.labels.gcCountSuffix}` : empty,
    badgeIntent: "neutral",
    description: info ? `${config.labels.gcNextTargetPrefix} ${formatCapacityMb(info.gc.nextGcMb, config)}` : config.emptyStates.noData
  }
  panelMap.disk = {
    ...panelMap.disk,
    badge: info ? formatNumber(derived.diskRows.length, config) : empty,
    badgeIntent: derived.diskRows[0]?.state.intent || "neutral",
    description: derived.diskRows.length
      ? `${formatNumber(derived.diskRows.length, config)} ${config.labels.diskMountCountSuffix}`
      : config.emptyStates.disk
  }
  panelMap.build = {
    ...panelMap.build,
    badge: info?.build.version || config.format.buildFallbackVersion,
    badgeIntent: "neutral",
    description: info?.build.module || info?.build.path || config.emptyStates.build
  }

  return panelMap
}

function createFieldItems(
  group: "environment" | "gc" | "heap" | "ram",
  info: SystemServerInfo | null,
  config: ServerStatusDashboardConfig
): AoiKeyValueItem[] {
  if (!info) {
    return []
  }
  return config.fieldGroups[group].map((field) => ({
    label: field.label,
    monospace: field.monospace,
    value: valueForField(field.id, info, config)
  }))
}

function createBuildItems(info: SystemServerInfo | null, config: ServerStatusDashboardConfig): AoiKeyValueItem[] {
  if (!info) {
    return []
  }

  const baseItems = config.fieldGroups.build.map((field) => ({
    label: field.label,
    monospace: field.monospace,
    value: valueForField(field.id, info, config)
  }))
  const settingItems = info.build.settings.map((setting) => ({
    label: setting.key,
    monospace: true,
    value: setting.value || config.format.emptyText
  }))

  return [...baseItems, ...settingItems]
}

function valueForField(id: string, info: SystemServerInfo, config: ServerStatusDashboardConfig) {
  switch (id) {
    case "build.goVersion":
      return info.build.goVersion || config.format.emptyText
    case "build.module":
      return info.build.module || config.format.emptyText
    case "build.path":
      return info.build.path || config.format.emptyText
    case "build.version":
      return info.build.version || config.format.buildFallbackVersion
    case "cpu.cores":
      return info.cpu.cores || info.os.numCpu || config.format.emptyText
    case "gc.lastGcAt":
      return formatMaybeDate(info.gc.lastGcAt, config)
    case "gc.pauseTotalNs":
      return formatPause(info.gc.pauseTotalNs, config)
    case "memory.allocMb":
      return formatCapacityMb(info.memory.allocMb, config)
    case "memory.heapIdleMb":
      return formatCapacityMb(info.memory.heapIdleMb, config)
    case "memory.heapObjects":
      return formatNumber(info.memory.heapObjects, config)
    case "memory.heapReleasedMb":
      return formatCapacityMb(info.memory.heapReleasedMb, config)
    case "memory.sysMb":
      return formatCapacityMb(info.memory.sysMb, config)
    case "memory.totalAllocMb":
      return formatCapacityMb(info.memory.totalAllocMb, config)
    case "os.compiler":
      return info.os.compiler || config.format.emptyText
    case "os.numCpu":
      return info.os.numCpu || config.format.emptyText
    case "os.numGoroutine":
      return formatNumber(info.os.numGoroutine, config)
    case "ram.freeEstimateMb":
      return info.ram.totalMb ? formatCapacityMb(Math.max(0, info.ram.totalMb - info.ram.usedMb), config) : config.format.emptyText
    case "ram.totalMb":
      return formatCapacityMb(info.ram.totalMb, config)
    case "ram.usedMb":
      return formatCapacityMb(info.ram.usedMb, config)
    case "ram.usedPercent":
      return formatPercent(info.ram.usedPercent, config)
    case "refreshedAt":
      return formatMaybeDate(info.refreshedAt, config)
    case "runtime.startTime":
      return formatMaybeDate(info.runtime.startTime, config)
    default:
      return config.format.emptyText
  }
}

function createCpuCoreRows(values: number[], config: ServerStatusDashboardConfig): ServerStatusCpuCoreRow[] {
  return values.map((value, index) => {
    const finite = finiteNumber(value)
    return {
      displayValue: formatPercent(finite, config),
      index,
      label: `${config.labels.cpuCorePrefix} ${formatNumber(index + 1, config)}`,
      state: stateFromPercent(finite, "cpuCore", config),
      value: boundedPercent(finite)
    }
  })
}

function createDiskRows(items: SystemServerDiskInfo[], config: ServerStatusDashboardConfig): ServerStatusDiskRow[] {
  return items
    .map((item) => {
      const percent = finiteNumber(item.usedPercent)
      return {
        displayPercent: formatPercent(percent, config),
        displaySize: formatDiskSize(item, config),
        fsType: item.fsType || config.format.emptyText,
        mountPoint: item.mountPoint || config.format.emptyText,
        state: stateFromPercent(percent, "diskUsed", config),
        value: boundedPercent(percent)
      }
    })
    .sort((left, right) =>
      right.state.weight - left.state.weight
      || right.value - left.value
      || left.mountPoint.localeCompare(right.mountPoint, config.format.locale)
    )
}

function formatWorstDisk(items: SystemServerDiskInfo[], config: ServerStatusDashboardConfig) {
  const worst = createDiskRows(items, config)[0]
  return worst ? worst.displayPercent : config.format.emptyText
}

function formatDiskSize(item: SystemServerDiskInfo, config: ServerStatusDashboardConfig) {
  return `${formatCapacityMb(item.usedMb, config)}${config.format.capacity.separator}${formatCapacityMb(item.totalMb, config)}`
}

function stateFromPercent(
  value: NumericValue,
  thresholdId: ServerStatusThresholdId,
  config: ServerStatusDashboardConfig
): ServerStatusMetricState {
  // 状态判断集中在这里，避免页面或卡片组件重复写阈值比较导致语义漂移。
  const finite = finiteNumber(value)
  if (finite === null) {
    return unknownState(config)
  }

  const thresholds = config.thresholds[thresholdId]
  if (finite >= thresholds.critical) {
    return stateForLevel("critical", config)
  }
  if (finite >= thresholds.warning) {
    return stateForLevel("warning", config)
  }
  if (finite >= thresholds.notice) {
    return stateForLevel("notice", config)
  }
  return healthyState(config)
}

function worstState(states: ServerStatusMetricState[], config: ServerStatusDashboardConfig) {
  return [...states].sort((left, right) => right.weight - left.weight)[0] || unknownState(config)
}

function healthyState(config: ServerStatusDashboardConfig) {
  return stateForLevel("healthy", config)
}

function unknownState(config: ServerStatusDashboardConfig) {
  return stateForLevel("unknown", config)
}

function stateForLevel(level: ServerStatusHealthLevel, config: ServerStatusDashboardConfig): ServerStatusMetricState {
  const label = config.statusLabels[level]
  return {
    ...label,
    level,
    weight: config.statusWeights[level]
  }
}

function averagePercent(values?: number[]) {
  if (!values?.length) {
    return null
  }
  const finiteValues = values.filter((value) => Number.isFinite(value))
  if (!finiteValues.length) {
    return null
  }
  const total = finiteValues.reduce((sum, value) => sum + value, 0)
  return total / finiteValues.length
}

function ratioPercent(value?: number | null, total?: number | null) {
  const finiteValue = finiteNumber(value)
  const finiteTotal = finiteNumber(total)
  if (finiteValue === null || finiteTotal === null || finiteTotal <= 0) {
    return null
  }
  return finiteValue / finiteTotal * 100
}

function finiteNumber(value?: number | null): NumericValue {
  return typeof value === "number" && Number.isFinite(value) ? value : null
}

function boundedPercent(value?: number | null) {
  const finite = finiteNumber(value)
  if (finite === null) {
    return 0
  }
  return Math.min(100, Math.max(0, Math.round(finite)))
}

function formatPercent(value: NumericValue | undefined, config: ServerStatusDashboardConfig) {
  const finite = finiteNumber(value)
  if (finite === null) {
    return config.format.emptyText
  }
  return `${new Intl.NumberFormat(config.format.locale, {
    maximumFractionDigits: config.format.percentFractionDigits,
    minimumFractionDigits: Number.isInteger(finite) ? 0 : config.format.percentFractionDigits
  }).format(finite)}${config.format.percentUnit}`
}

function formatNumber(value: number | null | undefined, config: ServerStatusDashboardConfig) {
  const finite = finiteNumber(value)
  if (finite === null) {
    return config.format.emptyText
  }
  return new Intl.NumberFormat(config.format.locale).format(finite)
}

function formatCapacityMb(value: number | null | undefined, config: ServerStatusDashboardConfig) {
  const finite = finiteNumber(value)
  if (finite === null) {
    return config.format.emptyText
  }

  // 后端当前返回 MB 口径，UI 的 GB/MB 切换只处理展示，不改变接口语义。
  if (finite >= config.format.capacity.gbThresholdMb) {
    return `${new Intl.NumberFormat(config.format.locale, {
      maximumFractionDigits: config.format.capacity.fractionDigits
    }).format(finite / config.format.capacity.gbThresholdMb)} ${config.format.capacity.gbUnit}`
  }

  return `${formatNumber(finite, config)} ${config.format.capacity.mbUnit}`
}

function formatPause(value: number | null | undefined, config: ServerStatusDashboardConfig) {
  const finite = finiteNumber(value)
  if (finite === null || finite <= 0) {
    return `0 ${config.format.time.msUnit}`
  }

  const digits = finite < config.format.time.gcPauseCompactBelowNs ? config.format.percentFractionDigits : 0
  return `${new Intl.NumberFormat(config.format.locale, {
    maximumFractionDigits: digits,
    minimumFractionDigits: digits
  }).format(finite / config.format.time.nsPerMs)} ${config.format.time.msUnit}`
}

function formatMaybeDate(value: string | null | undefined, config: ServerStatusDashboardConfig) {
  const formatted = formatDateTime(value)
  return formatted || config.format.emptyText
}
