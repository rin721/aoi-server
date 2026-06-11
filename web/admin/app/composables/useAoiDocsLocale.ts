import type { AoiDocsLocale } from "~/types/docs"

export function normalizeAoiDocsLocale(value?: string | null): AoiDocsLocale {
  return value === "en" || value === "ja" ? value : "zh-CN"
}

export function useAoiDocsLocale() {
  const { locale } = useI18n()
  const appSettings = useAppSettingsStore()

  return computed<AoiDocsLocale>(() => {
    const preferredLocale = appSettings.hydrated ? appSettings.locale : locale.value

    return normalizeAoiDocsLocale(preferredLocale)
  })
}



