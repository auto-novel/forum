<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';

import { CATEGORIES, getPosts, type Post } from '@/api';
import CategorySidebar from './CategorySidebar.vue';
import PostList from './PostList.vue';

const PAGE_SIZE = 20;

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

const page = computed(() => {
  const value = Number(route.query.page);
  return Number.isInteger(value) && value > 0 ? value : 1;
});

const categoryNames = computed(
  () => new Map(categories.map((category) => [category.id, category.title])),
);
const totalPages = computed(() =>
  Math.max(1, Math.ceil(total.value / PAGE_SIZE)),
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

function selectCategory(slug: string) {
  void router.push({
    name: 'posts',
    query: { category: slug },
  });
}

function changePage(nextPage: number) {
  void router.push({
    name: 'posts',
    query: {
      category: selectedCategory.value,
      ...(nextPage > 1 ? { page: String(nextPage) } : {}),
    },
  });
  window.scrollTo({ top: 0, behavior: 'smooth' });
}

watch([selectedCategory, page], loadPosts);

onMounted(() => {
  void loadPosts();
});

onBeforeUnmount(() => {
  postsController?.abort();
});
</script>

<template>
  <div class="page-container py-4 md:py-6">
    <div
      class="grid gap-6 md:grid-cols-[4rem_minmax(0,1fr)] lg:grid-cols-[14rem_minmax(0,1fr)] lg:gap-8"
    >
      <CategorySidebar
        :categories="categories"
        :selected="selectedCategory"
        @select="selectCategory"
      />
      <PostList
        class="min-w-0"
        :posts="posts"
        :category-names="categoryNames"
        :loading="postsLoading"
        :error="postsError"
        :page="page"
        :total-pages="totalPages"
        @retry="loadPosts"
        @change-page="changePage"
      />
    </div>
  </div>
</template>
