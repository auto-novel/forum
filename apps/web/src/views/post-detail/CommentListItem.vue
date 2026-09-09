<script setup lang="ts">
import type { PostComment } from '@/api';

defineProps<{
  comment: PostComment;
}>();

function formatDate(value: string) {
  return new Intl.DateTimeFormat('zh-CN', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  }).format(new Date(value));
}
</script>

<template>
  <article
    class="px-4 py-5 sm:px-6"
    :class="
      comment.rootId != null
        ? 'ml-6 border-l-2 border-primary-soft sm:ml-12'
        : ''
    "
  >
    <header class="flex items-center gap-2 text-xs text-muted">
      <span
        class="grid size-8 place-items-center rounded-full bg-primary-soft font-semibold text-primary"
        aria-hidden="true"
      >
        {{ comment.authorUsername.slice(0, 1).toUpperCase() }}
      </span>
      <span class="font-medium text-ink">{{ comment.authorUsername }}</span>
      <span aria-hidden="true">·</span>
      <time :datetime="comment.createdAt">
        {{ formatDate(comment.createdAt) }}
      </time>
    </header>
    <p class="mt-3 whitespace-pre-wrap break-words text-sm leading-6 text-ink">
      {{ comment.content }}
    </p>
  </article>
</template>
