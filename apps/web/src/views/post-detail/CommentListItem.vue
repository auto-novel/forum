<script setup lang="ts">
import { computed, defineAsyncComponent, ref } from 'vue';

import { type PostComment } from '@/api';
import MarkdownContent from '@/components/markdown/MarkdownContent.vue';
import MarkdownEditor from '@/components/markdown/MarkdownEditor.vue';
import { useUnsavedChangesGuard } from '@/composables/useUnsavedChangesGuard';
import { notifyError, notifySuccess } from '@/notifications';
import { useCommentStore } from '@/stores/comment';
import { getApiErrorMessage } from '@/utils/apiError';

import CommentActions from './CommentActions.vue';
import CommentComposer from './CommentComposer.vue';

const MarkdownHelpDialog = defineAsyncComponent(
  () => import('@/components/markdown/MarkdownHelpDialog.vue'),
);
const props = defineProps<{
  comment: PostComment;
  locked: boolean;
  postId: number;
  replying: boolean;
}>();

const emit = defineEmits<{
  reply: [comment: PostComment];
  cancelReply: [];
  created: [comment: PostComment];
  statusChanged: [id: number, status: number];
}>();

const commentStore = useCommentStore();
const editing = ref(false);
const content = ref(props.comment.content);
const submitting = ref(false);
const commentActionClass =
  'inline-flex min-h-[1.875rem] items-center rounded-sm px-[0.55rem] text-xs font-semibold text-muted hover:bg-paper hover:text-primary focus-visible:outline-2 focus-visible:outline-offset-[-2px] focus-visible:outline-primary';

const isPublished = computed(() => props.comment.status === 0);
const hasUnsavedChanges = computed(
  () => editing.value && content.value !== props.comment.content,
);

useUnsavedChangesGuard(hasUnsavedChanges);

function formatDate(value: string) {
  return new Intl.DateTimeFormat('zh-CN', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  }).format(new Date(value));
}

function startEditing() {
  if (props.replying) emit('cancelReply');
  content.value = props.comment.content;
  editing.value = true;
}

async function saveEdit() {
  const value = content.value.trim();
  if (!value || submitting.value) return;
  submitting.value = true;
  try {
    await commentStore.updateComment(props.comment.id, value);
    editing.value = false;
    notifySuccess('评论修改已保存');
  } catch (reason) {
    notifyError(await getApiErrorMessage(reason, '更新评论失败'));
  } finally {
    submitting.value = false;
  }
}
</script>

<template>
  <article
    :id="`comment-${comment.id}`"
    class="py-4"
    :class="comment.rootId != null ? 'ml-6 pl-4 sm:ml-10 sm:pl-5' : ''"
  >
    <header class="flex items-center gap-2 text-xs text-muted">
      <span class="font-medium text-ink">{{ comment.authorUsername }}</span>
      <span aria-hidden="true">·</span>
      <time :datetime="comment.createdAt">
        {{ formatDate(comment.createdAt) }}
      </time>
      <CommentActions
        v-if="isPublished && !editing"
        :comment="comment"
        :locked="locked"
        :post-id="postId"
        :replying="replying"
        @reply="emit('reply', comment)"
        @edit="startEditing"
        @status-changed="emit('statusChanged', comment.id, $event)"
      />
    </header>
    <p v-if="!isPublished" class="mt-2 text-sm text-muted">
      {{ comment.status === 2 ? '该评论已删除' : '该评论已隐藏' }}
    </p>
    <form v-else-if="editing" class="mt-2" @submit.prevent="saveEdit">
      <MarkdownEditor
        v-model="content"
        mode="comment"
        :rows="5"
        :disabled="submitting"
      />
      <div class="mt-3 flex items-center justify-between gap-3">
        <MarkdownHelpDialog mode="comment" />
        <div class="flex gap-2">
          <button
            type="button"
            :class="commentActionClass"
            :disabled="submitting"
            @click="editing = false"
          >
            取消
          </button>
          <button
            type="submit"
            class="rounded-sm bg-primary px-3 py-1.5 text-xs font-medium text-white hover:bg-primary-hover disabled:opacity-50"
            :disabled="!content.trim() || submitting"
          >
            {{ submitting ? '保存中…' : '保存' }}
          </button>
        </div>
      </div>
    </form>
    <MarkdownContent
      v-else
      class="mt-2"
      mode="comment"
      :source="comment.content"
    />

    <CommentComposer
      v-if="replying"
      :id="`comment-reply-composer-${comment.id}`"
      :post-id="postId"
      :locked="locked"
      :reply-to="comment"
      @created="emit('created', $event)"
      @cancel-reply="emit('cancelReply')"
    />
  </article>
</template>
