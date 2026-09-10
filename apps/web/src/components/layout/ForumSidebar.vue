<script setup lang="ts">
import { SmartToyOutlined } from '@vicons/material';

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
        <SmartToyOutlined class="size-5" />
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
