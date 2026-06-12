export type Result<T = unknown> = {
  code: number
  data?: T
  message: string
  traceId?: string
  serverTime: number
}

export type ApiErrorPayload = {
  code: number | string
  endpoint: string
  message: string
  serverTime?: number
  statusCode: number
  traceId?: string
}

export type HealthStatus = {
  status: "ok"
}

export type ReadyStatus = {
  checks: Record<string, string>
  status: "ready" | "not_ready"
}

export type ID = string

export type LoginRequest = {
  captchaCode?: string
  captchaId?: string
  identifier: string
  mfaCode?: string
  orgCode?: string
  password: string
}

export type CaptchaChallenge = {
  captchaId?: string
  enabled: boolean
  expiresAt?: string
  image?: string
}

export type SignupRequest = {
  displayName?: string
  email: string
  orgCode: string
  orgName: string
  password: string
  username: string
}

export type InitialAdminSetupRequest = SignupRequest

export type SetupStatus = {
  required: boolean
}

export type SystemMenuItem = {
  code: string
  icon: string
  label: string
  mobile: boolean
  order: number
  path: string
  permission?: string
}

export type SystemMenuGroup = {
  code: string
  items: SystemMenuItem[]
  label: string
  order: number
}

export type SystemAPIEntry = {
  access: "authenticated" | "permission" | "public"
  code: string
  description: string
  group: string
  method: string
  order: number
  path: string
  permission?: string
  permissionRegistered: boolean
  synced: boolean
  syncedAt?: string
}

export type SystemAPIGroup = {
  code: string
  count: number
  items: SystemAPIEntry[]
  label: string
}

export type SystemConfigItem = {
  description: string
  key: string
  label: string
  secret: boolean
  source: string
  value: unknown
}

export type SystemConfigSection = {
  code: string
  description: string
  icon: string
  items: SystemConfigItem[]
  label: string
  order: number
}

export type SystemConfigSnapshot = {
  sections: SystemConfigSection[]
}

export type SystemServerOSInfo = {
  compiler: string
  goarch: string
  goos: string
  goVersion: string
  numCpu: number
  numGoroutine: number
}

export type SystemServerRuntimeInfo = {
  startTime: string
  uptime: string
  uptimeSeconds: number
}

export type SystemServerCPUInfo = {
  cores: number
  percent: number[]
}

export type SystemServerRAMInfo = {
  totalMb: number
  usedMb: number
  usedPercent: number
}

export type SystemServerDiskInfo = {
  fsType: string
  mountPoint: string
  totalGb: number
  totalMb: number
  usedGb: number
  usedMb: number
  usedPercent: number
}

export type SystemServerMemoryInfo = {
  allocMb: number
  heapAllocMb: number
  heapIdleMb: number
  heapInuseMb: number
  heapObjects: number
  heapReleasedMb: number
  heapSysMb: number
  stackInuseMb: number
  stackSysMb: number
  sysMb: number
  totalAllocMb: number
}

export type SystemServerGCInfo = {
  lastGcAt?: string
  nextGcMb: number
  numGc: number
  pauseTotalNs: number
}

export type SystemServerBuildInfo = {
  goVersion: string
  module: string
  path: string
  settings: Array<{ key: string, value: string }>
  version: string
}

export type SystemServerInfo = {
  build: SystemServerBuildInfo
  cpu: SystemServerCPUInfo
  disk: SystemServerDiskInfo[]
  gc: SystemServerGCInfo
  memory: SystemServerMemoryInfo
  os: SystemServerOSInfo
  ram: SystemServerRAMInfo
  refreshedAt: string
  runtime: SystemServerRuntimeInfo
}

export type SystemAPISyncResult = {
  created: number
  groups: SystemAPIGroup[]
  persisted: boolean
  stale: number
  storageStatus: string
  syncedAt: string
  total: number
  updated: number
}

export type SystemPermissionSyncItem = {
  code: string
  created: boolean
  description: string
  exists: boolean
  name: string
}

export type SystemPermissionSyncResult = {
  created: number
  items: SystemPermissionSyncItem[]
  persisted: boolean
  skipped: number
  storageStatus: string
  syncedAt: string
  total: number
}

export type SystemDictionaryItem = {
  id: ID
  dictionaryId: ID
  extra: string
  label: string
  sort: number
  status: Status
  value: string
  createdAt: string
  updatedAt: string
}

