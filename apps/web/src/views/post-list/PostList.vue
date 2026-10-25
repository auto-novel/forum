<script setup lang="ts">
import { onMounted, useTemplateRef, watch } from 'vue';
import { useRoute } from 'vue-router';
import type { PostSummary } from '@/api';
import {
  restoreScrollPosition,
  XAsyncContent,
  XPagination,
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

const paginationScope = useTemplateRef<HTMLElement>('paginationScope');

function focusList() {
  const element = paginationScope.value;
  if (!element) return;
  const document = element.ownerDocument;
  const activeElement = document.activeElement;
  if (
    activeElement &&
    activeElement !== document.body &&
    activeElement !== document.documentElement
  )
    return;
  element.focus({ preventScroll: true });
}

onMounted(focusList);

const emit = defineEmits<{
  retry: [];
  changePage: [page: number];
}>();
</script>

<template>
  <section
    ref="paginationScope"
    tabindex="0"
    aria-label="帖子列表"
    aria-live="polite"
  >
    <XAsyncContent
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
          <div v-for="index in 5" :key="index" class="py-2.5">
            <div class="h-5 w-3/4 skeleton bg-border" />
            <div class="mt-1.5 h-5 w-1/3 skeleton" />
          </div>
        </div>
      </template>

      <div
        class="hidden grid-cols-[minmax(0,1fr)_6rem_6rem] items-center gap-x-4 border-b border-divider pb-2 text-xs font-bold text-ink sm:grid"
        aria-hidden="true"
      >
        <span>帖子</span>
        <span class="text-center">查看</span>
        <span class="text-center">评论</span>
      </div>
      <div class="divide-y divide-divider">
        <PostListItem v-for="post in posts" :key="post.id" :post="post" />
      </div>
    </XAsyncContent>

    <XPagination
      v-if="!loading && !error && totalPages > 1"
      :page="page"
      :total-pages="totalPages"
      keyboard
      :keyboard-scope="paginationScope"
      @update:page="emit('changePage', $event)"
    />
  </section>
</template>
