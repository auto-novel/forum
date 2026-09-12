<script setup lang="ts">
import {
  ChatBubbleOutlineOutlined,
  VisibilityOutlined,
} from '@vicons/material';
import { RouterLink } from 'vue-router';

import type { Post } from '@/api';
import PostTagList from '@/components/PostTagList.vue';

defineProps<{
  post: Post;
}>();

function excerpt(content: string) {
  return content.replace(/\s+/g, ' ').trim().slice(0, 180);
}

function formatDate(value: string) {
  const date = new Date(value);
  const now = new Date();
  const sameYear = date.getFullYear() === now.getFullYear();
  return new Intl.DateTimeFormat('zh-CN', {
    year: sameYear ? undefined : 'numeric',
    month: 'short',
    day: 'numeric',
  }).format(date);
}
</script>

<template>
  <article
    class="group relative px-4 py-4 transition-colors duration-300 hover:bg-paper sm:px-5"
  >
    <PostTagList
      class="mb-2"
      :tags="post.tags"
      :pinned="post.pinOrder != null"
    />

    <h2 class="text-[17px] leading-snug font-semibold">
      <RouterLink
        :to="{ name: 'post-detail', params: { id: post.id } }"
        class="text-ink transition-colors duration-300 before:absolute before:inset-0 focus-visible:rounded-sm focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-primary group-hover:text-primary"
      >
        {{ post.title }}
      </RouterLink>
    </h2>
    <p class="mt-1.5 line-clamp-2 text-sm leading-6 text-muted">
      {{ excerpt(post.content) || '这篇帖子暂时没有摘要。' }}
    </p>

    <div
      class="mt-3 flex flex-wrap items-center justify-between gap-3 text-xs text-muted"
    >
      <div class="flex items-center gap-2">
        <span class="font-medium text-ink/80">{{ post.authorUsername }}</span>
        <span aria-hidden="true">·</span>
        <time :datetime="post.activeAt">{{ formatDate(post.activeAt) }}</time>
      </div>

      <div class="flex items-center gap-4" aria-label="帖子数据">
        <span class="inline-flex items-center gap-1.5">
          <VisibilityOutlined class="size-4" aria-hidden="true" />
          {{ post.viewsCount }}
        </span>
        <span class="inline-flex items-center gap-1.5">
          <ChatBubbleOutlineOutlined class="size-4" aria-hidden="true" />
          {{ post.commentsCount }}
        </span>
      </div>
    </div>
  </article>
</template>
