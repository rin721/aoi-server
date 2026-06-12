<script setup lang="ts">
const emit = defineEmits<{
  openSettings: []
}>()

const route = useRoute()
const auth = useAuthStore()
const { findItem } = useAdminNavigation()
const { darkActive, state } = useAdminUiPreferences()

const orgOptions = computed(() => auth.orgs.map((org) => ({
  label: `${org.name} (${org.code})`,
  value: String(org.id)
})))
const currentTitle = computed(() => findItem(route.path)?.label || "工作台")

async function switchOrg(value: string) {
  if (value && value !== auth.currentOrgId) {
    await auth.switchOrg(value)
  }
}

async function logout() {
  await auth.logout()
  await navigateTo("/login")
}
</script>

<template>
  <header class="admin-topbar">
    <div class="admin-topbar__left">
      <AoiIconButton
        :active="state.sidebarCollapsed"
        icon="panel-left"
        label="折叠侧栏"
        @click="state.sidebarCollapsed = !state.sidebarCollapsed"
      />
      <div class="admin-topbar__title">
        <span>后台管理</span>
        <strong>{{ currentTitle }}</strong>
      </div>
      </div>

    <div class="admin-topbar__actions">
      <div v-if="orgOptions.length" class="admin-topbar__org">
        <AoiSelect
          :model-value="auth.currentOrgId || ''"
          :options="orgOptions"
          appearance="outlined"
          label="组织"
          @update:model-value="switchOrg"
        />
      </div>
      <AoiIconButton
        :icon="darkActive ? 'moon' : 'sun'"
        label="切换明暗主题"
        @click="state.theme = darkActive ? 'light' : 'dark'"
      />
      <AoiIconButton icon="settings" label="打开设置" @click="emit('openSettings')" />
      <div class="admin-topbar__user">
        <span>{{ auth.user?.displayName || auth.user?.username || "未登录" }}</span>
        <small>{{ auth.user?.email || "anonymous" }}</small>
      </div>
      <AoiButton appearance="soft" icon="log-out" intent="neutral" :loading="auth.loading" @click="logout">
        退出
      </AoiButton>
    </div>
  </header>
</template>


