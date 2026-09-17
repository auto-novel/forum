<script setup lang="ts">
import {
  CommentOutlined,
  LockOutlined,
  MoreHorizOutlined,
  PushPinOutlined,
  VisibilityOutlined,
} from '@vicons/material';
import { NButton, NIcon, NTag, NText } from 'naive-ui';

import type { PostSummary } from '@/api';

defineProps<{
  post: PostSummary;
  categoryName: string;
}>();

const emit = defineEmits<{
  moderate: [post: PostSummary];
  reviewComments: [post: PostSummary];
}>();

function formatDate(value: string) {
  return new Intl.DateTimeFormat('zh-CN', {
    dateStyle: 'medium',
    timeStyle: 'short',
  }).format(new Date(value));
}
</script>

<template>
  <div class="post-row">
    <div class="post-content">
      <div class="post-badges">
        <n-tag size="small" :bordered="false">{{ categoryName }}</n-tag>
        <n-tag
          v-if="post.status === 1"
          size="small"
          type="warning"
          :bordered="false"
        >
          隐藏
        </n-tag>
        <n-tag
          v-if="post.status === 2"
          size="small"
          type="error"
          :bordered="false"
        >
          已删除
        </n-tag>
        <n-tag
          v-if="post.pinOrder != null"
          size="small"
          type="warning"
          :bordered="false"
        >
          <template #icon><n-icon :component="PushPinOutlined" /></template>
          置顶 {{ post.pinOrder }}
        </n-tag>
        <n-tag
          v-if="post.commentsLocked"
          size="small"
          type="error"
          :bordered="false"
        >
          <template #icon><n-icon :component="LockOutlined" /></template>
          已锁评
        </n-tag>
      </div>
      <n-text strong class="post-title">{{ post.title }}</n-text>
      <div class="post-footer">
        <n-text depth="3">
          {{ post.authorUsername }} · {{ formatDate(post.activeAt) }}
        </n-text>
        <div class="post-metrics">
          <span>
            <n-icon :component="VisibilityOutlined" />
            {{ post.viewsCount }}
          </span>
          <span>
            <n-icon :component="CommentOutlined" />
            {{ post.commentsCount }}
          </span>
          <n-button
            text
            type="primary"
            size="small"
            @click="emit('reviewComments', post)"
          >
            审核评论
          </n-button>
        </div>
      </div>
    </div>
    <n-button
      circle
      quaternary
      aria-label="管理帖子"
      @click="emit('moderate', post)"
    >
      <template #icon><n-icon :component="MoreHorizOutlined" /></template>
    </n-button>
  </div>
</template>

<style scoped>
.post-row,
.post-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.post-row {
  align-items: flex-start;
}

.post-content {
  min-width: 0;
  display: flex;
  flex: 1;
  flex-direction: column;
  gap: 8px;
}

.post-badges,
.post-metrics {
  display: flex;
  align-items: center;
  gap: 7px;
}

.post-title {
  font-size: 17px;
  line-height: 1.4;
}

.post-footer {
  margin-top: 3px;
}

.post-metrics span {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  color: var(--n-text-color-3);
  font-size: 12px;
}

@media (max-width: 680px) {
  .post-footer {
    align-items: flex-start;
    flex-direction: column;
    gap: 6px;
  }
}
</style>
