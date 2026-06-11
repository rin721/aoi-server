<script setup lang="ts">
withDefaults(defineProps<{
  description?: string
  icon?: string
  title: string
}>(), {
  description: undefined,
  icon: undefined
})
</script>

<template>
  <header class="page-header">
    <div v-if="icon" class="page-header__icon" aria-hidden="true">
      <AoiIcon :name="icon" decorative />
    </div>
    <div class="page-header__copy">
      <h1>{{ title }}</h1>
      <p v-if="description">{{ description }}</p>
    </div>
    <div v-if="$slots.actions" class="page-header__actions">
      <slot name="actions" />
    </div>
  </header>
</template>

<style scoped>
.page-header {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  margin-bottom: 18px;
}

.page-header__icon {
  display: grid;
  width: 40px;
  height: 40px;
  place-items: center;
  border: 1px solid var(--aoi-border);
  border-radius: var(--aoi-radius-control);
  background: var(--aoi-accent-10);
  color: var(--aoi-accent-60);
  font-size: 22px;
}

.page-header__copy {
  min-width: 0;
  flex: 1;
}

h1 {
  margin: 0;
  font-size: 26px;
  line-height: 1.2;
}

p {
  max-width: 760px;
  margin: 7px 0 0;
  color: var(--aoi-text-muted);
  line-height: 1.7;
}

.page-header__actions {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: 8px;
}

@media (max-width: 720px) {
  h1 {
    font-size: 22px;
  }

  .page-header {
    flex-wrap: wrap;
  }
}
</style>
