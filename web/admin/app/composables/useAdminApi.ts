import type {
  ApiErrorPayload,
  AuditLog,
  HealthStatus,
  LoginRequest,
  MFASetupPayload,
  Organization,
  OrganizationUser,
  Permission,
  ReadyStatus,
  Result,
  Role,
  Session,
  Todo,
  TokenPair,
  User
} from "~/types/api"

type RequestOptions = {
  auth?: boolean
  body?: unknown
  method?: "DELETE" | "GET" | "PATCH" | "POST" | "PUT"
  query?: Record<string, unknown>
  retryAuth?: boolean
}

let refreshPromise: Promise<boolean> | null = null

export function useAdminApi() {
  const config = useRuntimeConfig()
  const baseURL = computed(() => config.public.apiBaseURL || "")

  async function request<T>(endpoint: string, options: RequestOptions = {}): Promise<T> {
    return requestWithAuthRetry<T>(endpoint, options, options.retryAuth !== false)
  }

  async function requestWithAuthRetry<T>(endpoint: string, options: RequestOptions, allowRefresh: boolean): Promise<T> {
    try {
      const headers: Record<string, string> = {}
      const accessToken = options.auth === false ? "" : getStoredAccessToken()
      if (accessToken) {
        headers.Authorization = `Bearer ${accessToken}`
      }

      const response = await $fetch<Result<T>>(endpoint, {
        baseURL: baseURL.value || undefined,
        body: options.body as Record<string, unknown> | undefined,
        headers,
        method: options.method || "GET",
        query: options.query
      })

      if (response.code !== 0) {
        throw toAdminApiError({ data: response, statusCode: 400 }, endpoint)
      }

      return response.data as T
    } catch (error) {
      const normalized = toAdminApiError(error, endpoint)
      if (shouldRefresh(normalized, options, allowRefresh) && await refreshAccessToken()) {
        return requestWithAuthRetry<T>(endpoint, options, false)
      }
      if (normalized.statusCode === 401 && options.auth !== false) {
        clearSessionAndRedirect()
      }
      throw normalized
    }
  }

  async function refreshAccessToken() {
    if (!refreshPromise) {
      refreshPromise = refreshTokenPair().finally(() => {
        refreshPromise = null
      })
    }
    return refreshPromise
  }

  async function refreshTokenPair() {
    const refreshToken = getStoredRefreshToken()
    if (!refreshToken) {
      return false
    }

    try {
      const pair = await requestWithAuthRetry<TokenPair>("/api/v1/auth/refresh", {
        auth: false,
        body: { refreshToken },
        method: "POST",
        retryAuth: false
      }, false)
      saveStoredTokenPair(pair)
      return true
    } catch {
      clearSessionAndRedirect()
      return false
    }
  }

  return {
    acceptInvitation: (token: string, body: { displayName?: string, password: string, username: string }) =>
      request<{ email?: string, orgId?: number, sessionId?: number, userId?: number }>(`/api/v1/invitations/${encodeURIComponent(token)}/accept`, {
        auth: false,
        body,
        method: "POST"
      }),
    createOrganization: (body: { code: string, name: string }) =>
      request<Organization>("/api/v1/orgs", { body, method: "POST" }),
    createRole: (orgId: number, body: { code: string, description?: string, name: string, permissions: string[] }) =>
      request<Role>(`/api/v1/orgs/${orgId}/roles`, { body, method: "POST" }),
    createTodo: (body: { completed?: boolean, description?: string, title: string }) =>
      request<Todo>("/api/v1/demo/todos", { body, method: "POST" }),
    deleteTodo: (id: number) => request<{ deleted: boolean }>(`/api/v1/demo/todos/${id}`, { method: "DELETE" }),
    forgotPassword: (email: string) =>
      request<{ token: string }>("/api/v1/auth/password/forgot", { auth: false, body: { email }, method: "POST" }),
    getHealth: () => request<HealthStatus>("/health", { auth: false }),
    getMe: () => request<User>("/api/v1/me"),
    getReady: () => request<ReadyStatus>("/ready", { auth: false }),
    inviteUser: (orgId: number, body: { email: string, roleCode: string }) =>
      request<{ token: string }>(`/api/v1/orgs/${orgId}/users/invitations`, { body, method: "POST" }),
    listAuditLogs: (orgId: number, limit = 100) =>
      request<AuditLog[]>(`/api/v1/orgs/${orgId}/audit-logs`, { query: { limit } }),
    listMyOrganizations: () => request<Organization[]>("/api/v1/me/orgs"),
    listOrganizations: () => request<Organization[]>("/api/v1/orgs"),
    listPermissions: (orgId: number) => request<Permission[]>(`/api/v1/orgs/${orgId}/permissions`),
    listRoles: (orgId: number) => request<Role[]>(`/api/v1/orgs/${orgId}/roles`),
    listSessions: (orgId: number, userId?: number | null) =>
      request<Session[]>(`/api/v1/orgs/${orgId}/sessions`, { query: userId ? { userId } : undefined }),
    listTodos: () => request<Todo[]>("/api/v1/demo/todos"),
    listUsers: (orgId: number) => request<OrganizationUser[]>(`/api/v1/orgs/${orgId}/users`),
    login: (body: LoginRequest) => request<TokenPair>("/api/v1/auth/login", { auth: false, body, method: "POST" }),
    logout: () => request<{ loggedOut: boolean }>("/api/v1/auth/logout", { method: "POST", retryAuth: false }),
    refreshSession: (refreshToken: string) =>
      request<TokenPair>("/api/v1/auth/refresh", { auth: false, body: { refreshToken }, method: "POST", retryAuth: false }),
    resetPassword: (body: { newPassword: string, token: string }) =>
      request<{ reset: boolean }>("/api/v1/auth/password/reset", { auth: false, body, method: "POST" }),
    revokeSession: (orgId: number, sessionId: number) =>
      request<{ revoked: boolean }>(`/api/v1/orgs/${orgId}/sessions/${sessionId}`, { method: "DELETE" }),
    setupMFA: () => request<MFASetupPayload>("/api/v1/auth/mfa/setup", { method: "POST" }),
    switchOrg: (orgId: number) => request<TokenPair>("/api/v1/auth/switch-org", { body: { orgId }, method: "POST" }),
    updateTodo: (id: number, body: { completed?: boolean, description?: string, title?: string }) =>
      request<Todo>(`/api/v1/demo/todos/${id}`, { body, method: "PUT" }),
    verifyMFA: (code: string) => request<{ verified: boolean }>("/api/v1/auth/mfa/verify", { body: { code }, method: "POST" })
  }
}

