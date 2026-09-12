<script setup lang="ts">
import { SearchOutlined } from '@vicons/material';
import { onBeforeUnmount, ref, watch } from 'vue';

import { getCategoryTags, type CategoryTag, type PostSort } from '@/api';

const props = defineProps<{
  categoryId: number;
  query: string;
  tagId?: number;
  sort: PostSort;
}>();

const emit = defineEmits<{
  apply: [filters: { query: string; tagId?: number; sort: PostSort }];
}>();

const queryInput = ref(props.query);
const tagInput = ref(props.tagId ? String(props.tagId) : '');
const sortInput = ref<PostSort>(props.sort);
const tags = ref<CategoryTag[]>([]);
const tagsLoading = ref(false);
let tagsController: AbortController | undefined;

function syncInputs() {
  queryInput.value = props.query;
  tagInput.value = props.tagId ? String(props.tagId) : '';
  sortInput.value = props.sort;
}

async function loadTags() {
  tagsController?.abort();
  const controller = new AbortController();
  tagsController = controller;
  tagsLoading.value = true;
  try {
    tags.value = await getCategoryTags(props.categoryId, controller.signal);
  } catch (reason) {
    if (reason instanceof DOMException && reason.name === 'AbortError') return;
    tags.value = [];
  } finally {
    if (tagsController === controller) tagsLoading.value = false;
  }
}

function apply() {
  const tagId = Number(tagInput.value);
  emit('apply', {
    query: queryInput.value.trim(),
    tagId: Number.isSafeInteger(tagId) && tagId > 0 ? tagId : undefined,
    sort: sortInput.value,
  });
}

function reset() {
  queryInput.value = '';
  tagInput.value = '';
  sortInput.value = 'active';
  apply();
}

watch(() => [props.query, props.tagId, props.sort], syncInputs);
watch(
  () => props.categoryId,
  (_categoryId, previousCategoryId) => {
    if (previousCategoryId != null) tagInput.value = '';
    void loadTags();
  },
  { immediate: true },
);

onBeforeUnmount(() => tagsController?.abort());
</script>

<template>
  <form
    class="grid gap-3 rounded-sm bg-surface p-3 sm:grid-cols-[minmax(0,1fr)_10rem_10rem_auto]"
    role="search"
    @submit.prevent="apply"
  >
    <label class="relative min-w-0">
      <span class="sr-only">搜索帖子</span>
      <SearchOutlined
        class="pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted"
        aria-hidden="true"
      />
      <input
        v-model="queryInput"
        type="search"
        enterkeyhint="search"
        class="min-h-10 w-full rounded-sm border border-border bg-surface pr-3 pl-9 text-sm outline-none transition placeholder:text-muted/70 focus:border-primary focus:ring-2 focus:ring-primary/15"
        placeholder="搜索标题或正文，按 Enter 确认"
      />
    </label>

    <label>
      <span class="sr-only">按标签过滤</span>
      <select
        v-model="tagInput"
        class="min-h-10 w-full rounded-sm border border-border bg-surface px-3 text-sm outline-none transition focus:border-primary focus:ring-2 focus:ring-primary/15"
        :disabled="tagsLoading"
        @change="apply"
      >
        <option value="">{{ tagsLoading ? '加载标签中…' : '全部标签' }}</option>
        <option v-for="tag in tags" :key="tag.id" :value="String(tag.id)">
          {{ tag.name }}
        </option>
      </select>
    </label>

    <label>
      <span class="sr-only">帖子排序</span>
      <select
        v-model="sortInput"
        class="min-h-10 w-full rounded-sm border border-border bg-surface px-3 text-sm outline-none transition focus:border-primary focus:ring-2 focus:ring-primary/15"
        @change="apply"
      >
        <option value="active">最近活跃</option>
        <option value="newest">最新发布</option>
        <option value="views">浏览最多</option>
        <option value="comments">评论最多</option>
      </select>
    </label>

    <div
      v-if="queryInput || tagInput || sortInput !== 'active'"
      class="flex gap-2"
    >
      <button
        type="button"
        class="min-h-10 rounded-sm border border-border px-3 text-sm font-medium text-muted transition-colors hover:border-primary hover:text-primary"
        @click="reset"
      >
        重置
      </button>
    </div>
  </form>
</template>
