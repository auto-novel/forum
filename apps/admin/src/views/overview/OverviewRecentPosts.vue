<script setup lang="ts">
import { XButton } from '@novelia/web-kit';

import type { PostSummary } from '@/api';

defineProps<{
  posts: PostSummary[];
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
  <section
    class="rounded-lg bg-surface p-6 text-ink shadow-sm"
    aria-label="最近活跃"
    :aria-busy="loading"
  >
    <header class="mb-5 flex items-center justify-between gap-4">
      <div class="flex min-w-0 flex-col gap-[3px]">
        <h2 class="font-semibold">最近活跃</h2>
        <span class="text-xs text-muted">按最后活跃时间排序</span>
      </div>
      <XButton
        variant="ghost"
        size="sm"
        class="shrink-0"
        @click="emit('viewAll')"
      >
        查看全部
      </XButton>
    </header>

    <div v-if="loading" class="grid gap-[22px]">
      <div
        v-for="index in 4"
        :key="index"
        aria-hidden="true"
        class="grid gap-2"
      >
        <div class="skeleton h-3.5"></div>
        <div class="skeleton h-3.5 w-3/5"></div>
      </div>
    </div>
    <ul v-else-if="posts.length">
      <li v-for="post in posts" :key="post.id" class="py-3">
        <div class="flex items-center justify-between gap-4">
          <div class="flex min-w-0 flex-col gap-[3px]">
            <span class="truncate font-medium">{{ post.title }}</span>
            <span class="text-xs text-muted">
              {{ post.authorUsername }} · {{ formatDate(post.activeAt) }}
            </span>
          </div>
          <div
            class="flex shrink-0 items-center gap-2.5 max-[640px]:flex-col max-[640px]:items-end max-[640px]:gap-1"
          >
            <span
              v-if="post.pinOrder != null"
              class="rounded border border-warning/30 bg-warning-soft px-1.5 py-0.5 text-xs text-warning-text"
            >
              置顶 {{ post.pinOrder }}
            </span>
            <span class="text-muted">{{ post.commentsCount }} 评论</span>
          </div>
        </div>
      </li>
    </ul>
    <p v-else class="text-muted">暂无已发布帖子</p>
  </section>
</template>
