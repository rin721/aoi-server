<script setup lang="ts">
definePageMeta({
  layout: "auth",
  public: true
})

const api = useAdminApi()
const auth = useAuthStore()
const orgCode = ref("acme")
const orgName = ref("Acme Corp")
const username = ref("admin")
const displayName = ref("Admin")
const email = ref("admin@example.com")
const password = ref("")
const checking = ref(true)
const completed = ref(false)
const error = ref("")
const passwordMinLength = 8
const passwordError = computed(() =>
  password.value && password.value.length < passwordMinLength
    ? `密码至少需要 ${passwordMinLength} 位。`
    : undefined
)

const canSubmit = computed(() => Boolean(
  orgCode.value.trim()
  && orgName.value.trim()
  && username.value.trim()
  && email.value.trim()
  && password.value
  && !passwordError.value
))

async function checkSetupStatus() {
  checking.value = true
  error.value = ""
  try {
    const status = await api.getSetupStatus()
    completed.value = !status.required
    if (!status.required) {
      await navigateTo(auth.authenticated ? "/" : "/login")
    }
  } catch (err) {
    error.value = errorMessage(err)
  } finally {
    checking.value = false
  }
}

async function submit() {
  if (!canSubmit.value || completed.value) {
    if (passwordError.value) {
      error.value = passwordError.value
    }
    return
  }

  error.value = ""
  try {
    await auth.initialAdminSetup({
      displayName: displayName.value.trim() || undefined,
      email: email.value.trim(),
      orgCode: orgCode.value.trim(),
      orgName: orgName.value.trim(),
      password: password.value,
      username: username.value.trim()
    })
    await navigateTo("/")
  } catch (err) {
    error.value = errorMessage(err)
  }
}

onMounted(checkSetupStatus)

useHead({
  title: "首次初始化 - Aoi Admin"
})
</script>

<template>
  <form class="form-grid" @submit.prevent="submit">
    <div>
      <h2>首次初始化</h2>
      <p class="muted">创建第一个组织和 owner 管理员。</p>
    </div>

    <AoiStatusMessage tone="danger" :message="error" />
    <AoiStatusMessage v-if="completed" tone="success" message="系统已经完成初始化。" />

    <AoiTextField v-model="orgCode" label="组织 Code" icon="badge" placeholder="acme" :disabled="checking || completed" @enter="submit" />
    <AoiTextField v-model="orgName" label="组织名称" icon="building-2" placeholder="Acme Corp" :disabled="checking || completed" @enter="submit" />
    <AoiTextField v-model="username" label="用户名" icon="user" autocomplete="username" :disabled="checking || completed" @enter="submit" />
    <AoiTextField v-model="displayName" label="显示名称" icon="id-card" :disabled="checking || completed" @enter="submit" />
    <AoiTextField v-model="email" label="邮箱" type="email" icon="mail" autocomplete="email" placeholder="admin@example.com" :disabled="checking || completed" @enter="submit" />
    <AoiTextField
      v-model="password"
      label="密码"
      icon="key-round"
      type="password"
      autocomplete="new-password"
      :disabled="checking || completed"
      :supporting-text="`至少 ${passwordMinLength} 位`"
      :error-text="passwordError"
      @enter="submit"
    />

    <AoiButton type="submit" icon="rocket" :disabled="!canSubmit || checking || completed" :loading="auth.loading || checking">
      初始化并进入
    </AoiButton>

    <div class="auth-links">
      <AoiLink to="/login">返回登录</AoiLink>
    </div>
  </form>
</template>

<style scoped>
h2 {
  margin: 0;
  font-size: 24px;
}

.auth-links {
  color: var(--aoi-accent-60);
  font-weight: 760;
}
</style>




