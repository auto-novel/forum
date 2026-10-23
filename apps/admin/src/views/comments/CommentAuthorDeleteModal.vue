<script setup lang="ts">
import { XConfirmDialog } from '@novelia/web-kit';

import { useForumApi, type Comment } from '@/api';

const props = defineProps<{ target: Comment | null }>();
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
    await api.deleteCommentsByAuthor(target.authorId);
    emit('success', `已删除 ${target.authorUsername} 的全部评论`);
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
    title="确认删除该作者全部评论"
    :description="`将把 ${target?.authorUsername ?? ''}（ID: ${target?.authorId ?? '-'}）的全部评论（含回复）标记为删除，公开页面会显示占位；管理员仍可查看原文并逐条恢复。`"
    confirm-label="确认删除"
    danger
    :loading="saving"
    @update:open="(open) => !open && close()"
    @confirm="confirm"
  />
</template>
