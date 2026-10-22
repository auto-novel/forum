<script setup lang="ts">
import {
  ArticleOutlined,
  CategoryOutlined,
  ChatBubbleOutlineOutlined,
  LocalOfferOutlined,
} from '@vicons/material';
import { XButton } from '@novelia/web-kit';
import { computed, onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';

import { useForumApi, type PostSummary } from '@/api';
import AdminAlert from '@/components/AdminAlert.vue';

import OverviewMetrics from './OverviewMetrics.vue';
import OverviewRecentPosts from './OverviewRecentPosts.vue';

const api = useForumApi();
const router = useRouter();
const loading = ref(true);
const errorMessage = ref('');
const categoryCount = ref(0);
const tagCount = ref(0);
const postCount = ref(0);
const recentPosts = ref<PostSummary[]>([]);

const recentCommentCount = computed(() =>
  recentPosts.value.reduce((total, post) => total + post.commentsCount, 0),
);

const metrics = computed(() => [
  {
    label: '内容分类',
    value: categoryCount.value,
    icon: CategoryOutlined,
    tone: 'info',
  },
  {
    label: '可用标签',
    value: tagCount.value,
    icon: LocalOfferOutlined,
    tone: 'success',
  },
  {
    label: '已发布帖子',
    value: postCount.value,
    icon: ArticleOutlined,
    tone: 'warning',
  },
  {
    label: '近期帖评论',
    value: recentCommentCount.value,
    icon: ChatBubbleOutlineOutlined,
    tone: 'primary',
  },
]);

async function loadOverview() {
  loading.value = true;
  errorMessage.value = '';
  try {
    const [categories, posts] = await Promise.all([
      api.getCategories(),
      api.getPosts({ page: 1, pageSize: 6 }),
    ]);
    categoryCount.value = categories.length;
    tagCount.value = categories.reduce(
      (count, category) => count + category.tags.length,
      0,
    );
    postCount.value = posts.total;
    recentPosts.value = posts.items;
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : String(error);
  } finally {
    loading.value = false;
  }
}

onMounted(loadOverview);
</script>

<template>
  <main class="mx-auto flex max-w-[1000px] flex-col gap-5 text-ink">
    <AdminAlert v-if="errorMessage" type="error" title="概览加载失败">
      <div class="flex flex-wrap items-center justify-between gap-4">
        <span class="min-w-0 break-words">{{ errorMessage }}</span>
        <XButton
          size="sm"
          color="neutral"
          variant="outline"
          @click="loadOverview"
        >
          重新加载
        </XButton>
      </div>
    </AdminAlert>

    <OverviewMetrics :metrics="metrics" :loading="loading" />
    <OverviewRecentPosts
      :posts="recentPosts"
      :loading="loading"
      @view-all="router.push('/posts')"
    />
  </main>
</template>
