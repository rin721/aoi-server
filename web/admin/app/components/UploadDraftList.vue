<script setup lang="ts">
import type { UploadDraft } from "~/types/upload"

const props = defineProps<{
  activeId?: string
  drafts: UploadDraft[]
}>()

const emit = defineEmits<{
  select: [id: string]
}>()

const { t } = useI18n()
</script>

<template>
  <div class="upload-draft-list" :aria-label="t('upload.draftsTitle')">
    <AoiChoiceCard
      v-for="draft in props.drafts"
      :key="draft.id"
      :value="draft.id"
      :title="draft.title || t('common.states.unnamedDraft')"
      :description="`${draft.status === 'queued-local' ? t('common.states.queuedLocal') : t('common.states.draft')} · ${t('common.counts.tags', { count: draft.tags.length })}`"
      variant="compact"
      :selected="draft.id === props.activeId"
      @select="emit('select', $event)"
    />
  </div>
</template>

<style scoped>
.upload-draft-list {
  display: grid;
  gap: 8px;
}
</style>


