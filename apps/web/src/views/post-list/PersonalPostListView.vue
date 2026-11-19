<script setup lang="ts">
import { ArticleOutlined, StarBorderOutlined } from '@vicons/material';
import { useWebKit } from '@novelia/web-kit';
import { computed } from 'vue';
import { useRoute, useRouter } from 'vue-router';

import { usePostListQuery } from '@/stores/post';
import PostList from './PostList.vue';

const props = defineProps<{ kind: 'mine' | 'favorites' }>();
const PAGE_SIZE = 20;
const presentations = {
  mine: {
    routeName: 'my-posts',
    label: '我的帖子',
    icon: ArticleOutlined,
    signInTitle: '登录后查看帖子',
    emptyTitle: '还没有发表帖子',
    emptyDescription: '发表帖子后，它们会出现在这里。',
  },
  favorites: {
    routeName: 'favorites',
    label: '我的收藏',
    icon: StarBorderOutlined,
    signInTitle: '登录后查看收藏',
    emptyTitle: '暂无收藏',
    emptyDescription: '收藏感兴趣的帖子后，它们会出现在这里。',
  },
};
const presentation = computed(() => presentations[props.kind]);
const route = useRoute();
const router = useRouter();
const { whoami } = useWebKit();

const page = computed(() => {
  const value = Number(route.query.page);
  return Number.isInteger(value) && value > 0 ? value : 1;
});

const { posts, total, loading, error, retry } = usePostListQuery(
  () => props.kind,
  () => ({ page: page.value, pageSize: PAGE_SIZE }),
);
const totalPages = computed(() =>
  Math.max(1, Math.ceil(total.value / PAGE_SIZE)),
);

function changePage(nextPage: number) {
  void router.push({
    name: presentation.value.routeName,
    query: nextPage > 1 ? { page: String(nextPage) } : {},
  });
  window.scrollTo({ top: 0, behavior: 'smooth' });
}
</script>

<template>
  <section :aria-label="presentation.label">
    <section v-if="!whoami.isSignedIn" class="py-8 text-center">
      <div
        class="mx-auto grid size-12 place-items-center rounded-full bg-primary/10 text-primary"
        aria-hidden="true"
      >
        <component :is="presentation.icon" class="size-6" />
      </div>
      <h2 class="mt-4 text-lg font-semibold text-default">
        {{ presentation.signInTitle }}
      </h2>
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
      :empty-title="presentation.emptyTitle"
      :empty-description="presentation.emptyDescription"
      @retry="retry"
      @change-page="changePage"
    />
  </section>
</template>
