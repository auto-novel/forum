<script setup lang="ts">
import { AddOutlined } from '@vicons/material';
import { storeToRefs } from 'pinia';
import { computed, watch } from 'vue';
import { RouterLink, useRoute, useRouter } from 'vue-router';

import { type PostSort } from '@/api';
import AppButton from '@/components/AppButton.vue';
import { useCategoryStore } from '@/stores/category';
import { usePostStore } from '@/stores/post';
import PostFilters from './PostFilters.vue';
import PostList from './PostList.vue';

const PAGE_SIZE = 20;
const POST_SORTS = new Set<PostSort>(['active', 'newest', 'views', 'comments']);

const route = useRoute();
const router = useRouter();
const categoryStore = useCategoryStore();
const postStore = usePostStore();
const {
  listPosts: posts,
  listTotal: total,
  listLoading: postsLoading,
  listError: postsError,
} = storeToRefs(postStore);

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

async function loadPosts() {
  await postStore.loadCategoryPosts({
    page: page.value,
    pageSize: PAGE_SIZE,
    category: selectedCategory.value,
    query: searchQuery.value || undefined,
    tagIds: selectedTagId.value ? [selectedTagId.value] : undefined,
    sort: selectedSort.value,
  });
}

function applyFilters(filters: {
  query: string;
  tagId?: number;
  sort: PostSort;
}) {
  void router.push({
    name: 'posts',
    params: { slug: selectedCategory.value },
    query: {
      ...(filters.query ? { q: filters.query } : {}),
      ...(filters.tagId ? { tag: String(filters.tagId) } : {}),
      ...(filters.sort !== 'active' ? { sort: filters.sort } : {}),
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
  document.querySelector('main')?.scrollTo({ top: 0, behavior: 'smooth' });
}

watch(
  [selectedCategory, searchQuery, selectedTagId, selectedSort, page],
  loadPosts,
  { immediate: true },
);
</script>

<template>
  <div class="page-container py-4 md:py-6">
    <div class="min-w-0 space-y-4">
      <div class="flex justify-end">
        <AppButton
          :as="RouterLink"
          :to="{
            name: 'post-create',
            query: { category: selectedCategory },
          }"
        >
          <AddOutlined class="size-4" aria-hidden="true" />
          发表帖子
        </AppButton>
      </div>
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
  </div>
</template>
