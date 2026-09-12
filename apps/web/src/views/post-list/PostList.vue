<script setup lang="ts">
import { Inventory2Outlined } from '@vicons/material';

import type { Post } from '@/api';
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
    <div
      v-if="loading"
      class="divide-y divide-divider"
      aria-label="正在加载帖子"
    >
      <div v-for="index in 5" :key="index" class="px-5 py-5">
        <div class="h-3 w-24 animate-pulse rounded-sm bg-divider" />
        <div class="mt-3 h-5 w-3/4 animate-pulse rounded-sm bg-border" />
        <div class="mt-3 h-4 w-full animate-pulse rounded-sm bg-divider" />
        <div class="mt-2 h-4 w-2/3 animate-pulse rounded-sm bg-divider" />
      </div>
    </div>

    <div
      v-else-if="error"
      class="grid min-h-80 place-items-center p-8 text-center"
    >
      <div>
        <div
          class="mx-auto grid size-12 place-items-center rounded-full bg-red-50 text-red-500"
        >
          !
        </div>
        <h2 class="mt-4 text-lg font-semibold">帖子列表加载失败</h2>
        <p class="mt-2 text-sm text-muted">{{ error }}</p>
        <button
          type="button"
          class="mt-5 rounded-sm bg-primary px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-primary-hover"
          @click="emit('retry')"
        >
          再试一次
        </button>
      </div>
    </div>

    <div
      v-else-if="!posts.length"
      class="grid min-h-80 place-items-center p-8 text-center"
    >
      <div>
        <div
          class="mx-auto grid size-12 place-items-center rounded-full bg-paper text-muted"
          aria-hidden="true"
        >
          <Inventory2Outlined class="size-6" />
        </div>
        <h2 class="mt-4 text-lg font-semibold">
          {{ emptyTitle ?? '暂无帖子' }}
        </h2>
        <p class="mt-2 text-sm text-muted">
          {{ emptyDescription ?? '这个分类暂时没有帖子。' }}
        </p>
      </div>
    </div>

    <div v-else class="divide-y divide-divider">
      <PostListItem v-for="post in posts" :key="post.id" :post="post" />
    </div>

    <PaginationControls
      v-if="!loading && !error && totalPages > 1"
      :page="page"
      :total-pages="totalPages"
      @change="emit('changePage', $event)"
    />
  </section>
</template>
