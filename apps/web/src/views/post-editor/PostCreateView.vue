<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';

import { authUser, createPost } from '@/api';
import { notifyError, notifySuccess } from '@/notifications';
import { useCategoryStore } from '@/stores/category';
import { getApiErrorMessage } from '@/utils/apiError';

import PostForm from './PostForm.vue';

interface PostDraft {
  title: string;
  category: string;
  tagIds: number[];
  content: string;
}

const route = useRoute();
const router = useRouter();
const categoryStore = useCategoryStore();
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
const draftKey = computed(() =>
  authUser.value ? `forum:post-draft:${authUser.value.id}` : '',
);

function readDraft(key: string): PostDraft | undefined {
  if (!key) return;
  try {
    const value = JSON.parse(
      localStorage.getItem(key) ?? 'null',
    ) as Partial<PostDraft> | null;
    if (
      !value ||
      typeof value.title !== 'string' ||
      typeof value.content !== 'string'
    )
      return;
    const categoryItem =
      categoryStore.categories.find((item) => item.slug === value.category) ??
      categoryStore.defaultCategory;
    const category = categoryItem.slug;
    const validIds = new Set(categoryItem.tags.map((tag) => tag.id));
    const tagIds = Array.isArray(value.tagIds)
      ? value.tagIds.filter(
          (id): id is number =>
            Number.isSafeInteger(id) && id > 0 && validIds.has(id),
        )
      : [];
    return { title: value.title, category, tagIds, content: value.content };
  } catch {
    return;
  }
}

function writeDraft(key: string, draft?: PostDraft) {
  if (!key) return;
  try {
    if (draft && (draft.title.trim() || draft.content.trim())) {
      localStorage.setItem(key, JSON.stringify(draft));
    } else {
      localStorage.removeItem(key);
    }
  } catch {
    // Draft persistence is optional when browser storage is unavailable.
  }
}

watch(
  draftKey,
  (key) => {
    const draft = readDraft(key);
    if (!draft) return;
    title.value = draft.title;
    categorySlug.value = draft.category;
    selectedTagIds.value = draft.tagIds;
    content.value = draft.content;
  },
  { immediate: true },
);

watch(
  [title, categorySlug, selectedTagIds, content],
  () => {
    writeDraft(draftKey.value, {
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
    writeDraft(draftKey.value);
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
      <section class="rounded-sm bg-surface px-4 py-5 sm:px-6 sm:py-7">
        <h1 class="text-2xl font-bold tracking-tight text-ink">发表帖子</h1>
        <p class="mt-1 text-sm text-muted">
          分享内容前，请选择最合适的讨论分类。
        </p>

        <div
          v-if="!authUser"
          class="mt-6 rounded-md border border-divider bg-paper px-4 py-4 text-sm text-muted"
        >
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
                class="block min-h-10 w-full rounded-md border border-border bg-surface px-3 text-sm text-ink outline-none transition focus:border-primary focus:ring-2 focus:ring-primary/15 disabled:cursor-not-allowed disabled:bg-paper"
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

          <template #hint>
            <p class="text-xs text-muted">
              内容支持 Markdown，草稿会自动保存在本机。
            </p>
          </template>
        </PostForm>
      </section>
    </div>
  </div>
</template>
