<script setup lang="ts">
definePageMeta({
  layout: "auth",
  public: true
})

const route = useRoute()
const api = useAdminApi()
const username = ref("")
const displayName = ref("")
const password = ref("")
const loading = ref(false)
const error = ref("")
const success = ref("")

async function submit() {
  if (!username.value.trim() || !password.value) {
    return
  }

  loading.value = true
  error.value = ""
  success.value = ""
  try {
    await api.acceptInvitation(String(route.params.token), {
      displayName: displayName.value.trim() || undefined,
      password: password.value,
      username: username.value.trim()
    })
    success.value = "邀请已接受，请使用新账号登录。"
  } catch (err) {
    error.value = errorMessage(err)
  } finally {
    loading.value = false
  }
}

useHead({
  title: "接受邀请 - Aoi Admin"
})
</script>

<template>
  <form class="form-grid" @submit.prevent="submit">
    <div>
      <h2>接受组织邀请</h2>
      <p class="muted">创建账号并加入邀请指定的组织。</p>
    </div>
    <AoiStatusMessage tone="danger" :message="error" />
    <AoiStatusMessage tone="success" :message="success" />
    <AoiTextField v-model="username" label="用户名" icon="user" @enter="submit" />
    <AoiTextField v-model="displayName" label="显示名称" icon="id-card" @enter="submit" />
    <AoiTextField v-model="password" label="密码" type="password" icon="key-round" @enter="submit" />
    <div class="action-row">
      <AoiButton type="submit" icon="user-check" :loading="loading">接受邀请</AoiButton>
      <AoiButton appearance="soft" intent="neutral" to="/login">返回登录</AoiButton>
    </div>
  </form>
</template>

<style scoped>
h2 {
  margin: 0;
  font-size: 24px;
}
</style>
