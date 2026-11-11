<script setup lang="ts">
import { ref, useId } from 'vue';

import type { PostComment } from '@/api';
import {
  CommentRulesNotice,
  Toast,
  XButton,
  XMarkdownHelpDialog,
  getApiErrorMessage,
  useMarkdownDraft,
  useWebKit,
} from '@novelia/web-kit';
import MarkdownEditor from '@/components/markdown/MarkdownEditor.vue';
import { useCommentStore } from '@/stores/comment';
import { useCommentValidation } from './useCommentValidation';

const props = defineProps<{
  postId: number;
  locked: boolean;
  replyTo?: PostComment;
}>();

const emit = defineEmits<{
  created: [comment: PostComment];
  cancelReply: [];
}>();

const { whoami } = useWebKit();
const commentStore = useCommentStore();
const { text: content, prepareSubmission } = useMarkdownDraft({
  key: () => {
    const user = whoami.value.user;
    return user
      ? `forum:comment:${props.postId}:${user.id}:${props.replyTo ? `reply:${props.replyTo.id}` : 'root'}`
      : undefined;
  },
  initialValue: () => {
    const replyTo = props.replyTo;
    return replyTo?.rootId != null && replyTo.authorId !== whoami.value.user?.id
      ? `@${replyTo.authorUsername} `
      : '';
  },
});
const submitting = ref(false);
const commentHintId = useId();
const commentPlaceholders = [
  '允许别人喜欢你不喜欢的，也允许别人不喜欢你喜欢的',
  '就事论事，请勿人身攻击',
];
const commentPlaceholder =
  commentPlaceholders[Math.floor(Math.random() * commentPlaceholders.length)];
const {
  length: commentLength,
  hint: commentHint,
  canSubmit,
} = useCommentValidation(content, submitting);
async function submitComment() {
  const value = content.value.trim();
  if (!canSubmit.value || props.locked) return;
  const completeSubmission = prepareSubmission();
  submitting.value = true;
  await Toast.promise(
    commentStore.createComment(props.postId, {
      content: value,
      rootId: props.replyTo?.rootId ?? props.replyTo?.id,
    }),
    {
      success: '评论已发表',
      error: (reason) => getApiErrorMessage(reason, '评论发布失败'),
    },
  )
    .then((comment) => {
      completeSubmission();
      emit('created', comment);
    })
    .catch(() => {});
  submitting.value = false;
}
</script>

<template>
  <section :class="replyTo ? 'mt-3' : 'border-b border-divider py-4'">
    <div
      v-if="locked"
      :class="['py-2 text-sm text-warning', { 'mt-4': replyTo }]"
    >
      评论区已锁定，暂时无法发表新评论。
    </div>

    <div
      v-else-if="!whoami.isSignedIn"
      :class="['py-2 text-sm text-muted', { 'mt-4': replyTo }]"
    >
      登录后即可参与评论，请使用页面右上角的登录入口。
    </div>

    <form v-else @submit.prevent="submitComment">
      <CommentRulesNotice :to="{ name: 'community-rules' }" class="mb-3" />
      <MarkdownEditor
        v-model="content"
        mode="comment"
        :rows="3"
        :placeholder="
          replyTo ? `回复 @${replyTo.authorUsername}…` : commentPlaceholder
        "
        :disabled="submitting"
        :described-by="commentHint ? commentHintId : undefined"
        :invalid="commentLength > 1000"
      />
      <div class="mt-2 flex items-center justify-between gap-3 text-xs">
        <p
          v-if="commentHint"
          :id="commentHintId"
          :class="content ? 'text-warning' : 'text-muted'"
          aria-live="polite"
        >
          {{ commentHint }}
        </p>
        <p class="ml-auto text-muted">{{ commentLength }} / 1000</p>
      </div>
      <div class="mt-3 flex flex-wrap items-center justify-between gap-3">
        <XMarkdownHelpDialog mode="comment" />
        <div class="flex items-center gap-2">
          <XButton
            v-if="replyTo"
            color="neutral"
            variant="outline"
            :disabled="submitting"
            @click="emit('cancelReply')"
          >
            取消
          </XButton>
          <XButton type="submit" :disabled="!canSubmit">
            {{ submitting ? '发表中…' : '发表' }}
          </XButton>
        </div>
      </div>
    </form>
  </section>
</template>
