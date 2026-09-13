<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';

import { authUser, createPost } from '@/api';
import { notifyError, notifySuccess } from '@/notifications';
import { useCategoryStore } from '@/stores/category';
import { useDraftStore } from '@/stores/draft';
import { getApiErrorMessage } from '@/utils/apiError';

import PostForm from './PostForm.vue';

const route = useRoute();
const router = useRouter();
const categoryStore = useCategoryStore();
const draftStore = useDraftStore();
const initialCategory =
  categoryStore.categories.find(
    (category) => category.slug === route.query.category,
  )?.slug ?? categoryStore.defaultCategory.slug;
const title = ref('');
const categorySlug = ref(initialCategory);
const selectedTagIds = ref<number[]>([]);
const content = ref('');
const submitting = ref(false);

const selectedCategory = computed(
  () =>
    categoryStore.categories.find(
      (category) => category.slug === categorySlug.value,
    ) ?? categoryStore.defaultCategory,
);
const tags = computed(() => selectedCategory.value.tags);
const draftUserId = computed(() => authUser.value?.id ?? 0);

watch(
  draftUserId,
  (userId) => {
    if (!userId) return;
    const draft = draftStore.getPostDraft(userId);
    if (!draft) return;
    const categoryItem =
      categoryStore.categories.find((item) => item.slug === draft.category) ??
      categoryStore.defaultCategory;
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
    if (!draftUserId.value) return;
    draftStore.savePostDraft(draftUserId.value, {
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
  if (!authUser.value || submitting.value) return;
  submitting.value = true;
  try {
    const post = await createPost({
      category: categorySlug.value,
      title: title.value.trim(),
      content: content.value,
      tagIds: selectedTagIds.value,
    });
    draftStore.clearPostDraft(draftUserId.value);
    title.value = '';
    content.value = '';
    selectedTagIds.value = [];
    notifySuccess('帖子已发布');
    await router.replace({ name: 'post-detail', params: { id: post.id } });
  } catch (error) {
    notifyError(await getApiErrorMessage(error, '帖子发布失败'));
  } finally {
    submitting.value = false;
  }
}
</script>

<template>
  <div class="page-container py-4 md:py-6">
    <div class="mx-auto max-w-4xl">
      <section>
        <h1 class="text-xl font-bold text-ink">发表帖子</h1>
        <p class="mt-1 text-sm text-muted">
          分享内容前，请选择最合适的讨论分类。
        </p>

        <div v-if="!authUser" class="mt-6 py-4 text-sm text-muted">
          登录后才能发表帖子，请使用页面右上角的登录入口。
        </div>

        <PostForm
          v-else
          v-model:title="title"
          v-model:content="content"
          v-model:tag-ids="selectedTagIds"
          class="mt-6"
          :tags="tags"
          :submitting="submitting"
          submit-label="发布帖子"
          submitting-label="发布中…"
          title-placeholder="用一句话概括你想讨论的内容"
          content-placeholder="详细说明你想分享或讨论的内容…"
          show-title-count
          @submit="submitPost"
        >
          <template #category>
            <div>
              <label
                for="post-category"
                class="mb-2 block text-sm font-semibold text-ink"
              >
                分类
              </label>
              <select
                id="post-category"
                v-model="categorySlug"
                class="block min-h-10 w-full rounded-md border border-border bg-transparent px-3 text-sm text-ink outline-none transition focus:border-primary focus:ring-2 focus:ring-primary/15 disabled:cursor-not-allowed disabled:opacity-50"
                :disabled="submitting"
                @change="changeCategory"
              >
                <option
                  v-for="category in categoryStore.categories"
                  :key="category.id"
                  :value="category.slug"
                >
                  {{ category.title }}
                </option>
              </select>
            </div>
          </template>
        </PostForm>
      </section>
    </div>
  </div>
</template>
