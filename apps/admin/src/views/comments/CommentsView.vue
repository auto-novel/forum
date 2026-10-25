<script setup lang="ts">
import { XButton } from '@novelia/web-kit';
import AdminAlert from '@/components/AdminAlert.vue';
import { computed, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';

import {
  useForumApi,
  type Comment,
  type CommentStatusFilter,
  type CommentSubjectFilter,
  type CommentStatusName,
} from '@/api';

import CommentFilters from './CommentFilters.vue';
import CommentList from './CommentList.vue';
import CommentModerationModal, {
  type CommentModerationTarget,
} from './CommentModerationModal.vue';
import CommentAuthorDeleteModal from './CommentAuthorDeleteModal.vue';

const PAGE_SIZE = 20;
const api = useForumApi();
const route = useRoute();
const router = useRouter();
const subjectTypeInput = ref<CommentSubjectFilter>('');
const subjectKeyInput = ref('');
const activeSubjectType = ref<CommentSubjectFilter>('');
const activeSubjectKey = ref('');
const queryInput = ref('');
const authorNameInput = ref('');
const statusInput = ref<CommentStatusFilter | ''>('');
const activeQuery = ref('');
const activeAuthorName = ref('');
const activeStatus = ref<CommentStatusFilter | ''>('');
const hasFilters = computed(() =>
  Boolean(
    activeSubjectType.value ||
    activeSubjectKey.value ||
    activeQuery.value ||
    activeAuthorName.value ||
    activeStatus.value,
  ),
);
const comments = ref<Comment[]>([]);
const total = ref(0);
const page = ref(1);
const loading = ref(false);
const moderationSaving = ref(false);
const authorDeletionSaving = ref(false);
const errorMessage = ref('');
const successMessage = ref('');
const actionErrorMessage = ref('');
const pendingAction = ref<CommentModerationTarget | null>(null);
const pendingAuthorDeletion = ref<Comment | null>(null);
const actionsDisabled = computed(
  () => moderationSaving.value || authorDeletionSaving.value,
);
let requestId = 0;

async function loadComments() {
  const currentRequestId = ++requestId;
  loading.value = true;
  errorMessage.value = '';
  try {
    const result = await api.getComments({
      subjectType: activeSubjectType.value,
      subjectKey: activeSubjectKey.value,
      page: page.value,
      pageSize: PAGE_SIZE,
      query: activeQuery.value,
      authorName: activeAuthorName.value,
      status: activeStatus.value || undefined,
    });
    if (currentRequestId !== requestId) return;
    if (result.total > 0 && page.value > Math.ceil(result.total / PAGE_SIZE)) {
      page.value = Math.ceil(result.total / PAGE_SIZE);
      await loadComments();
      return;
    }
    comments.value = result.items;
    total.value = result.total;
  } catch (error) {
    if (currentRequestId !== requestId) return;
    comments.value = [];
    total.value = 0;
    errorMessage.value = error instanceof Error ? error.message : String(error);
  } finally {
    if (currentRequestId === requestId) loading.value = false;
  }
}

function search() {
  const query = {
    q: queryInput.value.trim() || undefined,
    author: authorNameInput.value.trim() || undefined,
    subject_type: subjectTypeInput.value || undefined,
    subject_key: subjectKeyInput.value || undefined,
    status: statusInput.value || undefined,
  };
  if (
    !route.query.post &&
    ['q', 'author', 'subject_type', 'subject_key', 'status'].every(
      (key) => route.query[key] === query[key as keyof typeof query],
    )
  ) {
    page.value = 1;
    void loadComments();
  } else {
    void router.replace({ query });
  }
}

function resetFilters() {
  queryInput.value = '';
  authorNameInput.value = '';
  statusInput.value = '';
  subjectTypeInput.value = '';
  subjectKeyInput.value = '';
  search();
}

function filterSubject(comment: Comment) {
  subjectTypeInput.value = String(comment.subjectType) as CommentSubjectFilter;
  subjectKeyInput.value = comment.subjectKey;
  search();
}

function changePage(nextPage: number) {
  page.value = nextPage;
  void loadComments();
}

function requestModeration(comment: Comment, status: CommentStatusName) {
  actionErrorMessage.value = '';
  successMessage.value = '';
  pendingAction.value = { comment, status };
}

async function handleModerationSuccess(message: string) {
  pendingAction.value = null;
  successMessage.value = message;
  await loadComments();
}

function requestAuthorDeletion(comment: Comment) {
  actionErrorMessage.value = '';
  successMessage.value = '';
  pendingAuthorDeletion.value = comment;
}

async function handleAuthorDeletionSuccess(message: string) {
  pendingAuthorDeletion.value = null;
  successMessage.value = message;
  await loadComments();
}

watch(
  () => route.query,
  (query) => {
    const text = (value: unknown) => (typeof value === 'string' ? value : '');
    const postId = Number(text(query.post));
    const legacyPost = Number.isSafeInteger(postId) && postId > 0;
    const subjectType = text(query.subject_type);
    activeSubjectType.value = subjectTypeInput.value =
      subjectType === '0' || subjectType === '1'
        ? subjectType
        : legacyPost
          ? '0'
          : '';
    activeSubjectKey.value = subjectKeyInput.value =
      text(query.subject_key) || (legacyPost ? String(postId) : '');
    activeQuery.value = queryInput.value = text(query.q).trim();
    activeAuthorName.value = authorNameInput.value = text(query.author).trim();
    const status = text(query.status);
    activeStatus.value = statusInput.value =
      status === '0' || status === '1' || status === '2' ? status : '';
    page.value = 1;
    void loadComments();
  },
  { immediate: true },
);
</script>

<template>
  <div class="mx-auto flex w-full max-w-[1000px] flex-col gap-4">
    <CommentFilters
      v-model:subject-type="subjectTypeInput"
      v-model:subject-key="subjectKeyInput"
      v-model:query="queryInput"
      v-model:author-name="authorNameInput"
      v-model:status="statusInput"
      @search="search"
    />

    <AdminAlert
      v-if="successMessage"
      type="success"
      closable
      @close="successMessage = ''"
    >
      {{ successMessage }}
    </AdminAlert>
    <AdminAlert v-if="errorMessage" type="error" title="评论加载失败">
      <div class="flex flex-col items-start gap-2">
        <p>{{ errorMessage }}</p>
        <XButton
          size="sm"
          color="neutral"
          variant="outline"
          @click="loadComments"
        >
          重新加载
        </XButton>
      </div>
    </AdminAlert>

    <AdminAlert
      v-if="actionErrorMessage"
      type="error"
      closable
      @close="actionErrorMessage = ''"
    >
      {{ actionErrorMessage }}
    </AdminAlert>

    <CommentList
      v-if="!errorMessage"
      :comments="comments"
      :loading="loading"
      :total="total"
      :page="page"
      :page-size="PAGE_SIZE"
      :subject-type="activeSubjectType"
      :subject-key="activeSubjectKey"
      :has-filters="hasFilters"
      :actions-disabled="actionsDisabled"
      @update-page="changePage"
      @moderate="requestModeration"
      @delete-author="requestAuthorDeletion"
      @filter-subject="filterSubject"
      @reset-filters="resetFilters"
    />

    <CommentModerationModal
      v-model:saving="moderationSaving"
      :target="pendingAction"
      @close="pendingAction = null"
      @success="handleModerationSuccess"
      @error="actionErrorMessage = $event"
    />

    <CommentAuthorDeleteModal
      v-model:saving="authorDeletionSaving"
      :target="pendingAuthorDeletion"
      @close="pendingAuthorDeletion = null"
      @success="handleAuthorDeletionSuccess"
      @error="actionErrorMessage = $event"
    />
  </div>
</template>
