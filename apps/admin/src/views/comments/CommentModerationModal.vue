<script setup lang="ts">
import { NModal } from 'naive-ui';

import { useForumApi, type Comment } from '@/api';

export interface CommentModerationTarget {
  comment: Comment;
  status: 'hidden' | 'deleted';
}

const props = defineProps<{ target: CommentModerationTarget | null }>();
const emit = defineEmits<{
  close: [];
  success: [message: string];
  error: [message: string];
}>();

const api = useForumApi();
const saving = defineModel<boolean>('saving', { default: false });

function close() {
  if (!saving.value) emit('close');
}

async function confirm() {
  if (!props.target) return;
  saving.value = true;
  try {
    await api.setCommentStatus(props.target.comment.id, props.target.status);
    emit(
      'success',
      props.target.status === 'hidden' ? '评论已隐藏' : '评论已删除',
    );
  } catch (error) {
    emit('error', error instanceof Error ? error.message : String(error));
  } finally {
    saving.value = false;
  }
}
</script>

<template>
  <n-modal
    :show="target !== null"
    preset="dialog"
    type="warning"
    :title="target?.status === 'hidden' ? '确认隐藏评论' : '确认删除评论'"
    positive-text="确认处理"
    negative-text="取消"
    :loading="saving"
    :mask-closable="!saving"
    :close-on-esc="!saving"
    @positive-click="confirm"
    @negative-click="close"
    @mask-click="close"
  >
    处理后，该评论将显示占位，原文不再展示。
  </n-modal>
</template>
