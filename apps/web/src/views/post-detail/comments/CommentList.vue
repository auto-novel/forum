<script setup lang="ts">
import { ChatBubbleOutlineOutlined } from '@vicons/material';

import { XPagination } from '@novelia/web-kit';

import type { PostComment } from '@/api';
import { formatCount } from '@/utils/formatCount';

import CommentThread from './CommentThread.vue';

defineProps<{
  comments: PostComment[];
  highlighted?: boolean;
  highlightVersion?: number;
  loading: boolean;
  error?: string;
  page: number;
  commentsCount: number;
  totalPages: number;
  locked: boolean;
  postId: number;
  composing: boolean;
  replyToId?: number;
}>();

const emit = defineEmits<{
  retry: [];
  comment: [];
  changePage: [page: number];
  reply: [comment: PostComment];
  cancelReply: [];
  created: [comment: PostComment];
  statusChanged: [id: number, status: number];
  authorCommentsDeleted: [];
}>();
</script>

<template>
  <section
    tabindex="0"
    aria-label="评论列表"
    class="mt-5 border-t border-divider"
    aria-live="polite"
  >
    <header
      :class="[
        'flex items-center justify-between gap-4 py-4',
        { 'border-b border-divider': !composing },
      ]"
    >
      <h2
        :key="highlightVersion"
        class="comments-title text-lg font-semibold text-ink"
        :class="{ 'comments-title-highlighted': highlighted }"
      >
        评论
        <span class="text-muted">{{ formatCount(commentsCount) }}</span>
      </h2>
      <UButton
        size="lg"
        aria-controls="comment-composer"
        :aria-expanded="composing"
        @click="emit('comment')"
      >
        <ChatBubbleOutlineOutlined class="size-4" aria-hidden="true" />
        {{ composing ? '取消评论' : '发表评论' }}
      </UButton>
    </header>

    <slot name="composer" />

    <div v-if="loading" class="divide-y divide-divider" role="status">
      <span class="sr-only">正在加载评论…</span>
      <div v-for="index in 3" :key="index" class="py-4">
        <USkeleton class="h-3 w-28" />
        <USkeleton class="mt-3 h-4 w-full" />
        <USkeleton class="mt-2 h-4 w-2/3" />
      </div>
    </div>
    <UError
      :ui="{
        root: 'min-h-0 py-8',
        statusMessage: 'text-lg font-medium sm:text-xl',
        message: 'mt-2 text-sm',
        links: 'mt-5 flex-wrap gap-3',
      }"
      v-else-if="error"
      :error="{ statusMessage: '', message: error }"
      class="min-h-52 p-6"
      :clear="false"
      :as="{ root: 'div', statusMessage: 'h2' }"
      role="alert"
    >
      <template #links>
        <UButton size="lg" @click="emit('retry')">重新加载评论</UButton>
      </template>
    </UError>
    <UEmpty
      variant="naked"
      v-else-if="!comments.length"
      size="sm"
      title="还没有评论"
      description="这里暂时安安静静的。"
    >
      <template #leading>
        <span
          class="grid size-11 place-items-center rounded-full bg-primary-soft text-primary"
          aria-hidden="true"
        >
          <ChatBubbleOutlineOutlined class="size-5" />
        </span>
      </template>
    </UEmpty>
    <div v-else class="divide-y divide-divider">
      <CommentThread
        v-for="comment in comments"
        :key="comment.id"
        :comment="comment"
        :locked="locked"
        :post-id="postId"
        :reply-to-id="replyToId"
        @reply="emit('reply', $event)"
        @cancel-reply="emit('cancelReply')"
        @created="emit('created', $event)"
        @status-changed="(id, status) => emit('statusChanged', id, status)"
        @author-comments-deleted="emit('authorCommentsDeleted')"
      />
    </div>

    <XPagination
      v-if="!loading && !error && totalPages > 1"
      :page="page"
      :total-pages="totalPages"
      keyboard="window"
      @update:page="emit('changePage', $event)"
    />
  </section>
</template>

<style scoped>
.comments-title,
.comments-title span {
  transition: color 350ms ease-out;
}

.comments-title-highlighted {
  color: var(--color-warning);
  animation: comments-title-blink 600ms ease-in-out 2;
}

.comments-title-highlighted span {
  color: var(--color-warning);
}

@keyframes comments-title-blink {
  0%,
  100% {
    opacity: 1;
  }
  50% {
    opacity: 0.4;
  }
}

@media (prefers-reduced-motion: reduce) {
  .comments-title,
  .comments-title span {
    transition: none;
  }

  .comments-title-highlighted {
    animation: none;
  }
}
</style>
