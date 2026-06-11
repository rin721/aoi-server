<script setup lang="ts">
defineProps<{
  label: string
  modelValue?: string | number | null
  options: Array<{ label: string, value: string | number }>
}>()

const emit = defineEmits<{
  "update:modelValue": [value: string]
}>()
</script>

<template>
  <label class="aoi-select">
    <span class="aoi-select__label">{{ label }}</span>
    <span class="aoi-select__control">
      <select :value="modelValue ?? ''" @change="emit('update:modelValue', ($event.target as HTMLSelectElement).value)">
        <option value="">请选择</option>
        <option v-for="option in options" :key="option.value" :value="option.value">
          {{ option.label }}
        </option>
      </select>
      <AoiIcon name="chevron-down" decorative />
    </span>
  </label>
</template>

<style scoped>
.aoi-select {
  display: grid;
  gap: 6px;
}

.aoi-select__label {
  color: var(--aoi-text-muted);
  font-size: 12px;
  font-weight: 800;
}

.aoi-select__control {
  display: flex;
  min-height: 38px;
  align-items: center;
  border: 1px solid var(--aoi-border);
  border-radius: var(--aoi-radius-control);
  background: rgba(255, 255, 255, .78);
  color: var(--aoi-icon);
  padding: 0 10px;
}

select {
  width: 100%;
  min-width: 0;
  appearance: none;
  border: 0;
  background: transparent;
  color: var(--aoi-text);
  outline: 0;
  padding: 9px 0;
}
</style>
