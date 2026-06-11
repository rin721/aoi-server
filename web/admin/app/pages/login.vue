<script setup lang="ts">
definePageMeta({
  layout: "auth",
  public: true
})

const route = useRoute()
const auth = useAuthStore()
const identifier = ref("")
const password = ref("")
const orgCode = ref("")
const mfaCode = ref("")
const error = ref("")
const needsMFA = ref(false)

const canSubmit = computed(() => identifier.value.trim() && password.value)

async function submit() {
  if (!canSubmit.value) {
    return
  }

  error.value = ""
  try {
    await auth.login({
      identifier: identifier.value.trim(),
      mfaCode: mfaCode.value.trim() || undefined,
      orgCode: orgCode.value.trim() || undefined,
      password: password.value
    })
    await navigateTo(String(route.query.redirect || "/"))
  } catch (err) {
    const message = errorMessage(err)
    needsMFA.value = message.toLowerCase().includes("mfa")
    error.value = needsMFA.value ? "当前账号需要 MFA 验证，请输入动态验证码后重试。" : message
  }
}

useHead({
  title: "登录 - Aoi Admin"
})
</script>

<template>
  <form class="form-grid" @submit.prevent="submit">
    <div>
      <h2>登录管理台</h2>
      <p class="muted">使用 IAM 初始化账号进入组织管理视图。</p>
    </div>

    <AoiStatusMessage tone="danger" :message="error" />
    <AoiTextField v-model="identifier" label="用户名或邮箱" icon="user" autocomplete="username" placeholder="admin@example.com" @enter="submit" />
    <AoiTextField v-model="password" label="密码" icon="key-round" type="password" autocomplete="current-password" @enter="submit" />
    <AoiTextField v-model="orgCode" label="组织 Code" icon="building-2" placeholder="acme" @enter="submit" />
    <AoiTextField v-if="needsMFA || mfaCode" v-model="mfaCode" label="MFA 验证码" icon="shield-check" placeholder="123456" @enter="submit" />

    <AoiButton type="submit" icon="log-in" :disabled="!canSubmit" :loading="auth.loading">
      登录
    </AoiButton>

    <div class="auth-links">
      <NuxtLink to="/password/forgot">忘记密码</NuxtLink>
      <NuxtLink to="/password/reset">已有重置令牌</NuxtLink>
    </div>
  </form>
</template>

<style scoped>
h2 {
  margin: 0;
  font-size: 24px;
}

.auth-links {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  color: var(--aoi-accent-60);
  font-weight: 760;
}
</style>
