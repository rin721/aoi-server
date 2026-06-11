<script setup lang="ts">
const props = withDefaults(defineProps<{
  autocomplete?: string
  disabled?: boolean
  error?: string
  icon?: string
  label: string
  modelValue?: string | number
  placeholder?: string
  rows?: number
  type?: string
}>(), {
  autocomplete: undefined,
  disabled: false,
  error: undefined,
  icon: undefined,
  modelValue: "",
  placeholder: undefined,
  rows: undefined,
  type: "text"
})

const emit = defineEmits<{
  enter: [event: KeyboardEvent]
  "update:modelValue": [value: string]
}>()

function updateValue(event: Event) {
  emit("update:modelValue", (event.target as HTMLInputElement | HTMLTextAreaElement).value)
}

function onKeydown(event: KeyboardEvent) {
  if (event.key === "Enter") {
    emit("enter", event)
  }
}
</script>

<template>
  <label class="aoi-field">
    <span class="aoi-field__label">{{ label }}</span>
    <span class="aoi-field__control" :class="{ 'aoi-field__control--error': error }">
      <AoiIcon v-if="icon" :name="icon" decorative />
      <textarea
        v-if="type === 'textarea'"
        :value="modelValue"
        :disabled="disabled"
        :placeholder="placeholder"
        :rows="rows || 4"
        @input="updateValue"
        @keydown="onKeydown"
      />
      <input
        v-else
        :value="modelValue"
        :autocomplete="autocomplete"
        :disabled="disabled"
        :placeholder="placeholder"
        :type="type"
        @input="updateValue"
        @keydown="onKeydown"
      >
    </span>
    <span v-if="error" class="aoi-field__error">{{ error }}</span>
  </label>
</template>

<style scoped>
.aoi-field {
  display: grid;
  gap: 6px;
}

.aoi-field__label {
  color: var(--aoi-text-muted);
  font-size: 12px;
  font-weight: 800;
}

.aoi-field__control {
  display: flex;
  min-height: 38px;
  align-items: center;
  gap: 8px;
  border: 1px solid var(--aoi-border);
  border-radius: var(--aoi-radius-control);
  background: rgba(255, 255, 255, .78);
  color: var(--aoi-icon);
  padding: 0 10px;
  transition:
    border-color var(--aoi-motion-fast) var(--aoi-ease-out),
    box-shadow var(--aoi-motion-fast) var(--aoi-ease-out);
}

.aoi-field__control:focus-within {
  border-color: color-mix(in srgb, var(--aoi-accent-60) 55%, var(--aoi-border));
  box-shadow: 0 0 0 3px var(--aoi-focus);
}

.aoi-field__control--error {
  border-color: color-mix(in srgb, var(--aoi-danger) 62%, var(--aoi-border));
}

input,
textarea {
  width: 100%;
  min-width: 0;
  border: 0;
  background: transparent;
  color: var(--aoi-text);
  outline: 0;
  padding: 9px 0;
}

textarea {
  resize: vertical;
}

input::placeholder,
textarea::placeholder {
  color: color-mix(in srgb, var(--aoi-text-muted) 64%, transparent);
}

.aoi-field__error {
  color: var(--aoi-danger);
  font-size: 12px;
}
</style>
