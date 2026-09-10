<script setup lang="ts">
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
        <svg
          v-if="category.slug === 'novel'"
          viewBox="0 0 24 24"
          class="size-4"
          fill="none"
        >
          <path
            d="M4.5 5.5c2.7-.7 5.2.1 7.5 2.2v11c-2.3-2.1-4.8-2.9-7.5-2.2v-11Zm15 0c-2.7-.7-5.2.1-7.5 2.2v11c2.3-2.1 4.8-2.9 7.5-2.2v-11Z"
            stroke="currentColor"
            stroke-width="1.7"
            stroke-linecap="round"
            stroke-linejoin="round"
          />
        </svg>
        <svg
          v-else-if="category.slug === 'guide'"
          viewBox="0 0 24 24"
          class="size-4"
          fill="none"
        >
          <circle
            cx="12"
            cy="12"
            r="8"
            stroke="currentColor"
            stroke-width="1.7"
          />
          <path
            d="m14.8 9.2-1.5 4.1-4.1 1.5 1.5-4.1 4.1-1.5Z"
            stroke="currentColor"
            stroke-width="1.7"
            stroke-linejoin="round"
          />
        </svg>
        <svg v-else viewBox="0 0 24 24" class="size-4" fill="none">
          <path
            d="M5 6.5h14v9H11l-4.5 3v-3H5v-9Z"
            stroke="currentColor"
            stroke-width="1.7"
            stroke-linejoin="round"
          />
          <path
            d="M8.5 10h7M8.5 12.5h4"
            stroke="currentColor"
            stroke-width="1.7"
            stroke-linecap="round"
          />
        </svg>
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
