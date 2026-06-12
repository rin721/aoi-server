type AdminThemeMode = "light" | "dark" | "system"
type AdminDensity = "comfortable" | "compact"
type AdminInfoLayout = "adaptive" | "list" | "masonry"

type AdminAccentPreset = {
  accent10: string
  accent20: string
  accent40: string
  accent50: string
  accent60: string
  label: string
  value: string
}

type AdminUiPreferences = {
  accent: string
  contrast: boolean
  density: AdminDensity
  infoLayout: AdminInfoLayout
  reducedMotion: boolean
  sidebarCollapsed: boolean
  theme: AdminThemeMode
  watermark: boolean
}

const storageKey = "aoi.admin.ui.v1"

const defaults: AdminUiPreferences = {
  accent: "blue",
  contrast: false,
  density: "comfortable",
  infoLayout: "adaptive",
  reducedMotion: false,
  sidebarCollapsed: false,
  theme: "light",
  watermark: false
}

export const adminAccentPresets: AdminAccentPreset[] = [
  {
    accent10: "#eff6ff",
    accent20: "#bfdbfe",
    accent40: "#60a5fa",
    accent50: "#3b82f6",
    accent60: "#2563eb",
    label: "经典蓝",
    value: "blue"
  },
  {
    accent10: "#ecfdf5",
    accent20: "#a7f3d0",
    accent40: "#34d399",
    accent50: "#10b981",
    accent60: "#059669",
    label: "青绿",
    value: "green"
  },
  {
    accent10: "#fff7ed",
    accent20: "#fed7aa",
    accent40: "#fb923c",
    accent50: "#f97316",
    accent60: "#ea580c",
    label: "暖橙",
    value: "orange"
  },
  {
    accent10: "#fdf2f8",
    accent20: "#fbcfe8",
    accent40: "#f472b6",
    accent50: "#ec4899",
    accent60: "#db2777",
    label: "桃粉",
    value: "pink"
  }
]

const defaultAccent = adminAccentPresets[0] as AdminAccentPreset
const state = reactive<AdminUiPreferences>({ ...defaults })
let initialized = false

export function useAdminUiPreferences() {
  const systemDark = useMediaQuery("(prefers-color-scheme: dark)")

  if (!initialized && import.meta.client) {
    initialized = true
    load()
  }

  const activeAccent = computed(() =>
    adminAccentPresets.find((preset) => preset.value === state.accent) || defaultAccent
  )
  const darkActive = computed(() => state.theme === "dark" || (state.theme === "system" && systemDark.value))

  watch(
    () => ({ ...state, dark: darkActive.value }),
    () => {
      apply(activeAccent.value || defaultAccent, darkActive.value)
      save()
    },
    { deep: true, immediate: import.meta.client }
  )

  function load() {
    try {
      const raw = localStorage.getItem(storageKey)
      if (!raw) {
        return
      }
      Object.assign(state, defaults, JSON.parse(raw))
    } catch {
      Object.assign(state, defaults)
    }
  }

  function save() {
    if (!import.meta.client) {
      return
    }
    localStorage.setItem(storageKey, JSON.stringify(state))
  }

  function reset() {
    Object.assign(state, defaults)
  }

  function apply(preset: AdminAccentPreset, dark: boolean) {
    if (!import.meta.client) {
      return
    }

    const root = document.documentElement
    root.classList.toggle("dark", dark)
    root.dataset.aoiColorfulNav = "false"
    root.dataset.aoiAdminDensity = state.density
    root.dataset.aoiAdminInfoLayout = state.infoLayout
    root.dataset.aoiAdminReducedMotion = String(state.reducedMotion)
    root.dataset.aoiContrast = state.contrast ? "high" : "normal"
    root.dataset.aoiAdminWatermark = String(state.watermark)
    root.style.setProperty("--aoi-user-bg-image", "none")
    root.style.setProperty("--aoi-user-bg-opacity", "0")
    root.style.setProperty("--aoi-user-bg-blur", "0px")
    root.style.setProperty("--aoi-user-bg-dim", "0")
    root.style.setProperty("--aoi-accent-10", preset.accent10)
    root.style.setProperty("--aoi-accent-20", preset.accent20)
    root.style.setProperty("--aoi-accent-40", preset.accent40)
    root.style.setProperty("--aoi-accent-50", preset.accent50)
    root.style.setProperty("--aoi-accent-60", preset.accent60)
    root.style.setProperty("--aoi-active-color", dark ? preset.accent40 : preset.accent60)
    root.style.setProperty("--md-sys-color-primary", dark ? preset.accent40 : preset.accent60)
    root.style.setProperty("--md-sys-color-primary-container", preset.accent10)
  }

  return {
    accentPresets: adminAccentPresets,
    activeAccent,
    darkActive,
    reset,
    state
  }
}


