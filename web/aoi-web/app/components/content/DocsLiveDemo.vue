<script setup lang="ts">
import type { DefineComponent } from "vue"

const props = withDefaults(defineProps<{
  client?: boolean | string
  code?: string
  description?: string
  language?: string
  meta?: string
  title?: string
  unwrap?: boolean | string
}>(), {
  client: false,
  code: "",
  description: undefined,
  language: "mdc",
  meta: "",
  title: undefined,
  unwrap: false
})

const { t } = useI18n()

const aoiComponentModules = import.meta.glob("../aoi/Aoi*.vue", {
  eager: true,
  import: "default"
}) as Record<string, DefineComponent<any, any, any>>

const liveDemoComponents = Object.fromEntries(Object.entries(aoiComponentModules)
  .filter(([path]) => !path.includes(".client."))
  .map(([path, component]) => [path.match(/\/([^/]+)\.vue$/)?.[1] || path, component]))

const source = computed(() => props.code.trim())
const isClientOnly = computed(() => props.client === true || props.client === "" || props.client === "true")
const resolvedUnwrap = computed(() => props.unwrap === true ? "*" : props.unwrap || false)
const sourceLabel = computed(() => props.title
  ? `${props.title} - ${t("docs.liveDemo.source")}`
  : t("docs.liveDemo.source"))

function errorMessage(error: unknown) {
  const message = typeof error === "object" && error && "message" in error
    ? String((error as { message?: unknown }).message || "")
    : ""

  return message
    ? `${t("docs.liveDemo.error")} ${message}`
    : t("docs.liveDemo.error")
}
</script>

<template>
  <AoiSurface class="docs-live-demo" surface="panel" padding="none">
    <header v-if="props.title || props.description" class="docs-live-demo__header">
      <p class="docs-live-demo__eyebrow">{{ t("docs.liveDemo.preview") }}</p>
      <h3 v-if="props.title">{{ props.title }}</h3>
      <p v-if="props.description">{{ props.description }}</p>
    </header>

    <section class="docs-live-demo__preview" :aria-label="t('docs.liveDemo.preview')">
      <div v-if="!source" class="docs-live-demo__preview-body">
        <AoiStatusMessage as="div" intent="warning" icon="triangle-alert" :message="t('docs.liveDemo.empty')" />
      </div>

      <MDC v-else :value="source" tag="div">
        <template #default="{ body, data, error }">
          <div class="docs-live-demo__preview-body">
            <AoiStatusMessage
              v-if="error"
              as="div"
              intent="danger"
              icon="circle-alert"
              :message="errorMessage(error)"
            />
            <NuxtErrorBoundary v-else :key="source">
              <ClientOnly v-if="isClientOnly">
                <MDCRenderer
                  v-if="body"
                  :body="body"
                  :components="liveDemoComponents"
                  :data="data"
                  :unwrap="resolvedUnwrap"
                  class="docs-live-demo__rendered"
                  tag="div"
                />
                <template #fallback>
                  <AoiStatusMessage as="div" intent="info" icon="loader-circle" :message="t('docs.liveDemo.clientFallback')" />
                </template>
              </ClientOnly>
              <MDCRenderer
                v-else-if="body"
                :body="body"
                :components="liveDemoComponents"
                :data="data"
                :unwrap="resolvedUnwrap"
                class="docs-live-demo__rendered"
                tag="div"
              />
              <AoiStatusMessage v-else as="div" intent="warning" icon="triangle-alert" :message="t('docs.liveDemo.empty')" />

              <template #error="{ error: renderError }">
                <AoiStatusMessage
                  as="div"
                  intent="danger"
                  icon="circle-alert"
                  :message="errorMessage(renderError)"
                />
              </template>
            </NuxtErrorBoundary>
          </div>
        </template>
      </MDC>
    </section>

    <section class="docs-live-demo__source" :aria-label="t('docs.liveDemo.source')">
      <div class="docs-live-demo__source-header">
        <span>{{ t("docs.liveDemo.source") }}</span>
        <code>{{ props.language }}</code>
      </div>
      <AoiCodeBlock :code="source" :label="sourceLabel" />
    </section>
  </AoiSurface>
</template>

<style scoped>
.docs-live-demo {
  display: grid;
  overflow: hidden;
  margin: 20px 0;
}

.docs-live-demo__header,
.docs-live-demo__preview,
.docs-live-demo__source {
  display: grid;
  min-width: 0;
  gap: 10px;
  padding: 14px;
}

.docs-live-demo__header {
  border-bottom: 1px solid var(--aoi-border);
  background: var(--aoi-surface-muted);
}

.docs-live-demo__header h3,
.docs-live-demo__header p,
.docs-live-demo__eyebrow {
  margin: 0;
}

.docs-live-demo__header h3 {
  color: var(--aoi-text);
  font-size: 16px;
}

.docs-live-demo__header p {
  color: var(--aoi-text-muted);
  line-height: 1.7;
}

.docs-live-demo__eyebrow {
  color: var(--aoi-active-color) !important;
  font-size: 12px;
  font-weight: 820;
  text-transform: uppercase;
}

.docs-live-demo__preview {
  background:
    linear-gradient(90deg, color-mix(in srgb, var(--aoi-accent-10) 64%, transparent), transparent 56%),
    var(--aoi-card-bg);
}

.docs-live-demo__preview-body,
.docs-live-demo__rendered {
  display: flex;
  min-width: 0;
  flex-wrap: wrap;
  gap: 10px;
  align-items: center;
}

.docs-live-demo__rendered {
  width: 100%;
}

.docs-live-demo__source {
  border-top: 1px solid var(--aoi-border);
  background: var(--aoi-bg);
}

.docs-live-demo__source-header {
  display: flex;
  min-width: 0;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  color: var(--aoi-text-muted);
  font-size: 12px;
  font-weight: 760;
}

.docs-live-demo__source-header code {
  border-radius: var(--aoi-radius-xs);
  background: var(--aoi-accent-10);
  color: var(--aoi-active-color);
  padding: 2px 5px;
}

.docs-live-demo__source :deep(.aoi-code-block) {
  max-height: min(42vh, 420px);
}
</style>


