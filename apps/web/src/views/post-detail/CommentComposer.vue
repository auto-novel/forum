<script setup lang="ts">
import { computed, defineAsyncComponent, ref, watch } from 'vue';

import { authUser, type PostComment } from '@/api';
import MarkdownEditor from '@/components/markdown/MarkdownEditor.vue';
import { notifyError, notifySuccess } from '@/notifications';
import { useCommentStore } from '@/stores/comment';
import { useDraftStore } from '@/stores/draft';
import { getApiErrorMessage } from '@/utils/apiError';

const MarkdownHelpDialog = defineAsyncComponent(
  () => import('@/components/markdown/MarkdownHelpDialog.vue'),
);

const props = defineProps<{
  postId: number;
  locked: boolean;
  replyTo?: PostComment;
}>();

const emit = defineEmits<{
  created: [comment: PostComment];
  cancelReply: [];
}>();

const commentStore = useCommentStore();
const draftStore = useDraftStore();
const content = ref('');
const submitting = ref(false);
const draftKey = computed(() =>
  authUser.value
    ? `${props.postId}:${authUser.value.id}:${props.replyTo?.rootId ?? props.replyTo?.id ?? 'root'}`
    : '',
);

watch(
  draftKey,
  (key) => {
    const draft = key ? draftStore.getCommentDraft(key) : '';
    content.value =
      draft || (props.replyTo ? `@${props.replyTo.authorUsername} ` : '');
  },
  { immediate: true },
);

watch(content, (value) => {
  if (draftKey.value) draftStore.saveCommentDraft(draftKey.value, value);
});

async function submitComment() {
  const value = content.value.trim();
  if (!value || submitting.value || props.locked) return;
  submitting.value = true;
  try {
    const comment = await commentStore.createComment(props.postId, {
      content: value,
      rootId: props.replyTo?.rootId ?? props.replyTo?.id,
    });
    content.value = '';
    draftStore.clearCommentDraft(draftKey.value);
    notifySuccess('评论已发表');
    emit('created', comment);
  } catch (error) {
    notifyError(await getApiErrorMessage(error, '评论发布失败'));
  } finally {
    submitting.value = false;
  }
}
</script>

<template>
  <section :class="replyTo ? 'mt-3' : 'mt-5 border-t border-divider pt-5'">
    <h2 v-if="!replyTo" class="font-semibold text-ink">发表评论</h2>

    <div v-if="locked" class="mt-4 py-2 text-sm text-orange-700">
      评论区已锁定，暂时无法发表新评论。
    </div>

    <div v-else-if="!authUser" class="mt-4 py-2 text-sm text-muted">
      登录后即可参与评论，请使用页面右上角的登录入口。
    </div>

    <form v-else :class="{ 'mt-4': !replyTo }" @submit.prevent="submitComment">
      <MarkdownEditor
        v-model="content"
        mode="comment"
        :rows="3"
        :placeholder="
          replyTo
            ? `回复 @${replyTo.authorUsername}…`
            : '友善交流，分享你的想法…'
        "
        :disabled="submitting"
      />
      <div class="mt-3 flex flex-wrap items-center justify-between gap-3">
        <MarkdownHelpDialog mode="comment" />
        <div class="flex items-center gap-2">
          <button
            v-if="replyTo"
            type="button"
            class="rounded-sm border border-border px-4 py-2 text-sm font-medium text-muted transition-colors hover:border-primary hover:text-primary disabled:cursor-not-allowed disabled:opacity-50"
            :disabled="submitting"
            @click="emit('cancelReply')"
          >
            取消
          </button>
          <button
            type="submit"
            class="rounded-sm bg-primary px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-primary-hover disabled:cursor-not-allowed disabled:opacity-50"
            :disabled="!content.trim() || submitting"
          >
            {{ submitting ? '发表中…' : '发表' }}
          </button>
        </div>
      </div>
    </form>
  </section>
</template>
