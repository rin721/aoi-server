<script setup lang="ts">
definePageMeta({
  layout: "auth",
  public: true
})

const api = useAdminApi()
const route = useRoute()
const token = ref("")
const newPassword = ref("")
const loading = ref(false)
const error = ref("")
const success = ref("")

async function submit() {
  if (!token.value.trim() || !newPassword.value) {
    return
  }

  loading.value = true
  error.value = ""
  success.value = ""
  try {
    await api.resetPassword({
      newPassword: newPassword.value,
      token: token.value.trim()
    })
    success.value = "密码已重置，可以返回登录。"
  } catch (err) {
    error.value = errorMessage(err)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  token.value = String(route.query.token || token.value || "")
})

useHead({
  title: "重置密码 - Aoi Admin"
})
</script>

<template>
  <form class="form-grid" @submit.prevent="submit">
    <div>
      <h2>重置密码</h2>
      <p class="muted">输入找回密码接口返回的 token 和新密码。</p>
    </div>
    <AoiStatusMessage tone="danger" :message="error" />
    <AoiStatusMessage tone="success" :message="success" />
    <AoiTextField v-model="token" label="重置令牌" icon="ticket" @enter="submit" />
    <AoiTextField v-model="newPassword" label="新密码" type="password" icon="key-round" @enter="submit" />
    <div class="action-row">
      <AoiButton type="submit" icon="rotate-cw" :loading="loading">重置密码</AoiButton>
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




