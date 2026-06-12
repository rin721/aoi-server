import type { AoiIntent } from "~/types/ui"

export type ServerStatusHealthLevel = "healthy" | "notice" | "warning" | "critical" | "unknown"
export type ServerStatusKpiId = "uptime" | "cpuAverage" | "ramUsed" | "diskWorst" | "heapUsed" | "refreshedAt"
export type ServerStatusPanelId = "environment" | "cpu" | "ram" | "heap" | "gc" | "disk" | "build"
export type ServerStatusThresholdId = "cpuAverage" | "cpuCore" | "ramUsed" | "diskUsed" | "heapUsed"

export type ServerStatusThreshold = {
  critical: number
  notice: number
  warning: number
}

export type ServerStatusLabelConfig = {
  description: string
  icon: string
  intent: AoiIntent
  label: string
}

export type ServerStatusFieldConfig = {
  id: string
  label: string
  monospace?: boolean
}

export type ServerStatusDashboardConfig = {
  emptyStates: {
    build: string
    cpu: string
    disk: string
    loading: string
    noData: string
  }
  fieldGroups: Record<"build" | "environment" | "gc" | "heap" | "ram", ServerStatusFieldConfig[]>
  format: {
    buildFallbackVersion: string
    capacity: {
      fractionDigits: number
      gbThresholdMb: number
      gbUnit: string
      mbUnit: string
      separator: string
    }
    emptyText: string
    locale: string
    percentFractionDigits: number
    percentUnit: string
    time: {
      gcPauseCompactBelowNs: number
      msUnit: string
      nsPerMs: number
    }
  }
  kpis: Array<{
    icon: string
    id: ServerStatusKpiId
    label: string
  }>
  labels: {
    anomalyPrefix: string
    autoRefreshDisabled: string
    autoRefreshEnabled: string
    cpuCorePrefix: string
    cpuSampleCountSuffix: string
    diskMountCountSuffix: string
    gcCountSuffix: string
    gcNextTargetPrefix: string
    pageDescription: string
    pageTitle: string
    refreshAction: string
  }
  panels: Array<{
    icon: string
    id: ServerStatusPanelId
    title: string
  }>
  refresh: {
    autoEnabled: boolean
    intervalMs: number
    manualCooldownMs: number
  }
  statusLabels: Record<ServerStatusHealthLevel, ServerStatusLabelConfig>
  statusWeights: Record<ServerStatusHealthLevel, number>
  thresholds: Record<ServerStatusThresholdId, ServerStatusThreshold>
}

