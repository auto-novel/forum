<script setup lang="ts">
import { ChatBubbleOutlineOutlined } from '@vicons/material';

import type { PostComment } from '@/api';
import AsyncContent from '@/components/AsyncContent.vue';
import PaginationControls from '@/components/PaginationControls.vue';

import CommentListItem from './CommentListItem.vue';

defineProps<{
  comments: PostComment[];
  loading: boolean;
  error?: string;
  page: number;
  total: number;
  totalPages: number;
  locked: boolean;
}>();

const emit = defineEmits<{
  retry: [];
  changePage: [page: number];
  reply: [comment: PostComment];
  updated: [comment: PostComment];
  statusChanged: [id: number, status: number];
}>();
</script>

<template>
  <section class="mt-5 rounded-sm bg-surface" aria-live="polite">
    <header class="border-b border-divider px-4 py-4 sm:px-6">
      <h2 class="font-semibold text-ink">
        评论
        <span class="text-muted">{{ total }}</span>
      </h2>
    </header>

    <AsyncContent
      :loading="loading"
      :error="error"
      :empty="!comments.length"
      size="compact"
      error-title="评论加载失败"
      retry-label="重新加载评论"
      empty-title="还没有评论"
      empty-description="这里暂时安安静静的。"
      @retry="emit('retry')"
    >
      <template #loading>
        <div class="divide-y divide-divider" aria-label="正在加载评论">
          <div v-for="index in 3" :key="index" class="px-4 py-5 sm:px-6">
            <div class="h-3 w-28 animate-pulse rounded-sm bg-border" />
            <div class="mt-3 h-4 w-full animate-pulse rounded-sm bg-divider" />
            <div class="mt-2 h-4 w-2/3 animate-pulse rounded-sm bg-divider" />
          </div>
        </div>
      </template>

      <template #empty-icon>
        <div
          class="mx-auto grid size-11 place-items-center rounded-full bg-paper text-muted"
          aria-hidden="true"
        >
          <ChatBubbleOutlineOutlined class="size-5" />
        </div>
      </template>

      <div class="divide-y divide-divider">
        <CommentListItem
          v-for="comment in comments"
          :key="comment.id"
          :comment="comment"
          :locked="locked"
          @reply="emit('reply', $event)"
          @updated="emit('updated', $event)"
          @status-changed="(id, status) => emit('statusChanged', id, status)"
        />
      </div>
    </AsyncContent>

    <PaginationControls
      v-if="!loading && !error && totalPages > 1"
      :page="page"
      :total-pages="totalPages"
      @change="emit('changePage', $event)"
    />
  </section>
</template>
