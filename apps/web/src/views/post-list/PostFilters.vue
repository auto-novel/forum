<script setup lang="ts">
import { ExpandMoreOutlined, SearchOutlined } from '@vicons/material';
import { computed, ref, watch } from 'vue';

import { type PostSort } from '@/api';
import { useCategoryStore } from '@/stores/category';

const props = defineProps<{
  categoryId: number;
  query: string;
  tagId?: number;
  sort: PostSort;
}>();

const emit = defineEmits<{
  apply: [filters: { query: string; tagId?: number; sort: PostSort }];
}>();
const categoryStore = useCategoryStore();

const queryInput = ref(props.query);
const tagInput = ref(props.tagId ? String(props.tagId) : '');
const sortInput = ref<PostSort>(props.sort);
const tags = computed(() => categoryStore.tagsByCategoryId(props.categoryId));

function syncInputs() {
  queryInput.value = props.query;
  tagInput.value = props.tagId ? String(props.tagId) : '';
  sortInput.value = props.sort;
}

function apply() {
  const tagId = Number(tagInput.value);
  emit('apply', {
    query: queryInput.value.trim(),
    tagId: Number.isSafeInteger(tagId) && tagId > 0 ? tagId : undefined,
    sort: sortInput.value,
  });
}

watch(() => [props.query, props.tagId, props.sort], syncInputs);
watch(
  () => props.categoryId,
  (_categoryId, previousCategoryId) => {
    if (previousCategoryId != null) tagInput.value = '';
  },
);
</script>

<template>
  <form
    class="flex min-w-0 flex-wrap gap-2 sm:flex-nowrap"
    role="search"
    @submit.prevent="apply"
  >
    <label class="relative w-full min-w-0 sm:flex-1">
      <span class="sr-only">搜索帖子</span>
      <SearchOutlined
        class="pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted"
        aria-hidden="true"
      />
      <input
        v-model="queryInput"
        type="search"
        enterkeyhint="search"
        class="min-h-10 w-full rounded-sm border border-border bg-transparent pr-3 pl-9 text-sm outline-none transition placeholder:text-muted/70 focus:border-primary focus:ring-2 focus:ring-primary/15 sm:pr-10"
        placeholder="搜索标题或正文"
      />
      <span
        class="pointer-events-none absolute top-1/2 right-3 hidden -translate-y-1/2 text-sm text-muted/70 sm:block"
        aria-hidden="true"
      >
        ↵
      </span>
    </label>

    <label
      v-if="tags.length"
      class="relative min-w-0 flex-1 sm:w-32 sm:flex-none"
    >
      <span class="sr-only">按标签过滤</span>
      <select
        v-model="tagInput"
        class="min-h-10 w-full appearance-none truncate rounded-sm border border-border pr-9 pl-3 text-sm font-normal outline-none transition focus:border-primary focus:ring-2 focus:ring-primary/15"
        :class="
          tagInput ? 'bg-primary/5 text-primary' : 'bg-transparent text-ink'
        "
        @change="apply"
      >
        <option value="">全部标签</option>
        <option v-for="tag in tags" :key="tag.id" :value="String(tag.id)">
          {{ tag.name }}
        </option>
      </select>
      <ExpandMoreOutlined
        class="pointer-events-none absolute top-1/2 right-3 size-4 -translate-y-1/2 text-muted"
        aria-hidden="true"
      />
    </label>

    <label class="relative min-w-0 flex-1 sm:w-32 sm:flex-none">
      <span class="sr-only">帖子排序</span>
      <select
        v-model="sortInput"
        class="min-h-10 w-full appearance-none truncate rounded-sm border border-border pr-9 pl-3 text-sm font-normal outline-none transition focus:border-primary focus:ring-2 focus:ring-primary/15"
        :class="
          sortInput !== 'active'
            ? 'bg-primary/5 text-primary'
            : 'bg-transparent text-ink'
        "
        @change="apply"
      >
        <option value="active">最近活跃</option>
        <option value="newest">最新发布</option>
        <option value="views">浏览最多</option>
        <option value="comments">评论最多</option>
      </select>
      <ExpandMoreOutlined
        class="pointer-events-none absolute top-1/2 right-3 size-4 -translate-y-1/2 text-muted"
        aria-hidden="true"
      />
    </label>
  </form>
</template>