export const SERVER_STATUS_DASHBOARD_CONFIG = {
  emptyStates: {
    build: "当前构建信息没有返回可展示字段。",
    cpu: "CPU 采样暂不可用，刷新后会再次尝试读取。",
    disk: "当前运行环境没有返回可读磁盘分区。",
    loading: "服务器状态加载中。",
    noData: "暂无服务器状态。"
  },
  fieldGroups: {
    build: [
      { id: "build.path", label: "主包", monospace: true },
      { id: "build.module", label: "模块", monospace: true },
      { id: "build.goVersion", label: "Go 版本", monospace: true },
      { id: "build.version", label: "构建版本", monospace: true }
    ],
    environment: [
      { id: "runtime.startTime", label: "启动时间" },
      { id: "refreshedAt", label: "刷新时间" },
      { id: "cpu.cores", label: "物理核心" },
      { id: "os.numCpu", label: "逻辑 CPU" },
      { id: "os.compiler", label: "编译器", monospace: true },
      { id: "os.numGoroutine", label: "协程", monospace: true }
    ],
    gc: [
      { id: "gc.lastGcAt", label: "最近 GC" },
      { id: "gc.pauseTotalNs", label: "暂停总时长", monospace: true },
      { id: "memory.heapIdleMb", label: "空闲堆", monospace: true },
      { id: "memory.heapReleasedMb", label: "已释放堆", monospace: true }
    ],
    heap: [
      { id: "memory.allocMb", label: "当前分配", monospace: true },
      { id: "memory.sysMb", label: "系统占用", monospace: true },
      { id: "memory.totalAllocMb", label: "累计分配", monospace: true },
      { id: "memory.heapObjects", label: "对象数量", monospace: true }
    ],
    ram: [
      { id: "ram.totalMb", label: "总内存", monospace: true },
      { id: "ram.usedMb", label: "已使用", monospace: true },
      { id: "ram.freeEstimateMb", label: "可用估算", monospace: true },
      { id: "ram.usedPercent", label: "使用率", monospace: true }
    ]
  },
  format: {
    buildFallbackVersion: "devel",
    capacity: {
      fractionDigits: 1,
      gbThresholdMb: 1024,
      gbUnit: "GB",
      mbUnit: "MB",
      separator: " / "
    },
    emptyText: "-",
    locale: "zh-CN",
    percentFractionDigits: 1,
    percentUnit: "%",
    time: {
      gcPauseCompactBelowNs: 10_000_000,
      msUnit: "ms",
      nsPerMs: 1_000_000
    }
  },
  kpis: [
    { icon: "clock-3", id: "uptime", label: "运行时长" },
    { icon: "activity", id: "cpuAverage", label: "CPU 平均" },
    { icon: "database", id: "ramUsed", label: "主机内存" },
    { icon: "hard-drive", id: "diskWorst", label: "磁盘水位" },
    { icon: "server", id: "heapUsed", label: "Go 堆" },
    { icon: "refresh-cw", id: "refreshedAt", label: "最近刷新" }
  ],
  labels: {
    anomalyPrefix: "异常指标",
    autoRefreshDisabled: "自动刷新关闭",
    autoRefreshEnabled: "自动刷新开启",
    cpuCorePrefix: "Core",
    cpuSampleCountSuffix: "个逻辑核心采样",
    diskMountCountSuffix: "个挂载点",
    gcCountSuffix: "次",
    gcNextTargetPrefix: "下次目标",
    pageDescription: "展示当前后端进程、主机资源和构建信息快照。",
    pageTitle: "服务器状态",
    refreshAction: "刷新"
  },
  panels: [
    { icon: "server", id: "environment", title: "运行环境" },
    { icon: "activity", id: "cpu", title: "CPU 负载" },
    { icon: "database", id: "ram", title: "主机内存" },
    { icon: "server", id: "heap", title: "Go 堆内存" },
    { icon: "refresh-cw", id: "gc", title: "GC 状态" },
    { icon: "hard-drive", id: "disk", title: "磁盘空间" },
    { icon: "code", id: "build", title: "构建信息" }
  ],
  refresh: {
    autoEnabled: false,
    intervalMs: 30_000,
    manualCooldownMs: 1_000
  },
  statusLabels: {
    critical: { description: "指标已进入高风险区间。", icon: "circle-alert", intent: "danger", label: "严重" },
    healthy: { description: "关键指标处于正常范围。", icon: "check", intent: "success", label: "正常" },
    notice: { description: "指标开始接近关注区间。", icon: "info", intent: "info", label: "关注" },
    unknown: { description: "当前指标暂无可用数据。", icon: "badge-help", intent: "neutral", label: "未知" },
    warning: { description: "指标已进入预警区间。", icon: "circle-alert", intent: "warning", label: "预警" }
  },
  statusWeights: {
    critical: 4,
    healthy: 0,
    notice: 1,
    unknown: 2,
    warning: 3
  },
  thresholds: {
    cpuAverage: { critical: 90, notice: 60, warning: 80 },
    cpuCore: { critical: 92, notice: 65, warning: 82 },
    diskUsed: { critical: 95, notice: 75, warning: 88 },
    heapUsed: { critical: 95, notice: 70, warning: 85 },
    ramUsed: { critical: 95, notice: 70, warning: 85 }
  }
} satisfies ServerStatusDashboardConfig
