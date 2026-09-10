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
      class="flex h-16 flex-none items-center"
      :class="collapsed ? 'justify-center px-2' : 'px-5'"
    >
      <span
        class="grid size-8 place-items-center rounded-md bg-primary-soft text-primary"
        aria-hidden="true"
      >
        <svg viewBox="0 0 24 24" class="size-5" fill="none">
          <path
            d="M12 4V2m-1 0h2M7 8h10a3 3 0 0 1 3 3v6a3 3 0 0 1-3 3H7a3 3 0 0 1-3-3v-6a3 3 0 0 1 3-3Z"
            stroke="currentColor"
            stroke-width="1.8"
            stroke-linecap="round"
            stroke-linejoin="round"
          />
          <circle cx="9" cy="14" r="1" fill="currentColor" />
          <circle cx="15" cy="14" r="1" fill="currentColor" />
        </svg>
      </span>
      <span
        v-if="!collapsed"
        class="ml-2.5 text-sm font-bold tracking-tight text-ink"
      >
        论坛
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
