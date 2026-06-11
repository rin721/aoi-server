<script setup lang="ts">
definePageMeta({
  layout: "auth",
  public: true
})

const api = useAdminApi()
const email = ref("")
const loading = ref(false)
const error = ref("")
const token = ref("")
const resetUrl = ref("")
const success = ref("")

async function submit() {
  if (!email.value.trim()) {
    return
  }

  loading.value = true
  error.value = ""
  token.value = ""
  resetUrl.value = ""
  success.value = ""
  try {
    const result = await api.forgotPassword(email.value.trim())
    token.value = result.token || ""
    resetUrl.value = result.url || ""
    success.value = token.value || resetUrl.value ? "重置令牌已创建。" : "如果邮箱存在，系统会发送重置链接。"
  } catch (err) {
    error.value = errorMessage(err)
  } finally {
    loading.value = false
  }
}

useHead({
  title: "找回密码 - Aoi Admin"
})
</script>

<template>
  <form class="form-grid" @submit.prevent="submit">
    <div>
      <h2>找回密码</h2>
      <p class="muted">输入邮箱后获取密码重置链接。</p>
    </div>
    <AoiStatusMessage tone="danger" :message="error" />
    <AoiStatusMessage tone="success" :message="success" />
    <AoiStatusMessage v-if="token" tone="success" :message="`调试令牌：${token}`" />
    <AoiStatusMessage v-if="resetUrl" tone="success" :message="`调试链接：${resetUrl}`" />
    <AoiTextField v-model="email" label="邮箱" type="email" icon="mail" placeholder="admin@example.com" @enter="submit" />
    <div class="action-row">
      <AoiButton type="submit" icon="send" :loading="loading">创建令牌</AoiButton>
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
