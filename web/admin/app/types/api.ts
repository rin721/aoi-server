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

export type LoginRequest = {
  identifier: string
  mfaCode?: string
  orgCode?: string
  password: string
}

export type SwitchOrgRequest = {
  orgId: number
}

export type TokenPair = {
  accessExpiresAt: string
  accessToken: string
  refreshExpiresAt: string
  refreshToken: string
}

export type User = {
  id: number
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
  id: number
  code: string
  name: string
  status: Status
  createdAt: string
  updatedAt: string
}

export type OrganizationUser = {
  roles: string[]
  user: User
}

export type Role = {
  id: number
  orgId: number
  code: string
  name: string
  description: string
  system: boolean
  createdAt: string
  updatedAt: string
}

export type Permission = {
  id: number
  code: string
  name: string
  description: string
  createdAt: string
  updatedAt: string
}

export type Session = {
  id: number
  userId: number
  orgId: number
  userAgent: string
  ipAddress: string
  expiresAt: string
  revokedAt?: string | null
  lastUsedAt?: string | null
  createdAt: string
  updatedAt: string
}

export type AuditLog = {
  id: number
  orgId?: number | null
  userId?: number | null
  action: string
  resource: string
  resourceId: string
  ipAddress: string
  userAgent: string
  metadata: string
  createdAt: string
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
