<script setup lang="ts">
import { XInput } from '@novelia/web-kit';
import FilterRow from '@/components/FilterRow.vue';
import FilterChoiceGroup from '@/components/FilterChoiceGroup.vue';
import type { CommentStatusFilter, CommentSubjectFilter } from '@/api';

const emit = defineEmits<{ search: [] }>();
const query = defineModel<string>('query', { required: true });
const authorName = defineModel<string>('authorName', { required: true });
const subjectType = defineModel<CommentSubjectFilter>('subjectType', {
  required: true,
});
const subjectKey = defineModel<string>('subjectKey', { required: true });
const subjectOptions = [
  { label: '全部', value: '' },
  { label: '帖子', value: '0' },
  { label: '小说', value: '1' },
];
const status = defineModel<CommentStatusFilter | ''>('status', {
  required: true,
});
const statusOptions = [
  { label: '全部', value: '' },
  { label: '正常发布', value: '0' },
  { label: '隐藏', value: '1' },
  { label: '删除', value: '2' },
];
function changeStatus(value: string) {
  status.value = value as CommentStatusFilter | '';
  emit('search');
}

function changeSubjectType(value: string) {
  subjectType.value = value as CommentSubjectFilter;
  emit('search');
}
</script>

<template>
  <div class="filters">
    <FilterRow label="搜索">
      <XInput
        v-model="query"
        class="field-input"
        clearable
        placeholder="搜索评论内容"
        @change="emit('search')"
        @clear="emit('search')"
        @keyup.enter="emit('search')"
      />
    </FilterRow>
    <FilterRow label="作者">
      <XInput
        v-model="authorName"
        class="field-input"
        clearable
        placeholder="搜索作者名字"
        @change="emit('search')"
        @clear="emit('search')"
        @keyup.enter="emit('search')"
      />
    </FilterRow>
    <FilterRow label="Subject">
      <FilterChoiceGroup
        :value="subjectType"
        :options="subjectOptions"
        @update:value="changeSubjectType"
      />
    </FilterRow>
    <FilterRow label="资源标识">
      <XInput
        v-model="subjectKey"
        class="field-input"
        clearable
        placeholder="精确匹配 subject key（帖子 ID / 小说标识）"
        @change="emit('search')"
        @clear="emit('search')"
        @keyup.enter="emit('search')"
      />
    </FilterRow>
    <FilterRow label="状态">
      <FilterChoiceGroup
        :value="status"
        :options="statusOptions"
        @update:value="changeStatus"
      />
    </FilterRow>
  </div>
</template>

<style scoped>
.filters {
  display: grid;
  gap: 12px;
}
.field-input {
  width: min(400px, 100%);
}
</style>
