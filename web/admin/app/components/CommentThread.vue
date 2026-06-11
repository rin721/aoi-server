<script setup lang="ts">
import type { CommentSortMode, LocalComment } from "~/types/comments"

const props = withDefaults(defineProps<{
  comments: LocalComment[]
  hydrated?: boolean
  sortMode?: CommentSortMode
}>(), {
  hydrated: false,
  sortMode: "newest"
})

const emit = defineEmits<{
  delete: [commentId: string]
  edit: [commentId: string, body: string]
  "update:sortMode": [value: CommentSortMode]
}>()

const { t } = useI18n()

const sortValue = computed({
  get: () => props.sortMode,
  set: (value) => emit("update:sortMode", value as CommentSortMode)
})

const sortOptions = computed(() => [
  { label: t("comments.newest"), value: "newest" },
  { label: t("comments.oldest"), value: "oldest" }
])
</script>

<template>
  <section class="comment-thread" aria-labelledby="comment-thread-title">
    <AoiSection
      as="div"
      :title="t('comments.title')"
      :description="t('common.counts.comments', { count: comments.length })"
      title-id="comment-thread-title"
      :reveal="false"
    >
      <template #actions>
        <AoiSelect
          v-model="sortValue"
          class="comment-thread__sort"
          :label="t('comments.sortLabel')"
          appearance="outlined"
          :options="sortOptions"
          :disabled="!hydrated || comments.length < 2"
        />
      </template>
    </AoiSection>

    <PageState
      v-if="hydrated && comments.length === 0"
      icon="message-circle"
      :title="t('comments.emptyTitle')"
      :description="t('comments.emptyDescription')"
    />

    <AoiContentGrid v-else-if="hydrated" min-width="100%" gap="compact" :mobile-columns="1">
      <AoiReveal
        v-for="(comment, index) in comments"
        :key="comment.id"
        class="comment-thread__item"
        :index="index"
        variant="rise"
      >
        <CommentItem
          :comment="comment"
          @delete="emit('delete', $event)"
          @edit="(commentId, body) => emit('edit', commentId, body)"
        />
      </AoiReveal>
    </AoiContentGrid>
  </section>
</template>

<style scoped>
.comment-thread {
  display: grid;
  gap: 12px;
}

.comment-thread__sort {
  width: min(180px, 100%);
}

.comment-thread__item {
  min-width: 0;
}

@media (max-width: 620px) {
  .comment-thread__sort {
    width: 100%;
  }
}
</style>


