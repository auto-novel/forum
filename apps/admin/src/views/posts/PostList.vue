<script setup lang="ts">
import { XButton, XPagination } from '@novelia/web-kit';
import { computed } from 'vue';

import AdminSpinner from '@/components/AdminSpinner.vue';
import type { PostStatus, PostSummary } from '@/api';

import PostListItem from './PostListItem.vue';

const props = defineProps<{
  posts: PostSummary[];
  loading: boolean;
  total: number;
  page: number;
  pageSize: number;
  hasFilters: boolean;
  categoryNames: Map<number, string>;
  busyPostId: number | null;
}>();

const emit = defineEmits<{
  updatePage: [page: number];
  resetFilters: [];
  setStatus: [post: PostSummary, status: PostStatus];
  setCommentsLocked: [post: PostSummary, locked: boolean];
  setPinOrder: [post: PostSummary, pinOrder: number | null];
  reviewComments: [post: PostSummary];
}>();

const pageCount = computed(() =>
  Math.max(1, Math.ceil(props.total / props.pageSize)),
);
const rangeStart = computed(() =>
  props.total === 0 ? 0 : (props.page - 1) * props.pageSize + 1,
);
const rangeEnd = computed(() =>
  Math.min(props.page * props.pageSize, props.total),
);
</script>

<template>
  <div class="flex flex-col gap-3">
    <p
      v-if="!loading || posts.length"
      class="flex min-h-7 items-center text-muted"
    >
      <template v-if="total">
        显示 {{ rangeStart }}–{{ rangeEnd }}，共 {{ total }} 篇帖子
      </template>
      <template v-else>共 0 篇帖子</template>
    </p>

    <section class="border-y border-border bg-surface text-ink">
      <div
        v-if="loading && !posts.length"
        class="divide-y divide-border px-2.5 md:px-4"
      >
        <div v-for="index in 5" :key="index" class="py-4.5">
          <div aria-hidden="true" class="grid gap-2">
            <div class="skeleton h-3.5"></div>
            <div class="skeleton h-3.5"></div>
            <div class="skeleton h-3.5 w-3/5"></div>
          </div>
        </div>
      </div>
      <div v-else class="relative" :aria-busy="loading || undefined">
        <div
          v-if="loading"
          class="absolute inset-0 z-10 cursor-progress bg-surface/60"
          aria-hidden="true"
        ></div>
        <AdminSpinner
          v-if="loading"
          label="正在加载帖子"
          class="absolute top-4 right-4 z-20"
        />
        <ul v-if="posts.length" class="divide-y divide-divider">
          <li
            v-for="post in posts"
            :key="post.id"
            class="px-2.5 py-3 md:px-4 md:py-3.5"
          >
            <PostListItem
              :post="post"
              :category-name="
                categoryNames.get(post.categoryId) ?? `分类 ${post.categoryId}`
              "
              :actions-disabled="busyPostId !== null || loading"
              :saving="busyPostId === post.id"
              @set-status="(item, status) => emit('setStatus', item, status)"
              @set-comments-locked="
                (item, locked) => emit('setCommentsLocked', item, locked)
              "
              @set-pin-order="
                (item, pinOrder) => emit('setPinOrder', item, pinOrder)
              "
              @review-comments="emit('reviewComments', $event)"
            />
          </li>
        </ul>
        <div
          v-else
          class="flex flex-col items-center gap-4 px-4 py-16 text-muted"
        >
          <p>{{ hasFilters ? '没有符合条件的帖子' : '暂无帖子' }}</p>
          <XButton
            v-if="hasFilters"
            color="neutral"
            variant="outline"
            size="sm"
            @click="emit('resetFilters')"
          >
            清除筛选
          </XButton>
        </div>
      </div>
    </section>

    <XPagination
      v-if="total > pageSize"
      :page="page"
      :total-pages="pageCount"
      class="border-t-0 px-0 sm:px-0"
      @update:page="emit('updatePage', $event)"
    />
  </div>
</template>