export type SystemDictionary = {
  id: ID
  code: string
  description: string
  name: string
  status: Status
  items: SystemDictionaryItem[]
  createdAt: string
  updatedAt: string
}

export type SystemDictionaryCatalog = {
  items: SystemDictionary[]
  storageStatus: string
  total: number
}

export type SystemOperationRecord = {
  id: ID
  userId: ID
  username: string
  ipAddress: string
  method: string
  path: string
  status: number
  latencyMs: number
  userAgent: string
  errorMessage: string
  body: string
  response: string
  traceId: string
  createdAt: string
}

export type SystemOperationRecordPage = {
  items: SystemOperationRecord[]
  page: number
  pageSize: number
  storageStatus: string
  total: number
}

export type SystemParameter = {
  id: ID
  name: string
  key: string
  value: string
  description: string
  createdAt: string
  updatedAt: string
}

export type SystemParameterPage = {
  items: SystemParameter[]
  page: number
  pageSize: number
  storageStatus: string
  total: number
}

export type SystemVersionRecord = {
  id: ID
  apiCount: number
  createdAt: string
  createdBy: ID
  createdByUsername: string
  description: string
  dictionaryCount: number
  menuCount: number
  source: "export" | "import"
  updatedAt: string
  versionCode: string
  versionName: string
}

export type SystemVersionPage = {
  items: SystemVersionRecord[]
  page: number
  pageSize: number
  storageStatus: string
  total: number
}

export type SystemVersionPackageInfo = {
  code: string
  description: string
  exportTime: string
  name: string
}

export type SystemVersionPackage = {
  apis: SystemAPIEntry[]
  dictionaries: SystemDictionary[]
  menus: SystemMenuGroup[]
  version: SystemVersionPackageInfo
}

export type SystemVersionDetail = {
  item: SystemVersionRecord
  package: SystemVersionPackage
}

export type SystemVersionSourceCatalog = {
  apiCount: number
  apis: SystemAPIGroup[]
  dictionaries: SystemDictionary[]
  dictionaryCount: number
  menuCount: number
  menus: SystemMenuGroup[]
  storageStatus: string
}

export type SystemVersionImportResult = {
  apisSkipped: number
  dictionariesCreated: number
  dictionariesSkipped: number
  dictionaryItemsCreated: number
  importedAt: string
  item: SystemVersionRecord
  menusSkipped: number
  storageStatus: string
}

export type SystemMediaCategory = {
  id: ID
  parentId: ID
  name: string
  sort: number
  children?: SystemMediaCategory[]
  createdAt: string
  updatedAt: string
}

export type SystemMediaCategoryCatalog = {
  items: SystemMediaCategory[]
  storageStatus: string
  total: number
}

export type SystemMediaAsset = {
  id: ID
  categoryId: ID
  displayName: string
  originalName: string
  storageKey: string
  url: string
  mimeType: string
  extension: string
  sizeBytes: number
  source: "resumable" | "upload" | "url"
  external: boolean
  uploadedBy: ID
  uploadedByUsername: string
  createdAt: string
  updatedAt: string
}

export type SystemMediaAssetPage = {
  items: SystemMediaAsset[]
  objectStorage: string
  page: number
  pageSize: number
  storageStatus: string
  total: number
  uploadMaxBytes: number
  uploadMaxMb: number
  uploadUnavailable: boolean
}

export type SystemMediaURLImportResult = {
  imported: number
  items: SystemMediaAsset[]
  storageStatus: string
}

export type SystemMediaUploadSession = {
  id: ID
  categoryId: ID
  chunkSize: number
  chunkTotal: number
  completedAt?: string
  createdAt: string
  displayName: string
  expiresAt: string
  extension: string
  fileHash: string
  fileName: string
  finalAssetId: ID
  mimeType: string
  sizeBytes: number
  status: "aborted" | "active" | "completed" | "expired"
  updatedAt: string
  uploadedBy: ID
  uploadedByUsername: string
}

export type SystemMediaResumableCheckResult = {
  asset?: SystemMediaAsset
  chunkSize: number
  missingChunks: number[]
  objectStorage: string
  progress: number
  session: SystemMediaUploadSession
  storageStatus: string
  uploadMaxBytes: number
  uploadMaxMb: number
  uploadedChunks: number[]
  uploadUnavailable: boolean
}

