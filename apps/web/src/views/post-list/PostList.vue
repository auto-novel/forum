<script setup lang="ts">
import { computed, onMounted, watch } from 'vue';
import { useRoute } from 'vue-router';
import type { PostSummary } from '@/api';
import {
  restoreScrollPosition,
  XButton,
  XError,
  XEmpty,
  XSkeleton,
  XPagination,
  XSeparator,
} from '@novelia/web-kit';

import PostListItem from './PostListItem.vue';

const props = defineProps<{
  posts: PostSummary[];
  loading: boolean;
  error?: string;
  page: number;
  totalPages: number;
  emptyTitle?: string;
  emptyDescription?: string;
}>();

// 新页面请求期间查询会暂时清空总数，保留已知页数以免卸载分页。
const paginationTotalPages = computed<number>((previous) =>
  props.loading ? (previous ?? props.totalPages) : props.totalPages,
);

const route = useRoute();

async function restoreScroll() {
  if (props.loading || props.error) return;
  const path = route.fullPath;
  const { postListScrollTop: scrollTop, position: historyPosition } =
    window.history.state ?? {};
  if (typeof scrollTop !== 'number' || !Number.isFinite(scrollTop)) return;
  const position = await restoreScrollPosition(
    { top: scrollTop, left: 0 },
    () =>
      route.fullPath === path &&
      window.history.state?.position === historyPosition &&
      window.history.state?.postListScrollTop === scrollTop &&
      !props.loading &&
      !props.error,
  );
  if (!position) return;
  window.scrollTo(position);
  const state = { ...window.history.state };
  delete state.postListScrollTop;
  window.history.replaceState(state, '');
}

onMounted(restoreScroll);
watch(() => props.loading, restoreScroll, { flush: 'post' });

const emit = defineEmits<{
  retry: [];
  changePage: [page: number];
}>();
</script>

<template>
  <section aria-label="帖子列表" aria-live="polite">
    <XPagination
      v-if="!error && paginationTotalPages > 1"
      :page="page"
      :total-pages="paginationTotalPages"
      keyboard
      class="mt-6"
      :ui="{ root: 'border-t-0 p-0 sm:px-0' }"
      aria-label="帖子列表顶部分页"
      @update:page="emit('changePage', $event)"
    />

    <XSeparator
      :style="{
        margin: !error && paginationTotalPages > 1 ? '12px 0 0' : '24px 0 0',
      }"
    />

    <div v-if="loading" class="divide-y divide-divider" role="status">
      <span class="sr-only">正在加载帖子…</span>
      <div v-for="index in 5" :key="index" class="py-2.5">
        <XSkeleton class="h-5 w-3/4 bg-border" />
        <XSkeleton class="mt-1.5 h-5 w-1/3" />
      </div>
    </div>
    <XError v-else-if="error" title="帖子列表加载失败" :description="error">
      <XButton @click="emit('retry')">再试一次</XButton>
    </XError>
    <XEmpty
      v-else-if="!posts.length"
      :title="emptyTitle ?? '暂无帖子'"
      :description="emptyDescription ?? '这个分类暂时没有帖子。'"
    />
    <template v-else>
      <div
        class="hidden grid-cols-[minmax(0,1fr)_6rem_6rem] items-center gap-x-4 border-b border-divider py-2 text-xs font-bold text-ink sm:grid"
        aria-hidden="true"
      >
        <span>帖子</span>
        <span class="text-center">查看</span>
        <span class="text-center">评论</span>
      </div>
      <div class="divide-y divide-divider">
        <PostListItem v-for="post in posts" :key="post.id" :post="post" />
      </div>
    </template>

    <XPagination
      v-if="!error && paginationTotalPages > 1"
      :page="page"
      :total-pages="paginationTotalPages"
      keyboard="window"
      aria-label="帖子列表底部分页"
      @update:page="emit('changePage', $event)"
    />
  </section>
</template>
