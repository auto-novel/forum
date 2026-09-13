<script setup lang="ts">
import { computed, defineAsyncComponent, onBeforeUnmount, ref } from 'vue';

import { authUser, type PostComment } from '@/api';
import MarkdownContent from '@/components/markdown/MarkdownContent.vue';
import MarkdownEditor from '@/components/markdown/MarkdownEditor.vue';
import ActionMenu from '@/components/ActionMenu.vue';
import ActionMenuItem from '@/components/ActionMenuItem.vue';
import ConfirmDialog from '@/components/ConfirmDialog.vue';
import { useUnsavedChangesGuard } from '@/composables/useUnsavedChangesGuard';
import { notifyError, notifySuccess } from '@/notifications';
import { useCommentStore } from '@/stores/comment';
import { getApiErrorMessage } from '@/utils/apiError';

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
const now = ref(Date.now());
const confirmationAction = ref<'delete' | 'hide'>();
const commentActionClass =
  'inline-flex min-h-[1.875rem] items-center rounded-sm px-[0.55rem] text-xs font-semibold text-muted hover:bg-paper hover:text-primary focus-visible:outline-2 focus-visible:outline-offset-[-2px] focus-visible:outline-primary';

const modificationDeadline =
  new Date(props.comment.createdAt).getTime() + 20 * 60_000;
const expiryTimer = window.setTimeout(
  () => {
    now.value = Date.now();
  },
  Math.max(0, modificationDeadline - Date.now() + 50),
);

const isPublished = computed(() => props.comment.status === 0);
const isAdmin = computed(() => authUser.value?.role === 'admin');
const isOwner = computed(() => authUser.value?.id === props.comment.authorId);
const withinModificationWindow = computed(
  () => now.value <= modificationDeadline,
);
const canEdit = computed(
  () => (isOwner.value || isAdmin.value) && withinModificationWindow.value,
);
const hasMenu = computed(() => canEdit.value || isAdmin.value);
const hasUnsavedChanges = computed(
  () => editing.value && content.value !== props.comment.content,
);
const confirmation = computed(() =>
  confirmationAction.value === 'delete'
    ? {
        title: '删除评论',
        description: '确定删除这条评论吗？删除后无法恢复。',
        confirmLabel: '删除评论',
      }
    : {
        title: '隐藏评论',
        description: '确定隐藏这条评论吗？隐藏后可在管理端恢复。',
        confirmLabel: '隐藏评论',
      },
);

useUnsavedChangesGuard(hasUnsavedChanges);

onBeforeUnmount(() => window.clearTimeout(expiryTimer));

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

async function removeComment() {
  submitting.value = true;
  try {
    await commentStore.deleteComment(props.comment.id, isAdmin.value);
    notifySuccess('评论已删除');
    emit('statusChanged', props.comment.id, 2);
  } catch (reason) {
    notifyError(await getApiErrorMessage(reason, '删除评论失败'));
  } finally {
    submitting.value = false;
  }
}

async function hideComment() {
  submitting.value = true;
  try {
    await commentStore.hideComment(props.comment.id);
    notifySuccess('评论已隐藏');
    emit('statusChanged', props.comment.id, 1);
  } catch (reason) {
    notifyError(await getApiErrorMessage(reason, '隐藏评论失败'));
  } finally {
    submitting.value = false;
  }
}

function confirmAction() {
  const action = confirmationAction.value;
  confirmationAction.value = undefined;
  if (action === 'delete') void removeComment();
  else if (action === 'hide') void hideComment();
}

function handleConfirmationOpenChange(open: boolean) {
  if (!open) confirmationAction.value = undefined;
}
</script>

<template>
  <article
    :id="`comment-${comment.id}`"
    class="py-4"
    :class="
      comment.rootId != null
        ? 'ml-6 border-l-2 border-primary-soft pl-4 sm:ml-10 sm:pl-5'
        : ''
    "
  >
    <header class="flex items-center gap-2 text-xs text-muted">
      <span class="font-medium text-ink">{{ comment.authorUsername }}</span>
      <span aria-hidden="true">·</span>
      <time :datetime="comment.createdAt">
        {{ formatDate(comment.createdAt) }}
      </time>
    </header>
    <p v-if="!isPublished" class="mt-3 text-sm text-muted">
      {{ comment.status === 2 ? '该评论已删除' : '该评论已隐藏' }}
    </p>
    <form v-else-if="editing" class="mt-3" @submit.prevent="saveEdit">
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
      class="mt-3"
      mode="comment"
      :source="comment.content"
    />

    <div v-if="isPublished && !editing" class="mt-3 flex items-center gap-1">
      <button
        v-if="authUser && !locked"
        type="button"
        :class="commentActionClass"
        :aria-expanded="replying"
        @click="emit('reply', comment)"
      >
        回复
      </button>
      <button
        v-if="canEdit"
        type="button"
        :class="commentActionClass"
        @click="startEditing"
      >
        编辑
      </button>
      <ActionMenu v-if="hasMenu" compact side="top" align="start">
        <ActionMenuItem v-if="isAdmin" @activate="confirmationAction = 'hide'">
          隐藏评论
        </ActionMenuItem>
        <ActionMenuItem
          danger
          :disabled="submitting"
          @activate="confirmationAction = 'delete'"
        >
          删除评论
        </ActionMenuItem>
      </ActionMenu>
    </div>
    <CommentComposer
      v-if="replying"
      :id="`comment-reply-composer-${comment.id}`"
      :post-id="postId"
      :locked="locked"
      :reply-to="comment"
      @created="emit('created', $event)"
      @cancel-reply="emit('cancelReply')"
    />
    <ConfirmDialog
      :open="confirmationAction != null"
      :title="confirmation.title"
      :description="confirmation.description"
      :confirm-label="confirmation.confirmLabel"
      :loading="submitting"
      danger
      @update:open="handleConfirmationOpenChange"
      @confirm="confirmAction"
    />
  </article>
</template>
