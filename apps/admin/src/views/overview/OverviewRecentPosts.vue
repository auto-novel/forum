<script setup lang="ts">
import {
  NButton,
  NCard,
  NList,
  NListItem,
  NSkeleton,
  NTag,
  NText,
} from 'naive-ui';

import type { Post } from '@/api';

defineProps<{
  posts: Post[];
  loading: boolean;
}>();

const emit = defineEmits<{ viewAll: [] }>();

function formatDate(value: string) {
  return new Intl.DateTimeFormat('zh-CN', {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  }).format(new Date(value));
}
</script>

<template>
  <n-card class="activity-card" :bordered="false">
    <template #header>
      <div class="section-heading">
        <div class="section-title">
          <n-text strong>最近活跃</n-text>
          <n-text depth="3" class="section-caption">按最后活跃时间排序</n-text>
        </div>
        <n-button text type="primary" @click="emit('viewAll')">
          查看全部
        </n-button>
      </div>
    </template>

    <div v-if="loading" class="skeleton-list">
      <n-skeleton v-for="index in 4" :key="index" text :repeat="2" />
    </div>
    <n-list v-else-if="posts.length" :show-divider="false">
      <n-list-item v-for="post in posts" :key="post.id">
        <div class="post-row">
          <div class="post-main">
            <n-text strong class="post-title">{{ post.title }}</n-text>
            <n-text depth="3" class="post-meta">
              {{ post.authorUsername }} · {{ formatDate(post.activeAt) }}
            </n-text>
          </div>
          <div class="post-stats">
            <n-tag v-if="post.pinOrder != null" size="small" type="warning">
              置顶 {{ post.pinOrder }}
            </n-tag>
            <n-text depth="3">{{ post.commentsCount }} 评论</n-text>
          </div>
        </div>
      </n-list-item>
    </n-list>
    <n-text v-else depth="3">暂无已发布帖子</n-text>
  </n-card>
</template>

<style scoped>
.activity-card {
  box-shadow: 0 8px 28px rgba(0, 0, 0, 0.04);
}

.section-heading,
.post-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.section-title,
.post-main {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 3px;
}

.section-caption,
.post-meta {
  font-size: 12px;
}

.post-title {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.post-stats {
  display: flex;
  align-items: center;
  flex: none;
  gap: 10px;
}

.skeleton-list {
  display: grid;
  gap: 22px;
}

@media (max-width: 640px) {
  .post-stats {
    align-items: flex-end;
    flex-direction: column;
    gap: 4px;
  }
}
</style>
