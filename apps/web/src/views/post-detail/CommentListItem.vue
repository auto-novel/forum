<script setup lang="ts">
import { computed, onBeforeUnmount, ref, useTemplateRef } from 'vue';

import {
  authUser,
  deletePostComment,
  setPostCommentStatus,
  updatePostComment,
  type PostComment,
} from '@/api';
import MarkdownContent from '@/components/markdown/MarkdownContent.vue';
import MarkdownEditor from '@/components/markdown/MarkdownEditor.vue';

const props = defineProps<{
  comment: PostComment;
  locked: boolean;
}>();

const emit = defineEmits<{
  reply: [comment: PostComment];
  updated: [comment: PostComment];
  deleted: [id: number];
}>();

const menu = useTemplateRef<HTMLDetailsElement>('menu');
const editing = ref(false);
const content = ref(props.comment.content);
const submitting = ref(false);
const actionError = ref('');
const now = ref(Date.now());

const modificationDeadline =
  new Date(props.comment.createdAt).getTime() + 20 * 60_000;
const expiryTimer = window.setTimeout(
  () => {
    now.value = Date.now();
  },
  Math.max(0, modificationDeadline - Date.now() + 50),
);

const isAdmin = computed(() => authUser.value?.role === 'admin');
const isOwner = computed(() => authUser.value?.id === props.comment.authorId);
const withinModificationWindow = computed(
  () => now.value <= modificationDeadline,
);
const canEdit = computed(
  () => (isOwner.value || isAdmin.value) && withinModificationWindow.value,
);
const hasMenu = computed(() => canEdit.value || isAdmin.value);

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

async function errorMessage(reason: unknown, fallback: string) {
  if (reason && typeof reason === 'object' && 'response' in reason) {
    const response = (reason as { response?: Response }).response;
    if (response) {
      try {
        return (await response.text()) || fallback;
      } catch {
        // Use the fallback below.
      }
    }
  }
  return reason instanceof Error ? reason.message : fallback;
}

function closeMenu() {
  if (menu.value) menu.value.open = false;
}

function startEditing() {
  closeMenu();
  content.value = props.comment.content;
  actionError.value = '';
  editing.value = true;
}

async function saveEdit() {
  const value = content.value.trim();
  if (!value || submitting.value) return;
  submitting.value = true;
  actionError.value = '';
  try {
    const comment = await updatePostComment(props.comment.id, value);
    editing.value = false;
    emit('updated', comment);
  } catch (reason) {
    actionError.value = await errorMessage(reason, '更新评论失败');
  } finally {
    submitting.value = false;
  }
}

async function removeComment() {
  closeMenu();
  if (!window.confirm('确定删除这条评论吗？删除后无法恢复。')) return;
  submitting.value = true;
  actionError.value = '';
  try {
    if (isAdmin.value) await setPostCommentStatus(props.comment.id, 'deleted');
    else await deletePostComment(props.comment.id);
    emit('deleted', props.comment.id);
  } catch (reason) {
    actionError.value = await errorMessage(reason, '删除评论失败');
  } finally {
    submitting.value = false;
  }
}

async function hideComment() {
  closeMenu();
  if (!window.confirm('确定隐藏这条评论吗？隐藏后可在管理端恢复。')) return;
  submitting.value = true;
  actionError.value = '';
  try {
    await setPostCommentStatus(props.comment.id, 'hidden');
    emit('deleted', props.comment.id);
  } catch (reason) {
    actionError.value = await errorMessage(reason, '隐藏评论失败');
  } finally {
    submitting.value = false;
  }
}
</script>

<template>
  <article
    :id="`comment-${comment.id}`"
    class="px-4 py-5 sm:px-6"
    :class="
      comment.rootId != null
        ? 'ml-6 border-l-2 border-primary-soft sm:ml-12'
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
    <form v-if="editing" class="mt-3" @submit.prevent="saveEdit">
      <MarkdownEditor
        v-model="content"
        mode="comment"
        :rows="5"
        :disabled="submitting"
      />
      <div class="mt-3 flex justify-end gap-2">
        <button
          type="button"
          class="comment-action"
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
    </form>
    <MarkdownContent
      v-else
      class="mt-3"
      mode="comment"
      :source="comment.content"
    />

    <div v-if="!editing" class="mt-3 flex items-center gap-1">
      <button
        v-if="authUser && !locked"
        type="button"
        class="comment-action"
        @click="emit('reply', comment)"
      >
        回复
      </button>
      <button
        v-if="canEdit"
        type="button"
        class="comment-action"
        @click="startEditing"
      >
        编辑
      </button>
      <details v-if="hasMenu" ref="menu" class="relative">
        <summary class="comment-action cursor-pointer list-none">
          更多 ···
        </summary>
        <div
          class="absolute bottom-[calc(100%+0.35rem)] left-0 z-20 w-32 rounded-md border border-border bg-surface p-1 shadow-xl"
        >
          <button
            v-if="isAdmin"
            type="button"
            class="comment-menu-item"
            @click="hideComment"
          >
            隐藏评论
          </button>
          <button
            type="button"
            class="comment-menu-item text-red-600"
            :disabled="submitting"
            @click="removeComment"
          >
            删除评论
          </button>
        </div>
      </details>
    </div>
    <p v-if="actionError" class="mt-2 text-xs text-red-600" role="alert">
      {{ actionError }}
    </p>
  </article>
</template>

<style scoped>
.comment-action {
  display: inline-flex;
  min-height: 1.875rem;
  align-items: center;
  border-radius: 0.25rem;
  padding-inline: 0.55rem;
  color: var(--color-muted);
  font-size: 0.75rem;
  font-weight: 600;
}

.comment-action:hover {
  background: var(--color-paper);
  color: var(--color-primary);
}

.comment-menu-item {
  display: block;
  width: 100%;
  border-radius: 0.25rem;
  padding: 0.45rem 0.6rem;
  font-size: 0.75rem;
  text-align: left;
}

.comment-menu-item:hover {
  background: var(--color-paper);
}

.comment-action:focus-visible,
.comment-menu-item:focus-visible {
  outline: 2px solid var(--color-primary);
  outline-offset: -2px;
}

summary::-webkit-details-marker {
  display: none;
}
</style>
