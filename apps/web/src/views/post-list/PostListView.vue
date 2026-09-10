<script setup lang="ts">
import { AddOutlined } from '@vicons/material';
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue';
import { RouterLink, useRoute, useRouter } from 'vue-router';

import { CATEGORIES, getPosts, type Post, type PostSort } from '@/api';
import PostFilters from './PostFilters.vue';
import PostList from './PostList.vue';

const PAGE_SIZE = 20;
const POST_SORTS = new Set<PostSort>(['active', 'newest', 'views', 'comments']);

const route = useRoute();
const router = useRouter();

const categories = CATEGORIES;
const posts = ref<Post[]>([]);
const total = ref(0);
const postsLoading = ref(true);
const postsError = ref('');
let postsController: AbortController | undefined;

const selectedCategory = computed(() => {
  const value = route.query.category;
  return (
    categories.find((category) => category.slug === value)?.slug ??
    categories[0].slug
  );
});

const selectedCategoryItem = computed(
  () =>
    categories.find((category) => category.slug === selectedCategory.value) ??
    categories[0],
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
  postsController?.abort();
  const controller = new AbortController();
  postsController = controller;
  postsLoading.value = true;
  postsError.value = '';
  try {
    const result = await getPosts(
      {
        page: page.value,
        pageSize: PAGE_SIZE,
        category: selectedCategory.value,
        query: searchQuery.value || undefined,
        tagIds: selectedTagId.value ? [selectedTagId.value] : undefined,
        sort: selectedSort.value,
      },
      controller.signal,
    );
    posts.value = result.items;
    total.value = result.total;
  } catch (error) {
    if (error instanceof DOMException && error.name === 'AbortError') return;
    postsError.value = error instanceof Error ? error.message : '无法加载帖子';
    posts.value = [];
    total.value = 0;
  } finally {
    if (postsController === controller) postsLoading.value = false;
  }
}

function applyFilters(filters: {
  query: string;
  tagId?: number;
  sort: PostSort;
}) {
  void router.push({
    name: 'posts',
    query: {
      category: selectedCategory.value,
      ...(filters.query ? { q: filters.query } : {}),
      ...(filters.tagId ? { tag: String(filters.tagId) } : {}),
      ...(filters.sort !== 'active' ? { sort: filters.sort } : {}),
    },
  });
}

function changePage(nextPage: number) {
  void router.push({
    name: 'posts',
    query: {
      category: selectedCategory.value,
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
);

onMounted(() => {
  void loadPosts();
});

onBeforeUnmount(() => {
  postsController?.abort();
});
</script>

<template>
  <div class="page-container py-4 md:py-6">
    <div class="min-w-0 space-y-4">
      <div class="flex justify-end">
        <RouterLink
          :to="{
            name: 'post-create',
            query: { category: selectedCategory },
          }"
          class="inline-flex min-h-10 items-center justify-center gap-1.5 rounded-sm bg-primary px-4 text-sm font-medium text-white transition-colors hover:bg-primary-hover focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary"
        >
          <AddOutlined class="size-4" aria-hidden="true" />
          发表帖子
        </RouterLink>
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
