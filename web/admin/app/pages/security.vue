<script setup lang="ts">
const api = useAdminApi()
const auth = useAuthStore()

const secret = ref("")
const otpauthUrl = ref("")
const mfaCode = ref("")
const loading = ref(false)
const verifying = ref(false)
const error = ref("")
const success = ref("")

const accountName = computed(() => auth.user?.displayName || auth.user?.username || "-")
const mfaEnabled = computed(() => Boolean(auth.user?.mfaEnabled))
const accessExpiresLabel = computed(() => formatDateTime(auth.accessExpiresAt))
const refreshExpiresLabel = computed(() => formatDateTime(auth.refreshExpiresAt))
const securityItems = computed(() => [
  { icon: "user", label: "账号", value: accountName.value },
  { icon: "mail", label: "邮箱", value: auth.user?.email || "-" },
  { icon: "building-2", label: "当前组织", value: auth.currentOrg?.name || "-" },
  { icon: "fingerprint", label: "当前会话", value: auth.sessionId || "-" },
  { icon: "clock", label: "Access 过期", value: accessExpiresLabel.value },
  { icon: "calendar-clock", label: "Refresh 过期", value: refreshExpiresLabel.value }
])

async function setupMFA() {
  loading.value = true
  error.value = ""
  success.value = ""
  try {
    const result = await api.setupMFA()
    secret.value = result.secret
    otpauthUrl.value = result.otpauthUrl
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
    mfaCode.value = ""
    await auth.fetchSession()
  } catch (err) {
    error.value = errorMessage(err)
  } finally {
    verifying.value = false
  }
}

async function logout() {
  await auth.logout()
  await navigateTo("/login")
}

async function openSessions() {
  await navigateTo("/admin/sessions")
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
      <article class="admin-card security-account-card">
        <div class="admin-card__header">
          <div>
            <h2>当前账号</h2>
            <span>{{ auth.user?.username || "-" }}</span>
          </div>
          <span class="badge" :class="mfaEnabled ? 'badge--success' : 'badge--warning'">
            MFA {{ mfaEnabled ? "已启用" : "未启用" }}
          </span>
        </div>

        <div class="security-summary-grid">
          <div v-for="item in securityItems" :key="item.label" class="security-summary-item">
            <AoiIcon :name="item.icon" decorative />
            <span>{{ item.label }}</span>
            <strong>{{ item.value }}</strong>
          </div>
        </div>

        <div class="security-actions">
          <AoiButton appearance="soft" icon="monitor-check" @click="openSessions">查看会话</AoiButton>
          <AoiButton appearance="soft" intent="danger" icon="log-out" :loading="auth.loading" @click="logout">
            退出登录
          </AoiButton>
        </div>
      </article>

      <article class="admin-card security-mfa-card">
        <div class="admin-card__header">
          <div>
            <h2>MFA</h2>
            <span>{{ mfaEnabled ? "登录时需要一次性验证码" : "建议启用一次性验证码保护账号" }}</span>
          </div>
        </div>
        <div class="admin-card__body form-grid">
          <AoiButton appearance="soft" icon="shield-plus" :loading="loading" @click="setupMFA">
            {{ mfaEnabled ? "轮换密钥" : "生成密钥" }}
          </AoiButton>
          <div v-if="secret" class="form-grid">
            <AoiTextField :model-value="secret" label="Secret" icon="key-round" disabled />
            <AoiTextField :model-value="otpauthUrl" label="otpauth URL" icon="link" disabled />
            <AoiTextField v-model="mfaCode" label="验证码" icon="shield-check" placeholder="123456" @enter="verifyMFA" />
            <AoiButton icon="check" :loading="verifying" :disabled="!mfaCode.trim()" @click="verifyMFA">验证并启用</AoiButton>
          </div>
        </div>
      </article>
    </section>
  </div>
</template>

<style scoped>
.security-workspace {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(320px, 420px);
  gap: 16px;
  align-items: start;
}

.security-account-card .admin-card__header > div,
.security-mfa-card .admin-card__header > div {
  display: grid;
  gap: 8px;
}

.security-account-card .admin-card__header span,
.security-mfa-card .admin-card__header span {
  color: var(--aoi-text-muted);
  font-size: 13px;
}

.security-summary-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
  padding: 16px;
}

.security-summary-item {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr);
  gap: 6px 10px;
  align-items: center;
  min-width: 0;
  padding: 12px;
  border: 1px solid var(--aoi-border);
  border-radius: var(--aoi-radius-md);
  background: var(--aoi-surface-soft);
}

.security-summary-item :deep(.aoi-icon) {
  color: var(--aoi-primary);
}

.security-summary-item span {
  color: var(--aoi-text-muted);
  font-size: 12px;
}

.security-summary-item strong {
  grid-column: 1 / -1;
  min-width: 0;
  overflow-wrap: anywhere;
  color: var(--aoi-text);
  font-size: 14px;
}

.security-actions {
  display: flex;
  gap: 10px;
  justify-content: flex-end;
  padding: 14px 16px;
  border-top: 1px solid var(--aoi-border);
}

@media (max-width: 1100px) {
  .security-workspace {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 720px) {
  .security-summary-grid {
    grid-template-columns: 1fr;
  }

  .security-actions {
    align-items: stretch;
    flex-direction: column;
  }
}
</style>
