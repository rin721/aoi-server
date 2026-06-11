<script setup lang="ts">
const auth = useAuthStore()

const orgOptions = computed(() => auth.orgs.map((org) => ({
  label: `${org.name} (${org.code})`,
  value: org.id
})))

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
  <header class="app-topbar">
    <div>
      <strong>{{ auth.currentOrg?.name || "未选择组织" }}</strong>
      <span class="muted">{{ auth.user?.email || "未登录" }}</span>
    </div>
    <div class="app-topbar__actions">
      <label v-if="orgOptions.length" class="app-topbar__select">
        <span>组织</span>
        <select :value="auth.currentOrgId || ''" @change="switchOrg(($event.target as HTMLSelectElement).value)">
          <option v-for="option in orgOptions" :key="option.value" :value="option.value">
            {{ option.label }}
          </option>
        </select>
      </label>
      <AoiButton appearance="soft" icon="log-out" intent="neutral" :loading="auth.loading" @click="logout">
        退出
      </AoiButton>
    </div>
  </header>
</template>

<style scoped>
.app-topbar {
  display: flex;
  min-height: var(--aoi-topbar-height);
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  border-bottom: 1px solid var(--aoi-border);
  background: rgba(255, 255, 255, .62);
  padding: 0 clamp(16px, 3vw, 34px);
  backdrop-filter: blur(16px);
}

.app-topbar strong,
.app-topbar span {
  display: block;
}

.app-topbar__actions {
  display: flex;
  align-items: center;
  gap: 10px;
}

.app-topbar__select {
  display: flex;
  align-items: center;
  gap: 8px;
  color: var(--aoi-text-muted);
  font-size: 12px;
  font-weight: 800;
}

select {
  height: 34px;
  min-width: 190px;
  border: 1px solid var(--aoi-border);
  border-radius: var(--aoi-radius-control);
  background: var(--aoi-surface);
  color: var(--aoi-text);
  padding: 0 8px;
}

@media (max-width: 720px) {
  .app-topbar {
    align-items: flex-start;
    flex-direction: column;
    padding-block: 12px;
  }

  .app-topbar__actions {
    width: 100%;
    flex-wrap: wrap;
  }
}
</style>
