<script setup lang="ts">
import { SearchOutlined } from '@vicons/material';
import { NButton, NIcon, NInputNumber } from 'naive-ui';

import FilterRow from '@/components/FilterRow.vue';

const emit = defineEmits<{ search: [] }>();
const postId = defineModel<number | null>('postId', { required: true });
</script>

<template>
  <FilterRow label="帖子">
    <div class="post-filter">
      <n-input-number
        v-model:value="postId"
        :min="1"
        :show-button="false"
        placeholder="输入帖子 ID"
        @keyup.enter="emit('search')"
      />
      <n-button
        type="primary"
        :disabled="postId == null"
        @click="emit('search')"
      >
        <template #icon><n-icon :component="SearchOutlined" /></template>
        查询
      </n-button>
    </div>
  </FilterRow>
</template>

<style scoped>
.post-filter {
  width: min(400px, 100%);
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: 10px;
}
</style>
