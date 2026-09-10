<script setup lang="ts">
import { ArrowBackOutlined } from '@vicons/material';
import { computed, onBeforeUnmount, ref, watch } from 'vue';
import { RouterLink, useRoute, useRouter } from 'vue-router';

import {
  authUser,
  CATEGORIES,
  createPost,
  getCategoryTags,
  type CategoryTag,
} from '@/api';
import MarkdownEditor from '@/components/markdown/MarkdownEditor.vue';

interface PostDraft {
  title: string;
  category: string;
  tagIds: number[];
  content: string;
}

const route = useRoute();
const router = useRouter();
const initialCategory =
  CATEGORIES.find((category) => category.slug === route.query.category)?.slug ??
  CATEGORIES[0].slug;
const title = ref('');
const categorySlug = ref(initialCategory);
const selectedTagIds = ref<number[]>([]);
const content = ref('');
const tags = ref<CategoryTag[]>([]);
const tagsLoading = ref(false);
const tagsError = ref('');
const submitting = ref(false);
const submitError = ref('');
let tagsController: AbortController | undefined;

const selectedCategory = computed(
  () =>
    CATEGORIES.find((category) => category.slug === categorySlug.value) ??
    CATEGORIES[0],
);
const activeTags = computed(() => tags.value.filter((tag) => tag.isActive));
const draftKey = computed(() =>
  authUser.value ? `forum:post-draft:${authUser.value.id}` : '',
);
const canSubmit = computed(
  () =>
    Boolean(title.value.trim() && content.value.trim()) &&
    !tagsLoading.value &&
    !submitting.value,
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
    const category = CATEGORIES.some((item) => item.slug === value.category)
      ? String(value.category)
      : initialCategory;
    const tagIds = Array.isArray(value.tagIds)
      ? value.tagIds.filter(
          (id): id is number => Number.isSafeInteger(id) && id > 0,
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

async function loadTags() {
  tagsController?.abort();
  const controller = new AbortController();
  tagsController = controller;
  tagsLoading.value = true;
  tagsError.value = '';
  try {
    tags.value = await getCategoryTags(
      selectedCategory.value.id,
      controller.signal,
    );
    const validIds = new Set(activeTags.value.map((tag) => tag.id));
    selectedTagIds.value = selectedTagIds.value.filter((id) =>
      validIds.has(id),
    );
  } catch (error) {
    if (error instanceof DOMException && error.name === 'AbortError') return;
    tags.value = [];
    tagsError.value = error instanceof Error ? error.message : '无法加载标签';
  } finally {
    if (tagsController === controller) tagsLoading.value = false;
  }
}

function changeCategory() {
  selectedTagIds.value = [];
}

async function responseErrorMessage(error: unknown) {
  if (error && typeof error === 'object' && 'response' in error) {
    const response = (error as { response?: Response }).response;
    if (response) {
      try {
        const message = await response.text();
        if (message) return message;
      } catch {
        // Fall back to the client error message when the body is unavailable.
      }
    }
  }
  return error instanceof Error ? error.message : '帖子发布失败';
}

async function submitPost() {
  if (!canSubmit.value || !authUser.value) return;
  submitting.value = true;
  submitError.value = '';
  try {
    const post = await createPost({
      category: categorySlug.value,
      title: title.value.trim(),
      content: content.value,
      tagIds: tagsError.value ? [] : selectedTagIds.value,
    });
    writeDraft(draftKey.value);
    title.value = '';
    content.value = '';
    selectedTagIds.value = [];
    await router.replace({ name: 'post-detail', params: { id: post.id } });
  } catch (error) {
    submitError.value = await responseErrorMessage(error);
  } finally {
    submitting.value = false;
  }
}

watch(categorySlug, loadTags, { immediate: true });

onBeforeUnmount(() => tagsController?.abort());
</script>

<template>
  <div class="page-container py-4 md:py-6">
    <div class="mx-auto max-w-4xl">
      <RouterLink
        :to="{ name: 'posts', params: { slug: categorySlug } }"
        class="mb-4 inline-flex items-center gap-1.5 rounded-sm text-sm font-medium text-muted transition-colors hover:text-primary focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-primary"
      >
        <ArrowBackOutlined class="size-4" aria-hidden="true" />
        返回帖子列表
      </RouterLink>

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

        <form v-else class="mt-6 space-y-5" @submit.prevent="submitPost">
          <div>
            <label
              for="post-title"
              class="mb-2 block text-sm font-semibold text-ink"
            >
              标题
            </label>
            <input
              id="post-title"
              v-model="title"
              type="text"
              maxlength="500"
              class="block min-h-10 w-full rounded-md border border-border bg-surface px-3 text-sm text-ink outline-none transition focus:border-primary focus:ring-2 focus:ring-primary/15 disabled:cursor-not-allowed disabled:bg-paper"
              placeholder="用一句话概括你想讨论的内容"
              :disabled="submitting"
              required
            />
            <p class="mt-1 text-right text-xs text-muted">
              {{ title.length }} / 500
            </p>
          </div>

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
                v-for="category in CATEGORIES"
                :key="category.id"
                :value="category.slug"
              >
                {{ category.title }}
              </option>
            </select>
          </div>

          <fieldset>
            <legend class="mb-2 text-sm font-semibold text-ink">
              标签（可选）
            </legend>
            <div
              v-if="tagsLoading"
              class="flex gap-2"
              aria-label="正在加载标签"
            >
              <span
                v-for="index in 3"
                :key="index"
                class="h-7 w-20 animate-pulse rounded-sm bg-divider"
              />
            </div>
            <div
              v-else-if="tagsError"
              class="flex items-center gap-3 text-sm text-red-600"
            >
              <span>{{ tagsError }}</span>
              <button
                type="button"
                class="font-medium text-primary hover:text-primary-hover"
                @click="loadTags"
              >
                重试
              </button>
            </div>
            <div v-else-if="activeTags.length" class="flex flex-wrap gap-2">
              <label
                v-for="tag in activeTags"
                :key="tag.id"
                class="cursor-pointer rounded-sm border px-3 py-1.5 text-xs font-medium transition-colors"
                :class="
                  selectedTagIds.includes(tag.id)
                    ? 'border-primary bg-primary-soft text-primary'
                    : 'border-border text-muted hover:border-primary hover:text-primary'
                "
              >
                <input
                  v-model="selectedTagIds"
                  type="checkbox"
                  class="sr-only"
                  :value="tag.id"
                  :disabled="submitting"
                />
                {{ tag.name }}
              </label>
            </div>
            <p v-else class="text-sm text-muted">这个分类暂时没有可用标签。</p>
          </fieldset>

          <div>
            <label class="mb-2 block text-sm font-semibold text-ink">
              正文
            </label>
            <MarkdownEditor
              v-model="content"
              mode="article"
              placeholder="详细说明你想分享或讨论的内容…"
              :rows="14"
              :disabled="submitting"
            />
          </div>

          <div
            class="flex flex-wrap items-center justify-between gap-3 border-t border-divider pt-5"
          >
            <p class="text-xs text-muted">
              内容支持 Markdown，草稿会自动保存在本机。
            </p>
            <button
              type="submit"
              class="rounded-sm bg-primary px-5 py-2.5 text-sm font-medium text-white transition-colors hover:bg-primary-hover disabled:cursor-not-allowed disabled:opacity-50"
              :disabled="!canSubmit"
            >
              {{ submitting ? '发布中…' : '发布帖子' }}
            </button>
          </div>
          <p v-if="submitError" class="text-sm text-red-600" role="alert">
            {{ submitError }}
          </p>
        </form>
      </section>
    </div>
  </div>
</template>
