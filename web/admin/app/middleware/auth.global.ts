export default defineNuxtRouteMiddleware(async (to) => {
  const auth = useAuthStore()
  const isPublic = Boolean(to.meta.public)

  if (import.meta.client && !auth.hydrated && (!isPublic || to.path === "/login")) {
    await auth.fetchSession()
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
})
