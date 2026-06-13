export type AdminAutoRefreshLabels = {
  disabled: string
  lastPrefix: string
  minutesUnit: string
  nextPrefix: string
  paused: string
  secondsUnit: string
  statusPrefix: string
  statusSeparator: string
  toggle: string
  unavailable: string
}

export type AdminAutoRefreshConfig = {
  clockTickMs: number
  defaultEnabled: boolean
  intervalMs: number
  labels: AdminAutoRefreshLabels
  locale: string
  manualCooldownMs: number
  minIntervalMs: number
  minuteMs: number
  secondMs: number
}

// 自动刷新被多个后台页面复用，时间单位和文案集中在这里，避免倒计时、状态标签和控件文案各自漂移。
export const ADMIN_AUTO_REFRESH_CONFIG = {
  clockTickMs: 1_000,
  defaultEnabled: true,
  intervalMs: 30_000,
  labels: {
    disabled: "自动刷新关闭",
    lastPrefix: "最近刷新",
    minutesUnit: "m",
    nextPrefix: "下次刷新",
    paused: "自动刷新暂停",
    secondsUnit: "s",
    statusPrefix: "自动刷新",
    statusSeparator: " · ",
    toggle: "自动刷新",
    unavailable: "-"
  },
  locale: "zh-CN",
  manualCooldownMs: 0,
  minIntervalMs: 1_000,
  minuteMs: 60_000,
  secondMs: 1_000
} satisfies AdminAutoRefreshConfig
