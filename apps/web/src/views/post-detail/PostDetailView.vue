<script setup lang="ts">
import { ArrowBackOutlined } from '@vicons/material';
import { computed, nextTick, onBeforeUnmount, ref, watch } from 'vue';
import { RouterLink, useRoute, useRouter } from 'vue-router';

import {
  CATEGORIES,
  getPost,
  getPostComments,
  type Post,
  type PostComment,
} from '@/api';

import CommentComposer from './CommentComposer.vue';
import CommentList from './CommentList.vue';
import PostActions from './PostActions.vue';
import PostContent from './PostContent.vue';
import PostEditForm from './PostEditForm.vue';

const COMMENT_PAGE_SIZE = 50;

const route = useRoute();
const router = useRouter();
const post = ref<Post>();
const comments = ref<PostComment[]>([]);
const commentsTotal = ref(0);
const postLoading = ref(true);
const commentsLoading = ref(true);
const postError = ref('');
const commentsError = ref('');
const editingPost = ref(false);
const replyTo = ref<PostComment>();
let postController: AbortController | undefined;
let commentsController: AbortController | undefined;

const postId = computed(() => {
  const value = Number(route.params.id);
  return Number.isSafeInteger(value) && value > 0 ? value : 0;
});

const commentPage = computed(() => {
  const value = Number(route.query.commentPage);
  return Number.isInteger(value) && value > 0 ? value : 1;
});

const commentTotalPages = computed(() =>
  Math.max(1, Math.ceil(commentsTotal.value / COMMENT_PAGE_SIZE)),
);

const category = computed(() =>
  CATEGORIES.find((item) => item.id === post.value?.categoryId),
);

async function loadPost() {
  postController?.abort();
  post.value = undefined;
  postError.value = '';
  if (!postId.value) {
    postLoading.value = false;
    postError.value = '帖子地址无效';
    return;
  }

  const controller = new AbortController();
  postController = controller;
  postLoading.value = true;
  try {
    post.value = await getPost(postId.value, controller.signal);
    document.title = `${post.value.title} | Novelia Forum`;
  } catch (error) {
    if (error instanceof DOMException && error.name === 'AbortError') return;
    postError.value = error instanceof Error ? error.message : '无法加载帖子';
  } finally {
    if (postController === controller) postLoading.value = false;
  }
}

async function loadComments() {
  commentsController?.abort();
  comments.value = [];
  commentsTotal.value = 0;
  commentsError.value = '';
  if (!postId.value) {
    commentsLoading.value = false;
    return;
  }

  const controller = new AbortController();
  commentsController = controller;
  commentsLoading.value = true;
  try {
    const result = await getPostComments(
      postId.value,
      { page: commentPage.value, pageSize: COMMENT_PAGE_SIZE },
      controller.signal,
    );
    comments.value = result.items;
    commentsTotal.value = result.total;
  } catch (error) {
    if (error instanceof DOMException && error.name === 'AbortError') return;
    commentsError.value =
      error instanceof Error ? error.message : '无法加载评论';
  } finally {
    if (commentsController === controller) commentsLoading.value = false;
  }
}

function changeCommentPage(nextPage: number) {
  void router.push({
    name: 'post-detail',
    params: { id: postId.value },
    query: nextPage > 1 ? { commentPage: String(nextPage) } : undefined,
  });
  document.querySelector('#comments')?.scrollIntoView({ behavior: 'smooth' });
}

async function handleCommentCreated(comment: PostComment) {
  if (!post.value) return;
  commentsController?.abort();
  commentsController = undefined;
  commentsLoading.value = false;
  commentsError.value = '';
  const nextTotal = commentsTotal.value + 1;
  const lastPage = Math.max(1, Math.ceil(nextTotal / COMMENT_PAGE_SIZE));
  post.value = { ...post.value, commentsCount: post.value.commentsCount + 1 };
  commentsTotal.value = nextTotal;

  if (comment.rootId != null) {
    const lastReplyIndex = comments.value.reduce(
      (result, item, index) =>
        item.id === comment.rootId || item.rootId === comment.rootId
          ? index
          : result,
      -1,
    );
    comments.value.splice(lastReplyIndex + 1, 0, comment);
  } else if (commentPage.value === lastPage) {
    comments.value.push(comment);
  } else {
    await router.push({
      name: 'post-detail',
      params: { id: postId.value },
      query: { commentPage: String(lastPage) },
    });
  }
  replyTo.value = undefined;
  await nextTick();
  document
    .querySelector(`#comment-${comment.id}`)
    ?.scrollIntoView({ behavior: 'smooth', block: 'center' });
}

