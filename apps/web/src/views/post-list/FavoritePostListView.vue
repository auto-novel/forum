<script setup lang="ts">
import { StarBorderOutlined } from '@vicons/material';
import { storeToRefs } from 'pinia';
import { computed, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';

import { authUser } from '@/api';
import { usePostStore } from '@/stores/post';
import PostList from './PostList.vue';

const PAGE_SIZE = 20;

const route = useRoute();
const router = useRouter();
const postStore = usePostStore();
const {
  listPosts: posts,
  listTotal: total,
  listLoading: loading,
  listError: error,
} = storeToRefs(postStore);

const page = computed(() => {
  const value = Number(route.query.page);
  return Number.isInteger(value) && value > 0 ? value : 1;
});

const totalPages = computed(() =>
  Math.max(1, Math.ceil(total.value / PAGE_SIZE)),
);

async function loadFavorites() {
  if (!authUser.value) {
    postStore.clearList();
    return;
  }
  await postStore.loadFavoritePosts({
    page: page.value,
    pageSize: PAGE_SIZE,
  });
}

function changePage(nextPage: number) {
  void router.push({
    name: 'favorites',
    query: nextPage > 1 ? { page: String(nextPage) } : {},
  });
  document.querySelector('main')?.scrollTo({ top: 0, behavior: 'smooth' });
}

watch([authUser, page], loadFavorites, { immediate: true });
</script>

<template>
  <div class="page-container py-4 md:py-6">
    <div class="mx-auto max-w-4xl">
      <header class="mb-5">
        <h1 class="text-2xl font-bold tracking-tight text-ink">我的收藏</h1>
        <p class="mt-1 text-sm text-muted">查看你收藏过的帖子。</p>
      </header>

      <section v-if="!authUser" class="py-8 text-center">
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
