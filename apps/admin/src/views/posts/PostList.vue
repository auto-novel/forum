<script setup lang="ts">
import { XButton } from '@novelia/web-kit';
import { NPagination, NSpin } from 'naive-ui';
import { computed } from 'vue';

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
  <div class="post-list">
    <p v-if="!loading || posts.length" class="list-summary text-muted">
      <template v-if="total">
        显示 {{ rangeStart }}–{{ rangeEnd }}，共 {{ total }} 篇帖子
      </template>
      <template v-else>共 0 篇帖子</template>
    </p>

    <section class="border-y border-border bg-surface text-ink">
      <div v-if="loading && !posts.length" class="skeleton-list">
        <div v-for="index in 5" :key="index" class="skeleton-row">
          <div aria-hidden="true" class="grid gap-2">
            <div class="skeleton h-3.5"></div>
            <div class="skeleton h-3.5"></div>
            <div class="skeleton h-3.5 w-3/5"></div>
          </div>
        </div>
      </div>
      <n-spin v-else :show="loading">
        <ul v-if="posts.length" class="divide-y divide-divider">
          <li v-for="post in posts" :key="post.id" class="list-item">
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
        <div v-else class="empty flex flex-col items-center gap-4 text-muted">
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
      </n-spin>
    </section>

    <div v-if="total > pageSize" class="pagination">
      <n-pagination
        :page="page"
        :page-count="pageCount"
        :page-slot="5"
        show-quick-jumper
        @update:page="emit('updatePage', $event)"
      >
        <template #goto>跳至</template>
      </n-pagination>
    </div>
  </div>
</template>

<style scoped>
.post-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.list-summary {
  min-height: 28px;
  display: flex;
  align-items: center;
}

.skeleton-list {
  padding: 0 16px;
}

.skeleton-row {
  padding: 18px 0;
}

.skeleton-row + .skeleton-row {
  border-top: 1px solid var(--color-border);
}

.list-item {
  padding: 14px 16px;
}

.empty {
  padding: 64px 16px;
}

.pagination {
  display: flex;
  overflow-x: auto;
}

@media (max-width: 767px) {
  .list-item {
    padding: 12px 10px;
  }

  .skeleton-list {
    padding-inline: 10px;
  }
}
</style>
