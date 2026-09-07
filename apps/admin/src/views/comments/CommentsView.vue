<script setup lang="ts">
import { NAlert, NButton, NSpace, NText } from 'naive-ui';
import { ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';

import { useForumApi, type Comment } from '@/api';

import CommentFilters from './CommentFilters.vue';
import CommentList from './CommentList.vue';
import CommentModerationModal, {
  type CommentModerationTarget,
} from './CommentModerationModal.vue';

const PAGE_SIZE = 20;
const api = useForumApi();
const route = useRoute();
const router = useRouter();
const postIdInput = ref<number | null>(null);
const activePostId = ref<number>();
const postLoaded = ref(false);
const comments = ref<Comment[]>([]);
const total = ref(0);
const page = ref(1);
const loading = ref(false);
const moderationSaving = ref(false);
const errorMessage = ref('');
const successMessage = ref('');
const pendingAction = ref<CommentModerationTarget | null>(null);
let requestId = 0;

async function loadComments() {
  if (activePostId.value == null) return;
  const currentRequestId = ++requestId;
  loading.value = true;
  errorMessage.value = '';
  try {
    const result = await api.getComments(
      activePostId.value,
      page.value,
      PAGE_SIZE,
    );
    if (currentRequestId !== requestId) return;
    postLoaded.value = true;
    comments.value = result.items;
    total.value = result.total;
  } catch (error) {
    if (currentRequestId !== requestId) return;
    postLoaded.value = false;
    comments.value = [];
    total.value = 0;
    errorMessage.value = error instanceof Error ? error.message : String(error);
  } finally {
    if (currentRequestId === requestId) loading.value = false;
  }
}

function searchPost() {
  if (postIdInput.value == null || postIdInput.value <= 0) return;
  if (String(route.query.post ?? '') === String(postIdInput.value)) {
    activePostId.value = postIdInput.value;
    page.value = 1;
    void loadComments();
  } else {
    void router.replace({ query: { post: postIdInput.value } });
  }
}

function changePage(nextPage: number) {
  page.value = nextPage;
  void loadComments();
}

function requestModeration(comment: Comment, status: 'hidden' | 'deleted') {
  pendingAction.value = { comment, status };
}

async function handleModerationSuccess(message: string) {
  pendingAction.value = null;
  successMessage.value = message;
  await loadComments();
}

watch(
  () => route.query.post,
  (value) => {
    const normalizedValue = Array.isArray(value) ? value[0] : value;
    const postId = Number(normalizedValue);
    if (!Number.isSafeInteger(postId) || postId <= 0) return;
    postIdInput.value = postId;
    activePostId.value = postId;
    page.value = 1;
    void loadComments();
  },
  { immediate: true },
);
</script>

<template>
  <n-space vertical :size="16" class="comments-page">
    <CommentFilters v-model:post-id="postIdInput" @search="searchPost" />

    <n-alert
      v-if="successMessage"
      type="success"
      closable
      @close="successMessage = ''"
    >
      {{ successMessage }}
    </n-alert>
    <n-alert v-if="errorMessage" type="error" title="评论加载失败">
      <n-space vertical size="small">
        <n-text>{{ errorMessage }}</n-text>
        <n-button size="small" @click="loadComments">重新加载</n-button>
      </n-space>
    </n-alert>

    <CommentList
      v-if="!errorMessage"
      :comments="comments"
      :loading="loading"
      :total="total"
      :page="page"
      :page-size="PAGE_SIZE"
      :post-id="activePostId"
      :loaded="postLoaded"
      @update-page="changePage"
      @moderate="requestModeration"
    />

    <CommentModerationModal
      v-model:saving="moderationSaving"
      :target="pendingAction"
      @close="pendingAction = null"
      @success="handleModerationSuccess"
      @error="errorMessage = $event"
    />
  </n-space>
</template>

<style scoped>
.comments-page {
  max-width: 1000px;
  margin-inline: auto;
}
</style>
