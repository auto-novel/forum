<script setup lang="ts">
import { XConfirmDialog } from '@novelia/web-kit';

import { useForumApi, type Comment, type CommentStatusName } from '@/api';

export interface CommentModerationTarget {
  comment: Comment;
  status: CommentStatusName;
}

const props = defineProps<{ target: CommentModerationTarget | null }>();
const emit = defineEmits<{
  close: [];
  success: [message: string];
  error: [message: string];
}>();

const api = useForumApi();
const saving = defineModel<boolean>('saving', { default: false });
let confirming = false;

function close() {
  if (!saving.value && !confirming) emit('close');
}

async function confirm() {
  const target = props.target;
  if (!target || saving.value || confirming) return;
  confirming = true;
  saving.value = true;
  try {
    await api.setCommentStatus(
      target.comment.id,
      target.status,
      target.comment.subjectType,
    );
    emit(
      'success',
      target.status === 'published'
        ? '评论已恢复'
        : target.status === 'hidden'
          ? '评论已隐藏'
          : '评论已删除',
    );
  } catch (error) {
    emit('error', error instanceof Error ? error.message : String(error));
  } finally {
    saving.value = false;
    confirming = false;
  }
}
</script>

<template>
  <XConfirmDialog
    :open="target !== null"
    :title="
      target?.status === 'published'
        ? '确认恢复评论'
        : target?.status === 'hidden'
          ? '确认隐藏评论'
          : '确认删除评论'
    "
    :description="`${target?.comment.authorUsername} 的评论 #${target?.comment.id}。${
      target?.status === 'published'
        ? '恢复后，评论原文将重新公开展示。'
        : '处理后，公开页面将显示占位；管理员仍可查看原文并恢复。'
    }`"
    confirm-label="确认处理"
    :danger="target?.status !== 'published'"
    :loading="saving"
    @update:open="(open) => !open && close()"
    @confirm="confirm"
  />
</template>
