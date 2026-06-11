<script setup lang="ts">
import type { PluginManifest } from "~/types/api"
import type { PluginProxyOptions } from "~/composables/useAdminApi"

type RemoteModule = {
  mount?: (container: HTMLElement, context: PluginContext) => void | (() => void) | Promise<void | (() => void)>
  unmount?: () => void | Promise<void>
}

type PluginContext = {
  plugin: PluginManifest
  organization: ReturnType<typeof useAuthStore>["currentOrg"]
  request: <T = unknown>(path: string, options?: PluginProxyOptions) => Promise<T>
  route: {
    params: Record<string, unknown>
    path: string
    query: Record<string, unknown>
  }
  theme: Record<string, string>
  user: ReturnType<typeof useAuthStore>["user"]
}

const route = useRoute()
const api = useAdminApi()
const auth = useAuthStore()
const container = ref<HTMLElement | null>(null)
const plugin = ref<PluginManifest | null>(null)
const loading = ref(false)
const error = ref("")

let cleanup: (() => void | Promise<void>) | null = null

async function loadPlugin() {
  const pluginId = String(route.params.pluginId || "")
  if (!pluginId) {
    error.value = "插件不存在。"
    return
  }

  loading.value = true
  error.value = ""
  await unmountPlugin()

  try {
    const manifest = await api.getPlugin(pluginId)
    plugin.value = manifest
    await nextTick()
    if (!container.value) {
      throw new Error("插件容器未就绪")
    }
    const entry = resolvePluginEntry(manifest)
    if (!entry) {
      throw new Error("插件未配置前端入口")
    }
    const remote = await import(/* @vite-ignore */ entry) as RemoteModule
    if (typeof remote.mount !== "function") {
      throw new Error("插件模块缺少 mount(container, context) 导出")
    }
    const unmount = await remote.mount(container.value, createPluginContext(manifest))
    cleanup = typeof unmount === "function" ? unmount : remote.unmount || null
  } catch (err) {
    error.value = errorMessage(err)
  } finally {
    loading.value = false
  }
}

function createPluginContext(manifest: PluginManifest): PluginContext {
  return {
    plugin: manifest,
    organization: auth.currentOrg,
    request: (path, options) => api.proxyPlugin(manifest.id, path, options),
    route: {
      params: route.params,
      path: route.path,
      query: route.query
    },
    theme: readThemeTokens(),
    user: auth.user
  }
}

function resolvePluginEntry(manifest: PluginManifest) {
  const entry = manifest.frontend?.entry || ""
  if (!entry) {
    return ""
  }
  if (/^https?:\/\//i.test(entry)) {
    return entry
  }
  return `${manifest.baseURL.replace(/\/$/, "")}/${entry.replace(/^\//, "")}`
}

function readThemeTokens() {
  if (!import.meta.client) {
    return {}
  }
  const styles = getComputedStyle(document.documentElement)
  const tokens = [
    "--aoi-accent-60",
    "--aoi-surface",
    "--aoi-text",
    "--aoi-text-muted",
    "--aoi-border",
    "--aoi-radius-control"
  ]
  return Object.fromEntries(tokens.map((token) => [token, styles.getPropertyValue(token).trim()]))
}

async function unmountPlugin() {
  if (!cleanup) {
    return
  }
  await cleanup()
  cleanup = null
}

onMounted(loadPlugin)
watch(() => route.fullPath, loadPlugin)
onBeforeUnmount(unmountPlugin)
</script>

<template>
  <div class="page-grid">
    <PageHeader :title="plugin?.name || '插件'" icon="blocks" :description="plugin ? `${plugin.id} · ${plugin.version}` : '加载插件中'">
      <template #actions>
        <AoiButton appearance="soft" icon="refresh-cw" :loading="loading" @click="loadPlugin">刷新</AoiButton>
      </template>
    </PageHeader>

    <AoiStatusMessage tone="danger" :message="error" />

    <section ref="container" class="plugin-host" />
  </div>
</template>

<style scoped>
.plugin-host {
  min-height: calc(100vh - var(--aoi-topbar-height) - 150px);
  border: 1px solid var(--aoi-border);
  border-radius: var(--aoi-radius-card);
  background: rgba(255, 255, 255, .72);
  overflow: hidden;
}
</style>
