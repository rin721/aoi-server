import { defineStore } from "pinia"
import type { ID, InitialAdminSetupRequest, Organization, TokenPair, User } from "~/types/admin"

export const useAuthStore = defineStore("auth", () => {
  const api = useAdminApi()
  const accessToken = ref("")
  const refreshToken = ref("")
  const accessExpiresAt = ref("")
  const refreshExpiresAt = ref("")
  const user = ref<User | null>(null)
  const orgs = ref<Organization[]>([])
  const currentOrgId = ref<ID | null>(null)
  const hydrated = ref(false)
  const loading = ref(false)

  const authenticated = computed(() => Boolean(accessToken.value && user.value))
  const currentOrg = computed(() => orgs.value.find((org) => org.id === currentOrgId.value) || orgs.value[0] || null)

  async function fetchSession() {
    loading.value = true
    try {
      const stored = loadStoredTokenPair()
      if (!stored) {
        clearSession()
        return
      }

      applyTokenPair(stored)
      await loadIdentity()
    } finally {
      hydrated.value = true
      loading.value = false
    }
  }

  async function login(payload: { identifier: string, mfaCode?: string, orgCode?: string, password: string }) {
    loading.value = true
    try {
      applyTokenPair(await api.login(payload))
      await loadIdentity()
    } finally {
      hydrated.value = true
      loading.value = false
    }
  }

  async function signup(payload: { displayName?: string, email: string, orgCode: string, orgName: string, password: string, username: string }) {
    loading.value = true
    try {
      applyTokenPair(await api.signup(payload))
      await loadIdentity()
    } finally {
      hydrated.value = true
      loading.value = false
    }
  }

  async function initialAdminSetup(payload: InitialAdminSetupRequest) {
    loading.value = true
    try {
      applyTokenPair(await api.initialAdminSetup(payload))
      await loadIdentity()
    } finally {
      hydrated.value = true
      loading.value = false
    }
  }

  async function logout() {
    loading.value = true
    try {
      await api.logout()
    } catch {
      // 前端会话始终清理，后端会话已过期或网络抖动不阻断退出。
    } finally {
      clearSession()
      hydrated.value = true
      loading.value = false
    }
  }

  async function switchOrg(orgId: ID) {
    loading.value = true
    try {
      applyTokenPair(await api.switchOrg(orgId))
      await loadIdentity()
    } finally {
      hydrated.value = true
      loading.value = false
    }
  }

  async function refreshTokens() {
    const storedRefreshToken = refreshToken.value || getStoredRefreshToken()
    if (!storedRefreshToken) {
      clearSession()
      return false
    }

    applyTokenPair(await api.refreshSession(storedRefreshToken))
    return true
  }

  async function loadIdentity() {
    const [me, organizations] = await Promise.all([
      api.getMe(),
      api.listMyOrganizations()
    ])
    const latestPair = loadStoredTokenPair()
    if (latestPair) {
      applyTokenPair(latestPair, false)
    }

    user.value = me
    orgs.value = organizations
    currentOrgId.value = currentOrgIdFromToken(accessToken.value) || organizations[0]?.id || null
  }

  function applyTokenPair(pair: TokenPair, persist = true) {
    accessToken.value = pair.accessToken
    refreshToken.value = pair.refreshToken
    accessExpiresAt.value = pair.accessExpiresAt
    refreshExpiresAt.value = pair.refreshExpiresAt
    currentOrgId.value = currentOrgIdFromToken(pair.accessToken)

    if (persist) {
      saveStoredTokenPair(pair)
    }
  }

  function clearSession() {
    accessToken.value = ""
    refreshToken.value = ""
    accessExpiresAt.value = ""
    refreshExpiresAt.value = ""
    user.value = null
    orgs.value = []
    currentOrgId.value = null
    clearStoredTokenPair()
  }

  return {
    accessExpiresAt,
    accessToken,
    authenticated,
    clearSession,
    currentOrg,
    currentOrgId,
    fetchSession,
    hydrated,
    initialAdminSetup,
    loading,
    login,
    logout,
    orgs,
    refreshExpiresAt,
    refreshToken,
    refreshTokens,
    signup,
    switchOrg,
    user
  }
})