function startReply(comment: PostComment) {
  replyTo.value = comment;
  void nextTick(() =>
    document
      .querySelector('#comment-composer')
      ?.scrollIntoView({ behavior: 'smooth', block: 'center' }),
  );
}

function handleCommentUpdated(comment: PostComment) {
  const index = comments.value.findIndex((item) => item.id === comment.id);
  if (index >= 0) comments.value[index] = comment;
}

function handleCommentStatusChanged(id: number, status: number) {
  const comment = comments.value.find((item) => item.id === id);
  if (!comment) return;
  const wasPublished = comment.status === 0;
  comment.status = status;
  comment.content = '';
  if (post.value && wasPublished) {
    post.value = {
      ...post.value,
      commentsCount: Math.max(0, post.value.commentsCount - 1),
    };
  }
  if (replyTo.value?.id === id) replyTo.value = undefined;
}

function scrollToComposer() {
  document
    .querySelector('#comment-composer')
    ?.scrollIntoView({ behavior: 'smooth', block: 'center' });
}

function handlePostSaved(value: Post) {
  post.value = value;
  editingPost.value = false;
  document.title = `${value.title} | Novelia Forum`;
}

function handlePostUpdated(value: Post) {
  post.value = value;
}

function leaveDeletedPost() {
  void router.replace({
    name: 'posts',
    params: { slug: category.value?.slug ?? CATEGORIES[0].slug },
  });
}

watch(postId, loadPost, { immediate: true });
watch([postId, commentPage], loadComments, { immediate: true });

onBeforeUnmount(() => {
  postController?.abort();
  commentsController?.abort();
});
</script>

<template>
  <div class="page-container py-4 md:py-6">
    <div class="mx-auto max-w-4xl">
      <RouterLink
        :to="{
          name: 'posts',
          params: { slug: category?.slug ?? CATEGORIES[0].slug },
        }"
        class="mb-4 inline-flex items-center gap-1.5 rounded-sm text-sm font-medium text-muted transition-colors hover:text-primary focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-primary"
      >
        <ArrowBackOutlined class="size-4" aria-hidden="true" />
        返回帖子列表
      </RouterLink>

      <div v-if="postLoading" class="rounded-sm bg-surface px-4 py-6 sm:px-6">
        <div class="h-3 w-24 animate-pulse rounded-sm bg-divider" />
        <div class="mt-4 h-8 w-4/5 animate-pulse rounded-sm bg-border" />
        <div class="mt-5 h-4 w-56 animate-pulse rounded-sm bg-divider" />
        <div class="my-6 h-px bg-divider" />
        <div class="h-4 w-full animate-pulse rounded-sm bg-divider" />
        <div class="mt-3 h-4 w-full animate-pulse rounded-sm bg-divider" />
        <div class="mt-3 h-4 w-2/3 animate-pulse rounded-sm bg-divider" />
      </div>

      <div
        v-else-if="postError"
        class="grid min-h-96 place-items-center rounded-sm bg-surface p-8 text-center"
      >
        <div>
          <div
            class="mx-auto grid size-12 place-items-center rounded-full bg-red-50 text-red-500"
          >
            !
          </div>
          <h1 class="mt-4 text-lg font-semibold">帖子加载失败</h1>
          <p class="mt-2 text-sm text-muted">{{ postError }}</p>
          <button
            type="button"
            class="mt-5 rounded-sm bg-primary px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-primary-hover"
            @click="loadPost"
          >
            再试一次
          </button>
        </div>
      </div>

      <template v-else-if="post">
        <PostEditForm
          v-if="editingPost"
          :post="post"
          @cancel="editingPost = false"
          @saved="handlePostSaved"
        />
        <PostContent
          v-else
          :post="post"
          :category-name="category?.title ?? '未分类'"
        />
        <PostActions
          v-if="!editingPost"
          :post="post"
          @edit="editingPost = true"
          @deleted="leaveDeletedPost"
          @updated="handlePostUpdated"
          @comment="scrollToComposer"
        />
        <template v-if="!editingPost">
          <CommentComposer
            id="comment-composer"
            :post-id="post.id"
            :locked="post.commentsLocked"
            :reply-to="replyTo"
            @created="handleCommentCreated"
            @cancel-reply="replyTo = undefined"
          />
          <div id="comments">
            <CommentList
              :comments="comments"
              :loading="commentsLoading"
              :error="commentsError"
              :page="commentPage"
              :total="commentsTotal"
              :total-pages="commentTotalPages"
              :locked="post.commentsLocked"
              @retry="loadComments"
              @change-page="changeCommentPage"
              @reply="startReply"
              @updated="handleCommentUpdated"
              @status-changed="handleCommentStatusChanged"
            />
          </div>
        </template>
      </template>
    </div>
  </div>
</template>
