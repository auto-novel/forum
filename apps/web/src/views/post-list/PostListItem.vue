<script setup lang="ts">
import { RouterLink } from 'vue-router';

import type { Post } from '@/api';

defineProps<{
  post: Post;
  categoryName: string;
}>();

const tagColors = [
  'bg-primary-soft text-primary',
  'bg-blue-50 text-blue-600',
  'bg-orange-50 text-orange-600',
  'bg-purple-50 text-purple-600',
  'bg-red-50 text-red-600',
];

function tagClass(color: number) {
  return tagColors[Math.abs(color) % tagColors.length];
}

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
    <div class="mb-2 flex flex-wrap items-center gap-1.5 text-xs">
      <span class="font-medium text-primary">{{ categoryName }}</span>
      <span
        v-if="post.pinOrder != null"
        class="rounded-sm bg-orange-50 px-2 py-0.5 font-medium text-orange-600"
      >
        置顶
      </span>
      <span
        v-for="tag in post.tags"
        :key="tag.id"
        class="rounded-sm px-2 py-0.5 font-medium"
        :class="tagClass(tag.color)"
      >
        {{ tag.name }}
      </span>
    </div>

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
        <span
          class="grid size-7 place-items-center rounded-full bg-primary-soft font-semibold text-primary"
          aria-hidden="true"
        >
          {{ post.authorUsername.slice(0, 1).toUpperCase() }}
        </span>
        <span class="font-medium text-ink/80">{{ post.authorUsername }}</span>
        <span aria-hidden="true">·</span>
        <time :datetime="post.activeAt">{{ formatDate(post.activeAt) }}</time>
      </div>

      <div class="flex items-center gap-4" aria-label="帖子数据">
        <span class="inline-flex items-center gap-1.5">
          <svg
            viewBox="0 0 24 24"
            class="size-4"
            fill="none"
            aria-hidden="true"
          >
            <path
              d="M3.5 12s3.2-5 8.5-5 8.5 5 8.5 5-3.2 5-8.5 5-8.5-5-8.5-5Z"
              stroke="currentColor"
              stroke-width="1.6"
            />
            <circle
              cx="12"
              cy="12"
              r="2"
              stroke="currentColor"
              stroke-width="1.6"
            />
          </svg>
          {{ post.viewsCount }}
        </span>
        <span class="inline-flex items-center gap-1.5">
          <svg
            viewBox="0 0 24 24"
            class="size-4"
            fill="none"
            aria-hidden="true"
          >
            <path
              d="M5 6.5h14v9H11l-4.5 3v-3H5v-9Z"
              stroke="currentColor"
              stroke-width="1.6"
              stroke-linejoin="round"
            />
          </svg>
          {{ post.commentsCount }}
        </span>
      </div>
    </div>
  </article>
</template>
