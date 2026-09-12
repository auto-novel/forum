<script setup lang="ts">
import { storeToRefs } from 'pinia';
import { computed, nextTick, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';

import { type Post, type PostComment } from '@/api';
import AsyncContent from '@/components/AsyncContent.vue';
import { useCategoryStore } from '@/stores/category';
import { useCommentStore } from '@/stores/comment';
import { usePostStore } from '@/stores/post';

import CommentComposer from './CommentComposer.vue';
import CommentList from './CommentList.vue';
import PostActions from './PostActions.vue';
import PostContent from './PostContent.vue';
import PostEditForm from './PostEditForm.vue';

const COMMENT_PAGE_SIZE = 50;

const route = useRoute();
const router = useRouter();
const categoryStore = useCategoryStore();
const commentStore = useCommentStore();
const postStore = usePostStore();
const {
  currentPost: post,
  detailLoading: postLoading,
  detailError: postError,
} = storeToRefs(postStore);
const editingPost = ref(false);
const replyTo = ref<PostComment>();

const postId = computed(() => {
  const value = Number(route.params.id);
  return Number.isSafeInteger(value) && value > 0 ? value : 0;
});

const commentPage = computed(() => {
  const value = Number(route.query.commentPage);
  return Number.isInteger(value) && value > 0 ? value : 1;
});

const commentState = computed(() =>
  commentStore.getPageState(postId.value, commentPage.value, COMMENT_PAGE_SIZE),
);
const comments = computed(() =>
  commentStore.getPageComments(
    postId.value,
    commentPage.value,
    COMMENT_PAGE_SIZE,
  ),
);
const commentsTotal = computed(() => commentState.value.total);
const commentsLoading = computed(() => commentState.value.loading);
const commentsError = computed(() => commentState.value.error);

const commentTotalPages = computed(() =>
  Math.max(1, Math.ceil(commentsTotal.value / COMMENT_PAGE_SIZE)),
);

const category = computed(() =>
  categoryStore.categories.find((item) => item.id === post.value?.categoryId),
);

async function loadPost() {
  const loadedPost = await postStore.loadPost(postId.value);
  if (loadedPost) document.title = `${loadedPost.title} | Novelia Forum`;
}

async function loadComments() {
  await commentStore.loadPage(
    postId.value,
    commentPage.value,
    COMMENT_PAGE_SIZE,
  );
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
  const lastPage = commentStore.registerCreatedComment(
    comment,
    commentPage.value,
    COMMENT_PAGE_SIZE,
  );
  postStore.setPost({
    ...post.value,
    commentsCount: post.value.commentsCount + 1,
  });

  if (comment.rootId == null && commentPage.value !== lastPage) {
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

function handleCommentStatusChanged(id: number) {
  if (post.value) {
    postStore.setPost({
      ...post.value,
      commentsCount: Math.max(0, post.value.commentsCount - 1),
    });
  }
  if (replyTo.value?.id === id) replyTo.value = undefined;
}

function scrollToComposer() {
  document
    .querySelector('#comment-composer')
    ?.scrollIntoView({ behavior: 'smooth', block: 'center' });
}

function handlePostSaved(value: Post) {
  postStore.setPost(value);
  editingPost.value = false;
  document.title = `${value.title} | Novelia Forum`;
}

function handlePostUpdated(value: Post) {
  postStore.setPost(value);
}

function leaveDeletedPost() {
  void router.replace({
    name: 'posts',
    params: {
      slug: category.value?.slug ?? categoryStore.defaultCategory.slug,
    },
  });
}

watch(postId, loadPost, { immediate: true });
watch([postId, commentPage], loadComments, { immediate: true });
</script>

<template>
  <div class="page-container py-4 md:py-6">
    <div class="mx-auto max-w-4xl">
      <AsyncContent
        :loading="postLoading"
        :error="postError"
        size="large"
        heading-tag="h1"
        error-title="帖子加载失败"
        state-class="rounded-sm bg-surface"
        @retry="loadPost"
      >
        <template #loading>
          <div class="rounded-sm bg-surface px-4 py-6 sm:px-6">
            <div class="h-3 w-24 animate-pulse rounded-sm bg-divider" />
            <div class="mt-4 h-8 w-4/5 animate-pulse rounded-sm bg-border" />
            <div class="mt-5 h-4 w-56 animate-pulse rounded-sm bg-divider" />
            <div class="my-6 h-px bg-divider" />
            <div class="h-4 w-full animate-pulse rounded-sm bg-divider" />
            <div class="mt-3 h-4 w-full animate-pulse rounded-sm bg-divider" />
            <div class="mt-3 h-4 w-2/3 animate-pulse rounded-sm bg-divider" />
          </div>
        </template>

        <template v-if="post">
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
                @status-changed="handleCommentStatusChanged"
              />
            </div>
          </template>
        </template>
      </AsyncContent>
    </div>
  </div>
</template>
