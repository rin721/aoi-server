export type AdminNavItem = {
  icon: string
  label: string
  mobile?: boolean
  to: string
}

export type AdminNavGroup = {
  label: string
  items: AdminNavItem[]
}

const baseGroups: AdminNavGroup[] = [
  {
    label: "工作台",
    items: [
      { icon: "layout-dashboard", label: "仪表盘", mobile: true, to: "/" },
      { icon: "building-2", label: "组织", mobile: true, to: "/organizations" },
      { icon: "users", label: "用户", mobile: true, to: "/users" },
      { icon: "shield-check", label: "角色权限", mobile: true, to: "/roles" }
    ]
  },
  {
    label: "安全审计",
    items: [
      { icon: "monitor-check", label: "会话", to: "/sessions" },
      { icon: "scroll-text", label: "审计日志", to: "/audit-logs" },
      { icon: "lock-keyhole", label: "安全", to: "/security" }
    ]
  }
]

export function useAdminNavigation() {
  const config = useRuntimeConfig()

  const groups = computed<AdminNavGroup[]>(() => {
    if (!config.public.showDemoTodo) {
      return baseGroups
    }

    return [
      ...baseGroups,
      {
        label: "示例",
        items: [
          { icon: "list-checks", label: "Demo Todo", to: "/todos" }
        ]
      }
    ]
  })

  const items = computed(() => groups.value.flatMap((group) => group.items))
  const mobileItems = computed(() => items.value.filter((item) => item.mobile).slice(0, 4))

  function findItem(path: string) {
    return items.value.find((item) => item.to === "/" ? path === "/" : path.startsWith(item.to))
  }

  return {
    findItem,
    groups,
    items,
    mobileItems
  }
}


