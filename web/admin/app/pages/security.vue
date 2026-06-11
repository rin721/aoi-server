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

useHead({
  title: "安全 - Aoi Admin"
})
</script>

<template>
  <div class="page-grid">
    <PageHeader title="安全" icon="lock-keyhole" description="管理当前账号的 MFA 状态，并可退出当前会话。" />

    <AoiStatusMessage tone="danger" :message="error" />
    <AoiStatusMessage tone="success" :message="success" />

    <section class="two-column-grid">
      <article class="admin-card">
        <div class="admin-card__header">
          <h2>当前账号</h2>
        </div>
        <div class="admin-card__body page-grid">
          <div>
            <strong>{{ auth.user?.displayName || auth.user?.username }}</strong>
            <p class="muted">{{ auth.user?.email }}</p>
          </div>
          <div class="action-row">
            <span class="badge" :class="auth.user?.mfaEnabled ? 'badge--success' : 'badge--warning'">
              MFA {{ auth.user?.mfaEnabled ? "已启用" : "未启用" }}
            </span>
            <span class="badge">{{ formatStatus(auth.user?.status) }}</span>
          </div>
          <AoiButton appearance="soft" intent="danger" icon="log-out" :loading="auth.loading" @click="logout">
            退出登录
          </AoiButton>
        </div>
      </article>

      <article class="admin-card">
        <div class="admin-card__header">
          <h2>MFA</h2>
        </div>
        <div class="admin-card__body form-grid">
          <AoiButton appearance="soft" icon="shield-plus" :loading="loading" @click="setupMFA">生成或轮换密钥</AoiButton>
          <div v-if="secret" class="form-grid">
            <AoiTextField :model-value="secret" label="Secret" icon="key-round" />
            <AoiTextField :model-value="otpauthUrl" label="otpauth URL" icon="link" />
            <AoiTextField v-model="mfaCode" label="验证码" icon="shield-check" placeholder="123456" @enter="verifyMFA" />
            <AoiButton icon="check" :loading="verifying" @click="verifyMFA">验证并启用</AoiButton>
          </div>
        </div>
      </article>
    </section>
  </div>
</template>
