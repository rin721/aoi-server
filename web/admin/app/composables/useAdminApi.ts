import type {
  ApiErrorPayload,
  APITokenPage,
  CaptchaChallenge,
  AuditLog,
  CreateAPITokenResult,
  HealthStatus,
  ID,
  InitialAdminSetupRequest,
  Invitation,
  LoginRequest,
  MFASetupPayload,
  NotificationDelivery,
  Organization,
  OrganizationUser,
  Permission,
  PluginHealthStatus,
  PluginManifest,
  ReadyStatus,
  Result,
  Role,
  Session,
  SignupRequest,
  SetupStatus,
  SystemAPIGroup,
  SystemAPISyncResult,
  SystemConfigSnapshot,
  SystemDictionary,
  SystemDictionaryCatalog,
  SystemDictionaryItem,
  SystemMediaAsset,
  SystemMediaAssetPage,
  SystemMediaCategory,
  SystemMediaCategoryCatalog,
  SystemMediaURLImportResult,
  SystemOperationRecordPage,
  SystemParameter,
  SystemParameterPage,
  SystemPermissionSyncResult,
  SystemMenuGroup,
  SystemServerInfo,
  SystemVersionDetail,
  SystemVersionImportResult,
  SystemVersionPackage,
  SystemVersionPage,
  SystemVersionSourceCatalog,
  Todo,
  TokenPair,
  User
} from "~/types/admin"

type RequestOptions = {
  auth?: boolean
  body?: unknown
  method?: "DELETE" | "GET" | "PATCH" | "POST" | "PUT"
  query?: Record<string, unknown>
  retryAuth?: boolean
}

type AuditLogQuery = {
  action?: string
  cursor?: ID | null
  from?: string
  limit?: number
  to?: string
  userId?: ID | null
}

