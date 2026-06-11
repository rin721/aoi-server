<script setup lang="ts">
defineOptions({
  inheritAttrs: false
})

const props = withDefaults(defineProps<{
  decorative?: boolean
  label?: string
  name: string
  size?: number | string
}>(), {
  decorative: false,
  label: undefined,
  size: "1em"
})

const resolvedName = computed(() => props.name.includes(":") ? props.name : `lucide:${props.name}`)
const resolvedSize = computed(() => typeof props.size === "number" ? `${props.size}px` : props.size)
const hidden = computed(() => props.decorative || !props.label)
</script>

<template>
  <span
    v-bind="$attrs"
    class="aoi-icon"
    :style="{ fontSize: resolvedSize }"
    :aria-hidden="hidden ? 'true' : undefined"
    :aria-label="!hidden ? label : undefined"
    :role="!hidden ? 'img' : undefined"
  >
    <Icon :name="resolvedName" />
  </span>
</template>

<style scoped>
.aoi-icon {
  display: inline-flex;
  width: 1em;
  height: 1em;
  align-items: center;
  justify-content: center;
  flex: 0 0 auto;
  color: currentColor;
}
</style>
