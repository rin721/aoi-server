<script setup lang="ts">
import type { AoiFieldAppearance } from "~/types/ui"

export interface AoiSelectOption {
  value: string
  label: string
  disabled?: boolean
}

type AoiSelectMenuPositioning = "absolute" | "fixed" | "popover"
type MaterialSelect = HTMLElement & {
  label?: string
  requestUpdate?: () => void
  select?: (value: string) => void
  selectedIndex?: number
  selectedOptions?: MaterialSelectOption[]
  value?: string
}
type MaterialSelectOption = HTMLElement & { displayText?: string, requestUpdate?: () => void, selected?: boolean, typeaheadText?: string, value?: string }

const props = withDefaults(defineProps<{
  modelValue?: string
  options?: AoiSelectOption[]
  label?: string
  appearance?: AoiFieldAppearance
  disabled?: boolean
  menuPositioning?: AoiSelectMenuPositioning
}>(), {
  modelValue: "",
  options: () => [],
  label: undefined,
  appearance: "filled",
  disabled: false,
  menuPositioning: "popover"
})

const emit = defineEmits<{
  "update:modelValue": [value: string]
}>()

const tagName = computed(() => props.appearance === "outlined" ? "md-outlined-select" : "md-filled-select")
const selectId = `aoi-select-${Math.random().toString(36).slice(2)}`
const selectRef = ref<MaterialSelect | null>(null)
const menuOpen = ref(false)
const layer = useAoiLayer("menu", menuOpen)

function onChange(event: Event) {
  const selectEl = event.currentTarget as MaterialSelect
  const selectedValue = selectEl.selectedOptions?.[0]?.value ?? selectEl.value ?? ""
  emit("update:modelValue", selectedValue)
}

function onOpening() {
  menuOpen.value = true
}

function onClosed() {
  menuOpen.value = false
}

function syncMaterialSelect() {
  nextTick(() => {
    const selectEl = resolveSelectElement()
    if (!selectEl) {
      return
    }

    selectEl.label = props.label || ""

    const optionElements = Array.from(selectEl.querySelectorAll("md-select-option")) as MaterialSelectOption[]
    props.options.forEach((option, index) => {
      const optionEl = optionElements[index]
      if (!optionEl) {
        return
      }
      optionEl.value = option.value
      optionEl.selected = option.value === props.modelValue
      optionEl.displayText = option.label
      optionEl.typeaheadText = option.label
      optionEl.requestUpdate?.()
    })

    const selectedIndex = props.options.findIndex((option) => option.value === props.modelValue)
    if (selectedIndex >= 0) {
      selectEl.select?.(props.modelValue)
      if (typeof selectEl.selectedIndex === "number") {
        selectEl.selectedIndex = selectedIndex
      }
    }

    selectEl.requestUpdate?.()
  })
}

function resolveSelectElement() {
  if (selectRef.value instanceof HTMLElement) {
    return selectRef.value
  }
  if (!import.meta.client) {
    return null
  }
  return document.querySelector(`[data-aoi-select-id="${selectId}"]`) as MaterialSelect | null
}

onMounted(syncMaterialSelect)
onUpdated(syncMaterialSelect)
watch(() => [props.modelValue, props.label, props.options], syncMaterialSelect, { deep: true })
</script>

<template>
  <component
    ref="selectRef"
    :is="tagName"
    class="aoi-text-field"
    :data-aoi-select-id="selectId"
    :value.attr="modelValue"
    :label.attr="label"
    :disabled="disabled || undefined"
    :menu-positioning="menuPositioning"
    :style="layer.style.value"
    @change="onChange"
    @opening="onOpening"
    @opened="onOpening"
    @closed="onClosed"
    @closing="onClosed"
  >
    <md-select-option
      v-for="(option, index) in options"
      :key="option.value"
      :value.attr="option.value"
      :disabled="option.disabled || undefined"
      :display-text.attr="option.label"
      :selected.attr="option.value === modelValue ? '' : null"
      :typeahead-text.attr="option.label"
    >
      <div slot="headline">{{ option.label }}</div>
    </md-select-option>
  </component>
</template>


