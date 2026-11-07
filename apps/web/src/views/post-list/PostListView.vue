<script setup lang="ts">
import {
  AddOutlined,
  CheckOutlined,
  ExpandMoreOutlined,
} from '@vicons/material';
import { computed } from 'vue';
import { RouterLink, useRoute, useRouter } from 'vue-router';

import { type PostSort } from '@/api';
import {
  XButton,
  XDropdownMenu,
  XDropdownMenuItem,
  XPage,
  XPageHeader,
  useWebKitViewport,
} from '@novelia/web-kit';
import { useCategoryStore } from '@/stores/category';
import { usePostListQuery } from '@/stores/post';
import PostFilters from './PostFilters.vue';
import PostList from './PostList.vue';

const PAGE_SIZE = 20;
const POST_SORTS = new Set<PostSort>([
  'active',
  'updated',
  'views',
  'comments',
]);

const route = useRoute();
const router = useRouter();
const { isMobile } = useWebKitViewport();
const categoryStore = useCategoryStore();

const selectedCategory = computed(() => {
  const value = route.params.slug;
  return (
    categoryStore.categories.find((category) => category.slug === value)
      ?.slug ?? categoryStore.defaultCategory.slug
  );
});

const selectedCategoryItem = computed(
  () =>
    categoryStore.categories.find(
      (category) => category.slug === selectedCategory.value,
    ) ?? categoryStore.defaultCategory,
);

const searchQuery = computed(() => {
  const value = route.query.q;
  return typeof value === 'string' ? value.trim() : '';
});

const selectedTagId = computed(() => {
  const value = Number(route.query.tag);
  return Number.isSafeInteger(value) && value > 0 ? value : undefined;
});

const selectedSort = computed<PostSort>(() => {
  const value = route.query.sort;
  return typeof value === 'string' && POST_SORTS.has(value as PostSort)
    ? (value as PostSort)
    : 'active';
});

const page = computed(() => {
  const value = Number(route.query.page);
  return Number.isInteger(value) && value > 0 ? value : 1;
});

const totalPages = computed(() =>
  Math.max(1, Math.ceil(total.value / PAGE_SIZE)),
);
const hasFilters = computed(() =>
  Boolean(searchQuery.value || selectedTagId.value),
);

const {
  posts,
  total,
  loading: postsLoading,
  error: postsError,
  retry: loadPosts,
} = usePostListQuery('category', () => ({
  page: page.value,
  pageSize: PAGE_SIZE,
  category: selectedCategory.value,
  query: searchQuery.value || undefined,
  tagIds: selectedTagId.value ? [selectedTagId.value] : undefined,
  sort: selectedSort.value,
}));

function applyFilters(filters: {
  category: string;
  query: string;
  tagId?: number;
  sort: PostSort;
}) {
  void router.push({
    name: 'posts',
    params: { slug: filters.category },
    query: {
      ...(filters.query ? { q: filters.query } : {}),
      ...(filters.tagId ? { tag: String(filters.tagId) } : {}),
      ...(filters.sort !== 'active' ? { sort: filters.sort } : {}),
    },
  });
}

function changeCategory(slug: string) {
  if (slug === selectedCategory.value) return;
  void router.push({
    name: 'posts',
    params: { slug },
    query: {
      ...(searchQuery.value ? { q: searchQuery.value } : {}),
      ...(selectedSort.value !== 'active' ? { sort: selectedSort.value } : {}),
    },
  });
}

function changePage(nextPage: number) {
  void router.push({
    name: 'posts',
    params: { slug: selectedCategory.value },
    query: {
      ...(searchQuery.value ? { q: searchQuery.value } : {}),
      ...(selectedTagId.value ? { tag: String(selectedTagId.value) } : {}),
      ...(selectedSort.value !== 'active' ? { sort: selectedSort.value } : {}),
      ...(nextPage > 1 ? { page: String(nextPage) } : {}),
    },
  });
  window.scrollTo({ top: 0, behavior: 'smooth' });
}
</script>

<template>
  <XPage>
    <div class="min-w-0 space-y-4">
      <XPageHeader
        :title="selectedCategoryItem.title"
        class="flex-nowrap items-center"
      >
        <template v-if="isMobile" #title>
          <XDropdownMenu align="start" class="min-w-40">
            <template #trigger>
              <button
                type="button"
                class="inline-flex min-h-11 max-w-full items-center gap-1 rounded-sm text-left focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary/40"
                aria-label="切换帖子板块"
              >
                <span class="truncate">{{ selectedCategoryItem.title }}</span>
                <ExpandMoreOutlined
                  class="size-6 shrink-0"
                  aria-hidden="true"
                />
              </button>
            </template>
            <XDropdownMenuItem
              v-for="category in categoryStore.categories"
              :key="category.slug"
              :aria-current="
                selectedCategory === category.slug ? 'page' : undefined
              "
              @select="changeCategory(category.slug)"
            >
              <span class="flex-1">{{ category.title }}</span>
              <CheckOutlined
                v-if="selectedCategory === category.slug"
                class="size-4 shrink-0"
                aria-hidden="true"
              />
            </XDropdownMenuItem>
          </XDropdownMenu>
        </template>
        <template #actions>
          <XButton
            v-if="categoryStore.canPublish(selectedCategory)"
            class="shrink-0 whitespace-nowrap"
            :as="RouterLink"
            :to="{
              name: 'post-create',
              query: { category: selectedCategory },
            }"
            aria-label="发帖"
          >
            <AddOutlined class="size-4" aria-hidden="true" />
            <span>发帖</span>
          </XButton>
        </template>
      </XPageHeader>

      <aside
        v-if="selectedCategory === 'feedback'"
        aria-label="意见反馈告示"
        class="space-y-2 text-sm leading-relaxed text-ink"
      >
        <p class="text-warning">
          FishHawk陷入加班地狱，网站开发速度大幅下降已成常态，论坛反馈目前没有精力维护，有问题加群@吧，群链接看小说站首页。
        </p>
      </aside>
      <PostFilters
        :category-id="selectedCategoryItem.id"
        :query="searchQuery"
        :tag-id="selectedTagId"
        :sort="selectedSort"
        @apply="applyFilters"
      />
      <PostList
        :posts="posts"
        :loading="postsLoading"
        :error="postsError"
        :page="page"
        :total-pages="totalPages"
        :empty-title="hasFilters ? '没有找到帖子' : undefined"
        :empty-description="
          hasFilters ? '可以尝试调整关键词或标签条件。' : undefined
        "
        @retry="loadPosts"
        @change-page="changePage"
      />
    </div>
  </XPage>
</template>
