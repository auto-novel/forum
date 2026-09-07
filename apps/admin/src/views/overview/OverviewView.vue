<script setup lang="ts">
import {
  ArticleOutlined,
  CategoryOutlined,
  ChatBubbleOutlineOutlined,
  LocalOfferOutlined,
} from '@vicons/material';
import {
  NAlert,
  NButton,
  NCard,
  NIcon,
  NList,
  NListItem,
  NSkeleton,
  NTag,
  NText,
} from 'naive-ui';
import { computed, onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';

import { useForumApi, type Post } from '@/api';

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

function formatDate(value: string) {
  return new Intl.DateTimeFormat('zh-CN', {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  }).format(new Date(value));
}

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
    <section class="hero">
      <div>
        <n-text depth="3" class="eyebrow">FORUM CONTROL CENTER</n-text>
        <n-text tag="h1" class="hero-title">社区运营，一目了然</n-text>
        <n-text depth="3" class="hero-copy">
          管理内容结构、处理帖子状态，并快速定位需要关注的讨论。
        </n-text>
      </div>
      <div class="hero-actions">
        <n-button secondary type="primary" @click="router.push('/categories')">
          管理分类
        </n-button>
        <n-button type="primary" @click="router.push('/posts')">
          查看帖子
        </n-button>
      </div>
    </section>

    <n-alert v-if="errorMessage" type="error" title="概览加载失败">
      <div class="error-content">
        <n-text>{{ errorMessage }}</n-text>
        <n-button size="small" @click="loadOverview">重新加载</n-button>
      </div>
    </n-alert>

    <section class="metric-grid" aria-label="论坛数据概览">
      <n-card v-for="metric in metrics" :key="metric.label" class="metric-card">
        <n-skeleton v-if="loading" text :repeat="2" />
        <div v-else class="metric-content">
          <div>
            <n-text depth="3">{{ metric.label }}</n-text>
            <n-text class="metric-value">
              {{ metric.value.toLocaleString() }}
            </n-text>
          </div>
          <span :class="['metric-icon', metric.tone]">
            <n-icon :component="metric.icon" />
          </span>
        </div>
      </n-card>
    </section>

    <n-card class="activity-card" :bordered="false">
      <template #header>
        <div class="section-heading">
          <div>
            <n-text strong>最近活跃</n-text>
            <n-text depth="3" class="section-caption">
              按最后活跃时间排序
            </n-text>
          </div>
          <n-button text type="primary" @click="router.push('/posts')">
            查看全部
          </n-button>
        </div>
      </template>

      <div v-if="loading" class="skeleton-list">
        <n-skeleton v-for="index in 4" :key="index" text :repeat="2" />
      </div>
      <n-list v-else-if="recentPosts.length" :show-divider="false">
        <n-list-item v-for="post in recentPosts" :key="post.id">
          <div class="post-row">
            <div class="post-main">
              <n-text strong class="post-title">{{ post.title }}</n-text>
              <n-text depth="3" class="post-meta">
                {{ post.authorUsername }} · {{ formatDate(post.activeAt) }}
              </n-text>
            </div>
            <div class="post-stats">
              <n-tag v-if="post.pinOrder != null" size="small" type="warning">
                置顶 {{ post.pinOrder }}
              </n-tag>
              <n-text depth="3">{{ post.commentsCount }} 评论</n-text>
            </div>
          </div>
        </n-list-item>
      </n-list>
      <n-text v-else depth="3">暂无已发布帖子</n-text>
    </n-card>
  </main>
</template>

<style scoped>
.overview-page {
  max-width: 1120px;
  margin-inline: auto;
  display: flex;
  flex-direction: column;
  gap: 20px;
}
.hero {
  position: relative;
  overflow: hidden;
  min-height: 142px;
  padding: 28px 30px;
  border: 1px solid rgba(24, 160, 88, 0.18);
  border-radius: 16px;
  background:
    radial-gradient(circle at 88% 10%, rgba(24, 160, 88, 0.2), transparent 32%),
    linear-gradient(135deg, rgba(24, 160, 88, 0.1), rgba(59, 130, 246, 0.05));
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 24px;
}
.eyebrow {
  display: block;
  margin-bottom: 8px;
  color: #18a058;
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.16em;
}
.hero-title {
  display: block;
  margin: 0 0 8px;
  font-size: clamp(24px, 4vw, 34px);
  line-height: 1.2;
  font-weight: 700;
}
.hero-copy {
  display: block;
  max-width: 540px;
}
.hero-actions {
  display: flex;
  flex: none;
  gap: 10px;
}
.error-content,
.section-heading,
.post-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}
.metric-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 14px;
}
.metric-card :deep(.n-card__content) {
  min-height: 108px;
  padding: 18px;
}
.metric-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}
.metric-content > div {
  display: flex;
  flex-direction: column;
  gap: 7px;
}
.metric-value {
  font-size: 28px;
  line-height: 1;
  font-weight: 700;
  font-variant-numeric: tabular-nums;
}
.metric-icon {
  width: 42px;
  height: 42px;
  border-radius: 12px;
  display: grid;
  place-items: center;
  flex: none;
  font-size: 21px;
}
.metric-icon.blue {
  color: #2080f0;
  background: rgba(32, 128, 240, 0.12);
}
.metric-icon.green {
  color: #18a058;
  background: rgba(24, 160, 88, 0.12);
}
.metric-icon.orange {
  color: #f0a020;
  background: rgba(240, 160, 32, 0.14);
}
.metric-icon.purple {
  color: #8b5cf6;
  background: rgba(139, 92, 246, 0.12);
}
.activity-card {
  box-shadow: 0 8px 28px rgba(0, 0, 0, 0.04);
}
.section-heading > div,
.post-main {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 3px;
}
.section-caption,
.post-meta {
  font-size: 12px;
}
.post-title {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.post-stats {
  display: flex;
  align-items: center;
  flex: none;
  gap: 10px;
}
.skeleton-list {
  display: grid;
  gap: 22px;
}
@media (max-width: 900px) {
  .metric-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}
@media (max-width: 640px) {
  .hero {
    padding: 22px;
    align-items: flex-start;
    flex-direction: column;
  }
  .hero-actions {
    width: 100%;
  }
  .hero-actions :deep(.n-button) {
    flex: 1;
  }
  .metric-grid {
    gap: 10px;
  }
  .post-stats {
    align-items: flex-end;
    flex-direction: column;
    gap: 4px;
  }
}
</style>
