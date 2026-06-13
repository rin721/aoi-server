<script setup lang="ts">
import { toDataURL } from "qrcode"
import type { AoiKeyValueItem } from "~/types/ui"

const api = useAdminApi()
const auth = useAuthStore()

const secret = ref("")
const otpauthUrl = ref("")
const mfaQrDataUrl = ref("")
const mfaCode = ref("")
const mfaSetupTab = ref<"qr" | "otpauth" | "secret">("qr")
const loading = ref(false)
const verifying = ref(false)
const error = ref("")
const success = ref("")
const mfaSetupTabs = [
  { value: "qr", label: "二维码", icon: "qr-code" },
  { value: "otpauth", label: "otpauth URL", icon: "link" },
  { value: "secret", label: "Secret", icon: "key-round" }
]

const accountName = computed(() => auth.user?.displayName || auth.user?.username || "-")
const mfaEnabled = computed(() => Boolean(auth.user?.mfaEnabled))
const accessExpiresLabel = computed(() => formatDateTime(auth.accessExpiresAt))
const refreshExpiresLabel = computed(() => formatDateTime(auth.refreshExpiresAt))
const securityItems = computed<AoiKeyValueItem[]>(() => [
  { icon: "user", label: "账号", value: accountName.value },
  { icon: "mail", label: "邮箱", value: auth.user?.email || "-" },
  { icon: "building-2", label: "当前组织", value: auth.currentOrg?.name || "-" },
  { icon: "fingerprint", label: "当前会话", value: auth.sessionId || "-", monospace: true },
  { icon: "clock", label: "Access 过期", value: accessExpiresLabel.value },
  { icon: "calendar-clock", label: "Refresh 过期", value: refreshExpiresLabel.value }
])

async function setupMFA() {
  loading.value = true
  error.value = ""
  success.value = ""
  mfaQrDataUrl.value = ""
  try {
    const result = await api.setupMFA()
    secret.value = result.secret
    otpauthUrl.value = result.otpauthUrl
    mfaSetupTab.value = "qr"
    try {
      mfaQrDataUrl.value = await toDataURL(result.otpauthUrl, {
        color: {
          dark: "#0f172a",
          light: "#ffffff"
        },
        errorCorrectionLevel: "M",
        margin: 2,
        width: 240
      })
    } catch {
      mfaSetupTab.value = "otpauth"
      error.value = "二维码生成失败，请使用 otpauth URL 或 Secret 手动录入。"
    }
    success.value = "MFA 密钥已生成，请录入验证器后提交验证码。"
  } catch (err) {
    error.value = errorMessage(err)
  } finally {
    loading.value = false
  }
}

async function verifyMFA() {
  if (!mfaCode.value.trim()) {
    return
  }

  verifying.value = true
  error.value = ""
  success.value = ""
  try {
    await api.verifyMFA(mfaCode.value.trim())
    success.value = "MFA 已启用。"
    secret.value = ""
    otpauthUrl.value = ""
    mfaQrDataUrl.value = ""
    mfaCode.value = ""
    mfaSetupTab.value = "qr"
    await auth.fetchSession()
  } catch (err) {
    error.value = errorMessage(err)
  } finally {
    verifying.value = false
  }
}

async function copyMFAValue(value: string, label: string) {
  if (!value) {
    return
  }
  try {
    if (!navigator.clipboard?.writeText) {
      throw new Error("clipboard unsupported")
    }
    await navigator.clipboard.writeText(value)
    success.value = `${label} 已复制。`
  } catch {
    success.value = `浏览器未允许复制，请手动选中 ${label}。`
  }
}

async function logout() {
  await auth.logout()
  await navigateTo("/login")
}

async function openSessions() {
  await navigateTo("/sessions")
}

useHead({
  title: "安全 - Aoi Admin"
})
</script>

