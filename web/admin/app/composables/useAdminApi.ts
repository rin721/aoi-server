import type {
  ApiErrorPayload,
  APITokenPage,
  CaptchaChallenge,
  DemoCustomer,
  DemoCustomerPage,
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
  OrganizationPage,
  OrganizationUser,
  OrganizationUserPage,
  Permission,
  PluginHealthStatus,
  PluginManifest,
  ReadyStatus,
  Result,
  Role,
  Session,
  SessionPage,
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
  SystemMediaResumableAbortResult,
  SystemMediaResumableCheckResult,
  SystemMediaResumableChunkResult,
  SystemMediaResumableCompleteResult,
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
import { ADMIN_API_ENDPOINTS } from "~/config/admin-api"

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
      const pair = await requestWithAuthRetry<TokenPair>(ADMIN_API_ENDPOINTS.auth.refresh, {
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
      request<{ email?: string, orgId?: ID, sessionId?: ID, userId?: ID }>(ADMIN_API_ENDPOINTS.invitations.accept(token), {
        auth: false,
        body,
        method: "POST"
      }),
    createOrganization: (body: { code: string, name: string }) =>
      request<Organization>(ADMIN_API_ENDPOINTS.orgs.collection, { body, method: "POST" }),
    createAPIToken: (orgId: ID, body: { days?: number, remark?: string, roleCode: string, userId: ID }) =>
      request<CreateAPITokenResult>(ADMIN_API_ENDPOINTS.orgs.apiTokens(orgId), { body, method: "POST" }),
    createRole: (orgId: ID, body: { code: string, description?: string, name: string, permissions: string[] }) =>
      request<Role>(ADMIN_API_ENDPOINTS.orgs.roles(orgId), { body, method: "POST" }),
    createTodo: (body: { completed?: boolean, description?: string, title: string }) =>
      request<Todo>(ADMIN_API_ENDPOINTS.demo.todos, { body, method: "POST" }),
    createDemoCustomer: (body: { customerName: string, customerPhoneData: string }) =>
      request<DemoCustomer>(ADMIN_API_ENDPOINTS.demo.customers, { body, method: "POST" }),
    deleteDemoCustomer: (id: number) =>
      request<{ deleted: boolean }>(ADMIN_API_ENDPOINTS.demo.customer(id), { method: "DELETE" }),
    deleteTodo: (id: number) => request<{ deleted: boolean }>(ADMIN_API_ENDPOINTS.demo.todo(id), { method: "DELETE" }),
    forgotPassword: (email: string) =>
      request<NotificationDelivery>(ADMIN_API_ENDPOINTS.auth.forgotPassword, { auth: false, body: { email }, method: "POST" }),
    getAuthCaptcha: () => request<CaptchaChallenge>(ADMIN_API_ENDPOINTS.auth.captcha, { auth: false }),
    getHealth: () => request<HealthStatus>(ADMIN_API_ENDPOINTS.health, { auth: false }),
    getMe: () => request<User>(ADMIN_API_ENDPOINTS.me.profile),
    getReady: () => request<ReadyStatus>(ADMIN_API_ENDPOINTS.ready, { auth: false }),
    getSetupStatus: () => request<SetupStatus>(ADMIN_API_ENDPOINTS.auth.setupStatus, { auth: false }),
    getSystemConfig: () => request<SystemConfigSnapshot>(ADMIN_API_ENDPOINTS.system.config),
    updateSystemConfig: (items: Array<{ key: string, value: unknown }>, options: { persist?: boolean } = {}) =>
      request<SystemConfigSnapshot>(ADMIN_API_ENDPOINTS.system.config, { body: { items, persist: Boolean(options.persist) }, method: "PATCH" }),
    getSystemServerInfo: () => request<SystemServerInfo>(ADMIN_API_ENDPOINTS.system.serverInfo),
    initialAdminSetup: (body: InitialAdminSetupRequest) =>
      request<TokenPair>(ADMIN_API_ENDPOINTS.auth.initialAdminSetup, { auth: false, body, method: "POST" }),
    inviteUser: (orgId: ID, body: { email: string, roleCode: string }) =>
      request<NotificationDelivery>(ADMIN_API_ENDPOINTS.orgs.userInvitations(orgId), { body, method: "POST" }),
    listAuditLogs: (orgId: ID, query: AuditLogQuery | number = 100) =>
      request<AuditLog[]>(ADMIN_API_ENDPOINTS.orgs.auditLogs(orgId), { query: typeof query === "number" ? { limit: query } : query }),
    listInvitations: (orgId: ID) => request<Invitation[]>(ADMIN_API_ENDPOINTS.orgs.invitations(orgId)),
    listMyOrganizations: () => request<Organization[]>(ADMIN_API_ENDPOINTS.me.organizations),
    listOrganizations: (query: { code?: string, desc?: boolean, keyword?: string, name?: string, orderKey?: string, page?: number, pageSize?: number, status?: string } = {}) =>
      request<OrganizationPage>(ADMIN_API_ENDPOINTS.orgs.collection, { query }),
    listPermissions: (orgId: ID) => request<Permission[]>(ADMIN_API_ENDPOINTS.orgs.permissions(orgId)),
    listAPITokens: (orgId: ID, query: { page?: number, pageSize?: number, status?: string, userId?: ID | null } = {}) =>
      request<APITokenPage>(ADMIN_API_ENDPOINTS.orgs.apiTokens(orgId), { query }),
    getPlugin: (pluginId: string) => request<PluginManifest>(ADMIN_API_ENDPOINTS.plugins.item(pluginId)),
    getPluginHealth: (pluginId: string) => request<PluginHealthStatus>(ADMIN_API_ENDPOINTS.plugins.health(pluginId)),
    listPlugins: () => request<PluginManifest[]>(ADMIN_API_ENDPOINTS.plugins.collection),
    proxyPlugin: <T = unknown>(pluginId: string, path: string, options: PluginProxyOptions = {}) =>
      request<T>(pluginProxyEndpoint(pluginId, path), {
        body: options.body,
        method: options.method || "GET",
        query: options.query
      }),
    listRoles: (orgId: ID) => request<Role[]>(ADMIN_API_ENDPOINTS.orgs.roles(orgId)),
    listSessions: (orgId: ID, query: { desc?: boolean, ipAddress?: string, keyword?: string, orderKey?: string, page?: number, pageSize?: number, scope?: string, status?: string, userId?: ID | null } = {}) =>
      request<SessionPage>(ADMIN_API_ENDPOINTS.orgs.sessions(orgId), { query }),
    listSystemAPIs: () => request<SystemAPIGroup[]>(ADMIN_API_ENDPOINTS.system.apis),
    listSystemDictionaries: () => request<SystemDictionaryCatalog>(ADMIN_API_ENDPOINTS.system.dictionaries),
    listSystemMenus: () => request<SystemMenuGroup[]>(ADMIN_API_ENDPOINTS.system.menus),
    listSystemOperationRecords: (query: { method?: string, page?: number, pageSize?: number, path?: string, status?: number | string, statusClass?: string } = {}) =>
      request<SystemOperationRecordPage>(ADMIN_API_ENDPOINTS.system.operationRecords, { query }),
    listSystemParameters: (query: { endCreatedAt?: string, key?: string, name?: string, page?: number, pageSize?: number, startCreatedAt?: string } = {}) =>
      request<SystemParameterPage>(ADMIN_API_ENDPOINTS.system.parameters, { query }),
    listSystemVersionSources: () => request<SystemVersionSourceCatalog>(ADMIN_API_ENDPOINTS.system.versionSources),
    listSystemVersions: (query: { endCreatedAt?: string, page?: number, pageSize?: number, startCreatedAt?: string, versionCode?: string, versionName?: string } = {}) =>
      request<SystemVersionPage>(ADMIN_API_ENDPOINTS.system.versions, { query }),
    createSystemDictionary: (body: { code: string, description?: string, name: string, status?: string }) =>
      request<SystemDictionary>(ADMIN_API_ENDPOINTS.system.dictionaries, { body, method: "POST" }),
    createSystemParameter: (body: { description?: string, key: string, name: string, value: string }) =>
      request<SystemParameter>(ADMIN_API_ENDPOINTS.system.parameters, { body, method: "POST" }),
    exportSystemVersion: (body: { apiCodes: string[], description?: string, dictionaryCodes: string[], menuCodes: string[], versionCode: string, versionName: string }) =>
      request<SystemVersionDetail>(ADMIN_API_ENDPOINTS.system.versionExport, { body, method: "POST" }),
    importSystemVersion: (versionData: string) =>
      request<SystemVersionImportResult>(ADMIN_API_ENDPOINTS.system.versionImport, { body: { versionData }, method: "POST" }),
    importSystemMediaURLs: (body: { categoryId?: ID | number, items?: Array<{ name?: string, url: string }>, text?: string }) =>
      request<SystemMediaURLImportResult>(ADMIN_API_ENDPOINTS.system.media.importURL, { body, method: "POST" }),
    abortSystemMediaResumableUpload: (body: { fileHash: string, sessionId: ID | number }) =>
      request<SystemMediaResumableAbortResult>(ADMIN_API_ENDPOINTS.system.media.resumableAbort, { body, method: "POST" }),
    checkSystemMediaResumableUpload: (body: { categoryId?: ID | number, chunkSize?: number, chunkTotal?: number, fileHash: string, fileName: string, sizeBytes: number }) =>
      request<SystemMediaResumableCheckResult>(ADMIN_API_ENDPOINTS.system.media.resumableCheck, { body, method: "POST" }),
    completeSystemMediaResumableUpload: (body: { fileHash: string, sessionId: ID | number }) =>
      request<SystemMediaResumableCompleteResult>(ADMIN_API_ENDPOINTS.system.media.resumableComplete, { body, method: "POST" }),
    createSystemDictionaryItem: (dictionaryId: ID, body: { extra?: string, label: string, sort?: number, status?: string, value: string }) =>
      request<SystemDictionaryItem>(ADMIN_API_ENDPOINTS.system.dictionaryItems(dictionaryId), { body, method: "POST" }),
    saveSystemMediaCategory: (body: { id?: ID | number, name: string, parentId?: ID | number, sort?: number }) =>
      request<SystemMediaCategory>(ADMIN_API_ENDPOINTS.system.media.categories, { body, method: "POST" }),
    deleteSystemVersion: (versionId: ID) =>
      request<{ deleted: boolean }>(ADMIN_API_ENDPOINTS.system.version(versionId), { method: "DELETE" }),
    deleteSystemVersions: (ids: ID[]) =>
      request<{ deleted: boolean }>(ADMIN_API_ENDPOINTS.system.versions, { body: { ids }, method: "DELETE" }),
    downloadSystemVersion: (versionId: ID) =>
      request<SystemVersionPackage>(ADMIN_API_ENDPOINTS.system.versionDownload(versionId)),
    downloadSystemMediaAsset: (assetId: ID) =>
      downloadWithAuthRetry(ADMIN_API_ENDPOINTS.system.media.assetDownload(assetId)),
    deleteSystemMediaAsset: (assetId: ID) =>
      request<{ deleted: boolean }>(ADMIN_API_ENDPOINTS.system.media.asset(assetId), { method: "DELETE" }),
    deleteSystemMediaCategory: (categoryId: ID) =>
      request<{ deleted: boolean }>(ADMIN_API_ENDPOINTS.system.media.category(categoryId), { method: "DELETE" }),
    deleteSystemDictionary: (dictionaryId: ID) =>
      request<{ deleted: boolean }>(ADMIN_API_ENDPOINTS.system.dictionary(dictionaryId), { method: "DELETE" }),
    deleteSystemDictionaryItem: (itemId: ID) =>
      request<{ deleted: boolean }>(ADMIN_API_ENDPOINTS.system.dictionaryItem(itemId), { method: "DELETE" }),
    deleteSystemOperationRecords: (ids: ID[]) =>
      request<{ deleted: boolean }>(ADMIN_API_ENDPOINTS.system.operationRecords, { body: { ids }, method: "DELETE" }),
    deleteSystemParameter: (parameterId: ID) =>
      request<{ deleted: boolean }>(ADMIN_API_ENDPOINTS.system.parameter(parameterId), { method: "DELETE" }),
    deleteSystemParameters: (ids: ID[]) =>
      request<{ deleted: boolean }>(ADMIN_API_ENDPOINTS.system.parameters, { body: { ids }, method: "DELETE" }),
    getSystemParameterByKey: (key: string) =>
      request<SystemParameter>(ADMIN_API_ENDPOINTS.system.parameterValue, { query: { key } }),
    getSystemVersion: (versionId: ID) =>
      request<SystemVersionDetail>(ADMIN_API_ENDPOINTS.system.version(versionId)),
    listSystemMediaAssets: (query: { categoryId?: ID | number, keyword?: string, page?: number, pageSize?: number } = {}) =>
      request<SystemMediaAssetPage>(ADMIN_API_ENDPOINTS.system.media.assets, { query }),
    listSystemMediaCategories: () => request<SystemMediaCategoryCatalog>(ADMIN_API_ENDPOINTS.system.media.categories),
    syncSystemAPIs: () => request<SystemAPISyncResult>(ADMIN_API_ENDPOINTS.system.apisSync, { method: "POST" }),
    syncSystemAPIPermissions: () => request<SystemPermissionSyncResult>(ADMIN_API_ENDPOINTS.system.apiPermissionsSync, { method: "POST" }),
    updateSystemDictionary: (dictionaryId: ID, body: { description?: string, name?: string, status?: string }) =>
      request<SystemDictionary>(ADMIN_API_ENDPOINTS.system.dictionary(dictionaryId), { body, method: "PATCH" }),
    updateSystemDictionaryItem: (itemId: ID, body: { extra?: string, label?: string, sort?: number, status?: string, value?: string }) =>
      request<SystemDictionaryItem>(ADMIN_API_ENDPOINTS.system.dictionaryItem(itemId), { body, method: "PATCH" }),
    updateSystemMediaAsset: (assetId: ID, body: { displayName: string }) =>
      request<SystemMediaAsset>(ADMIN_API_ENDPOINTS.system.media.asset(assetId), { body, method: "PATCH" }),
    updateSystemParameter: (parameterId: ID, body: { description?: string, key?: string, name?: string, value?: string }) =>
      request<SystemParameter>(ADMIN_API_ENDPOINTS.system.parameter(parameterId), { body, method: "PATCH" }),
    uploadSystemMediaAsset: (file: File, categoryId?: ID | number) => {
      const body = new FormData()
      body.append("file", file)
      if (categoryId) {
        body.append("categoryId", String(categoryId))
      }
      return request<SystemMediaAsset>(ADMIN_API_ENDPOINTS.system.media.assetUpload, { body, method: "POST" })
    },
    uploadSystemMediaChunk: (file: Blob, metadata: { chunkHash: string, chunkIndex: number, chunkTotal: number, fileHash: string, fileName: string, sessionId: ID | number }) => {
      const body = new FormData()
      body.append("file", file)
      body.append("chunkHash", metadata.chunkHash)
      body.append("chunkIndex", String(metadata.chunkIndex))
      body.append("chunkTotal", String(metadata.chunkTotal))
      body.append("fileHash", metadata.fileHash)
      body.append("fileName", metadata.fileName)
      body.append("sessionId", String(metadata.sessionId))
      return request<SystemMediaResumableChunkResult>(ADMIN_API_ENDPOINTS.system.media.resumableChunks, { body, method: "POST" })
    },
    getDemoCustomer: (id: number) => request<DemoCustomer>(ADMIN_API_ENDPOINTS.demo.customer(id)),
    listDemoCustomers: (query: { keyword?: string, page?: number, pageSize?: number } = {}) =>
      request<DemoCustomerPage>(ADMIN_API_ENDPOINTS.demo.customers, { query }),
    listTodos: () => request<Todo[]>(ADMIN_API_ENDPOINTS.demo.todos),
    listUsers: (orgId: ID, query: { desc?: boolean, displayName?: string, email?: string, keyword?: string, orderKey?: string, page?: number, pageSize?: number, roleCode?: string, status?: string, username?: string } = {}) =>
      request<OrganizationUserPage>(ADMIN_API_ENDPOINTS.orgs.users(orgId), { query }),
    login: (body: LoginRequest) => request<TokenPair>(ADMIN_API_ENDPOINTS.auth.login, { auth: false, body, method: "POST" }),
    logout: () => request<{ loggedOut: boolean }>(ADMIN_API_ENDPOINTS.auth.logout, { method: "POST", retryAuth: false }),
    refreshSession: (refreshToken: string) =>
      request<TokenPair>(ADMIN_API_ENDPOINTS.auth.refresh, { auth: false, body: { refreshToken }, method: "POST", retryAuth: false }),
    resetPassword: (body: { newPassword: string, token: string }) =>
      request<{ reset: boolean }>(ADMIN_API_ENDPOINTS.auth.passwordReset, { auth: false, body, method: "POST" }),
    revokeInvitation: (orgId: ID, invitationId: ID) =>
      request<{ revoked: boolean }>(ADMIN_API_ENDPOINTS.orgs.invitation(orgId, invitationId), { method: "DELETE" }),
    revokeAPIToken: (orgId: ID, tokenId: ID) =>
      request<{ revoked: boolean }>(ADMIN_API_ENDPOINTS.orgs.apiToken(orgId, tokenId), { method: "DELETE" }),
    revokeSession: (orgId: ID, sessionId: ID) =>
      request<{ revoked: boolean }>(ADMIN_API_ENDPOINTS.orgs.session(orgId, sessionId), { method: "DELETE" }),
    setupMFA: () => request<MFASetupPayload>(ADMIN_API_ENDPOINTS.auth.mfaSetup, { method: "POST" }),
    signup: (body: SignupRequest) => request<TokenPair>(ADMIN_API_ENDPOINTS.auth.signup, { auth: false, body, method: "POST" }),
    switchOrg: (orgId: ID) => request<TokenPair>(ADMIN_API_ENDPOINTS.auth.switchOrg, { body: { orgId }, method: "POST" }),
    updateOrganization: (orgId: ID, body: { name: string }) =>
      request<Organization>(ADMIN_API_ENDPOINTS.orgs.item(orgId), { body, method: "PATCH" }),
    updateRole: (orgId: ID, roleId: ID, body: { description?: string, name?: string, permissions?: string[] }) =>
      request<Role>(ADMIN_API_ENDPOINTS.orgs.role(orgId, roleId), { body, method: "PATCH" }),
    updateDemoCustomer: (id: number, body: { customerName?: string, customerPhoneData?: string }) =>
      request<DemoCustomer>(ADMIN_API_ENDPOINTS.demo.customer(id), { body, method: "PATCH" }),
    updateTodo: (id: number, body: { completed?: boolean, description?: string, title?: string }) =>
      request<Todo>(ADMIN_API_ENDPOINTS.demo.todo(id), { body, method: "PUT" }),
    updateUser: (orgId: ID, userId: ID, body: { roles?: string[], status?: string }) =>
      request<OrganizationUser>(ADMIN_API_ENDPOINTS.orgs.user(orgId, userId), { body, method: "PATCH" }),
    verifyMFA: (code: string) => request<{ verified: boolean }>(ADMIN_API_ENDPOINTS.auth.mfaVerify, { body: { code }, method: "POST" })
  }
}

function pluginProxyEndpoint(pluginId: string, path: string) {
  return ADMIN_API_ENDPOINTS.plugins.proxy(pluginId, path)
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
