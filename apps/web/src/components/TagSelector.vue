<script setup lang="ts">
import type { PostTag } from '@/api';

withDefaults(
  defineProps<{
    tags: PostTag[];
    label?: string;
    emptyText?: string;
    disabled?: boolean;
  }>(),
  {
    label: '标签（可选）',
    emptyText: '这个分类暂时没有可用标签。',
    disabled: false,
  },
);

const selectedIds = defineModel<number[]>({ required: true });
</script>

<template>
  <fieldset :disabled="disabled">
    <legend class="mb-2 text-sm font-semibold text-ink">{{ label }}</legend>
    <div v-if="tags.length" class="flex flex-wrap gap-2">
      <label
        v-for="tag in tags"
        :key="tag.id"
        class="cursor-pointer rounded-sm border px-3 py-1.5 text-xs font-medium transition-colors"
        :class="
          selectedIds.includes(tag.id)
            ? 'border-primary bg-primary-soft text-primary'
            : 'border-border text-muted hover:border-primary hover:text-primary'
        "
      >
        <input
          v-model="selectedIds"
          type="checkbox"
          class="sr-only"
          :value="tag.id"
        />
        {{ tag.name }}
      </label>
    </div>
    <p v-else class="text-sm text-muted">{{ emptyText }}</p>
  </fieldset>
</template>