function toAdminApiError(error: unknown, endpoint: string): ApiErrorPayload {
  if (isApiErrorPayload(error)) {
    return error
  }

  const fetchError = error as {
    data?: Result<unknown>
    message?: string
    status?: number
    statusCode?: number
    statusMessage?: string
  }
  const statusCode = fetchError.statusCode || fetchError.status || 500
  const payload = fetchError.data

  return {
    code: payload?.code ?? fetchError.statusMessage ?? "ADMIN_API_ERROR",
    endpoint,
    message: payload?.message || fetchError.message || "请求暂时失败，请稍后重试。",
    serverTime: payload?.serverTime,
    statusCode,
    traceId: payload?.traceId
  }
}

function shouldRefresh(error: ApiErrorPayload, options: RequestOptions, allowRefresh: boolean) {
  return error.statusCode === 401 && allowRefresh && options.auth !== false && Boolean(getStoredRefreshToken())
}

function clearSessionAndRedirect() {
  clearStoredTokenPair()
  if (!import.meta.client) {
    return
  }

  const auth = useAuthStore()
  auth.clearSession()

  const route = useRoute()
  if (route.path !== "/login") {
    void navigateTo({
      path: "/login",
      query: { redirect: route.fullPath }
    })
  }
}

function isApiErrorPayload(value: unknown): value is ApiErrorPayload {
  return Boolean(value && typeof value === "object" && "endpoint" in value && "statusCode" in value)
}