export type PluginProxyOptions = {
  body?: unknown
  method?: "DELETE" | "GET" | "PATCH" | "POST" | "PUT"
  query?: Record<string, unknown>
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

  async function downloadWithAuthRetry(endpoint: string, allowRefresh = true): Promise<{ blob: Blob, contentType: string, filename: string }> {
    try {
      const headers: Record<string, string> = {}
      const accessToken = getStoredAccessToken()
      if (accessToken) {
        headers.Authorization = `Bearer ${accessToken}`
      }
      const response = await fetch(resolveEndpointURL(endpoint, baseURL.value), { headers })
      if (!response.ok) {
        const payload = await readErrorPayload(response)
        throw toAdminApiError({
          data: payload,
          message: response.statusText,
          statusCode: response.status,
          statusMessage: response.statusText
        }, endpoint)
      }
      return {
        blob: await response.blob(),
        contentType: response.headers.get("content-type") || "application/octet-stream",
        filename: filenameFromContentDisposition(response.headers.get("content-disposition")) || "media-download"
      }
    } catch (error) {
      const normalized = toAdminApiError(error, endpoint)
      if (shouldRefresh(normalized, {}, allowRefresh) && await refreshAccessToken()) {
        return downloadWithAuthRetry(endpoint, false)
      }
      if (normalized.statusCode === 401) {
        clearSessionAndRedirect()
      }
      throw normalized
    }
  }

  return {
    acceptInvitation: (token: string, body: { displayName?: string, password: string, username: string }) =>
      request<{ email?: string, orgId?: ID, sessionId?: ID, userId?: ID }>(`/api/v1/invitations/${encodeURIComponent(token)}/accept`, {
        auth: false,
        body,
        method: "POST"
      }),
    createOrganization: (body: { code: string, name: string }) =>
      request<Organization>("/api/v1/orgs", { body, method: "POST" }),
    createAPIToken: (orgId: ID, body: { days?: number, remark?: string, roleCode: string, userId: ID }) =>
      request<CreateAPITokenResult>(`/api/v1/orgs/${orgId}/api-tokens`, { body, method: "POST" }),
    createRole: (orgId: ID, body: { code: string, description?: string, name: string, permissions: string[] }) =>
      request<Role>(`/api/v1/orgs/${orgId}/roles`, { body, method: "POST" }),
    createTodo: (body: { completed?: boolean, description?: string, title: string }) =>
      request<Todo>("/api/v1/demo/todos", { body, method: "POST" }),
    deleteTodo: (id: number) => request<{ deleted: boolean }>(`/api/v1/demo/todos/${id}`, { method: "DELETE" }),
    forgotPassword: (email: string) =>
      request<NotificationDelivery>("/api/v1/auth/password/forgot", { auth: false, body: { email }, method: "POST" }),
    getAuthCaptcha: () => request<CaptchaChallenge>("/api/v1/auth/captcha", { auth: false }),
    getHealth: () => request<HealthStatus>("/health", { auth: false }),
    getMe: () => request<User>("/api/v1/me"),
    getReady: () => request<ReadyStatus>("/ready", { auth: false }),
    getSetupStatus: () => request<SetupStatus>("/api/v1/auth/setup/status", { auth: false }),
    getSystemConfig: () => request<SystemConfigSnapshot>("/api/v1/system/config"),
    getSystemServerInfo: () => request<SystemServerInfo>("/api/v1/system/server-info"),
    initialAdminSetup: (body: InitialAdminSetupRequest) =>
      request<TokenPair>("/api/v1/auth/setup/initial-admin", { auth: false, body, method: "POST" }),
    inviteUser: (orgId: ID, body: { email: string, roleCode: string }) =>
      request<NotificationDelivery>(`/api/v1/orgs/${orgId}/users/invitations`, { body, method: "POST" }),
    listAuditLogs: (orgId: ID, query: AuditLogQuery | number = 100) =>
      request<AuditLog[]>(`/api/v1/orgs/${orgId}/audit-logs`, { query: typeof query === "number" ? { limit: query } : query }),
    listInvitations: (orgId: ID) => request<Invitation[]>(`/api/v1/orgs/${orgId}/invitations`),
    listMyOrganizations: () => request<Organization[]>("/api/v1/me/orgs"),
    listOrganizations: () => request<Organization[]>("/api/v1/orgs"),
    listPermissions: (orgId: ID) => request<Permission[]>(`/api/v1/orgs/${orgId}/permissions`),
    listAPITokens: (orgId: ID, query: { page?: number, pageSize?: number, status?: string, userId?: ID | null } = {}) =>
      request<APITokenPage>(`/api/v1/orgs/${orgId}/api-tokens`, { query }),
    getPlugin: (pluginId: string) => request<PluginManifest>(`/api/v1/plugins/${encodeURIComponent(pluginId)}`),
    getPluginHealth: (pluginId: string) => request<PluginHealthStatus>(`/api/v1/plugins/${encodeURIComponent(pluginId)}/health`),
    listPlugins: () => request<PluginManifest[]>("/api/v1/plugins"),
    proxyPlugin: <T = unknown>(pluginId: string, path: string, options: PluginProxyOptions = {}) =>
      request<T>(pluginProxyEndpoint(pluginId, path), {
        body: options.body,
        method: options.method || "GET",
        query: options.query
      }),
    listRoles: (orgId: ID) => request<Role[]>(`/api/v1/orgs/${orgId}/roles`),
    listSessions: (orgId: ID, userId?: ID | null) =>
      request<Session[]>(`/api/v1/orgs/${orgId}/sessions`, { query: userId ? { userId } : undefined }),
    listSystemAPIs: () => request<SystemAPIGroup[]>("/api/v1/system/apis"),
    listSystemDictionaries: () => request<SystemDictionaryCatalog>("/api/v1/system/dictionaries"),
    listSystemMenus: () => request<SystemMenuGroup[]>("/api/v1/system/menus"),
    listSystemOperationRecords: (query: { method?: string, page?: number, pageSize?: number, path?: string, status?: number | string, statusClass?: string } = {}) =>
      request<SystemOperationRecordPage>("/api/v1/system/operation-records", { query }),
    listSystemParameters: (query: { endCreatedAt?: string, key?: string, name?: string, page?: number, pageSize?: number, startCreatedAt?: string } = {}) =>
      request<SystemParameterPage>("/api/v1/system/parameters", { query }),
    listSystemVersionSources: () => request<SystemVersionSourceCatalog>("/api/v1/system/versions/sources"),
    listSystemVersions: (query: { endCreatedAt?: string, page?: number, pageSize?: number, startCreatedAt?: string, versionCode?: string, versionName?: string } = {}) =>
      request<SystemVersionPage>("/api/v1/system/versions", { query }),
    createSystemDictionary: (body: { code: string, description?: string, name: string, status?: string }) =>
      request<SystemDictionary>("/api/v1/system/dictionaries", { body, method: "POST" }),
    createSystemParameter: (body: { description?: string, key: string, name: string, value: string }) =>
      request<SystemParameter>("/api/v1/system/parameters", { body, method: "POST" }),
    exportSystemVersion: (body: { apiCodes: string[], description?: string, dictionaryCodes: string[], menuCodes: string[], versionCode: string, versionName: string }) =>
      request<SystemVersionDetail>("/api/v1/system/versions/export", { body, method: "POST" }),
    importSystemVersion: (versionData: string) =>
      request<SystemVersionImportResult>("/api/v1/system/versions/import", { body: { versionData }, method: "POST" }),
    importSystemMediaURLs: (body: { categoryId?: ID | number, items?: Array<{ name?: string, url: string }>, text?: string }) =>
      request<SystemMediaURLImportResult>("/api/v1/system/media/assets/import-url", { body, method: "POST" }),
    createSystemDictionaryItem: (dictionaryId: ID, body: { extra?: string, label: string, sort?: number, status?: string, value: string }) =>
      request<SystemDictionaryItem>(`/api/v1/system/dictionaries/${dictionaryId}/items`, { body, method: "POST" }),
    saveSystemMediaCategory: (body: { id?: ID | number, name: string, parentId?: ID | number, sort?: number }) =>
      request<SystemMediaCategory>("/api/v1/system/media/categories", { body, method: "POST" }),
    deleteSystemVersion: (versionId: ID) =>
      request<{ deleted: boolean }>(`/api/v1/system/versions/${versionId}`, { method: "DELETE" }),
    deleteSystemVersions: (ids: ID[]) =>
      request<{ deleted: boolean }>("/api/v1/system/versions", { body: { ids }, method: "DELETE" }),
    downloadSystemVersion: (versionId: ID) =>
      request<SystemVersionPackage>(`/api/v1/system/versions/${versionId}/download`),
    downloadSystemMediaAsset: (assetId: ID) =>
      downloadWithAuthRetry(`/api/v1/system/media/assets/${assetId}/download`),
    deleteSystemMediaAsset: (assetId: ID) =>
      request<{ deleted: boolean }>(`/api/v1/system/media/assets/${assetId}`, { method: "DELETE" }),
    deleteSystemMediaCategory: (categoryId: ID) =>
      request<{ deleted: boolean }>(`/api/v1/system/media/categories/${categoryId}`, { method: "DELETE" }),
    deleteSystemDictionary: (dictionaryId: ID) =>
      request<{ deleted: boolean }>(`/api/v1/system/dictionaries/${dictionaryId}`, { method: "DELETE" }),
    deleteSystemDictionaryItem: (itemId: ID) =>
      request<{ deleted: boolean }>(`/api/v1/system/dictionary-items/${itemId}`, { method: "DELETE" }),
    deleteSystemOperationRecords: (ids: ID[]) =>
      request<{ deleted: boolean }>("/api/v1/system/operation-records", { body: { ids }, method: "DELETE" }),
    deleteSystemParameter: (parameterId: ID) =>
      request<{ deleted: boolean }>(`/api/v1/system/parameters/${parameterId}`, { method: "DELETE" }),
    deleteSystemParameters: (ids: ID[]) =>
      request<{ deleted: boolean }>("/api/v1/system/parameters", { body: { ids }, method: "DELETE" }),
    getSystemParameterByKey: (key: string) =>
      request<SystemParameter>("/api/v1/system/parameters/value", { query: { key } }),
    getSystemVersion: (versionId: ID) =>
      request<SystemVersionDetail>(`/api/v1/system/versions/${versionId}`),
    listSystemMediaAssets: (query: { categoryId?: ID | number, keyword?: string, page?: number, pageSize?: number } = {}) =>
      request<SystemMediaAssetPage>("/api/v1/system/media/assets", { query }),
    listSystemMediaCategories: () => request<SystemMediaCategoryCatalog>("/api/v1/system/media/categories"),
    syncSystemAPIs: () => request<SystemAPISyncResult>("/api/v1/system/apis/sync", { method: "POST" }),
    syncSystemAPIPermissions: () => request<SystemPermissionSyncResult>("/api/v1/system/apis/permissions/sync", { method: "POST" }),
    updateSystemDictionary: (dictionaryId: ID, body: { description?: string, name?: string, status?: string }) =>
      request<SystemDictionary>(`/api/v1/system/dictionaries/${dictionaryId}`, { body, method: "PATCH" }),
    updateSystemDictionaryItem: (itemId: ID, body: { extra?: string, label?: string, sort?: number, status?: string, value?: string }) =>
      request<SystemDictionaryItem>(`/api/v1/system/dictionary-items/${itemId}`, { body, method: "PATCH" }),
    updateSystemMediaAsset: (assetId: ID, body: { displayName: string }) =>
      request<SystemMediaAsset>(`/api/v1/system/media/assets/${assetId}`, { body, method: "PATCH" }),
    updateSystemParameter: (parameterId: ID, body: { description?: string, key?: string, name?: string, value?: string }) =>
      request<SystemParameter>(`/api/v1/system/parameters/${parameterId}`, { body, method: "PATCH" }),
    uploadSystemMediaAsset: (file: File, categoryId?: ID | number) => {
      const body = new FormData()
      body.append("file", file)
      if (categoryId) {
        body.append("categoryId", String(categoryId))
      }
      return request<SystemMediaAsset>("/api/v1/system/media/assets/upload", { body, method: "POST" })
    },
    listTodos: () => request<Todo[]>("/api/v1/demo/todos"),
    listUsers: (orgId: ID) => request<OrganizationUser[]>(`/api/v1/orgs/${orgId}/users`),
    login: (body: LoginRequest) => request<TokenPair>("/api/v1/auth/login", { auth: false, body, method: "POST" }),
    logout: () => request<{ loggedOut: boolean }>("/api/v1/auth/logout", { method: "POST", retryAuth: false }),
    refreshSession: (refreshToken: string) =>
      request<TokenPair>("/api/v1/auth/refresh", { auth: false, body: { refreshToken }, method: "POST", retryAuth: false }),
    resetPassword: (body: { newPassword: string, token: string }) =>
      request<{ reset: boolean }>("/api/v1/auth/password/reset", { auth: false, body, method: "POST" }),
    revokeInvitation: (orgId: ID, invitationId: ID) =>
      request<{ revoked: boolean }>(`/api/v1/orgs/${orgId}/invitations/${invitationId}`, { method: "DELETE" }),
    revokeAPIToken: (orgId: ID, tokenId: ID) =>
      request<{ revoked: boolean }>(`/api/v1/orgs/${orgId}/api-tokens/${tokenId}`, { method: "DELETE" }),
    revokeSession: (orgId: ID, sessionId: ID) =>
      request<{ revoked: boolean }>(`/api/v1/orgs/${orgId}/sessions/${sessionId}`, { method: "DELETE" }),
    setupMFA: () => request<MFASetupPayload>("/api/v1/auth/mfa/setup", { method: "POST" }),
    signup: (body: SignupRequest) => request<TokenPair>("/api/v1/auth/signup", { auth: false, body, method: "POST" }),
    switchOrg: (orgId: ID) => request<TokenPair>("/api/v1/auth/switch-org", { body: { orgId }, method: "POST" }),
    updateOrganization: (orgId: ID, body: { name: string }) =>
      request<Organization>(`/api/v1/orgs/${orgId}`, { body, method: "PATCH" }),
    updateRole: (orgId: ID, roleId: ID, body: { description?: string, name?: string, permissions?: string[] }) =>
      request<Role>(`/api/v1/orgs/${orgId}/roles/${roleId}`, { body, method: "PATCH" }),
    updateTodo: (id: number, body: { completed?: boolean, description?: string, title?: string }) =>
      request<Todo>(`/api/v1/demo/todos/${id}`, { body, method: "PUT" }),
    updateUser: (orgId: ID, userId: ID, body: { roles?: string[], status?: string }) =>
      request<OrganizationUser>(`/api/v1/orgs/${orgId}/users/${userId}`, { body, method: "PATCH" }),
    verifyMFA: (code: string) => request<{ verified: boolean }>("/api/v1/auth/mfa/verify", { body: { code }, method: "POST" })
  }
}

function pluginProxyEndpoint(pluginId: string, path: string) {
  const cleanPath = path.startsWith("/") ? path : `/${path}`
  return `/api/v1/plugins/${encodeURIComponent(pluginId)}/proxy${cleanPath}`
}

function resolveEndpointURL(endpoint: string, baseURL: string) {
  if (!baseURL) {
    return endpoint
  }

  return new URL(endpoint, baseURL).toString()
}

async function readErrorPayload(response: Response): Promise<Result<unknown> | undefined> {
  try {
    return await response.json() as Result<unknown>
  } catch {
    return undefined
  }
}

function filenameFromContentDisposition(value: string | null) {
  if (!value) {
    return ""
  }

  const utf8Match = value.match(/filename\*=UTF-8''([^;]+)/i)
  if (utf8Match?.[1]) {
    try {
      return decodeURIComponent(utf8Match[1])
    } catch {
      return utf8Match[1]
    }
  }

  const quoted = value.match(/filename="([^"]+)"/i)
  if (quoted?.[1]) {
    return quoted[1]
  }

  const plain = value.match(/filename=([^;]+)/i)
  return plain?.[1]?.trim() || ""
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




