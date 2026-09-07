<script setup lang="ts">
import {
  NCard,
  NEmpty,
  NList,
  NListItem,
  NPagination,
  NSkeleton,
  NSpin,
  NTag,
  NText,
} from 'naive-ui';
import { computed } from 'vue';

import type { Comment } from '@/api';

import CommentListItem from './CommentListItem.vue';

const props = defineProps<{
  comments: Comment[];
  loading: boolean;
  total: number;
  page: number;
  pageSize: number;
  postId?: number;
  loaded: boolean;
}>();

const emit = defineEmits<{
  updatePage: [page: number];
  moderate: [comment: Comment, status: 'hidden' | 'deleted'];
}>();

const pageCount = computed(() =>
  Math.max(1, Math.ceil(props.total / props.pageSize)),
);

function forwardModeration(comment: Comment, status: 'hidden' | 'deleted') {
  emit('moderate', comment, status);
}
</script>

<template>
  <div class="comment-list">
    <div v-if="loaded" class="list-summary">
      <n-text depth="3">帖子 #{{ postId }}</n-text>
      <n-tag size="small" type="success" :bordered="false">
        {{ total }} 条评论
      </n-tag>
    </div>

    <n-card class="list-card" content-style="padding: 0;">
      <div v-if="loading && !comments.length" class="skeleton-list">
        <div v-for="index in 5" :key="index" class="skeleton-row">
          <n-skeleton text :repeat="3" />
        </div>
      </div>
      <n-spin v-else :show="loading">
        <n-list v-if="comments.length">
          <n-list-item v-for="comment in comments" :key="comment.id">
            <CommentListItem :comment="comment" @moderate="forwardModeration" />
          </n-list-item>
        </n-list>
        <n-empty
          v-else
          class="empty"
          :description="
            loaded ? '该帖子暂无公开评论' : '输入帖子 ID 后开始审核'
          "
        />
      </n-spin>
    </n-card>

    <div v-if="total > pageSize" class="pagination">
      <n-pagination
        :page="page"
        :page-count="pageCount"
        :page-slot="5"
        @update:page="emit('updatePage', $event)"
      />
    </div>
  </div>
</template>

<style scoped>
.comment-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.list-summary {
  min-height: 28px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.list-card {
  border-right: 0;
  border-left: 0;
}

.skeleton-list {
  padding: 0 16px;
}

.skeleton-row {
  padding: 18px 0;
}

.skeleton-row + .skeleton-row {
  border-top: 1px solid var(--n-border-color);
}

:deep(.n-list-item) {
  padding: 14px 16px;
}

.empty {
  padding: 64px 16px;
}

.pagination {
  display: flex;
  overflow-x: auto;
}

@media (max-width: 767px) {
  :deep(.n-list-item) {
    padding: 12px 10px;
  }

  .skeleton-list {
    padding-inline: 10px;
  }
}
</style>
