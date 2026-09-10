<script setup lang="ts">
import {
  ExploreOutlined,
  ForumOutlined,
  MenuBookOutlined,
} from '@vicons/material';

import type { Category } from '@/api';

defineProps<{
  categories: Category[];
  selected?: string;
  collapsed?: boolean;
}>();

const emit = defineEmits<{
  select: [slug: string];
}>();
</script>

<template>
  <nav class="grid gap-1" aria-label="帖子分类">
    <button
      v-for="category in categories"
      :key="category.id"
      type="button"
      class="category-item"
      :class="[
        selected === category.slug
          ? 'bg-primary-soft text-primary'
          : 'text-ink hover:bg-paper',
        collapsed ? 'justify-center px-0 lg:justify-start lg:px-3' : 'px-3',
      ]"
      :aria-label="collapsed ? category.title : undefined"
      :title="collapsed ? category.title : undefined"
      :aria-pressed="selected === category.slug"
      @click="emit('select', category.slug)"
    >
      <span
        class="grid size-7 flex-none place-items-center rounded-md text-primary transition-colors duration-300"
        :class="selected === category.slug ? 'bg-surface' : ''"
        aria-hidden="true"
      >
        <MenuBookOutlined v-if="category.slug === 'novel'" class="size-4" />
        <ExploreOutlined v-else-if="category.slug === 'guide'" class="size-4" />
        <ForumOutlined v-else class="size-4" />
      </span>
      <span :class="collapsed ? 'hidden lg:block' : ''" class="truncate">
        {{ category.title }}
      </span>
    </button>
  </nav>
</template>

<style scoped>
.category-item {
  display: flex;
  min-height: 2.75rem;
  align-items: center;
  gap: 0.7rem;
  border-radius: 0.25rem;
  font-size: 0.875rem;
  font-weight: 600;
  text-align: left;
  transition:
    color 0.3s cubic-bezier(0.4, 0, 0.2, 1),
    background-color 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.category-item:focus-visible {
  outline: 2px solid var(--color-primary);
  outline-offset: 2px;
}
</style>
