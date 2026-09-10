<script setup lang="ts">
import type { Category } from '@/api';

import CategoryNavigation from './CategoryNavigation.vue';

defineProps<{
  categories: Category[];
  selected?: string;
  collapsed?: boolean;
  fullWidth?: boolean;
}>();

const emit = defineEmits<{
  select: [slug: string];
}>();
</script>

<template>
  <aside
    class="flex h-full flex-col border-r border-divider bg-surface transition-[width] duration-300"
    :class="fullWidth ? 'w-full' : collapsed ? 'w-16' : 'w-56'"
    aria-label="论坛导航"
  >
    <div
      class="flex h-16 flex-none items-center border-b border-divider"
      :class="collapsed ? 'justify-center px-2' : 'px-5'"
    >
      <span
        v-if="collapsed"
        class="grid size-8 place-items-center rounded-md bg-primary-soft text-primary"
        aria-hidden="true"
      >
        <svg viewBox="0 0 24 24" class="size-5" fill="none">
          <path
            d="M6.5 5.5h11v9h-6L7 18v-3.5h-.5v-9Z"
            stroke="currentColor"
            stroke-width="1.8"
            stroke-linejoin="round"
          />
        </svg>
      </span>
      <span v-else class="text-sm font-bold tracking-tight text-ink">
        讨论分类
      </span>
    </div>

    <div class="min-h-0 flex-1 overflow-y-auto p-2">
      <CategoryNavigation
        :categories="categories"
        :selected="selected"
        :collapsed="collapsed"
        @select="emit('select', $event)"
      />
    </div>
  </aside>
</template>
