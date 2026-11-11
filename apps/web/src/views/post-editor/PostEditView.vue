<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';

import { useForumApi, type Post } from '@/api';
import { Toast, XPage, getApiErrorMessage, useWebKit } from '@novelia/web-kit';
import { useUnsavedChangesGuard } from '@/composables/useUnsavedChangesGuard';
import { useCategoryStore } from '@/stores/category';
import { usePostQuery, usePostStore } from '@/stores/post';

import PostForm from './PostForm.vue';

const forumApi = useForumApi();
const { whoami } = useWebKit();
const route = useRoute();
const router = useRouter();
const categoryStore = useCategoryStore();
const postStore = usePostStore();

const title = ref('');
const content = ref('');
const categorySlug = ref('');
const selectedTagIds = ref<number[]>([]);
const submitting = ref(false);
const savedSnapshot = ref('');

const postId = computed(() => {
  const value = Number(route.params.id);
  return Number.isSafeInteger(value) && value > 0 ? value : 0;
});
const categoriesReady = ref(false);
const {
  post,
  loading: postLoading,
  error,
  retry,
} = usePostQuery(postId, { editing: true, enabled: categoriesReady });
const loading = computed(() => !categoriesReady.value || postLoading.value);
const canEdit = computed(
  () =>
    (post.value != null &&
      categoryStore.canPublish(
        categoryStore.categories.find(
          (item) => item.id === post.value?.categoryId,
        )?.slug ?? '',
      ) &&
      whoami.value.user?.id === post.value.authorId) ||
    whoami.value.asAdmin,
);
const selectedCategory = computed(
  () =>
    categoryStore.categories.find(
      (category) => category.slug === categorySlug.value,
    ) ?? categoryStore.defaultCategory,
);
const tags = computed(() => selectedCategory.value.tags);
const detailRoute = computed(() => ({
  name: 'post-detail' as const,
  params: { id: postId.value },
}));
const hasUnsavedChanges = computed(
  () => savedSnapshot.value !== '' && savedSnapshot.value !== formSnapshot(),
);

useUnsavedChangesGuard(hasUnsavedChanges);

function formSnapshot() {
  return JSON.stringify([
    title.value,
    content.value,
    categorySlug.value,
    [...selectedTagIds.value].sort((left, right) => left - right),
  ]);
}

function applyPost(value: Post) {
  title.value = value.title;
  content.value = value.content;
  const category =
    categoryStore.categories.find((item) => item.id === value.categoryId) ??
    categoryStore.defaultCategory;
  categorySlug.value = category.slug;
  const validTagIds = new Set(category.tags.map((tag) => tag.id));
  selectedTagIds.value = value.tags
    .map((tag) => tag.id)
    .filter((id) => validTagIds.has(id));
  savedSnapshot.value = formSnapshot();
}

async function loadPost() {
  await categoryStore.initialize();
  await retry();
}

function changeCategory() {
  selectedTagIds.value = [];
}

async function save() {
  if (!post.value || submitting.value) return;
  submitting.value = true;
  await Toast.promise(
    forumApi.updatePost(post.value.id, {
      categoryId: selectedCategory.value.id,
      title: title.value.trim(),
      content: content.value,
      tagIds: selectedTagIds.value,
    }),
    {
      success: '帖子修改已保存',
      error: (reason) => getApiErrorMessage(reason, '更新帖子失败'),
    },
  )
    .then((updatedPost) => {
      applyPost(updatedPost);
      postStore.setPost(updatedPost);
      // 修改已保存，导航失败无需改变保存结果。
      return Toast.promise(
        router.replace({
          name: 'post-detail',
          params: { id: updatedPost.id },
        }),
        { error: '无法打开帖子详情' },
      );
    })
    .catch(() => {});
  submitting.value = false;
}

function cancel() {
  void router.push(detailRoute.value);
}

watch(
  [post, categoriesReady],
  ([value, ready]) => {
    if (!ready || !value || hasUnsavedChanges.value) return;
    applyPost(value);
  },
  { immediate: true },
);
watch(
  postId,
  () => {
    savedSnapshot.value = '';
  },
  { flush: 'sync' },
);
void categoryStore.initialize().then(() => {
  categoriesReady.value = true;
});
</script>

<template>
  <XPage>
    <UPageHeader title="编辑帖子" />
    <div v-if="loading" class="py-6" role="status">
      <span class="sr-only">正在加载帖子…</span>
      <USkeleton class="h-10 w-full" />
      <USkeleton class="mt-5 h-10 w-full" />
      <USkeleton class="mt-5 h-80 w-full" />
    </div>
    <UError
      :ui="{
        root: 'min-h-0 py-8',
        statusMessage: 'text-lg font-medium sm:text-xl',
        message: 'mt-2 text-sm',
        links: 'mt-5 flex-wrap gap-3',
      }"
      v-else-if="error"
      class="min-h-96"
      :error="{ statusMessage: '帖子加载失败', message: error }"
      :clear="false"
      :as="{ root: 'div', statusMessage: 'h2' }"
      role="alert"
    >
      <template #links>
        <UButton size="lg" @click="loadPost">再试一次</UButton>
      </template>
    </UError>
    <template v-else>
      <section v-if="post && canEdit">
        <PostForm
          v-model:title="title"
          v-model:category="categorySlug"
          v-model:content="content"
          v-model:tag-ids="selectedTagIds"
          :categories="categoryStore.writableCategories"
          :tags="tags"
          :submitting="submitting"
          submit-label="保存修改"
          submitting-label="保存中…"
          title-placeholder="用一句话概括你想讨论的内容"
          content-placeholder="详细说明你想分享或讨论的内容…"
          show-cancel
          @category-change="changeCategory"
          @submit="save"
          @cancel="cancel"
        />
      </section>

      <section v-else-if="post" class="py-8 text-center">
        <h1 class="text-2xl font-bold text-ink">无法编辑帖子</h1>
        <p class="mt-2 text-sm text-muted">
          {{
            whoami.isSignedIn
              ? '只有具有该板块发帖权限的作者或管理员可以编辑这篇帖子。'
              : '登录后才能编辑帖子，请使用页面右上角的登录入口。'
          }}
        </p>
        <UButton
          size="lg"
          :to="detailRoute"
          color="neutral"
          variant="outline"
          class="mt-5"
        >
          返回帖子详情
        </UButton>
      </section>
    </template>
  </XPage>
</template>
