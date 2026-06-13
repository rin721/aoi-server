<script setup lang="ts">
const props = withDefaults(defineProps<{
  autoRefresh?: {
    disabled?: boolean
    enabled: boolean
    lastRefreshedLabel: string
    nextRefreshLabel: string
    statusLabel: string
  }
  description?: string
  refreshDisabled?: boolean
  refreshLabel?: string
  refreshLoading?: boolean
  title: string
}>(), {
  autoRefresh: undefined,
  description: undefined,
  refreshDisabled: false,
  refreshLabel: "刷新",
  refreshLoading: false
})

const emit = defineEmits<{
  "refresh": []
  "update:autoRefreshEnabled": [value: boolean]
}>()
</script>

<template>
  <header v-aoi-reveal="'rise'" class="settings-page-header">
    <div class="settings-page-header__copy">
      <h1>{{ props.title }}</h1>
      <p v-if="props.description">{{ props.description }}</p>
    </div>
    <div v-if="props.autoRefresh || $slots.actions" class="settings-page-header__actions">
      <AdminAutoRefreshControls
        v-if="props.autoRefresh"
        :model-value="props.autoRefresh.enabled"
        :disabled="props.autoRefresh.disabled"
        :last-refreshed-label="props.autoRefresh.lastRefreshedLabel"
        :next-refresh-label="props.autoRefresh.nextRefreshLabel"
        :status-label="props.autoRefresh.statusLabel"
        @update:model-value="emit('update:autoRefreshEnabled', $event)"
      />
      <AoiButton
        v-if="props.autoRefresh"
        appearance="soft"
        icon="refresh-cw"
        :disabled="props.refreshDisabled"
        :loading="props.refreshLoading"
        @click="emit('refresh')"
      >
        {{ props.refreshLabel }}
      </AoiButton>
      <slot name="actions" />
    </div>
  </header>
</template>

<style scoped>
.settings-page-header {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: 7px;
  align-items: start;
}

.settings-page-header__copy {
  display: grid;
  gap: 7px;
}

.settings-page-header__copy h1,
.settings-page-header__copy p {
  margin: 0;
}

.settings-page-header__copy h1 {
  color: var(--aoi-text);
  font-size: 24px;
  line-height: 1.2;
}

.settings-page-header__copy p {
  color: var(--aoi-text-muted);
  line-height: 1.7;
}

.settings-page-header__actions {
  display: inline-flex;
  flex-wrap: wrap;
  gap: var(--aoi-grid-gap-compact);
  justify-content: end;
  min-width: 0;
}

@media (max-width: 639px) {
  .settings-page-header {
    grid-template-columns: 1fr;
  }

  .settings-page-header__actions {
    justify-content: start;
  }
}
</style>
