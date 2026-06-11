export type AdminNavItem = {
  icon: string
  label: string
  mobile?: boolean
  permission?: string
  to: string
}

export type AdminNavGroup = {
  code?: string
  label: string
  items: AdminNavItem[]
}

const baseGroups: AdminNavGroup[] = [
  {
    code: "workspace",
    label: "工作台",
    items: [
      { icon: "layout-dashboard", label: "仪表盘", mobile: true, to: "/" },
      { icon: "building-2", label: "组织", mobile: true, to: "/organizations" },
      { icon: "users", label: "用户", mobile: true, to: "/users" },
      { icon: "shield-check", label: "角色权限", mobile: true, to: "/roles" }
    ]
  },
  {
    code: "security",
    label: "安全审计",
    items: [
      { icon: "monitor-check", label: "会话", to: "/sessions" },
      { icon: "scroll-text", label: "审计日志", to: "/audit-logs" },
      { icon: "lock-keyhole", label: "安全", to: "/security" }
    ]
  },
  {
    code: "system",
    label: "系统管理",
    items: [
      { icon: "panel-left", label: "菜单管理", to: "/menus" },
      { icon: "code-2", label: "API 管理", to: "/apis" },
      { icon: "book-open", label: "字典管理", to: "/dictionaries" },
      { icon: "history", label: "操作历史", to: "/operation-records" },
      { icon: "compass", label: "参数管理", to: "/parameters" },
      { icon: "settings", label: "系统配置", to: "/system" },
      { icon: "activity", label: "服务器状态", to: "/server-info" }
    ]
  }
]

export function useAdminNavigation() {
  const config = useRuntimeConfig()
  const api = useAdminApi()
  const auth = useAuthStore()
  const serverGroups = useState<AdminNavGroup[] | null>("admin-navigation-groups", () => null)
  const serverLoaded = useState("admin-navigation-loaded", () => false)
  const serverLoading = useState("admin-navigation-loading", () => false)
  const serverScope = useState("admin-navigation-scope", () => "")

  const fallbackGroups = computed<AdminNavGroup[]>(() => {
    if (!config.public.showDemoTodo) {
      return baseGroups
    }

    return [
      ...baseGroups,
      {
        code: "examples",
        label: "示例",
        items: [
          { icon: "list-checks", label: "Demo Todo", to: "/todos" }
        ]
      }
    ]
  })

  async function loadNavigation() {
    if (!auth.authenticated || serverLoading.value || serverLoaded.value) {
      return
    }

    serverLoading.value = true
    try {
      const groups = await api.listSystemMenus()
      serverGroups.value = groups.map((group) => ({
        code: group.code,
        label: group.label,
        items: group.items.map((item) => ({
          icon: item.icon,
          label: item.label,
          mobile: item.mobile,
          permission: item.permission,
          to: item.path
        }))
      }))
    } catch {
      serverGroups.value = null
    } finally {
      serverLoaded.value = true
      serverLoading.value = false
    }
  }

  if (import.meta.client) {
    watch([
      () => auth.authenticated,
      () => auth.currentOrgId
    ], ([authenticated, currentOrgId]) => {
      const scope = authenticated ? String(currentOrgId || "") : ""
      if (scope && scope === serverScope.value && serverLoaded.value) {
        return
      }
      serverScope.value = scope
      serverLoaded.value = false
      if (!authenticated) {
        serverGroups.value = null
        return
      }
      void loadNavigation()
    }, { immediate: true })
  }

  const groups = computed<AdminNavGroup[]>(() => serverGroups.value?.length ? serverGroups.value : fallbackGroups.value)
  const items = computed(() => groups.value.flatMap((group) => group.items))
  const mobileItems = computed(() => items.value.filter((item) => item.mobile).slice(0, 4))

  function findItem(path: string) {
    return items.value.find((item) => item.to === "/" ? path === "/" : path.startsWith(item.to))
  }

  return {
    findItem,
    groups,
    items,
    loadNavigation,
    mobileItems
  }
}
