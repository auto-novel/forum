<script setup lang="ts">
import { DeleteOutlineOutlined, VisibilityOffOutlined } from '@vicons/material';
import { NButton, NIcon, NTag, NText } from 'naive-ui';

import type { Comment } from '@/api';

defineProps<{ comment: Comment }>();
const emit = defineEmits<{
  moderate: [comment: Comment, status: 'hidden' | 'deleted'];
}>();

function formatDate(value: string) {
  return new Intl.DateTimeFormat('zh-CN', {
    dateStyle: 'medium',
    timeStyle: 'short',
  }).format(new Date(value));
}
</script>

<template>
  <div class="comment-row">
    <div class="comment-header">
      <div class="comment-author">
        <n-text strong>{{ comment.authorUsername }}</n-text>
        <n-text depth="3" class="comment-meta">
          #{{ comment.id }} · {{ formatDate(comment.createdAt) }}
        </n-text>
      </div>
      <n-tag v-if="comment.rootId" size="small" :bordered="false">
        回复 #{{ comment.rootId }}
      </n-tag>
    </div>
    <n-text class="comment-content">
      {{
        comment.status === 0
          ? comment.content
          : comment.status === 2
            ? '该评论已删除'
            : '该评论已隐藏'
      }}
    </n-text>
    <div class="comment-actions">
      <n-button
        size="small"
        secondary
        type="warning"
        @click="emit('moderate', comment, 'hidden')"
      >
        <template #icon>
          <n-icon :component="VisibilityOffOutlined" />
        </template>
        隐藏
      </n-button>
      <n-button
        size="small"
        secondary
        type="error"
        @click="emit('moderate', comment, 'deleted')"
      >
        <template #icon>
          <n-icon :component="DeleteOutlineOutlined" />
        </template>
        删除
      </n-button>
    </div>
  </div>
</template>

<style scoped>
.comment-row {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.comment-header,
.comment-actions {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.comment-author {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 3px;
}

.comment-meta {
  font-size: 12px;
}

.comment-content {
  line-height: 1.72;
  white-space: pre-wrap;
  overflow-wrap: anywhere;
}

.comment-actions {
  justify-content: flex-end;
}
</style>
