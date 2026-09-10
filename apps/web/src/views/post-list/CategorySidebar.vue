<script setup lang="ts">
import { RouterLink } from 'vue-router';

import type { Category } from '@/api';

import CategoryNavigation from './CategoryNavigation.vue';

const props = defineProps<{
  categories: Category[];
  selected: string;
}>();

const emit = defineEmits<{
  select: [slug: string];
}>();

function changeCategory(event: Event) {
  emit('select', (event.target as HTMLSelectElement).value);
}
</script>

<template>
  <div class="flex gap-3 md:hidden">
    <label class="relative min-w-0 flex-1">
      <span class="sr-only">选择帖子分类</span>
      <select
        :value="selected"
        class="min-h-10 w-full appearance-none rounded-sm border border-divider bg-surface py-2 pr-10 pl-3 text-sm font-medium text-ink focus:border-primary focus:outline-none focus:ring-2 focus:ring-primary/20"
        @change="changeCategory"
      >
        <option
          v-for="category in categories"
          :key="category.id"
          :value="category.slug"
        >
          {{ category.title }}
        </option>
      </select>
      <svg
        viewBox="0 0 20 20"
        class="pointer-events-none absolute top-1/2 right-3 size-4 -translate-y-1/2 text-muted"
        fill="none"
        aria-hidden="true"
      >
        <path
          d="m6 8 4 4 4-4"
          stroke="currentColor"
          stroke-width="1.7"
          stroke-linecap="round"
          stroke-linejoin="round"
        />
      </svg>
    </label>
    <RouterLink
      :to="{
        name: 'post-create',
        query: { category: props.selected },
      }"
      class="inline-flex min-h-10 flex-none items-center gap-1.5 rounded-sm bg-primary px-3.5 text-sm font-medium text-white transition-colors hover:bg-primary-hover focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary"
    >
      <svg viewBox="0 0 20 20" class="size-4" fill="none" aria-hidden="true">
        <path
          d="M10 4v12M4 10h12"
          stroke="currentColor"
          stroke-width="1.7"
          stroke-linecap="round"
        />
      </svg>
      发表帖子
    </RouterLink>
  </div>

  <aside
    class="sticky top-24 hidden self-start md:grid md:gap-3"
    aria-label="帖子分类"
  >
    <RouterLink
      :to="{
        name: 'post-create',
        query: { category: props.selected },
      }"
      class="inline-flex min-h-10 items-center justify-center gap-1.5 rounded-sm bg-primary px-2 text-sm font-medium text-white transition-colors hover:bg-primary-hover focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary"
    >
      <svg viewBox="0 0 20 20" class="size-4" fill="none" aria-hidden="true">
        <path
          d="M10 4v12M4 10h12"
          stroke="currentColor"
          stroke-width="1.7"
          stroke-linecap="round"
        />
      </svg>
      <span class="hidden lg:inline">发表帖子</span>
    </RouterLink>
    <div class="rounded-sm bg-surface p-2">
      <CategoryNavigation
        :categories="categories"
        :selected="selected"
        collapsed
        @select="emit('select', $event)"
      />
    </div>
  </aside>
</template>
