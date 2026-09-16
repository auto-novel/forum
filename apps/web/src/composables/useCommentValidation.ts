import { computed, type Ref } from 'vue';

const MAX_COMMENT_LENGTH = 1000;

export function useCommentValidation(
  content: Ref<string>,
  submitting: Ref<boolean>,
) {
  const length = computed(() => Array.from(content.value.trim()).length);
  const hint = computed(() => {
    if (!length.value) return '请输入评论';
    if (length.value > MAX_COMMENT_LENGTH) return '评论不能超过 1000 字';
    return '';
  });
  const canSubmit = computed(
    () =>
      length.value > 0 &&
      length.value <= MAX_COMMENT_LENGTH &&
      !submitting.value,
  );

  return { length, hint, canSubmit };
}
