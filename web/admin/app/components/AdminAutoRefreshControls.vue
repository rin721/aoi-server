<script setup lang="ts">
import { ADMIN_AUTO_REFRESH_CONFIG } from "~/config/admin-auto-refresh"

const props = withDefaults(defineProps<{
  modelValue?: boolean
  disabled?: boolean
  label?: string
  lastRefreshedLabel: string
  nextRefreshLabel: string
  statusLabel: string
}>(), {
  disabled: false,
  label: ADMIN_AUTO_REFRESH_CONFIG.labels.toggle,
  modelValue: true
})

const emit = defineEmits<{
  "update:modelValue": [value: boolean]
}>()
</script>

<template>
  <div class="admin-auto-refresh">
    <AoiSwitch
      :model-value="props.modelValue"
      :disabled="props.disabled"
      :label="props.label"
      @update:model-value="emit('update:modelValue', $event)"
    />
    <AoiMetaPill appearance="outline" :intent="props.modelValue ? 'success' : 'neutral'" icon="refresh-cw">
      {{ props.statusLabel }}
    </AoiMetaPill>
    <AoiMetaPill appearance="plain" intent="neutral" icon="clock-3">
      {{ props.nextRefreshLabel }}
    </AoiMetaPill>
    <AoiMetaPill appearance="plain" intent="neutral" icon="history">
      {{ props.lastRefreshedLabel }}
    </AoiMetaPill>
  </div>
</template>

<style scoped>
.admin-auto-refresh {
  align-items: center;
  display: flex;
  flex-wrap: wrap;
  gap: var(--aoi-admin-auto-refresh-gap);
  min-width: 0;
}

.admin-auto-refresh :deep(.aoi-field-control) {
  min-height: var(--aoi-control-height-sm);
  white-space: nowrap;
}
</style>
