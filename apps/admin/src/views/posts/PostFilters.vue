<script setup lang="ts">
import { SearchOutlined } from '@vicons/material';
import { NIcon, NInput } from 'naive-ui';
import { computed } from 'vue';

import FilterChoiceGroup from '@/components/FilterChoiceGroup.vue';
import FilterRow from '@/components/FilterRow.vue';
import type { Category } from '@/api';

const props = defineProps<{ categories: Category[] }>();
const emit = defineEmits<{ search: [] }>();

const query = defineModel<string>('query', { required: true });
const category = defineModel<string>('category', { required: true });

const categoryOptions = computed(() => [
  { label: '全部', value: '' },
  ...props.categories.map((item) => ({
    label: item.slug,
    value: item.slug,
  })),
]);

function changeCategory(value: string) {
  category.value = value;
  emit('search');
}
</script>

<template>
  <div class="filters">
    <FilterRow label="搜索">
      <n-input
        v-model:value="query"
        class="query-input"
        clearable
        placeholder="搜索标题或正文"
        @change="emit('search')"
      >
        <template #suffix>
          <n-icon :component="SearchOutlined" />
        </template>
      </n-input>
    </FilterRow>

    <FilterRow label="分类">
      <FilterChoiceGroup
        :value="category"
        :options="categoryOptions"
        @update:value="changeCategory"
      />
    </FilterRow>
  </div>
</template>

<style scoped>
.filters {
  display: grid;
  width: 100%;
  gap: 16px;
}

.query-input {
  width: min(400px, 100%);
}
</style>
