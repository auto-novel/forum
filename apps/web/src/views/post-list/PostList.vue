<script setup lang="ts">
import type { Post } from '@/api';
import AsyncContent from '@/components/AsyncContent.vue';
import PaginationControls from '@/components/PaginationControls.vue';

import PostListItem from './PostListItem.vue';

defineProps<{
  posts: Post[];
  loading: boolean;
  error?: string;
  page: number;
  totalPages: number;
  emptyTitle?: string;
  emptyDescription?: string;
}>();

const emit = defineEmits<{
  retry: [];
  changePage: [page: number];
}>();
</script>

<template>
  <section class="overflow-hidden rounded-sm bg-surface" aria-live="polite">
    <AsyncContent
      :loading="loading"
      :error="error"
      :empty="!posts.length"
      error-title="帖子列表加载失败"
      :empty-title="emptyTitle ?? '暂无帖子'"
      :empty-description="emptyDescription ?? '这个分类暂时没有帖子。'"
      @retry="emit('retry')"
    >
      <template #loading>
        <div class="divide-y divide-divider" aria-label="正在加载帖子">
          <div v-for="index in 5" :key="index" class="px-5 py-5">
            <div class="h-3 w-24 animate-pulse rounded-sm bg-divider" />
            <div class="mt-3 h-5 w-3/4 animate-pulse rounded-sm bg-border" />
            <div class="mt-3 h-4 w-full animate-pulse rounded-sm bg-divider" />
            <div class="mt-2 h-4 w-2/3 animate-pulse rounded-sm bg-divider" />
          </div>
        </div>
      </template>

      <div class="divide-y divide-divider">
        <PostListItem v-for="post in posts" :key="post.id" :post="post" />
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
