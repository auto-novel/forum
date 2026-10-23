<script setup lang="ts">
import { XButton, XPagination } from '@novelia/web-kit';
import { computed } from 'vue';

import AdminSpinner from '@/components/AdminSpinner.vue';
import type { Comment, CommentStatusName } from '@/api';

import CommentListItem from './CommentListItem.vue';

const props = defineProps<{
  comments: Comment[];
  loading: boolean;
  total: number;
  page: number;
  pageSize: number;
  postId?: number;
  hasFilters: boolean;
  actionsDisabled: boolean;
}>();

const emit = defineEmits<{
  updatePage: [page: number];
  filterPost: [postId: number];
  resetFilters: [];
  moderate: [comment: Comment, status: CommentStatusName];
  deleteAuthor: [comment: Comment];
}>();

const pageCount = computed(() =>
  Math.max(1, Math.ceil(props.total / props.pageSize)),
);

function forwardModeration(comment: Comment, status: CommentStatusName) {
  emit('moderate', comment, status);
}
</script>

<template>
  <div class="flex flex-col gap-3">
    <div class="flex min-h-7 items-center justify-between gap-3">
      <span class="text-muted">
        {{ postId ? `帖子 #${postId}` : '全部帖子' }} · 最新评论优先
      </span>
      <span
        class="shrink-0 rounded bg-success/10 px-2 py-0.5 text-xs text-success"
      >
        {{ total }} 条评论
      </span>
    </div>

    <section class="rounded border-y border-border bg-surface">
      <div v-if="loading && !comments.length" class="px-2.5 md:px-4">
        <div
          v-for="index in 5"
          :key="index"
          class="border-t border-border py-[18px] first:border-t-0"
        >
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
          label="正在加载评论"
          class="absolute top-4 right-4 z-20"
        />
        <ul
          v-if="comments.length"
          class="m-0 list-none divide-y divide-border p-0"
        >
          <li
            v-for="comment in comments"
            :key="comment.id"
            class="px-2.5 py-3 md:px-4 md:py-3.5"
          >
            <CommentListItem
              :comment="comment"
              :actions-disabled="actionsDisabled"
              @moderate="forwardModeration"
              @delete-author="emit('deleteAuthor', $event)"
              @filter-post="emit('filterPost', $event)"
            />
          </li>
        </ul>
        <div
          v-else
          class="flex flex-col items-center gap-3 px-4 py-16 text-sm text-muted"
        >
          <p>{{ hasFilters ? '没有符合筛选条件的评论' : '暂无评论' }}</p>
          <XButton
            v-if="hasFilters"
            size="sm"
            color="neutral"
            variant="outline"
            @click="emit('resetFilters')"
          >
            重置筛选
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
