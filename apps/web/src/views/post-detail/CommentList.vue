<script setup lang="ts">
import { ChatBubbleOutlineOutlined } from '@vicons/material';

import type { PostComment } from '@/api';
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

    <div
      v-if="loading"
      class="divide-y divide-divider"
      aria-label="正在加载评论"
    >
      <div v-for="index in 3" :key="index" class="px-4 py-5 sm:px-6">
        <div class="h-3 w-28 animate-pulse rounded-sm bg-border" />
        <div class="mt-3 h-4 w-full animate-pulse rounded-sm bg-divider" />
        <div class="mt-2 h-4 w-2/3 animate-pulse rounded-sm bg-divider" />
      </div>
    </div>

    <div
      v-else-if="error"
      class="grid min-h-52 place-items-center p-6 text-center"
    >
      <div>
        <p class="text-sm text-muted">{{ error }}</p>
        <button
          type="button"
          class="mt-4 rounded-sm bg-primary px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-primary-hover"
          @click="emit('retry')"
        >
          重新加载评论
        </button>
      </div>
    </div>

    <div
      v-else-if="!comments.length"
      class="grid min-h-52 place-items-center p-6 text-center"
    >
      <div>
        <div
          class="mx-auto grid size-11 place-items-center rounded-full bg-paper text-muted"
          aria-hidden="true"
        >
          <ChatBubbleOutlineOutlined class="size-5" />
        </div>
        <p class="mt-3 text-sm font-medium text-ink">还没有评论</p>
        <p class="mt-1 text-xs text-muted">这里暂时安安静静的。</p>
      </div>
    </div>

    <div v-else class="divide-y divide-divider">
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

    <PaginationControls
      v-if="!loading && !error && totalPages > 1"
      :page="page"
      :total-pages="totalPages"
      @change="emit('changePage', $event)"
    />
  </section>
</template>
