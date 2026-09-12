<script setup lang="ts">
import { ArrowBackOutlined, StarBorderOutlined } from '@vicons/material';
import { computed, onBeforeUnmount, ref, watch } from 'vue';
import { RouterLink, useRoute, useRouter } from 'vue-router';

import { authUser, getFavoritePosts, type Post } from '@/api';
import { useCategoryStore } from '@/stores/category';
import PostList from './PostList.vue';

const PAGE_SIZE = 20;

const route = useRoute();
const router = useRouter();
const categoryStore = useCategoryStore();

const posts = ref<Post[]>([]);
const total = ref(0);
const loading = ref(false);
const error = ref('');
let postsController: AbortController | undefined;

const page = computed(() => {
  const value = Number(route.query.page);
  return Number.isInteger(value) && value > 0 ? value : 1;
});

const totalPages = computed(() =>
  Math.max(1, Math.ceil(total.value / PAGE_SIZE)),
);

async function loadFavorites() {
  postsController?.abort();
  postsController = undefined;
  error.value = '';

  if (!authUser.value) {
    posts.value = [];
    total.value = 0;
    loading.value = false;
    return;
  }

  const controller = new AbortController();
  postsController = controller;
  loading.value = true;

  try {
    const result = await getFavoritePosts(
      { page: page.value, pageSize: PAGE_SIZE },
      controller.signal,
    );
    posts.value = result.items;
    total.value = result.total;
  } catch (reason) {
    if (reason instanceof DOMException && reason.name === 'AbortError') return;
    error.value = reason instanceof Error ? reason.message : '无法加载收藏帖子';
    posts.value = [];
    total.value = 0;
  } finally {
    if (postsController === controller) loading.value = false;
  }
}

function changePage(nextPage: number) {
  void router.push({
    name: 'favorites',
    query: nextPage > 1 ? { page: String(nextPage) } : {},
  });
  document.querySelector('main')?.scrollTo({ top: 0, behavior: 'smooth' });
}

watch([authUser, page], loadFavorites, { immediate: true });

onBeforeUnmount(() => postsController?.abort());
</script>

<template>
  <div class="page-container py-4 md:py-6">
    <div class="mx-auto max-w-4xl">
      <RouterLink
        :to="{
          name: 'posts',
          params: { slug: categoryStore.defaultCategory.slug },
        }"
        class="mb-4 inline-flex items-center gap-1.5 rounded-sm text-sm font-medium text-muted transition-colors hover:text-primary focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-primary"
      >
        <ArrowBackOutlined class="size-4" aria-hidden="true" />
        返回帖子列表
      </RouterLink>

      <header class="mb-5">
        <h1 class="text-2xl font-bold tracking-tight text-ink">我的收藏</h1>
        <p class="mt-1 text-sm text-muted">查看你收藏过的帖子。</p>
      </header>

      <section
        v-if="!authUser"
        class="rounded-sm border border-divider bg-surface px-5 py-8 text-center"
      >
        <div
          class="mx-auto grid size-12 place-items-center rounded-full bg-primary-soft text-primary"
          aria-hidden="true"
        >
          <StarBorderOutlined class="size-6" />
        </div>
        <h2 class="mt-4 text-lg font-semibold text-ink">登录后查看收藏</h2>
        <p class="mt-2 text-sm text-muted">
          请使用页面右上角的登录入口登录账号。
        </p>
      </section>

      <PostList
        v-else
        :posts="posts"
        :loading="loading"
        :error="error"
        :page="page"
        :total-pages="totalPages"
        empty-title="暂无收藏"
        empty-description="收藏感兴趣的帖子后，它们会出现在这里。"
        @retry="loadFavorites"
        @change-page="changePage"
      />
    </div>
  </div>
</template>
