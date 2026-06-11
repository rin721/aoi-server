<script setup lang="ts">
definePageMeta({
  layout: "auth",
  public: true
})

const auth = useAuthStore()
const orgCode = ref("")
const orgName = ref("")
const username = ref("")
const displayName = ref("")
const email = ref("")
const password = ref("")
const error = ref("")

const canSubmit = computed(() =>
  orgCode.value.trim()
  && orgName.value.trim()
  && username.value.trim()
  && email.value.trim()
  && password.value
)

async function submit() {
  if (!canSubmit.value) {
    return
  }

  error.value = ""
  try {
    await auth.signup({
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

useHead({
  title: "注册 - Aoi Admin"
})
</script>

<template>
  <form class="form-grid" @submit.prevent="submit">
    <div>
      <h2>创建工作区</h2>
      <p class="muted">注册账号并创建首个组织。</p>
    </div>

    <AoiStatusMessage tone="danger" :message="error" />
    <AoiTextField v-model="orgCode" label="组织 Code" icon="badge" placeholder="acme" @enter="submit" />
    <AoiTextField v-model="orgName" label="组织名称" icon="building-2" placeholder="Acme Corp" @enter="submit" />
    <AoiTextField v-model="username" label="用户名" icon="user" autocomplete="username" @enter="submit" />
    <AoiTextField v-model="displayName" label="显示名称" icon="id-card" @enter="submit" />
    <AoiTextField v-model="email" label="邮箱" type="email" icon="mail" autocomplete="email" placeholder="owner@example.com" @enter="submit" />
    <AoiTextField v-model="password" label="密码" icon="key-round" type="password" autocomplete="new-password" @enter="submit" />

    <AoiButton type="submit" icon="rocket" :disabled="!canSubmit" :loading="auth.loading">
      创建并进入
    </AoiButton>

    <div class="auth-links">
      <NuxtLink to="/login">已有账号，返回登录</NuxtLink>
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




