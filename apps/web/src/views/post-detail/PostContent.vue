<script setup lang="ts">
import type { Post } from '@/api';
import MarkdownContent from '@/components/markdown/MarkdownContent.vue';

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

function formatDate(value: string) {
  return new Intl.DateTimeFormat('zh-CN', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  }).format(new Date(value));
}
</script>

<template>
  <article class="rounded-sm bg-surface px-4 py-5 sm:px-6 sm:py-7">
    <div class="flex flex-wrap items-center gap-1.5 text-xs">
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

    <h1
      class="mt-3 text-2xl leading-tight font-bold tracking-tight text-ink sm:text-3xl"
    >
      {{ post.title }}
    </h1>

    <div
      class="mt-5 flex flex-wrap items-center justify-between gap-3 text-xs text-muted"
    >
      <div class="flex items-center gap-2">
        <span class="font-medium text-ink/80">{{ post.authorUsername }}</span>
        <span aria-hidden="true">·</span>
        <time :datetime="post.createdAt">{{ formatDate(post.createdAt) }}</time>
      </div>
      <div class="flex items-center gap-4" aria-label="帖子数据">
        <span>{{ post.viewsCount }} 次浏览</span>
        <span>{{ post.commentsCount }} 条评论</span>
      </div>
    </div>

    <div class="my-6 h-px bg-divider" />
    <MarkdownContent mode="article" :source="post.content" />
  </article>
</template>
