<script setup lang="ts">
const { t } = useI18n()

const props = withDefaults(defineProps<{
  authorName: string
  disabled?: boolean
  maxAuthorLength?: number
  maxBodyLength?: number
}>(), {
  disabled: false,
  maxAuthorLength: 24,
  maxBodyLength: 500
})

const emit = defineEmits<{
  submit: [body: string]
  "update:authorName": [value: string]
}>()

const body = ref("")

const localAuthorName = computed({
  get: () => props.authorName,
  set: (value) => emit("update:authorName", value)
})

const trimmedBody = computed(() => body.value.trim())
const bodyLength = computed(() => body.value.length)
const isBodyTooLong = computed(() => bodyLength.value > props.maxBodyLength)
const canSubmit = computed(() => {
  return !props.disabled
    && localAuthorName.value.trim().length > 0
    && trimmedBody.value.length > 0
    && !isBodyTooLong.value
})

function submitComment() {
  if (!canSubmit.value) {
    return
  }

  emit("submit", trimmedBody.value)
  body.value = ""
}
</script>

<template>
  <AoiSurface
    as="form"
    class="comment-composer"
    surface="card"
    padding="md"
    reveal="rise"
    @submit.prevent="submitComment"
  >
    <div class="comment-composer__fields">
      <AoiTextField
        v-model="localAuthorName"
        appearance="outlined"
        :label="t('comments.authorLabel')"
        :disabled="disabled"
        :max-length="maxAuthorLength"
      />
      <AoiTextField
        v-model="body"
        appearance="outlined"
        :label="t('comments.bodyLabel')"
        :placeholder="t('comments.placeholder')"
        :disabled="disabled"
        :max-length="maxBodyLength"
        :supporting-text="`${bodyLength}/${maxBodyLength}`"
        :error-text="isBodyTooLong ? t('comments.tooLong') : undefined"
        multiline
        :rows="4"
      />
    </div>

    <AoiActionBar class="comment-composer__actions" align="between">
      <span class="comment-composer__hint">
        {{ t("comments.localOnly") }}
      </span>
      <AoiButton
        type="submit"
        icon="send"
        :disabled="!canSubmit"
      >
        {{ t("comments.publish") }}
      </AoiButton>
    </AoiActionBar>
  </AoiSurface>
</template>

<style scoped>
.comment-composer {
  display: grid;
  gap: 12px;
}

.comment-composer__fields {
  display: grid;
  gap: 12px;
}

.comment-composer__hint {
  color: var(--aoi-text-muted);
  font-size: 13px;
  line-height: 1.6;
}

@media (max-width: 620px) {
  .comment-composer__actions {
    align-items: stretch;
    flex-direction: column;
  }
}
</style>


