<script setup lang="ts">
import {
  ArticleOutlined,
  CategoryOutlined,
  ChatBubbleOutlineOutlined,
  LocalOfferOutlined,
} from '@vicons/material';
import { NAlert, NButton, NText } from 'naive-ui';
import { computed, onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';

import { useForumApi, type Post } from '@/api';

import OverviewMetrics from './OverviewMetrics.vue';
import OverviewRecentPosts from './OverviewRecentPosts.vue';

const api = useForumApi();
const router = useRouter();
const loading = ref(true);
const errorMessage = ref('');
const categoryCount = ref(0);
const tagCount = ref(0);
const postCount = ref(0);
const recentPosts = ref<Post[]>([]);

const recentCommentCount = computed(() =>
  recentPosts.value.reduce((total, post) => total + post.commentsCount, 0),
);

const metrics = computed(() => [
  {
    label: '内容分类',
    value: categoryCount.value,
    icon: CategoryOutlined,
    tone: 'blue',
  },
  {
    label: '可用标签',
    value: tagCount.value,
    icon: LocalOfferOutlined,
    tone: 'green',
  },
  {
    label: '已发布帖子',
    value: postCount.value,
    icon: ArticleOutlined,
    tone: 'orange',
  },
  {
    label: '近期帖评论',
    value: recentCommentCount.value,
    icon: ChatBubbleOutlineOutlined,
    tone: 'purple',
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
    const tagGroups = await Promise.all(
      categories.map((category) => api.getTags(category.id)),
    );
    categoryCount.value = categories.length;
    tagCount.value = tagGroups.flat().filter((tag) => tag.isActive).length;
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
  <main class="overview-page">
    <n-alert v-if="errorMessage" type="error" title="概览加载失败">
      <div class="error-content">
        <n-text>{{ errorMessage }}</n-text>
        <n-button size="small" @click="loadOverview">重新加载</n-button>
      </div>
    </n-alert>

    <OverviewMetrics :metrics="metrics" :loading="loading" />
    <OverviewRecentPosts
      :posts="recentPosts"
      :loading="loading"
      @view-all="router.push('/posts')"
    />
  </main>
</template>

<style scoped>
.overview-page {
  max-width: 1000px;
  margin-inline: auto;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.error-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}
</style>
