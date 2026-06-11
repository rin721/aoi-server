export default defineNuxtRouteMiddleware(async (to) => {
  const auth = useAuthStore()
  const isPublic = Boolean(to.meta.public)

  if (import.meta.client) {
    if (!auth.hydrated && (!isPublic || to.path === "/login" || to.path === "/setup")) {
      await auth.fetchSession()
    }

    if (!auth.authenticated) {
      const api = useAdminApi()
      try {
        const status = await api.getSetupStatus()
        if (status.required && to.path !== "/setup") {
          return navigateTo("/setup")
        }
        if (!status.required && to.path === "/setup") {
          return navigateTo("/login")
        }
      } catch {
        // setup 状态检查失败时继续原有认证流程，让具体页面展示 API 错误。
      }
    }
  }

  if (!isPublic && !auth.authenticated) {
    return navigateTo({
      path: "/login",
      query: { redirect: to.fullPath }
    })
  }

  if (to.path === "/login" && auth.authenticated) {
    return navigateTo("/")
  }

  if (to.path === "/setup" && auth.authenticated) {
    return navigateTo("/")
  }
})
