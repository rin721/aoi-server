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
  identifier: string
  mfaCode?: string
  orgCode?: string
  password: string
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

export type OrganizationUser = {
  membershipStatus: Status
  roles: string[]
  user: User
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

export type MFASetupPayload = {
  otpauthUrl: string
  secret: string
}

export type Status = "active" | "disabled" | "pending" | "used" | "revoked"



