<script setup lang="ts">
import type { PostTag } from '@/api';

defineProps<{
  tags: PostTag[];
  pinned?: boolean;
  categoryName?: string;
}>();

const tagColors = [
  'bg-primary-soft text-primary',
  'bg-blue-50 text-blue-600',
  'bg-orange-50 text-orange-600',
  'bg-purple-50 text-purple-600',
  'bg-red-50 text-red-600',
];

function tagClass(color: number) {
  return tagColors[Math.abs(color) % tagColors.length];
}
</script>

<template>
  <div
    v-if="categoryName || pinned || tags.length"
    class="flex flex-wrap items-center gap-1.5 text-xs"
  >
    <span v-if="categoryName" class="font-medium text-primary">
      {{ categoryName }}
    </span>
    <span
      v-if="pinned"
      class="rounded-sm bg-orange-50 px-2 py-0.5 font-medium text-orange-600"
    >
      置顶
    </span>
    <span
      v-for="tag in tags"
      :key="tag.id"
      class="rounded-sm px-2 py-0.5 font-medium"
      :class="tagClass(tag.color)"
    >
      {{ tag.name }}
    </span>
  </div>
</template>