<template>
  <div class="page-grid security-page">
    <PageHeader title="安全" icon="lock-keyhole" description="查看账号安全状态，管理 MFA，并进入会话撤销流程。">
      <template #actions>
        <AoiButton appearance="soft" icon="monitor-check" @click="openSessions">会话管理</AoiButton>
      </template>
    </PageHeader>

    <AoiStatusMessage tone="danger" :message="error" />
    <AoiStatusMessage tone="success" :message="success" />

    <section class="security-workspace">
      <AoiAdminCard
        class="security-account-card"
        :badge="`MFA ${mfaEnabled ? '已启用' : '未启用'}`"
        :badge-intent="mfaEnabled ? 'success' : 'warning'"
        :description="auth.user?.username || '-'"
        icon="user-round-check"
        title="当前账号"
      >
        <AoiKeyValueList :items="securityItems" layout="cards" />
        <div class="security-actions">
          <AoiButton appearance="soft" icon="monitor-check" @click="openSessions">查看会话</AoiButton>
          <AoiButton appearance="soft" intent="danger" icon="log-out" :loading="auth.loading" @click="logout">
            退出登录
          </AoiButton>
        </div>
      </AoiAdminCard>

      <AoiAdminCard
        class="security-mfa-card"
        :description="mfaEnabled ? '登录时需要一次性验证码' : '建议启用一次性验证码保护账号'"
        icon="shield-check"
        title="MFA"
      >
        <div class="form-grid">
          <AoiButton appearance="soft" icon="shield-plus" :loading="loading" @click="setupMFA">
            {{ mfaEnabled ? "轮换密钥" : "生成密钥" }}
          </AoiButton>
          <div v-if="secret && otpauthUrl" class="mfa-setup">
            <AoiTabs v-model="mfaSetupTab" :items="mfaSetupTabs" aria-label="MFA 录入方式" />
            <div class="mfa-setup__panel">
              <div v-if="mfaSetupTab === 'qr'" class="mfa-qr-panel">
                <div class="mfa-qr-panel__frame">
                  <img v-if="mfaQrDataUrl" :src="mfaQrDataUrl" alt="MFA otpauth 二维码" />
                  <AoiProgress v-else type="circular" indeterminate />
                </div>
                <p>使用验证器扫描二维码，随后输入 6 位验证码启用 MFA。</p>
              </div>
              <div v-else-if="mfaSetupTab === 'otpauth'" class="mfa-secret-panel">
                <AoiTextField :model-value="otpauthUrl" label="otpauth URL" icon="link" disabled multiline :rows="4" />
                <AoiButton appearance="soft" icon="copy" @click="copyMFAValue(otpauthUrl, 'otpauth URL')">复制 URL</AoiButton>
              </div>
              <div v-else class="mfa-secret-panel">
                <AoiTextField :model-value="secret" label="Secret" icon="key-round" disabled />
                <AoiButton appearance="soft" icon="copy" @click="copyMFAValue(secret, 'Secret')">复制 Secret</AoiButton>
              </div>
            </div>
            <AoiTextField v-model="mfaCode" label="验证码" icon="shield-check" placeholder="123456" @enter="verifyMFA" />
            <AoiButton icon="check" :loading="verifying" :disabled="!mfaCode.trim()" @click="verifyMFA">验证并启用</AoiButton>
          </div>
        </div>
      </AoiAdminCard>
    </section>
  </div>
</template>

<style scoped>
.security-workspace {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(var(--aoi-admin-security-side-min-width), var(--aoi-admin-security-side-width));
  gap: var(--aoi-admin-panel-gap);
  align-items: start;
}

.security-account-card :deep(.aoi-admin-card__body) {
  display: grid;
  gap: var(--aoi-admin-panel-gap);
}

.security-mfa-card :deep(.aoi-admin-card__body),
.mfa-setup,
.mfa-setup__panel,
.mfa-secret-panel {
  display: grid;
  gap: var(--aoi-admin-card-gap);
}

.mfa-qr-panel {
  display: grid;
  gap: var(--aoi-admin-card-gap);
  justify-items: center;
  text-align: center;
}

.mfa-qr-panel__frame {
  display: grid;
  width: min(100%, 260px);
  aspect-ratio: 1;
  align-items: center;
  justify-items: center;
  border: 1px solid var(--aoi-border);
  border-radius: var(--aoi-radius-card);
  background: #fff;
  padding: 10px;
}

.mfa-qr-panel__frame img {
  display: block;
  width: 100%;
  max-width: 240px;
  height: auto;
}

.mfa-qr-panel p {
  max-width: 30rem;
  margin: 0;
  color: var(--aoi-text-muted);
  font-size: 13px;
  line-height: 1.6;
}

.security-actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--aoi-admin-card-gap);
  justify-content: flex-end;
}

@media (max-width: 1100px) {
  .security-workspace {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 720px) {
  .security-actions {
    align-items: stretch;
    flex-direction: column;
  }
}
</style>
