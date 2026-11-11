<script setup lang="ts">
import { useTemplateRef } from 'vue';

import FilterRow from '@/components/FilterRow.vue';
import FilterChoiceGroup from '@/components/FilterChoiceGroup.vue';
import type { CommentStatusFilter, CommentSubjectFilter } from '@/api';

const emit = defineEmits<{ search: [] }>();
const queryInput = useTemplateRef<{ inputRef: HTMLInputElement | null }>(
  'queryInput',
);
const authorNameInput = useTemplateRef<{ inputRef: HTMLInputElement | null }>(
  'authorNameInput',
);
const subjectKeyInput = useTemplateRef<{ inputRef: HTMLInputElement | null }>(
  'subjectKeyInput',
);
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
      <UInput
        ref="queryInput"
        v-model="query"
        aria-label="搜索评论内容"
        class="field-input"
        placeholder="搜索评论内容"
        @change="emit('search')"
        @keyup.enter="emit('search')"
      >
        <template #trailing>
          <UButton
            v-if="query"
            color="neutral"
            variant="link"
            size="xs"
            square
            icon="i-lucide-x"
            aria-label="清除搜索"
            @mousedown.prevent
            @click="
              query = '';
              queryInput?.inputRef?.focus();
              emit('search');
            "
          />
        </template>
      </UInput>
    </FilterRow>
    <FilterRow label="作者">
      <UInput
        ref="authorNameInput"
        v-model="authorName"
        aria-label="搜索作者名字"
        class="field-input"
        placeholder="搜索作者名字"
        @change="emit('search')"
        @keyup.enter="emit('search')"
      >
        <template #trailing>
          <UButton
            v-if="authorName"
            color="neutral"
            variant="link"
            size="xs"
            square
            icon="i-lucide-x"
            aria-label="清除作者"
            @mousedown.prevent
            @click="
              authorName = '';
              authorNameInput?.inputRef?.focus();
              emit('search');
            "
          />
        </template>
      </UInput>
    </FilterRow>
    <FilterRow label="Subject">
      <FilterChoiceGroup
        :value="subjectType"
        :options="subjectOptions"
        @update:value="changeSubjectType"
      />
    </FilterRow>
    <FilterRow label="资源标识">
      <UInput
        ref="subjectKeyInput"
        v-model="subjectKey"
        aria-label="资源标识"
        class="field-input"
        placeholder="精确匹配 subject key（帖子 ID / 小说标识）"
        @change="emit('search')"
        @keyup.enter="emit('search')"
      >
        <template #trailing>
          <UButton
            v-if="subjectKey"
            color="neutral"
            variant="link"
            size="xs"
            square
            icon="i-lucide-x"
            aria-label="清除资源标识"
            @mousedown.prevent
            @click="
              subjectKey = '';
              subjectKeyInput?.inputRef?.focus();
              emit('search');
            "
          />
        </template>
      </UInput>
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
