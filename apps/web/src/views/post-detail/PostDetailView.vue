<script setup lang="ts">
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
import PostContent from './PostContent.vue';

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
  const nextTotal = post.value.commentsCount + 1;
  const lastPage = Math.max(1, Math.ceil(nextTotal / COMMENT_PAGE_SIZE));
  post.value = { ...post.value, commentsCount: nextTotal };
  commentsTotal.value = nextTotal;

  if (commentPage.value === lastPage) {
    comments.value.push(comment);
  } else {
    await router.push({
      name: 'post-detail',
      params: { id: postId.value },
      query: { commentPage: String(lastPage) },
    });
  }
  await nextTick();
  document.querySelector('#comments')?.scrollIntoView({ behavior: 'smooth' });
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
          query: category ? { category: category.slug } : undefined,
        }"
        class="mb-4 inline-flex items-center gap-1.5 rounded-sm text-sm font-medium text-muted transition-colors hover:text-primary focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-primary"
      >
        <svg viewBox="0 0 20 20" class="size-4" fill="none" aria-hidden="true">
          <path
            d="m11.5 5-5 5 5 5"
            stroke="currentColor"
            stroke-width="1.7"
            stroke-linecap="round"
            stroke-linejoin="round"
          />
        </svg>
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
        <PostContent
          :post="post"
          :category-name="category?.title ?? '未分类'"
        />
        <CommentComposer
          :post-id="post.id"
          :locked="post.commentsLocked"
          @created="handleCommentCreated"
        />
        <div id="comments">
          <CommentList
            :comments="comments"
            :loading="commentsLoading"
            :error="commentsError"
            :page="commentPage"
            :total="commentsTotal"
            :total-pages="commentTotalPages"
            @retry="loadComments"
            @change-page="changeCommentPage"
          />
        </div>
      </template>
    </div>
  </div>
</template>