export type SystemMediaResumableChunkResult = {
  chunkIndex: number
  missingChunks: number[]
  progress: number
  sessionId: ID
  status: string
  storageStatus: string
  uploadedChunks: number[]
}

export type SystemMediaResumableCompleteResult = {
  asset: SystemMediaAsset
  sessionId: ID
  storageStatus: string
}

export type SystemMediaResumableAbortResult = {
  sessionId: ID
  status: string
  storageStatus: string
}

export type SwitchOrgRequest = {
  orgId: ID
}

export type TokenPair = {
  accessExpiresAt: string
  accessToken: string
  refreshExpiresAt: string
  refreshToken: string
}

export type User = {
  id: ID
  username: string
  email: string
  displayName: string
  status: Status
  mfaEnabled: boolean
  lockedUntil?: string | null
  lastLoginAt?: string | null
  createdAt: string
  updatedAt: string
}

export type Organization = {
  id: ID
  code: string
  name: string
  status: Status
  createdAt: string
  updatedAt: string
}

export type OrganizationPage = {
  items: Organization[]
  page: number
  pageSize: number
  storageStatus: string
  total: number
}

export type OrganizationUser = {
  membershipStatus: Status
  roles: string[]
  user: User
}

export type OrganizationUserPage = {
  items: OrganizationUser[]
  page: number
  pageSize: number
  storageStatus: string
  total: number
}

export type Role = {
  id: ID
  orgId: ID
  code: string
  name: string
  description: string
  system: boolean
  permissions?: string[]
  createdAt: string
  updatedAt: string
}

export type Permission = {
  id: ID
  code: string
  name: string
  description: string
  createdAt: string
  updatedAt: string
}

export type Session = {
  id: ID
  userId: ID
  orgId: ID
  userAgent: string
  ipAddress: string
  expiresAt: string
  revokedAt?: string | null
  lastUsedAt?: string | null
  createdAt: string
  updatedAt: string
}

export type APIToken = {
  id: ID
  orgId: ID
  userId: ID
  username: string
  userDisplayName: string
  roleCode: string
  tokenPrefix: string
  status: Status
  expiresAt?: string | null
  lastUsedAt?: string | null
  lastUsedIpAddress: string
  revokedAt?: string | null
  revokedBy?: ID | null
  remark: string
  createdBy: ID
  createdAt: string
  updatedAt: string
}

export type APITokenPage = {
  items: APIToken[]
  page: number
  pageSize: number
  storageStatus: string
  total: number
}

export type CreateAPITokenResult = {
  item: APIToken
  token: string
}

export type AuditLog = {
  id: ID
  orgId?: ID | null
  userId?: ID | null
  action: string
  resource: string
  resourceId: string
  ipAddress: string
  userAgent: string
  metadata: string
  createdAt: string
}

export type Invitation = {
  id: ID
  orgId: ID
  email: string
  roleCode: string
  status: Status
  invitedBy: ID
  acceptedBy?: ID | null
  expiresAt: string
  createdAt: string
  updatedAt: string
}

export type NotificationDelivery = {
  token?: string
  url?: string
}

export type PluginManifest = {
  id: string
  name: string
  version: string
  baseURL: string
  healthPath: string
  frontend: {
    entry?: string
  }
  menus: PluginMenu[]
  permissions: PluginPermission[]
  proxy: {
    prefixes: string[]
  }
  secretRef?: string
}

export type PluginMenu = {
  code: string
  label: string
  icon?: string
  path: string
  permission?: string
  order?: number
}

export type PluginPermission = {
  code: string
  name: string
  description?: string
}

export type PluginHealthStatus = {
  id: string
  status: "ok" | "unhealthy"
  statusCode: number
  durationMs: number
  error?: string
}

export type Todo = {
  id: number
  title: string
  description?: string
  completed: boolean
  createdAt: string
  updatedAt: string
}

export type DemoCustomer = {
  id: number
  customerName: string
  customerPhoneData: string
  ownerUserId: ID
  ownerUsername: string
  ownerRoleCode: string
  orgId: ID
  createdAt: string
  updatedAt: string
}

export type DemoCustomerPage = {
  items: DemoCustomer[]
  page: number
  pageSize: number
  storageStatus: string
  total: number
}

export type MFASetupPayload = {
  otpauthUrl: string
  secret: string
}

export type Status = "active" | "disabled" | "expired" | "pending" | "used" | "revoked"



