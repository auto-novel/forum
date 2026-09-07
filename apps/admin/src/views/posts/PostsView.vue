<script setup lang="ts">
import { NAlert, NButton, NSpace, NText } from 'naive-ui';
import { computed, onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';

import { useForumApi, type Category, type Post } from '@/api';

import PostFilters from './PostFilters.vue';
import PostList from './PostList.vue';
import PostModerationModal from './PostModerationModal.vue';

const PAGE_SIZE = 20;
const api = useForumApi();
const router = useRouter();
const categories = ref<Category[]>([]);
const posts = ref<Post[]>([]);
const loading = ref(true);
const moderationSaving = ref(false);
const errorMessage = ref('');
const successMessage = ref('');
const page = ref(1);
const total = ref(0);
const queryInput = ref('');
const categoryInput = ref('');
const query = ref('');
const category = ref('');
const selectedPost = ref<Post | null>(null);
let requestId = 0;

const categoryMap = computed(
  () => new Map(categories.value.map((item) => [item.id, item.slug])),
);
const hasFilters = computed(() => Boolean(query.value || category.value));

async function loadPosts() {
  const currentRequestId = ++requestId;
  loading.value = true;
  errorMessage.value = '';
  try {
    const result = await api.getPosts({
      page: page.value,
      pageSize: PAGE_SIZE,
      query: query.value,
      category: category.value,
    });
    if (currentRequestId !== requestId) return;
    posts.value = result.items;
    total.value = result.total;
  } catch (error) {
    if (currentRequestId !== requestId) return;
    posts.value = [];
    total.value = 0;
    errorMessage.value = error instanceof Error ? error.message : String(error);
  } finally {
    if (currentRequestId === requestId) loading.value = false;
  }
}

function search() {
  query.value = queryInput.value.trim();
  category.value = categoryInput.value;
  page.value = 1;
  void loadPosts();
}

function resetFilters() {
  queryInput.value = '';
  categoryInput.value = '';
  search();
}

function changePage(nextPage: number) {
  page.value = nextPage;
  void loadPosts();
}

function reviewComments(post: Post) {
  void router.push({ name: 'comments', query: { post: post.id } });
}

async function handleModerationSuccess(message: string) {
  selectedPost.value = null;
  successMessage.value = message;
  await loadPosts();
}

async function initialize() {
  try {
    categories.value = await api.getCategories();
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : String(error);
  }
  await loadPosts();
}

onMounted(initialize);
</script>

<template>
  <n-space vertical :size="16" class="posts-page">
    <PostFilters
      v-model:query="queryInput"
      v-model:category="categoryInput"
      :categories="categories"
      @search="search"
    />

    <n-alert
      v-if="successMessage"
      type="success"
      closable
      @close="successMessage = ''"
    >
      {{ successMessage }}
    </n-alert>
    <n-alert v-if="errorMessage" type="error" title="帖子列表加载失败">
      <n-space vertical size="small">
        <n-text>{{ errorMessage }}</n-text>
        <n-button size="small" @click="loadPosts">重新加载</n-button>
      </n-space>
    </n-alert>

    <PostList
      v-if="!errorMessage"
      :posts="posts"
      :loading="loading"
      :total="total"
      :page="page"
      :page-size="PAGE_SIZE"
      :has-filters="hasFilters"
      :category-names="categoryMap"
      @update-page="changePage"
      @reset-filters="resetFilters"
      @moderate="selectedPost = $event"
      @review-comments="reviewComments"
    />

    <PostModerationModal
      v-model:saving="moderationSaving"
      :post="selectedPost"
      @close="selectedPost = null"
      @success="handleModerationSuccess"
      @error="errorMessage = $event"
    />
  </n-space>
</template>

<style scoped>
.posts-page {
  max-width: 1000px;
  margin-inline: auto;
}
</style>
