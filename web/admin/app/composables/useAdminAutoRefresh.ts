import { ADMIN_AUTO_REFRESH_CONFIG, type AdminAutoRefreshConfig } from "~/config/admin-auto-refresh"
import type { MaybeRefOrGetter } from "vue"

export type AdminAutoRefreshLoadOptions = {
  silent?: boolean
}

export type AdminAutoRefreshOptions = {
  blocked?: MaybeRefOrGetter<boolean>
  config?: AdminAutoRefreshConfig
  defaultEnabled?: boolean
  intervalMs?: number
  load: (options?: AdminAutoRefreshLoadOptions) => Promise<void>
  manualCooldownMs?: number
}

export function useAdminAutoRefresh(options: AdminAutoRefreshOptions) {
  const config = options.config || ADMIN_AUTO_REFRESH_CONFIG
  const intervalMs = Math.max(config.minIntervalMs, options.intervalMs ?? config.intervalMs)
  const manualCooldownMs = Math.max(0, options.manualCooldownMs ?? config.manualCooldownMs)
  const enabled = ref(options.defaultEnabled ?? config.defaultEnabled)
  const refreshing = ref(false)
  const lastRefreshedAt = ref<number | null>(null)
  const manualCooldownUntil = ref(0)
  const nextRefreshAt = ref<number | null>(null)
  const now = ref(Date.now())
  let manualCooldownTimer: number | undefined
  let refreshTimer: number | undefined
  let clockTimer: number | undefined

  const isBlocked = computed(() => Boolean(options.blocked ? toValue(options.blocked) : false))
  const manualCooldownActive = computed(() => manualCooldownMs > 0 && now.value < manualCooldownUntil.value)
  const refreshDisabled = computed(() => refreshing.value || isBlocked.value || manualCooldownActive.value)
  const intervalLabel = computed(() => formatAutoRefreshDuration(intervalMs, config))
  const statusLabel = computed(() => {
    if (!enabled.value) {
      return config.labels.disabled
    }
    if (isBlocked.value) {
      return config.labels.paused
    }
    return `${config.labels.statusPrefix}${config.labels.statusSeparator}${intervalLabel.value}`
  })
  const lastRefreshedLabel = computed(() => {
    const value = lastRefreshedAt.value ? formatAutoRefreshTime(lastRefreshedAt.value, config) : config.labels.unavailable
    return `${config.labels.lastPrefix} ${value}`
  })
  const nextRefreshLabel = computed(() => {
    if (!enabled.value || !nextRefreshAt.value) {
      return `${config.labels.nextPrefix} ${config.labels.unavailable}`
    }
    if (isBlocked.value) {
      return `${config.labels.nextPrefix} ${config.labels.paused}`
    }
    const seconds = Math.max(0, Math.ceil((nextRefreshAt.value - now.value) / config.secondMs))
    return `${config.labels.nextPrefix} ${seconds}${config.labels.secondsUnit}`
  })

  async function refreshNow(trigger?: unknown) {
    const refreshRequest = resolveAutoRefreshRequest(trigger)
    if (refreshing.value || isBlocked.value || (refreshRequest.manual && manualCooldownActive.value)) {
      return
    }
    clearRefreshTimer()
    refreshing.value = true
    try {
      await options.load({ silent: refreshRequest.silent })
      lastRefreshedAt.value = Date.now()
    } finally {
      refreshing.value = false
      if (refreshRequest.manual && manualCooldownMs > 0) {
        startManualCooldown()
      }
      scheduleNext()
    }
  }

  function start() {
    enabled.value = true
    scheduleNext()
  }

  function stop() {
    enabled.value = false
    nextRefreshAt.value = null
    clearRefreshTimer()
  }

  function scheduleNext(delay = intervalMs) {
    if (!import.meta.client) {
      return
    }
    clearRefreshTimer()
    if (!enabled.value || isBlocked.value || document.visibilityState === "hidden") {
      nextRefreshAt.value = null
      return
    }
    nextRefreshAt.value = Date.now() + delay
    refreshTimer = window.setTimeout(() => {
      void refreshNow({ silent: true })
    }, delay)
  }

  function clearRefreshTimer() {
    if (refreshTimer) {
      window.clearTimeout(refreshTimer)
      refreshTimer = undefined
    }
  }

  function startManualCooldown() {
    manualCooldownUntil.value = Date.now() + manualCooldownMs
    if (!import.meta.client) {
      return
    }
    if (manualCooldownTimer) {
      window.clearTimeout(manualCooldownTimer)
    }
    // clockTickMs 驱动普通倒计时；这里单独在冷却边界刷新一次，避免短冷却被 tick 粒度拉长。
    manualCooldownTimer = window.setTimeout(() => {
      now.value = Date.now()
      manualCooldownTimer = undefined
    }, manualCooldownMs)
  }

  function handleVisibilityChange() {
    if (document.visibilityState === "hidden") {
      clearRefreshTimer()
      return
    }
    if (!enabled.value || isBlocked.value) {
      return
    }
    if (!nextRefreshAt.value || Date.now() >= nextRefreshAt.value) {
      void refreshNow({ silent: true })
      return
    }
    scheduleNext(nextRefreshAt.value - Date.now())
  }

  onMounted(() => {
    clockTimer = window.setInterval(() => {
      now.value = Date.now()
    }, config.clockTickMs)
    document.addEventListener("visibilitychange", handleVisibilityChange)
    scheduleNext()
  })

  onBeforeUnmount(() => {
    clearRefreshTimer()
    if (clockTimer) {
      window.clearInterval(clockTimer)
      clockTimer = undefined
    }
    if (manualCooldownTimer) {
      window.clearTimeout(manualCooldownTimer)
      manualCooldownTimer = undefined
    }
    document.removeEventListener("visibilitychange", handleVisibilityChange)
  })

  watch(enabled, (value) => {
    if (value) {
      scheduleNext()
      return
    }
    nextRefreshAt.value = null
    clearRefreshTimer()
  })

  watch(isBlocked, (blocked) => {
    if (blocked) {
      nextRefreshAt.value = null
      clearRefreshTimer()
      return
    }
    if (enabled.value) {
      scheduleNext()
    }
  })

  return {
    enabled,
    intervalLabel,
    isBlocked,
    lastRefreshedLabel,
    nextRefreshLabel,
    refreshDisabled,
    refreshing,
    refreshNow,
    start,
    statusLabel,
    stop
  }
}

function formatAutoRefreshDuration(value: number, config: AdminAutoRefreshConfig) {
  if (value >= config.minuteMs && value % config.minuteMs === 0) {
    return `${value / config.minuteMs}${config.labels.minutesUnit}`
  }
  return `${Math.round(value / config.secondMs)}${config.labels.secondsUnit}`
}

function formatAutoRefreshTime(value: number, config: AdminAutoRefreshConfig) {
  return new Date(value).toLocaleTimeString(config.locale, {
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit"
  })
}

type AdminAutoRefreshRequest = AdminAutoRefreshLoadOptions & {
  manual: boolean
}

function resolveAutoRefreshRequest(trigger: unknown): AdminAutoRefreshRequest {
  if (!trigger || typeof trigger !== "object") {
    return { manual: false }
  }
  if (typeof Event !== "undefined" && trigger instanceof Event) {
    return { manual: true }
  }
  if (!("silent" in trigger)) {
    return { manual: false }
  }
  return {
    manual: false,
    silent: (trigger as AdminAutoRefreshLoadOptions).silent === true
  }
}
