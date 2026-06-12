import type { TokenPair } from "~/types/admin"

const storageKey = "aoi-admin-session"

type AccessTokenClaims = {
  exp?: number
  orgId?: string | number
  sessionId?: string | number
  userId?: string | number
}

export type StoredAdminSession = TokenPair & {
  savedAt: string
}

export function loadStoredTokenPair(): TokenPair | null {
  if (!import.meta.client) {
    return null
  }

  const raw = sessionStorage.getItem(storageKey)
  if (!raw) {
    return null
  }

  try {
    const parsed = JSON.parse(raw) as StoredAdminSession
    if (!parsed.accessToken || !parsed.refreshToken) {
      clearStoredTokenPair()
      return null
    }
    return parsed
  } catch {
    clearStoredTokenPair()
    return null
  }
}

export function saveStoredTokenPair(pair: TokenPair) {
  if (!import.meta.client) {
    return
  }

  const stored: StoredAdminSession = {
    ...pair,
    savedAt: new Date().toISOString()
  }
  sessionStorage.setItem(storageKey, JSON.stringify(stored))
}

export function clearStoredTokenPair() {
  if (import.meta.client) {
    sessionStorage.removeItem(storageKey)
  }
}

export function getStoredAccessToken() {
  return loadStoredTokenPair()?.accessToken || ""
}

export function getStoredRefreshToken() {
  return loadStoredTokenPair()?.refreshToken || ""
}

export function decodeAccessTokenClaims(accessToken = getStoredAccessToken()): AccessTokenClaims | null {
  const [, payload] = accessToken.split(".")
  if (!payload) {
    return null
  }

  try {
    const json = decodeBase64URL(payload)
    return JSON.parse(json) as AccessTokenClaims
  } catch {
    return null
  }
}

export function currentOrgIdFromToken(accessToken = getStoredAccessToken()) {
  const orgId = decodeAccessTokenClaims(accessToken)?.orgId
  return typeof orgId === "string" ? orgId : null
}

export function currentSessionIdFromToken(accessToken = getStoredAccessToken()) {
  const sessionId = decodeAccessTokenClaims(accessToken)?.sessionId
  return typeof sessionId === "string" ? sessionId : null
}

function decodeBase64URL(value: string) {
  const normalized = value.replace(/-/g, "+").replace(/_/g, "/")
  const padded = normalized.padEnd(normalized.length + ((4 - normalized.length % 4) % 4), "=")
  return decodeURIComponent(
    atob(padded)
      .split("")
      .map((char) => `%${char.charCodeAt(0).toString(16).padStart(2, "0")}`)
      .join("")
  )
}




