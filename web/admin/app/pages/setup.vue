<script setup lang="ts">
import type { PasswordPolicy } from "~/types/admin"

definePageMeta({
  layout: "auth",
  public: true
})

const api = useAdminApi()
const auth = useAuthStore()
const navigation = useAdminNavigation()
const orgCode = ref("acme")
const orgName = ref("Acme Corp")
const username = ref("admin")
const displayName = ref("Admin")
const email = ref("admin@example.com")
const password = ref("")
const checking = ref(true)
const completed = ref(false)
const submitting = ref(false)
const error = ref("")
const fallbackPasswordPolicy: PasswordPolicy = {
  minLength: 8,
  requireLower: false,
  requireNumber: false,
  requireSymbol: false,
  requireUpper: false
}
const passwordPolicy = ref<PasswordPolicy>({ ...fallbackPasswordPolicy })
const normalizedPasswordPolicy = computed(() => normalizePasswordPolicy(passwordPolicy.value))
const passwordRequirementItems = computed(() => passwordRequirements(normalizedPasswordPolicy.value))
const passwordRequirementText = computed(() => passwordRequirementItems.value.join("，"))
const passwordError = computed(() => {
  if (!password.value) {
    return undefined
  }
  const missing = missingPasswordRequirements(password.value, normalizedPasswordPolicy.value)
  return missing.length ? `密码需要${missing.join("、")}。` : undefined
})

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
    passwordPolicy.value = normalizePasswordPolicy(status.passwordPolicy)
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
  if (!canSubmit.value || completed.value || submitting.value) {
    if (passwordError.value) {
      error.value = passwordError.value
    }
    return
  }

  error.value = ""
  submitting.value = true
  try {
    await auth.initialAdminSetup({
      displayName: displayName.value.trim() || undefined,
      email: email.value.trim(),
      orgCode: orgCode.value.trim(),
      orgName: orgName.value.trim(),
      password: password.value,
      username: username.value.trim()
    })
    await navigation.loadNavigation()
    await navigateTo("/")
  } catch (err) {
    error.value = errorMessage(err)
  } finally {
    submitting.value = false
  }
}

onMounted(checkSetupStatus)

useHead({
  title: "首次初始化 - Aoi Admin"
})

function normalizePasswordPolicy(policy?: Partial<PasswordPolicy>): PasswordPolicy {
  return {
    minLength: Math.max(1, Number(policy?.minLength || fallbackPasswordPolicy.minLength)),
    requireLower: Boolean(policy?.requireLower),
    requireNumber: Boolean(policy?.requireNumber),
    requireSymbol: Boolean(policy?.requireSymbol),
    requireUpper: Boolean(policy?.requireUpper)
  }
}

function passwordRequirements(policy: PasswordPolicy) {
  const items = [`至少 ${policy.minLength} 位`]
  if (policy.requireLower) {
    items.push("包含小写字母")
  }
  if (policy.requireUpper) {
    items.push("包含大写字母")
  }
  if (policy.requireNumber) {
    items.push("包含数字")
  }
  if (policy.requireSymbol) {
    items.push("包含符号")
  }
  return items
}

function missingPasswordRequirements(value: string, policy: PasswordPolicy) {
  const missing: string[] = []
  if (Array.from(value).length < policy.minLength) {
    missing.push(`至少 ${policy.minLength} 位`)
  }
  if (policy.requireLower && !/[a-z]/.test(value)) {
    missing.push("包含小写字母")
  }
  if (policy.requireUpper && !/[A-Z]/.test(value)) {
    missing.push("包含大写字母")
  }
  if (policy.requireNumber && !/\d/.test(value)) {
    missing.push("包含数字")
  }
  if (policy.requireSymbol && !/[^A-Za-z0-9\s]/.test(value)) {
    missing.push("包含符号")
  }
  return missing
}
</script>

<template>
  <form class="form-grid" @submit.prevent="submit">
    <div>
      <h2>首次初始化</h2>
      <p class="muted">创建第一个组织和 owner 管理员。</p>
    </div>

    <AoiStatusMessage tone="danger" :message="error" />
    <AoiStatusMessage v-if="completed" tone="success" message="系统已经完成初始化。" />

    <AoiTextField v-model="orgCode" label="组织 Code" icon="badge" placeholder="acme" :disabled="checking || completed || submitting" @enter="submit" />
    <AoiTextField v-model="orgName" label="组织名称" icon="building-2" placeholder="Acme Corp" :disabled="checking || completed || submitting" @enter="submit" />
    <AoiTextField v-model="username" label="用户名" icon="user" autocomplete="username" :disabled="checking || completed || submitting" @enter="submit" />
    <AoiTextField v-model="displayName" label="显示名称" icon="id-card" :disabled="checking || completed || submitting" @enter="submit" />
    <AoiTextField v-model="email" label="邮箱" type="email" icon="mail" autocomplete="email" placeholder="admin@example.com" :disabled="checking || completed || submitting" @enter="submit" />
    <AoiTextField
      v-model="password"
      label="密码"
      icon="key-round"
      type="password"
      autocomplete="new-password"
      :disabled="checking || completed || submitting"
      :supporting-text="passwordRequirementText"
      :error-text="passwordError"
      @enter="submit"
    />

    <AoiButton type="submit" icon="rocket" :disabled="!canSubmit || checking || completed || submitting" :loading="auth.loading || checking || submitting">
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
