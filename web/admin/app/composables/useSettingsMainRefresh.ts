import { useAdminAutoRefresh, type AdminAutoRefreshOptions } from "~/composables/useAdminAutoRefresh"

export type SettingsMainRefreshOptions = AdminAutoRefreshOptions

// 设置页只在主页面创建刷新状态，派生面板通过主 store/computed 跟随。
export function useSettingsMainRefresh(options: SettingsMainRefreshOptions) {
  return useAdminAutoRefresh(options)
}
