<script setup lang="ts">
import {
  ChatBubbleOutlineOutlined,
  DeleteOutlineOutlined,
  SearchOutlined,
  VisibilityOffOutlined,
} from '@vicons/material';
import {
  NAlert,
  NButton,
  NCard,
  NEmpty,
  NIcon,
  NInputNumber,
  NModal,
  NPagination,
  NSkeleton,
  NTag,
  NText,
} from 'naive-ui';
import { computed, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';

import { useForumApi, type Comment } from '@/api';

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
const saving = ref(false);
const errorMessage = ref('');
const successMessage = ref('');
const pendingAction = ref<{
  comment: Comment;
  status: 'hidden' | 'deleted';
}>();

const pageCount = computed(() =>
  Math.max(1, Math.ceil(total.value / PAGE_SIZE)),
);

function formatDate(value: string) {
  return new Intl.DateTimeFormat('zh-CN', {
    dateStyle: 'medium',
    timeStyle: 'short',
  }).format(new Date(value));
}

async function loadComments() {
  if (activePostId.value == null) return;
  loading.value = true;
  errorMessage.value = '';
  try {
    const commentResult = await api.getComments(
      activePostId.value,
      page.value,
      PAGE_SIZE,
    );
    postLoaded.value = true;
    comments.value = commentResult.items;
    total.value = commentResult.total;
  } catch (error) {
    postLoaded.value = false;
    comments.value = [];
    total.value = 0;
    errorMessage.value = error instanceof Error ? error.message : String(error);
  } finally {
    loading.value = false;
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

async function confirmModeration() {
  if (!pendingAction.value) return;
  saving.value = true;
  errorMessage.value = '';
  try {
    await api.setCommentStatus(
      pendingAction.value.comment.id,
      pendingAction.value.status,
    );
    successMessage.value =
      pendingAction.value.status === 'hidden' ? '评论已隐藏' : '评论已删除';
    pendingAction.value = undefined;
    await loadComments();
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : String(error);
  } finally {
    saving.value = false;
  }
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
  <main class="comments-page">
    <section class="page-intro">
      <div>
        <n-text tag="h1" class="page-title">评论审核</n-text>
        <n-text depth="3">输入帖子 ID，查看讨论并处理违规内容。</n-text>
      </div>
    </section>

    <n-card class="search-card" :bordered="false">
      <div class="search-row">
        <span class="search-icon">
          <n-icon :component="ChatBubbleOutlineOutlined" />
        </span>
        <div class="search-copy">
          <n-text strong>定位帖子评论区</n-text>
          <n-text depth="3">评论接口按帖子组织，请先输入目标帖子 ID。</n-text>
        </div>
        <n-input-number
          v-model:value="postIdInput"
          :min="1"
          :show-button="false"
          placeholder="帖子 ID"
          @keyup.enter="searchPost"
        />
        <n-button
          type="primary"
          :disabled="postIdInput == null"
          @click="searchPost"
        >
          <template #icon><n-icon :component="SearchOutlined" /></template>
          查询
        </n-button>
      </div>
    </n-card>

    <n-alert
      v-if="successMessage"
      type="success"
      closable
      @close="successMessage = ''"
    >
      {{ successMessage }}
    </n-alert>
    <n-alert v-if="errorMessage" type="error" title="评论加载失败">
      {{ errorMessage }}
    </n-alert>

    <n-card v-if="postLoaded" size="small" class="post-context">
      <div>
        <n-text depth="3">正在审核</n-text>
        <n-text strong>帖子 #{{ activePostId }}</n-text>
      </div>
      <n-tag type="success" size="small">{{ total }} 条评论</n-tag>
    </n-card>

    <div v-if="loading" class="comment-list">
      <n-card v-for="index in 4" :key="index">
        <n-skeleton text :repeat="3" />
      </n-card>
    </div>
    <n-empty
      v-else-if="activePostId == null"
      class="empty-state"
      description="输入帖子 ID 后开始审核"
    />
    <n-empty
      v-else-if="!comments.length && !errorMessage"
      class="empty-state"
      description="该帖子暂无公开评论"
    />
    <div v-else class="comment-list">
      <n-card
        v-for="comment in comments"
        :key="comment.id"
        class="comment-card"
      >
        <div class="comment-header">
          <div>
            <n-text strong>{{ comment.authorUsername }}</n-text>
            <n-text depth="3" class="comment-meta">
              #{{ comment.id }} · {{ formatDate(comment.createdAt) }}
            </n-text>
          </div>
          <n-tag v-if="comment.rootId" size="small" :bordered="false">
            回复 #{{ comment.rootId }}
          </n-tag>
        </div>
        <n-text class="comment-content">{{ comment.content }}</n-text>
        <div class="comment-actions">
          <n-button
            size="small"
            secondary
            type="warning"
            @click="pendingAction = { comment, status: 'hidden' }"
          >
            <template #icon>
              <n-icon :component="VisibilityOffOutlined" />
            </template>
            隐藏
          </n-button>
          <n-button
            size="small"
            secondary
            type="error"
            @click="pendingAction = { comment, status: 'deleted' }"
          >
            <template #icon>
              <n-icon :component="DeleteOutlineOutlined" />
            </template>
            删除
          </n-button>
        </div>
      </n-card>
    </div>

    <div v-if="total > PAGE_SIZE" class="pagination">
      <n-pagination
        v-model:page="page"
        :page-count="pageCount"
        @update:page="loadComments"
      />
    </div>

    <n-modal
      :show="pendingAction != null"
      preset="dialog"
      type="warning"
      :title="
        pendingAction?.status === 'hidden' ? '确认隐藏评论' : '确认删除评论'
      "
      positive-text="确认处理"
      negative-text="取消"
      :loading="saving"
      @positive-click="confirmModeration"
      @negative-click="pendingAction = undefined"
      @mask-click="pendingAction = undefined"
    >
      处理后，该评论将立即从公开讨论中移除。
    </n-modal>
  </main>
</template>

<style scoped>
.comments-page {
  max-width: 900px;
  margin-inline: auto;
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.page-title {
  display: block;
  margin: 0 0 5px;
  font-size: 23px;
  font-weight: 680;
}
.search-card {
  background: linear-gradient(
    135deg,
    rgba(32, 128, 240, 0.08),
    rgba(24, 160, 88, 0.05)
  );
}
.search-row {
  display: grid;
  grid-template-columns: auto minmax(180px, 1fr) 180px auto;
  align-items: center;
  gap: 14px;
}
.search-icon {
  width: 44px;
  height: 44px;
  border-radius: 12px;
  color: #2080f0;
  background: rgba(32, 128, 240, 0.12);
  display: grid;
  place-items: center;
  font-size: 22px;
}
.search-copy,
.post-context > div,
.comment-header > div {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 3px;
}
.post-context :deep(.n-card__content),
.comment-header,
.comment-actions {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}
.comment-list {
  display: grid;
  gap: 10px;
}
.comment-card :deep(.n-card__content) {
  display: flex;
  flex-direction: column;
  gap: 14px;
}
.comment-meta {
  font-size: 12px;
}
.comment-content {
  line-height: 1.72;
  white-space: pre-wrap;
  overflow-wrap: anywhere;
}
.comment-actions {
  justify-content: flex-end;
  padding-top: 2px;
}
.empty-state {
  padding: 72px 16px;
}
.pagination {
  display: flex;
  justify-content: center;
}
@media (max-width: 700px) {
  .search-row {
    grid-template-columns: auto minmax(0, 1fr);
  }
  .search-row :deep(.n-input-number) {
    grid-column: 1 / 2;
  }
  .search-row > :last-child {
    grid-column: 2 / 3;
  }
}
</style>
