<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';

import { useForumApi } from '@/api';
import { usePostStore } from '@/stores/post';
import { Toast, XPage, getApiErrorMessage, useWebKit } from '@novelia/web-kit';
import { useCategoryStore } from '@/stores/category';
import { usePostDraftStore } from '@/stores/draft';
import PostPublishingNotice from './PostPublishingNotice.vue';

import PostForm from './PostForm.vue';

const forumApi = useForumApi();
const { whoami } = useWebKit();
const postStore = usePostStore();
const route = useRoute();
const router = useRouter();
const categoryStore = useCategoryStore();
const draftStore = usePostDraftStore();
const initialCategory =
  categoryStore.writableCategories.find(
    (category) => category.slug === route.query.category,
  )?.slug ?? categoryStore.defaultCategory.slug;
const title = ref('');
const categorySlug = ref(initialCategory);
const selectedTagIds = ref<number[]>([]);
const content = ref('');
const submitting = ref(false);

const selectedCategory = computed(
  () =>
    categoryStore.writableCategories.find(
      (category) => category.slug === categorySlug.value,
    ) ?? categoryStore.defaultCategory,
);
const tags = computed(() => selectedCategory.value.tags);
const draftUserId = computed(() => whoami.value.user?.id ?? 0);
const categoriesReady = computed(() => categoryStore.items.length > 0);

watch(
  [draftUserId, categoriesReady],
  ([userId, ready]) => {
    title.value = '';
    categorySlug.value = initialCategory;
    selectedTagIds.value = [];
    content.value = '';
    if (!userId || !ready) return;
    const draft = draftStore.get(userId);
    if (!draft) return;
    const categoryItem =
      categoryStore.writableCategories.find(
        (item) => item.slug === draft.category,
      ) ?? categoryStore.defaultCategory;
    const validIds = new Set(categoryItem.tags.map((tag) => tag.id));
    title.value = draft.title;
    categorySlug.value = categoryItem.slug;
    selectedTagIds.value = draft.tagIds.filter((id) => validIds.has(id));
    content.value = draft.content;
  },
  { immediate: true },
);

watch(
  [title, categorySlug, selectedTagIds, content],
  () => {
    if (!draftUserId.value || !categoriesReady.value) return;
    draftStore.save(draftUserId.value, {
      title: title.value,
      category: categorySlug.value,
      tagIds: selectedTagIds.value,
      content: content.value,
    });
  },
  { deep: true },
);

function changeCategory() {
  selectedTagIds.value = [];
}

async function submitPost() {
  if (!whoami.value.isSignedIn || !categoriesReady.value || submitting.value)
    return;
  submitting.value = true;
  await Toast.promise(
    forumApi.createPost({
      categoryId: selectedCategory.value.id,
      title: title.value.trim(),
      content: content.value,
      tagIds: selectedTagIds.value,
    }),
    {
      success: '帖子已发布',
      error: (reason) => getApiErrorMessage(reason, '帖子发布失败'),
    },
  )
    .then((post) => {
      postStore.setPost(post);
      draftStore.clear(draftUserId.value);
      title.value = '';
      content.value = '';
      selectedTagIds.value = [];
      // 帖子已发布，导航失败无需改变提交结果。
      return Toast.promise(
        router.replace({ name: 'post-detail', params: { id: post.id } }),
        { error: '无法打开帖子详情' },
      );
    })
    .catch(() => {});
  submitting.value = false;
}
</script>

<template>
  <XPage>
    <UPageHeader title="发表帖子" />
    <section>
      <div v-if="!whoami.isSignedIn" class="py-4 text-sm text-muted">
        登录后才能发表帖子，请使用页面右上角的登录入口。
      </div>

      <p
        v-else-if="!categoriesReady && categoryStore.loading"
        class="py-4 text-sm text-muted"
        role="status"
      >
        正在加载分类…
      </p>
      <UError
        :ui="{
          root: 'min-h-0 py-8',
          statusMessage: 'text-lg font-medium sm:text-xl',
          message: 'mt-2 text-sm',
          links: 'mt-5 flex-wrap gap-3',
        }"
        v-else-if="!categoriesReady && categoryStore.error"
        :error="{ statusMessage: '分类加载失败', message: categoryStore.error }"
        :clear="false"
        :as="{ root: 'div', statusMessage: 'h2' }"
        role="alert"
      >
        <template #links>
          <UButton size="lg" @click="categoryStore.refresh">再试一次</UButton>
        </template>
      </UError>
      <template v-else>
        <PostForm
          v-model:title="title"
          v-model:category="categorySlug"
          v-model:content="content"
          v-model:tag-ids="selectedTagIds"
          :categories="categoryStore.writableCategories"
          :tags="tags"
          :submitting="submitting"
          submit-label="发布帖子"
          submitting-label="发布中…"
          title-placeholder="用一句话概括你想讨论的内容"
          content-placeholder="详细说明你想分享或讨论的内容…"
          @category-change="changeCategory"
          @submit="submitPost"
        >
          <template #notice>
            <PostPublishingNotice />
          </template>
        </PostForm>
      </template>
    </section>
  </XPage>
</template>
