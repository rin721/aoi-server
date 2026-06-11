<script setup lang="ts">
import type { Organization } from "~/types/api"

const api = useAdminApi()
const auth = useAuthStore()
const organizations = ref<Organization[]>([])
const code = ref("")
const name = ref("")
const loading = ref(false)
const saving = ref(false)
const error = ref("")
const success = ref("")

async function load() {
  loading.value = true
  error.value = ""
  try {
    organizations.value = await api.listOrganizations()
  } catch (err) {
    error.value = errorMessage(err)
  } finally {
    loading.value = false
  }
}

async function createOrg() {
  if (!code.value.trim() || !name.value.trim()) {
    return
  }

  saving.value = true
  error.value = ""
  success.value = ""
  try {
    const org = await api.createOrganization({
      code: code.value.trim(),
      name: name.value.trim()
    })
    success.value = `组织 ${org.name} 已创建。`
    code.value = ""
    name.value = ""
    await Promise.all([load(), auth.fetchSession()])
  } catch (err) {
    error.value = errorMessage(err)
  } finally {
    saving.value = false
  }
}

onMounted(load)

useHead({
  title: "组织 - Aoi Admin"
})
</script>

<template>
  <div class="page-grid">
    <PageHeader title="组织" icon="building-2" description="查看可管理组织，创建新组织，并切换当前访问上下文。">
      <template #actions>
        <AoiButton appearance="soft" icon="refresh-cw" :loading="loading" @click="load">刷新</AoiButton>
      </template>
    </PageHeader>

    <AoiStatusMessage tone="danger" :message="error" />
    <AoiStatusMessage tone="success" :message="success" />

    <section class="two-column-grid">
      <article class="admin-card">
        <div class="admin-card__header">
          <h2>组织列表</h2>
          <span class="badge">{{ organizations.length }} 个</span>
        </div>
        <div class="data-table-wrap">
          <table class="data-table">
            <thead>
              <tr>
                <th>ID</th>
                <th>Code</th>
                <th>名称</th>
                <th>状态</th>
                <th>创建时间</th>
                <th>操作</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="org in organizations" :key="org.id">
                <td>{{ org.id }}</td>
                <td class="mono">{{ org.code }}</td>
                <td>{{ org.name }}</td>
                <td><span class="badge" :class="org.status === 'active' ? 'badge--success' : 'badge--warning'">{{ formatStatus(org.status) }}</span></td>
                <td>{{ formatDateTime(org.createdAt) }}</td>
                <td>
                  <AoiButton appearance="soft" icon="repeat" :disabled="org.id === auth.currentOrgId" @click="auth.switchOrg(org.id)">
                    切换
                  </AoiButton>
                </td>
              </tr>
              <tr v-if="!organizations.length">
                <td colspan="6" class="muted">暂无组织。</td>
              </tr>
            </tbody>
          </table>
        </div>
      </article>

      <article class="admin-card">
        <div class="admin-card__header">
          <h2>创建组织</h2>
        </div>
        <form class="admin-card__body form-grid" @submit.prevent="createOrg">
          <AoiTextField v-model="code" label="组织 Code" icon="badge" placeholder="acme" />
          <AoiTextField v-model="name" label="组织名称" icon="building" placeholder="Acme Corp" />
          <AoiButton type="submit" icon="plus" :loading="saving" :disabled="!code || !name">创建组织</AoiButton>
        </form>
      </article>
    </section>
  </div>
</template>
