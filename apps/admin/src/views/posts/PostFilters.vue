<script setup lang="ts">
import { SearchOutlined } from '@vicons/material';
import { XButton, XInput } from '@novelia/web-kit';
import { computed, ref } from 'vue';

import FilterChoiceGroup from '@/components/FilterChoiceGroup.vue';
import FilterRow from '@/components/FilterRow.vue';
import type { CategoryListItem, PostSort, PostStatusFilter } from '@/api';
import { categoryOrder, categoryTitle } from '@/category';

const props = defineProps<{ categories: CategoryListItem[] }>();
const emit = defineEmits<{ search: [] }>();

const query = defineModel<string>('query', { required: true });
const category = defineModel<string>('category', { required: true });
const status = defineModel<PostStatusFilter | ''>('status', { required: true });
const tagId = defineModel<number | null>('tagId', { required: true });
const authorName = defineModel<string>('authorName', { required: true });
const sort = defineModel<PostSort>('sort', { required: true });

const statusOptions = [
  { label: '全部', value: '' },
  { label: '正常发布', value: '0' },
  { label: '隐藏', value: '1' },
  { label: '删除', value: '2' },
];

const categoryOptions = computed(() => [
  { label: '全部', value: '' },
  ...[...props.categories]
    .sort((left, right) => categoryOrder(left.slug) - categoryOrder(right.slug))
    .map((item) => ({
      label: categoryTitle(item.slug),
      value: item.slug,
    })),
]);
const TAG_PREVIEW_LIMIT = 8;
const showAllTags = ref(false);
const tagChoices = computed(() =>
  [...props.categories]
    .sort((left, right) => categoryOrder(left.slug) - categoryOrder(right.slug))
    .filter((item) => !category.value || item.slug === category.value)
    .flatMap((item) =>
      item.tags.map((tag) => ({
        label: `${tag.name} · ${categoryTitle(item.slug)}`,
        value: String(tag.id),
      })),
    ),
);
const tagOptions = computed(() => {
  const selectedValue = tagId.value == null ? '' : String(tagId.value);
  const visible = showAllTags.value
    ? tagChoices.value
    : tagChoices.value.slice(0, TAG_PREVIEW_LIMIT);
  const selected = tagChoices.value.find(
    (option) => option.value === selectedValue,
  );
  if (selected && !visible.some((option) => option.value === selectedValue)) {
    return [{ label: '全部', value: '' }, ...visible, selected];
  }
  return [{ label: '全部', value: '' }, ...visible];
});
const hasMoreTags = computed(() => tagChoices.value.length > TAG_PREVIEW_LIMIT);
const sortOptions: { label: string; value: PostSort }[] = [
  { label: '最近活跃', value: 'active' },
  { label: '最新发布', value: 'newest' },
  { label: '浏览最多', value: 'views' },
  { label: '评论最多', value: 'comments' },
];

function changeCategory(value: string) {
  showAllTags.value = false;
  category.value = value;
  if (
    tagId.value != null &&
    !props.categories.some(
      (item) =>
        (value === '' || item.slug === value) &&
        item.tags.some((tag) => tag.id === tagId.value),
    )
  )
    tagId.value = null;
  emit('search');
}

function changeTag(value: string) {
  const id = Number(value);
  tagId.value = value !== '' && Number.isSafeInteger(id) ? id : null;
  emit('search');
}

function changeStatus(value: string) {
  status.value = value as PostStatusFilter | '';
  emit('search');
}

function changeSort(value: string) {
  sort.value = value as PostSort;
  emit('search');
}
</script>

<template>
  <div class="filters">
    <FilterRow label="搜索">
      <XInput
        v-model="query"
        class="query-input"
        clearable
        placeholder="搜索标题或正文"
        @change="emit('search')"
        @clear="emit('search')"
      >
        <template #suffix>
          <SearchOutlined class="size-5" aria-hidden="true" />
        </template>
      </XInput>
    </FilterRow>

    <FilterRow label="作者">
      <XInput
        v-model="authorName"
        class="query-input"
        clearable
        placeholder="搜索作者名字"
        @change="emit('search')"
        @clear="emit('search')"
      />
    </FilterRow>

    <FilterRow label="标签">
      <div class="flex flex-col items-start gap-1">
        <FilterChoiceGroup
          :value="tagId == null ? '' : String(tagId)"
          :options="tagOptions"
          @update:value="changeTag"
        />
        <XButton
          v-if="hasMoreTags"
          color="neutral"
          variant="ghost"
          size="sm"
          :aria-expanded="showAllTags"
          @click="showAllTags = !showAllTags"
        >
          {{ showAllTags ? '收起' : '更多标签' }}
        </XButton>
      </div>
    </FilterRow>
    <FilterRow label="分类">
      <FilterChoiceGroup
        :value="category"
        :options="categoryOptions"
        @update:value="changeCategory"
      />
    </FilterRow>
    <FilterRow label="状态">
      <FilterChoiceGroup
        :value="status"
        :options="statusOptions"
        @update:value="changeStatus"
      />
    </FilterRow>
    <FilterRow label="排序">
      <FilterChoiceGroup
        :value="sort"
        :options="sortOptions"
        @update:value="changeSort"
      />
    </FilterRow>
  </div>
</template>

<style scoped>
.filters {
  display: grid;
  width: 100%;
  gap: 12px;
}

.query-input {
  width: min(400px, 100%);
}
</style>
