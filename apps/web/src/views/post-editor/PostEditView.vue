<script setup lang="ts">
import { storeToRefs } from 'pinia';
import { computed, ref, watch } from 'vue';
import { RouterLink, useRoute, useRouter } from 'vue-router';

import { authUser, updatePost } from '@/api';
import AsyncContent from '@/components/AsyncContent.vue';
import { notifyError, notifySuccess } from '@/notifications';
import { useCategoryStore } from '@/stores/category';
import { usePostStore } from '@/stores/post';
import { getApiErrorMessage } from '@/utils/apiError';

import PostForm from './PostForm.vue';

const route = useRoute();
const router = useRouter();
const categoryStore = useCategoryStore();
const postStore = usePostStore();
const {
  currentPost: post,
  detailLoading: loading,
  detailError: error,
} = storeToRefs(postStore);
const title = ref('');
const content = ref('');
const selectedTagIds = ref<number[]>([]);
const submitting = ref(false);

const postId = computed(() => {
  const value = Number(route.params.id);
  return Number.isSafeInteger(value) && value > 0 ? value : 0;
});
const canEdit = computed(
  () =>
    post.value != null &&
    (authUser.value?.id === post.value.authorId ||
      authUser.value?.role === 'admin'),
);
const tags = computed(() =>
  post.value ? categoryStore.tagsByCategoryId(post.value.categoryId) : [],
);
const detailRoute = computed(() => ({
  name: 'post-detail' as const,
  params: { id: postId.value },
}));

async function loadPost() {
  await postStore.loadPost(postId.value);
}

async function save() {
  if (!post.value || submitting.value) return;
  submitting.value = true;
  try {
    const updatedPost = await updatePost(post.value.id, {
      title: title.value.trim(),
      content: content.value,
      tagIds: selectedTagIds.value,
    });
    postStore.setPost(updatedPost);
    notifySuccess('帖子修改已保存');
    await router.replace({
      name: 'post-detail',
      params: { id: updatedPost.id },
    });
  } catch (reason) {
    notifyError(await getApiErrorMessage(reason, '更新帖子失败'));
  } finally {
    submitting.value = false;
  }
}

function cancel() {
  void router.push(detailRoute.value);
}

watch(
  post,
  (value) => {
    if (!value) return;
    title.value = value.title;
    content.value = value.content;
    const validTagIds = new Set(tags.value.map((tag) => tag.id));
    selectedTagIds.value = value.tags
      .map((tag) => tag.id)
      .filter((id) => validTagIds.has(id));
  },
  { immediate: true },
);
watch(postId, loadPost, { immediate: true });
</script>

<template>
  <div class="page-container py-4 md:py-6">
    <div class="mx-auto max-w-4xl">
      <AsyncContent
        :loading="loading"
        :error="error"
        size="large"
        heading-tag="h1"
        error-title="帖子加载失败"
        @retry="loadPost"
      >
        <template #loading>
          <div class="py-6">
            <div class="h-7 w-32 animate-pulse rounded-sm bg-border" />
            <div class="mt-6 h-10 w-full animate-pulse rounded-sm bg-divider" />
            <div class="mt-5 h-10 w-full animate-pulse rounded-sm bg-divider" />
            <div class="mt-5 h-80 w-full animate-pulse rounded-sm bg-divider" />
          </div>
        </template>

        <section v-if="post && canEdit">
          <h1 class="text-xl font-bold text-ink">编辑帖子</h1>
          <PostForm
            v-model:title="title"
            v-model:content="content"
            v-model:tag-ids="selectedTagIds"
            class="mt-5"
            :tags="tags"
            :submitting="submitting"
            submit-label="保存修改"
            submitting-label="保存中…"
            tag-label="标签"
            show-cancel
            @submit="save"
            @cancel="cancel"
          />
        </section>

        <section v-else-if="post" class="py-8 text-center">
          <h1 class="text-xl font-bold text-ink">无法编辑帖子</h1>
          <p class="mt-2 text-sm text-muted">
            {{
              authUser
                ? '只有帖子作者或管理员可以编辑这篇帖子。'
                : '登录后才能编辑帖子，请使用页面右上角的登录入口。'
            }}
          </p>
          <RouterLink
            :to="detailRoute"
            class="mt-5 inline-flex rounded-sm border border-border px-4 py-2 text-sm font-medium text-muted transition-colors hover:border-primary hover:text-primary"
          >
            返回帖子详情
          </RouterLink>
        </section>
      </AsyncContent>
    </div>
  </div>
</template>
